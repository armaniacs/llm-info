package integration

import (
	"os"
	"testing"
	"time"

	"github.com/armaniacs/llm-info/internal/config"
)

func TestConfigPriority(t *testing.T) {
	// テスト用の環境変数を設定
	testEnvVars := map[string]string{
		"LLM_INFO_URL":           "https://env.example.com",
		"LLM_INFO_API_KEY":       "env-api-key",
		"LLM_INFO_TIMEOUT":       "30s",
		"LLM_INFO_OUTPUT_FORMAT": "json",
		"LLM_INFO_SORT_BY":       "max_tokens",
		"LLM_INFO_FILTER":        "gpt",
	}

	// テスト前に環境変数を保存
	originalVars := make(map[string]string)
	for key, value := range testEnvVars {
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

	tests := []struct {
		name     string
		cliArgs  *config.CLIArgs
		expected map[string]interface{}
	}{
		{
			name:    "environment variables only",
			cliArgs: &config.CLIArgs{},
			expected: map[string]interface{}{
				"url":           "https://env.example.com",
				"api_key":       "env-api-key",
				"timeout":       30 * time.Second,
				"output_format": "json",
				"sort_by":       "max_tokens",
				"filter":        "gpt",
			},
		},
		{
			name: "CLI overrides environment",
			cliArgs: &config.CLIArgs{
				URL:          "https://cli.example.com",
				APIKey:       "cli-api-key",
				Timeout:      60 * time.Second,
				OutputFormat: "table",
				SortBy:       "name",
				Filter:       "claude",
			},
			expected: map[string]interface{}{
				"url":           "https://cli.example.com",
				"api_key":       "cli-api-key",
				"timeout":       60 * time.Second,
				"output_format": "table",
				"sort_by":       "name",
				"filter":        "claude",
			},
		},
		{
			name: "partial CLI override",
			cliArgs: &config.CLIArgs{
				URL:          "https://cli.example.com",
				OutputFormat: "table",
			},
			expected: map[string]interface{}{
				"url":           "https://cli.example.com",
				"api_key":       "env-api-key",
				"timeout":       30 * time.Second,
				"output_format": "table",
				"sort_by":       "max_tokens",
				"filter":        "gpt",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 設定マネージャーの初期化
			configManager := config.NewManager("")

			// 設定の解決
			resolvedConfig, err := configManager.ResolveConfig(test.cliArgs)
			if err != nil {
				t.Fatalf("ResolveConfig() error = %v", err)
			}

			// 結果の検証
			if resolvedConfig.Gateway.URL != test.expected["url"] {
				t.Errorf("URL = %v, expected %v", resolvedConfig.Gateway.URL, test.expected["url"])
			}

			if resolvedConfig.Gateway.APIKey != test.expected["api_key"] {
				t.Errorf("APIKey = %v, expected %v", resolvedConfig.Gateway.APIKey, test.expected["api_key"])
			}

			if resolvedConfig.Gateway.Timeout != test.expected["timeout"] {
				t.Errorf("Timeout = %v, expected %v", resolvedConfig.Gateway.Timeout, test.expected["timeout"])
			}

			if resolvedConfig.OutputFormat != test.expected["output_format"] {
				t.Errorf("OutputFormat = %v, expected %v", resolvedConfig.OutputFormat, test.expected["output_format"])
			}

			if resolvedConfig.SortBy != test.expected["sort_by"] {
				t.Errorf("SortBy = %v, expected %v", resolvedConfig.SortBy, test.expected["sort_by"])
			}

			if resolvedConfig.Filter != test.expected["filter"] {
				t.Errorf("Filter = %v, expected %v", resolvedConfig.Filter, test.expected["filter"])
			}
		})
	}
}

func TestConfigSourceInfo(t *testing.T) {
	// テスト用の環境変数を設定
	os.Setenv("LLM_INFO_URL", "https://env.example.com")
	os.Setenv("LLM_INFO_OUTPUT_FORMAT", "json")
	defer func() {
		os.Unsetenv("LLM_INFO_URL")
		os.Unsetenv("LLM_INFO_OUTPUT_FORMAT")
	}()

	// 設定マネージャーの初期化
	configManager := config.NewManager("")

	// CLI引数
	cliArgs := &config.CLIArgs{
		URL:          "https://cli.example.com",
		OutputFormat: "table",
	}

	// 設定の解決
	resolvedConfig, err := configManager.ResolveConfig(cliArgs)
	if err != nil {
		t.Fatalf("ResolveConfig() error = %v", err)
	}

	// 設定ソース情報の取得
	sourceInfo := configManager.GetConfigSourceInfo(resolvedConfig)

	// ソース情報にCLIが含まれていることを確認
	if !contains(sourceInfo, "command line") {
		t.Error("Source info should contain 'command line'")
	}

	// ソース情報に環境変数が含まれていることを確認
	if !contains(sourceInfo, "environment variable") {
		t.Error("Source info should contain 'environment variable'")
	}

	// ソース情報にデフォルトが含まれていることを確認
	if !contains(sourceInfo, "default") {
		t.Error("Source info should contain 'default'")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cliArgs *config.CLIArgs
		wantErr bool
	}{
		{
			name: "valid config",
			cliArgs: &config.CLIArgs{
				URL:    "https://example.com",
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			cliArgs: &config.CLIArgs{
				APIKey: "test-key",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			cliArgs: &config.CLIArgs{
				URL:    "invalid-url",
				APIKey: "test-key",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 設定マネージャーの初期化
			configManager := config.NewManager("")

			// 設定の解決
			_, err := configManager.ResolveConfig(test.cliArgs)

			if (err != nil) != test.wantErr {
				t.Errorf("ResolveConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestEnvironmentVariableValidation(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		cliArgs *config.CLIArgs
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"LLM_INFO_URL":           "https://example.com",
				"LLM_INFO_API_KEY":       "test-key",
				"LLM_INFO_TIMEOUT":       "30s",
				"LLM_INFO_OUTPUT_FORMAT": "json",
			},
			wantErr: false,
			cliArgs: &config.CLIArgs{},
		},
		{
			name: "invalid timeout",
			envVars: map[string]string{
				"LLM_INFO_URL":     "https://example.com",
				"LLM_INFO_API_KEY": "test-key",
				"LLM_INFO_TIMEOUT": "invalid",
			},
			wantErr: true,
			cliArgs: &config.CLIArgs{},
		},
		{
			name: "invalid output format",
			envVars: map[string]string{
				"LLM_INFO_URL":           "https://example.com",
				"LLM_INFO_API_KEY":       "test-key",
				"LLM_INFO_OUTPUT_FORMAT": "invalid",
			},
			wantErr: true,
			cliArgs: &config.CLIArgs{},
		},
		{
			name: "CLI overrides invalid env",
			envVars: map[string]string{
				"LLM_INFO_TIMEOUT": "invalid",
			},
			wantErr: false,
			cliArgs: &config.CLIArgs{
				URL:     "https://example.com",
				APIKey:  "test-key",
				Timeout: 30 * time.Second,
			},
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

			// 設定マネージャーの初期化
			configManager := config.NewManager("")

			// 設定の解決
			_, err := configManager.ResolveConfig(test.cliArgs)

			if (err != nil) != test.wantErr {
				t.Errorf("ResolveConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// contains は文字列に部分文字列が含まれているかチェックするヘルパー関数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
