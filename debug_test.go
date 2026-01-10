package main

import (
	"fmt"
	"os"
	"github.com/armaniacs/llm-info/internal/config"
)

func main() {
	// テスト用の環境変数を設定
	os.Setenv("LLM_INFO_URL", "https://example.com")
	os.Setenv("LLM_INFO_API_KEY", "test-key")
	os.Setenv("LLM_INFO_TIMEOUT", "invalid")
	defer func() {
		os.Unsetenv("LLM_INFO_URL")
		os.Unsetenv("LLM_INFO_API_KEY")
		os.Unsetenv("LLM_INFO_TIMEOUT")
	}()

	// 設定マネージャーの初期化
	configManager := config.NewManager("")

	// 設定の解決
	_, err := configManager.ResolveConfig(&config.CLIArgs{})
	
	fmt.Printf("Error: %v\n", err)
}
