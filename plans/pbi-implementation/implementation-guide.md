# LLM-Info 機能拡張版 - 実装ガイド

## 概要

このガイドはLLM-Info CLIツールの機能拡張版を実装するための包括的な手順とベストプラクティスをまとめたものです。5つのPBI（プロダクトバックログアイテム）を3つのフェーズで実装します。

## プロジェクト概要

### 目標
- LiteLLM互換でないゲートウェイへの対応
- 設定ファイルと環境変数による柔軟な設定管理
- 高度なフィルタリングと表示機能
- 詳細なエラーハンドリングとユーザビリティ向上

### 実装フェーズ
1. **Phase 1: 基本拡張**（1週間）
   - PBI 1: 標準互換モードと自動フォールバック
   - PBI 2: 設定ファイル機能

2. **Phase 2: 環境対応**（1週間）
   - PBI 3: 環境変数サポート
   - PBI 5: エラーハンドリングとユーザビリティ向上

3. **Phase 3: 高度な機能**（1週間）
   - PBI 4: 高度な表示機能

## 全体実装手順

### 事前準備

#### 1. 開発環境のセットアップ
```bash
# リポジトリのクローン
git clone <repository-url>
cd llm-info

# 依存関係のインストール
go mod tidy

# テストの実行
go test ./...

# ビルドの確認
go build -o bin/llm-info ./cmd/llm-info
```

#### 2. ブランチ戦略
```bash
# メインブランチから機能ブランチを作成
git checkout -b feature/enhanced-llm-info

# 各PBIごとにサブブランチを作成
git checkout -b feature/pbi-1-standard-compatibility
git checkout -b feature/pbi-2-config-file
# ...
```

#### 3. 開発ツールの設定
```bash
# 静的解析ツール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# テストカバレッジツール
go install github.com/wadey/gocovmerge@latest

# ドキュメント生成ツール
go install golang.org/x/tools/cmd/godoc@latest
```

### Phase 1: 基本拡張の実装

#### PBI 1: 標準互換モードと自動フォールバック

**実装順序**:
1. エンドポイント管理機能の実装
2. OpenAI標準クライアントの実装
3. フォールバックロジックの実装
4. レスポンス正規化の実装
5. テストの実装

**詳細手順**:
```bash
# 1. エンドポイント管理機能の実装
cat > internal/api/endpoints.go << 'EOF'
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
EOF

# 2. OpenAI標準クライアントの実装
cat > internal/api/standard_client.go << 'EOF'
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
EOF

# 3. 既存クライアントの修正
# internal/api/client.go にフォールバックロジックを追加

# 4. テストの実装
cat > internal/api/endpoints_test.go << 'EOF'
package api

import (
    "testing"
)

func TestGetEndpoints(t *testing.T) {
    baseURL := "https://api.example.com"
    endpoints := GetEndpoints(baseURL)
    
    if len(endpoints) != 2 {
        t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
    }
    
    if endpoints[0].Type != LiteLLM {
        t.Errorf("Expected first endpoint to be LiteLLM, got %v", endpoints[0].Type)
    }
    
    if endpoints[1].Type != OpenAIStandard {
        t.Errorf("Expected second endpoint to be OpenAIStandard, got %v", endpoints[1].Type)
    }
}
EOF
```

#### PBI 2: 設定ファイル機能

**実装順序**:
1. 設定構造体の定義
2. 設定ファイル読み書き機能の実装
3. 設定バリデーションの実装
4. 設定マネージャーの拡張
5. コマンドラインインターフェースの修正
6. テストの実装

