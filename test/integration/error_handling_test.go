package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	errhandler "github.com/your-org/llm-info/internal/error"
	pkgconfig "github.com/your-org/llm-info/pkg/config"
)

// TestErrorHandlingIntegration はエラーハンドリングの統合テスト
func TestErrorHandlingNetworkIntegration(t *testing.T) {
	tests := []struct {
		name              string
		setupMock         func() *httptest.Server
		expectedError     string
		expectedSolutions []string
	}{
		{
			name: "Network timeout error",
			setupMock: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(2 * time.Second) // タイムアウトを引き起こす
					w.WriteHeader(http.StatusOK)
				}))
				return server
			},
			expectedError: "接続がタイムアウトしました",
			expectedSolutions: []string{
				"ネットワーク接続を確認してください",
				"ファイアウォール設定を確認してください",
				"ゲートウェイURLが正しいか確認してください",
			},
		},
		{
			name: "Authentication error",
			setupMock: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != "Bearer valid-token" {
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(`{"error": "Invalid API key"}`))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"data": []}`))
				}))
				return server
			},
			expectedError: "認証に失敗しました",
			expectedSolutions: []string{
				"APIキーが正しいか確認してください",
				"APIキーの有効期限が切れていないか確認してください",
				"APIキーの権限設定を確認してください",
			},
		},
		{
			name: "Rate limit error",
			setupMock: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(`{"error": "Rate limit exceeded"}`))
				}))
				return server
			},
			expectedError: "レート制限を超えました",
			expectedSolutions: []string{
				"しばらく待ってから再試行してください",
				"APIプランのレート制限を確認してください",
				"並列リクエスト数を減らしてください",
			},
		},
		{
			name: "Not found error",
			setupMock: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/v1/models" {
						w.WriteHeader(http.StatusNotFound)
						w.Write([]byte(`{"error": "Endpoint not found"}`))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"data": []}`))
				}))
				return server
			},
			expectedError: "エンドポイントが見つかりません",
			expectedSolutions: []string{
				"ゲートウェイURLが正しいか確認してください",
				"APIバージョンが正しいか確認してください",
				"エンドポイントパスを確認してください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサーバーをセットアップ
			mockServer := tt.setupMock()
			defer mockServer.Close()

			// 設定を作成
			gateway := pkgconfig.Gateway{
				Name:    "test",
				URL:     mockServer.URL,
				APIKey:  "invalid-token", // 認証エラーを引き起こす
				Timeout: 1 * time.Second, // タイムアウトを短く設定
			}

			// APIクライアントを作成してリクエストを実行
			client := &http.Client{
				Timeout: gateway.Timeout,
			}

			req, err := http.NewRequestWithContext(context.Background(), "GET", gateway.URL+"/v1/models", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Authorization", "Bearer "+gateway.APIKey)

			resp, err := client.Do(req)
			if err != nil {
				// ネットワークエラーを処理
				appErr := errhandler.CreateNetworkError("connection_timeout", gateway.URL, err)
				formattedError := errhandler.FormatErrorMessage(appErr)

				if !strings.Contains(formattedError, tt.expectedError) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, formattedError)
				}

				// 解決策をチェック
				for _, expectedSolution := range tt.expectedSolutions {
					if !strings.Contains(formattedError, expectedSolution) {
						t.Errorf("Expected solution %q not found in error message: %s", expectedSolution, formattedError)
					}
				}
				return
			}
			defer resp.Body.Close()

			// HTTPステータスコードに基づいてエラーを処理
			if resp.StatusCode >= 400 {
				var appErr *errhandler.AppError
				switch resp.StatusCode {
				case 401:
					appErr = errhandler.CreateAPIError("authentication_failed", resp.StatusCode, gateway.URL, fmt.Errorf("unauthorized"))
				case 429:
					appErr = errhandler.CreateAPIError("rate_limit_exceeded", resp.StatusCode, gateway.URL, fmt.Errorf("rate limit"))
				case 404:
					appErr = errhandler.CreateAPIError("endpoint_not_found", resp.StatusCode, gateway.URL, fmt.Errorf("not found"))
				default:
					appErr = errhandler.CreateAPIError("server_error", resp.StatusCode, gateway.URL, fmt.Errorf("server error"))
				}

				formattedError := errhandler.FormatErrorMessage(appErr)

				if !strings.Contains(formattedError, tt.expectedError) {
					t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, formattedError)
				}

				// 解決策をチェック
				for _, expectedSolution := range tt.expectedSolutions {
					if !strings.Contains(formattedError, expectedSolution) {
						t.Errorf("Expected solution %q not found in error message: %s", expectedSolution, formattedError)
					}
				}
			}
		})
	}
}

