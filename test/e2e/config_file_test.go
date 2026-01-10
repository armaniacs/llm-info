package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigFileExecution(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "test-gateway"
    url: "https://httpbin.org/status/200"
    api_key: "test-key"
    timeout: "5s"
default_gateway: "test-gateway"
global:
  timeout: "10s"
  output_format: "json"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// 設定ファイルを使用してコマンドを実行
	cmd := exec.Command(binPath, "--config", configPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with config file: %v, output: %s", err, string(output))
	}

	// ヘルプが表示されることを確認
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output))
	}
}

func TestConfigFileWithInvalidGateway(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "test-gateway"
    url: "https://invalid-url-that-does-not-exist.com"
    api_key: "test-key"
    timeout: "1s"
default_gateway: "test-gateway"
global:
  timeout: "10s"
  output_format: "table"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// 設定ファイルを使用してコマンドを実行（エラーが発生することを確認）
	cmd := exec.Command(binPath, "--config", configPath)
	output, err := cmd.CombinedOutput()

	// エラーが発生することを確認
	if err == nil {
		t.Error("Expected error when using invalid gateway, but command succeeded")
	}

	// エラーメッセージが出力されることを確認
	outputStr := string(output)
	if !strings.Contains(outputStr, "Failed") && !strings.Contains(outputStr, "error") {
		t.Errorf("Expected error message in output, got: %s", outputStr)
	}
}

func TestConfigFileWithMultipleGateways(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "gateway1"
    url: "https://httpbin.org/status/200"
    api_key: "key1"
    timeout: "5s"
  - name: "gateway2"
    url: "https://httpbin.org/status/200"
    api_key: "key2"
    timeout: "10s"
default_gateway: "gateway1"
global:
  timeout: "10s"
  output_format: "json"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// デフォルトゲートウェイを使用してコマンドを実行
	cmd := exec.Command(binPath, "--config", configPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with default gateway: %v, output: %s", err, string(output))
	}

	// ヘルプが表示されることを確認
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output))
	}

	// gateway2を指定してコマンドを実行
	cmd2 := exec.Command(binPath, "--config", configPath, "--gateway", "gateway2", "--help")
	output2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with gateway2: %v, output: %s", err, string(output2))
	}

	// ヘルプが表示されることを確認
	if !strings.Contains(string(output2), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output2))
	}
}

func TestConfigFileWithCommandLineOverride(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "config-gateway"
    url: "https://httpbin.org/status/200"
    api_key: "config-key"
    timeout: "5s"
default_gateway: "config-gateway"
global:
  timeout: "10s"
  output_format: "json"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// コマンドライン引数でURLを上書きして実行
	cmd := exec.Command(binPath, "--config", configPath, "--url", "https://httpbin.org/status/200", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with CLI override: %v, output: %s", err, string(output))
	}

	// ヘルプが表示されることを確認
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output))
	}
}

func TestLegacyConfigFileCompatibility(t *testing.T) {
	// テスト用の古い形式設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-legacy-config.yaml")

	configContent := `
gateways:
  - name: "legacy-gateway"
    url: "https://httpbin.org/status/200"
    api_key: "legacy-key"
    timeout: "5s"
default_gateway: "legacy-gateway"
common:
  timeout: "10s"
  output:
    format: "json"
    table:
      always_show: ["name"]
      show_if_available: ["max_tokens"]
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test legacy config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// 古い形式の設定ファイルを使用してコマンドを実行
	cmd := exec.Command(binPath, "--config", configPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with legacy config file: %v, output: %s", err, string(output))
	}

	// ヘルプが表示されることを確認
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output))
	}
}

func TestConfigFileNotFound(t *testing.T) {
	// 存在しない設定ファイルパスを指定
	nonExistentPath := filepath.Join(t.TempDir(), "non-existent-config.yaml")

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// 存在しない設定ファイルを指定してコマンドを実行
	cmd := exec.Command(binPath, "--config", nonExistentPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run llm-info with non-existent config file: %v, output: %s", err, string(output))
	}

	// ヘルプが表示されることを確認（デフォルト設定が使用される）
	if !strings.Contains(string(output), "Usage:") {
		t.Errorf("Expected help output, got: %s", string(output))
	}
}

func TestConfigFileValidation(t *testing.T) {
	// 無効な設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.yaml")

	configContent := `
gateways:
  - name: ""
    url: "invalid-url"
    api_key: "test-key"
    timeout: "-1s"
default_gateway: "non-existent"
global:
  timeout: "0s"
  output_format: "invalid"
  sort_by: "invalid"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// llm-infoバイナリのパスを取得
	binPath := "../../llm-info"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Skip("llm-info binary not found, skipping E2E test")
	}

	// 無効な設定ファイルを使用してコマンドを実行
	cmd := exec.Command(binPath, "--config", configPath)
	output, err := cmd.CombinedOutput()

	// エラーが発生することを確認
	if err == nil {
		t.Error("Expected error when using invalid config, but command succeeded")
	}

	// エラーメッセージが出力されることを確認
	outputStr := string(output)
	if !strings.Contains(outputStr, "Failed") && !strings.Contains(outputStr, "error") {
		t.Errorf("Expected error message in output, got: %s", outputStr)
	}
}
