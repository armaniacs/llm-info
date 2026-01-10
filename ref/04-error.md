# エラーハンドリングリファレンス

## 概要

llm-infoのエラーハンドリングシステムは、エラーの分類、詳細なメッセージ生成、解決策の提示を行います。

## エラー分類

### ErrorType

```go
type ErrorType int

const (
    ErrorTypeUnknown ErrorType = iota  // 不明なエラー
    ErrorTypeNetwork                   // ネットワーク関連
    ErrorTypeAPI                       // API関連
    ErrorTypeConfig                    // 設定関連
    ErrorTypeUser                      // ユーザー入力関連
    ErrorTypeSystem                    // システム関連
)
```

### ErrorSeverity

```go
type ErrorSeverity int

const (
    SeverityInfo    ErrorSeverity = iota // 情報
    SeverityWarning                      // 警告
    SeverityError                        // エラー
    SeverityFatal                        // 致命的エラー
)
```

## AppError 構造体

```go
type AppError struct {
    Type        ErrorType              // エラータイプ
    Severity    ErrorSeverity          // 重大度
    Code        string                 // エラーコード
    Message     string                 // エラーメッセージ
    OriginalErr error                  // 元のエラー
    Suggestion  string                 // 解決策（旧形式）
    StatusCode  int                    // HTTPステータスコード
    Context     map[string]interface{} // コンテキスト情報
    Solutions   []string               // 解決策リスト
    HelpURL     string                 // ヘルプURL
}
```

### メソッド

#### Error()

エラーメッセージを返します。

```go
func (e *AppError) Error() string
```

**出力例**:
```
invalid_url: URLの形式が不正です (caused by: invalid URL format)
```

#### WithCause()

原因エラーを設定します。

```go
func (e *AppError) WithCause(cause error) *AppError
```

#### WithContext()

コンテキスト情報を追加します。

```go
func (e *AppError) WithContext(key string, value interface{}) *AppError
```

**使用例**:
```go
err.WithContext("url", gatewayURL).
    WithContext("status_code", 404)
```

#### WithSolution()

解決策を追加します。

```go
func (e *AppError) WithSolution(solution string) *AppError
```

**使用例**:
```go
err.WithSolution("URLを確認してください").
    WithSolution("ネットワーク接続を確認してください")
```

#### WithHelpURL()

ヘルプURLを設定します。

```go
func (e *AppError) WithHelpURL(url string) *AppError
```

## エラーの作成

### 基本的な作成

```go
err := error.NewAppError(
    error.ErrorTypeNetwork,
    error.SeverityError,
    "connection_timeout",
    "接続がタイムアウトしました",
)
```

### メソッドチェーン

```go
err := error.NewAppError(
    error.ErrorTypeAPI,
    error.SeverityError,
    "authentication_failed",
    "認証に失敗しました",
).WithCause(originalErr).
  WithContext("url", gatewayURL).
  WithSolution("APIキーを確認してください").
  WithSolution("APIキーに必要な権限があることを確認してください").
  WithHelpURL("https://github.com/armaniacs/llm-info/wiki/authentication")
```

### 便利な関数

**実装場所**: `internal/error/messages.go`

```go
// ネットワークエラーの作成
func CreateNetworkError(code, context string, err error) *AppError

// APIエラーの作成
func CreateAPIError(code string, statusCode int, context string, err error) *AppError

// 設定エラーの作成
func CreateConfigError(code, context string, err error) *AppError

// ユーザーエラーの作成
func CreateUserError(code, context string, err error) *AppError

// システムエラーの作成
func CreateSystemError(code, context string, err error) *AppError
```

**使用例**:
```go
err := errhandler.CreateNetworkError("connection_timeout", gatewayURL, originalErr)
```

## エラーの検出と自動分類

### WrapErrorWithDetection()

エラーメッセージから自動的にエラータイプを検出します。

**実装場所**: `internal/error/messages.go:37-60`

```go
func WrapErrorWithDetection(err error, context string) *AppError
```

**処理フロー**:
```
1. エラーメッセージを解析
   ↓
2. エラータイプを検出
   ├→ Network: timeout, connection refused, DNS等
   ├→ API: 401, 403, 404, 429, 500等
   ├→ Config: config, YAML等
   ├→ User: validation, invalid等
   └→ System: permission, disk等
   ↓
3. 適切なAppErrorを作成
   ↓
4. 解決策とヘルプURLを追加
```

