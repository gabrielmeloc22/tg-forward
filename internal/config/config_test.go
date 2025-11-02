package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"TG_USER_APP_ID":        "12345",
				"TG_USER_APP_HASH":      "abcdef123456",
				"TG_USER_PHONE":         "+1234567890",
				"TG_BOT_TOKEN":          "123456:ABC-DEF",
				"TG_BOT_TARGET_CHAT_ID": "987654321",
				"API_PORT":              "8080",
				"API_TOKEN":             "test-token",
			},
			wantErr: false,
		},
		{
			name: "invalid app_id",
			envVars: map[string]string{
				"TG_USER_APP_ID":        "not-a-number",
				"TG_USER_APP_HASH":      "abcdef123456",
				"TG_USER_PHONE":         "+1234567890",
				"TG_BOT_TOKEN":          "123456:ABC-DEF",
				"TG_BOT_TARGET_CHAT_ID": "987654321",
				"API_TOKEN":             "test-token",
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			envVars: map[string]string{
				"TG_USER_APP_ID": "12345",
			},
			wantErr: true,
		},
		{
			name: "target username instead of chat_id",
			envVars: map[string]string{
				"TG_USER_APP_ID":         "12345",
				"TG_USER_APP_HASH":       "abcdef123456",
				"TG_USER_PHONE":          "+1234567890",
				"TG_BOT_TOKEN":           "123456:ABC-DEF",
				"TG_BOT_TARGET_USERNAME": "@testuser",
				"API_TOKEN":              "test-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer os.Clearenv()

			_, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Telegram: TelegramConfig{
					User: UserConfig{
						AppID:   12345,
						AppHash: "test",
						Phone:   "+123",
					},
					Bot: BotConfig{
						Token:        "token",
						TargetChatID: 123,
					},
				},
				API: APIConfig{
					Port:  "8080",
					Token: "test-token",
				},
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			config: Config{
				Telegram: TelegramConfig{
					User: UserConfig{
						AppHash: "test",
						Phone:   "+123",
					},
					Bot: BotConfig{
						Token:        "token",
						TargetChatID: 123,
					},
				},
				API: APIConfig{
					Port:  "8080",
					Token: "test-token",
				},
			},
			wantErr: true,
		},
		{
			name: "missing bot token",
			config: Config{
				Telegram: TelegramConfig{
					User: UserConfig{
						AppID:   12345,
						AppHash: "test",
						Phone:   "+123",
					},
					Bot: BotConfig{
						TargetChatID: 123,
					},
				},
				API: APIConfig{
					Port:  "8080",
					Token: "test-token",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
