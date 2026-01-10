# PBI 5: エラーハンドリングとユーザビリティ向上 - 実装計画

## PBI概要と目的

**タイトル**: 詳細なエラーメッセージとヘルプ機能

**目的**: エラーが発生した際に具体的な原因と解決策を提示し、迅速な問題解決を支援すること。

**ビジネス価値**:
- ユーザー体験：分かりやすいエラーメッセージ
- 問題解決：自己解決能力の向上
- 学習支援：ツールの効果的な使用方法の提供

## 現状の課題

1. エラーメッセージが技術的で分かりにくい
2. 解決策の提示がない
3. ヘルプ機能が不十分
4. 使用例が少ない
5. エラーの種類ごとの適切なハンドリングがない

## 実装計画

### 1. エラー分類とメッセージ設計

#### エラー種別の定義
1. **ネットワークエラー**
   - 接続タイムアウト
   - DNS解決失敗
   - TLS証明書エラー
   - 接続拒否

2. **APIエラー**
   - 認証失敗
   - レート制限
   - エンドポイント不在
   - レスポンス形式エラー

3. **設定エラー**
   - 設定ファイル不在
   - 設定形式エラー
   - 必須項目不在
   - 環境変数エラー

4. **ユーザーエラー**
   - 不正なコマンドライン引数
   - 不正なフィルタ構文
   - 不正なソート条件

5. **システムエラー**
   - ファイルアクセス権限
   - メモリ不足
   - 予期せぬエラー

### 2. コード構造

#### 既存ファイル修正
- `internal/error/handler.go` - エラーハンドラーの拡張
- `cmd/llm-info/main.go` - ヘルプ機能の拡張

#### 新規ファイル作成
- `internal/error/types.go` - エラー種別の定義
- `internal/error/messages.go` - エラーメッセージの定義
- `internal/error/solutions.go` - 解決策の定義
- `cmd/llm-info/help.go` - ヘルプ機能の実装

### 3. 詳細実装

#### internal/error/types.go
```go
package error

import "fmt"

// ErrorType はエラー種別を表す
type ErrorType int

const (
    ErrorTypeUnknown ErrorType = iota
    ErrorTypeNetwork
    ErrorTypeAPI
    ErrorTypeConfig
    ErrorTypeUser
    ErrorTypeSystem
)

// ErrorSeverity はエラーの重大度を表す
type ErrorSeverity int

const (
    SeverityInfo ErrorSeverity = iota
    SeverityWarning
    SeverityError
    SeverityFatal
)

// AppError はアプリケーションエラーを表す
type AppError struct {
    Type        ErrorType
    Severity    ErrorSeverity
    Code        string
    Message     string
    Cause       error
    Context     map[string]interface{}
    Solutions   []string
    HelpURL     string
}

// Error はerrorインターフェースを実装する
func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap はエラーのアンラップをサポートする
func (e *AppError) Unwrap() error {
    return e.Cause
}

// NewAppError は新しいアプリケーションエラーを作成する
func NewAppError(errorType ErrorType, severity ErrorSeverity, code, message string) *AppError {
    return &AppError{
        Type:     errorType,
        Severity: severity,
        Code:     code,
        Message:  message,
        Context:  make(map[string]interface{}),
    }
}

// WithCause は原因エラーを設定する
func (e *AppError) WithCause(cause error) *AppError {
    e.Cause = cause
    return e
}

// WithContext はコンテキスト情報を追加する
func (e *AppError) WithContext(key string, value interface{}) *AppError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    e.Context[key] = value
    return e
}

// WithSolution は解決策を追加する
func (e *AppError) WithSolution(solution string) *AppError {
    e.Solutions = append(e.Solutions, solution)
    return e
}

// WithHelpURL はヘルプURLを設定する
func (e *AppError) WithHelpURL(url string) *AppError {
    e.HelpURL = url
    return e
}
```

