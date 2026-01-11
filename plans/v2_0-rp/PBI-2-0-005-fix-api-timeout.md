# プロダクトバックログアイテム: API呼び出しタイムアウト問題の修正

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 2（最高優先度）

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## 依存PBI
PBI-2-0-002: 技術調査 - OpenAI API基本連携
PBI-2-0-003: Context Window推定アルゴリズム実装
PBI-2-0-004: Max Output Tokens推定実装

## ユーザーストーリー

開発者として、probe コマンドが無限にハングしないようにしたい、なぜならタイムアウト後に適切にエラーを返し、ツールを実用可能にしたいから

## ビジネス価値

- **実用性確保**: ツールが実際に使えるようになる（現在は使用不可）
- **信頼性向上**: 予測可能な動作により、本番環境での採用が可能に
- **開発効率**: デバッグ時間の削減（無限ハングによる時間浪費を防止）
- **v2.0リリース**: この修正なしではv2.0をリリースできない

## 問題の背景

### 現在の状況
```
$ llm-info probe-context --model GLM-4.6 --config test/env/llm-info.yaml
Probing context window...
[無限にハング、Ctrl+Cで強制終了するしかない]
```

### 根本原因（調査済み）

#### Critical Issue 1: タイムアウトなしのContext
**ファイル**: `internal/api/probe_client.go:106-107`

```go
httpReq, err := http.NewRequestWithContext(
    context.Background(),  // ← タイムアウトなし！
    "POST",
    url,
    bytes.NewBuffer(jsonBody),
)
```

**問題**: `context.Background()`には有効期限がなく、ネットワーク問題時に無限に待機

#### High Priority Issue 2: Nil Pointer Dereference
**ファイル**: `internal/probe/context_probe.go:145`

```go
return &BoundarySearchResult{
    Value: response.Usage.PromptTokens,  // ← response.Usageがnilの場合panic
    ...
}
```

**問題**: APIレスポンスにUsageフィールドがない場合にpanicが発生し、エラーが適切に伝播しない

## BDD受け入れシナリオ

```gherkin
Scenario: 設定されたタイムアウト時間で処理が停止する
  Given タイムアウトを5秒に設定する
  When probe-contextコマンドを実行する
  And APIレスポンスが遅延する
  Then 5秒後にタイムアウトエラーが表示される
  And プロセスが正常終了する
  And 無限にハングしない

Scenario: タイムアウト時に明確なエラーメッセージを表示
  Given APIが応答しない状態
  When probe-contextコマンドを実行する
  Then "context deadline exceeded"エラーメッセージが表示される
  And 終了コード1で終了する

Scenario: Usageフィールドがnilの場合でもpanicしない
  Given APIレスポンスにUsageフィールドがない
  When probe-contextコマンドを実行する
  Then "Response missing usage information"エラーが表示される
  And プログラムがpanicしない
  And 適切にエラーハンドリングされる

Scenario: 正常なレスポンスは引き続き動作する
  Given 正常なAPIレスポンスが返る
  When probe-contextコマンドを実行する
  Then 探索が正常に完了する
  And 推定結果が表示される
  And エラーが発生しない
```

## 受け入れ基準

### 機能要件
- [ ] probe-context コマンドが設定されたタイムアウト時間で停止する
- [ ] probe-max-output コマンドが設定されたタイムアウト時間で停止する
- [ ] タイムアウト時に"context deadline exceeded"エラーを表示する
- [ ] response.Usage が nil の場合でも panic しない
- [ ] 正常なAPIレスポンスは引き続き動作する

### 品質要件
- [ ] タイムアウトは設定ファイルおよびCLI引数で変更可能
- [ ] エラーメッセージがユーザーフレンドリー
- [ ] 既存の動作に regression がない
- [ ] 単体テストがパスする

## 実装内容

### 変更1: ProbeClient にタイムアウトContextを追加

**ファイル**: `/Users/y-araki/Playground/llm-info-v1_0/internal/api/probe_client.go`

#### Import の追加（行3-11）

