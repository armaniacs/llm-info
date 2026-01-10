# LLM-Info MVP 開発ガイド

## 開発アプローチ

### Outside-In開発

このプロジェクトでは「Outside-In開発」アプローチを採用します。ユーザーインターフェース（CLI）から開発を始め、必要な機能を内側に向かって実装していきます。

1. **CLIインターフェース**から始める
2. **必要なデータ構造**を定義する
3. **API通信層**を実装する
4. **表示機能**を完成させる

### Red-Green-Refactorサイクル

各機能をTDD（テスト駆動開発）で実装します：

1. **Red**: 失敗するテストを書く
2. **Green**: テストを通過させる最小限のコードを書く
3. **Refactor**: コードを改善する

## 詳細実装ガイド

### Phase 1: 基本構造

#### 1.1 プロジェクト初期化

```bash
# プロジェクトディレクトリで実行
go mod init github.com/your-org/llm-info

# 必要な依存関係を追加
go get github.com/olekukonko/tablewriter

# ディレクトリ構造を作成
mkdir -p cmd/llm-info
mkdir -p internal/{api,model,config,ui}
mkdir -p pkg/config
mkdir -p test/{integration,e2e}
```

#### 1.2 エントリーポイントの実装

**ファイル**: `cmd/llm-info/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/your-org/llm-info/internal/api"
	"github.com/your-org/llm-info/internal/config"
	"github.com/your-org/llm-info/internal/model"
	"github.com/your-org/llm-info/internal/ui"
)

func main() {
	// コマンドライン引数の定義
	var (
		url     = flag.String("url", "", "Base URL of the LLM gateway (required)")
		apiKey  = flag.String("api-key", "", "API key for authentication")
		timeout = flag.Duration("timeout", 10*time.Second, "Request timeout")
	)
	
	flag.Parse()
	
	// 必須引数のチェック
	if *url == "" {
		fmt.Fprintf(os.Stderr, "Error: --url is required\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// 設定の作成
	cfg := config.New(*url, *apiKey, *timeout)
	
	// APIクライアントの作成
	client := api.NewClient(cfg)
	
	// モデル情報の取得
	modelInfos, err := client.GetModelInfo()
	if err != nil {
		log.Fatalf("Failed to get model info: %v", err)
	}
	
	// データモデルへの変換
	models := model.FromAPIResponse(modelInfos.Models)
	
	// テーブル表示
	ui.RenderTable(models)
}
```

#### 1.3 設定管理の実装

**ファイル**: `internal/config/config.go`

```go
package config

import "time"

// Config はアプリケーション設定を保持します
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// New は新しい設定を作成します
func New(baseURL, apiKey string, timeout time.Duration) *Config {
	return &Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: timeout,
	}
}
```

### Phase 2: API通信

#### 2.1 APIクライアントの実装

**ファイル**: `internal/api/client.go`

```go
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

// Client はAPIクライアントです
type Client struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// NewClient は新しいAPIクライアントを作成します
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		timeout: cfg.Timeout,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// GetModelInfo はモデル情報を取得します
func (c *Client) GetModelInfo() (*ModelInfoResponse, error) {
	url := fmt.Sprintf("%s/model/info", c.baseURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 認証ヘッダーの設定
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
	req.Header.Set("Content-Type", "application/json")
	
	// リクエスト送信
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// ステータスコードのチェック
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	// レスポンスのパース
	var response ModelInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &response, nil
}
```

#### 2.2 APIレスポンスの定義

**ファイル**: `internal/api/response.go`

```go
package api

// ModelInfoResponse はAPIレスポンスの構造体です
type ModelInfoResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelInfo は個別のモデル情報です
type ModelInfo struct {
	ID        string  `json:"id"`
	MaxTokens int     `json:"max_tokens"`
	Mode      string  `json:"mode"`
	InputCost float64 `json:"input_cost"`
}
```

### Phase 3: データ処理と表示

#### 3.1 データモデルの実装

**ファイル**: `internal/model/model.go`

