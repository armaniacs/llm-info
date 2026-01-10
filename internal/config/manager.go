package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/armaniacs/llm-info/pkg/config"
)

// ConfigSource は設定ソースの種類を表す
type ConfigSource int

const (
	SourceDefault ConfigSource = iota
	SourceFile
	SourceEnv
	SourceCLI
)

// ResolvedConfig は解決された設定を表す
type ResolvedConfig struct {
	Gateway      *config.GatewayConfig
	OutputFormat string
	SortBy       string
	Filter       string
	Columns      string
	LogLevel     string
	UserAgent    string
	Sources      map[string]ConfigSource
}

// Manager は設定管理機能を提供します
type Manager struct {
	appConfig  *config.AppConfig
	fileConfig *config.FileConfig
	newConfig  *config.Config // 新しい形式の設定
	path       string
}

// NewManager は新しい設定マネージャーを作成します
func NewManager(configPath string) *Manager {
	return &Manager{
		appConfig: config.NewAppConfig(),
		path:      configPath,
	}
}

// NewManagerWithDefaults はデフォルト設定で新しい設定マネージャーを作成します
func NewManagerWithDefaults() *Manager {
	return &Manager{
		appConfig: config.NewAppConfig(),
		path:      GetDefaultConfigPath(),
	}
}

// Load は設定を読み込みます
func (m *Manager) Load() error {
	configPath := m.path
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	// まず新しい形式の設定ファイルを試す
	newConfig, err := LoadConfigFromFile(configPath)
	if err != nil {
		// 新しい形式で読み込めない場合は古い形式を試す
		legacyConfig, legacyErr := LoadLegacyConfigFromFile(configPath)
		if legacyErr != nil {
			return fmt.Errorf("failed to load config file (tried both new and legacy formats): %w", err)
		}

		// 古い形式の設定を検証
		if err := ValidateLegacyConfig(legacyConfig); err != nil {
			return fmt.Errorf("invalid legacy config: %w", err)
		}

		m.fileConfig = legacyConfig
		m.newConfig = ConvertLegacyToNew(legacyConfig)
	} else {
		// 新しい形式の設定を検証
		if err := ValidateConfig(newConfig); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		m.newConfig = newConfig
	}

	m.appConfig.ConfigFile = configPath
	return nil
}

// LoadFromFile はファイルから設定を読み込みます（後方互換性）
func (m *Manager) LoadFromFile(configPath string) error {
	m.path = configPath
	return m.Load()
}

// LoadFromEnv は環境変数から設定を読み込みます
func (m *Manager) LoadFromEnv() {
	if url := os.Getenv("LLM_INFO_URL"); url != "" {
		m.appConfig.BaseURL = url
	}
	if apiKey := os.Getenv("LLM_INFO_API_KEY"); apiKey != "" {
		m.appConfig.APIKey = apiKey
	}
	if gateway := os.Getenv("LLM_INFO_GATEWAY"); gateway != "" {
		m.appConfig.Gateway = gateway
	}
	if timeout := os.Getenv("LLM_INFO_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			m.appConfig.Timeout = duration
		}
	}
	if format := os.Getenv("LLM_INFO_OUTPUT_FORMAT"); format != "" {
		m.appConfig.OutputFormat = format
	}
}

// ResolveConfig はすべての設定ソースから最終的な設定を解決する
func (m *Manager) ResolveConfig(cliArgs *CLIArgs) (*ResolvedConfig, error) {
	resolved := &ResolvedConfig{
		Sources: make(map[string]ConfigSource),
	}

	// 1. デフォルト値を設定
	if err := m.applyDefaults(resolved); err != nil {
		return nil, err
	}

	// 2. 設定ファイルから適用
	if err := m.applyFileConfig(resolved); err != nil {
		return nil, err
	}

	// 3. 環境変数から適用
	if err := m.applyEnvConfig(resolved); err != nil {
		return nil, err
	}

	// 4. コマンドライン引数から適用
	if err := m.applyCLIConfig(resolved, cliArgs); err != nil {
		return nil, err
	}

	// 5. 最終的な設定の検証
	if err := m.validateResolvedConfig(resolved); err != nil {
		return nil, err
	}

	return resolved, nil
}

