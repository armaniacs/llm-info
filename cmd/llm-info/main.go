package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/armaniacs/llm-info/internal/api"
	"github.com/armaniacs/llm-info/internal/config"
	internalConfig "github.com/armaniacs/llm-info/internal/config"
	errhandler "github.com/armaniacs/llm-info/internal/error"
	"github.com/armaniacs/llm-info/internal/model"
	"github.com/armaniacs/llm-info/internal/ui"
	pkgconfig "github.com/armaniacs/llm-info/pkg/config"
)

const version = "1.0.0"

func main() {
	// エラーハンドラーの初期化
	verbose := os.Getenv("LLM_INFO_DEBUG") != "" || os.Getenv("LLM_INFO_VERBOSE") != ""
	errorHandler := errhandler.NewHandler(verbose)

	// パニックから回復
	defer errorHandler.Recover()

	// コマンドライン引数の定義
	var (
		url          = flag.String("url", "", "Base URL of the LLM gateway")
		apiKey       = flag.String("api-key", "", "API key for authentication")
		timeout      = flag.Duration("timeout", 10*time.Second, "Request timeout (default: 10s)")
		configFile   = flag.String("config", "", "Path to config file")
		gateway      = flag.String("gateway", "", "Gateway name to use from config")
		outputFormat = flag.String("format", "table", "Output format (table, json)")
		sortBy       = flag.String("sort", "", "Sort models by field (name, max_tokens, mode, input_cost). Use - prefix for descending order")
		filter       = flag.String("filter", "", "Filter models (e.g., 'name:gpt,tokens>1000,mode:chat')")
		columns      = flag.String("columns", "", "Specify columns to display (e.g., 'name,max_tokens')")
		showHelp     = flag.Bool("help", false, "Show help")
		showVersion  = flag.Bool("version", false, "Show version")
		showSources  = flag.Bool("show-sources", false, "Show configuration sources")
		verboseFlag  = flag.Bool("verbose", false, "Show verbose logs")
		initConfig   = flag.Bool("init-config", false, "Create config file template")
		checkConfig  = flag.Bool("check-config", false, "Validate config file")
		listGateways = flag.Bool("list-gateways", false, "List configured gateways")
		helpTopic    = flag.String("help-topic", "", "Show help for specific topic (filter, sort, config, examples, errors)")
	)

	// ヘルププロバイダーの初期化
	helpProvider := NewHelpProvider(version)

	flag.Usage = func() {
		helpProvider.ShowGeneralHelp()
	}

	flag.Parse()

	// 詳細モードの設定
	if *verboseFlag {
		verbose = true
		errorHandler = errhandler.NewHandler(true)
	}

	// トピック別ヘルプの表示
	if *helpTopic != "" {
		helpProvider.ShowTopicHelp(*helpTopic)
		os.Exit(0)
	}

	// ヘルプの表示
	if *showHelp {
		helpProvider.ShowGeneralHelp()
		os.Exit(0)
	}

	// バージョンの表示
	if *showVersion {
		helpProvider.ShowVersion()
		os.Exit(0)
	}

	// 設定ファイルテンプレートの作成
	if *initConfig {
		if err := createConfigTemplate(); err != nil {
			os.Exit(errorHandler.Handle(err))
		}
		os.Exit(0)
	}

	// 設定ファイルの検証
	if *checkConfig {
		if err := validateConfigFile(*configFile); err != nil {
			os.Exit(errorHandler.Handle(err))
		}
		os.Exit(0)
	}

	// ゲートウェイ一覧の表示
	if *listGateways {
		if err := listConfiguredGateways(*configFile); err != nil {
			os.Exit(errorHandler.Handle(err))
		}
		os.Exit(0)
	}

	// 設定マネージャーの初期化
	configPath := *configFile
	if configPath == "" {
		configPath = internalConfig.GetDefaultConfigPath()
	}
	configManager := internalConfig.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := configManager.Load(); err != nil {
		// 設定ファイルが存在しない場合は警告のみ表示
		if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "config file not found") {
			appErr := errhandler.CreateConfigError("config_file_not_found", configPath, err)
			os.Exit(errorHandler.Handle(appErr))
		}
	}

	// コマンドライン引数の構造体を作成
	cliArgs := &config.CLIArgs{
		URL:          *url,
		APIKey:       *apiKey,
		Timeout:      *timeout,
		Gateway:      *gateway,
		OutputFormat: *outputFormat,
		SortBy:       *sortBy,
		Filter:       *filter,
		Columns:      *columns,
	}

	// 設定の解決（優先順位: CLI > 環境変数 > 設定ファイル > デフォルト）
	resolvedConfig, err := configManager.ResolveConfig(cliArgs)
	if err != nil {
		appErr := errhandler.CreateConfigError("missing_required_field", configPath, err)
		os.Exit(errorHandler.Handle(appErr))
	}

	// 設定ソース情報の表示
	if *showSources {
		fmt.Println(configManager.GetConfigSourceInfo(resolvedConfig))
		os.Exit(0)
	}

	// 従来の設定構造体に変換（既存コードとの互換性のため）
	cfg := config.New(resolvedConfig.Gateway.URL, resolvedConfig.Gateway.APIKey, resolvedConfig.Gateway.Timeout)

	// URLの形式を検証
	if err := validateURL(resolvedConfig.Gateway.URL); err != nil {
		appErr := errhandler.CreateUserError("invalid_argument", resolvedConfig.Gateway.URL, err)
		os.Exit(errorHandler.Handle(appErr))
	}

	// APIクライアントの作成
	client := api.NewClient(cfg)

	// エンドポイントURLを表示（エラー時にも表示するため）
	if err := ui.DisplayEndpoint(resolvedConfig.Gateway.URL); err != nil {
		// URL表示エラーは処理を継続
		fmt.Fprintf(os.Stderr, "Warning: failed to display endpoint: %v\n", err)
	}

	// モデル情報の取得（フォールバック機能付き）
	if verbose {
		fmt.Printf("Fetching model information from %s...\n", resolvedConfig.Gateway.URL)
	}
	response, err := client.FetchModelsWithFallback()
	if err != nil {
		// 新しいエラーハンドリングを使用
		appErr := errhandler.WrapErrorWithDetection(err, resolvedConfig.Gateway.URL)
		os.Exit(errorHandler.Handle(appErr))
	}

	// APIレスポンスをアプリケーションモデルに変換
	models := model.FromAPIResponse(response.Models)

	// 高度なフィルタリング
	if resolvedConfig.Filter != "" {
		filterCriteria, err := ui.ParseFilterString(resolvedConfig.Filter)
		if err != nil {
			appErr := errhandler.CreateUserError("invalid_filter_syntax", resolvedConfig.Filter, err)
			os.Exit(errorHandler.Handle(appErr))
		}
		models = ui.Filter(models, filterCriteria)
	}

	// 高度なソート
	if resolvedConfig.SortBy != "" {
		sortCriteria, err := ui.ParseSortString(resolvedConfig.SortBy)
		if err != nil {
			appErr := errhandler.CreateUserError("invalid_sort_field", resolvedConfig.SortBy, err)
			os.Exit(errorHandler.Handle(appErr))
		}
		ui.Sort(models, sortCriteria)
	}

	// 結果の表示
	if len(models) == 0 {
		fmt.Printf("⚠️  No models found. The gateway may not have any models configured.\n")
		fmt.Printf("💡 Try using --filter to adjust search criteria or check the gateway configuration.\n")
		os.Exit(0)
	}

	if verbose {
		fmt.Printf("✅ Found %d models:\n\n", len(models))
	}

	// 表示オプションの準備
	renderOptions := &ui.RenderOptions{
		Filter:  resolvedConfig.Filter,
		Sort:    resolvedConfig.SortBy,
		Columns: resolvedConfig.Columns,
	}

	// 出力形式に応じて表示
	switch resolvedConfig.OutputFormat {
	case "json":
		if err := ui.RenderJSONWithOptions(models, renderOptions); err != nil {
			appErr := errhandler.CreateSystemError("unexpected_error", "JSON rendering", err)
			os.Exit(errorHandler.Handle(appErr))
		}
	default:
		if err := ui.RenderTableWithOptions(models, renderOptions); err != nil {
			appErr := errhandler.CreateSystemError("unexpected_error", "table rendering", err)
			os.Exit(errorHandler.Handle(appErr))
		}
	}
}

