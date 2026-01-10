# PBI 3: 環境変数サポート - 実装計画

## PBI概要と目的

**タイトル**: 環境変数による設定機能

**目的**: 環境変数でゲートウェイ情報を設定し、CI/CDパイプラインやコンテナ環境で利用できるようにすること。

**ビジネス価値**:
- 環境対応：CI/CDやコンテナ環境での利用
- セキュリティ：APIキーを環境変数で管理
- 柔軟性：設定方法の選択肢拡大

## 現状の課題

1. 現在はコマンドライン引数と設定ファイルのみ対応
2. CI/CDパイプラインでの利用が困難
3. コンテナ環境での設定が煩雑
4. APIキーをファイルに保存するセキュリティリスク

## 実装計画

### 1. 環境変数の設計

#### サポートする環境変数
```bash
# 基本設定
LLM_INFO_URL              # ゲートウェイURL
LLM_INFO_API_KEY          # APIキー
LLM_INFO_TIMEOUT          # リクエストタイムアウト（秒）
LLM_INFO_DEFAULT_GATEWAY  # デフォルトゲートウェイ名

# 表示設定
LLM_INFO_OUTPUT_FORMAT    # 出力形式（table/json）
LLM_INFO_SORT_BY          # ソート項目
LLM_INFO_FILTER           # フィルタ条件

# 詳細設定
LLM_INFO_CONFIG_PATH      # 設定ファイルパス
LLM_INFO_LOG_LEVEL        # ログレベル
LLM_INFO_USER_AGENT       # ユーザーエージェント
```

### 2. コード構造

#### 既存ファイル修正
- `internal/config/env.go` - 環境変数読み込み機能の拡張
- `internal/config/manager.go` - 設定優先順位の実装
- `cmd/llm-info/main.go` - 環境変数処理の統合

### 3. 詳細実装

#### internal/config/env.go（拡張）
```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

// EnvConfig は環境変数から読み込んだ設定を表す
type EnvConfig struct {
    URL             string
    APIKey          string
    Timeout         time.Duration
    DefaultGateway  string
    OutputFormat    string
    SortBy          string
    Filter          string
    ConfigPath      string
    LogLevel        string
    UserAgent       string
}

// LoadEnvConfig は環境変数から設定を読み込む
func LoadEnvConfig() *EnvConfig {
    return &EnvConfig{
        URL:            os.Getenv("LLM_INFO_URL"),
        APIKey:         os.Getenv("LLM_INFO_API_KEY"),
        Timeout:        parseTimeout(os.Getenv("LLM_INFO_TIMEOUT")),
        DefaultGateway: os.Getenv("LLM_INFO_DEFAULT_GATEWAY"),
        OutputFormat:   os.Getenv("LLM_INFO_OUTPUT_FORMAT"),
        SortBy:         os.Getenv("LLM_INFO_SORT_BY"),
        Filter:         os.Getenv("LLM_INFO_FILTER"),
        ConfigPath:     os.Getenv("LLM_INFO_CONFIG_PATH"),
        LogLevel:       os.Getenv("LLM_INFO_LOG_LEVEL"),
        UserAgent:      os.Getenv("LLM_INFO_USER_AGENT"),
    }
}

// parseTimeout はタイムアウト文字列をtime.Durationに変換する
func parseTimeout(s string) time.Duration {
    if s == "" {
        return 0
    }
    
    // 秒数を指定
    if seconds, err := strconv.Atoi(s); err == nil {
        return time.Duration(seconds) * time.Second
    }
    
    // duration形式を指定
    if duration, err := time.ParseDuration(s); err == nil {
        return duration
    }
    
    return 0
}

// IsSet は環境変数が設定されているかチェックする
func (e *EnvConfig) IsSet() bool {
    return e.URL != "" || e.APIKey != "" || e.DefaultGateway != ""
}

// ToGatewayConfig は環境変数設定をGatewayConfigに変換する
func (e *EnvConfig) ToGatewayConfig() *GatewayConfig {
    return &GatewayConfig{
        URL:     e.URL,
        APIKey:  e.APIKey,
        Timeout: e.Timeout,
    }
}

// Validate は環境変数設定を検証する
func (e *EnvConfig) Validate() error {
    if e.URL != "" {
        if _, err := url.Parse(e.URL); err != nil {
            return fmt.Errorf("invalid LLM_INFO_URL: %w", err)
        }
    }
    
    if e.Timeout < 0 {
        return fmt.Errorf("LLM_INFO_TIMEOUT must be positive")
    }
    
    if e.OutputFormat != "" {
        validFormats := []string{"table", "json"}
        if !contains(validFormats, e.OutputFormat) {
            return fmt.Errorf("invalid LLM_INFO_OUTPUT_FORMAT: %s (valid: %s)", 
                e.OutputFormat, strings.Join(validFormats, ", "))
        }
    }
    
    return nil
}

// contains はスライスに文字列が含まれているかチェックする
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

#### internal/config/manager.go（拡張）
```go
package config

