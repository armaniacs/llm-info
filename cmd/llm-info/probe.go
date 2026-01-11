package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	internalConfig "github.com/armaniacs/llm-info/internal/config"
	"github.com/armaniacs/llm-info/internal/logging"
	"github.com/armaniacs/llm-info/internal/probe"
	"github.com/armaniacs/llm-info/internal/storage"
	"github.com/armaniacs/llm-info/internal/ui"
	"github.com/armaniacs/llm-info/pkg/config"
)

func init() {
	// サブコマンド登録
	subcommands["probe"] = probeCommand
	subcommands["probe-context"] = probeContextCommand
	subcommands["probe-max-output"] = probeMaxOutputCommand
}

// probeCommand はprobeサブコマンドを実行する（統合版）
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
	logDir := probeCmd.String("log-dir", "", "Directory to save probe logs")
	saveResult := probeCmd.Bool("save-result", false, "Save probe results to file")
	noLog := probeCmd.Bool("no-log", false, "Disable logging")
	contextOnly := probeCmd.Bool("context-only", false, "Probe only context window")
	outputOnly := probeCmd.Bool("output-only", false, "Probe only max output tokens")
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

	// オプションの排他性チェック
	if *contextOnly && *outputOnly {
		fmt.Fprintf(os.Stderr, "Error: --context-only and --output-only cannot be used together\n\n")
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
		showIntegratedExecutionPlan(*model, resolved, *contextOnly, *outputOnly)
		return nil
	}

	// APIクライアントを作成
	cfg := &config.AppConfig{
		BaseURL: resolved.Gateway.URL,
		APIKey:  resolved.Gateway.APIKey,
		Timeout: resolved.Gateway.Timeout,
	}

	client := api.NewProbeClient(cfg)

	// デフォルトのログ設定を取得
	probeConfig := internalConfig.GetDefaultProbeConfig()

	// ログ設定の準備
	var logger logging.ProbeLogger
	if !*noLog {
		// CLI引数で設定を上書き
		if *logDir != "" {
			probeConfig.Log.Dir = *logDir
		}

		// ログ作成
		logger, err = logging.NewProbeLogger(probeConfig.Log.ConvertToProbeLogConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create logger: %v\n", err)
			logger = nil
		} else {
			defer logger.Close()
		}
	}

	// 結果保存の準備
	var resultStorage storage.ResultStorage
	if *saveResult {
		resultStorage, err = storage.NewResultStorage(probeConfig.Result.Dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create result storage: %v\n", err)
			resultStorage = nil
		}
	}

	// 統合結果の構造体
	var contextResult *probe.ContextWindowResult
	var outputResult *probe.MaxOutputResult
	var totalDuration time.Duration
	var totalTrials int

	// 実行モードに応じて探索を実行
	if *contextOnly {
		// Context Windowのみ測定
		if *verbose {
			fmt.Printf("Probing context window for model %s...\n", *model)
		}

		start := time.Now()
		prober := probe.NewContextWindowProbe(client)
		contextResult, err = prober.Probe(*model, *verbose)
		if err != nil {
			return fmt.Errorf("failed to probe context window: %w", err)
		}
		totalDuration = time.Since(start)
		totalTrials = len(contextResult.TrialHistory)

	} else if *outputOnly {
		// Max Output Tokensのみ測定
		if *verbose {
			fmt.Printf("Probing max output tokens for model %s...\n", *model)
		}

		start := time.Now()
		prober := probe.NewMaxOutputTokensProbe(client)
		outputResult, err = prober.ProbeOutputTokens(*model, *verbose)
		if err != nil {
			return fmt.Errorf("failed to probe max output tokens: %w", err)
		}
		totalDuration = time.Since(start)
		totalTrials = len(outputResult.TrialHistory)

	} else {
		// 両方を測定（デフォルト）
		if *verbose {
			fmt.Printf("Probing both context window and max output tokens for model %s...\n", *model)
		}

		// 1. Context Window測定（時間がかかる方を先に）
		start := time.Now()
		prober := probe.NewContextWindowProbe(client)
		contextResult, err = prober.Probe(*model, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to probe context window: %v\n", err)
			contextResult = nil
		}
		contextDuration := time.Since(start)

		// 2. Max Output Tokens測定
		start = time.Now()
		maxProber := probe.NewMaxOutputTokensProbe(client)
		outputResult, err = maxProber.ProbeOutputTokens(*model, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to probe max output tokens: %v\n", err)
			outputResult = nil
		}
		outputDuration := time.Since(start)

		totalDuration = contextDuration + outputDuration
		if contextResult != nil {
			totalTrials += len(contextResult.TrialHistory)
		}
		if outputResult != nil {
			totalTrials += len(outputResult.TrialHistory)
		}
	}

	// 統合結果を表示
	formatter := ui.NewTableFormatter()
	output := formatter.FormatIntegratedResult(*model, contextResult, outputResult, totalDuration, totalTrials)
	fmt.Println(output)

	// ログ保存
	if logger != nil {
		if contextResult != nil {
			// Context Windowの試行履歴をログ
			for i, trial := range contextResult.TrialHistory {
				logEntry := logging.TrialLogEntry{
					Index:        i,
					TokenCount:   trial.TokenCount,
					Success:      trial.Success,
					Message:      trial.Message,
					Timestamp:    time.Now(),
					Duration:     contextResult.Duration / time.Duration(len(contextResult.TrialHistory)),
				}
				if err := logger.LogTrial(*model, resolved.Gateway.Name, "context", logEntry); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to log context trial: %v\n", err)
				}
			}

			// Context Windowの最終結果をログ
			if err := logger.LogResult(*model, resolved.Gateway.Name, "context", contextResult); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to log context result: %v\n", err)
			}
		}

		if outputResult != nil {
			// Max Outputの試行履歴をログ
			for i, trial := range outputResult.TrialHistory {
				logEntry := logging.TrialLogEntry{
					Index:        i,
					TokenCount:   trial.TokenCount,
					Success:      trial.Success,
					Message:      trial.Message,
					Timestamp:    time.Now(),
					Duration:     outputResult.Duration / time.Duration(len(outputResult.TrialHistory)),
				}
				if err := logger.LogTrial(*model, resolved.Gateway.Name, "max_output", logEntry); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to log max output trial: %v\n", err)
				}
			}

			// Max Outputの最終結果をログ
			if err := logger.LogResult(*model, resolved.Gateway.Name, "max_output", outputResult); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to log max output result: %v\n", err)
			}
		}
	}

	// 結果保存
	if resultStorage != nil {
		provider := extractProviderName(resolved.Gateway.URL)

		if contextResult != nil {
			if err := resultStorage.SaveContextResult(provider, *model, contextResult); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save context result: %v\n", err)
			} else if *verbose {
				fmt.Printf("Context result saved to: %s\n", probeConfig.Result.Dir)
			}
		}

		if outputResult != nil {
			if err := resultStorage.SaveMaxOutputResult(provider, *model, outputResult); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save max output result: %v\n", err)
			} else if *verbose {
				fmt.Printf("Max output result saved to: %s\n", probeConfig.Result.Dir)
			}
		}
	}

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
	logDir := probeCmd.String("log-dir", "", "Directory to save probe logs")
	saveResult := probeCmd.Bool("save-result", false, "Save probe results to file")
	noLog := probeCmd.Bool("no-log", false, "Disable logging")
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

	// デフォルトのログ設定を取得
	probeConfig := internalConfig.GetDefaultProbeConfig()

	// ログ保存処理
	if !*noLog {
		// CLI引数で設定を上書き
		if *logDir != "" {
			probeConfig.Log.Dir = *logDir
		}

		// ログ作成
		logger, err := logging.NewProbeLogger(probeConfig.Log.ConvertToProbeLogConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create logger: %v\n", err)
		} else {
			defer logger.Close()

			// 試行履歴をログ
			for i, trial := range result.TrialHistory {
				logEntry := logging.TrialLogEntry{
					Index:        i,
					TokenCount:   trial.TokenCount,
					Success:      trial.Success,
					Message:      trial.Message,
					Timestamp:    time.Now(), // 実際は各試行のタイムスタンプを使用すべき
					Duration:     result.Duration / time.Duration(len(result.TrialHistory)),
				}
				if err := logger.LogTrial(*model, resolved.Gateway.Name, "context", logEntry); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to log trial: %v\n", err)
				}
			}

			// 最終結果をログ
			if err := logger.LogResult(*model, resolved.Gateway.Name, "context", result); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to log result: %v\n", err)
			}
		}
	}

	// 結果保存処理
	if *saveResult {
		resultStorage, err := storage.NewResultStorage(probeConfig.Result.Dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create result storage: %v\n", err)
		} else {
			// Provider名を取得（gateway名から推測）
			provider := extractProviderName(resolved.Gateway.URL)
			if err := resultStorage.SaveContextResult(provider, *model, result); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save result: %v\n", err)
			} else if *verbose {
				fmt.Printf("Result saved to: %s\n", probeConfig.Result.Dir)
			}
		}
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
	logDir := probeCmd.String("log-dir", "", "Directory to save probe logs")
	saveResult := probeCmd.Bool("save-result", false, "Save probe results to file")
	noLog := probeCmd.Bool("no-log", false, "Disable logging")
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

	// デフォルトのログ設定を取得
	probeConfig := internalConfig.GetDefaultProbeConfig()

	// ログ保存処理
	if !*noLog {
		// CLI引数で設定を上書き
		if *logDir != "" {
			probeConfig.Log.Dir = *logDir
		}

		// ログ作成
		logger, err := logging.NewProbeLogger(probeConfig.Log.ConvertToProbeLogConfig())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create logger: %v\n", err)
		} else {
			defer logger.Close()

			// 試行履歴をログ
			for i, trial := range result.TrialHistory {
				logEntry := logging.TrialLogEntry{
					Index:        i,
					TokenCount:   trial.TokenCount,
					Success:      trial.Success,
					Message:      trial.Message,
					Timestamp:    time.Now(),
					Duration:     result.Duration / time.Duration(len(result.TrialHistory)),
				}
				if err := logger.LogTrial(*model, resolved.Gateway.Name, "max_output", logEntry); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to log trial: %v\n", err)
				}
			}

			// 最終結果をログ
			if err := logger.LogResult(*model, resolved.Gateway.Name, "max_output", result); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to log result: %v\n", err)
			}
		}
	}

	// 結果保存処理
	if *saveResult {
		resultStorage, err := storage.NewResultStorage(probeConfig.Result.Dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create result storage: %v\n", err)
		} else {
			// Provider名を取得（gateway名から推測）
			provider := extractProviderName(resolved.Gateway.URL)
			if err := resultStorage.SaveMaxOutputResult(provider, *model, result); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save result: %v\n", err)
			} else if *verbose {
				fmt.Printf("Result saved to: %s\n", probeConfig.Result.Dir)
			}
		}
	}

	return nil
}

