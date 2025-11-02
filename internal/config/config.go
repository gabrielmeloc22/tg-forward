package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	API      APIConfig      `yaml:"api"`
}

type TelegramConfig struct {
	User UserConfig `yaml:"user"`
	Bot  BotConfig  `yaml:"bot"`
}

type UserConfig struct {
	AppID       int    `yaml:"app_id"`
	AppHash     string `yaml:"app_hash"`
	Phone       string `yaml:"phone"`
	SessionFile string `yaml:"session_file"`
}

type BotConfig struct {
	Token          string `yaml:"token"`
	TargetChatID   int64  `yaml:"target_chat_id"`
	TargetUsername string `yaml:"target_username"`
}

type APIConfig struct {
	Port  string `yaml:"port"`
	Token string `yaml:"token"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
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