#### internal/error/messages.go
```go
package error

import (
    "fmt"
    "net/url"
    "strings"
)

// ErrorMessage はエラーメッセージを管理する
type ErrorMessage struct {
    Type    ErrorType
    Code    string
    Message string
}

var errorMessages = map[ErrorType]map[string]string{
    ErrorTypeNetwork: {
        "connection_timeout":    "接続がタイムアウトしました",
        "dns_resolution_failed": "DNS解決に失敗しました",
        "tls_certificate_error": "TLS証明書エラーが発生しました",
        "connection_refused":    "接続が拒否されました",
        "unknown_host":          "不明なホストです",
    },
    ErrorTypeAPI: {
        "authentication_failed": "認証に失敗しました",
        "rate_limit_exceeded":   "レート制限を超えました",
        "endpoint_not_found":    "エンドポイントが見つかりません",
        "invalid_response":      "無効なレスポンス形式です",
        "server_error":          "サーバーエラーが発生しました",
    },
    ErrorTypeConfig: {
        "config_file_not_found": "設定ファイルが見つかりません",
        "invalid_config_format": "設定ファイルの形式が無効です",
        "missing_required_field": "必須項目が設定されていません",
        "invalid_env_variable":  "環境変数の値が無効です",
    },
    ErrorTypeUser: {
        "invalid_argument":      "無効な引数です",
        "invalid_filter_syntax": "フィルタ構文が無効です",
        "invalid_sort_field":    "無効なソートフィールドです",
        "gateway_not_found":     "指定されたゲートウェイが見つかりません",
    },
    ErrorTypeSystem: {
        "permission_denied":     "ファイルアクセス権限がありません",
        "disk_full":             "ディスク容量が不足しています",
        "memory_insufficient":   "メモリが不足しています",
        "unexpected_error":      "予期せぬエラーが発生しました",
    },
}

// GetErrorMessage はエラーメッセージを取得する
func GetErrorMessage(errorType ErrorType, code string) string {
    if messages, ok := errorMessages[errorType]; ok {
        if message, ok := messages[code]; ok {
            return message
        }
    }
    return "不明なエラーが発生しました"
}

// FormatErrorMessage はエラーメッセージをフォーマットする
func FormatErrorMessage(err *AppError) string {
    var builder strings.Builder
    
    // 基本メッセージ
    builder.WriteString(fmt.Sprintf("❌ %s\n", err.Message))
    
    // コンテキスト情報
    if len(err.Context) > 0 {
        builder.WriteString("\n📋 詳細情報:\n")
        for key, value := range err.Context {
            builder.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
        }
    }
    
    // 解決策
    if len(err.Solutions) > 0 {
        builder.WriteString("\n💡 解決策:\n")
        for i, solution := range err.Solutions {
            builder.WriteString(fmt.Sprintf("   %d. %s\n", i+1, solution))
        }
    }
    
    // ヘルプURL
    if err.HelpURL != "" {
        builder.WriteString(fmt.Sprintf("\n📖 詳細なヘルプ: %s\n", err.HelpURL))
    }
    
    return builder.String()
}

// CreateNetworkError はネットワークエラーを作成する
func CreateNetworkError(code string, url string, cause error) *AppError {
    message := GetErrorMessage(ErrorTypeNetwork, code)
    err := NewAppError(ErrorTypeNetwork, SeverityError, code, message).
        WithCause(cause).
        WithContext("url", url)
    
    // エラーコードごとの解決策を追加
    switch code {
    case "connection_timeout":
        err = err.WithSolution("ネットワーク接続を確認してください").
            WithSolution("ファイアウォール設定を確認してください").
            WithSolution("ゲートウェイURLが正しいか確認してください")
    case "dns_resolution_failed":
        err = err.WithSolution("DNS設定を確認してください").
            WithSolution("ホスト名が正しいか確認してください").
            WithSolution("インターネット接続を確認してください")
    case "tls_certificate_error":
        err = err.WithSolution("サーバーの証明書が有効か確認してください").
            WithSolution("システムの時刻が正しいか確認してください")
    case "connection_refused":
        err = err.WithSolution("ゲートウェイサーバーが起動しているか確認してください").
            WithSolution("ポート番号が正しいか確認してください").
            WithSolution("ファイアウォール設定を確認してください")
    }
    
    return err.WithHelpURL("https://github.com/armaniacs/llm-info/wiki/network-errors")
}

// CreateAPIError はAPIエラーを作成する
func CreateAPIError(code string, statusCode int, url string, cause error) *AppError {
    message := GetErrorMessage(ErrorTypeAPI, code)
    err := NewAppError(ErrorTypeAPI, SeverityError, code, message).
        WithCause(cause).
        WithContext("url", url).
        WithContext("status_code", statusCode)
    
    // エラーコードごとの解決策を追加
    switch code {
    case "authentication_failed":
        err = err.WithSolution("APIキーが正しいか確認してください").
            WithSolution("APIキーの有効期限が切れていないか確認してください").
            WithSolution("APIキーの権限設定を確認してください")
    case "rate_limit_exceeded":
        err = err.WithSolution("しばらく待ってから再試行してください").
            WithSolution("APIプランのレート制限を確認してください").
            WithSolution("並列リクエスト数を減らしてください")
    case "endpoint_not_found":
        err = err.WithSolution("ゲートウェイURLが正しいか確認してください").
            WithSolution("APIバージョンが正しいか確認してください").
            WithSolution("エンドポイントパスを確認してください")
    }
    
    return err.WithHelpURL("https://github.com/armaniacs/llm-info/wiki/api-errors")
}

// CreateConfigError は設定エラーを作成する
func CreateConfigError(code string, configPath string, cause error) *AppError {
    message := GetErrorMessage(ErrorTypeConfig, code)
    err := NewAppError(ErrorTypeConfig, SeverityError, code, message).
        WithCause(cause).
        WithContext("config_path", configPath)
    
    // エラーコードごとの解決策を追加
    switch code {
    case "config_file_not_found":
        err = err.WithSolution("設定ファイルを作成してください: llm-info --init-config").
            WithSolution("設定ファイルパスが正しいか確認してください").
            WithSolution("環境変数LLM_INFO_CONFIG_PATHを確認してください")
    case "invalid_config_format":
        err = err.WithSolution("YAML形式が正しいか確認してください").
            WithSolution("設定ファイルの構文を確認してください").
            WithSolution("例: https://github.com/armaniacs/llm-info/blob/main/configs/example.yaml")
    case "missing_required_field":
        err = err.WithSolution("必須項目（url, api_keyなど）が設定されているか確認してください").
            WithSolution("設定ファイルのテンプレートを確認してください")
    }
    
    return err.WithHelpURL("https://github.com/armaniacs/llm-info/wiki/config-errors")
}

// CreateUserError はユーザーエラーを作成する
func CreateUserError(code string, argument string, cause error) *AppError {
    message := GetErrorMessage(ErrorTypeUser, code)
    err := NewAppError(ErrorTypeUser, SeverityWarning, code, message).
        WithCause(cause).
        WithContext("argument", argument)
    
    // エラーコードごとの解決策を追加
    switch code {
    case "invalid_argument":
        err = err.WithSolution("コマンドライン引数が正しいか確認してください").
            WithSolution("ヘルプを確認してください: llm-info --help")
    case "invalid_filter_syntax":
        err = err.WithSolution("フィルタ構文を確認してください").
            WithSolution("例: --filter \"name:gpt,tokens>1000\"").
            WithSolution("ヘルプを確認してください: llm-info --help filter")
    case "invalid_sort_field":
        err = err.WithSolution("ソートフィールドを確認してください").
            WithSolution("使用可能なフィールド: name, tokens, cost, mode").
            WithSolution("ヘルプを確認してください: llm-info --help sort")
    case "gateway_not_found":
        err = err.WithSolution("ゲートウェイ名が正しいか確認してください").
            WithSolution("利用可能なゲートウェイを確認してください: llm-info --list-gateways")
    }
    
    return err.WithHelpURL("https://github.com/armaniacs/llm-info/wiki/usage")
}
```