// showProbeHelp はprobeコマンドのヘルプを表示する
func showProbeHelp() {
	fmt.Println(`llm-info probe - Probe model constraints via actual API behavior

USAGE:
    llm-info probe --model <MODEL_ID> [flags]

FLAGS:
    --model string              Target model ID (required)
    --url string                 Base URL of the LLM gateway
    --api-key string             API key for authentication
    --gateway string             Gateway name to use from config
    --timeout duration           Request timeout (default: 30s)
    --dry-run                   Show execution plan without making actual API calls
    --verbose                   Show verbose logs
    --log-dir string            Directory to save probe logs
    --save-result               Save probe results to file
    --no-log                   Disable logging
    --context-only              Probe only context window
    --output-only               Probe only max output tokens
    --config string              Path to config file
    --help                      Show help for probe command

EXAMPLES:
    # Basic usage (probes both context window and max output tokens)
    llm-info probe --model gpt-4o-mini

    # With custom gateway
    llm-info probe --model gpt-4o-mini --gateway production

    # Dry run to see execution plan
    llm-info probe --model gpt-4o-mini --dry-run

    # Verbose output
    llm-info probe --model gpt-4o-mini --verbose

    # Probe only context window
    llm-info probe --model gpt-4o-mini --context-only

    # Probe only max output tokens
    llm-info probe --model gpt-4o-mini --output-only

    # Save results and logs
    llm-info probe --model gpt-4o-mini --save-result --log-dir ./logs`)
}