// validateURL はURLの形式を検証します
func validateURL(urlStr string) error {
	// URLの形式を検証
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// スキームのチェック
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// ホストのチェック
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	return nil
}

// createConfigTemplate は設定ファイルのテンプレートを作成します
func createConfigTemplate() error {
	helpProvider := NewHelpProvider(version)
	helpProvider.ShowConfigTemplate()

	configPath := config.GetDefaultConfigPath()

	// ディレクトリが存在しない場合は作成
	configDir := configPath[:len(configPath)-len("/llm-info.yaml")]
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 設定ファイルが既に存在するか確認
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  設定ファイルは既に存在します: %s\n", configPath)
		fmt.Printf("上書きしますか？ [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("キャンセルしました。")
			return nil
		}
	}

	// テンプレート内容をファイルに書き込み
	templateContent := `# llm-info 設定ファイル
# このファイルを ~/.config/llm-info/llm-info.yaml に保存してください

# ゲートウェイ設定
gateways:
  # 本番環境ゲートウェイ
  - name: "production"
    url: "https://api.example.com"
    api_key: "your-production-api-key"
    timeout: "10s"
    description: "本番環境ゲートウェイ"
  
  # 開発環境ゲートウェイ
  - name: "development"
    url: "https://dev-api.example.com"
    api_key: "your-development-api-key"
    timeout: "5s"
    description: "開発環境ゲートウェイ"

# デフォルトゲートウェイ
default_gateway: "production"

# グローバル設定
global:
  timeout: "10s"
  output_format: "table"
  sort_by: "name"
  columns: "name,tokens,cost,mode"
  verbose: false
`

	if err := os.WriteFile(configPath, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✅ 設定ファイルを作成しました: %s\n", configPath)
	fmt.Println("このファイルを編集して、ご自身の環境に合わせて設定してください。")

	return nil
}

