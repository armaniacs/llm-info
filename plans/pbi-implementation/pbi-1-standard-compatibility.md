# PBI 1: 標準互換モードと自動フォールバック機能 - 実装計画

## PBI概要と目的

**タイトル**: OpenAI標準互換モードの追加と自動フォールバック機能

**目的**: LiteLLM互換でないゲートウェイにも対応し、利用可能なLLMサービスの幅を広げること。

**ビジネス価値**:
- 互換性向上：より多くのLLMゲートウェイに対応
- ユーザー体験：自動フォールバックにより手動設定が不要
- 柔軟性：様々なゲートウェイ環境で利用可能

## 現状の課題

1. 現在の実装はLiteLLMの`/model/info`エンドポイントにのみ対応
2. OpenAI標準の`/v1/models`エンドポイントをサポートしていない
3. エンドポイントが利用できない場合のフォールバック機能がない
4. ユーザーが手動でエンドポイントを切り替える必要がある

## 実装計画

### 1. APIクライアントの拡張

#### 新規ファイル作成
- `internal/api/endpoints.go` - エンドポイント管理機能
- `internal/api/standard_client.go` - OpenAI標準互換クライアント

#### 既存ファイル修正
- `internal/api/client.go` - フォールバックロジックの追加
- `internal/api/response.go` - レスポンス形式の正規化処理

### 2. コード構造

#### internal/api/endpoints.go
```go
package api

import "fmt"

// EndpointType はAPIエンドポイントの種類を表す
type EndpointType int

const (
    LiteLLM EndpointType = iota
    OpenAIStandard
)

// Endpoint はAPIエンドポイント情報を表す
type Endpoint struct {
    Type EndpointType
    Path string
}

// GetEndpoints は利用可能なエンドポイント一覧を返す
func GetEndpoints(baseURL string) []Endpoint {
    return []Endpoint{
        {Type: LiteLLM, Path: fmt.Sprintf("%s/model/info", baseURL)},
        {Type: OpenAIStandard, Path: fmt.Sprintf("%s/v1/models", baseURL)},
    }
}
```

#### internal/api/standard_client.go
```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
)

// StandardResponse はOpenAI標準APIのレスポンス形式を表す
type StandardResponse struct {
    Object string `json:"object"`
    Data   []struct {
        ID      string `json:"id"`
        Object  string `json:"object"`
        Created int64  `json:"created"`
        OwnedBy string `json:"owned_by"`
    } `json:"data"`
}

// FetchStandardModels はOpenAI標準エンドポイントからモデル情報を取得する
func (c *Client) FetchStandardModels() (*StandardResponse, error) {
    endpoint := fmt.Sprintf("%s/v1/models", c.BaseURL)
    
    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    if c.APIKey != "" {
        req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
    }
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to send request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
    }
    
    var result StandardResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &result, nil
}
```

#### internal/api/client.go（修正）
```go
package api

import (
    "fmt"
    "log"
)

// FetchModelsWithFallback はLiteLLMエンドポイントを試行し、失敗した場合はOpenAI標準エンドポイントにフォールバックする
func (c *Client) FetchModelsWithFallback() ([]ModelInfo, error) {
    // まずLiteLLMエンドポイントを試行
    models, err := c.FetchModels()
    if err == nil {
        return models, nil
    }
    
    log.Printf("LiteLLM endpoint failed, falling back to OpenAI standard endpoint: %v", err)
    
    // OpenAI標準エンドポイントを試行
    standardResp, err := c.FetchStandardModels()
    if err != nil {
        return nil, fmt.Errorf("both endpoints failed: LiteLLM error: %v, Standard error: %w", err, err)
    }
    
    // 標準レスポンスを内部形式に変換
    return c.convertStandardResponse(standardResp), nil
}

// convertStandardResponse はOpenAI標準レスポンスを内部形式に変換する
func (c *Client) convertStandardResponse(resp *StandardResponse) []ModelInfo {
    var models []ModelInfo
    for _, data := range resp.Data {
        models = append(models, ModelInfo{
            ID:        data.ID,
            Name:      data.ID,
            MaxTokens: 0, // 標準APIでは提供されない
            Mode:      "chat", // デフォルト値
            InputCost: 0,     // 標準APIでは提供されない
        })
    }
    return models
}
```

### 3. テスト戦略

#### 単体テスト
- `internal/api/endpoints_test.go` - エンドポイント生成ロジックのテスト
- `internal/api/standard_client_test.go` - OpenAI標準クライアントのテスト
- `internal/api/client_test.go` - フォールバックロジックのテスト

#### 統合テスト
- `test/integration/endpoint_fallback_test.go` - エンドポイント切り替えの統合テスト

#### E2Eテスト
- `test/e2e/standard_compatibility_test.go` - 標準互換ゲートウェイでのテスト

### 4. 必要なファイルの新規作成・修正

#### 新規作成ファイル
1. `internal/api/endpoints.go`
2. `internal/api/standard_client.go`
3. `internal/api/endpoints_test.go`
4. `internal/api/standard_client_test.go`
5. `test/integration/endpoint_fallback_test.go`
6. `test/e2e/standard_compatibility_test.go`

#### 修正ファイル
1. `internal/api/client.go`
2. `internal/api/response.go`
3. `internal/api/client_test.go`

### 5. 受け入れ基準チェックリスト

- [ ] OpenAI標準の`/v1/models`エンドポイントに対応
- [ ] `/model/info`失敗時に自動的に`/v1/models`へフォールバック
- [ ] フォールバック時にユーザーに通知
- [ ] 両モードでのエラーハンドリング
- [ ] 標準APIレスポンスの正規化
- [ ] 単体テストカバレッジ80%以上
- [ ] 統合テストの実装
- [ ] E2Eテストの実装

### 6. 実装手順

1. **エンドポイント管理機能の実装**
   - `internal/api/endpoints.go`の作成
   - エンドポイント種別の定義

2. **OpenAI標準クライアントの実装**
   - `internal/api/standard_client.go`の作成
   - 標準APIレスポンス構造の定義

3. **フォールバックロジックの実装**
   - `internal/api/client.go`の修正
   - フォールバック処理と通知機能

4. **レスポンス正規化の実装**
   - `internal/api/response.go`の修正
   - 標準APIレスポンスの変換処理

5. **テストの実装**
   - 単体テストの作成
   - 統合テストの作成
   - E2Eテストの作成

6. **ドキュメント更新**
   - APIドキュメントの更新
   - 使用例の追加

### 7. リスクと対策

#### リスク
1. OpenAI標準APIとLiteLLM APIのレスポンス形式の差異
2. フォールバック時のパフォーマンス影響
3. エラーハンドリングの複雑化

#### 対策
1. レスポンス正規化レイヤーの実装
2. タイムアウト設定の最適化
3. エラーメッセージの明確化とログ出力

### 8. 成功指標

- OpenAI標準互換ゲートウェイでの動作成功率100%
- フォールバック機能の正常動作率100%
- エンドポイント切り替えのレスポンスタイム2秒以内
- テストカバレッジ80%以上