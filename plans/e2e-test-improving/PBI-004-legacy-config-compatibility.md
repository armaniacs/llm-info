# PBI-004: レガー設定ファイルの互換性を改善

## 現在の問題

### 失敗しているテスト
- TestLegacyConfigCompatibility - YAMLパースエラーが発生

### エラーメッセージ
```
Failed to load legacy config file: failed to load config file (tried both new and legacy formats):
failed to parse config file: yaml: line 2: found character that cannot start any token
```

### 問題の分析
- レガー設定ファイルのフォーマットが不明
- YAMLパース時にエラーが発生
- 新旧フォーマットの移行パスが不完全

## 実装計画

### Phase 1: レガー設定ファイルの調査

#### 1.1 既存のレガー設定例を特定
```bash
# テストで使用されているレガー設定ファイルの場所を特定
find . -name "*legacy*" -type f
grep -r "LoadLegacyConfigFrom" . --include="*.go"
```

#### 1.2 考えられるレガー形式
A) 単純なキー・バリュー形式：
```yaml
base_url: https://api.example.com
api_key: your-key
timeout: 30s
```

B) 設定のネスト形式：
```yaml
llm_info:
  base_url: https://api.example.com
  api_key: your-key
  timeout: 30s
```

C) フラットな形式：
```yaml
url: https://api.example.com
key: your-key
timeout: 30s
```

### Phase 2: パーサーの改善

#### 2.1 internal/config/file.go - LoadLegacyConfigFromFile の改善
```go
func LoadLegacyConfigFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    // まず新しい形式で試す
    var newConfig Config
    if err := yaml.Unmarshal(data, &newConfig); err == nil {
        return &newConfig, nil
    }

    // 次に既知のレガー形式を試す
    return tryLegacyFormats(data, path)
}

// tryLegacyFormats は複数のレガー形式を試行
func tryLegacyFormats(data []byte, path string) (*Config, error) {
    formats := []func([]byte) (*Config, error){
        tryLegacyFlatFormat,
        tryLegacyNestedFormat,
        tryLegacyFileConfigFormat,
    }

    for i, formatFunc := range formats {
        if config, err := formatFunc(data); err == nil {
            // 成功した場合は移行を提案
            log.Printf("Successfully parsed legacy config using format %d from %s", i+1, path)
            log.Printf("Consider migrating to the new format for better compatibility")
            return config, nil
        }
    }

    return nil, fmt.Errorf("failed to parse config file as any known format from %s", path)
}
```

