package telegram

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type MessageHandler func(ctx context.Context, message *tg.Message) error

type Client struct {
	client        *telegram.Client
	phone         string
	handler       MessageHandler
	api           *tg.Client
	botID         int64
	sessionString string
	sessionStore  session.Storage
}

func NewClient(appID int, appHash, phone string, handler MessageHandler, botID int64, sessionString string) *Client {
	return &Client{
		phone:         phone,
		handler:       handler,
		botID:         botID,
		sessionString: sessionString,
	}
}

func (c *Client) Run(ctx context.Context, appID int, appHash string) error {
	dispatcher := tg.NewUpdateDispatcher()

	sessionStorage := &session.StorageMemory{}
	usingEnvSession := c.sessionString != ""

	if usingEnvSession {
		sessionData, err := decodeTelethonSession(c.sessionString)
		if err != nil {
			return fmt.Errorf("failed to decode session string: %w", err)
		}

		loader := session.Loader{Storage: sessionStorage}
		if err := loader.Save(ctx, sessionData); err != nil {
			return fmt.Errorf("failed to save decoded session: %w", err)
		}

		log.Println("Using Telethon session string from environment variable")
	} else {
		log.Println("No session string provided - will require 2FA authentication")
	}

	c.sessionStore = sessionStorage

	client := telegram.NewClient(appID, appHash, telegram.Options{
		UpdateHandler:  dispatcher,
		SessionStorage: sessionStorage,
	})
	c.client = client

	dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		if msg.Out {
			return nil
		}

		if peerUser, ok := msg.FromID.(*tg.PeerUser); ok {
			if peerUser.UserID == c.botID {
				log.Printf("Ignoring channel message from bot (ID: %d)", c.botID)
				return nil
			}
		}

		if c.handler != nil {
			return c.handler(ctx, msg)
		}
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		authStatus, err := c.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to check auth status: %w", err)
		}
		wasNotAuthorized := !authStatus.Authorized

		if err := c.authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		c.api = client.API()

		log.Println("Successfully authenticated as user")

		if wasNotAuthorized && !usingEnvSession {
			self, err := c.api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
			if err == nil {
				_ = self
			}

			if err := c.printSessionString(ctx); err != nil {
				log.Printf("Warning: Failed to print session string: %v", err)
			}
		}

		<-ctx.Done()
		return ctx.Err()
	})
}

func (c *Client) printSessionString(ctx context.Context) error {
	loader := session.Loader{Storage: c.sessionStore}
	sessionData, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	if sessionData.Addr == "" {
		cfg := c.client.Config()
		for _, dc := range cfg.DCOptions {
			if dc.ID == sessionData.DC {
				sessionData.Addr = fmt.Sprintf("%s:%d", dc.IPAddress, dc.Port)
				break
			}
		}
	}

	sessionString, err := encodeTelethonSession(sessionData)
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	separator := "================================================================================"
	fmt.Println("\n" + separator)
	fmt.Println("🔑 Session authenticated successfully!")
	fmt.Println(separator)
	fmt.Println("\nAdd this to your environment to avoid 2FA on every restart:")
	fmt.Printf("\nTG_USER_SESSION=%s\n", sessionString)
	fmt.Printf("\nSession size: %d characters\n", len(sessionString))
	fmt.Println("\n" + separator + "\n")

	return nil
}

func (c *Client) authenticate(ctx context.Context) error {
	flow := auth.NewFlow(
		auth.Constant(c.phone, "", auth.CodeAuthenticatorFunc(func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
			var code string
			fmt.Print("Enter code: ")
			if _, err := fmt.Scanln(&code); err != nil {
				return "", err
			}
			return code, nil
		})),
		auth.SendCodeOptions{},
	)

	if err := c.client.Auth().IfNecessary(ctx, flow); err != nil {
		return err
	}

	return nil
}

func decodeTelethonSession(sessionStr string) (*session.Data, error) {
	return session.TelethonSession(sessionStr)
}

func encodeTelethonSession(data *session.Data) (string, error) {
	if data.Addr == "" {
		return "", fmt.Errorf("session address is empty - session may not be fully initialized yet")
	}

	host, portStr, err := net.SplitHostPort(data.Addr)
	if err != nil {
		return "", fmt.Errorf("invalid address format: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("invalid port: %w", err)
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", host)
	}

	var buf []byte
	buf = append(buf, byte(data.DC))

	if ip4 := ip.To4(); ip4 != nil {
		buf = append(buf, ip4...)
	} else {
		buf = append(buf, ip.To16()...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	buf = append(buf, portBytes...)

	buf = append(buf, data.AuthKey...)

	return "1" + base64.URLEncoding.EncodeToString(buf), nil
}
