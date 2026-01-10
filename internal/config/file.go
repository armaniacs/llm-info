package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/armaniacs/llm-info/pkg/config"
	"gopkg.in/yaml.v3"
	"errors"
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
func LoadLegacyConfigFromFile(path string) (*config.Config, error) {
	if path == "" {
		path = GetConfigPath()
	}

	// ファイルが存在しない場合はデフォルト設定を返す
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getLegacyDefaultConfigAsNewFormat(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// まずレガー形式を試す
	if cfg, err := tryLegacyFormats(data, path); err == nil {
		return cfg, nil
	}

	// 次に新しい形式を試す
	var newConfig config.Config
	if err := yaml.Unmarshal(data, &newConfig); err == nil {
		return &newConfig, nil
	}

	return nil, enhanceParseError(path, fmt.Errorf("failed to parse config file as any known format"))
}

// tryLegacyFormats は複数のレガー形式を試行
func tryLegacyFormats(data []byte, path string) (*config.Config, error) {
		formats := []func([]byte) (*config.Config, error){
		tryLegacyFlatFormat,
		tryLegacyNestedFormat,
		tryLegacyFileConfigFormat,
	}

	for i, formatFunc := range formats {
				if config, err := formatFunc(data); err == nil {
			// 成功した場合は移行を提案
			log.Printf("Successfully parsed legacy config using format %d from %s", i+1, path)
			log.Printf("Consider migrating to the new format for better compatibility")
			return config, nil
		}
	}

	return nil, enhanceParseError(path, fmt.Errorf("failed to parse config file as any known legacy format"))
}

// tryLegacyFlatFormat - 単純なキー・バリュー形式
func tryLegacyFlatFormat(data []byte) (*config.Config, error) {
		var legacy struct {
		BaseURL     string        `yaml:"base_url"`
		URL         string        `yaml:"url"`
		APIKey      string        `yaml:"api_key"`
		Key         string        `yaml:"key"`
		Timeout     string        `yaml:"timeout"`
		OutputFormat string        `yaml:"output_format"`
		SortBy      string        `yaml:"sort_by"`
	}

	if err := yaml.Unmarshal(data, &legacy); err != nil {
				return nil, err
	}

	// 新形式に変換
	cfg := &config.Config{
		DefaultGateway: "default",
		Global: config.Global{
			OutputFormat: legacy.OutputFormat,
			SortBy:       legacy.SortBy,
		},
	}

	// Timeoutのパース
	var timeout time.Duration
	if legacy.Timeout != "" {
		if parsed, err := time.ParseDuration(legacy.Timeout); err == nil {
			timeout = parsed
			cfg.Global.Timeout = timeout
		}
	}

	// URLの処理
	url := legacy.URL
	if url == "" {
		url = legacy.BaseURL
	}

	if url != "" {
		cfg.Gateways = []config.Gateway{
			{
				Name:    "default",
				URL:     url,
				APIKey:  legacy.APIKey,
				Timeout: timeout,
			},
		}
	}

	return cfg, nil
}

// tryLegacyNestedFormat - ネストされた形式
func tryLegacyNestedFormat(data []byte) (*config.Config, error) {
		var legacy struct {
		LLMInfo struct {
			BaseURL      string        `yaml:"base_url"`
			URL          string        `yaml:"url"`
			APIKey       string        `yaml:"api_key"`
			Timeout      time.Duration `yaml:"timeout"`
			OutputFormat string        `yaml:"output_format"`
			SortBy       string        `yaml:"sort_by"`
		} `yaml:"llm_info"`
	}

	if err := yaml.Unmarshal(data, &legacy); err != nil {
				return nil, err
	}

	
	// 新形式に変換
	cfg := &config.Config{
		DefaultGateway: "default",
		Global: config.Global{
			OutputFormat: legacy.LLMInfo.OutputFormat,
			SortBy:       legacy.LLMInfo.SortBy,
			Timeout:      legacy.LLMInfo.Timeout,
		},
	}

	url := legacy.LLMInfo.URL
	if url == "" {
		url = legacy.LLMInfo.BaseURL
	}

	if url != "" {
		cfg.Gateways = []config.Gateway{
			{
				Name:    "default",
				URL:     url,
				APIKey:  legacy.LLMInfo.APIKey,
				Timeout: legacy.LLMInfo.Timeout,
			},
		}
	}

	return cfg, nil
}

// tryLegacyFileConfigFormat - FileConfig構造体を使用
func tryLegacyFileConfigFormat(data []byte) (*config.Config, error) {
	var fileConfig config.FileConfig
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return nil, err
	}

	// 新形式に変換
	cfg := &config.Config{
		DefaultGateway: fileConfig.DefaultGateway,
		Global: config.Global{
			Timeout:      10 * 1000000000, // デフォルト値
			OutputFormat: "table",         // デフォルト値
			SortBy:       "name",          // デフォルト値
		},
	}

	// Gatewaysを変換
	if len(fileConfig.Gateways) > 0 {
		cfg.Gateways = make([]config.Gateway, len(fileConfig.Gateways))
		for i, gw := range fileConfig.Gateways {
			cfg.Gateways[i] = config.Gateway{
				Name:    gw.Name,
				URL:     gw.URL,
				APIKey:  gw.APIKey,
				Timeout: gw.Timeout,
			}
		}
	}

	// Common設定をGlobalに変換
	if fileConfig.Common.Timeout > 0 {
		cfg.Global.Timeout = fileConfig.Common.Timeout
	}
	if fileConfig.Common.Output.Format != "" {
		cfg.Global.OutputFormat = fileConfig.Common.Output.Format
	}

	return cfg, nil
}

// getLegacyDefaultConfigAsNewFormat は古い形式のデフォルト設定をConfig形式で返す
func getLegacyDefaultConfigAsNewFormat() *config.Config {
	return &config.Config{
		Gateways: []config.Gateway{
			{
				Name: "default",
				URL:  "https://api.example.com/v1",
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

// enhanceParseError は解析エラーを向上させる
func enhanceParseError(path string, originalErr error) error {
	var typeErr *yaml.TypeError

	if errors.As(originalErr, &typeErr) {
		return fmt.Errorf("YAML structure error in %s. Please check the format:\n%v",
			path, originalErr)
	}

	return fmt.Errorf("failed to parse config file %s: %w", path, originalErr)
}

// LoadLegacyConfigFromString は文字列からレガシー設定を読み込む（テスト用）
func LoadLegacyConfigFromString(content string) (*config.Config, error) {
	data := []byte(content)

	// まずレガー形式を試す
	if cfg, err := tryLegacyFormats(data, "string"); err == nil {
		return cfg, nil
	}

	// 次に新しい形式を試す
	var newConfig config.Config
	if err := yaml.Unmarshal(data, &newConfig); err == nil {
		return &newConfig, nil
	}

	return nil, fmt.Errorf("failed to parse config as any known format")
}
