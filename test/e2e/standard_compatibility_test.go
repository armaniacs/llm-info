package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestStandardCompatibilityE2E(t *testing.T) {
	// このテストは実際のOpenAI互換APIエンドポイントが必要なため、
	// 環境変数が設定されている場合のみ実行
	apiURL := os.Getenv("LLM_INFO_TEST_API_URL")
	if apiURL == "" {
		t.Skip("Skipping E2E test: LLM_INFO_TEST_API_URL not set")
	}

	apiKey := os.Getenv("LLM_INFO_TEST_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping E2E test: LLM_INFO_TEST_API_KEY not set")
	}

	tests := []struct {
		name           string
		args           []string
		expectSuccess  bool
		expectedOutput []string
	}{
		{
			name:          "Basic model listing",
			args:          []string{"--url", apiURL, "--api-key", apiKey},
			expectSuccess: true,
			expectedOutput: []string{
				"Fetching model information",
				"Found",
				"models",
			},
		},
		{
			name:          "JSON output format",
			args:          []string{"--url", apiURL, "--api-key", apiKey, "--format", "json"},
			expectSuccess: true,
			expectedOutput: []string{
				"[",
				"]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// llm-infoコマンドの実行
			args := append([]string{"run", "cmd/llm-info/main.go"}, tt.args...)
			cmd := exec.Command("go", args...)
			output, err := cmd.CombinedOutput()

			// 成功/失敗のチェック
			if (err == nil) != tt.expectSuccess {
				t.Errorf("Expected success: %v, got error: %v", tt.expectSuccess, err)
				t.Logf("Command output: %s", string(output))
				return
			}

			// 出力内容のチェック
			outputStr := string(output)
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, outputStr)
				}
			}

			t.Logf("Command output: %s", outputStr)
		})
	}
}

func TestFallbackBehaviorE2E(t *testing.T) {
	// フォールバック動作をテストするためのモックサーバーを起動
	// このテストは実際のサーバーを必要としない

	// LiteLLMエンドポイントが存在しない場合のテスト
	t.Run("LiteLLM endpoint not found, fallback to standard", func(t *testing.T) {
		// このテストはモックサーバーを使用するため、実際のAPIキーは不要
		// ただし、テスト用のモックサーバーを別途実装する必要がある
		t.Skip("Skipping fallback test: requires mock server implementation")
	})
}

func TestRealWorldGateways(t *testing.T) {
	// 実際のゲートウェイでのテスト（オプション）
	gatewayURLs := []string{
		os.Getenv("OPENAI_API_URL"),
		os.Getenv("AZURE_OPENAI_API_URL"),
		os.Getenv("LOCAL_LLM_API_URL"),
	}

	for i, url := range gatewayURLs {
		if url == "" {
			continue
		}

		t.Run(fmt.Sprintf("Gateway %d: %s", i+1, url), func(t *testing.T) {
			apiKey := os.Getenv(fmt.Sprintf("GATEWAY_%d_API_KEY", i+1))
			if apiKey == "" {
				t.Skip("API key not provided for gateway")
			}

			args := []string{"run", "cmd/llm-info/main.go",
				"--url", url,
				"--api-key", apiKey,
				"--timeout", "30s"}
			cmd := exec.Command("go", args...)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Failed to get models from gateway: %v", err)
				t.Logf("Output: %s", string(output))
				return
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Found") {
				t.Errorf("Expected output to contain 'Found', got: %s", outputStr)
			}

			t.Logf("Gateway %d output: %s", i+1, outputStr)
		})
	}
}
