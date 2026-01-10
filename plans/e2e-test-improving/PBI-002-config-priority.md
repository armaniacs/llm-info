# PBI-002: 設定優先順位の実装を修正

## 現在の問題

### 失敗しているテスト
1. TestConfigPriority/partial_CLI_override - CLIで部分的な設定をした場合、環境変数からの値が保持されない
2. TestConfigSourceInfo - ソース情報に"environment variable"が含まれない
3. TestConfigValidation/invalid_URL - URL検証が期待通り動作しない

### 問題の根本原因
- `applyCLIConfig()`がCLI引数が一つでもあるとGateway全体を上書きする
- Gateway単位でソースを管理しているため、個別のフィールドのソースを追跡できない

## 実装計画

### Phase 1: GatewayConfigの拡張

#### 1.1 pkg/config/config.go の修正
```go
type GatewayConfig struct {
    Name        string        `yaml:"name"`
    URL         string        `yaml:"url"`
    APIKey      string        `yaml:"api_key"`
    Timeout     time.Duration `yaml:"timeout"`

    // ソース追跡（JSON/YAML出力から除外）
    URLSource     ConfigSource `json:"-" yaml:"-"`
    APIKeySource  ConfigSource `json:"-" yaml:"-"`
    TimeoutSource ConfigSource `json:"-" yaml:"-"`
}

// GetURLSource はURLの設定ソースを返す
func (g *GatewayConfig) GetURLSource() ConfigSource {
    return g.URLSource
}

// GetAPIKeySource はAPIKeyの設定ソースを返す
func (g *GatewayConfig) GetAPIKeySource() ConfigSource {
    return g.APIKeySource
}

// GetTimeoutSource はTimeoutの設定ソースを返す
func (g *GatewayConfig) GetTimeoutSource() ConfigSource {
    return g.TimeoutSource
}
```

### Phase 2: マージロジックの改善

#### 2.1 internal/config/manager.go - applyEnvConfig
```go
func (m *Manager) applyEnvConfig(resolved *ResolvedConfig) error {
    envConfig := LoadEnvConfig()

    // Gatewayの初期化（初回のみ）
    if resolved.Gateway == nil {
        resolved.Gateway = &config.GatewayConfig{}
    }

    // 各フィールドを個別に設定
    if envConfig.URL != "" {
        resolved.Gateway.URL = envConfig.URL
        resolved.Gateway.URLSource = SourceEnv
        resolved.Sources["gateway.url"] = SourceEnv
    }

    if envConfig.APIKey != "" {
        resolved.Gateway.APIKey = envConfig.APIKey
        resolved.Gateway.APIKeySource = SourceEnv
        resolved.Sources["gateway.api_key"] = SourceEnv
    }

    if envConfig.Timeout > 0 {
        resolved.Gateway.Timeout = envConfig.Timeout
        resolved.Gateway.TimeoutSource = SourceEnv
        resolved.Sources["gateway.timeout"] = SourceEnv
    }

    // APIキーが空でもURLが設定されている場合は、ソース情報を保持
    if envConfig.URL != "" && envConfig.APIKey == "" && resolved.Sources["gateway"] == 0 {
        resolved.Sources["gateway"] = SourceEnv
    }

    // その他の設定...
}
```

#### 2.2 internal/config/manager.go - applyCLIConfig
```go
func (m *Manager) applyCLIConfig(resolved *ResolvedConfig, cliArgs *CLIArgs) error {
    if cliArgs == nil {
        return nil
    }

    // Gatewayの初期化（初回のみ）
    if resolved.Gateway == nil {
        resolved.Gateway = &config.GatewayConfig{}
    }

    // CLIで指定されたフィールドのみ更新
    if cliArgs.URL != "" {
        resolved.Gateway.URL = cliArgs.URL
        resolved.Gateway.URLSource = SourceCLI
        resolved.Sources["gateway.url"] = SourceCLI
    }

    if cliArgs.APIKey != "" {
        resolved.Gateway.APIKey = cliArgs.APIKey
        resolved.Gateway.APIKeySource = SourceCLI
        resolved.Sources["gateway.api_key"] = SourceCLI
    }

    if cliArgs.Timeout > 0 {
        resolved.Gateway.Timeout = cliArgs.Timeout
        resolved.Gateway.TimeoutSource = SourceCLI
        resolved.Sources["gateway.timeout"] = SourceCLI
    }

    // Gateway全体のソース情報（後方互換性）
    if cliArgs.URL != "" || cliArgs.APIKey != "" {
        resolved.Sources["gateway"] = SourceCLI
    }

    // その他の設定...
}
```

### Phase 3: ソース情報表示の改善

#### 3.1 internal/config/manager.go - GetConfigSourceInfo
```go
func (m *Manager) GetConfigSourceInfo(resolved *ResolvedConfig) string {
    info := "Configuration sources:\n"

    // Gatewayの詳細なソース情報
    if resolved.Gateway != nil {
        if resolved.Gateway.URLSource > 0 {
            info += fmt.Sprintf("  gateway.url: %s\n", m.getSourceName(resolved.Gateway.URLSource))
        }
        if resolved.Gateway.APIKeySource > 0 {
            info += fmt.Sprintf("  gateway.api_key: %s\n", m.getSourceName(resolved.Gateway.APIKeySource))
        }
        if resolved.Gateway.TimeoutSource > 0 {
            info += fmt.Sprintf("  gateway.timeout: %s\n", m.getSourceName(resolved.Gateway.TimeoutSource))
        }
    }

    // その他の設定
    for key, source := range resolved.Sources {
        if !strings.HasPrefix(key, "gateway.") && key != "gateway" {
            info += fmt.Sprintf("  %s: %s\n", key, m.getSourceName(source))
        }
    }

    return info
}

// getSourceName はConfigSourceを文字列に変換
func (m *Manager) getSourceName(source ConfigSource) string {
    switch source {
    case SourceDefault:
        return "default"
    case SourceFile:
        return "config file"
    case SourceEnv:
        return "environment variable"
    case SourceCLI:
        return "command line"
    default:
        return "unknown"
    }
}
```

## 検証テスト

### TestConfigSourceInfoの期待値
```
Configuration sources:
  gateway.url: command line
  gateway.api_key: environment variable
  gateway.timeout: environment variable
  output_format: command line
  sort_by: default
```

### TestConfigPriority/partial_CLI_override
- CLI: URL, OutputFormat
- 環境変数: URL, APIKey, Timeout, OutputFormat, SortBy
- 期待結果: URLとOutputFormatはCLI、APIKeyとTimeoutとSortByは環境変数

## 実装手順

1. GatewayConfigにソースフィールドを追加
2. applyEnvConfigを修正して個別フィールドを更新
3. applyCLIConfigを修正して部分的な更新をサポート
4. GetConfigSourceInfoを修正して詳細なソースを表示
5. テストを実行して修正を確認

## リスク

1. **後方互換性**: 新しいフィールドがJSON/YAMLに含まれる可能性
   - 対策: `json:"-"`と`yaml:"-"`タグを使用
2. **既存のテスト**: Sourcesマップを使用するテストへの影響
   - 対策: 移行期間中は新旧両方をサポート

## 成功条件

- [ ] TestConfigPriority/partial_CLI_overrideがパス
- [ ] TestConfigSourceInfoがパス
- [ ] TestConfigValidation/invalid_URLがパス
- [ ] 手動テストで部分的なCLI overrideが動作
- [ ] ソース情報が正しく表示される