package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/your-org/llm-info/pkg/config"
	"gopkg.in/yaml.v3"
)

// GetConfigPath は設定ファイルのパスを返す
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "llm-info", "llm-info.yaml")
}

// LoadConfigFromFile はファイルから設定を読み込む
func LoadConfigFromFile(path string) (*config.Config, error) {
	if path == "" {
		path = GetConfigPath()
	}

	// ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfigToFile は設定をファイルに保存する
func SaveConfigToFile(cfg *config.Config, path string) error {
	if path == "" {
		path = GetConfigPath()
	}

	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getDefaultConfig はデフォルト設定を返す
func getDefaultConfig() *config.Config {
	return &config.Config{
		Gateways: []config.Gateway{
			{
				Name:    "default",
				URL:     "https://api.example.com",
				APIKey:  "",
				Timeout: 10 * 1000000000, // 10秒（ナノ秒）
			},
		},
		DefaultGateway: "default",
		Global: config.Global{
			Timeout:      10 * 1000000000, // 10秒（ナノ秒）
			OutputFormat: "table",
			SortBy:       "name",
		},
	}
}

// LoadLegacyConfigFromFile は古い形式の設定ファイルを読み込む（後方互換性）
func LoadLegacyConfigFromFile(path string) (*config.FileConfig, error) {
	if path == "" {
		path = GetConfigPath()
	}

	// ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getLegacyDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig config.FileConfig
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &fileConfig, nil
}

// getLegacyDefaultConfig は古い形式のデフォルト設定を返す
func getLegacyDefaultConfig() *config.FileConfig {
	return &config.FileConfig{
		Gateways: []config.GatewayConfig{
			{
				Name:    "default",
				URL:     "https://api.example.com/v1",
				APIKey:  "",
				Timeout: 10 * 1000000000, // 10秒（ナノ秒）
			},
		},
		DefaultGateway: "default",
		Common: config.CommonConfig{
			Timeout: 10 * 1000000000, // 10秒（ナノ秒）
			Output: config.OutputConfig{
				Format: "table",
				Table: config.TableConfig{
					AlwaysShow:      []string{"name"},
					ShowIfAvailable: []string{"max_tokens", "mode", "input_cost"},
				},
			},
		},
	}
}

// ConvertLegacyToNew は古い形式の設定を新しい形式に変換する
func ConvertLegacyToNew(legacy *config.FileConfig) *config.Config {
	if legacy == nil {
		return getDefaultConfig()
	}

	// ゲートウェイ設定の変換
	gateways := make([]config.Gateway, len(legacy.Gateways))
	for i, gw := range legacy.Gateways {
		gateways[i] = config.Gateway{
			Name:    gw.Name,
			URL:     gw.URL,
			APIKey:  gw.APIKey,
			Timeout: gw.Timeout,
		}
	}

	// グローバル設定の変換
	global := config.Global{
		Timeout:      10 * 1000000000, // デフォルト値
		OutputFormat: "table",         // デフォルト値
		SortBy:       "name",          // デフォルト値
	}

	if legacy.Common.Timeout > 0 {
		global.Timeout = legacy.Common.Timeout
	}

	if legacy.Common.Output.Format != "" {
		global.OutputFormat = legacy.Common.Output.Format
	}

	// タイムアウトが0または負の場合はデフォルト値を使用
	if global.Timeout <= 0 {
		global.Timeout = 10 * 1000000000
	}

	return &config.Config{
		Gateways:       gateways,
		DefaultGateway: legacy.DefaultGateway,
		Global:         global,
	}
}