```go
package model

import (
	"github.com/your-org/llm-info/internal/api"
)

// Model はアプリケーション内のモデルデータです
type Model struct {
	Name      string
	MaxTokens int
	Mode      string
	InputCost float64
}

// FromAPIResponse はAPIレスポンスをアプリケーションモデルに変換します
func FromAPIResponse(apiModels []api.ModelInfo) []Model {
	models := make([]Model, len(apiModels))
	for i, apiModel := range apiModels {
		models[i] = Model{
			Name:      apiModel.ID,
			MaxTokens: apiModel.MaxTokens,
			Mode:      apiModel.Mode,
			InputCost: apiModel.InputCost,
		}
	}
	return models
}
```

#### 3.2 テーブル表示の実装

**ファイル**: `internal/ui/table.go`

```go
package ui

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/your-org/llm-info/internal/model"
)

// RenderTable はモデル情報をテーブル形式で表示します
func RenderTable(models []model.Model) {
	if len(models) == 0 {
		fmt.Println("No models found.")
		return
	}
	
	// 利用可能なデータをチェック
	var hasMaxTokens, hasMode, hasInputCost bool
	for _, model := range models {
		if model.MaxTokens > 0 {
			hasMaxTokens = true
		}
		if model.Mode != "" {
			hasMode = true
		}
		if model.InputCost > 0 {
			hasInputCost = true
		}
	}
	
	// ヘッダーを動的に構成
	headers := []string{"Model Name"}
	if hasMaxTokens {
		headers = append(headers, "Max Tokens")
	}
	if hasMode {
		headers = append(headers, "Mode")
	}
	if hasInputCost {
		headers = append(headers, "Input Cost")
	}
	
	// テーブルの設定
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	
	// データ行を動的に構成
	for _, model := range models {
		var row []string
		row = append(row, model.Name)
		
		if hasMaxTokens {
			row = append(row, strconv.Itoa(model.MaxTokens))
		}
		if hasMode {
			row = append(row, model.Mode)
		}
		if hasInputCost {
			row = append(row, fmt.Sprintf("%.6f", model.InputCost))
		}
		
		table.Append(row)
	}
	
	table.Render()
}
```

### Phase 4: テスト実装

#### 4.1 APIクライアントのテスト

**ファイル**: `internal/api/client_test.go`

```go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/llm-info/internal/config"
)

func TestClient_GetModelInfo(t *testing.T) {
	// モックサーバーの作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストの検証
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/model/info" {
			t.Errorf("Expected /model/info path, got %s", r.URL.Path)
		}
		
		// レスポンスの返信
		response := ModelInfoResponse{
			Models: []ModelInfo{
				{
					ID:        "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// テスト用クライアントの作成
	cfg := config.New(server.URL, "test-api-key", 5*time.Second)
	client := NewClient(cfg)
	
	// テスト実行
	result, err := client.GetModelInfo()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if len(result.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(result.Models))
	}
	
	model := result.Models[0]
	if model.ID != "gpt-4" {
		t.Errorf("Expected model ID 'gpt-4', got '%s'", model.ID)
	}
	if model.MaxTokens != 8192 {
		t.Errorf("Expected max tokens 8192, got %d", model.MaxTokens)
	}
}

func TestClient_GetModelInfo_APIError(t *testing.T) {
	// エラーを返すモックサーバー
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	
	cfg := config.New(server.URL, "test-api-key", 5*time.Second)
	client := NewClient(cfg)
	
	_, err := client.GetModelInfo()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
```

#### 4.2 データモデルのテスト

**ファイル**: `internal/model/model_test.go`

```go
package model

import (
	"testing"

	"github.com/your-org/llm-info/internal/api"
)

func TestFromAPIResponse(t *testing.T) {
	apiModels := []api.ModelInfo{
		{
			ID:        "gpt-4",
			MaxTokens: 8192,
			Mode:      "chat",
			InputCost: 0.00003,
		},
		{
			ID:        "claude-3",
			MaxTokens: 200000,
			Mode:      "chat",
			InputCost: 0.000015,
		},
	}
	
	models := FromAPIResponse(apiModels)
	
	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}
	
	if models[0].Name != "gpt-4" {
		t.Errorf("Expected first model name 'gpt-4', got '%s'", models[0].Name)
	}
	
	if models[1].MaxTokens != 200000 {
		t.Errorf("Expected second model max tokens 200000, got %d", models[1].MaxTokens)
	}
}
```