```go
import (
    "bytes"
    "context"  // ← 追加
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/armaniacs/llm-info/pkg/config"
)
```

#### ProbeModel 関数の修正（行104-114）

**変更前**:
```go
// HTTPリクエストを作成
url := fmt.Sprintf("%s/v1/chat/completions", pc.config.BaseURL)
httpReq, err := http.NewRequestWithContext(
    context.Background(),
    "POST",
    url,
    bytes.NewBuffer(jsonBody),
)
```

**変更後**:
```go
// HTTPリクエストを作成（タイムアウト付き）
url := fmt.Sprintf("%s/v1/chat/completions", pc.config.BaseURL)

// タイムアウト付きContextを作成
ctx, cancel := context.WithTimeout(context.Background(), pc.config.Timeout)
defer cancel()

httpReq, err := http.NewRequestWithContext(
    ctx,  // タイムアウト付きContextを使用
    "POST",
    url,
    bytes.NewBuffer(jsonBody),
)
```

**理由**:
- `context.WithTimeout`により、設定されたタイムアウト時間で自動的にリクエストがキャンセルされる
- `defer cancel()`でリソースリークを防止
- Go標準のタイムアウトパターン

**リスク**: LOW - Go標準パターン、広く使用されている

---

### 変更2: Context Probe に nil チェックを追加

**ファイル**: `/Users/y-araki/Playground/llm-info-v1_0/internal/probe/context_probe.go`

#### testWithTokenCount 関数の修正（行143-151）

**変更前**:
```go
// 成功した場合
return &BoundarySearchResult{
    Value:        response.Usage.PromptTokens,
    Success:      true,
    ErrorMessage: "",
    Source:        "success",
    Trials:        1,
    EstimatedTokens: response.Usage.TotalTokens,
}, nil
```

**変更後**:
```go
// 成功した場合 - Usageフィールドのnilチェック
if response.Usage == nil {
    return &BoundarySearchResult{
        Value:           0,
        Success:         false,
        ErrorMessage:    "Response missing usage information",
        Source:          "api_error",
        Trials:          1,
        EstimatedTokens: 0,
    }, nil
}

return &BoundarySearchResult{
    Value:           response.Usage.PromptTokens,
    Success:         true,
    ErrorMessage:    "",
    Source:          "success",
    Trials:          1,
    EstimatedTokens: response.Usage.TotalTokens,
}, nil
```

**理由**:
- APIレスポンスが不完全な場合でもpanicを防止
- 適切なエラーメッセージでデバッグが容易に
- 防御的プログラミングのベストプラクティス

**リスク**: LOW - 既存動作に影響なし（Usageがあれば従来通り動作）

**注**: max_output_probe.go は既に184行目で`if response.Usage != nil`チェックが実装済みのため修正不要

---

## テスト戦略

### 単体テスト

#### 新規ファイル: `internal/api/probe_client_test.go`

```go
package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/armaniacs/llm-info/pkg/config"
)

func TestProbeClient_TimeoutWorks(t *testing.T) {
    // タイムアウトが実際に機能することをテスト
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(5 * time.Second) // 遅いAPIをシミュレート
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    cfg := &config.AppConfig{
        BaseURL: server.URL,
        APIKey:  "test",
        Timeout: 100 * time.Millisecond, // 短いタイムアウト
    }

    client := NewProbeClient(cfg)
    _, err := client.ProbeModel("test-model")

    // タイムアウトエラーが発生することを確認
    if err == nil {
        t.Error("Expected timeout error, got nil")
    }
    if !strings.Contains(err.Error(), "deadline exceeded") &&
       !strings.Contains(err.Error(), "context deadline") {
        t.Errorf("Expected deadline exceeded error, got: %v", err)
    }
}

func TestProbeClient_NormalRequestStillWorks(t *testing.T) {
    // 正常なリクエストが引き続き動作することをテスト
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        resp := ProbeResponse{
            ID:      "test-123",
            Model:   "test-model",
            Choices: []ChatChoice{{FinishReason: "stop"}},
            Usage: &Usage{
                PromptTokens:     10,
                CompletionTokens: 5,
                TotalTokens:      15,
            },
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()

    cfg := &config.AppConfig{
        BaseURL: server.URL,
        APIKey:  "test",
        Timeout: 10 * time.Second,
    }

    client := NewProbeClient(cfg)
    result, err := client.ProbeModel("test-model")

    // 正常に完了することを確認
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    if result == nil {
        t.Error("Expected result, got nil")
    }
    if result.Usage == nil {
        t.Error("Expected Usage field, got nil")
    }
    if result.Usage.PromptTokens != 10 {
        t.Errorf("Expected PromptTokens=10, got %d", result.Usage.PromptTokens)
    }
}
```

