# プロダクトバックログアイテム: Max Output Tokens推定実装

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 4

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能
PBI-2-0-002: 技術調査 - OpenAI API基本連携
PBI-2-0-003: Context Window推定アルゴリズム実装

## ユーザーストーリー

LLM開発者として、モデルの最大出力トークン数を知りたい、なから長文生成を要求した際の上限を把握し、適切なプロンプト設計をしたいから

## ビジネス価値

- **品質保証**: レスポンスが途中で途切れるのを防止
- **UX向上**: ユーザーに不足感を与えないレスポンス設計
- **コスト最適化**: 不要なリトライ回数を削減

## BDD受け入れシナリオ

### 基本推定シナリオ

```gherkin
Scenario: validationエラーからmax output tokensを検出する
  Given gpt-4o-miniモデルを使用する
  When max_output_tokensに100000を設定してリクエストする
  Then "max_output_tokens must be <= 16384"エラーを受信する
  And 16384という値を正しく抽出する
  And それを推定値として採用する

Scenario: incomplete statusから出力制限を検出する
  Then 1000トークンの出力を要求するが500で打ち切られる
  And APIレスポンスのstatusが"incomplete"である
  Then incomplete_details.reasonが"max_output_tokens"である
  And 実際の出力トークン数を記録する
```

### 探索戦略シナリオ

```gherkin
Scenario: 徐々に出力要求を増やす探索を実行する
  Given 初期値256トークンから探索を開始する
  When 各試行でx2に増やしていく
  Then 8192で失敗した場合、4096と8192の間で二分探索する
  And 4096が成功で5120が失敗ならば4608を試行する
  And 最終的に正確な上限値が特定される

Scenario: 小さなモデルでも効率的に探索する
  Given 2048トークン未満の上限を持つモデルを指定する
  When 探索を開始する
  Then 256から始めて512で打ち切られる
  And 384と256の間で確定する
```

## 受け入れ基準

### 検出要件
- [ ] APIバリデーションエラーから数値を抽出できる
- [ ] incomplete statusを検出してreasonを取得できる
- [ ] 両方の方法で検出できた場合はvalidation_errorを優先する
- [ ] 検出不能な場合はnullと理由を記録する

### 探索要件
- [ ] 256から開始して指数的に増やす
- [ ] 境界を特定後、二分探索で収束させる
- [ ] 総試行回数はcontext window探索より少なくする
- [ ] 各試行で十分な長さの入力を与える

### 品質要件
- [ ] 同じ条件で常に同じ結果になる
- [ ] 推定値は実際の値から±16以内の精度
- [ ] methodとconfidenceを適切に設定する

### 出力要件
- [ ] estimated_max_output_tokensを数値で出力する
- [ ] evidence（validation_error/max_output_incomplete）を記録
- [ ] observed_incomplete_reasonを記録する
- [ ] 実際に生成できた最大トークン数を記録する

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- gpt-4o-miniで上限16Kの検証
- 小さいモデルでの探索挙動確認
- --context-windowと同時実行のテスト

統合テスト:
- バリデーションエラーのモックテスト
- incompleteレスポンスのモックテスト
- 二分探索の収束ロジックテスト

単体テスト:
- MaxTokensProbeのメインロジック
- ValidationExtractorのエラー抽出
- IncompleteDetectorのstatus解析
- OutputResultFormatterの整形
```

## 実装アプローチ

### コアロジック
```go
type MaxOutputTokensProbe struct {
    client     OpenAIClient
    generator  TestDataGenerator
    searcher   BoundarySearcher
}

func (p *MaxOutputTokensProbe) Probe(model string) (Result, error) {
    // 1. 十分な入力長を確保する（context windowの半分）
    // 2. 256から指数探索で上限を見つける
    // 3. 二分探索で精度を上げる
    // 4. 検証方法を判定して記録
}
```

### 入力テキスト戦略
- Context Window推定結果を利用して入力長を決定する
- 入力長: 推定されたcontext windowの50%を使用（出力に十分な領域を確保）
- Context Window推定がまだの場合は、モデルごとの既知値を使用（gpt-4o: 127000, gpt-3.5-turbo: 16000など）
- 入力テキストはContext Windowと同じテストデータを流用する

### 探索パラメータ
- 初期値: 256 tokens
- 増加率: x2
- 収束条件: 差分 ≤ 16
- 最小入力: 推定context windowの30%

## 技術仕様

### APIエラーメッセージ
```regex
max_output_tokens must be <= (\d+)
maximum output tokens is (\d+)
the value of max_output_tokens should be <= (\d+)
```

### APIレスポンス構造
```json
{
  "id": "resp_...",
  "status": "incomplete",
  "incomplete_details": {
    "reason": "max_output_tokens"
  },
  "usage": {
    "prompt_tokens": 1000,
    "completion_tokens": 1024,
    "total_tokens": 2024
  }
}
```

### 結果フォーマット
```json
{
  "estimated_max_output_tokens": 16384,
  "evidence": "validation_error",
  "observed_incomplete_reason": "max_output_tokens",
  "max_successfully_generated": 8192
}
```

## 見積もり

**ストーリーポイント**: 5

内訳:
- APIレスポンス解析: 2
- 探索ロジック実装: 2
- 結果整形: 1

## 技術的考慮事項

### 新規ファイル
- `internal/probe/max_output_probe.go`: max output探索
- `internal/probe/validation_extractor.go`: エラー抽出
- `internal/probe/incomplete_detector.go`: status検出

### 既存ファイル活用
- `boundary_searcher.go`: 共通の二分探索ロジック
- `data_generator.go`: 入力テキスト生成

### 特殊考慮
- max_output_tokensが大きいモデルではコスト上昇に注意
- 入力が不足するケースのハンドリング
- completion_tokensの正確な計算

## Definition of Done

- [ ] gpt-4o-miniで16384±16が検出できる
- [ ] バリデーションエラーとincomplete両方に対応
- [ ] context windowとの同時実行で競合しない
- [ ] 総APIコール回数が15回以内

## 次のPBIへの準備

- [ ] 抽出ロジックがNeedle position探索で流用可能
- [ ] 各種パラメータが設定ファイルで変更可能
- [ ] 結果のマージ処理が実装できていること

## 注意事項

- max_output_tokensはモデルによっては設定されない場合がある
- 出力制限はユーザー毎、時間毎に変わる可能性がある（記録必須）
- max_output_tokensが非常に大きいモデルでは上限テストを避ける配慮が必要