#### 4.3 統合テスト

**ファイル**: `test/integration/integration_test.go`

```go
package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/your-org/llm-info/internal/api"
	"github.com/your-org/llm-info/internal/config"
	"github.com/your-org/llm-info/internal/model"
	"github.com/your-org/llm-info/internal/ui"
)

func TestEndToEndFlow(t *testing.T) {
	// モックサーバーのセットアップ
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := api.ModelInfoResponse{
			Models: []api.ModelInfo{
				{
					ID:        "gpt-4",
					MaxTokens: 8192,
					Mode:      "chat",
					InputCost: 0.00003,
				},
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// 設定とクライアントの作成
	cfg := config.New(server.URL, "test-key", 5*time.Second)
	client := api.NewClient(cfg)
	
	// API呼び出し
	apiResponse, err := client.GetModelInfo()
	if err != nil {
		t.Fatalf("Failed to get model info: %v", err)
	}
	
	// データ変換
	models := model.FromAPIResponse(apiResponse.Models)
	if len(models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(models))
	}
	
	// 出力のキャプチャ
	var buf bytes.Buffer
	oldStdout := os.Stdout
	os.Stdout = &buf
	defer func() { os.Stdout = oldStdout }()
	
	// テーブル表示
	ui.RenderTable(models)
	
	// 出力の検証
	output := buf.String()
	if !strings.Contains(output, "gpt-4") {
		t.Errorf("Expected output to contain 'gpt-4', got: %s", output)
	}
	if !strings.Contains(output, "8192") {
		t.Errorf("Expected output to contain '8192', got: %s", output)
	}
}
```

## ビルドと実行

### Makefile

```makefile
.PHONY: build test clean lint install run-example

BINARY_NAME=llm-info
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/llm-info/main.go

test:
	go test -v -cover ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html

lint:
	golangci-lint run

install:
	go install cmd/llm-info/main.go

run-example:
	$(BUILD_DIR)/$(BINARY_NAME) --url https://gateway.aipf-dev.sakuraha.jp/v1

dev:
	go run cmd/llm-info/main.go --url https://gateway.aipf-dev.sakuraha.jp/v1
```

### 実行例

```bash
# ビルド
make build

# 実行
./bin/llm-info --url https://gateway.aipf-dev.sakuraha.jp/v1 --api-key your-api-key

# 開発モードで実行
make dev

# テスト実行
make test

# カバレッジ付きテスト
make test-coverage
```

## トラブルシューティング

### よくある問題

1. **インポートエラー**: `go.mod`ファイルのモジュール名を確認
2. **API接続エラー**: タイムアウト値やURLを確認
3. **JSONパースエラー**: APIレスポンスの形式を確認
4. **テーブル表示が崩れる**: `tablewriter`の設定を確認

### デバッグ手法

1. **ログ出力**: `fmt.Printf`や`log.Printf`でデバッグ情報を出力
2. **HTTPデバッグ**: `httputil.DumpRequest`でリクエスト内容を確認
3. **テストデバッグ**: `t.Logf`でテスト中の情報を出力

## 次のステップ

1. **Codeモードへの切り替え**: このガイドを元に実装を開始
2. **Phase 1から着手**: プロジェクト構造のセットアップ
3. **テスト駆動開発**: 各機能をRed-Green-Refactorサイクルで実装
4. **定期的なレビュー**: 各フェーズ完了時にコードレビューを実施

---

## 付録: 開発ベストプラクティス

### コーディング規約

1. **パッケージ名**: 小文字で短い名前を使用
2. **関数名**: キャメルケース、エクスポートする場合は大文字で開始
3. **エラーハンドリング**: 常にエラーをチェックし、適切に処理
4. **コメント**: エクスポートする関数には`godoc`形式のコメントを記述

### Gitワークフロー

1. **ブランチ戦略**: `feature/機能名`でブランチを作成
2. **コミットメッセージ**: `type: description`形式で記述
3. **プルリクエスト**: レビューを必須化

### パフォーマンス考慮事項

1. **HTTPクライアント**: 再利用可能なクライアントを使用
2. **JSONパース**: 効率的なパース処理を実装
3. **メモリ使用**: 大量データを扱う際はストリーミングを検討