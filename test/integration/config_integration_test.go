package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

func TestConfigFileAndCommandLineIntegration(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "test-gateway"
    url: "https://test.example.com"
    api_key: "test-key"
    timeout: "5s"
  - name: "another-gateway"
    url: "https://another.example.com"
    api_key: "another-key"
    timeout: "10s"
default_gateway: "test-gateway"
global:
  timeout: "10s"
  output_format: "json"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// 設定マネージャーの初期化
	manager := config.NewManager(configPath)

	// 環境変数を設定
	testVars := map[string]string{
		"LLM_INFO_URL":           "https://env.example.com",
		"LLM_INFO_API_KEY":       "env-api-key",
		"LLM_INFO_TIMEOUT":       "15s",
		"LLM_INFO_GATEWAY":       "another-gateway",
		"LLM_INFO_OUTPUT_FORMAT": "table",
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

	// 環境変数の読み込み
	manager.LoadFromEnv()

	// 設定ファイルの読み込み
	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	// 環境変数でゲートウェイが指定されている場合、そのゲートウェイを適用
	if err := manager.ApplyGateway("another-gateway"); err != nil {
		t.Fatalf("Failed to apply gateway: %v", err)
	}

	// 設定の取得と検証
	appConfig := manager.GetConfig()
	if err := manager.ValidateConfig(); err != nil {
		t.Fatalf("Invalid configuration: %v", err)
	}

	// 環境変数が設定ファイルより優先されることを確認
	if appConfig.BaseURL != "https://env.example.com" {
		t.Errorf("Expected BaseURL from environment variable, got %q", appConfig.BaseURL)
	}

	if appConfig.APIKey != "env-api-key" {
		t.Errorf("Expected APIKey from environment variable, got %q", appConfig.APIKey)
	}

	if appConfig.Timeout != 15*time.Second {
		t.Errorf("Expected Timeout from environment variable, got %v", appConfig.Timeout)
	}

	// 設定ファイルから出力形式が読み込まれる（環境変数で上書きされていない）
	if appConfig.OutputFormat != "json" {
		t.Errorf("Expected OutputFormat from config file, got %q", appConfig.OutputFormat)
	}
}

func TestConfigFileFallback(t *testing.T) {
	// 存在しない設定ファイルパスを指定
	nonExistentPath := filepath.Join(t.TempDir(), "non-existent-config.yaml")

	// 設定マネージャーの初期化
	manager := config.NewManager(nonExistentPath)

	// 設定ファイルの読み込み（エラーにならないことを確認）
	if err := manager.Load(); err != nil {
		t.Errorf("Load() with non-existent file should not return error, got: %v", err)
	}

	// デフォルト設定が使用されることを確認
	newConfig := manager.GetNewConfig()
	if newConfig == nil {
		t.Error("GetNewConfig() returned nil")
		return
	}

	if len(newConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in default config, got %d", len(newConfig.Gateways))
	}

	if newConfig.Gateways[0].Name != "default" {
		t.Errorf("Expected gateway name 'default', got %q", newConfig.Gateways[0].Name)
	}

	if newConfig.DefaultGateway != "default" {
		t.Errorf("Expected default gateway 'default', got %q", newConfig.DefaultGateway)
	}
}

func TestLegacyConfigCompatibility(t *testing.T) {
	// テスト用の古い形式設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-legacy-config.yaml")

	configContent := `gateways:
	 - name: "legacy-gateway"
	   url: "https://legacy.example.com/v1"
	   api_key: "legacy-key"
	   timeout: "5s"
default_gateway: "legacy-gateway"
common:
	 timeout: "10s"
	 output:
	   format: "json"
	   table:
	     always_show: ["name"]
	     show_if_available: ["max_tokens"]`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test legacy config file: %v", err)
	}

	// 設定マネージャーの初期化
	manager := config.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load legacy config file: %v", err)
	}

	// 古い形式の設定が読み込まれていることを確認
	fileConfig := manager.GetFileConfig()
	if fileConfig == nil {
		t.Error("GetFileConfig() returned nil")
		return
	}

	if len(fileConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in legacy config, got %d", len(fileConfig.Gateways))
	}

	if fileConfig.Gateways[0].Name != "legacy-gateway" {
		t.Errorf("Expected gateway name 'legacy-gateway', got %q", fileConfig.Gateways[0].Name)
	}

	// 新しい形式の設定に変換されていることを確認
	newConfig := manager.GetNewConfig()
	if newConfig == nil {
		t.Error("GetNewConfig() returned nil")
		return
	}

	if len(newConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in converted config, got %d", len(newConfig.Gateways))
	}

	if newConfig.Gateways[0].Name != "legacy-gateway" {
		t.Errorf("Expected gateway name 'legacy-gateway', got %q", newConfig.Gateways[0].Name)
	}

	if newConfig.Global.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got %q", newConfig.Global.OutputFormat)
	}

	// ゲートウェイ設定が適用できることを確認
	if err := manager.ApplyGateway("legacy-gateway"); err != nil {
		t.Errorf("Failed to apply legacy gateway: %v", err)
	}

	appConfig := manager.GetConfig()
	if appConfig.BaseURL != "https://legacy.example.com/v1" {
		t.Errorf("Expected BaseURL 'https://legacy.example.com/v1', got %q", appConfig.BaseURL)
	}
}