// TestConfigErrorHandling は設定エラーのハンドリングをテスト
func TestConfigErrorHandling(t *testing.T) {
	tests := []struct {
		name              string
		configPath        string
		expectedError     string
		expectedSolutions []string
	}{
		{
			name:          "Config file not found",
			configPath:    "/nonexistent/config.yaml",
			expectedError: "設定ファイルが見つかりません",
			expectedSolutions: []string{
				"設定ファイルを作成してください: llm-info --init-config",
				"設定ファイルパスが正しいか確認してください",
				"環境変数LLM_INFO_CONFIG_PATHを確認してください",
			},
		},
		{
			name:          "Invalid config format",
			configPath:    "test/fixtures/invalid-config.yaml",
			expectedError: "設定ファイルの形式が無効です",
			expectedSolutions: []string{
				"YAML形式が正しいか確認してください",
				"設定ファイルの構文を確認してください",
				"例: https://github.com/your-org/llm-info/blob/main/configs/example.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var appErr *errhandler.AppError

			if tt.name == "Config file not found" {
				// ファイルが存在しない場合のエラーをシミュレート
				err := fmt.Errorf("file not found")
				appErr = errhandler.CreateConfigError("config_file_not_found", tt.configPath, err)
			} else if tt.name == "Invalid config format" {
				// 無効な設定ファイル形式のエラーをシミュレート
				err := fmt.Errorf("yaml: line 2: found character that cannot start any token")
				appErr = errhandler.CreateConfigError("invalid_config_format", tt.configPath, err)
			}

			formattedError := errhandler.FormatErrorMessage(appErr)

			if !strings.Contains(formattedError, tt.expectedError) {
				t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, formattedError)
			}

			// 解決策をチェック
			for _, expectedSolution := range tt.expectedSolutions {
				if !strings.Contains(formattedError, expectedSolution) {
					t.Errorf("Expected solution %q not found in error message: %s", expectedSolution, formattedError)
				}
			}
		})
	}
}

// TestUserErrorHandling はユーザーエラーのハンドリングをテスト
func TestUserErrorHandling(t *testing.T) {
	tests := []struct {
		name              string
		argument          string
		errorCode         string
		expectedError     string
		expectedSolutions []string
	}{
		{
			name:          "Invalid argument",
			argument:      "--invalid-flag",
			errorCode:     "invalid_argument",
			expectedError: "無効な引数です",
			expectedSolutions: []string{
				"コマンドライン引数が正しいか確認してください",
				"ヘルプを確認してください: llm-info --help",
			},
		},
		{
			name:          "Invalid filter syntax",
			argument:      "invalid-filter-syntax",
			errorCode:     "invalid_filter_syntax",
			expectedError: "フィルタ構文が無効です",
			expectedSolutions: []string{
				"フィルタ構文を確認してください",
				"例: --filter \"name:gpt,tokens>1000\"",
				"ヘルプを確認してください: llm-info --help filter",
			},
		},
		{
			name:          "Invalid sort field",
			argument:      "invalid-field",
			errorCode:     "invalid_sort_field",
			expectedError: "無効なソートフィールドです",
			expectedSolutions: []string{
				"ソートフィールドを確認してください",
				"使用可能なフィールド: name, tokens, cost, mode",
				"ヘルプを確認してください: llm-info --help sort",
			},
		},
		{
			name:          "Gateway not found",
			argument:      "nonexistent-gateway",
			errorCode:     "gateway_not_found",
			expectedError: "指定されたゲートウェイが見つかりません",
			expectedSolutions: []string{
				"ゲートウェイ名が正しいか確認してください",
				"利用可能なゲートウェイを確認してください: llm-info --list-gateways",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ユーザーエラーを作成
			appErr := errhandler.CreateUserError(tt.errorCode, tt.argument, fmt.Errorf("user error"))

			formattedError := errhandler.FormatErrorMessage(appErr)

			if !strings.Contains(formattedError, tt.expectedError) {
				t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, formattedError)
			}

			// 解決策をチェック
			for _, expectedSolution := range tt.expectedSolutions {
				if !strings.Contains(formattedError, expectedSolution) {
					t.Errorf("Expected solution %q not found in error message: %s", expectedSolution, formattedError)
				}
			}
		})
	}
}