#### internal/error/solutions.go
```go
package error

import (
    "fmt"
    "runtime"
    "strings"
)

// SolutionProvider は解決策を提供する
type SolutionProvider struct {
    osInfo string
}

// NewSolutionProvider は新しい解決策プロバイダーを作成する
func NewSolutionProvider() *SolutionProvider {
    return &SolutionProvider{
        osInfo: getOSInfo(),
    }
}

// GetGeneralSolutions は一般的な解決策を返す
func (sp *SolutionProvider) GetGeneralSolutions() []string {
    solutions := []string{
        "最新バージョンにアップデートしてください: llm-info --version",
        "詳細なログを確認してください: llm-info --verbose",
        "設定ファイルを確認してください: llm-info --check-config",
    }
    
    // OS固有の解決策を追加
    if strings.Contains(sp.osInfo, "windows") {
        solutions = append(solutions, "Windowsファイアウォール設定を確認してください")
    } else {
        solutions = append(solutions, "iptables設定を確認してください")
    }
    
    return solutions
}

// GetNetworkSolutions はネットワーク関連の解決策を返す
func (sp *SolutionProvider) GetNetworkSolutions(url string) []string {
    parsedURL, err := url.Parse(url)
    if err != nil {
        return sp.GetGeneralSolutions()
    }
    
    solutions := []string{
        fmt.Sprintf("ホスト %s に到達できるか確認してください", parsedURL.Host),
        "プロキシ設定を確認してください",
        "DNS設定を確認してください",
    }
    
    if parsedURL.Scheme == "https" {
        solutions = append(solutions, "TLS/SSL設定を確認してください")
    }
    
    return solutions
}

// GetConfigSolutions は設定関連の解決策を返す
func (sp *SolutionProvider) GetConfigSolutions(configPath string) []string {
    return []string{
        fmt.Sprintf("設定ファイル %s が存在するか確認してください", configPath),
        "設定ファイルのパーミッションを確認してください",
        "設定ファイルのYAML構文を確認してください",
        "llm-info --init-config で初期設定を作成してください",
    }
}

// getOSInfo はOS情報を取得する
func getOSInfo() string {
    return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
```

