# プロダクトバックログアイテム: テーブル形式での結果出力

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 3

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## 依存PBI
PBI-2-0-005: API呼び出しタイムアウト問題の修正（必須）

## ユーザーストーリー

開発者として、探索結果を見やすいテーブル形式で表示したい、なぜならCLIで直接結果を確認して迅速に判断したいから

## ビジネス価値

- **UX向上**: CLIでの直接確認が容易になり、作業効率が向上
- **可読性**: 構造化された表示により、結果の理解が即座に可能
- **採用促進**: 見やすい出力により、チーム内での採用ハードルが低下
- **デバッグ支援**: 探索過程が一目で分かり、問題特定が容易

## BDD受け入れシナリオ

### Context Window探索

```gherkin
Scenario: Context Window探索結果をテーブル形式で表示
  Given probe-contextコマンドを--verboseなしで実行する
  When 探索が完了する
  Then 結果が以下の形式でテーブル表示される
    """
    Context Window Probe Results
    ============================
    Model:                GLM-4.6
    Estimated Context:    127,000 tokens
    Method Confidence:    high
    Trials:               12
    Duration:             45.3s
    Max Input at Success: 126,800 tokens

    Status: ✓ Success
    """

Scenario: Context Window探索の詳細情報表示
  Given probe-contextコマンドを--verboseで実行する
  When 探索が完了する
  Then テーブル出力に加えて探索履歴が表示される
  And 各試行の成功/失敗が表示される
  And APIコール回数が表示される

Scenario: Context Window探索のエラー表示
  Given 探索が失敗した場合
  When エラーが発生する
  Then テーブル形式でエラー情報が表示される
    """
    Context Window Probe Results
    ============================
    Model:   GLM-4.6
    Status:  ✗ Failed
    Error:   API timeout after 30s
    Trials:  5

    Suggestion: Increase timeout with --timeout flag
    """
```

### Max Output Tokens探索

```gherkin
Scenario: Max Output探索結果をテーブル形式で表示
  Given probe-max-outputコマンドを実行する
  When 探索が完了する
  Then 結果が以下の形式でテーブル表示される
    """
    Max Output Tokens Probe Results
    ================================
    Model:                GLM-4.6
    Max Output Tokens:    16,384
    Evidence:             validation_error
    Trials:               8
    Duration:             30.2s
    Max Successfully Gen: 8,192 tokens

    Status: ✓ Success
    """

Scenario: 複数の検出方法が使われた場合の表示
  Given バリデーションエラーとincomplete両方で検出された場合
  When 結果を表示する
  Then 優先された検出方法が表示される
  And 副次的な検出方法も記録される
```

## 受け入れ基準

### 機能要件
- [ ] probe-context の結果がテーブル形式で表示される
- [ ] probe-max-output の結果がテーブル形式で表示される
- [ ] 成功時と失敗時で適切なフォーマットが選択される
- [ ] 数値が3桁区切り（カンマ）で表示される
- [ ] ステータスマーク（✓/✗）が表示される

### 詳細表示（verbose）要件
- [ ] --verboseフラグで詳細情報が追加表示される
- [ ] 探索履歴（各試行の結果）が表示される
- [ ] APIコール回数と総消費時間が表示される
- [ ] 検出方法の詳細が表示される

### UX要件
- [ ] 80文字幅のターミナルで適切に表示される
- [ ] ヘッダーと本文が視覚的に区別できる
- [ ] 重要な情報（推定値、ステータス）が強調表示される
- [ ] エラー時は改善提案が表示される

### 品質要件
- [ ] 日本語と英語が混在しても整列される
- [ ] 長いモデル名でもレイアウトが崩れない
- [ ] カラー出力が無効な環境でも読みやすい

## 実装内容

### 新規ファイル

#### 1. internal/ui/table_formatter.go

テーブル整形の核となるロジック

