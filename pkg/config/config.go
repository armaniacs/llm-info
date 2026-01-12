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
	Cost         CostConfig    `yaml:"cost"`
}

// ConfigSource は設定ソースの種類を表す
type ConfigSource int

const (
	SourceDefault ConfigSource = iota
	SourceFile
	SourceEnv
	SourceCLI
)

// GatewayConfig は実行時に使用するゲートウェイ設定を表す
type GatewayConfig struct {
	Name    string        `yaml:"name"`
	URL     string        `yaml:"url"`
	APIKey  string        `yaml:"api_key"`
	Timeout time.Duration `yaml:"timeout"`

	// ソース追跡（JSON/YAML出力から除外）
	URLSource     ConfigSource `json:"-" yaml:"-"`
	APIKeySource  ConfigSource `json:"-" yaml:"-"`
	TimeoutSource ConfigSource `json:"-" yaml:"-"`
}

// GetURLSource はURLの設定ソースを返す
func (g *GatewayConfig) GetURLSource() ConfigSource {
	return g.URLSource
}

// GetAPIKeySource はAPIKeyの設定ソースを返す
func (g *GatewayConfig) GetAPIKeySource() ConfigSource {
	return g.APIKeySource
}

// GetTimeoutSource はTimeoutの設定ソースを返す
func (g *GatewayConfig) GetTimeoutSource() ConfigSource {
	return g.TimeoutSource
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

// CostConfig はコスト計算の設定です
type CostConfig struct {
	WarningThreshold float64              `yaml:"warning_threshold"` // Default: 0.05
	Pricing         map[string]Pricing    `yaml:"pricing"`
	Enabled         bool                  `yaml:"enabled"`
}

// Pricing はモデルごとの料金レートです
type Pricing struct {
	InputPricePer1K  float64 `yaml:"input_price_per_1k"`   // e.g., 0.0025 for gpt-4
	OutputPricePer1K float64 `yaml:"output_price_per_1k"`  // e.g., 0.01 for gpt-4
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