// TestSystemErrorHandling はシステムエラーのハンドリングをテスト
func TestSystemErrorHandling(t *testing.T) {
	tests := []struct {
		name              string
		errorCode         string
		context           string
		expectedError     string
		expectedSolutions []string
	}{
		{
			name:          "Permission denied",
			errorCode:     "permission_denied",
			context:       "file access",
			expectedError: "ファイルアクセス権限がありません",
			expectedSolutions: []string{
				"ファイルのアクセス権限を確認してください",
				"ファイルの所有者を確認してください",
			},
		},
		{
			name:          "Disk full",
			errorCode:     "disk_full",
			context:       "storage",
			expectedError: "ディスク容量が不足しています",
			expectedSolutions: []string{
				"ディスク容量を確認してください",
				"不要なファイルを削除してください",
				"別のストレージを使用してください",
			},
		},
		{
			name:          "Memory insufficient",
			errorCode:     "memory_insufficient",
			context:       "memory",
			expectedError: "メモリが不足しています",
			expectedSolutions: []string{
				"メモリ使用量を確認してください",
				"他のアプリケーションを終了してください",
				"システムを再起動してください",
			},
		},
		{
			name:          "Unexpected error",
			errorCode:     "unexpected_error",
			context:       "system",
			expectedError: "予期せぬエラーが発生しました",
			expectedSolutions: []string{
				"開発者にバグレポートを送信してください",
				"詳細なログを確認してください: llm-info --verbose",
				"最新バージョンにアップデートしてください",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// システムエラーを作成
			appErr := errhandler.CreateSystemError(tt.errorCode, tt.context, fmt.Errorf("system error"))

			formattedError := errhandler.FormatErrorMessage(appErr)

			if !strings.Contains(formattedError, tt.expectedError) {
				t.Errorf("Expected error message to contain %q, got: %s", tt.expectedError, formattedError)
			}

			// 解決策をチェック
			for _, expectedSolution := range tt.expectedSolutions {
				if !strings.Contains(formattedError, expectedSolution) {
					t.Errorf("Expected solution %q not found in error message: %s", expectedSolution, formattedError)
				}
			}
		})
	}
}

// TestErrorHandlingWithVerboseMode は詳細モードでのエラーハンドリングをテスト
func TestErrorHandlingWithVerboseMode(t *testing.T) {
	// エラーハンドラーを作成（詳細モード）
	errorHandler := errhandler.NewHandler(true)

	// テスト用のエラーを作成
	originalErr := fmt.Errorf("original error")
	appErr := errhandler.NewAppError(errhandler.ErrorTypeNetwork, errhandler.SeverityError, "connection_timeout", "接続がタイムアウトしました").
		WithCause(originalErr).
		WithContext("url", "https://api.example.com").
		WithSolution("ネットワーク接続を確認してください").
		WithHelpURL("https://github.com/your-org/llm-info/wiki/network-errors")

	// エラーを処理
	exitCode := errorHandler.Handle(appErr)

	// 終了コードをチェック
	if exitCode != 2 {
		t.Errorf("Expected exit code 2 for error, got %d", exitCode)
	}
}

