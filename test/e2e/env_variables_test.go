package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestEnvironmentVariables(t *testing.T) {
	// ビルド済みバイナリのパス
	binaryPath := "../../llm-info"

	// テスト用の環境変数を設定
	testEnvVars := map[string]string{
		"LLM_INFO_URL":           "https://api.example.com",
		"LLM_INFO_API_KEY":       "test-api-key",
		"LLM_INFO_TIMEOUT":       "15s",
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
		name        string
		envVars     map[string]string
		args        []string
		expectError bool
		expectOut   string
	}{
		{
			name: "basic environment variables",
			args: []string{"--help"},
			envVars: map[string]string{
				"LLM_INFO_URL":     "https://api.example.com",
				"LLM_INFO_API_KEY": "test-api-key",
			},
			expectError: false,
			expectOut:   "Usage:",
		},
		{
			name: "show sources with environment variables",
			args: []string{"--show-sources"},
			envVars: map[string]string{
				"LLM_INFO_URL":           "https://api.example.com",
				"LLM_INFO_API_KEY":       "test-api-key",
				"LLM_INFO_OUTPUT_FORMAT": "json",
			},
			expectError: false,
			expectOut:   "environment variable",
		},
		{
			name: "CLI overrides environment variables",
			args: []string{
				"--url", "https://cli.example.com",
				"--api-key", "cli-api-key",
				"--show-sources",
			},
			envVars: map[string]string{
				"LLM_INFO_URL":     "https://env.example.com",
				"LLM_INFO_API_KEY": "env-api-key",
			},
			expectError: false,
			expectOut:   "command line",
		},
		{
			name: "invalid environment variable",
			args: []string{"--help"},
			envVars: map[string]string{
				"LLM_INFO_TIMEOUT": "invalid",
			},
			expectError: false, // helpは常に成功するはず
			expectOut:   "Usage:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// テスト固有の環境変数を設定
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range test.envVars {
					if originalValue, exists := originalVars[key]; exists {
						if originalValue == "" {
							os.Unsetenv(key)
						} else {
							os.Setenv(key, originalValue)
						}
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			// コマンド実行
			cmd := exec.Command(binaryPath, test.args...)
			cmd.Env = os.Environ()

			output, err := cmd.CombinedOutput()

			if test.expectError && err == nil {
				t.Errorf("Expected error but command succeeded")
			}

			if !test.expectError && err != nil {
				t.Errorf("Command failed: %v, output: %s", err, string(output))
			}

			if test.expectOut != "" && !strings.Contains(string(output), test.expectOut) {
				t.Errorf("Expected output to contain '%s', got: %s", test.expectOut, string(output))
			}
		})
	}
}

func TestEnvironmentVariablePriority(t *testing.T) {
	// ビルド済みバイナリのパス
	binaryPath := "../../llm-info"

	// テスト用の環境変数を設定
	envVars := map[string]string{
		"LLM_INFO_URL":           "https://env.example.com",
		"LLM_INFO_API_KEY":       "env-api-key",
		"LLM_INFO_OUTPUT_FORMAT": "json",
	}

	// テスト前に環境変数を保存
	originalVars := make(map[string]string)
	for key, value := range envVars {
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

	// 環境変数のみの場合
	cmd := exec.Command(binaryPath, "--show-sources")
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, string(output))
	}

	// 環境変数が設定ソースに含まれていることを確認
	if !strings.Contains(string(output), "environment variable") {
		t.Errorf("Expected output to contain 'environment variable', got: %s", string(output))
	}

	// CLI引数が環境変数を上書きすることを確認
	cmd = exec.Command(binaryPath,
		"--url", "https://cli.example.com",
		"--show-sources")
	cmd.Env = os.Environ()

	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v, output: %s", err, string(output))
	}

	// CLIが設定ソースに含まれていることを確認
	if !strings.Contains(string(output), "command line") {
		t.Errorf("Expected output to contain 'command line', got: %s", string(output))
	}
}

func TestEnvironmentVariableValidation(t *testing.T) {
	// ビルド済みバイナリのパス
	binaryPath := "../../llm-info"

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "valid timeout",
			envVars: map[string]string{
				"LLM_INFO_TIMEOUT": "30s",
			},
			expectError: false,
		},
		{
			name: "invalid timeout",
			envVars: map[string]string{
				"LLM_INFO_TIMEOUT": "invalid",
			},
			expectError: true,
		},
		{
			name: "valid output format",
			envVars: map[string]string{
				"LLM_INFO_OUTPUT_FORMAT": "json",
			},
			expectError: false,
		},
		{
			name: "invalid output format",
			envVars: map[string]string{
				"LLM_INFO_OUTPUT_FORMAT": "invalid",
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// テスト用の環境変数を設定
			for key, value := range test.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range test.envVars {
					os.Unsetenv(key)
				}
			}()

			// コマンド実行（実際のAPI呼び出しはしないように、存在しないURLを使用）
			cmd := exec.Command(binaryPath,
				"--url", "https://nonexistent.example.com",
				"--timeout", "1s") // すぐにタイムアウトするように設定
			cmd.Env = os.Environ()

			// タイムアウトを設定
			timer := time.AfterFunc(5*time.Second, func() {
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			})
			defer timer.Stop()

			output, err := cmd.CombinedOutput()

			if test.expectError {
				// エラーが期待される場合、環境変数のバリデーションエラーかタイムアウトエラーのいずれかを受け入れる
				if err == nil {
					t.Errorf("Expected error but command succeeded")
				}
				// 環境変数のバリデーションエラーが出力に含まれているか確認
				if strings.Contains(string(output), "invalid") ||
					strings.Contains(string(output), "timeout") {
					// 期待されるエラー
				} else {
					t.Errorf("Expected validation error, got: %v, output: %s", err, string(output))
				}
			} else {
				// エラーが期待されない場合、タイムアウトエラーのみを受け入れる
				if err != nil && !strings.Contains(string(output), "timeout") {
					t.Errorf("Unexpected error: %v, output: %s", err, string(output))
				}
			}
		})
	}
}

func TestEnvironmentVariableHelp(t *testing.T) {
	// ビルド済みバイナリのパス
	binaryPath := "../../llm-info"

	// ヘルプコマンドを実行
	cmd := exec.Command(binaryPath, "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	// 環境変数のヘルプが含まれていることを確認
	expectedEnvVars := []string{
		"LLM_INFO_URL",
		"LLM_INFO_API_KEY",
		"LLM_INFO_TIMEOUT",
		"LLM_INFO_DEFAULT_GATEWAY",
		"LLM_INFO_OUTPUT_FORMAT",
		"LLM_INFO_SORT_BY",
		"LLM_INFO_FILTER",
		"LLM_INFO_CONFIG_PATH",
		"LLM_INFO_LOG_LEVEL",
		"LLM_INFO_USER_AGENT",
	}

	for _, envVar := range expectedEnvVars {
		if !strings.Contains(string(output), envVar) {
			t.Errorf("Expected help to contain '%s', got: %s", envVar, string(output))
		}
	}

	// 環境変数の例が含まれていることを確認
	expectedExamples := []string{
		"export LLM_INFO_URL=",
		"export LLM_INFO_API_KEY=",
		"export LLM_INFO_OUTPUT_FORMAT=",
		"export LLM_INFO_SORT_BY=",
		"export LLM_INFO_FILTER=",
	}

	for _, example := range expectedExamples {
		if !strings.Contains(string(output), example) {
			t.Errorf("Expected help to contain example '%s', got: %s", example, string(output))
		}
	}
}