import (
    "fmt"
)

// ConfigSource は設定ソースの種類を表す
type ConfigSource int

const (
    SourceDefault ConfigSource = iota
    SourceFile
    SourceEnv
    SourceCLI
)

// ResolvedConfig は解決された設定を表す
type ResolvedConfig struct {
    Gateway       *GatewayConfig
    OutputFormat  string
    SortBy        string
    Filter        string
    LogLevel      string
    UserAgent     string
    Sources       map[string]ConfigSource
}

// ResolveConfig はすべての設定ソースから最終的な設定を解決する
func (m *Manager) ResolveConfig(cliArgs *CLIArgs) (*ResolvedConfig, error) {
    resolved := &ResolvedConfig{
        Sources: make(map[string]ConfigSource),
    }
    
    // 1. デフォルト値を設定
    if err := m.applyDefaults(resolved); err != nil {
        return nil, err
    }
    
    // 2. 設定ファイルから適用
    if err := m.applyFileConfig(resolved); err != nil {
        return nil, err
    }
    
    // 3. 環境変数から適用
    if err := m.applyEnvConfig(resolved); err != nil {
        return nil, err
    }
    
    // 4. コマンドライン引数から適用
    if err := m.applyCLIConfig(resolved, cliArgs); err != nil {
        return nil, err
    }
    
    // 5. 最終的な設定の検証
    if err := m.validateResolvedConfig(resolved); err != nil {
        return nil, err
    }
    
    return resolved, nil
}

// applyDefaults はデフォルト値を適用する
func (m *Manager) applyDefaults(resolved *ResolvedConfig) error {
    resolved.OutputFormat = "table"
    resolved.SortBy = "name"
    resolved.Sources["output_format"] = SourceDefault
    resolved.Sources["sort_by"] = SourceDefault
    return nil
}

// applyFileConfig は設定ファイルから設定を適用する
func (m *Manager) applyFileConfig(resolved *ResolvedConfig) error {
    if m.config == nil {
        return nil
    }
    
    // グローバル設定を適用
    if m.config.Global.OutputFormat != "" {
        resolved.OutputFormat = m.config.Global.OutputFormat
        resolved.Sources["output_format"] = SourceFile
    }
    
    if m.config.Global.SortBy != "" {
        resolved.SortBy = m.config.Global.SortBy
        resolved.Sources["sort_by"] = SourceFile
    }
    
    return nil
}

// applyEnvConfig は環境変数から設定を適用する
func (m *Manager) applyEnvConfig(resolved *ResolvedConfig) error {
    envConfig := LoadEnvConfig()
    
    if envConfig.URL != "" || envConfig.APIKey != "" {
        resolved.Gateway = envConfig.ToGatewayConfig()
        resolved.Sources["gateway"] = SourceEnv
    }
    
    if envConfig.OutputFormat != "" {
        resolved.OutputFormat = envConfig.OutputFormat
        resolved.Sources["output_format"] = SourceEnv
    }
    
    if envConfig.SortBy != "" {
        resolved.SortBy = envConfig.SortBy
        resolved.Sources["sort_by"] = SourceEnv
    }
    
    if envConfig.Filter != "" {
        resolved.Filter = envConfig.Filter
        resolved.Sources["filter"] = SourceEnv
    }
    
    if envConfig.LogLevel != "" {
        resolved.LogLevel = envConfig.LogLevel
        resolved.Sources["log_level"] = SourceEnv
    }
    
    if envConfig.UserAgent != "" {
        resolved.UserAgent = envConfig.UserAgent
        resolved.Sources["user_agent"] = SourceEnv
    }
    
    return nil
}

// applyCLIConfig はコマンドライン引数から設定を適用する
func (m *Manager) applyCLIConfig(resolved *ResolvedConfig, cliArgs *CLIArgs) error {
    if cliArgs == nil {
        return nil
    }
    
    // ゲートウェイ設定
    if cliArgs.URL != "" || cliArgs.APIKey != "" {
        resolved.Gateway = &GatewayConfig{
            URL:     cliArgs.URL,
            APIKey:  cliArgs.APIKey,
            Timeout: cliArgs.Timeout,
        }
        resolved.Sources["gateway"] = SourceCLI
    } else if cliArgs.Gateway != "" {
        // ゲートウェイ名が指定された場合は設定ファイルから取得
        gatewayConfig, err := m.GetGatewayConfig(cliArgs.Gateway)
        if err != nil {
            return fmt.Errorf("failed to get gateway config: %w", err)
        }
        resolved.Gateway = gatewayConfig
        resolved.Sources["gateway"] = SourceCLI
    }
    
    // その他の設定
    if cliArgs.OutputFormat != "" {
        resolved.OutputFormat = cliArgs.OutputFormat
        resolved.Sources["output_format"] = SourceCLI
    }
    
    if cliArgs.SortBy != "" {
        resolved.SortBy = cliArgs.SortBy
        resolved.Sources["sort_by"] = SourceCLI
    }
    
    if cliArgs.Filter != "" {
        resolved.Filter = cliArgs.Filter
        resolved.Sources["filter"] = SourceCLI
    }
    
    return nil
}

