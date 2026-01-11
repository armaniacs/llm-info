package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	internalConfig "github.com/armaniacs/llm-info/internal/config"
	pkgconfig "github.com/armaniacs/llm-info/pkg/config"
)

func init() {
	// サブコマンド登録
	subcommands["probe"] = probeCommand
}

// probeCommand はprobeサブコマンドを実行する
func probeCommand(args []string) error {
	// probeコマンド用のフラグを定義
	probeCmd := flag.NewFlagSet("probe", flag.ExitOnError)
	model := probeCmd.String("model", "", "Target model ID (required)")
	baseURL := probeCmd.String("url", "", "Base URL of the LLM gateway")
	apiKey := probeCmd.String("api-key", "", "API key for authentication")
	gateway := probeCmd.String("gateway", "", "Gateway name to use from config")
	timeout := probeCmd.Duration("timeout", 30*time.Second, "Request timeout (default: 30s)")
	dryRun := probeCmd.Bool("dry-run", false, "Show execution plan without making actual API calls")
	verbose := probeCmd.Bool("verbose", false, "Show verbose logs")
	configFile := probeCmd.String("config", "", "Path to config file")
	showHelp := probeCmd.Bool("help", false, "Show help for probe command")

	// フラグを解析
	probeCmd.Parse(args)

	// ヘルプ表示
	if *showHelp {
		showProbeHelp()
		return nil
	}

	// 必須引数のチェック
	if *model == "" {
		fmt.Fprintf(os.Stderr, "Error: --model is required\n\n")
		showProbeHelp()
		os.Exit(1)
	}

	// 設定マネージャーの準備
	configPath := *configFile
	if configPath == "" {
		configPath = internalConfig.GetDefaultConfigPath()
	}
	configManager := internalConfig.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := configManager.Load(); err != nil {
		// 設定ファイルが存在しない場合は警告のみ表示
		if !strings.Contains(err.Error(), "no such file or directory") &&
		   !strings.Contains(err.Error(), "config file not found") {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config file: %v\n", err)
		}
	}

	// コマンドライン引数の構造体を作成
	cliArgs := &internalConfig.CLIArgs{
		URL:          *baseURL,
		APIKey:       *apiKey,
		Timeout:      *timeout,
		Gateway:      *gateway,
		OutputFormat: "json", // probeではjson固定
	}

	// 設定の解決
	resolved, err := configManager.ResolveConfig(cliArgs)
	if err != nil {
		return fmt.Errorf("failed to resolve config: %w", err)
	}

	// APIクライアントの作成
	cfg := pkgconfig.NewAppConfig()
	cfg.BaseURL = resolved.Gateway.URL
	cfg.APIKey = resolved.Gateway.APIKey
	cfg.Timeout = resolved.Gateway.Timeout
	client := api.NewProbeClient(cfg)

	// Dry-runモードの場合は実行計画を表示
	if *dryRun {
		showExecutionPlan(*model, resolved)
		return nil
	}

	// API呼び出しの実行
	if *verbose {
		fmt.Printf("Probing model %s...\n", *model)
	}

	response, err := client.ProbeModel(*model)
	if err != nil {
		return err
	}

	// 結果の表示
	displayProbeResult(*model, response, *verbose)

	return nil
}

// showProbeHelp はprobeコマンドのヘルプを表示する
func showProbeHelp() {
	fmt.Println(`llm-info probe - Probe model constraints via actual API behavior

USAGE:
    llm-info probe --model <MODEL_ID> [flags]

FLAGS:
    --model string      Target model ID (required)
    --url string         Base URL of the LLM gateway
    --api-key string     API key for authentication
    --gateway string     Gateway name to use from config
    --timeout duration   Request timeout (default: 30s)
    --dry-run           Show execution plan without making actual API calls
    --verbose           Show verbose logs
    --config string      Path to config file
    --help              Show help for probe command

EXAMPLES:
    # Basic usage
    llm-info probe --model gpt-4o-mini

    # With custom gateway
    llm-info probe --model gpt-4o-mini --gateway production

    # Dry run to see execution plan
    llm-info probe --model gpt-4o-mini --dry-run

    # Verbose output
    llm-info probe --model gpt-4o-mini --verbose
`)
}

// showExecutionPlan は実行計画を表示する
func showExecutionPlan(model string, config *internalConfig.ResolvedConfig) {
	fmt.Printf("Execution Plan:\n")
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  URL: %s\n", config.Gateway.URL)
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.Gateway.APIKey))
	fmt.Printf("  Timeout: %s\n", config.Gateway.Timeout)
	fmt.Printf("\nAPI Calls:\n")
	fmt.Printf("  1. POST %s/v1/chat/completions\n", config.Gateway.URL)
	fmt.Printf("     - Model: %s\n", model)
	fmt.Printf("     - Input: \"test\"\n")
	fmt.Printf("     - Max Output Tokens: 16\n")
	fmt.Printf("     - Temperature: 0\n")
	fmt.Printf("\nDry run complete. Use --dry-run=false to execute actual API calls.\n")
}

// displayProbeResult は探索結果を表示する
func displayProbeResult(model string, response *api.ProbeResponse, verbose bool) {
	fmt.Printf("Probe Results for %s:\n\n", model)

	fmt.Printf("API Response:\n")
	fmt.Printf("  ID: %s\n", response.ID)
	fmt.Printf("  Model: %s\n", response.Model)

	if response.Error != nil {
		fmt.Printf("  Error: %s\n", response.Error.Message)
	} else {
		if len(response.Choices) > 0 {
			fmt.Printf("  Finish Reason: %s\n", response.Choices[0].FinishReason)
			fmt.Printf("  Content: %s\n", response.Choices[0].Message.Content)
		}
		if response.Usage != nil {
			fmt.Printf("  Prompt Tokens: %d\n", response.Usage.PromptTokens)
			fmt.Printf("  Completion Tokens: %d\n", response.Usage.CompletionTokens)
			fmt.Printf("  Total Tokens: %d\n", response.Usage.TotalTokens)
		}
	}
}

// maskAPIKey はAPIキーをマスキングする
func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}