// applyDefaults はデフォルト値を適用する
func (m *Manager) applyDefaults(resolved *ResolvedConfig) error {
	resolved.OutputFormat = "table"
	resolved.SortBy = "name"
	resolved.Sources["output_format"] = SourceDefault
	resolved.Sources["sort_by"] = SourceDefault
	return nil
}

// applyFileConfig は設定ファイルから設定を適用する
func (m *Manager) applyFileConfig(resolved *ResolvedConfig) error {
	if m.newConfig == nil {
		return nil
	}

	// グローバル設定を適用
	if m.newConfig.Global.OutputFormat != "" {
		resolved.OutputFormat = m.newConfig.Global.OutputFormat
		resolved.Sources["output_format"] = SourceFile
	}

	if m.newConfig.Global.SortBy != "" {
		resolved.SortBy = m.newConfig.Global.SortBy
		resolved.Sources["sort_by"] = SourceFile
	}

	// デフォルトゲートウェイを適用（まだゲートウェイが設定されていない場合）
	if resolved.Gateway == nil && m.newConfig.DefaultGateway != "" {
		for _, gw := range m.newConfig.Gateways {
			if gw.Name == m.newConfig.DefaultGateway {
				resolved.Gateway = &config.GatewayConfig{
					Name:    gw.Name,
					URL:     gw.URL,
					APIKey:  gw.APIKey,
					Timeout: gw.Timeout,
				}
				resolved.Sources["gateway"] = SourceFile
				break
			}
		}
	}

	return nil
}

// applyEnvConfig は環境変数から設定を適用する
func (m *Manager) applyEnvConfig(resolved *ResolvedConfig) error {
	envConfig := LoadEnvConfig()

	if envConfig.URL != "" || envConfig.APIKey != "" {
		resolved.Gateway = &config.GatewayConfig{
			URL:     envConfig.URL,
			APIKey:  envConfig.APIKey,
			Timeout: envConfig.Timeout,
		}
		resolved.Sources["gateway"] = SourceEnv
	}

	if envConfig.OutputFormat != "" {
		resolved.OutputFormat = envConfig.OutputFormat
		resolved.Sources["output_format"] = SourceEnv
	}

	if envConfig.SortBy != "" {
		resolved.SortBy = envConfig.SortBy
		resolved.Sources["sort_by"] = SourceEnv
	}

	if envConfig.Filter != "" {
		resolved.Filter = envConfig.Filter
		resolved.Sources["filter"] = SourceEnv
	}

	if envConfig.LogLevel != "" {
		resolved.LogLevel = envConfig.LogLevel
		resolved.Sources["log_level"] = SourceEnv
	}

	if envConfig.UserAgent != "" {
		resolved.UserAgent = envConfig.UserAgent
		resolved.Sources["user_agent"] = SourceEnv
	}

	return nil
}

// applyCLIConfig はコマンドライン引数から設定を適用する
func (m *Manager) applyCLIConfig(resolved *ResolvedConfig, cliArgs *CLIArgs) error {
	if cliArgs == nil {
		return nil
	}

	// ゲートウェイ設定
	if cliArgs.URL != "" || cliArgs.APIKey != "" {
		// 既存のゲートウェイ設定があれば保存してからマージ
		if resolved.Gateway == nil {
			resolved.Gateway = &config.GatewayConfig{}
		}

		// ソースが既にENVである場合はマージを示す
		if resolved.Sources["gateway"] == SourceEnv {
			resolved.Sources["gateway"] = SourceCLI // CLIが優先
		} else {
			resolved.Sources["gateway"] = SourceCLI
		}

		if cliArgs.URL != "" {
			resolved.Gateway.URL = cliArgs.URL
		}

		if cliArgs.APIKey != "" {
			resolved.Gateway.APIKey = cliArgs.APIKey
		}

		if cliArgs.Timeout > 0 {
			resolved.Gateway.Timeout = cliArgs.Timeout
		}
	} else if cliArgs.Gateway != "" {
		// ゲートウェイ名が指定された場合は設定ファイルから取得
		gatewayConfig, err := m.GetGatewayConfig(cliArgs.Gateway)
		if err != nil {
			return fmt.Errorf("failed to get gateway config: %w", err)
		}
		resolved.Gateway = gatewayConfig
		resolved.Sources["gateway"] = SourceCLI
	}

	// その他の設定
	if cliArgs.OutputFormat != "" {
		resolved.OutputFormat = cliArgs.OutputFormat
		resolved.Sources["output_format"] = SourceCLI
	}

	if cliArgs.SortBy != "" {
		resolved.SortBy = cliArgs.SortBy
		resolved.Sources["sort_by"] = SourceCLI
	}

	if cliArgs.Filter != "" {
		resolved.Filter = cliArgs.Filter
		resolved.Sources["filter"] = SourceCLI
	}

	if cliArgs.Columns != "" {
		resolved.Columns = cliArgs.Columns
		resolved.Sources["columns"] = SourceCLI
	}

	return nil
}