**詳細手順**:
```bash
# 1. 設定構造体の定義
cat > internal/config/types.go << 'EOF'
package config

import "time"

// Config はアプリケーション設定全体を表す
type Config struct {
    Gateways       []Gateway `yaml:"gateways"`
    DefaultGateway string    `yaml:"default_gateway"`
    Global         Global    `yaml:"global"`
}

// Gateway は個別のゲートウェイ設定を表す
type Gateway struct {
    Name    string        `yaml:"name"`
    URL     string        `yaml:"url"`
    APIKey  string        `yaml:"api_key"`
    Timeout time.Duration `yaml:"timeout"`
}

// Global はグローバル設定を表す
type Global struct {
    Timeout     time.Duration `yaml:"timeout"`
    OutputFormat string       `yaml:"output_format"`
    SortBy      string       `yaml:"sort_by"`
}
EOF

# 2. 設定ファイル読み書き機能の実装
cat > internal/config/file.go << 'EOF'
package config

import (
    "fmt"
    "os"
    "path/filepath"
    
    "gopkg.in/yaml.v3"
)

// GetConfigPath は設定ファイルのパスを返す
func GetConfigPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(home, ".config", "llm-info", "llm-info.yaml")
}

// LoadConfigFromFile はファイルから設定を読み込む
func LoadConfigFromFile(path string) (*Config, error) {
    if path == "" {
        path = GetConfigPath()
    }
    
    // ファイルが存在しない場合はデフォルト設定を返す
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return getDefaultConfig(), nil
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }
    
    return &config, nil
}
EOF

# 3. 設定バリデーションの実装
cat > internal/config/validator.go << 'EOF'
package config

import (
    "fmt"
    "net/url"
)

// ValidateConfig は設定値を検証する
func ValidateConfig(config *Config) error {
    if len(config.Gateways) == 0 {
        return fmt.Errorf("at least one gateway must be configured")
    }
    
    // ゲートウェイ名の重複チェック
    names := make(map[string]bool)
    for _, gw := range config.Gateways {
        if names[gw.Name] {
            return fmt.Errorf("duplicate gateway name: %s", gw.Name)
        }
        names[gw.Name] = true
        
        if err := validateGateway(&gw); err != nil {
            return fmt.Errorf("gateway %s: %w", gw.Name, err)
        }
    }
    
    return nil
}

// validateGateway は個別のゲートウェイ設定を検証する
func validateGateway(gw *Gateway) error {
    if gw.Name == "" {
        return fmt.Errorf("gateway name cannot be empty")
    }
    
    if gw.URL == "" {
        return fmt.Errorf("gateway URL cannot be empty")
    }
    
    if _, err := url.Parse(gw.URL); err != nil {
        return fmt.Errorf("invalid URL format: %w", err)
    }
    
    return nil
}
EOF
```

### Phase 2: 環境対応の実装

#### PBI 3: 環境変数サポート

**実装順序**:
1. 環境変数読み込み機能の拡張
2. 設定優先順位の実装
3. コマンドラインインターフェースの修正
4. 設定ソース情報の実装
5. テストの実装

#### PBI 5: エラーハンドリングとユーザビリティ向上

**実装順序**:
1. エラー種別の定義
2. エラーメッセージの実装
3. 解決策プロバイダーの実装
4. エラーハンドラーの拡張
5. ヘルプ機能の実装
6. テストの実装

### Phase 3: 高度な機能の実装

#### PBI 4: 高度な表示機能

**実装順序**:
1. フィルタリング機能の実装
2. ソート機能の実装
3. カラム管理機能の実装
4. テーブル表示機能の拡張
5. JSON出力機能の拡張
6. コマンドラインインターフェースの修正
7. テストの実装

## ベストプラクティス

### 1. コーディング規約

#### Goのコーディング規約
```bash
# コードフォーマット
go fmt ./...

# 静的解析
golangci-lint run

# インポートの整理
goimports -w .
```

#### コード構造のガイドライン
- パッケージは単一責任の原則に従う
- インターフェースは小さく、具体的にする
- エラーハンドリングは明示的に行う
- テストは各パッケージに配置する

### 2. テスト戦略

#### テストピラミッド
```
    E2Eテスト (10%)
   ─────────────────
  統合テスト (20%)
 ───────────────────────
単体テスト (70%)
```

#### テストカバレッジ目標
- 単体テストカバレッジ: 80%以上
- 統合テスト: 主要なユースケースをカバー
- E2Eテスト: クリティカルパスをカバー

#### テスト実行コマンド
```bash
# 単体テスト実行
go test ./...

# カバレッジレポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 統合テスト実行
go test -tags=integration ./test/integration/...

# E2Eテスト実行
go test -tags=e2e ./test/e2e/...
```

### 3. バージョン管理

#### コミットメッセージ規約
```
<type>(<scope>): <subject>

<body>

<footer>
```

**タイプ**:
- feat: 新機能
- fix: バグ修正
- docs: ドキュメント更新
- style: コードスタイル修正
- refactor: リファクタリング
- test: テスト関連
- chore: ビルドプロセスや補助ツールの変更

**例**:
```
feat(api): add OpenAI standard endpoint support

Implement fallback mechanism from LiteLLM to OpenAI standard
endpoint when the former is unavailable.

Closes #123
```

#### ブランチ命名規約
- `feature/<feature-name>`: 新機能開発
- `bugfix/<bug-description>`: バグ修正
- `hotfix/<hotfix-description>`: 緊急修正
- `release/<version>`: リリース準備