// validateResolvedConfig は解決された設定を検証する
func (m *Manager) validateResolvedConfig(resolved *ResolvedConfig) error {
    if resolved.Gateway == nil {
        return fmt.Errorf("no gateway configuration found")
    }
    
    if resolved.Gateway.URL == "" {
        return fmt.Errorf("gateway URL is required")
    }
    
    return nil
}

// GetConfigSourceInfo は設定ソース情報を返す
func (m *Manager) GetConfigSourceInfo(resolved *ResolvedConfig) string {
    info := "Configuration sources:\n"
    
    for key, source := range resolved.Sources {
        sourceName := ""
        switch source {
        case SourceDefault:
            sourceName = "default"
        case SourceFile:
            sourceName = "config file"
        case SourceEnv:
            sourceName = "environment variable"
        case SourceCLI:
            sourceName = "command line"
        }
        info += fmt.Sprintf("  %s: %s\n", key, sourceName)
    }
    
    return info
}
```

### 4. テスト戦略

#### 単体テスト
- `internal/config/env_test.go` - 環境変数読み込みのテスト
- `internal/config/manager_test.go` - 設定優先順位のテスト

#### 統合テスト
- `test/integration/env_priority_test.go` - 環境変数と他の設定ソースの優先順位テスト

#### E2Eテスト
- `test/e2e/env_variables_test.go` - 環境変数を使用したコマンド実行テスト

### 5. 必要なファイルの新規作成・修正

#### 新規作成ファイル
1. `test/integration/env_priority_test.go`
2. `test/e2e/env_variables_test.go`

#### 修正ファイル
1. `internal/config/env.go`
2. `internal/config/manager.go`
3. `cmd/llm-info/main.go`
4. `internal/config/env_test.go`
5. `internal/config/manager_test.go`

### 6. 受け入れ基準チェックリスト

- [ ] `LLM_INFO_URL`と`LLM_INFO_API_KEY`環境変数に対応
- [ ] 設定の優先順位：コマンドライン > 環境変数 > 設定ファイル
- [ ] 環境変数のバリデーション
- [ ] すべての設定項目の環境変数対応
- [ ] 環境変数設定の優先順位表示機能
- [ ] 単体テストカバレッジ80%以上
- [ ] 統合テストの実装
- [ ] E2Eテストの実装

### 7. 実装手順

1. **環境変数読み込み機能の拡張**
   - `internal/config/env.go`の修正
   - 新しい環境変数の追加

2. **設定優先順位の実装**
   - `internal/config/manager.go`の修正
   - 設定解決ロジックの実装

3. **コマンドラインインターフェースの修正**
   - `cmd/llm-info/main.go`の修正
   - 環境変数処理の統合

4. **設定ソース情報の実装**
   - 設定ソースの可視化機能
   - デバッグ情報の提供

5. **テストの実装**
   - 単体テストの拡張
   - 統合テストの作成
   - E2Eテストの作成

6. **ドキュメント更新**
   - 環境変数一覧のドキュメント化
   - 使用例の追加

### 8. リスクと対策

#### リスク
1. 環境変数の型変換エラー
2. 設定優先順位の複雑化
3. 環境変数名の競合

#### 対策
1. 堅牢な型変換とエラーハンドリング
2. 明確な優先順位ドキュメントとテスト
3. プレフィックスによる名前空間の確保

### 9. 成功指標

- 環境変数からの正常な読み込み成功率100%
- 設定優先順位の正しい動作率100%
- 環境変数バリデーションのエラー検出率100%
- テストカバレッジ80%以上

### 10. 使用例

#### 基本使用例
```bash
# 環境変数でゲートウェイを設定
export LLM_INFO_URL="https://api.example.com"
export LLM_INFO_API_KEY="your-api-key"
llm-info

# 出力形式を指定
export LLM_INFO_OUTPUT_FORMAT="json"
llm-info

# ソートとフィルタを指定
export LLM_INFO_SORT_BY="max_tokens"
export LLM_INFO_FILTER="gpt"
llm-info
```

#### CI/CDパイプラインでの使用例
```yaml
# GitHub Actions
- name: List available models
  env:
    LLM_INFO_URL: ${{ secrets.LLM_GATEWAY_URL }}
    LLM_INFO_API_KEY: ${{ secrets.LLM_API_KEY }}
    LLM_INFO_OUTPUT_FORMAT: json
  run: |
    llm-info > models.json
```

#### Dockerコンテナでの使用例
```dockerfile
FROM alpine:latest
RUN apk add --no-cache llm-info
ENV LLM_INFO_URL="https://api.example.com"
ENV LLM_INFO_API_KEY=""
CMD ["llm-info"]