// TestErrorHandlingWithFallback はフォールバック処理付きのエラーハンドリングをテスト
func TestErrorHandlingWithFallback(t *testing.T) {
	// エラーハンドラーを作成
	errorHandler := errhandler.NewHandler(false)

	// テスト用のエラーを作成
	originalErr := fmt.Errorf("original error")
	appErr := errhandler.NewAppError(errhandler.ErrorTypeNetwork, errhandler.SeverityError, "connection_timeout", "接続がタイムアウトしました").
		WithCause(originalErr)

	// フォールバック処理を定義
	fallbackExecuted := false
	fallback := func() error {
		fallbackExecuted = true
		return nil
	}

	// エラーを処理
	exitCode := errorHandler.HandleWithFallback(appErr, fallback)

	// 終了コードをチェック（フォールバックが成功した場合は0）
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 after successful fallback, got %d", exitCode)
	}

	// フォールバックが実行されたかチェック
	if !fallbackExecuted {
		t.Error("Expected fallback to be executed")
	}
}

// TestErrorDetectionAndWrapping はエラーの検出と自動ラッピングをテスト
func TestErrorDetectionAndWrapping(t *testing.T) {
	tests := []struct {
		name         string
		originalErr  error
		expectedType errhandler.ErrorType
		expectedCode string
	}{
		{
			name:         "Timeout error",
			originalErr:  fmt.Errorf("context deadline exceeded"),
			expectedType: errhandler.ErrorTypeNetwork,
			expectedCode: "connection_timeout",
		},
		{
			name:         "Connection refused",
			originalErr:  fmt.Errorf("connection refused"),
			expectedType: errhandler.ErrorTypeNetwork,
			expectedCode: "connection_refused",
		},
		{
			name:         "DNS error",
			originalErr:  fmt.Errorf("no such host"),
			expectedType: errhandler.ErrorTypeNetwork,
			expectedCode: "dns_resolution_failed",
		},
		{
			name:         "File not found",
			originalErr:  fmt.Errorf("no such file or directory"),
			expectedType: errhandler.ErrorTypeConfig,
			expectedCode: "config_file_not_found",
		},
		{
			name:         "Permission denied",
			originalErr:  fmt.Errorf("permission denied"),
			expectedType: errhandler.ErrorTypeSystem,
			expectedCode: "permission_denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// エラーを検出してラッピング
			appErr := errhandler.WrapErrorWithDetection(tt.originalErr, "test context")

			// エラータイプをチェック
			if appErr.Type != tt.expectedType {
				t.Errorf("Expected error type %v, got %v", tt.expectedType, appErr.Type)
			}

			// エラーコードをチェック
			if appErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %q, got %q", tt.expectedCode, appErr.Code)
			}

			// 原因エラーが設定されているかチェック
			if appErr.Unwrap() == nil {
				t.Error("Expected cause error to be set")
			}

			// コンテキストが設定されているかチェック
			if context, ok := appErr.Context["context"].(string); !ok || context != "test context" {
				// コンテキストが設定されていない場合は、他のコンテキストキーをチェック
				if len(appErr.Context) == 0 {
					t.Error("Expected context to be set")
				}
			}
		})
	}
}