// validateConfigFile は設定ファイルを検証します
func validateConfigFile(configFile string) error {
	configPath := configFile
	if configPath == "" {
		configPath = internalConfig.GetDefaultConfigPath()
	}

	fmt.Printf("設定ファイルを検証します: %s\n", configPath)

	// 設定マネージャーの初期化
	configManager := internalConfig.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// ダミーのCLI引数を作成して設定を解決
	cliArgs := &internalConfig.CLIArgs{
		URL:          "",
		APIKey:       "",
		Timeout:      10 * time.Second,
		Gateway:      "",
		OutputFormat: "table",
		SortBy:       "",
		Filter:       "",
		Columns:      "",
	}

	// 設定の解決
	resolvedConfig, err := configManager.ResolveConfig(cliArgs)
	if err != nil {
		return fmt.Errorf("設定の解決に失敗しました: %w", err)
	}

	// URLの検証
	if resolvedConfig.Gateway.URL == "" {
		return fmt.Errorf("ゲートウェイURLが設定されていません")
	}

	if err := validateURL(resolvedConfig.Gateway.URL); err != nil {
		return fmt.Errorf("無効なゲートウェイURL: %w", err)
	}

	// APIキーの検証
	if resolvedConfig.Gateway.APIKey == "" {
		return fmt.Errorf("APIキーが設定されていません")
	}

	fmt.Println("✅ 設定ファイルは有効です")
	fmt.Printf("ゲートウェイURL: %s\n", resolvedConfig.Gateway.URL)
	fmt.Printf("タイムアウト: %s\n", resolvedConfig.Gateway.Timeout)

	return nil
}

// listConfiguredGateways は設定済みのゲートウェイを一覧表示します
func listConfiguredGateways(configFile string) error {
	configPath := configFile
	if configPath == "" {
		configPath = internalConfig.GetDefaultConfigPath()
	}

	fmt.Printf("設定済みゲートウェイ一覧: %s\n", configPath)

	// 設定マネージャーの初期化
	configManager := internalConfig.NewManager(configPath)

	// 設定ファイルの読み込み
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// 設定ファイルからゲートウェイ情報を取得
	var gateways []pkgconfig.GatewayConfig
	var defaultGateway string

	// 新しい形式の設定を試す
	if newConfig := configManager.GetNewConfig(); newConfig != nil {
		for _, gw := range newConfig.Gateways {
			gateways = append(gateways, pkgconfig.GatewayConfig{
				Name:    gw.Name,
				URL:     gw.URL,
				APIKey:  gw.APIKey,
				Timeout: gw.Timeout,
			})
		}
		defaultGateway = newConfig.DefaultGateway
	} else if fileConfig := configManager.GetFileConfig(); fileConfig != nil {
		gateways = fileConfig.Gateways
		defaultGateway = fileConfig.DefaultGateway
	}

	if len(gateways) == 0 {
		fmt.Println("⚠️  ゲートウェイが設定されていません")
		fmt.Println("設定ファイルにゲートウェイを追加してください。")
		return nil
	}

	fmt.Printf("デフォルトゲートウェイ: %s\n\n", defaultGateway)
	fmt.Println("利用可能なゲートウェイ:")

	for _, gateway := range gateways {
		fmt.Printf("  - %s\n", gateway.Name)
		fmt.Printf("    URL: %s\n", gateway.URL)
		if gateway.Timeout != 0 {
			fmt.Printf("    タイムアウト: %s\n", gateway.Timeout)
		}
		fmt.Println()
	}

	return nil
}
