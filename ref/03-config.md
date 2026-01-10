# 設定管理リファレンス

## 概要

llm-infoの設定管理システムは、複数のソースからの設定を統合し、優先順位に基づいて最終的な設定を解決します。

## 設定ソース

### 優先順位 (高→低)

1. **コマンドライン引数** (最優先)
2. **環境変数**
3. **設定ファイル**
4. **デフォルト値** (最下位)

## パッケージ構成

```
internal/config/
├── manager.go          # 設定マネージャー（統合管理）
├── manager_test.go     # マネージャーのテスト
├── file.go            # YAML設定ファイルの読み書き
├── file_test.go       # ファイル読み込みのテスト
├── env.go             # 環境変数からの読み込み
├── env_test.go        # 環境変数のテスト
├── validator.go       # 設定の検証
├── validator_test.go  # 検証ロジックのテスト
└── config.go          # 基本構造体

pkg/config/
└── config.go          # 公開インターフェース
```

## 設定マネージャー

### Manager 構造体

```go
type Manager struct {
    appConfig  *config.AppConfig   // アプリケーション設定
    fileConfig *config.FileConfig  // ファイル設定（旧形式）
    newConfig  *config.Config      // ファイル設定（新形式）
    path       string              // 設定ファイルパス
}
```

**初期化**:
```go
manager := config.NewManager(configPath)
// または
manager := config.NewManagerWithDefaults()
```

### 主要メソッド

#### 1. Load()

設定ファイルを読み込みます。

**実装場所**: `internal/config/manager.go:58-92`

**処理フロー**:
```
1. 新形式の設定ファイルを試行
   ↓
2. 失敗の場合
   ├→ 旧形式（レガシー）を試行
   ├→ 旧形式も失敗 → エラー
   └→ 成功 → 旧形式から新形式に変換
   ↓
3. 設定の検証
   ↓
4. マネージャーに保存
```

**シグネチャ**:
```go
func (m *Manager) Load() error
```

#### 2. ResolveConfig()

すべての設定ソースから最終的な設定を解決します。

**実装場所**: `internal/config/manager.go:121-153`

**処理フロー**:
```
1. デフォルト値を設定
   ↓
2. 設定ファイルから適用
   ↓
3. 環境変数から適用
   ↓
4. コマンドライン引数から適用
   ↓
5. 最終的な設定の検証
   ↓
6. ResolvedConfig を返す
```

**シグネチャ**:
```go
func (m *Manager) ResolveConfig(cliArgs *CLIArgs) (*ResolvedConfig, error)
```

**返り値**:
```go
type ResolvedConfig struct {
    Gateway      *config.GatewayConfig // ゲートウェイ設定
    OutputFormat string                // 出力形式
    SortBy       string                // ソート項目
    Filter       string                // フィルタ条件
    Columns      string                // 表示列
    LogLevel     string                // ログレベル
    UserAgent    string                // ユーザーエージェント
    Sources      map[string]ConfigSource // 設定ソース情報
}
```

#### 3. GetConfigSourceInfo()

設定のソース情報を返します。

**実装場所**: `internal/config/manager.go:302-322`

**シグネチャ**:
```go
func (m *Manager) GetConfigSourceInfo(resolved *ResolvedConfig) string
```

**出力例**:
```
Configuration sources:
  gateway: command line
  output_format: config file
  sort_by: default
```

## 設定ファイル

### デフォルトパス

```
~/.config/llm-info/llm-info.yaml
```

### ファイル形式

**新形式** (推奨):
```yaml
# ゲートウェイ設定
gateways:
  - name: "production"
    url: "https://api.example.com"
    api_key: "your-production-api-key"
    timeout: "10s"
    description: "本番環境ゲートウェイ"

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
  columns: "name,max_tokens,mode,input_cost"
  verbose: false
```

**旧形式** (後方互換性):
```yaml
gateways:
  - name: "default"
    url: "https://api.example.com/v1"
    api_key: "your-api-key-here"
    timeout: "10s"

default_gateway: "default"

common:
  timeout: "10s"
  output:
    format: "table"
```

### 設定ファイルの作成

**コマンド**:
```bash
llm-info --init-config
```

**実装**:
```go
manager.CreateExampleConfig()
```

### 設定ファイルの検証

**コマンド**:
```bash
llm-info --check-config
```

**実装**:
`cmd/llm-info/main.go:92-98` の `validateConfigFile()`

## 環境変数

### 対応環境変数