```go
package ui

import (
    "fmt"
    "strings"
    "time"
)

// TableFormatter はテーブル形式の出力を整形する
type TableFormatter struct {
    width int // テーブル幅（デフォルト80）
}

// NewTableFormatter は新しいTableFormatterを作成
func NewTableFormatter() *TableFormatter {
    return &TableFormatter{
        width: 80,
    }
}

// FormatContextWindowResult はContext Window探索結果を整形
func (tf *TableFormatter) FormatContextWindowResult(result *ContextWindowResult) string {
    var sb strings.Builder

    // ヘッダー
    sb.WriteString("Context Window Probe Results\n")
    sb.WriteString(strings.Repeat("=", 28) + "\n")

    // データ行
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Model:", result.Model))
    sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Estimated Context:", formatNumber(result.MaxContextTokens)))
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Method Confidence:", result.MethodConfidence))
    sb.WriteString(fmt.Sprintf("%-22s %d\n", "Trials:", result.Trials))
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Duration:", formatDuration(result.Duration)))

    if result.MaxInputAtSuccess > 0 {
        sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Max Input at Success:", formatNumber(result.MaxInputAtSuccess)))
    }

    sb.WriteString("\n")

    // ステータス
    if result.Success {
        sb.WriteString("Status: ✓ Success\n")
    } else {
        sb.WriteString("Status: ✗ Failed\n")
        if result.ErrorMessage != "" {
            sb.WriteString(fmt.Sprintf("Error:  %s\n", result.ErrorMessage))
        }
    }

    return sb.String()
}

// FormatMaxOutputResult はMax Output探索結果を整形
func (tf *TableFormatter) FormatMaxOutputResult(result *MaxOutputResult) string {
    var sb strings.Builder

    // ヘッダー
    sb.WriteString("Max Output Tokens Probe Results\n")
    sb.WriteString(strings.Repeat("=", 32) + "\n")

    // データ行
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Model:", result.Model))
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Max Output Tokens:", formatNumber(result.MaxOutputTokens)))
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Evidence:", result.Evidence))
    sb.WriteString(fmt.Sprintf("%-22s %d\n", "Trials:", result.Trials))
    sb.WriteString(fmt.Sprintf("%-22s %s\n", "Duration:", formatDuration(result.Duration)))

    if result.MaxSuccessfullyGenerated > 0 {
        sb.WriteString(fmt.Sprintf("%-22s %s tokens\n", "Max Successfully Gen:", formatNumber(result.MaxSuccessfullyGenerated)))
    }

    sb.WriteString("\n")

    // ステータス
    if result.Success {
        sb.WriteString("Status: ✓ Success\n")
    } else {
        sb.WriteString("Status: ✗ Failed\n")
        if result.ErrorMessage != "" {
            sb.WriteString(fmt.Sprintf("Error:  %s\n", result.ErrorMessage))
        }
    }

    return sb.String()
}

// formatNumber は数値を3桁区切りで整形
func formatNumber(n int) string {
    if n == 0 {
        return "0"
    }

    s := fmt.Sprintf("%d", n)
    var result []rune

    for i, r := range reverse(s) {
        if i > 0 && i%3 == 0 {
            result = append(result, ',')
        }
        result = append(result, r)
    }

    return string(reverse(string(result)))
}

// formatDuration は時間を読みやすく整形
func formatDuration(d time.Duration) string {
    if d < time.Second {
        return fmt.Sprintf("%dms", d.Milliseconds())
    }
    return fmt.Sprintf("%.1fs", d.Seconds())
}

// reverse は文字列を反転
func reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

// FormatVerboseHistory は探索履歴を整形（verbose用）
func (tf *TableFormatter) FormatVerboseHistory(trials []TrialInfo) string {
    var sb strings.Builder

    sb.WriteString("\nSearch History:\n")
    sb.WriteString(strings.Repeat("-", 60) + "\n")
    sb.WriteString(fmt.Sprintf("%-8s %-15s %-12s %s\n", "Trial", "Tokens", "Result", "Message"))
    sb.WriteString(strings.Repeat("-", 60) + "\n")

    for i, trial := range trials {
        status := "✓"
        if !trial.Success {
            status = "✗"
        }

        msg := trial.Message
        if len(msg) > 30 {
            msg = msg[:27] + "..."
        }

        sb.WriteString(fmt.Sprintf("%-8d %-15s %-12s %s\n",
            i+1,
            formatNumber(trial.TokenCount),
            status,
            msg,
        ))
    }

    return sb.String()
}

// TrialInfo は試行情報
type TrialInfo struct {
    TokenCount int
    Success    bool
    Message    string
}
```

#### 2. internal/probe/result.go

結果構造体の拡張

```go
package probe

import "time"

// ContextWindowResult はContext Window探索結果
type ContextWindowResult struct {
    Model              string
    MaxContextTokens   int
    MethodConfidence   string
    Trials             int
    Duration           time.Duration
    MaxInputAtSuccess  int
    Success            bool
    ErrorMessage       string
}

// MaxOutputResult はMax Output探索結果
type MaxOutputResult struct {
    Model                   string
    MaxOutputTokens         int
    Evidence                string
    Trials                  int
    Duration                time.Duration
    MaxSuccessfullyGenerated int
    Success                 bool
    ErrorMessage            string
}
```

