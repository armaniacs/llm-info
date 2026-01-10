package config

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		apiKey  string
		timeout time.Duration
	}{
		{
			name:    "basic config",
			baseURL: "https://api.example.com",
			apiKey:  "test-key",
			timeout: 10 * time.Second,
		},
		{
			name:    "empty API key",
			baseURL: "https://gateway.example.com/v1",
			apiKey:  "",
			timeout: 5 * time.Second,
		},
		{
			name:    "long timeout",
			baseURL: "https://slow-api.example.com",
			apiKey:  "slow-key",
			timeout: 60 * time.Second,
		},
		{
			name:    "zero timeout",
			baseURL: "https://fast-api.example.com",
			apiKey:  "fast-key",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := New(tt.baseURL, tt.apiKey, tt.timeout)

			if cfg.BaseURL != tt.baseURL {
				t.Errorf("New() BaseURL = %q, expected %q", cfg.BaseURL, tt.baseURL)
			}

			if cfg.APIKey != tt.apiKey {
				t.Errorf("New() APIKey = %q, expected %q", cfg.APIKey, tt.apiKey)
			}

			if cfg.Timeout != tt.timeout {
				t.Errorf("New() Timeout = %v, expected %v", cfg.Timeout, tt.timeout)
			}
		})
	}
}

func TestConfigFields(t *testing.T) {
	baseURL := "https://test.example.com"
	apiKey := "test-api-key"
	timeout := 30 * time.Second

	cfg := New(baseURL, apiKey, timeout)

	// 構造体のフィールドが正しく設定されていることを確認
	if cfg.BaseURL == "" {
		t.Error("BaseURL should not be empty")
	}

	if cfg.APIKey == "" && apiKey != "" {
		t.Error("APIKey should be set when provided")
	}

	if cfg.Timeout == 0 && timeout != 0 {
		t.Error("Timeout should be set when provided")
	}
}

func TestConfigWithSpecialCharacters(t *testing.T) {
	baseURL := "https://api.example.com/v1/special?param=value"
	apiKey := "special-key-with-123456789!@#$%^&*()"
	timeout := 15 * time.Second

	cfg := New(baseURL, apiKey, timeout)

	if cfg.BaseURL != baseURL {
		t.Errorf("New() with special characters BaseURL = %q, expected %q", cfg.BaseURL, baseURL)
	}

	if cfg.APIKey != apiKey {
		t.Errorf("New() with special characters APIKey = %q, expected %q", cfg.APIKey, apiKey)
	}

	if cfg.Timeout != timeout {
		t.Errorf("New() with special characters Timeout = %v, expected %v", cfg.Timeout, timeout)
	}
}