#### internal/error/handler.go（拡張）
```go
package error

import (
    "fmt"
    "log"
    "os"
    "runtime/debug"
)

// Handler はエラーハンドラーを表す
type Handler struct {
    verbose       bool
    solutionProvider *SolutionProvider
}

// NewHandler は新しいエラーハンドラーを作成する
func NewHandler(verbose bool) *Handler {
    return &Handler{
        verbose:       verbose,
        solutionProvider: NewSolutionProvider(),
    }
}

// Handle はエラーを処理する
func (h *Handler) Handle(err error) int {
    if err == nil {
        return 0
    }
    
    var appErr *AppError
    if !AsAppError(err, &appErr) {
        // アプリケーションエラーでない場合はラップする
        appErr = NewAppError(ErrorTypeUnknown, SeverityError, "unexpected_error", "予期せぬエラーが発生しました").
            WithCause(err).
            WithSolution("開発者にエラーレポートを送信してください").
            WithHelpURL("https://github.com/armaniacs/llm-info/issues")
    }
    
    // エラーメッセージを表示
    fmt.Fprintln(os.Stderr, FormatErrorMessage(appErr))
    
    // 詳細モードの場合は追加情報を表示
    if h.verbose {
        h.printVerboseInfo(appErr)
    }
    
    // 重大度に応じて終了コードを返す
    switch appErr.Severity {
    case SeverityInfo:
        return 0
    case SeverityWarning:
        return 1
    case SeverityError:
        return 2
    case SeverityFatal:
        return 3
    default:
        return 2
    }
}

// HandleWithFallback はエラーを処理し、フォールバック処理を実行する
func (h *Handler) HandleWithFallback(err error, fallback func() error) int {
    if err == nil {
        return 0
    }
    
    // エラーを処理
    exitCode := h.Handle(err)
    
    // フォールバック処理を実行
    if fallback != nil {
        fmt.Fprintln(os.Stderr, "\n🔄 フォールバック処理を実行します...")
        if fallbackErr := fallback(); fallbackErr != nil {
            fmt.Fprintf(os.Stderr, "❌ フォールバック処理も失敗しました: %v\n", fallbackErr)
            return 3
        }
        fmt.Fprintln(os.Stderr, "✅ フォールバック処理が成功しました")
        return 0
    }
    
    return exitCode
}

// printVerboseInfo は詳細情報を表示する
func (h *Handler) printVerboseInfo(err *AppError) {
    fmt.Fprintln(os.Stderr, "\n🔍 詳細情報:")
    
    // スタックトレース
    if err.Cause != nil {
        fmt.Fprintf(os.Stderr, "原因エラー: %v\n", err.Cause)
    }
    
    // デバッグ情報
    fmt.Fprintf(os.Stderr, "エラータイプ: %v\n", err.Type)
    fmt.Fprintf(os.Stderr, "エラーコード: %s\n", err.Code)
    fmt.Fprintf(os.Stderr, "重大度: %v\n", err.Severity)
    
    // スタックトレース
    if h.verbose {
        fmt.Fprintln(os.Stderr, "\nスタックトレース:")
        debug.PrintStack()
    }
}

// AsAppError はエラーをAppErrorに変換する
func AsAppError(err error, target **AppError) bool {
    if err == nil {
        return false
    }
    
    if appErr, ok := err.(*AppError); ok {
        *target = appErr
        return true
    }
    
    // エラーチェーンをたどってAppErrorを探す
    for {
        if unwrapped := fmt.Unwrap(err); unwrapped != nil {
            if appErr, ok := unwrapped.(*AppError); ok {
                *target = appErr
                return true
            }
            err = unwrapped
            continue
        }
        break
    }
    
    return false
}

// Recover はパニックから回復する
func (h *Handler) Recover() {
    if r := recover(); r != nil {
        err := NewAppError(ErrorTypeSystem, SeverityFatal, "panic", "アプリケーションがクラッシュしました").
            WithCause(fmt.Errorf("panic: %v", r)).
            WithSolution("開発者にバグレポートを送信してください").
            WithHelpURL("https://github.com/armaniacs/llm-info/issues")
        
        log.Printf("Panic recovered: %v\n", r)
        debug.PrintStack()
        
        os.Exit(h.Handle(err))
    }
}
```