### 変更ファイル

#### 1. cmd/llm-info/probe.go

probe-context と probe-max-output の出力部分を修正

**probe-context の変更**:
```go
// 変更前（行185-195付近）:
result := prober.Probe(*model, *verbose)
fmt.Printf("Result: %+v\n", result)

// 変更後:
result, err := prober.Probe(*model, *verbose)
if err != nil {
    fmt.Fprintf(os.Stderr, "Probe failed: %v\n", err)
    os.Exit(1)
}

// テーブル形式で出力
formatter := ui.NewTableFormatter()
output := formatter.FormatContextWindowResult(result)
fmt.Println(output)

// verbose時は履歴も表示
if *verbose && len(result.Trials) > 0 {
    history := formatter.FormatVerboseHistory(result.TrialHistory)
    fmt.Println(history)
}
```

**probe-max-output の変更**:
```go
// 同様にテーブル形式出力に変更
result, err := prober.Probe(*model, *verbose)
if err != nil {
    fmt.Fprintf(os.Stderr, "Probe failed: %v\n", err)
    os.Exit(1)
}

formatter := ui.NewTableFormatter()
output := formatter.FormatMaxOutputResult(result)
fmt.Println(output)

if *verbose && len(result.TrialHistory) > 0 {
    history := formatter.FormatVerboseHistory(result.TrialHistory)
    fmt.Println(history)
}
```

#### 2. internal/probe/context_probe.go

Probe関数の戻り値を拡張

```go
// 変更前:
func (p *ContextWindowProbe) Probe(model string, verbose bool) *ContextWindowResult

// 変更後:
func (p *ContextWindowProbe) Probe(model string, verbose bool) (*ContextWindowResult, error)
```

結果構造体にTrialHistoryフィールドを追加:
```go
type ContextWindowResult struct {
    // 既存フィールド...
    TrialHistory []ui.TrialInfo // 追加
}
```

#### 3. internal/probe/max_output_probe.go

同様にProbe関数の戻り値とTrialHistoryを追加

---

## テスト戦略

### 単体テスト

#### internal/ui/table_formatter_test.go

```go
package ui

import (
    "strings"
    "testing"
    "time"
)

func TestTableFormatter_FormatContextWindowResult(t *testing.T) {
    formatter := NewTableFormatter()

    result := &ContextWindowResult{
        Model:              "GLM-4.6",
        MaxContextTokens:   127000,
        MethodConfidence:   "high",
        Trials:             12,
        Duration:           45*time.Second + 300*time.Millisecond,
        MaxInputAtSuccess:  126800,
        Success:            true,
        ErrorMessage:       "",
    }

    output := formatter.FormatContextWindowResult(result)

    // 期待される要素が含まれることを確認
    if !strings.Contains(output, "GLM-4.6") {
        t.Error("Output should contain model name")
    }
    if !strings.Contains(output, "127,000") {
        t.Error("Output should contain formatted context tokens")
    }
    if !strings.Contains(output, "✓ Success") {
        t.Error("Output should contain success status")
    }
    if !strings.Contains(output, "45.3s") {
        t.Error("Output should contain formatted duration")
    }
}

func TestFormatNumber(t *testing.T) {
    tests := []struct {
        input    int
        expected string
    }{
        {0, "0"},
        {123, "123"},
        {1234, "1,234"},
        {123456, "123,456"},
        {1234567, "1,234,567"},
    }

    for _, tt := range tests {
        result := formatNumber(tt.input)
        if result != tt.expected {
            t.Errorf("formatNumber(%d) = %s; want %s", tt.input, result, tt.expected)
        }
    }
}

func TestFormatDuration(t *testing.T) {
    tests := []struct {
        input    time.Duration
        expected string
    }{
        {500 * time.Millisecond, "500ms"},
        {1 * time.Second, "1.0s"},
        {45*time.Second + 300*time.Millisecond, "45.3s"},
    }

    for _, tt := range tests {
        result := formatDuration(tt.input)
        if result != tt.expected {
            t.Errorf("formatDuration(%v) = %s; want %s", tt.input, result, tt.expected)
        }
    }
}
```

### 統合テスト（手動）

