package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	validConfig := `
telegram:
  user:
    app_id: 12345
    app_hash: "abcdef123456"
    phone: "+1234567890"
  bot:
    token: "123456:ABC-DEF"
    target_chat_id: 987654321
api:
  port: "8080"
  token: "test-token"
`

	invalidConfig := `
telegram:
  user:
    app_id: 0
`

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid config",
			content: validConfig,
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			content: "invalid: [yaml",
			wantErr: true,
		},
		{
			name:    "missing required fields",
			content: invalidConfig,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err := Load(configPath)
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