#### cmd/llm-info/help.go
```go
package main

import (
    "fmt"
    "os"
    "strings"
    "text/tabwriter"
)

// HelpProvider はヘルプ機能を提供する
type HelpProvider struct {
    version string
}

// NewHelpProvider は新しいヘルププロバイダーを作成する
func NewHelpProvider(version string) *HelpProvider {
    return &HelpProvider{version: version}
}

// ShowGeneralHelp は一般的なヘルプを表示する
func (hp *HelpProvider) ShowGeneralHelp() {
    fmt.Printf(`llm-info - LLMゲートウェイ情報可視化ツール (バージョン: %s)

使用方法:
  llm-info [flags]

フラグ:
`, hp.version)
    
    w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
    fmt.Fprintln(w, "  --url string\tゲートウェイのURL")
    fmt.Fprintln(w, "  --api-key string\tAPIキー")
    fmt.Fprintln(w, "  --gateway string\t使用するゲートウェイ名")
    fmt.Fprintln(w, "  --timeout duration\tリクエストタイムアウト (デフォルト: 10s)")
    fmt.Fprintln(w, "  --format string\t出力形式 (table|json) (デフォルト: table)")
    fmt.Fprintln(w, "  --filter string\tフィルタ条件")
    fmt.Fprintln(w, "  --sort string\tソート条件")
    fmt.Fprintln(w, "  --columns string\t表示するカラム (カンマ区切り)")
    fmt.Fprintln(w, "  --config string\t設定ファイルパス")
    fmt.Fprintln(w, "  --verbose\t詳細なログを表示")
    fmt.Fprintln(w, "  --help\tヘルプを表示")
    fmt.Fprintln(w, "  --version\tバージョンを表示")
    w.Flush()
    
    fmt.Println(`
使用例:
  # 基本使用
  llm-info --url https://api.example.com --api-key your-key
  
  # 設定ファイルを使用
  llm-info --gateway production
  
  # フィルタリングとソート
  llm-info --filter "gpt" --sort "tokens"
  
  # JSON出力
  llm-info --format json
  
