package config

import (
	"os"
	"testing"
	"time"
)

func TestEnvConfig_Load(t *testing.T) {
	// テスト用の環境変数を設定
	testVars := map[string]string{
		"LLM_INFO_URL":           "https://test.example.com/v1",
		"LLM_INFO_API_KEY":       "test-api-key",
		"LLM_INFO_TIMEOUT":       "15s",
		"LLM_INFO_GATEWAY":       "test-gateway",
		"LLM_INFO_OUTPUT_FORMAT": "json",
		"LLM_INFO_CONFIG_FILE":   "/path/to/config.yaml",
	}

	// テスト前に環境変数を保存
	originalVars := make(map[string]string)
	for key, value := range testVars {
		originalVars[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// テスト後に環境変数を復元
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// 環境変数設定のテスト
	envConfig := NewEnvConfig()
	envConfig.Load()

	// すべての変数が読み込まれたことを確認
	if !envConfig.IsSet() {
		t.Error("environment variables should be set")
	}

	// 値の確認
	if envConfig.GetString("base_url", "") != "https://test.example.com/v1" {
		t.Errorf("Expected base_url 'https://test.example.com/v1', got %q", envConfig.GetString("base_url", ""))
	}
	if envConfig.GetString("api_key", "") != "test-api-key" {
		t.Errorf("Expected api_key 'test-api-key', got %q", envConfig.GetString("api_key", ""))
	}
	if envConfig.GetString("gateway", "") != "test-gateway" {
		t.Errorf("Expected gateway 'test-gateway', got %q", envConfig.GetString("gateway", ""))
	}
	if envConfig.GetString("output_format", "") != "json" {
		t.Errorf("Expected output_format 'json', got %q", envConfig.GetString("output_format", ""))
	}
	if envConfig.GetString("config_file", "") != "/path/to/config.yaml" {
		t.Errorf("Expected config_file '/path/to/config.yaml', got %q", envConfig.GetString("config_file", ""))
	}
}

func TestEnvConfig_GetStringBackwardCompatibility(t *testing.T) {
	// 環境変数を設定してテスト
	os.Setenv("LLM_INFO_URL", "test_value")
	defer os.Unsetenv("LLM_INFO_URL")

	envConfig := NewEnvConfig()

	// 設定されているキーのテスト
	if value := envConfig.GetString("base_url", "default"); value != "test_value" {
		t.Errorf("GetString() = %q, expected %q", value, "test_value")
	}

	// 設定されていないキーのテスト
	if value := envConfig.GetString("nonexistent", "default"); value != "default" {
		t.Errorf("GetString() with nonexistent key = %q, expected %q", value, "default")
	}
}

func TestEnvConfig_GetDurationBackwardCompatibility(t *testing.T) {
	// 環境変数を設定してテスト
	os.Setenv("LLM_INFO_TIMEOUT", "10s")
	defer os.Unsetenv("LLM_INFO_TIMEOUT")

	envConfig := NewEnvConfig()

	// 有効な時間間隔のテスト
	if duration := envConfig.GetDuration("timeout", 5*time.Second); duration != 10*time.Second {
		t.Errorf("GetDuration() = %v, expected %v", duration, 10*time.Second)
	}

	// 設定されていないキーのテスト
	if duration := envConfig.GetDuration("nonexistent", 5*time.Second); duration != 5*time.Second {
		t.Errorf("GetDuration() with nonexistent key = %v, expected %v", duration, 5*time.Second)
	}
}

func TestEnvConfig_GetBoolBackwardCompatibility(t *testing.T) {
	envConfig := NewEnvConfig()

	// 設定されていないキーのテスト（常にデフォルト値が返される）
	if value := envConfig.GetBool("nonexistent", true); value != true {
		t.Errorf("GetBool() with nonexistent key = %v, expected %v", value, true)
	}
}

func TestEnvConfig_IsSetBackwardCompatibility(t *testing.T) {
	// 環境変数を設定してテスト
	os.Setenv("LLM_INFO_URL", "test_value")
	defer os.Unsetenv("LLM_INFO_URL")

	envConfig := NewEnvConfig()

	// 設定されているキーのテスト
	if !envConfig.IsSetKey("base_url") {
		t.Error("IsSetKey() should return true for set key")
	}

	// 設定されていないキーのテスト
	if envConfig.IsSetKey("nonexistent") {
		t.Error("IsSetKey() should return false for nonexistent key")
	}
}

func TestEnvConfig_GetAllBackwardCompatibility(t *testing.T) {
	// 環境変数を設定してテスト
	os.Setenv("LLM_INFO_URL", "value1")
	os.Setenv("LLM_INFO_API_KEY", "value2")
	defer func() {
		os.Unsetenv("LLM_INFO_URL")
		os.Unsetenv("LLM_INFO_API_KEY")
	}()

	envConfig := NewEnvConfig()
	all := envConfig.GetAll()

	if len(all) != 2 {
		t.Errorf("GetAll() returned %d items, expected 2", len(all))
	}

	if all["base_url"] != "value1" {
		t.Errorf("GetAll()[base_url] = %q, expected %q", all["base_url"], "value1")
	}

	if all["api_key"] != "value2" {
		t.Errorf("GetAll()[api_key] = %q, expected %q", all["api_key"], "value2")
	}
}

func TestValidateEnvVars(t *testing.T) {
	// テスト用の環境変数を設定
	testVars := map[string]string{
		"LLM_INFO_TIMEOUT":       "15s",
		"LLM_INFO_OUTPUT_FORMAT": "json",
	}

	// テスト前に環境変数を保存
	originalVars := make(map[string]string)
	for key, value := range testVars {
		originalVars[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// テスト後に環境変数を復元
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// 有効な環境変数のテスト
	if err := ValidateEnvVars(); err != nil {
		t.Errorf("ValidateEnvVars() with valid vars returned error: %v", err)
	}

	// 無効なタイムアウトのテスト
	os.Setenv("LLM_INFO_TIMEOUT", "invalid")
	if err := ValidateEnvVars(); err == nil {
		t.Error("ValidateEnvVars() with invalid timeout should return error")
	}

	// 修正
	os.Setenv("LLM_INFO_TIMEOUT", "15s")

	// 無効な出力形式のテスト
	os.Setenv("LLM_INFO_OUTPUT_FORMAT", "invalid")
	if err := ValidateEnvVars(); err == nil {
		t.Error("ValidateEnvVars() with invalid output format should return error")
	}

	// 修正
	os.Setenv("LLM_INFO_OUTPUT_FORMAT", "json")

	// 存在しない設定ファイルのテスト
	os.Setenv("LLM_INFO_CONFIG_FILE", "/nonexistent/config.yaml")
	if err := ValidateEnvVars(); err == nil {
		t.Error("ValidateEnvVars() with nonexistent config file should return error")
	}
}

func TestPrintEnvHelp(t *testing.T) {
	// このテストはヘルプメッセージがパニックを起こさないことを確認するだけ
	// 実際の出力内容は手動で確認する必要がある
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintEnvHelp() panicked: %v", r)
		}
	}()

	PrintEnvHelp()
}

func TestLoadEnvConfig(t *testing.T) {
	// テスト用の環境変数を設定
	testVars := map[string]string{
		"LLM_INFO_URL":             "https://test.example.com",
		"LLM_INFO_API_KEY":         "test-key",
		"LLM_INFO_TIMEOUT":         "30s",
		"LLM_INFO_DEFAULT_GATEWAY": "test-gateway",
		"LLM_INFO_OUTPUT_FORMAT":   "json",
		"LLM_INFO_SORT_BY":         "max_tokens",
		"LLM_INFO_FILTER":          "gpt",
		"LLM_INFO_CONFIG_PATH":     "/path/to/config",
		"LLM_INFO_LOG_LEVEL":       "debug",
		"LLM_INFO_USER_AGENT":      "test-agent",
	}

	// テスト前に環境変数を保存
	originalVars := make(map[string]string)
	for key, value := range testVars {
		originalVars[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	// テスト後に環境変数を復元
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	envConfig := LoadEnvConfig()

	// 値が正しく読み込まれたか確認
	if envConfig.URL != "https://test.example.com" {
		t.Errorf("Expected URL 'https://test.example.com', got '%s'", envConfig.URL)
	}

	if envConfig.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", envConfig.APIKey)
	}

	if envConfig.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", envConfig.Timeout)
	}

	if envConfig.DefaultGateway != "test-gateway" {
		t.Errorf("Expected default gateway 'test-gateway', got '%s'", envConfig.DefaultGateway)
	}

	if envConfig.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got '%s'", envConfig.OutputFormat)
	}

	if envConfig.SortBy != "max_tokens" {
		t.Errorf("Expected sort by 'max_tokens', got '%s'", envConfig.SortBy)
	}

	if envConfig.Filter != "gpt" {
		t.Errorf("Expected filter 'gpt', got '%s'", envConfig.Filter)
	}

	if envConfig.ConfigPath != "/path/to/config" {
		t.Errorf("Expected config path '/path/to/config', got '%s'", envConfig.ConfigPath)
	}

	if envConfig.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", envConfig.LogLevel)
	}

	if envConfig.UserAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got '%s'", envConfig.UserAgent)
	}
}