#### 2.2 各レガー形式のパーサー実装
```go
// tryLegacyFlatFormat - 単純なキー・バリュー形式
func tryLegacyFlatFormat(data []byte) (*Config, error) {
    var legacy struct {
        BaseURL     string        `yaml:"base_url"`
        URL         string        `yaml:"url"`
        APIKey      string        `yaml:"api_key"`
        Key         string        `yaml:"key"`
        Timeout     time.Duration `yaml:"timeout"`
        OutputFormat string        `yaml:"output_format"`
        SortBy      string        `yaml:"sort_by"`
    }

    if err := yaml.Unmarshal(data, &legacy); err != nil {
        return nil, err
    }

    // 新形式に変換
    config := &Config{
        DefaultGateway: "default",
        Global: Global{
            OutputFormat: legacy.OutputFormat,
            SortBy:       legacy.SortBy,
            Timeout:      legacy.Timeout,
        },
    }

    // URLの処理
    url := legacy.URL
    if url == "" {
        url = legacy.BaseURL
    }

    if url != "" {
        config.Gateways = []Gateway{
            {
                Name:    "default",
                URL:     url,
                APIKey:  legacy.APIKey,
                Timeout: legacy.Timeout,
            },
        }
    }

    return config, nil
}

// tryLegacyNestedFormat - ネストされた形式
func tryLegacyNestedFormat(data []byte) (*Config, error) {
    var legacy struct {
        LLMInfo struct {
            BaseURL      string        `yaml:"base_url"`
            APIKey       string        `yaml:"api_key"`
            Timeout      time.Duration `yaml:"timeout"`
            OutputFormat string        `yaml:"output_format"`
            SortBy       string        `yaml:"sort_by"`
        } `yaml:"llm_info"`
    }

    if err := yaml.Unmarshal(data, &legacy); err != nil {
        return nil, err
    }

    // 新形式に変換
    config := &Config{
        DefaultGateway: "default",
        Global: Global{
            OutputFormat: legacy.LLMInfo.OutputFormat,
            SortBy:       legacy.LLMInfo.SortBy,
            Timeout:      legacy.LLMInfo.Timeout,
        },
    }

    if legacy.LLMInfo.BaseURL != "" {
        config.Gateways = []Gateway{
            {
                Name:    "default",
                URL:     legacy.LLMInfo.BaseURL,
                APIKey:  legacy.LLMInfo.APIKey,
                Timeout: legacy.LLMInfo.Timeout,
            },
        }
    }

    return config, nil
}

// tryLegacyFileConfigFormat - FileConfig構造体を使用
func tryLegacyFileConfigFormat(data []byte) (*Config, error) {
    var fileConfig FileConfig
    if err := yaml.Unmarshal(data, &fileConfig); err != nil {
        return nil, err
    }

    // 新形式に変換
    config := &Config{
        Gateways:       fileConfig.Gateways,
        DefaultGateway: fileConfig.DefaultGateway,
    }

    // Common設定をGlobalに変換
    if fileConfig.Common.Timeout > 0 {
        config.Global.Timeout = fileConfig.Common.Timeout
    }
    if fileConfig.Common.Output.Format != "" {
        config.Global.OutputFormat = fileConfig.Common.Output.Format
    }

    return config, nil
}
```

### Phase 3: エラーハンドリングの改善

#### 3.1 詳細なエラー情報の提供
```go
func enhanceParseError(path string, originalErr error) error {
    var typeErr *yaml.TypeError
    var syntaxErr *yaml.SyntaxError

    if errors.As(originalErr, &syntaxErr) {
        line := syntaxErr.Line
        snippet := getYAMLLineSnippet(path, line)
        return fmt.Errorf("YAML syntax error at line %d in %s:\n%s\nError: %v",
            line, path, snippet, originalErr)
    }

    if errors.As(originalErr, &typeErr) {
        return fmt.Errorf("YAML structure error in %s. Please check the format:\n%v",
            path, originalErr)
    }

    return fmt.Errorf("failed to parse config file %s: %w", path, originalErr)
}

// getYAMLLineSnippet はエラー行の前後を取得
func getYAMLLineSnippet(path string, line int) string {
    content, err := os.ReadFile(path)
    if err != nil {
        return ""
    }

    lines := strings.Split(string(content), "\n")
    start := max(0, line-3)
    end := min(len(lines), line+2)

    var snippet strings.Builder
    for i := start; i < end; i++ {
        prefix := "  "
        if i == line-1 {
            prefix = "> "
        }
        snippet.WriteString(fmt.Sprintf("%s%d: %s\n", prefix, i+1, lines[i]))
    }

    return snippet.String()
}
```

### Phase 4: テストの改善

#### 4.1 test/integration/config_integration_test.go
```go
func TestLegacyConfigCompatibility(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        wantErr  bool
        expected map[string]interface{}
    }{
        {
            name: "flat legacy format",
            content: `
base_url: https://api.example.com
api_key: test-key
timeout: 30s
output_format: json
`,
            wantErr: false,
            expected: map[string]interface{}{
                "url":           "https://api.example.com",
                "api_key":       "test-key",
                "timeout":       30 * time.Second,
                "output_format": "json",
            },
        },
        {
            name: "nested legacy format",
            content: `
llm_info:
  base_url: https://api.example.com
  api_key: test-key
  timeout: 30s
`,
            wantErr: false,
            expected: map[string]interface{}{
                "url":      "https://api.example.com",
                "api_key":  "test-key",
                "timeout":  30 * time.Second,
            },
        },
        {
            name: "invalid legacy format",
            content: `
