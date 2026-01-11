package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	internalConfig "github.com/armaniacs/llm-info/internal/config"
	"github.com/armaniacs/llm-info/internal/probe"
	"github.com/armaniacs/llm-info/internal/ui"
	"github.com/armaniacs/llm-info/pkg/config"
)

func init() {
	// サブコマンド登録
	subcommands["probe"] = probeCommand
	subcommands["probe-context"] = probeContextCommand
	subcommands["probe-max-output"] = probeMaxOutputCommand
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

	// Dry-runモードの場合は実行計画を表示
	if *dryRun {
		showExecutionPlan(*model, resolved)
		return nil
	}

	// API呼び出の実行
	if *verbose {
		fmt.Printf("Probing model %s...\n", *model)
	}

	// TODO: ここにContext Window探索の実装を追加
	fmt.Printf("Context window probing not yet implemented. Use --dry-run to see execution plan.\n")

	return nil
}

// probeContextCommand はcontext window探索を実行する
func probeContextCommand(args []string) error {
	// probe-contextコマンド用のフラグを定義
	probeCmd := flag.NewFlagSet("probe-context", flag.ExitOnError)
	model := probeCmd.String("model", "", "Target model ID (required)")
	baseURL := probeCmd.String("url", "", "Base URL of the LLM gateway")
	apiKey := probeCmd.String("api-key", "", "API key for authentication")
	gateway := probeCmd.String("gateway", "", "Gateway name to use from config")
	timeout := probeCmd.Duration("timeout", 30*time.Second, "Request timeout (default: 30s)")
	dryRun := probeCmd.Bool("dry-run", false, "Show execution plan without making actual API calls")
	verbose := probeCmd.Bool("verbose", false, "Show verbose logs")
	configFile := probeCmd.String("config", "", "Path to config file")
	showHelp := probeCmd.Bool("help", false, "Show help for probe-context command")

	// フラグを解析
	probeCmd.Parse(args)

	// ヘルプ表示
	if *showHelp {
		showProbeContextHelp()
		return nil
	}

	// 必須引数のチェック
	if *model == "" {
		fmt.Fprintf(os.Stderr, "Error: --model is required\n\n")
		showProbeContextHelp()
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

	// Dry-runモードの場合は実行計画を表示
	if *dryRun {
		showContextExecutionPlan(*model, resolved)
		return nil
	}

	// APIクライアントを作成
	cfg := &config.AppConfig{
		BaseURL: resolved.Gateway.URL,
		APIKey:  resolved.Gateway.APIKey,
		Timeout: resolved.Gateway.Timeout,
	}

	client := api.NewProbeClient(cfg)

	// Context Window Proberを作成
	prober := probe.NewContextWindowProbe(client)

	// 実行
	if *verbose {
		fmt.Printf("Probing context window for model %s...\n", *model)
	}

	result, err := prober.Probe(*model, *verbose)
	if err != nil {
		return fmt.Errorf("failed to probe context window: %w", err)
	}

	// テーブル形式で出力
	formatter := ui.NewTableFormatter()
	output := formatter.FormatContextWindowResult(result)
	fmt.Println(output)

	// verbose時は履歴も表示
	if *verbose && len(result.TrialHistory) > 0 {
		history := formatter.FormatVerboseHistory(result.TrialHistory)
		fmt.Println(history)
	}

	return nil
}

// probeMaxOutputCommand はmax output tokens探索を実行する
func probeMaxOutputCommand(args []string) error {
	// probe-max-outputコマンド用のフラグを定義
	probeCmd := flag.NewFlagSet("probe-max-output", flag.ExitOnError)
	model := probeCmd.String("model", "", "Target model ID (required)")
	baseURL := probeCmd.String("url", "", "Base URL of the LLM gateway")
	apiKey := probeCmd.String("api-key", "", "API key for authentication")
	gateway := probeCmd.String("gateway", "", "Gateway name to use from config")
	timeout := probeCmd.Duration("timeout", 30*time.Second, "Request timeout (default: 30s)")
	dryRun := probeCmd.Bool("dry-run", false, "Show execution plan without making actual API calls")
	verbose := probeCmd.Bool("verbose", false, "Show verbose logs")
	configFile := probeCmd.String("config", "", "Path to config file")
	showHelp := probeCmd.Bool("help", false, "Show help for probe-max-output command")

	// フラグを解析
	probeCmd.Parse(args)

	// ヘルプ表示
	if *showHelp {
		showProbeMaxOutputHelp()
		return nil
	}

	// 必須引数のチェック
	if *model == "" {
		fmt.Fprintf(os.Stderr, "Error: --model is required\n\n")
		showProbeMaxOutputHelp()
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

	// Dry-runモードの場合は実行計画を表示
	if *dryRun {
		showMaxOutputExecutionPlan(*model, resolved)
		return nil
	}

	// APIクライアントを作成
	cfg := &config.AppConfig{
		BaseURL: resolved.Gateway.URL,
		APIKey:  resolved.Gateway.APIKey,
		Timeout: resolved.Gateway.Timeout,
	}

	client := api.NewProbeClient(cfg)

	// Max Output Tokens Proberを作成
	prober := probe.NewMaxOutputTokensProbe(client)

	// 実行
	if *verbose {
		fmt.Printf("Probing max output tokens for model %s...\n", *model)
	}

	result, err := prober.ProbeOutputTokens(*model, *verbose)
	if err != nil {
		return fmt.Errorf("failed to probe max output tokens: %w", err)
	}

	// テーブル形式で出力
	formatter := ui.NewTableFormatter()
	output := formatter.FormatMaxOutputResult(result)
	fmt.Println(output)

	// verbose時は履歴も表示
	if *verbose && len(result.TrialHistory) > 0 {
		history := formatter.FormatVerboseHistory(result.TrialHistory)
		fmt.Println(history)
	}

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

    # Context window probing (not yet implemented)
    llm-info probe-context --model gpt-4o-mini
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

// showProbeContextHelp はprobe-contextコマンドのヘルプを表示する
func showProbeContextHelp() {
	fmt.Println(`llm-info probe-context - Probe context window constraints via actual API behavior

USAGE:
    llm-info probe-context --model <MODEL_ID> [flags]

FLAGS:
    --model string      Target model ID (required)
    --url string         Base URL of the LLM gateway
    --api-key string     API key for authentication
    --gateway string     Gateway name to use from config
    --timeout duration   Request timeout (default: 30s)
    --dry-run           Show execution plan without making actual API calls
    --verbose           Show verbose logs
    --config string      Path to config file
    --help              Show help for probe-context command

EXAMPLES:
    # Basic usage
    llm-info probe-context --model gpt-4o-mini

    # With custom gateway
    llm-info probe-context --model gpt-4o-mini --gateway production

    # Dry run to see execution plan
    llm-info probe-context --model gpt-4o-mini --dry-run

    # Verbose output
    llm-info probe-context --model gpt-4o-mini --verbose
`)
}

// showContextExecutionPlan はContext Window探索の実行計画を表示する
func showContextExecutionPlan(model string, config *internalConfig.ResolvedConfig) {
	fmt.Printf("Context Window Probe Execution Plan:\n")
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  URL: %s\n", config.Gateway.URL)
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.Gateway.APIKey))
	fmt.Printf("  Timeout: %s\n", config.Gateway.Timeout)
	fmt.Printf("\nProbe Phases:\n")
	fmt.Printf("  1. Exponential Search: Find upper bound by doubling token count\n")
	fmt.Printf("  2. Binary Search: Refine boundary within ±1024 tokens\n")
	fmt.Printf("  3. Error Analysis: Extract token limits from error messages\n")
	fmt.Printf("\nAPI Calls:\n")
	fmt.Printf("  POST %s/v1/chat/completions\n", config.Gateway.URL)
	fmt.Printf("  - Test data generation with Japanese text\n")
	fmt.Printf("  - Needle-in-haystack methodology\n")
	fmt.Printf("  - Rate limited: 1 second between calls\n")
	fmt.Printf("\nDry run complete. Use --dry-run=false to execute actual API calls.\n")
}

// showProbeMaxOutputHelp はprobe-max-outputコマンドのヘルプを表示する
func showProbeMaxOutputHelp() {
	fmt.Println(`llm-info probe-max-output - Probe max output tokens constraints via actual API behavior

USAGE:
    llm-info probe-max-output --model <MODEL_ID> [flags]

FLAGS:
    --model string      Target model ID (required)
    --url string         Base URL of the LLM gateway
    --api-key string     API key for authentication
    --gateway string     Gateway name to use from config
    --timeout duration   Request timeout (default: 30s)
    --dry-run           Show execution plan without making actual API calls
    --verbose           Show verbose logs
    --config string      Path to config file
    --help              Show help for probe-max-output command

EXAMPLES:
    # Basic usage
    llm-info probe-max-output --model gpt-4o-mini

    # With custom gateway
    llm-info probe-max-output --model gpt-4o-mini --gateway production

    # Dry run to see execution plan
    llm-info probe-max-output --model gpt-4o-mini --dry-run

    # Verbose output
    llm-info probe-max-output --model gpt-4o-mini --verbose

DESCRIPTION:
    This command determines the maximum number of tokens a model can generate
    in a single response through binary search and exponential search.
    It detects limits via validation errors and incomplete response status.
`)
}

// showMaxOutputExecutionPlan はMax Output Tokens探索の実行計画を表示する
func showMaxOutputExecutionPlan(model string, config *internalConfig.ResolvedConfig) {
	fmt.Printf("Max Output Tokens Probe Execution Plan:\n")
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  URL: %s\n", config.Gateway.URL)
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.Gateway.APIKey))
	fmt.Printf("  Timeout: %s\n", config.Gateway.Timeout)
	fmt.Printf("\nProbe Phases:\n")
	fmt.Printf("  1. Exponential Search: Find upper bound by doubling token count (256→512→1024...)\n")
	fmt.Printf("  2. Binary Search: Refine boundary within detection range\n")
	fmt.Printf("  3. Evidence Analysis: Determine limit source (error/incomplete)\n")
	fmt.Printf("\nAPI Calls:\n")
	fmt.Printf("  POST %s/v1/chat/completions\n", config.Gateway.URL)
	fmt.Printf("  - Fixed input length (1000 tokens)\n")
	fmt.Printf("  - Varying max_tokens parameter\n")
	fmt.Printf("  - Check for validation errors and incomplete status\n")
	fmt.Printf("  - Rate limited: 0.5s between calls\n")
	fmt.Printf("\nDetection Methods:\n")
	fmt.Printf("  - Validation Error: Extract from error messages (e.g., 'max_output_tokens must be <= 16384')\n")
	fmt.Printf("  - Incomplete Status: Check response.finish_reason='length'\n")
	fmt.Printf("\nDry run complete. Use --dry-run=false to execute actual API calls.\n")
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