詳細なヘルプ:
  llm-info --help filter    # フィルタ構文のヘルプ
  llm-info --help sort      # ソートオプションのヘルプ
  llm-info --help config    # 設定ファイルのヘルプ
  llm-info --help examples  # 使用例のヘルプ
`)
}

// ShowFilterHelp はフィルタ構文のヘルプを表示する
func (hp *HelpProvider) ShowFilterHelp() {
    fmt.Println(`フィルタ構文ヘルプ

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

使用例:
  llm-info --filter "gpt"                           # GPTモデルのみ
  llm-info --filter "name:gpt,tokens>1000"          # GPTでトークン数>1000
  llm-info --filter "exclude:beta,cost<0.01"        # ベータ版除外でコスト<0.01
  llm-info --filter "mode:chat,tokens>4000"         # チャットモードでトークン数>4000
`)
}

// ShowSortHelp はソートオプションのヘルプを表示する
func (hp *HelpProvider) ShowSortHelp() {
    fmt.Println(`ソートオプションヘルプ

基本構文:
  --sort "フィールド"        # 昇順
  --sort "-フィールド"       # 降順

使用可能なフィールド:
  name, model              モデル名
  tokens, max_tokens       最大トークン数
  cost, input_cost         入力コスト
  mode                     モード

使用例:
  llm-info --sort "name"           # 名前の昇順
  llm-info --sort "-tokens"        # トークン数の降順
  llm-info --sort "cost"           # コストの昇順
`)
}

// ShowConfigHelp は設定ファイルのヘルプを表示する
func (hp *HelpProvider) ShowConfigHelp() {
    fmt.Println(`設定ファイルヘルプ

設定ファイルの場所:
  ~/.config/llm-info/llm-info.yaml

設定ファイル形式:
  gateways:
    - name: "production"
      url: "https://api.example.com"
      api_key: "your-api-key"
      timeout: "10s"
    - name: "development"
      url: "https://dev-api.example.com"
      api_key: "dev-api-key"
      timeout: "5s"
  
  default_gateway: "production"
  
  global:
    timeout: "10s"
    output_format: "table"
    sort_by: "name"

コマンド:
  llm-info --init-config     # 設定ファイルのテンプレートを作成
  llm-info --check-config    # 設定ファイルを検証
  llm-info --list-gateways   # 登録済みゲートウェイを一覧表示
`)
}

// ShowExamplesHelp は使用例のヘルプを表示する
func (hp *HelpProvider) ShowExamplesHelp() {
    fmt.Println(`使用例ヘルプ

基本使用例:
  # 直接指定
  llm-info --url https://api.openai.com --api-key sk-xxx
  
  # 設定ファイル使用
  llm-info --gateway production
  
  # 環境変数使用
  export LLM_INFO_URL="https://api.example.com"
  export LLM_INFO_API_KEY="your-key"
  llm-info

フィルタリング例:
  # GPTモデルのみ
  llm-info --filter "gpt"
  
  # 高トークン数モデル
  llm-info --filter "tokens>32000"
  
  # 安価なモデル
  llm-info --filter "cost<0.001"
  
  # 複合条件
  llm-info --filter "name:gpt-4,tokens>8000"

ソート例:
  # トークン数の降順
  llm-info --sort "-tokens"
  
  # コストの昇順
  llm-info --sort "cost"

出力形式例:
  # JSON出力
  llm-info --format json
  
  # 特定カラムのみ
  llm-info --columns "name,tokens"
  
  # スクリプトでの使用
  llm-info --format json | jq '.models[] | select(.max_tokens > 10000)'

CI/CDでの使用例:
  # GitHub Actions
  - name: List models
    env:
      LLM_INFO_URL: ${{ secrets.API_URL }}
      LLM_INFO_API_KEY: ${{ secrets.API_KEY }}
    run: llm-info --format json > models.json
  
  # Docker
  docker run --rm \
    -e LLM_INFO_URL="https://api.example.com" \
    -e LLM_INFO_API_KEY="your-key" \
    llm-info:latest