#### 新規ファイル: `internal/probe/context_probe_test.go`

```go
package probe

import (
    "testing"

    "github.com/armaniacs/llm-info/internal/api"
)

func TestContextProbe_NilUsageHandling(t *testing.T) {
    // response.Usageがnilでもpanicしないことをテスト
    prober := &ContextWindowProbe{}

    // nil Usageのレスポンス
    response := &api.ProbeResponse{
        ID:     "test",
        Usage:  nil, // nil!
        Choices: []api.ChatChoice{{FinishReason: "stop"}},
    }

    // testWithTokenCountの内部ロジックをシミュレート
    if response.Usage == nil {
        // エラーハンドリングが正しく動作することを確認
        result := &BoundarySearchResult{
            Value:           0,
            Success:         false,
            ErrorMessage:    "Response missing usage information",
            Source:          "api_error",
            Trials:          1,
            EstimatedTokens: 0,
        }

        if result.Success {
            t.Error("Expected Success=false when Usage is nil")
        }
        if result.ErrorMessage != "Response missing usage information" {
            t.Errorf("Expected specific error message, got: %s", result.ErrorMessage)
        }
    }
}
```

### 統合テスト（手動）

```bash
# テスト環境
cd /Users/y-araki/Playground/llm-info-v1_0

# ビルド
go build -o llm-info cmd/llm-info/*.go

# テスト1: タイムアウトが動作することを確認
echo "Test 1: Timeout behavior"
./llm-info probe-context --model GLM-4.6 \
    --config test/env/llm-info.yaml \
    --timeout 5s \
    --verbose

# 期待結果:
# - 5秒後にタイムアウトエラー表示
# - 無限ハングしない
# - "context deadline exceeded" メッセージ

# テスト2: 正常動作の確認
echo "Test 2: Normal operation"
./llm-info probe-context --model GLM-4.6 \
    --config test/env/llm-info.yaml \
    --verbose

# 期待結果:
# - 探索が正常に完了
# - 推定結果が表示される
# - エラーなし

# テスト3: Max Output探索
echo "Test 3: Max output probe"
./llm-info probe-max-output --model GLM-4.6 \
    --config test/env/llm-info.yaml \
    --verbose

# 期待結果:
# - 探索が正常に完了
# - 推定結果が表示される
# - エラーなし

# テスト4: 短いタイムアウトでの動作確認
echo "Test 4: Short timeout"
./llm-info probe-context --model GLM-4.6 \
    --config test/env/llm-info.yaml \
    --timeout 1s

# 期待結果:
# - 1秒後にタイムアウト
# - 適切なエラーメッセージ

# テスト5: 複数回実行してpanic確認
echo "Test 5: Panic check (5 runs)"
for i in {1..5}; do
    echo "Run $i"
    ./llm-info probe-context --model GLM-4.6 \
        --config test/env/llm-info.yaml || echo "Run $i failed gracefully"
done

# 期待結果:
# - panicが発生しない
# - エラーは適切にハンドリングされる
```

---

## 見積もり

**ストーリーポイント**: 1（約75分）

内訳:
- コード変更: 15分
  - probe_client.go: 5分
  - context_probe.go: 5分
  - import追加: 1分
  - 動作確認: 4分
- 単体テスト作成: 30分
- 統合テスト実行: 20分
- ドキュメント更新: 10分