// validateResolvedConfig は解決された設定を検証する
func (m *Manager) validateResolvedConfig(resolved *ResolvedConfig) error {
	if resolved.Gateway == nil {
		return fmt.Errorf("no gateway configuration found")
	}

	if resolved.Gateway.URL == "" {
		return fmt.Errorf("gateway URL is required")
	}

	return nil
}

// GetConfigSourceInfo は設定ソース情報を返す
func (m *Manager) GetConfigSourceInfo(resolved *ResolvedConfig) string {
	info := "Configuration sources:\n"

	for key, source := range resolved.Sources {
		sourceName := ""
		switch source {
		case SourceDefault:
			sourceName = "default"
		case SourceFile:
			sourceName = "config file"
		case SourceEnv:
			sourceName = "environment variable"
		case SourceCLI:
			sourceName = "command line"
		}
		info += fmt.Sprintf("  %s: %s\n", key, sourceName)
	}

	return info
}

// CLIArgs はコマンドライン引数を表す
type CLIArgs struct {
	URL          string
	APIKey       string
	Timeout      time.Duration
	Gateway      string
	OutputFormat string
	SortBy       string
	Filter       string
	Columns      string
}

// ApplyGateway は指定されたゲートウェイ設定を適用します
func (m *Manager) ApplyGateway(gatewayName string) error {
	var gateways []config.GatewayConfig

	// 新しい形式の設定があれば使用
	if m.newConfig != nil {
		// 新しい形式から古い形式に変換
		for _, gw := range m.newConfig.Gateways {
			gateways = append(gateways, config.GatewayConfig{
				Name:    gw.Name,
				URL:     gw.URL,
				APIKey:  gw.APIKey,
				Timeout: gw.Timeout,
			})
		}

		// ゲートウェイ名が指定されていない場合はデフォルトを使用
		if gatewayName == "" {
			gatewayName = m.newConfig.DefaultGateway
		}
	} else if m.fileConfig != nil {
		gateways = m.fileConfig.Gateways

		// ゲートウェイ名が指定されていない場合はデフォルトを使用
		if gatewayName == "" {
			gatewayName = m.fileConfig.DefaultGateway
		}
	} else {
		return fmt.Errorf("no config file loaded")
	}

	// 指定されたゲートウェイを検索
	for _, gateway := range gateways {
		if gateway.Name == gatewayName {
			// コマンドライン引数で上書きされていない場合のみ適用
			if m.appConfig.BaseURL == "" {
				m.appConfig.BaseURL = gateway.URL
			}
			if m.appConfig.APIKey == "" {
				m.appConfig.APIKey = gateway.APIKey
			}
			if m.appConfig.Timeout == 10*time.Second { // デフォルト値の場合のみ上書き
				m.appConfig.Timeout = gateway.Timeout
			}

			// グローバル設定を適用（新しい形式の場合）
			if m.newConfig != nil {
				if m.appConfig.OutputFormat == "table" { // デフォルト値の場合のみ上書き
					m.appConfig.OutputFormat = m.newConfig.Global.OutputFormat
				}
			}

			return nil
		}
	}

	return fmt.Errorf("gateway '%s' not found in config", gatewayName)
}