`)
}

// ShowTopicHelp はトピック別のヘルプを表示する
func (hp *HelpProvider) ShowTopicHelp(topic string) {
    switch strings.ToLower(topic) {
    case "filter":
        hp.ShowFilterHelp()
    case "sort":
        hp.ShowSortHelp()
    case "config":
        hp.ShowConfigHelp()
    case "examples":
        hp.ShowExamplesHelp()
    default:
        fmt.Printf("不明なトピック: %s\n", topic)
        fmt.Println("利用可能なトピック: filter, sort, config, examples")
    }
}
```

### 4. テスト戦略

#### 単体テスト
- `internal/error/types_test.go` - エラー種別のテスト
- `internal/error/messages_test.go` - エラーメッセージのテスト
- `internal/error/handler_test.go` - エラーハンドラーのテスト

#### 統合テスト
- `test/integration/error_handling_test.go` - エラーハンドリングの統合テスト

### 5. 必要なファイルの新規作成・修正

#### 新規作成ファイル
1. `internal/error/types.go`
2. `internal/error/messages.go`
3. `internal/error/solutions.go`
4. `cmd/llm-info/help.go`
5. `internal/error/types_test.go`
6. `internal/error/messages_test.go`
7. `test/integration/error_handling_test.go`

#### 修正ファイル
1. `internal/error/handler.go`
2. `cmd/llm-info/main.go`
3. `internal/error/handler_test.go`

### 6. 受け入れ基準チェックリスト

- [ ] エラー種別ごとの詳細メッセージ
- [ ] 解決策の提示
- [ ] 包括的なヘルプ機能
- [ ] 使用例の提示
- [ ] エラーコードの一貫性
- [ ] 多言語対応の準備
- [ ] 単体テストカバレッジ80%以上
- [ ] 統合テストの実装

### 7. 実装手順

1. **エラー種別の定義**
   - `internal/error/types.go`の作成
   - エラー構造体の定義

2. **エラーメッセージの実装**
   - `internal/error/messages.go`の作成
   - エラーメッセージと解決策の定義

3. **解決策プロバイダーの実装**
   - `internal/error/solutions.go`の作成
   - 動的な解決策生成機能

4. **エラーハンドラーの拡張**
   - `internal/error/handler.go`の修正
   - 詳細なエラー表示機能

5. **ヘルプ機能の実装**
   - `cmd/llm-info/help.go`の作成
   - トピック別ヘルプ機能

6. **コマンドラインインターフェースの修正**
   - `cmd/llm-info/main.go`の修正
   - ヘルプオプションの拡張

7. **テストの実装**
   - 単体テストの作成
   - 統合テストの作成

8. **ドキュメント更新**
   - エラーコード一覧の作成
   - トラブルシューティングガイドの作成

### 8. リスクと対策

#### リスク
1. エラーメッセージの多言語対応
2. 解決策の陳腐化
3. ヘルプ情報のメンテナンスコスト

#### 対策
1. 国際化フレームワークの導入準備
2. 定期的な解決策の見直しプロセス
3. ドキュメント生成の自動化

### 9. 成功指標

- エラー種別の識別率100%
- 解決策の提示率100%
- ユーザー満足度の向上（アンケート等）
- サポート問い合わせの削減率

### 10. 使用例

#### エラーメッセージの例
```bash
$ llm-info --url https://invalid-host
❌ 接続がタイムアウトしました

📋 詳細情報:
   url: https://invalid-host

💡 解決策:
   1. ネットワーク接続を確認してください
   2. ファイアウォール設定を確認してください
   3. ゲートウェイURLが正しいか確認してください

📖 詳細なヘルプ: https://github.com/armaniacs/llm-info/wiki/network-errors
```

#### ヘルプ機能の例
```bash
$ llm-info --help filter
フィルタ構文ヘルプ

基本構文:
  --filter "条件1,条件2,..."

使用可能な条件:
  name:パターン          モデル名でフィルタ（部分一致）
  exclude:パターン       モデル名で除外（部分一致）
  tokens>数値           最大トークン数が指定値より大きい
  ...