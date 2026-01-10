package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvConfig は環境変数から読み込んだ設定を表す
type EnvConfig struct {
	URL            string
	APIKey         string
	Timeout        time.Duration
	DefaultGateway string
	OutputFormat   string
	SortBy         string
	Filter         string
	ConfigPath     string
	LogLevel       string
	UserAgent      string
}

// LoadEnvConfig は環境変数から設定を読み込む
func LoadEnvConfig() *EnvConfig {
	return &EnvConfig{
		URL:            os.Getenv("LLM_INFO_URL"),
		APIKey:         os.Getenv("LLM_INFO_API_KEY"),
		Timeout:        parseTimeout(os.Getenv("LLM_INFO_TIMEOUT")),
		DefaultGateway: os.Getenv("LLM_INFO_DEFAULT_GATEWAY"),
		OutputFormat:   os.Getenv("LLM_INFO_OUTPUT_FORMAT"),
		SortBy:         os.Getenv("LLM_INFO_SORT_BY"),
		Filter:         os.Getenv("LLM_INFO_FILTER"),
		ConfigPath:     os.Getenv("LLM_INFO_CONFIG_PATH"),
		LogLevel:       os.Getenv("LLM_INFO_LOG_LEVEL"),
		UserAgent:      os.Getenv("LLM_INFO_USER_AGENT"),
	}
}

// parseTimeout はタイムアウト文字列をtime.Durationに変換する
func parseTimeout(s string) time.Duration {
	if s == "" {
		return 0
	}

	// 秒数を指定
	if seconds, err := strconv.Atoi(s); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// duration形式を指定
	if duration, err := time.ParseDuration(s); err == nil {
		return duration
	}

	return 0
}

// IsSet は環境変数が設定されているかチェックする
func (e *EnvConfig) IsSet() bool {
	return e.URL != "" || e.APIKey != "" || e.DefaultGateway != ""
}

// ToConfig は環境変数設定をConfigに変換する
func (e *EnvConfig) ToConfig() *Config {
	return &Config{
		BaseURL: e.URL,
		APIKey:  e.APIKey,
		Timeout: e.Timeout,
	}
}

// Validate は環境変数設定を検証する
func (e *EnvConfig) Validate() error {
	if e.URL != "" {
		if _, err := url.Parse(e.URL); err != nil {
			return fmt.Errorf("invalid LLM_INFO_URL: %w", err)
		}
	}

	if e.Timeout < 0 {
		return fmt.Errorf("LLM_INFO_TIMEOUT must be positive")
	}

	if e.OutputFormat != "" {
		validFormats := []string{"table", "json"}
		if !contains(validFormats, e.OutputFormat) {
			return fmt.Errorf("invalid LLM_INFO_OUTPUT_FORMAT: %s (valid: %s)",
				e.OutputFormat, strings.Join(validFormats, ", "))
		}
	}

	return nil
}

// contains はスライスに文字列が含まれているかチェックする
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// NewEnvConfig は新しい環境変数設定を作成します（後方互換性）
func NewEnvConfig() *EnvConfig {
	return LoadEnvConfig()
}

// Load は環境変数を読み込みます（後方互換性）
func (e *EnvConfig) Load() {
	// 新しい実装では何もしない（コンストラクタで既に読み込んでいる）
}

// GetString は文字列値を取得します（後方互換性）
func (e *EnvConfig) GetString(key string, defaultValue string) string {
	switch key {
	case "base_url":
		if e.URL != "" {
			return e.URL
		}
	case "api_key":
		if e.APIKey != "" {
			return e.APIKey
		}
	case "gateway":
		if e.DefaultGateway != "" {
			return e.DefaultGateway
		}
	case "output_format":
		if e.OutputFormat != "" {
			return e.OutputFormat
		}
	case "config_file":
		if e.ConfigPath != "" {
			return e.ConfigPath
		}
	}
	return defaultValue
}

