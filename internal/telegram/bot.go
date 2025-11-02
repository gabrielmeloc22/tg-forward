package telegram

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api            *tgbotapi.BotAPI
	targetChatID   int64
	targetUsername string
}

func NewBot(token string, targetChatID int64, targetUsername string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized bot account: %s (ID: %d)", api.Self.UserName, api.Self.ID)

	return &Bot{
		api:            api,
		targetChatID:   targetChatID,
		targetUsername: targetUsername,
	}, nil
}

func (b *Bot) GetBotID() int64 {
	return b.api.Self.ID
}

func (b *Bot) ForwardMessage(text string) error {
	var msg tgbotapi.MessageConfig

	if b.targetChatID != 0 {
		msg = tgbotapi.NewMessage(b.targetChatID, text)
	} else {
		msg = tgbotapi.NewMessageToChannel("@"+b.targetUsername, text)
	}

	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message forwarded successfully")
	return nil
}