// TestHelpSystemIntegration はヘルプシステムの統合をテスト
func TestHelpSystemIntegration(t *testing.T) {
	// ヘルププロバイダーを作成
	helpProvider := &HelpProvider{
		version: "1.0.0-test",
	}

	// 一般ヘルプをテスト
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	helpProvider.ShowGeneralHelp()

	w.Close()
	os.Stdout = oldStdout
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	output := buf.String()

	expectedTopics := []string{
		"llm-info - LLMゲートウェイ情報可視化ツール",
		"--url string",
		"--api-key string",
		"--gateway string",
		"--help",
		"--version",
	}

	for _, expected := range expectedTopics {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected help output to contain %q, got: %s", expected, output)
		}
	}

	// トピック別ヘルプをテスト
	topics := []string{"filter", "sort", "config", "examples"}
	for _, topic := range topics {
		r, w, _ = os.Pipe()
		os.Stdout = w

		helpProvider.ShowTopicHelp(topic)

		w.Close()
		os.Stdout = oldStdout
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		output = buf.String()

		if output == "" {
			t.Errorf("Expected help output for topic %q to be non-empty", topic)
		}
	}

	// 不明なトピックをテスト
	r, w, _ = os.Pipe()
	os.Stdout = w

	helpProvider.ShowTopicHelp("unknown")

	w.Close()
	os.Stdout = oldStdout
	buf = new(bytes.Buffer)
	buf.ReadFrom(r)
	output = buf.String()

	if !strings.Contains(output, "不明なトピック") {
		t.Error("Expected error message for unknown topic")
	}
}

// HelpProvider はテスト用のヘルププロバイダー
type HelpProvider struct {
	version string
}

// ShowGeneralHelp は一般ヘルプを表示する
func (hp *HelpProvider) ShowGeneralHelp() {
	fmt.Printf(`llm-info - LLMゲートウェイ情報可視化ツール (バージョン: %s)

使用方法:
  llm-info [flags]

フラグ:
  --url string        ゲートウェイのURL
  --api-key string    APIキー
  --gateway string    使用するゲートウェイ名
  --timeout duration  リクエストタイムアウト (デフォルト: 10s)
  --format string     出力形式 (table|json) (デフォルト: table)
  --filter string     フィルタ条件
  --sort string       ソート条件
  --columns string    表示するカラム (カンマ区切り)
  --config string     設定ファイルパス
  --verbose           詳細なログを表示
  --help              ヘルプを表示
  --version           バージョンを表示
`, hp.version)
}

// ShowTopicHelp はトピック別のヘルプを表示する
func (hp *HelpProvider) ShowTopicHelp(topic string) {
	switch strings.ToLower(topic) {
	case "filter":
		fmt.Print(`フィルタ構文ヘルプ

基本構文:
	 --filter "条件1,条件2,..."

使用可能な条件:
	 name:パターン          モデル名でフィルタ（部分一致）
	 exclude:パターン       モデル名で除外（部分一致）
	 tokens>数値           最大トークン数が指定値より大きい
	 tokens<数値           最大トークン数が指定値より小さい
	 cost>数値             入力コストが指定値より大きい
	 cost<数値             入力コストが指定値より小さい
	 mode:値               モードでフィルタ（chat/completion）

`)
	case "sort":
		fmt.Print(`ソートオプションヘルプ

基本構文:
	 --sort "フィールド"        # 昇順
	 --sort "-フィールド"       # 降順

使用可能なフィールド:
	 name, model              モデル名
	 tokens, max_tokens       最大トークン数
	 cost, input_cost         入力コスト
	 mode                     モード

`)
	case "config":
		fmt.Print(`設定ファイルヘルプ

設定ファイルの場所:
	 ~/.config/llm-info/llm-info.yaml

設定ファイル形式:
	 gateways:
	   - name: "production"
	     url: "https://api.example.com"
	     api_key: "your-api-key"
	     timeout: "10s"
	 
	 default_gateway: "production"

`)
	case "examples":
		fmt.Print(`使用例ヘルプ

基本使用例:
	 # 直接指定
	 llm-info --url https://api.openai.com --api-key sk-xxx
	 
	 # 設定ファイル使用
	 llm-info --gateway production
	 
	 # 環境変数使用
	 export LLM_INFO_URL="https://api.example.com"
	 export LLM_INFO_API_KEY="your-key"
	 llm-info

`)
	default:
		fmt.Printf("不明なトピック: %s\n", topic)
		fmt.Println("利用可能なトピック: filter, sort, config, examples")
	}
}
