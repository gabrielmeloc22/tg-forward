package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Telegram TelegramConfig
	API      APIConfig
}

type TelegramConfig struct {
	User UserConfig
	Bot  BotConfig
}

type UserConfig struct {
	AppID   int
	AppHash string
	Phone   string
	Session string
}

type BotConfig struct {
	Token          string
	TargetChatID   int64
	TargetUsername string
}

type APIConfig struct {
	Port  string
	Token string
}

func Load() (*Config, error) {
	cfg := &Config{}

	appID, err := strconv.Atoi(getEnv("TG_USER_APP_ID", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid TG_USER_APP_ID: must be a number")
	}
	cfg.Telegram.User.AppID = appID
	cfg.Telegram.User.AppHash = getEnv("TG_USER_APP_HASH", "")
	cfg.Telegram.User.Phone = getEnv("TG_USER_PHONE", "")
	cfg.Telegram.User.Session = getEnv("TG_USER_SESSION", "")

	cfg.Telegram.Bot.Token = getEnv("TG_BOT_TOKEN", "")

	targetChatID := getEnv("TG_BOT_TARGET_CHAT_ID", "")
	if targetChatID != "" {
		chatID, err := strconv.ParseInt(targetChatID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid TG_BOT_TARGET_CHAT_ID: must be a number")
		}
		cfg.Telegram.Bot.TargetChatID = chatID
	}
	cfg.Telegram.Bot.TargetUsername = getEnv("TG_BOT_TARGET_USERNAME", "")

	cfg.API.Port = getEnv("API_PORT", "8080")
	cfg.API.Token = getEnv("API_TOKEN", "")

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.Telegram.User.AppID == 0 {
		return fmt.Errorf("telegram.user.app_id is required")
	}
	if c.Telegram.User.AppHash == "" {
		return fmt.Errorf("telegram.user.app_hash is required")
	}
	if c.Telegram.User.Phone == "" {
		return fmt.Errorf("telegram.user.phone is required")
	}
	if c.Telegram.Bot.Token == "" {
		return fmt.Errorf("telegram.bot.token is required")
	}
	if c.Telegram.Bot.TargetChatID == 0 && c.Telegram.Bot.TargetUsername == "" {
		return fmt.Errorf("either telegram.bot.target_chat_id or telegram.bot.target_username is required")
	}
	if c.API.Token == "" {
		return fmt.Errorf("api.token is required")
	}

	return nil
}
