package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/pkg/config"
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
	Sources      map[string]config.ConfigSource
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

		// 古い形式の設定は新しい形式に変換済みなので、そのまま使用
		m.newConfig = legacyConfig
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
		Sources: make(map[string]config.ConfigSource),
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
	resolved.Sources["output_format"] = config.SourceDefault
	resolved.Sources["sort_by"] = config.SourceDefault
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
		resolved.Sources["output_format"] = config.SourceFile
	}

	if m.newConfig.Global.SortBy != "" {
		resolved.SortBy = m.newConfig.Global.SortBy
		resolved.Sources["sort_by"] = config.SourceFile
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
				resolved.Gateway.URLSource = config.SourceFile
				resolved.Gateway.APIKeySource = config.SourceFile
				resolved.Gateway.TimeoutSource = config.SourceFile
				resolved.Sources["gateway"] = config.SourceFile
				resolved.Sources["gateway.url"] = config.SourceFile
				resolved.Sources["gateway.api_key"] = config.SourceFile
				resolved.Sources["gateway.timeout"] = config.SourceFile
				break
			}
		}
	}

	return nil
}

// applyEnvConfig は環境変数から設定を適用する
func (m *Manager) applyEnvConfig(resolved *ResolvedConfig) error {
	envConfig := LoadEnvConfig()

	// Gatewayの初期化（初回のみ）
	if resolved.Gateway == nil {
		resolved.Gateway = &config.GatewayConfig{}
	}

	// 各フィールドを個別に設定
	if envConfig.URL != "" {
		resolved.Gateway.URL = envConfig.URL
		resolved.Gateway.URLSource = config.SourceEnv
		resolved.Sources["gateway.url"] = config.SourceEnv
	}

	if envConfig.APIKey != "" {
		resolved.Gateway.APIKey = envConfig.APIKey
		resolved.Gateway.APIKeySource = config.SourceEnv
		resolved.Sources["gateway.api_key"] = config.SourceEnv
	}

	if envConfig.TimeoutString != "" {
		// タイムアウトが設定されている場合（有効か無効かに関わらず）
		resolved.Gateway.Timeout = envConfig.Timeout
		resolved.Gateway.TimeoutSource = config.SourceEnv
		resolved.Sources["gateway.timeout"] = config.SourceEnv
	}

	// APIキーが空でもURLが設定されている場合は、ソース情報を保持
	if envConfig.URL != "" && envConfig.APIKey == "" && resolved.Sources["gateway"] == 0 {
		resolved.Sources["gateway"] = config.SourceEnv
	}

	if envConfig.OutputFormat != "" {
		resolved.OutputFormat = envConfig.OutputFormat
		resolved.Sources["output_format"] = config.SourceEnv
	}

	if envConfig.SortBy != "" {
		resolved.SortBy = envConfig.SortBy
		resolved.Sources["sort_by"] = config.SourceEnv
	}

	if envConfig.Filter != "" {
		resolved.Filter = envConfig.Filter
		resolved.Sources["filter"] = config.SourceEnv
	}

	if envConfig.LogLevel != "" {
		resolved.LogLevel = envConfig.LogLevel
		resolved.Sources["log_level"] = config.SourceEnv
	}

	if envConfig.UserAgent != "" {
		resolved.UserAgent = envConfig.UserAgent
		resolved.Sources["user_agent"] = config.SourceEnv
	}

	return nil
}

