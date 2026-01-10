package config

import "time"

// Config はアプリケーション設定を保持します
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// New は新しい設定を作成します
func New(baseURL, apiKey string, timeout time.Duration) *Config {
	return &Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: timeout,
	}
}
