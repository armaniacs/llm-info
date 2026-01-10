# API通信リファレンス

## 概要

llm-infoのAPI通信層は、LLMゲートウェイとの通信を担当します。LiteLLM互換エンドポイントとOpenAI標準エンドポイントの両方をサポートし、自動フォールバック機能を提供します。

## パッケージ構成

```
internal/api/
├── client.go              # LiteLLMクライアント
├── client_test.go         # クライアントのテスト
├── standard_client.go     # OpenAI標準クライアント
├── standard_client_test.go # 標準クライアントのテスト
├── endpoints.go           # エンドポイント定義
├── endpoints_test.go      # エンドポイントのテスト
└── response.go            # レスポンス構造体
```

## エンドポイント

### 1. LiteLLM互換エンドポイント

**パス**: `/model/info`

**レスポンス形式**:
```json
{
  "models": [
    {
      "id": "gpt-4",
      "max_tokens": 8192,
      "mode": "chat",
      "input_cost": 0.00003
    }
  ]
}
```

**特徴**:
- 詳細なメタデータを提供
- `max_tokens`、`input_cost`、`mode` 等の情報を含む
- LiteLLM固有のエンドポイント

### 2. OpenAI標準エンドポイント

**パス**: `/v1/models`

**レスポンス形式**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1234567890,
      "owned_by": "openai"
    }
  ]
}
```

**特徴**:
- OpenAI標準互換
- 基本情報のみ提供
- ほぼすべてのゲートウェイで実装されている

## クライアント実装

### Client 構造体

```go
type Client struct {
    baseURL string        // ゲートウェイのベースURL
    apiKey  string        // 認証用APIキー
    timeout time.Duration // リクエストタイムアウト
    client  *http.Client  // HTTPクライアント
}
```

**初期化**:
```go
client := api.NewClient(cfg)
```

### 主要メソッド

#### 1. FetchModelsWithFallback()

自動フォールバック機能を持つメイン取得メソッド。

**実装場所**: `internal/api/client.go:120-150`

**処理フロー**:

```
1. OpenAI標準エンドポイントを試行
   ↓
2. 成功の場合
   ├→ 基本情報を内部形式に変換
   ├→ LiteLLMで詳細情報の追加取得を試行
   │   ├→ 成功: 詳細情報を返す
   │   └→ 失敗: 基本情報を返す（エラーなし）
   └→ 返却

3. 失敗の場合
   ├→ 警告メッセージを表示
   ├→ LiteLLMエンドポイントを試行
   │   ├→ 成功: 詳細情報を返す
   │   └→ 失敗: エラーを返す（両方失敗）
   └→ 返却
```

**シグネチャ**:
```go
func (c *Client) FetchModelsWithFallback() (*ModelInfoResponse, error)
```

**返り値**:
- 成功: `*ModelInfoResponse` （モデル情報）
- 失敗: `error` （詳細なエラーメッセージ）

**エラーケース**:
- 両エンドポイント失敗: `fmt.Errorf("both endpoints failed: Standard error: %v, LiteLLM error: %w", standardErr, litellmErr)`

#### 2. GetModelInfo()

LiteLLM互換エンドポイントから情報を取得。

**実装場所**: `internal/api/client.go:33-92`

**シグネチャ**:
```go
func (c *Client) GetModelInfo() (*ModelInfoResponse, error)
```

**処理内容**:
1. `/model/info` へGETリクエスト
2. Bearer認証ヘッダーの設定
3. ステータスコードのチェック
4. JSONレスポンスのパース
5. エラーハンドリング

**エラーハンドリング**:
- ステータスコード非200: 詳細なエラーメッセージを生成
- JSONパースエラー: レスポンスのプレビューを含めたエラー
- 空レスポンス: 専用のエラーメッセージ

#### 3. FetchStandardModels()

OpenAI標準エンドポイントから情報を取得。

**実装場所**: `internal/api/standard_client.go:21-60`

**シグネチャ**:
```go
func (c *Client) FetchStandardModels() (*StandardResponse, error)
```

**処理内容**:
1. `/v1/models` へGETリクエスト
2. Bearer認証ヘッダーの設定
3. ステータスコードのチェック
4. JSONレスポンスのパース

#### 4. convertStandardResponse()

OpenAI標準レスポンスを内部形式に変換。

**実装場所**: `internal/api/standard_client.go:62-74`

**シグネチャ**:
```go
func (c *Client) convertStandardResponse(resp *StandardResponse) *ModelInfoResponse
```

**変換ルール**:
```go
ModelInfo{
    ID:        data.ID,        // そのまま使用
    MaxTokens: 0,              // 標準APIでは提供されない
    Mode:      "chat",         // デフォルト値
    InputCost: 0,              // 標準APIでは提供されない
}
```

## レスポンス構造体

### ModelInfoResponse

LiteLLM互換レスポンスの構造体。

```go
type ModelInfoResponse struct {
    Models []ModelInfo `json:"models"`
}
```

### ModelInfo

モデル情報の構造体。

```go
type ModelInfo struct {
    ID        string  `json:"id"`          // モデルID
    MaxTokens int     `json:"max_tokens"`  // 最大トークン数
    Mode      string  `json:"mode"`        // モード（chat等）
    InputCost float64 `json:"input_cost"`  // 入力コスト
}
```

### StandardResponse

OpenAI標準レスポンスの構造体。

```go
type StandardResponse struct {
    Object string `json:"object"` // "list"
    Data   []struct {
        ID      string `json:"id"`       // モデルID
        Object  string `json:"object"`   // "model"
        Created int64  `json:"created"`  // 作成日時
        OwnedBy string `json:"owned_by"` // 所有者
    } `json:"data"`
}
```

## 認証

### Bearer認証

**ヘッダー形式**:
```
Authorization: Bearer <api-key>
```

**実装**:
```go
if c.apiKey != "" {
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
}
```

**設定方法**:
1. コマンドライン: `--api-key <key>`
2. 環境変数: `LLM_INFO_API_KEY=<key>`
3. 設定ファイル: `api_key: <key>`

## エラーハンドリング

### HTTPステータスコード

APIクライアントは以下のステータスコードを適切に処理します：

| コード | 説明 | デフォルトメッセージ |
|-------|------|-------------------|
| 400 | Bad Request | Invalid parameters or request format |
| 401 | Unauthorized | Invalid or missing API key |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | The requested endpoint does not exist |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server encountered an error |
| 502 | Bad Gateway | Invalid response received |
| 503 | Service Unavailable | Server temporarily unavailable |
| 504 | Gateway Timeout | Server took too long to respond |

**実装**: `internal/api/client.go:94-118` の `getDefaultStatusMessage()`

### エラーメッセージの生成

エラーレスポンスのボディを読み取り、詳細なエラーメッセージを生成：

```go
errorBody, _ := io.ReadAll(resp.Body)
errorMsg := string(errorBody)