```bash
# ビルド
cd /Users/y-araki/Playground/llm-info-v1_0
go build -o llm-info cmd/llm-info/*.go

# テスト1: Context Window探索のテーブル出力
echo "Test 1: Context Window table output"
./llm-info probe-context --model GLM-4.6 \
    --config test/env/llm-info.yaml

# 期待結果:
# - テーブル形式で表示
# - 数値がカンマ区切り
# - ✓ Success マーク

# テスト2: verbose モードでの詳細表示
echo "Test 2: Verbose mode"
./llm-info probe-context --model GLM-4.6 \
    --config test/env/llm-info.yaml \
    --verbose

# 期待結果:
# - テーブル出力に加えて探索履歴表示
# - 各試行の成功/失敗

# テスト3: Max Output探索のテーブル出力
echo "Test 3: Max Output table output"
./llm-info probe-max-output --model GLM-4.6 \
    --config test/env/llm-info.yaml

# 期待結果:
# - テーブル形式で表示
# - Evidence行が表示される

# テスト4: エラー時のテーブル出力
echo "Test 4: Error output"
./llm-info probe-context --model invalid-model \
    --config test/env/llm-info.yaml

# 期待結果:
# - テーブル形式でエラー表示
# - ✗ Failed マーク
# - エラーメッセージ表示
```

---

## 見積もり

**ストーリーポイント**: 2（約4時間）

内訳:
- TableFormatter実装: 90分
- 結果構造体拡張: 30分
- probe.go統合: 45分
- 単体テスト作成: 45分
- 統合テスト実行: 30分
- ドキュメント更新: 30分

---

## リスク評価

### リスク1: 日本語と英語の整列問題

**確率**: Medium
**影響**: Low
**シナリオ**: 日本語文字とASCII文字の幅が異なり、整列が崩れる

**軽減策**:
- 固定幅フォントを前提とした設計
- 必要に応じてrunewidth パッケージを使用
- テストで日本語を含むケースを検証

### リスク2: ターミナル幅の違い

**確率**: Low
**影響**: Low
**シナリオ**: 80文字未満のターミナルでレイアウトが崩れる

**軽減策**:
- 80文字を基準に設計（最も一般的）
- 必要に応じてターミナル幅を検出

### リスク3: カラー出力の非対応

**確率**: Low
**影響**: Low
**シナリオ**: カラーコードが表示される環境

**軽減策**:
- v2.0ではカラー出力なし（記号のみ使用）
- v2.1でカラー対応を検討

### 総合リスク: LOW

---

## Definition of Done

### コード
- [ ] internal/ui/table_formatter.go を作成
- [ ] internal/probe/result.go を作成
- [ ] cmd/llm-info/probe.go を修正（probe-context）
- [ ] cmd/llm-info/probe.go を修正（probe-max-output）
- [ ] internal/probe/context_probe.go を修正
- [ ] internal/probe/max_output_probe.go を修正
- [ ] コードレビュー完了
- [ ] 全変更がコミット済み

### テスト
- [ ] internal/ui/table_formatter_test.go を作成
- [ ] 全単体テストがパスする
- [ ] 統合テスト（手動）で4つのテストケースが成功
- [ ] 日本語を含む出力が正しく整列される

### 動作確認
- [ ] probe-context がテーブル形式で出力される
- [ ] probe-max-output がテーブル形式で出力される
- [ ] --verbose で詳細情報が表示される
- [ ] エラー時も適切にフォーマットされる
- [ ] 数値が3桁区切りで表示される

### ドキュメント
- [ ] このPBIファイルに実装記録を追記
- [ ] plans/v2_0-rp/README.md を更新
- [ ] USAGE.md に出力例を追加

---

## 次のステップ

このPBI完了後:
1. v2.0リリース判定（PBI-2-0-005 + PBI-2-0-006完了）
2. 最終E2Eテスト
3. ドキュメント整備
4. v2.0リリース

---

## 備考

### 設計判断

#### なぜJSONでなくテーブル形式を優先？

ユーザーフィードバック：「CLIで直接確認したい」
- 開発者が結果を即座に理解できる
- パイプ処理よりも直接確認が優先
- JSON出力はv2.1で追加予定

#### なぜカラー出力なし？

- v2.0は最小限の機能で早期リリース
- 記号（✓/✗）でも十分な可読性
- カラー対応はv2.1で検討

---

## 実装記録

### 2026-01-11 (予定)

**実装者**: (To be filled)

**実装内容**: (To be filled)

**遭遇した問題と解決策**: (To be filled)

**テスト結果**: (To be filled)

**備考**: (To be filled)