| 環境変数 | 説明 | デフォルト値 |
|----------|------|-------------|
| `LLM_INFO_URL` | ゲートウェイのベースURL | - |
| `LLM_INFO_API_KEY` | 認証用APIキー | - |
| `LLM_INFO_TIMEOUT` | リクエストタイムアウト | 10s |
| `LLM_INFO_DEFAULT_GATEWAY` | デフォルトゲートウェイ名 | default |
| `LLM_INFO_OUTPUT_FORMAT` | 出力形式 (table, json) | table |
| `LLM_INFO_SORT_BY` | ソート項目 | name |
| `LLM_INFO_FILTER` | フィルタ条件 | - |
| `LLM_INFO_COLUMNS` | 表示列 | - |
| `LLM_INFO_CONFIG_PATH` | 設定ファイルのパス | ~/.config/llm-info/llm-info.yaml |
| `LLM_INFO_VERBOSE` | 詳細ログを有効にする | false |
| `LLM_INFO_DEBUG` | デバッグモードを有効にする | false |
| `LLM_INFO_USER_AGENT` | ユーザーエージェント | llm-info/1.0.0 |

### 環境変数の読み込み

**実装場所**: `internal/config/env.go`

**主要関数**:
```go
func LoadEnvConfig() *EnvConfig
```

**処理**:
```go
url := os.Getenv("LLM_INFO_URL")
apiKey := os.Getenv("LLM_INFO_API_KEY")
timeout, _ := time.ParseDuration(os.Getenv("LLM_INFO_TIMEOUT"))
```

### 使用例

```bash
# 基本設定
export LLM_INFO_URL="https://api.example.com"
export LLM_INFO_API_KEY="your-api-key"
export LLM_INFO_TIMEOUT="15s"

# 出力設定
export LLM_INFO_OUTPUT_FORMAT="json"
export LLM_INFO_SORT_BY="max_tokens"
export LLM_INFO_FILTER="gpt"

# 実行
llm-info
```

## コマンドライン引数

### CLIArgs 構造体

```go
type CLIArgs struct {
    URL          string        // ゲートウェイURL
    APIKey       string        // APIキー
    Timeout      time.Duration // タイムアウト
    Gateway      string        // ゲートウェイ名
    OutputFormat string        // 出力形式
    SortBy       string        // ソート項目
    Filter       string        // フィルタ条件
    Columns      string        // 表示列
}
```

### 引数定義

**実装場所**: `cmd/llm-info/main.go:31-48`

```go
var (
    url          = flag.String("url", "", "Base URL of the LLM gateway")
    apiKey       = flag.String("api-key", "", "API key for authentication")
    timeout      = flag.Duration("timeout", 10*time.Second, "Request timeout")
    configFile   = flag.String("config", "", "Path to config file")
    gateway      = flag.String("gateway", "", "Gateway name to use")
    outputFormat = flag.String("format", "table", "Output format (table, json)")
    sortBy       = flag.String("sort", "", "Sort models by field")
    filter       = flag.String("filter", "", "Filter models")
    columns      = flag.String("columns", "", "Specify columns to display")
    // ... その他オプション
)
```

### 使用例

```bash
# 直接URL指定
llm-info --url https://api.example.com --api-key sk-1234567890

# ゲートウェイ名指定
llm-info --gateway production

# 出力形式とフィルタ
llm-info --url https://api.example.com --format json --filter "name:gpt"

# タイムアウト指定
llm-info --url https://api.example.com --timeout 30s
```

## ゲートウェイ設定

### GatewayConfig 構造体

```go
type GatewayConfig struct {
    Name    string        // ゲートウェイ名
    URL     string        // ベースURL
    APIKey  string        // APIキー
    Timeout time.Duration // タイムアウト
}
```

### ゲートウェイの取得

**実装場所**: `internal/config/manager.go:435-472`

```go
func (m *Manager) GetGatewayConfig(name string) (*config.GatewayConfig, error)
```

**処理**:
1. 新形式の設定があれば使用
2. なければ旧形式から取得
3. ゲートウェイ名が空の場合はデフォルトゲートウェイを使用
4. 指定されたゲートウェイを検索して返す

### ゲートウェイ一覧

**実装場所**: `internal/config/manager.go:474-500`

```go
func (m *Manager) ListGateways() []string
```

**使用例**:
```bash
llm-info --list-gateways
```

**出力例**:
```
利用可能なゲートウェイ:
  - production
    URL: https://api.example.com
    タイムアウト: 10s

  - development
    URL: https://dev-api.example.com
    タイムアウト: 5s
```

## 設定の検証

### Validator

**実装場所**: `internal/config/validator.go`

**主要関数**:
```go
func ValidateConfig(cfg *config.Config) error
func ValidateLegacyConfig(cfg *config.FileConfig) error
```

### 検証ルール

