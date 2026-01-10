package config

import (
	"time"
)

// Config はアプリケーション設定全体を表す
type Config struct {
	Gateways       []Gateway `yaml:"gateways"`
	DefaultGateway string    `yaml:"default_gateway"`
	Global         Global    `yaml:"global"`
}

// Gateway は個別のゲートウェイ設定を表す
type Gateway struct {
	Name    string        `yaml:"name"`
	URL     string        `yaml:"url"`
	APIKey  string        `yaml:"api_key"`
	Timeout time.Duration `yaml:"timeout"`
}

// Global はグローバル設定を表す
type Global struct {
	Timeout      time.Duration `yaml:"timeout"`
	OutputFormat string        `yaml:"output_format"`
	SortBy       string        `yaml:"sort_by"`
}

// GatewayConfig は実行時に使用するゲートウェイ設定を表す
type GatewayConfig struct {
	Name    string        `yaml:"name"`
	URL     string        `yaml:"url"`
	APIKey  string        `yaml:"api_key"`
	Timeout time.Duration `yaml:"timeout"`
}

// FileConfig は設定ファイルの構造体です（後方互換性のため）
type FileConfig struct {
	Gateways       []GatewayConfig `yaml:"gateways"`
	DefaultGateway string          `yaml:"default_gateway"`
	Common         CommonConfig    `yaml:"common"`
}

// CommonConfig は共通設定です（後方互換性のため）
type CommonConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	Output  OutputConfig  `yaml:"output"`
}

// OutputConfig は出力設定です（後方互換性のため）
type OutputConfig struct {
	Format string      `yaml:"format"`
	Table  TableConfig `yaml:"table"`
}

// TableConfig はテーブル出力設定です（後方互換性のため）
type TableConfig struct {
	AlwaysShow      []string `yaml:"always_show"`
	ShowIfAvailable []string `yaml:"show_if_available"`
}

// AppConfig はアプリケーション設定です
type AppConfig struct {
	// 現在の設定
	BaseURL string
	APIKey  string
	Timeout time.Duration

	// 設定ファイル関連
	ConfigFile string
	Gateway    string

	// 出力設定
	OutputFormat string
}

// NewAppConfig は新しいアプリケーション設定を作成します
func NewAppConfig() *AppConfig {
	return &AppConfig{
		Timeout:      10 * time.Second,
		OutputFormat: "table",
	}
}