- invalid yaml
- starts with dash
- no proper structure
`,
            wantErr: true,
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            // 一時的な設定ファイルを作成
            tmpFile := filepath.Join(t.TempDir(), "config.yaml")
            err := os.WriteFile(tmpFile, []byte(test.content), 0644)
            require.NoError(t, err)
            defer os.Remove(tmpFile)

            // レガー設定の読み込み
            config, err := LoadLegacyConfigFromFile(tmpFile)

            if test.wantErr {
                assert.Error(t, err)
                // エラーメッセージが役立つことを確認
                assert.Contains(t, err.Error(), "please check the format")
            } else {
                assert.NoError(t, err)

                // 値の検証
                require.NotNil(t, config.DefaultGateway)
                require.NotEmpty(t, config.Gateways)

                gateway := config.Gateways[0]
                if url, ok := test.expected["url"].(string); ok {
                    assert.Equal(t, url, gateway.URL)
                }
                if key, ok := test.expected["api_key"].(string); ok {
                    assert.Equal(t, key, gateway.APIKey)
                }
            }
        })
    }
}
```

## 移行支援ツール

### config-migrate コマンドの提案
```bash
# 既存の設定ファイルを新しい形式に変換
llm-info --migrate-config

# 変換の差分を確認
llm-info --migrate-config --dry-run
```

## 実装手順

1. レガー形式の調査と特定
2. tryLegacyFormats()と各形式のパーサーを実装
3. エラーハンドリングの改善
4. テストケースの追加と更新
5. 移行ツールの検討（オプション）

## 実装記録

### [2026-01-10 19:49:00]

**実装者**: Claude Code

**実装内容**:
- `internal/config/file.go:115-394`: LoadLegacyConfigFromFileを修正して複数のレガー形式をサポート
- `internal/config/file.go:145-198`: tryLegacyFlatFormatを実装（base_url, api_key, timeout, output_format）
- `internal/config/file.go:200-268`: tryLegacyNestedFormatを実装（llm_infoネスト形式）
- `internal/config/file.go:270-336`: tryLegacyFileConfigFormatを実装（FileConfig形式）
- `internal/config/file.go:357-367`: enhanceParseErrorを実装してエラーメッセージを改善
- `internal/config/manager.go:74-87`: Loadメソッドを修正してレガシー形式を優先
- `internal/config/manager.go:422-425`: SetNewConfigメソッドを追加
- `test/integration/config_integration_test.go:143-198`: TestLegacyConfigCompatibilityを修正

**遭遇した問題と解決策**:
- **問題**: YAMLパース時、time.Durationが数値だとパースできない
  **解決策**: timeoutをstringとして読み込み、time.ParseDurationで変換
- **問題**: 新形式がレガシー形式より先にパースされる
  **解決策**: LoadLegacyConfigFromFileの処理順序を逆転
- **問題**: Gatewayが作成されない
  **解決策**: 各レガー形式のパーサーでGateway構造を正しく構築
- **問題**: テストが余計なパターンにマッチする
  **解決策**: tryLegacyNestedFormatが空のllm_infoでも成功しないように修正

**テスト結果**:
- TestLegacyConfigCompatibility: ✅ 成功
- internal/configパッケージの全テスト: ✅ 成功
- 3つのレガー形式（flat, nested, FileConfig）をサポート

**受け入れ基準の達成状況**:
- [x] TestLegacyConfigCompatibilityがパス
- [x] すべての既知のレガー形式が読み込める
- [x] エラーメッセージが改善される
- [x] 新旧形式の混在環境で動作する

**備考**:
現在、3つのレガー形式をサポート:
1. Flat形式: base_url, api_key, timeout, output_format
2. Nested形式: llm_info配下の設定
3. FileConfig形式: Gateways, DefaultGateway, Common構造

## 成功条件

- [x] TestLegacyConfigCompatibilityがパス
- [x] すべての既知のレガー形式が読み込める
- [x] エラーメッセージが改善される
- [x] 新旧形式の混在環境で動作する

## 注意事項

- レガー形式のサポートは段階的に廃止することを検討
- ユーザーには新しい形式への移行を推奨
- パフォーマンスへの影響を最小限に抑える