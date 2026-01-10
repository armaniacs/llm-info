package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	// DefaultBinaryPath はデフォルトのバイナリパス
	DefaultBinaryPath = "test/bin/llm-info"
	// TestConfigPath はテスト用設定パス
	TestConfigPath = "test/configs/test.yaml"
)

// SetupTestEnvironment はE2Eテスト環境をセットアップする
func SetupTestEnvironment(t *testing.T) string {
	t.Helper()

	// バイナリの存在確認
	binaryPath := os.Getenv("LLM_INFO_BIN_PATH")
	if binaryPath == "" {
		binaryPath = DefaultBinaryPath
	}

	// バイナリが存在しない場合はビルド
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Logf("Binary not found at %s, building...", binaryPath)

		// プロジェクトルートでビルド
		pwd, _ := os.Getwd()
		projectRoot := filepath.Dir(filepath.Dir(pwd))
		buildCmd := exec.Command("go", "build", "-o", binaryPath, "cmd/llm-info/*.go")
		buildCmd.Dir = projectRoot
		if err := buildCmd.Run(); err != nil {
			t.Fatalf("Failed to build binary: %v", err)
		}
	}

	return binaryPath
}

// CleanTestEnvironment はE2Eテスト環境をクリーンアップする
func CleanTestEnvironment(t *testing.T) {
	t.Helper()

	// テンポラリな設定のクリーンアップ
	os.RemoveAll("test/tmp")
}

// WithTestConfig はテスト用設定を環境変数に設定する関数を返す
func WithTestConfig() func() {
	origConfig := os.Getenv("LLM_INFO_CONFIG_PATH")
	os.Setenv("LLM_INFO_CONFIG_PATH", TestConfigPath)

	return func() {
		if origConfig != "" {
			os.Setenv("LLM_INFO_CONFIG_PATH", origConfig)
		} else {
			os.Unsetenv("LLM_INFO_CONFIG_PATH")
		}
	}
}

// CreateTempDir は一時的なテストディレクトリを作成する
func CreateTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("test/tmp", "e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}