func TestGatewayOverride(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "default-gateway"
    url: "https://default.example.com"
    api_key: "default-key"
    timeout: "5s"
  - name: "override-gateway"
    url: "https://override.example.com"
    api_key: "override-key"
    timeout: "10s"
default_gateway: "default-gateway"
global:
  timeout: "10s"
  output_format: "table"
  sort_by: "name"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// 設定マネージャーの初期化
	manager := config.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	// デフォルトゲートウェイを適用
	if err := manager.ApplyGateway(""); err != nil {
		t.Fatalf("Failed to apply default gateway: %v", err)
	}

	appConfig := manager.GetConfig()
	if appConfig.BaseURL != "https://default.example.com" {
		t.Errorf("Expected default gateway URL, got %q", appConfig.BaseURL)
	}

	// コマンドライン引数でURLを上書き
	manager.SetBaseURL("https://cli-override.example.com")

	// ゲートウェイ設定を再適用（上書きされないことを確認）
	if err := manager.ApplyGateway("override-gateway"); err != nil {
		t.Fatalf("Failed to apply override gateway: %v", err)
	}

	// コマンドライン引数が優先されることを確認
	if appConfig.BaseURL != "https://cli-override.example.com" {
		t.Errorf("Expected CLI override URL, got %q", appConfig.BaseURL)
	}

	// APIキーは上書きされていないため、ゲートウェイ設定が使用される
	if appConfig.APIKey != "default-key" {
		t.Errorf("Expected default gateway API key, got %q", appConfig.APIKey)
	}
}

func TestMultipleConfigSources(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
gateways:
  - name: "config-gateway"
    url: "https://config.example.com"
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

	// 設定マネージャーの初期化
	manager := config.NewManager(configPath)

	// 環境変数を設定
	os.Setenv("LLM_INFO_API_KEY", "env-key")
	defer os.Unsetenv("LLM_INFO_API_KEY")

	// 環境変数の読み込み
	manager.LoadFromEnv()

	// 設定ファイルの読み込み
	if err := manager.Load(); err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	// ゲートウェイ設定を適用
	if err := manager.ApplyGateway(""); err != nil {
		t.Fatalf("Failed to apply gateway: %v", err)
	}

	// 設定の取得と検証
	appConfig := manager.GetConfig()

	// 設定ファイルからURLが読み込まれる
	if appConfig.BaseURL != "https://config.example.com" {
		t.Errorf("Expected URL from config file, got %q", appConfig.BaseURL)
	}

	// 環境変数からAPIキーが読み込まれる
	if appConfig.APIKey != "env-key" {
		t.Errorf("Expected API key from environment variable, got %q", appConfig.APIKey)
	}

	// 設定ファイルから出力形式が読み込まれる
	if appConfig.OutputFormat != "json" {
		t.Errorf("Expected output format from config file, got %q", appConfig.OutputFormat)
	}

	// コマンドライン引数で出力形式を上書き
	manager.SetOutputFormat("table")

	// コマンドライン引数が優先される
	if appConfig.OutputFormat != "table" {
		t.Errorf("Expected output format from CLI argument, got %q", appConfig.OutputFormat)
	}
}