// applyCLIConfig はコマンドライン引数から設定を適用する
func (m *Manager) applyCLIConfig(resolved *ResolvedConfig, cliArgs *CLIArgs) error {
	if cliArgs == nil {
		return nil
	}

	// Gatewayの初期化（初回のみ）
	if resolved.Gateway == nil {
		resolved.Gateway = &config.GatewayConfig{}
	}

	// CLIで指定されたフィールドのみ更新
	if cliArgs.URL != "" {
		resolved.Gateway.URL = cliArgs.URL
		resolved.Gateway.URLSource = config.SourceCLI
		resolved.Sources["gateway.url"] = config.SourceCLI
	}

	if cliArgs.APIKey != "" {
		resolved.Gateway.APIKey = cliArgs.APIKey
		resolved.Gateway.APIKeySource = config.SourceCLI
		resolved.Sources["gateway.api_key"] = config.SourceCLI
	}

	if cliArgs.Timeout > 0 {
		resolved.Gateway.Timeout = cliArgs.Timeout
		resolved.Gateway.TimeoutSource = config.SourceCLI
		resolved.Sources["gateway.timeout"] = config.SourceCLI
	}

	// Gateway全体のソース情報（後方互換性）
	if cliArgs.URL != "" || cliArgs.APIKey != "" {
		resolved.Sources["gateway"] = config.SourceCLI
	} else if cliArgs.Gateway != "" {
		// ゲートウェイ名が指定された場合は設定ファイルから取得
		gatewayConfig, err := m.GetGatewayConfig(cliArgs.Gateway)
		if err != nil {
			return fmt.Errorf("failed to get gateway config: %w", err)
		}
		resolved.Gateway = gatewayConfig
		resolved.Sources["gateway"] = config.SourceCLI
	}

	// その他の設定
	if cliArgs.OutputFormat != "" {
		resolved.OutputFormat = cliArgs.OutputFormat
		resolved.Sources["output_format"] = config.SourceCLI
	}

	if cliArgs.SortBy != "" {
		resolved.SortBy = cliArgs.SortBy
		resolved.Sources["sort_by"] = config.SourceCLI
	}

	if cliArgs.Filter != "" {
		resolved.Filter = cliArgs.Filter
		resolved.Sources["filter"] = config.SourceCLI
	}

	if cliArgs.Columns != "" {
		resolved.Columns = cliArgs.Columns
		resolved.Sources["columns"] = config.SourceCLI
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

	// GatewayのURLも検証
	if !isValidURL(resolved.Gateway.URL) {
		return fmt.Errorf("invalid gateway URL: %q", resolved.Gateway.URL)
	}

	// 出力形式の検証
	if resolved.OutputFormat != "" {
		validFormats := []string{"table", "json"}
		if !contains(validFormats, resolved.OutputFormat) {
			return fmt.Errorf("invalid output format: %s (valid: %s)",
				resolved.OutputFormat, strings.Join(validFormats, ", "))
		}
	}

	// CLIで上書きされなかった環境変数の値を検証
	envConfig := LoadEnvConfig()

	// URLがCLIで上書きされていない場合に限り検証
	if resolved.Gateway != nil && resolved.Sources["gateway.url"] == config.SourceEnv && envConfig.URL != "" {
		if !isValidURL(envConfig.URL) {
			return fmt.Errorf("invalid LLM_INFO_URL from environment variables: %q", envConfig.URL)
		}
	}

	// TimeoutがCLIで上書きされていない場合に限り検証
	if resolved.Gateway != nil && resolved.Sources["gateway.timeout"] == config.SourceEnv && envConfig.TimeoutString != "" {
		// 値が設定されているがパースに失敗した場合
		if envConfig.Timeout == 0 && envConfig.TimeoutString != "" && envConfig.TimeoutString != "0" {
			return fmt.Errorf("invalid LLM_INFO_TIMEOUT from environment variables: %q", envConfig.TimeoutString)
		}
		// パース成功でも負の値は不正
		if envConfig.Timeout < 0 {
			return fmt.Errorf("LLM_INFO_TIMEOUT from environment variables must be positive, got: %v", envConfig.Timeout)
		}
	}

	// OutputFormatがCLIで上書きされていない場合に限り検証
	if resolved.Sources["output_format"] == config.SourceEnv && envConfig.OutputFormat != "" {
		validFormats := []string{"table", "json"}
		if !contains(validFormats, envConfig.OutputFormat) {
			return fmt.Errorf("invalid LLM_INFO_OUTPUT_FORMAT from environment variables: %s (valid: %s)",
				envConfig.OutputFormat, strings.Join(validFormats, ", "))
		}
	}

	return nil
}

// GetConfigSourceInfo は設定ソース情報を返す
func (m *Manager) GetConfigSourceInfo(resolved *ResolvedConfig) string {
	info := "Configuration sources:\n"

	// Gatewayの詳細なソース情報
	if resolved.Gateway != nil {
		if resolved.Gateway.URLSource > 0 {
			info += fmt.Sprintf("  gateway.url: %s\n", m.getSourceName(resolved.Gateway.URLSource))
		}
		if resolved.Gateway.APIKeySource > 0 {
			info += fmt.Sprintf("  gateway.api_key: %s\n", m.getSourceName(resolved.Gateway.APIKeySource))
		}
		if resolved.Gateway.TimeoutSource > 0 {
			info += fmt.Sprintf("  gateway.timeout: %s\n", m.getSourceName(resolved.Gateway.TimeoutSource))
		}
	}

	// その他の設定
	for key, source := range resolved.Sources {
		if !strings.HasPrefix(key, "gateway.") && key != "gateway" {
			info += fmt.Sprintf("  %s: %s\n", key, m.getSourceName(source))
		}
	}

	return info
}

// getSourceName はConfigSourceを文字列に変換
func (m *Manager) getSourceName(source config.ConfigSource) string {
	switch source {
	case config.SourceDefault:
		return "default"
	case config.SourceFile:
		return "config file"
	case config.SourceEnv:
		return "environment variable"
	case config.SourceCLI:
		return "command line"
	default:
		return "unknown"
	}
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

// SetNewConfig は新しい設定を設定します
func (m *Manager) SetNewConfig(cfg *config.Config) {
	m.newConfig = cfg
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
