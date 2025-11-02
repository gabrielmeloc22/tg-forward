package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

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
	sessionPath   string
	sessionString string
}

func NewClient(appID int, appHash, phone string, handler MessageHandler, botID int64, sessionPath, sessionString string) *Client {
	return &Client{
		phone:         phone,
		handler:       handler,
		botID:         botID,
		sessionPath:   sessionPath,
		sessionString: sessionString,
	}
}

func (c *Client) Run(ctx context.Context, appID int, appHash string) error {
	dispatcher := tg.NewUpdateDispatcher()

	var sessionStorage session.Storage
	if c.sessionString != "" {
		data, err := base64.StdEncoding.DecodeString(c.sessionString)
		if err != nil {
			return fmt.Errorf("failed to decode session string: %w", err)
		}
		sessionStorage = &sessionFromEnv{data: data}
		log.Println("Using session from environment variable")
	} else {
		sessionStorage = &session.FileStorage{Path: c.sessionPath}
		log.Printf("Using session file: %s", c.sessionPath)
	}

	client := telegram.NewClient(appID, appHash, telegram.Options{
		UpdateHandler:  dispatcher,
		SessionStorage: sessionStorage,
	})
	c.client = client

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			return nil
		}

		if msg.Out {
			return nil
		}

		if peerUser, ok := msg.FromID.(*tg.PeerUser); ok {
			if peerUser.UserID == c.botID {
				log.Printf("Ignoring message from bot (ID: %d)", c.botID)
				return nil
			}
		}

		if c.handler != nil {
			return c.handler(ctx, msg)
		}
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := c.authenticate(ctx); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		c.api = client.API()

		log.Println("Successfully authenticated as user")

		<-ctx.Done()
		return ctx.Err()
	})
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

type sessionFromEnv struct {
	data []byte
}

func (s *sessionFromEnv) LoadSession(ctx context.Context) ([]byte, error) {
	return s.data, nil
}

func (s *sessionFromEnv) StoreSession(ctx context.Context, data []byte) error {
	s.data = data
	return nil
}