// GetDuration は時間間隔値を取得します（後方互換性）
func (e *EnvConfig) GetDuration(key string, defaultValue time.Duration) time.Duration {
	if key == "timeout" && e.Timeout > 0 {
		return e.Timeout
	}
	return defaultValue
}

// GetBool は真偽値を取得します（後方互換性）
func (e *EnvConfig) GetBool(key string, defaultValue bool) bool {
	// 現在の実装では真偽値の環境変数はサポートしていない
	return defaultValue
}

// IsSetKey はキーが設定されているかチェックします（後方互換性）
func (e *EnvConfig) IsSetKey(key string) bool {
	switch key {
	case "base_url":
		return e.URL != ""
	case "api_key":
		return e.APIKey != ""
	case "gateway":
		return e.DefaultGateway != ""
	case "output_format":
		return e.OutputFormat != ""
	case "config_file":
		return e.ConfigPath != ""
	}
	return false
}

// GetAll はすべての環境変数設定を返します（後方互換性）
func (e *EnvConfig) GetAll() map[string]string {
	result := make(map[string]string)
	if e.URL != "" {
		result["base_url"] = e.URL
	}
	if e.APIKey != "" {
		result["api_key"] = e.APIKey
	}
	if e.DefaultGateway != "" {
		result["gateway"] = e.DefaultGateway
	}
	if e.OutputFormat != "" {
		result["output_format"] = e.OutputFormat
	}
	if e.ConfigPath != "" {
		result["config_file"] = e.ConfigPath
	}
	return result
}

// PrintEnvHelp は環境変数のヘルプを表示します
func PrintEnvHelp() {
	fmt.Println("環境変数:")
	fmt.Println("  LLM_INFO_URL              LLMゲートウェイのベースURL")
	fmt.Println("  LLM_INFO_API_KEY          認証に使用するAPIキー")
	fmt.Println("  LLM_INFO_TIMEOUT          リクエストタイムアウト (例: 10s, 1m)")
	fmt.Println("  LLM_INFO_DEFAULT_GATEWAY  デフォルトゲートウェイ名")
	fmt.Println("  LLM_INFO_OUTPUT_FORMAT    出力形式 (table, json)")
	fmt.Println("  LLM_INFO_SORT_BY          ソート項目 (name, max_tokens, mode)")
	fmt.Println("  LLM_INFO_FILTER           フィルタ条件")
	fmt.Println("  LLM_INFO_CONFIG_PATH      設定ファイルのパス")
	fmt.Println("  LLM_INFO_LOG_LEVEL        ログレベル")
	fmt.Println("  LLM_INFO_USER_AGENT       ユーザーエージェント")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  export LLM_INFO_URL=https://api.example.com/v1")
	fmt.Println("  export LLM_INFO_API_KEY=your-api-key")
	fmt.Println("  export LLM_INFO_TIMEOUT=15s")
	fmt.Println("  export LLM_INFO_OUTPUT_FORMAT=json")
	fmt.Println("  export LLM_INFO_SORT_BY=max_tokens")
	fmt.Println("  export LLM_INFO_FILTER=gpt")
	fmt.Println("  llm-info")
}

// ValidateEnvVars は環境変数の妥当性を検証します
func ValidateEnvVars() error {
	// タイムアウトの検証
	if timeout := os.Getenv("LLM_INFO_TIMEOUT"); timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			return fmt.Errorf("invalid LLM_INFO_TIMEOUT value: %s (%w)", timeout, err)
		}
	}

	// 出力形式の検証
	if format := os.Getenv("LLM_INFO_OUTPUT_FORMAT"); format != "" {
		validFormats := []string{"table", "json"}
		isValid := false
		for _, validFormat := range validFormats {
			if format == validFormat {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid LLM_INFO_OUTPUT_FORMAT value: %s (valid formats: %v)", format, validFormats)
		}
	}

	// 設定ファイルの存在チェック
	if configFile := os.Getenv("LLM_INFO_CONFIG_FILE"); configFile != "" {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", configFile)
		}
	}

	return nil
}