---

## リスク評価

### リスク1: タイムアウトが早すぎる

**確率**: Medium
**影響**: Medium
**シナリオ**: 正常なAPIレスポンスがタイムアウトにより失敗

**軽減策**:
- デフォルトタイムアウトを30秒に設定（現在のconfig.Timeout値を維持）
- CLI引数でタイムアウトを調整可能にする
- テストで様々なタイムアウト値を検証

### リスク2: nil チェックによる false negative

**確率**: Low
**影響**: Low
**シナリオ**: 実際にはUsageがあるべきなのに、APIが返さないケース

**軽減策**:
- エラーメッセージを明確にして原因特定を容易に
- verboseモードでレスポンス全体をログ出力
- APIプロバイダーのドキュメントを確認

### リスク3: 既存動作の regression

**確率**: Low
**影響**: Medium
**シナリオ**: 変更により正常だったケースが動作しなくなる

**軽減策**:
- 既存の統合テストを全て実行
- 手動E2Eテストで正常動作を確認
- Git で差分を慎重にレビュー

### 総合リスク: LOW
- 変更範囲が限定的（2ファイル、合計20行程度）
- Go標準パターンの使用
- 防御的プログラミング（既存動作に影響なし）

---

## ロールバック計画

### 問題が発生した場合

```bash
cd /Users/y-araki/Playground/llm-info-v1_0

# 変更を一時保存
git diff HEAD > /tmp/timeout-fix.patch

# 問題が発生したらロールバック
git checkout HEAD -- internal/api/probe_client.go \
    internal/probe/context_probe.go

# 必要に応じて再適用
git apply /tmp/timeout-fix.patch
```

### ロールバックが必要なシナリオ

1. **タイムアウトが頻発する**: デフォルトタイムアウト値を調整
2. **nil チェックで常に失敗**: nilチェックロジックを見直し
3. **予期しない動作**: 完全ロールバックして再設計

---

## Definition of Done

### コード
- [ ] probe_client.go に context.WithTimeout を追加
- [ ] probe_client.go に context パッケージをimport
- [ ] context_probe.go に response.Usage の nil チェックを追加
- [ ] コードレビュー完了
- [ ] 全変更がコミット済み

### テスト
- [ ] 単体テスト (probe_client_test.go) を作成
- [ ] 単体テスト (context_probe_test.go) を作成
- [ ] 全単体テストがパスする
- [ ] 統合テスト（手動）で5つのテストケースが成功
- [ ] タイムアウト動作を確認
- [ ] panic が発生しないことを確認

### 動作確認
- [ ] probe-context が実際のAPIで動作する
- [ ] probe-max-output が実際のAPIで動作する
- [ ] タイムアウト時に適切なエラーメッセージが表示される
- [ ] 正常なレスポンスは引き続き動作する

### ドキュメント
- [ ] このPBIファイルに実装記録を追記
- [ ] plans/v2_0-rp/README.md を更新
- [ ] 必要に応じてUSAGE.mdを更新

---

## 次のステップ

このPBI完了後:
1. **PBI-2-0-006**: テーブル形式での結果出力を実装
2. v2.0リリース判定
3. ドキュメント整備
4. v2.0リリース

---

## 備考

### なぜこのPBIが最高優先度か

1. **ブロッカー**: この修正なしではツールが使用不可
2. **短時間で完了**: 75分で実装可能
3. **高い価値**: ツールを実用可能にする
4. **低リスク**: 変更範囲が限定的

### 実装時の注意点

- タイムアウト値はconfig.Timeoutを使用（既存設定との整合性）
- エラーメッセージはユーザーフレンドリーに
- verboseモードでの詳細ログ出力を考慮
- 既存のテストが全てパスすることを確認

---

## 実装記録

### 2026-01-11 (予定)

**実装者**: (To be filled)

**実装内容**: (To be filled)

**遭遇した問題と解決策**: (To be filled)

**テスト結果**: (To be filled)

**備考**: (To be filled)