**検出パターン**:

| パターン | エラータイプ | エラーコード |
|---------|------------|------------|
| timeout, deadline exceeded | Network | connection_timeout |
| connection refused | Network | connection_refused |
| no such host, DNS | Network | dns_error |
| 401, unauthorized | API | authentication_failed |
| 403, forbidden | API | authorization_failed |
| 404, not found | API | endpoint_not_found |
| 429, rate limit | API | rate_limit_exceeded |
| 500, 502, 503, 504 | API | server_error |
| config, YAML | Config | config_error |
| validation, invalid | User | validation_error |
| permission denied | System | permission_denied |

### DetectErrorType()

エラーメッセージからエラータイプとコードを検出します。

**実装場所**: `internal/error/messages.go:62-133`

```go
func DetectErrorType(err error) (ErrorType, string)
```

## エラーハンドラー

### Handler 構造体

```go
type Handler struct {
    verbose bool  // 詳細モード
}
```

**初期化**:
```go
handler := error.NewHandler(verbose)
```

### Handle()

エラーを処理して表示し、終了コードを返します。

**実装場所**: `internal/error/handler.go:195-231`

```go
func (h *Handler) Handle(err error) int
```

**処理フロー**:
```
1. エラーがnilの場合は0を返す
   ↓
2. AppErrorでない場合はラップ
   ↓
3. エラーメッセージを標準エラー出力に表示
   ↓
4. 詳細モードの場合は追加情報を表示
   ↓
5. 重大度に応じた終了コードを返す
```

**終了コード**:
| 重大度 | 終了コード |
|--------|----------|
| SeverityInfo | 0 |
| SeverityWarning | 1 |
| SeverityError | 2 |
| SeverityFatal | 3 |

### HandleWithFallback()

エラーを処理し、フォールバック処理を実行します。

**実装場所**: `internal/error/handler.go:233-254`

```go
func (h *Handler) HandleWithFallback(err error, fallback func() error) int
```

**処理フロー**:
```
1. エラーを処理
   ↓
2. フォールバック関数を実行
   ├→ 成功: 0を返す
   └→ 失敗: 3を返す
```

**使用例**:
```go
exitCode := handler.HandleWithFallback(err, func() error {
    return retryOperation()
})
```

### Recover()

パニックから回復します。

**実装場所**: `internal/error/handler.go:304-317`

```go
func (h *Handler) Recover()
```

**使用例**:
```go
defer errorHandler.Recover()
```

## エラーメッセージのフォーマット

### FormatErrorMessage()

エラーメッセージを整形して返します。

**実装場所**: `internal/error/handler.go:319-351`

```go
func FormatErrorMessage(err *AppError) string
```

**出力形式**:
```
❌ <メッセージ>

📋 詳細情報:
   <key>: <value>
   ...

💡 解決策:
   1. <solution1>
   2. <solution2>
   ...

📖 詳細なヘルプ: <helpURL>
```

**出力例**:
```
❌ 接続がタイムアウトしました

📋 詳細情報:
   url: https://api.example.com

💡 解決策:
   1. ネットワーク接続を確認してください
   2. タイムアウト値を増やしてみてください
   3. ゲートウェイが稼働していることを確認してください

📖 詳細なヘルプ: https://github.com/armaniacs/llm-info/wiki/network-errors
```

## 解決策の定義

### GetSolutions()

エラーコードに応じた解決策を返します。

**実装場所**: `internal/error/solutions.go:10-64`

```go
func GetSolutions(code string) []string
```

**定義されている解決策**:

| エラーコード | 解決策 |
|-------------|--------|
| connection_timeout | ネットワーク接続確認、タイムアウト増加、ゲートウェイ稼働確認 |
| connection_refused | URL確認、サービス稼働確認、ファイアウォール確認 |
| dns_error | URL綴り確認、DNS設定確認、インターネット接続確認 |
| authentication_failed | APIキー確認、権限確認、キー有効期限確認 |
| authorization_failed | APIキー権限確認、アカウント状態確認 |
| endpoint_not_found | URL確認、エンドポイント互換性確認 |
| rate_limit_exceeded | 待機後再試行、プラン確認、リクエスト頻度削減 |
| server_error | 待機後再試行、ステータスページ確認、サポート連絡 |
| config_file_not_found | パス確認、ファイル存在確認、サンプル作成 |
| invalid_config_format | YAML構文確認、サンプル参照、ドキュメント参照 |
| missing_required_field | 必須フィールド確認、設定完全性確認 |
| invalid_argument | 引数確認、ヘルプ参照、値形式確認 |
| invalid_filter_syntax | フィルタ構文確認、例参照 |
| invalid_sort_field | フィールド名確認、利用可能フィールド確認 |