if errorMsg == "" {
    errorMsg = getDefaultStatusMessage(resp.StatusCode)
}

return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, errorMsg)
```

### JSONパースエラー

JSONパースに失敗した場合、レスポンスのプレビューを含めます：

```go
preview := string(body)
if len(preview) > 200 {
    preview = preview[:200] + "..."
}
return nil, fmt.Errorf("failed to decode JSON response: %w. Response preview: %s", err, preview)
```

## タイムアウト管理

### デフォルトタイムアウト

**デフォルト値**: 10秒

**設定方法**:
```go
client := &http.Client{
    Timeout: cfg.Timeout,
}
```

### タイムアウトのカスタマイズ

1. コマンドライン: `--timeout 30s`
2. 環境変数: `LLM_INFO_TIMEOUT=30s`
3. 設定ファイル: `timeout: 30s`

## 使用例

### 基本的な使用

```go
// 設定の作成
cfg := config.New("https://api.example.com", "api-key", 10*time.Second)

// クライアントの初期化
client := api.NewClient(cfg)

// モデル情報の取得（自動フォールバック）
response, err := client.FetchModelsWithFallback()
if err != nil {
    // エラーハンドリング
    return err
}

// モデル情報の使用
for _, model := range response.Models {
    fmt.Printf("Model: %s, MaxTokens: %d\n", model.ID, model.MaxTokens)
}
```

### 手動でのエンドポイント選択

```go
// LiteLLMエンドポイントのみ使用
response, err := client.GetModelInfo()

// OpenAI標準エンドポイントのみ使用
standardResp, err := client.FetchStandardModels()
baseModels := client.convertStandardResponse(standardResp)
```

## テスト

### ユニットテスト

**場所**: `internal/api/client_test.go`

**テストケース**:
1. 両エンドポイント成功
2. 標準成功、LiteLLM失敗
3. 標準失敗、LiteLLM成功
4. 両エンドポイント失敗

**モックサーバーの使用**:
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/model/info" {
        // LiteLLMエンドポイントのモック
    } else if r.URL.Path == "/v1/models" {
        // OpenAI標準エンドポイントのモック
    }
}))
defer server.Close()
```

### 統合テスト

**場所**: `test/integration/endpoint_fallback_test.go`

**テストシナリオ**:
- フォールバック動作の検証
- エンドポイント優先度の検証
- エラーハンドリングの検証

## フォールバック機能の詳細

### 設計判断 (2026-01-10)

**変更前の順序**:
1. LiteLLM → 失敗 → OpenAI標準

**変更後の順序**:
1. OpenAI標準 → 成功 → LiteLLM詳細情報追加取得
2. OpenAI標準 → 失敗 → LiteLLM

**変更理由**:
- OpenAI標準エンドポイントの方が広く実装されている
- 最初の試行での成功率が向上
- 警告メッセージの頻度が減少
- 詳細情報は可能な限り取得（二次試行）

### パフォーマンス影響

**標準成功時のリクエスト**:
- 標準エンドポイント: 1回
- LiteLLMエンドポイント: 1回（詳細情報取得試行）
- 合計: 2回のHTTPリクエスト

**レスポンス時間**:
- 最大で約2倍（両方のエンドポイントを試行する場合）
- タイムアウトにより上限は制限される

## 対応ゲートウェイ

### テスト済み

- **OpenRouter** (https://openrouter.ai/api)
  - OpenAI標準エンドポイントのみ実装
  - 基本情報のみ取得可能

### 互換性

**LiteLLM互換**:
- `/model/info` エンドポイントを実装しているゲートウェイ

**OpenAI標準互換**:
- `/v1/models` エンドポイントを実装しているゲートウェイ
- OpenAI API互換のほとんどのゲートウェイ

## 拡張ポイント

### 新しいエンドポイントの追加

1. 新しいクライアントファイルを作成（例: `graphql_client.go`）
2. `FetchModelsWithFallback()` にフォールバックロジックを追加
3. レスポンス構造体を定義
4. 変換関数を実装

### 並行リクエスト

将来的な改善として、標準とLiteLLMを並行リクエストする実装を検討：

```go
// 並行リクエストの例（未実装）
ch := make(chan result, 2)
go func() { ch <- fetchStandard() }()
go func() { ch <- fetchLiteLLM() }()

// 両方の結果を待つ
result1, result2 := <-ch, <-ch
// 詳細情報を優先して返す
```

## 関連ドキュメント

- [アーキテクチャリファレンス](01-architecture.md)
- [エラーハンドリングリファレンス](04-error.md)
- [仕様書](../plans/spec.md)
