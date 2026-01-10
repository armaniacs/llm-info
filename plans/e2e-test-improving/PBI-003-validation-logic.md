# PBI-003: 設定検証ロジックを修正

## 現在の問題

### 失敗しているテスト
1. TestConfigValidation/invalid_URL - 無効なURLが検証されない
2. TestEnvironmentVariableValidation/invalid_timeout - 不正なタイムアウトが検証されない
3. TestEnvironmentVariableValidation/invalid_output_format - 不正な出力形式が検証されない

### 問題の分析
- Validate()メソッドが期待通りにエラーを返していない
- URLパースが寛大すぎる（"invalid-url"でもパース可能）
- 環境変数の検証タイミングの問題

## 実装計画

### Phase 1: URL検証の強化

#### 1.1 internal/config/env.go - Validateメソッドの改善
```go
func (e *EnvConfig) Validate() error {
    // URL検証の改善
    if e.URL != "" {
        // より厳格なURL検証
        if !isValidURL(e.URL) {
            return fmt.Errorf("invalid LLM_INFO_URL: %q", e.URL)
        }

        // スキームのチェック
        parsed, err := url.Parse(e.URL)
        if err != nil {
            return fmt.Errorf("invalid LLM_INFO_URL format: %w", err)
        }

        // HTTP/HTTPSのみ許可
        if parsed.Scheme != "http" && parsed.Scheme != "https" {
            return fmt.Errorf("LLM_INFO_URL must use http or https scheme, got: %s", parsed.Scheme)
        }

        // ホスト名の存在チェック
        if parsed.Host == "" {
            return fmt.Errorf("LLM_INFO_URL must have a valid host")
        }
    }

    // タイムアウト検証
    if e.Timeout < 0 {
        return fmt.Errorf("LLM_INFO_TIMEOUT must be positive, got: %v", e.Timeout)
    }

    // 出力形式検証
    if e.OutputFormat != "" {
        validFormats := []string{"table", "json"}
        if !contains(validFormats, e.OutputFormat) {
            return fmt.Errorf("invalid LLM_INFO_OUTPUT_FORMAT: %s (valid: %s)",
                e.OutputFormat, strings.Join(validFormats, ", "))
        }
    }

    return nil
}

// isValidURL はURL形式をより厳密に検証
func isValidURL(urlStr string) bool {
    // 明らかに無効なパターンをチェック
    invalidPatterns := []string{
        " ",  // スペースを含む
        "\t", // タブを含む
        "\n", // 改行を含む
        "//", // スキームなしで始まる
    }

    for _, pattern := range invalidPatterns {
        if strings.Contains(urlStr, pattern) {
            return false
        }
    }

    // url.Parseで検証
    _, err := url.Parse(urlStr)
    return err == nil
}
```

### Phase 2: 環境変数検証の統合

#### 2.1 internal/config/manager.go - validateResolvedConfigの改善
```go
func (m *Manager) validateResolvedConfig(resolved *ResolvedConfig) error {
    // Gateway設定の検証
    if resolved.Gateway != nil {
        if resolved.Gateway.URL == "" {
            return fmt.Errorf("gateway URL is required")
        }

        // GatewayのURLも検証
        if !isValidURL(resolved.Gateway.URL) {
            return fmt.Errorf("invalid gateway URL: %q", resolved.Gateway.URL)
        }

        if resolved.Gateway.Timeout < 0 {
            return fmt.Errorf("gateway timeout must be positive, got: %v", resolved.Gateway.Timeout)
        }
    }

    // 出力形式の検証
    if resolved.OutputFormat != "" {
        validFormats := []string{"table", "json"}
        if !contains(validFormats, resolved.OutputFormat) {
            return fmt.Errorf("invalid output format: %s (valid: %s)",
                resolved.OutputFormat, strings.Join(validFormats, ", "))
        }
    }

    return nil
}
```

### Phase 3: テストの期待値調整

#### 3.1 test/integration/env_priority_test.go - TestConfigValidation
```go
{
    name:    "invalid URL",
    envVars: map[string]string{"LLM_INFO_URL": "http://invalid url with spaces"},
    wantErr: true,
},
{
    name:    "URL without http scheme",
    envVars: map[string]string{"LLM_INFO_URL": "ftp://example.com"},
    wantErr: true,
},
{
    name:    "URL without host",
    envVars: map[string]string{"LLM_INFO_URL": "http://"},
    wantErr: true,
},
```

## テストケースの追加

### 新しいテストシナリオ
1. **URL検証の厳格化**
   - スペースを含むURL
   - ファイルプロトコル (file://)
   - 不正なスキーム (ftp://)
   - ホスト名なしのURL

2. **タイムアウト検証**
   - 負の値
   - 極端に大きな値
   - 0秒（許可するかを決定）

3. **出力形式検証**
   - 大文字・小文字の混合
   - 不明な形式
   - ヌル文字列（許可）

## 実装手順

1. `isValidURL()`ヘルパー関数の実装
2. `EnvConfig.Validate()`の改善
3. 環境変数テストのデータ更新
4. `validateResolvedConfig()`の改善
5. すべての検証関連テストを実行

## エラーメッセージの改善

### 改善前
```
Error: invalid LLM_INFO_URL: %v
```

### 改善後
```
Error: invalid LLM_INFO_URL format: "ht tp://invalid url"
       URL must not contain spaces and must use http or https scheme
```

## 検証方法

### テスト実行
```bash
go test ./internal/config -run TestEnvConfig_ValidateNew -v
go test ./test/integration -run TestConfigValidation -v
go test ./test/integration -run TestEnvironmentVariableValidation -v
```

### 手動テスト
```bash
# 無効なURL
export LLM_INFO_URL="http://invalid url"
llm-info

# 無効な出力形式
export LLM_INFO_OUTPUT_FORMAT="xml"
llm-info

# 負のタイムアウト
export LLM_INFO_TIMEOUT="-1s"
llm-info
```

## 成功条件

- [ ] TestConfigValidation/invalid_URLがパス
- [ ] TestEnvironmentVariableValidation/invalid_timeoutがパス
- [ ] TestEnvironmentVariableValidation/invalid_output_formatがパス
- [ ] エラーメッセージが理解しやすい
- [ ] 有効な設定値が正しく受け入れられる

## 備考

- 検証は厳格にしすぎるとユーザビリティが低下する可能性
- エラーメッセージでは具体的な修正案を提示
- 開発時には検証を緩和するオプションを検討（DEBUGモード等）