func TestParseTimeout(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"", 0},
		{"10", 10 * time.Second},
		{"10s", 10 * time.Second},
		{"1m", 1 * time.Minute},
		{"invalid", 0},
	}

	for _, test := range tests {
		result := parseTimeout(test.input)
		if result != test.expected {
			t.Errorf("parseTimeout(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestEnvConfig_IsSetNew(t *testing.T) {
	// 環境変数が設定されていない場合
	envConfig := LoadEnvConfig()
	if envConfig.IsSet() {
		t.Error("Expected IsSet to return false when no environment variables are set")
	}

	// 環境変数を設定
	os.Setenv("LLM_INFO_URL", "https://test.example.com")
	defer os.Unsetenv("LLM_INFO_URL")

	envConfig = LoadEnvConfig()
	if !envConfig.IsSet() {
		t.Error("Expected IsSet to return true when environment variables are set")
	}
}

func TestEnvConfig_ToConfig(t *testing.T) {
	os.Setenv("LLM_INFO_URL", "https://test.example.com")
	os.Setenv("LLM_INFO_API_KEY", "test-key")
	os.Setenv("LLM_INFO_TIMEOUT", "30s")
	defer func() {
		os.Unsetenv("LLM_INFO_URL")
		os.Unsetenv("LLM_INFO_API_KEY")
		os.Unsetenv("LLM_INFO_TIMEOUT")
	}()

	envConfig := LoadEnvConfig()
	config := envConfig.ToConfig()

	if config.BaseURL != "https://test.example.com" {
		t.Errorf("Expected BaseURL 'https://test.example.com', got '%s'", config.BaseURL)
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got '%s'", config.APIKey)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", config.Timeout)
	}
}

func TestEnvConfig_ValidateNew(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "valid config",
			envVars: map[string]string{"LLM_INFO_URL": "https://example.com"},
			wantErr: false,
		},
		{
			name:    "invalid URL",
			envVars: map[string]string{"LLM_INFO_URL": "invalid-url"},
			wantErr: true,
		},
		{
			name:    "negative timeout",
			envVars: map[string]string{"LLM_INFO_TIMEOUT": "-1s"},
			wantErr: true,
		},
		{
			name:    "invalid output format",
			envVars: map[string]string{"LLM_INFO_OUTPUT_FORMAT": "invalid"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 環境変数を設定
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range test.envVars {
					os.Unsetenv(key)
				}
			}()

			envConfig := LoadEnvConfig()
			err := envConfig.Validate()

			if (err != nil) != test.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