1. **必須フィールドのチェック**
   - ゲートウェイURLが必須
   - ゲートウェイ設定が存在すること

2. **値の妥当性チェック**
   - タイムアウトが正の値
   - 出力形式が有効な値 (table, json)
   - URLスキームが http/https

3. **ゲートウェイ設定のチェック**
   - 重複するゲートウェイ名がないこと
   - デフォルトゲートウェイが存在すること

**実装例**:
```go
func (m *Manager) validateResolvedConfig(resolved *ResolvedConfig) error {
    if resolved.Gateway == nil {
        return fmt.Errorf("no gateway configuration found")
    }

    if resolved.Gateway.URL == "" {
        return fmt.Errorf("gateway URL is required")
    }

    return nil
}
```

## 設定ソースの追跡

### ConfigSource 型

```go
type ConfigSource int

const (
    SourceDefault ConfigSource = iota
    SourceFile
    SourceEnv
    SourceCLI
)
```

### ソース情報の記録

**実装**:
```go
resolved.Sources["gateway"] = SourceCLI
resolved.Sources["output_format"] = SourceFile
resolved.Sources["sort_by"] = SourceDefault
```

### ソース情報の表示

**コマンド**:
```bash
llm-info --show-sources
```

**出力例**:
```
Configuration sources:
  gateway: command line
  output_format: config file
  sort_by: default
  filter: environment variable
```

## 設定の適用順序

### applyDefaults()

**実装場所**: `internal/config/manager.go:155-162`

**デフォルト値**:
```go
resolved.OutputFormat = "table"
resolved.SortBy = "name"
```

### applyFileConfig()

**実装場所**: `internal/config/manager.go:164-198`

**処理**:
1. グローバル設定を適用
2. デフォルトゲートウェイを適用

### applyEnvConfig()

**実装場所**: `internal/config/manager.go:200-239`

**処理**:
1. 環境変数から設定を読み込み
2. 各設定項目を上書き

### applyCLIConfig()

**実装場所**: `internal/config/manager.go:241-287`

**処理**:
1. コマンドライン引数から設定を読み込み
2. 各設定項目を上書き（最優先）
3. ゲートウェイ名が指定された場合は設定ファイルから取得

## 後方互換性

### レガシー設定形式のサポート

**実装場所**: `internal/config/file.go`

**変換関数**:
```go
func ConvertLegacyToNew(legacy *config.FileConfig) *config.Config
```

**処理**:
1. 旧形式の設定を読み込み
2. 新形式の構造体に変換
3. 検証を実施

## 使用例

### 基本的な使用

```go
// マネージャーの初期化
manager := config.NewManager(configPath)

// 設定ファイルの読み込み
if err := manager.Load(); err != nil {
    // エラーハンドリング
}

// コマンドライン引数の作成
cliArgs := &config.CLIArgs{
    URL:          *url,
    APIKey:       *apiKey,
    Timeout:      *timeout,
    Gateway:      *gateway,
    OutputFormat: *outputFormat,
}

// 設定の解決
resolved, err := manager.ResolveConfig(cliArgs)
if err != nil {
    // エラーハンドリング
}

// 設定の使用
fmt.Printf("Gateway URL: %s\n", resolved.Gateway.URL)
fmt.Printf("Output Format: %s\n", resolved.OutputFormat)
```

### 設定ソースの確認

```go
// ソース情報の取得
sourceInfo := manager.GetConfigSourceInfo(resolved)
fmt.Println(sourceInfo)
```

## テスト

### ユニットテスト

**場所**: `internal/config/*_test.go`

**テストケース**:
- 設定ファイルの読み込み
- 環境変数からの読み込み
- 設定の優先順位
- 設定の検証
- エラーハンドリング

### 統合テスト

**場所**: `test/integration/config_integration_test.go`

**テストシナリオ**:
- 複数ソースの統合
- ゲートウェイ設定の切り替え
- 設定ファイルとCLIの組み合わせ

## トラブルシューティング

### 設定が反映されない

**確認事項**:
1. 設定の優先順位を確認（CLI > 環境変数 > ファイル）
2. `--show-sources` で設定ソースを確認
3. 設定ファイルのパスが正しいか確認

### 設定ファイルが読み込めない

**確認事項**:
1. ファイルが存在するか確認
2. YAMLの構文が正しいか確認
3. パーミッションが適切か確認

**検証コマンド**:
```bash
llm-info --check-config --config /path/to/config.yaml
```

## 関連ドキュメント

- [アーキテクチャリファレンス](01-architecture.md)
- [使用ガイド](../USAGE.md)
- [設定ファイル例](../configs/example.yaml)
