# プロダクトバックログアイテム: Context Window推定アルゴリズム実装

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 3

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能
PBI-2-0-002: 技術調査 - OpenAI API基本連携

## ユーザーストーリー

LLM開発者として、モデルの最大コンテキスト長を正確に知りたい、なぜならcontext overflowエラーを防ぎ、適切な入力サイズを設計したいから

## ビジネス価値

- **エラ防止**: 本番でのcontext overflowエラーを根絶
- **コスト削減**: 不必要に大きなモデル選択を防止
- **パフォーマンス最適化**: 最大限活用できる入力サイズの特定

## BDD受け入れシナリオ

### 基本探索シナリオ

```gherkin
Scenario: 二分探索でcontext windowを推定する
  Given gpt-4oモデルのcontext windowが128000未満であることが知られている
  When 4Kから開始して青空文庫の「吾輩は猫である」などの著作権フリー文章を徐々に増やす探索を実行する
  Then 探索は127000付近で収束する
  And 誤差は128トークン以内である
  And "high confidence"の評価がつく

Scenario: エラーメッセージから上限値を抽出する
  Given context windowを超えるリクエストを送信する
  When APIが"maximum context length is 128000 tokens"エラーを返す
  Then 128000という数字を正しく抽出する
  And この値を探索の上界として利用する

Scenario: 小さいモデルの探索を効率的に行う
  Given 最大contextが8192以下のモデルを指定する
  When 探索を開始する
  Then 初期値は4096から始まる
  And 探索は5回以内で完了する
```

### 境界値テストシナリオ

```gherkin
Scenario: 直前の成功値が探索結果となる
  Given 127000トークンで成功する
  And 128100トークンで失敗する
  When 二分探索で127500を試行する
  Then 127500が成功すればこれが最終値となる
  And 127500が失敗すれば125250を次の試行値とする

Scenario: API使用量を計算して記録する
  Given 探索が10回の試行で完了した
  When 各試行で1トークン$0.000001のコストがかかる
  Then 総コストを計算してログに表示する
  And コスト警告（$0.05超過）が出ないことを確認する
```

## 受け入れ基準

### アルゴリズム要件
- [ ] 指数探索（4K, 8K, 16K...）で上限を特定する
- [ ] 二分探索で精度を128トークン以内に収束させる
- [ ] APIエラーメッセージから数値を正規表現で抽出する
- [ ] 成功時のトークン数をusage.prompt_tokensから取得する

### テストデータ要件
- [ ] 意味のある日本語文章を生成する
- [ ] トークン数を指定して正確なサイズで生成する
- [ ] 同じ内容で常に同じトークン数になること

### 品質要件
- [ ] 同じ条件で実行しても常に同じ結果になる
- [ ] maximum trials（40）を超えないように停止する
- [ ] 探索履歴をverboseモードで表示できる

### 出力要件
- [ ] estimated_max_context_tokensを数値で出力する
- [ ] method_confidence（high/medium/low）を判定する
- [ ] max_input_tokens_at_successを記録する
- [ ] 探索回数と時間を記録する

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- gpt-4o-miniで実際の推定を実行
- gpt-3.5-turboで小さいモデルをテスト
- --verboseでのログ出力を確認

統合テスト:
- モックAPIで境界値シナリオをテスト
- エラーメッセージ抽出ロジックをテスト
- 二分探索の収束アルゴリズムをテスト

単体テスト:
- BoundarySearcherの各種メソッド
- TokenGeneratorの文章生成ロジック
- TokenCounterでのtoken数算出
- ConfidenceCalculatorの信頼度判定
```

## 実装アプローチ

### コアアルゴリズム
```go
type ContextWindowProbe struct {
    client    OpenAIClient
    generator TestDataGenerator
    searcher  BoundarySearcher
}

// 探索メインロジック
func (p *ContextWindowProbe) Probe(model string) (Result, error) {
    // 1. 指数探索で上限を見つける
    // 2. 二分探索で境界を絞る
    // 3. 結果の信頼度を評価
}
```

### TDD実装段階
1. **Red**: モックAPIで常に失敗するテスト
2. **Green**: 最小成功ケース（固定値を返す）
3. **Refactor**: 実際の探索ロジックを实现

### 最適化ポイント
- 無駄なAPIコールを削減する早期打ち切り
- 徐々に小さくするステップサイズの調整
- レート制限を考慮した待機時間

## 技術仕様

### 探索パラメータ
- 初期値: 4096 tokens
- 指数増加: x2
- 収束条件: 上界と下界の差 ≤ 128
- 最大試行: 40回
- API間隔: 1秒

### テストデータ構成
```
[Preamble] + [Body x N] + [Needle] + [Question]
- Preamble: "以下の内容を記憶してください。"
- Body: 青空文庫「吾輩は猫である」「坊っちゃん」などの著作権フリー文章
- Needle: "本日の日付は2024年1月11日、ラッキーカラーは青です"
- Question: "ラッキーカラーは何色でしたか？"
```

### テストデータの具体例

青空文庫から取得した文章の例：
```
吾輩は猫である。名前はまだ無い。
どこで生れたかとんと見当がつかぬ。
何でも薄暗いじめじめした所でニャーニャー泣いていた事だけは記憶している。
吾輩はここで始めて人間というものを見た。
```

このようなテキストを繰り返し使用し、必要なトークン数まで埋めていく。

### エラーメッセージ抽出
```regex
maximum context length is (\d+) tokens
your request resulted in (\d+) tokens
this model's maximum context length is (\d+) tokens
```

## 見積もり

**ストーリーポイント**: 8

内訳:
- テストデータ生成: 2
- 境界探索アルゴリズム: 3
- エラー抽出と処理: 2
- 結果整形と出力: 1

## 技術的考慮事項

### 新規ファイル
- `internal/probe/context_probe.go`: Context window探索
- `internal/probe/data_generator.go`: テストデータ生成
- `internal/probe/boundary_searcher.go`: 二分探索ロジック
- `internal/probe/result_formatter.go`: 結果整形

### 既存ファイルの変更
- `cmd/llm-info/probe.go`: context windowサブコマンド

### 性能考慮
- Goroutineによる並列化は行わない（レート制限）
- 必要最小限のメモリ使用量
- キャッシュは実装しない（毎回正確性優先）

## Definition of Done

- [ ] gpt-4o-miniの推定で127000±128の結果が出る
- [ ] 探索履歴がverboseモードで追跡可能
- [ ] 全てのエラーパターンで停止できる
- [ ] コストが$0.05以内で完了する
- [ ] テストカバレッジが95%以上

## 次のPBIへの準備

- [ ] BoundarySearcherがmax output tokens探索でも使えること
- [ ] TestDataGeneratorがneedle positionに対応可能な構造であること
- [ ] Resultフォーマットがmax output tokensと共通化できること