## 詳細モード

### printVerboseInfo()

詳細情報を表示します。

**実装場所**: `internal/error/handler.go:256-275`

**表示内容**:
- 原因エラー
- エラータイプ
- エラーコード
- 重大度
- スタックトレース

**出力例**:
```
🔍 詳細情報:
原因エラー: dial tcp: lookup api.example.com: no such host
エラータイプ: 1
エラーコード: dns_error
重大度: 2

スタックトレース:
goroutine 1 [running]:
runtime/debug.Stack()
    /usr/local/go/src/runtime/debug/stack.go:24 +0x65
...
```

## 使用例

### 基本的なエラーハンドリング

```go
// エラーハンドラーの初期化
verbose := os.Getenv("LLM_INFO_DEBUG") != ""
errorHandler := errhandler.NewHandler(verbose)

// エラー発生時
if err != nil {
    appErr := errhandler.WrapErrorWithDetection(err, gatewayURL)
    os.Exit(errorHandler.Handle(appErr))
}
```

### カスタムエラーの作成

```go
err := errhandler.CreateNetworkError("connection_timeout", gatewayURL, originalErr)
err.WithSolution("タイムアウト値を増やしてみてください")
err.WithHelpURL("https://example.com/help")
os.Exit(errorHandler.Handle(err))
```

### フォールバック付きエラーハンドリング

```go
exitCode := errorHandler.HandleWithFallback(err, func() error {
    // フォールバック処理
    return tryAlternativeEndpoint()
})
os.Exit(exitCode)
```

## テスト

### ユニットテスト

**場所**: `internal/error/*_test.go`

**テストケース**:
- エラーの作成とフォーマット
- エラー検出ロジック
- 解決策の取得
- エラーハンドラーの動作

### 統合テスト

**場所**: `test/integration/error_handling_test.go`

**テストシナリオ**:
- ネットワークエラーのハンドリング
- APIエラーのハンドリング
- 設定エラーのハンドリング
- ユーザーエラーのハンドリング
- システムエラーのハンドリング

## エラーコード一覧

### ネットワーク関連

- `connection_timeout`: 接続タイムアウト
- `connection_refused`: 接続拒否
- `dns_error`: DNS解決エラー
- `network_unreachable`: ネットワーク到達不可

### API関連

- `authentication_failed`: 認証失敗
- `authorization_failed`: 認可失敗
- `endpoint_not_found`: エンドポイント未検出
- `rate_limit_exceeded`: レート制限超過
- `server_error`: サーバーエラー

### 設定関連

- `config_file_not_found`: 設定ファイル未検出
- `invalid_config_format`: 設定形式不正
- `missing_required_field`: 必須フィールド欠損

### ユーザー入力関連

- `invalid_argument`: 引数不正
- `invalid_filter_syntax`: フィルタ構文不正
- `invalid_sort_field`: ソートフィールド不正
- `gateway_not_found`: ゲートウェイ未検出

### システム関連

- `permission_denied`: 権限不足
- `disk_full`: ディスク容量不足
- `unexpected_error`: 予期せぬエラー

## ベストプラクティス

### 1. 適切なエラータイプの使用

```go
// 良い例
err := errhandler.CreateNetworkError("connection_timeout", url, originalErr)

// 悪い例
err := errors.New("connection timeout")
```

### 2. コンテキスト情報の追加

```go
err.WithContext("url", gatewayURL).
    WithContext("timeout", timeout.String()).
    WithContext("attempt", attemptNumber)
```

### 3. 解決策の提供

```go
err.WithSolution("ネットワーク接続を確認してください").
    WithSolution("タイムアウト値を増やしてみてください")
```

### 4. ヘルプURLの提供

```go
err.WithHelpURL("https://github.com/armaniacs/llm-info/wiki/network-errors")
```

## 関連ドキュメント

- [アーキテクチャリファレンス](01-architecture.md)
- [API通信リファレンス](02-api.md)
- [トラブルシューティングガイド](../USAGE.md#トラブルシューティング)