### 4. CI/CDパイプライン

#### GitHub Actionsの設定例
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Build
      run: go build -o bin/llm-info ./cmd/llm-info
```

### 5. ドキュメント管理

#### ドキュメント構造
```
docs/
├── api/              # APIドキュメント
├── user-guide/       # ユーザーガイド
├── developer-guide/  # 開発者ガイド
├── examples/         # 使用例
└── troubleshooting/  # トラブルシューティング
```

#### ドキュメント生成
```bash
# Goドキュメント生成
godoc -http=:6060

# MarkdownからHTML生成
pandoc README.md -o README.html

# APIドキュメント生成
swag init -g cmd/llm-info/main.go
```

### 6. パフォーマンス最適化

#### プロファイリング
```bash
# CPUプロファイリング
go test -cpuprofile=cpu.prof -bench=.

# メモリプロファイリング
go test -memprofile=mem.prof -bench=.

# プロファイル結果の確認
go tool pprof cpu.prof
go tool pprof mem.prof
```

#### ベンチマークテスト
```go
func BenchmarkFilterModels(b *testing.B) {
    models := generateTestModels(1000)
    criteria := &FilterCriteria{NamePattern: "gpt"}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Filter(models, criteria)
    }
}
```

### 7. セキュリティ考慮事項

#### セキュリティチェックリスト
- [ ] APIキーの適切な扱い
- [ ] 入力値のバリデーション
- [ ] SQLインジェクション対策
- [ ] XSS対策
- [ ] 依存関係の脆弱性チェック

#### セキュリティスキャン
```bash
# 依存関係の脆弱性チェック
go list -json -m all | nancy sleuth

# 静的セキュリティ解析
gosec ./...
```

## 品質保証

### 1. コードレビュープロセス

#### レビューチェックリスト
- [ ] コードが機能要件を満たしているか
- [ ] テストが十分に記述されているか
- [ ] エラーハンドリングが適切か
- [ ] パフォーマンス上の問題がないか
- [ ] セキュリティ上の問題がないか
- [ ] ドキュメントが更新されているか

### 2. リリースプロセス

#### リリースチェックリスト
- [ ] すべてのテストが通過しているか
- [ ] コードカバレッジが基準を満たしているか
- [ ] パフォーマンスが基準を満たしているか
- [ ] セキュリティスキャンで問題がないか
- [ ] ドキュメントが更新されているか
- [ ] バージョン番号が適切に更新されているか
- [ ] リリースノートが作成されているか

#### リリースコマンド
```bash
# バージョンタグの作成
git tag -a v1.0.0 -m "Release version 1.0.0"

# タグのプッシュ
git push origin v1.0.0

# リリースビルド
go build -ldflags "-X main.version=v1.0.0" -o bin/llm-info ./cmd/llm-info

# リリースアーカイブの作成
tar -czf llm-info-v1.0.0-darwin-amd64.tar.gz llm-info
```

## トラブルシューティング

### 1. よくある問題

#### ビルドエラー
```bash
# 問題: 依存関係の解決に失敗する
# 解決策:
go mod tidy
go mod download

# 問題: インポートエラー
# 解決策:
goimports -w .
```

#### テストエラー
```bash
# 問題: テストがタイムアウトする
# 解決策:
go test -timeout=30s ./...

# 問題: テストが並列実行で失敗する
# 解決策:
go test -p=1 ./...
```

### 2. デバッグ手法

#### ログ出力
```go
// デバッグログの出力
log.Printf("Debug: %+v", data)

// 構造化ログ
logger.Debug("Processing request",
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
)
```

#### デバッガの使用
```bash
# Delveデバッガのインストール
go install github.com/go-delve/delve/cmd/dlv@latest

# デバッグ実行
dlv debug ./cmd/llm-info
```

## まとめ

この実装ガイドに従うことで、LLM-Info CLIツールの機能拡張版を体系的に実装できます。各フェーズで品質を確保しながら、段階的に機能を追加していくことが重要です。

### 成功の鍵
1. **段階的な実装**: フェーズごとに機能を追加し、品質を確保する
2. **テスト重視**: 十分なテストカバレッジを維持する
3. **ドキュメント整備**: 適切なドキュメントを維持する
4. **コードレビュー**: チームでのコードレビューを徹底する
5. **継続的改善**: フィードバックを反映し、継続的に改善する

このガイドを参考に、高品質なLLM-Info CLIツールの機能拡張版を実装してください。