// GetConfig は現在のアプリケーション設定を返します
func (m *Manager) GetConfig() *config.AppConfig {
	return m.appConfig
}

// GetFileConfig はファイル設定を返します
func (m *Manager) GetFileConfig() *config.FileConfig {
	return m.fileConfig
}

// GetNewConfig は新しい形式の設定を返します
func (m *Manager) GetNewConfig() *config.Config {
	return m.newConfig
}

// SetBaseURL はベースURLを設定します
func (m *Manager) SetBaseURL(url string) {
	m.appConfig.BaseURL = url
}

// SetAPIKey はAPIキーを設定します
func (m *Manager) SetAPIKey(apiKey string) {
	m.appConfig.APIKey = apiKey
}

// SetTimeout はタイムアウトを設定します
func (m *Manager) SetTimeout(timeout time.Duration) {
	m.appConfig.Timeout = timeout
}

// SetGateway はゲートウェイ名を設定します
func (m *Manager) SetGateway(gateway string) {
	m.appConfig.Gateway = gateway
}

// SetOutputFormat は出力形式を設定します
func (m *Manager) SetOutputFormat(format string) {
	m.appConfig.OutputFormat = format
}

// GetGatewayConfig は指定されたゲートウェイ設定を返します
func (m *Manager) GetGatewayConfig(name string) (*config.GatewayConfig, error) {
	var gateways []config.GatewayConfig

	// 新しい形式の設定があれば使用
	if m.newConfig != nil {
		// 新しい形式から古い形式に変換
		for _, gw := range m.newConfig.Gateways {
			gateways = append(gateways, config.GatewayConfig{
				Name:    gw.Name,
				URL:     gw.URL,
				APIKey:  gw.APIKey,
				Timeout: gw.Timeout,
			})
		}
	} else if m.fileConfig != nil {
		gateways = m.fileConfig.Gateways
	} else {
		return nil, fmt.Errorf("no config file loaded")
	}

	// ゲートウェイ名が指定されていない場合はデフォルトを使用
	if name == "" {
		if m.newConfig != nil {
			name = m.newConfig.DefaultGateway
		} else if m.fileConfig != nil {
			name = m.fileConfig.DefaultGateway
		}
	}

	for _, gw := range gateways {
		if gw.Name == name {
			return &gw, nil
		}
	}

	return nil, fmt.Errorf("gateway '%s' not found", name)
}

// ListGateways は利用可能なゲートウェイ一覧を返します
func (m *Manager) ListGateways() []string {
	var gateways []config.GatewayConfig

	// 新しい形式の設定があれば使用
	if m.newConfig != nil {
		// 新しい形式から古い形式に変換
		for _, gw := range m.newConfig.Gateways {
			gateways = append(gateways, config.GatewayConfig{
				Name:    gw.Name,
				URL:     gw.URL,
				APIKey:  gw.APIKey,
				Timeout: gw.Timeout,
			})
		}
	} else if m.fileConfig != nil {
		gateways = m.fileConfig.Gateways
	} else {
		return nil
	}

	var names []string
	for _, gw := range gateways {
		names = append(names, gw.Name)
	}
	return names
}

// CreateExampleConfig は例設定ファイルを作成します
func (m *Manager) CreateExampleConfig() error {
	configPath := m.path
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	defaultConfig := getDefaultConfig()
	return SaveConfigToFile(defaultConfig, configPath)
}

// GetDefaultConfigPath はデフォルトの設定ファイルパスを返します
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "llm-info", "llm-info.yaml")
}

// ValidateConfig は設定の妥当性を検証します
func (m *Manager) ValidateConfig() error {
	if m.appConfig.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	// タイムアウトの妥当性チェック
	if m.appConfig.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// 出力形式の妥当性チェック
	validFormats := []string{"table", "json"}
	isValidFormat := false
	for _, format := range validFormats {
		if m.appConfig.OutputFormat == format {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("invalid output format: %s (valid formats: %v)", m.appConfig.OutputFormat, validFormats)
	}

	return nil
}