// showIntegratedExecutionPlan は統合探索の実行計画を表示する
func showIntegratedExecutionPlan(model string, config *internalConfig.ResolvedConfig, contextOnly, outputOnly bool) {
	fmt.Printf("Model Constraints Probe Execution Plan:\n")
	fmt.Printf("  Model: %s\n", model)
	fmt.Printf("  URL: %s\n", config.Gateway.URL)
	fmt.Printf("  API Key: %s\n", maskAPIKey(config.Gateway.APIKey))
	fmt.Printf("  Timeout: %s\n", config.Gateway.Timeout)

	if contextOnly {
		fmt.Printf("\nProbe Mode: Context Window Only\n")
		fmt.Printf("\nProbe Phases:\n")
		fmt.Printf("  1. Exponential Search: Find upper bound by doubling token count\n")
		fmt.Printf("  2. Binary Search: Refine boundary within ±1024 tokens\n")
		fmt.Printf("  3. Error Analysis: Extract token limits from error messages\n")
		fmt.Printf("\nAPI Calls:\n")
		fmt.Printf("  POST %s/v1/chat/completions\n", config.Gateway.URL)
		fmt.Printf("  - Test data generation with Japanese text\n")
		fmt.Printf("  - Needle-in-haystack methodology\n")
		fmt.Printf("  - Rate limited: 1 second between calls\n")
	} else if outputOnly {
		fmt.Printf("\nProbe Mode: Max Output Tokens Only\n")
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
	} else {
		fmt.Printf("\nProbe Mode: Both Context Window and Max Output Tokens\n")
		fmt.Printf("\nExecution Order:\n")
		fmt.Printf("  1. Context Window Probing (time-intensive phase first)\n")
		fmt.Printf("  2. Max Output Tokens Probing\n")
		fmt.Printf("\nContext Window Probe Phases:\n")
		fmt.Printf("  1. Exponential Search: Find upper bound by doubling token count\n")
		fmt.Printf("  2. Binary Search: Refine boundary within ±1024 tokens\n")
		fmt.Printf("  3. Error Analysis: Extract token limits from error messages\n")
		fmt.Printf("\nMax Output Tokens Probe Phases:\n")
		fmt.Printf("  1. Exponential Search: Find upper bound by doubling token count (256→512→1024...)\n")
		fmt.Printf("  2. Binary Search: Refine boundary within detection range\n")
		fmt.Printf("  3. Evidence Analysis: Determine limit source (error/incomplete)\n")
		fmt.Printf("\nAPI Calls:\n")
		fmt.Printf("  POST %s/v1/chat/completions\n", config.Gateway.URL)
		fmt.Printf("  - Context: Variable input lengths with fixed output\n")
		fmt.Printf("  - Max Output: Fixed input with variable max_tokens\n")
		fmt.Printf("  - Rate limiting: 1s (context) / 0.5s (output) between calls\n")
	}

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
    --log-dir string     Directory to save probe logs
    --save-result       Save probe results to file
    --no-log           Disable logging
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
    llm-info probe-context --model gpt-4o-mini --verbose`)
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
    --log-dir string     Directory to save probe logs
    --save-result       Save probe results to file
    --no-log           Disable logging
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
    It detects limits via validation errors and incomplete response status.`)
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

// extractProviderName はURLからプロバイダ名を抽出する
func extractProviderName(url string) string {
	if strings.Contains(url, "openai") {
		return "openai"
	}
	if strings.Contains(url, "anthropic") {
		return "anthropic"
	}
	if strings.Contains(url, "google") {
		return "google"
	}
	if strings.Contains(url, "azure") {
		return "azure"
	}
	// デフォルトは "unknown"
	return "unknown"
}