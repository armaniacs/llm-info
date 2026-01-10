package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/your-org/llm-info/pkg/config"
)

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// パスが期待される形式であることを確認
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	expected := filepath.Join(home, ".config", "llm-info", "llm-info.yaml")
	if path != expected {
		t.Errorf("GetConfigPath() = %s, want %s", path, expected)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "llm-info-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用設定ファイルのパス
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// ファイルが存在しない場合のテスト
	cfg, err := LoadConfigFromFile(configPath)
	if err != nil {
		t.Errorf("LoadConfigFromFile() with non-existent file returned error: %v", err)
	}

	if cfg == nil {
		t.Error("LoadConfigFromFile() returned nil config")
	}

	// デフォルト設定の確認
	if len(cfg.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in default config, got %d", len(cfg.Gateways))
	}

	if cfg.Gateways[0].Name != "default" {
		t.Errorf("Expected gateway name 'default', got '%s'", cfg.Gateways[0].Name)
	}

	if cfg.DefaultGateway != "default" {
		t.Errorf("Expected default gateway 'default', got '%s'", cfg.DefaultGateway)
	}

	if cfg.Global.OutputFormat != "table" {
		t.Errorf("Expected output format 'table', got '%s'", cfg.Global.OutputFormat)
	}
}

func TestSaveConfigToFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "llm-info-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用設定ファイルのパス
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// テスト用設定
	testConfig := &config.Config{
		Gateways: []config.Gateway{
			{
				Name:    "test-gateway",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: 5 * time.Second,
			},
		},
		DefaultGateway: "test-gateway",
		Global: config.Global{
			Timeout:      5 * time.Second,
			OutputFormat: "json",
			SortBy:       "max_tokens",
		},
	}

	// 設定を保存
	err = SaveConfigToFile(testConfig, configPath)
	if err != nil {
		t.Fatalf("SaveConfigToFile() failed: %v", err)
	}

	// ファイルが存在することを確認
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// 設定を読み込んで確認
	loadedConfig, err := LoadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFromFile() failed: %v", err)
	}

	if len(loadedConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway, got %d", len(loadedConfig.Gateways))
	}

	if loadedConfig.Gateways[0].Name != "test-gateway" {
		t.Errorf("Expected gateway name 'test-gateway', got '%s'", loadedConfig.Gateways[0].Name)
	}

	if loadedConfig.Gateways[0].URL != "https://test.example.com" {
		t.Errorf("Expected URL 'https://test.example.com', got '%s'", loadedConfig.Gateways[0].URL)
	}

	if loadedConfig.Gateways[0].APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", loadedConfig.Gateways[0].APIKey)
	}

	if loadedConfig.DefaultGateway != "test-gateway" {
		t.Errorf("Expected default gateway 'test-gateway', got '%s'", loadedConfig.DefaultGateway)
	}

	if loadedConfig.Global.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got '%s'", loadedConfig.Global.OutputFormat)
	}

	if loadedConfig.Global.SortBy != "max_tokens" {
		t.Errorf("Expected sort by 'max_tokens', got '%s'", loadedConfig.Global.SortBy)
	}
}

func TestLoadLegacyConfigFromFile(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "llm-info-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// テスト用設定ファイルのパス
	configPath := filepath.Join(tmpDir, "test-legacy-config.yaml")

	// ファイルが存在しない場合のテスト
	cfg, err := LoadLegacyConfigFromFile(configPath)
	if err != nil {
		t.Errorf("LoadLegacyConfigFromFile() with non-existent file returned error: %v", err)
	}

	if cfg == nil {
		t.Error("LoadLegacyConfigFromFile() returned nil config")
	}

	// デフォルト設定の確認
	if len(cfg.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in default legacy config, got %d", len(cfg.Gateways))
	}

	if cfg.Gateways[0].Name != "default" {
		t.Errorf("Expected gateway name 'default', got '%s'", cfg.Gateways[0].Name)
	}

	if cfg.DefaultGateway != "default" {
		t.Errorf("Expected default gateway 'default', got '%s'", cfg.DefaultGateway)
	}

	if cfg.Common.Output.Format != "table" {
		t.Errorf("Expected output format 'table', got '%s'", cfg.Common.Output.Format)
	}
}

func TestConvertLegacyToNew(t *testing.T) {
	// テスト用の古い形式設定
	legacyConfig := &config.FileConfig{
		Gateways: []config.GatewayConfig{
			{
				Name:    "test-gateway",
				URL:     "https://test.example.com",
				APIKey:  "test-key",
				Timeout: 5 * time.Second,
			},
		},
		DefaultGateway: "test-gateway",
		Common: config.CommonConfig{
			Timeout: 5 * time.Second,
			Output: config.OutputConfig{
				Format: "json",
				Table: config.TableConfig{
					AlwaysShow:      []string{"name"},
					ShowIfAvailable: []string{"max_tokens"},
				},
			},
		},
	}

	// 新しい形式に変換
	newConfig := ConvertLegacyToNew(legacyConfig)

	if newConfig == nil {
		t.Error("ConvertLegacyToNew() returned nil")
	}

	// 変換結果の確認
	if len(newConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway, got %d", len(newConfig.Gateways))
	}

	if newConfig.Gateways[0].Name != "test-gateway" {
		t.Errorf("Expected gateway name 'test-gateway', got '%s'", newConfig.Gateways[0].Name)
	}

	if newConfig.Gateways[0].URL != "https://test.example.com" {
		t.Errorf("Expected URL 'https://test.example.com', got '%s'", newConfig.Gateways[0].URL)
	}

	if newConfig.Gateways[0].APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", newConfig.Gateways[0].APIKey)
	}

	if newConfig.DefaultGateway != "test-gateway" {
		t.Errorf("Expected default gateway 'test-gateway', got '%s'", newConfig.DefaultGateway)
	}

	if newConfig.Global.OutputFormat != "json" {
		t.Errorf("Expected output format 'json', got '%s'", newConfig.Global.OutputFormat)
	}

	if newConfig.Global.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", newConfig.Global.Timeout)
	}
}

func TestConvertLegacyToNewWithNil(t *testing.T) {
	// nilを渡した場合のテスト
	newConfig := ConvertLegacyToNew(nil)

	if newConfig == nil {
		t.Error("ConvertLegacyToNew() with nil returned nil")
	}

	// デフォルト設定が返されることを確認
	if len(newConfig.Gateways) != 1 {
		t.Errorf("Expected 1 gateway in default config, got %d", len(newConfig.Gateways))
	}

	if newConfig.Gateways[0].Name != "default" {
		t.Errorf("Expected gateway name 'default', got '%s'", newConfig.Gateways[0].Name)
	}
}
