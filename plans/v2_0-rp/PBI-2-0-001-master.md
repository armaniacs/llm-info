# プロダクトバックログアイテム: v2.0モデル制約値推定機能

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 1

## ユーザーストーリー

LLMアプリケーション開発者として、APIの実挙動からモデルの制約値を推定する機能がほしい、なぜなら公式ドキュメントにない正確な制約値を取得して、本番環境でのエラーやパフォーマンス問題を防ぎたいから

## ビジネス価値

- **開発効率向上**: 手動での制約値調査時間を大幅に削減（数時間から数分へ）
- **エラ削減**: 本番環境におけるcontext overflowエラーの未然防止
- **コスト最適化**: モデル仕様に基づいた適切な選択による課金最適化
- **品質保証**: CI/CDパイプラインでの制約値自動検証の実現

## BDD受け入れシナリオ

### メインシナリオ: Context Window推定

```gherkin
Scenario: モデルのcontext windowを正常に推定する
  Given ユーザーが有効なAPIキーを設定している
  And "gpt-4o"モデルを指定してllm-info probeを実行する
  When システムが二分探索でcontextの上限に到達する
  Then 推定結果がJSON形式で出力される
  And "estimated_max_context_tokens"が128000のような数値で含まれる
  And "method_confidence"が"high"または"medium"で含まれる

Scenario: 推定過程の詳細をverboseモードで表示する
  Given ユーザーが"--verbose"オプションを指定する
  When context window推定が実行される
  Then 各試行のログが逐次表示される
  And 成功/失敗の判定理由が表示される
  And 最終的な推計結果が要約される

Scenario: テーブル形式で結果を出力する
  Given ユーザーが"--format table"オプションを指定する
  When 推定が完了する
  Then 結果が見やすいテーブル形式で表示される
  And Model, Estimated Context, Method, Confidenceが列として含まれる
```

### エラーシナリオ

```gherkin
Scenario: APIキーが未設定の場合に即座に失敗する
  Given APIキーが設定されていない
  When llm-info probeが実行される
  Then "API key is required"というメッセージが表示される
  And 終了コード1で即座に終了する

Scenario: 存在しないモデルを指定した場合
  Given ユーザーが"invalid-model"を指定する
  When 探索が開始される
  Then "Model not found"エラーが表示される
  And APIコールは1回で停止する

Scenario: コスト制限を超過しそうな場合
  Given 試行回数が40回に達した
  まだ収束していない場合
  Then 試行を停止する
  And "Max trials reached"という警告を表示する
  And 途中経過を含む結果を出力する
```

## 受け入れ基準

### 機能要件
- [ ] `--model`で指定したモデルのcontext windowを推定できる
- [ ] max output tokensの上限を観測できる
- [ ] 二分探索で効率的に収束する（128トークン以内の精度）
- [ ] API応答に含まれるtoken情報を抽出して利用できる

### 安全要件
- [ ] APIキー未設定時に即座に失敗する
- [ ] 最大試行回数（デフォルト40）を超えない
- [ ] HTTPタイムアウト（30秒）で自動停止する
- [ ] 入力テキスト内容はログに残さない

### 出力要件
- [ ] JSON形式とテーブル形式で出力できる
- [ ] verboseモードで詳細な試行ログを表示できる
- [ ] 結果ファイルに出力できる（--outオプション）

### 品質要件
- [ ] 温度パラメータ0で再現性を確保する
- [ ] 乱数を使用しない決定論的な挙動
- [ ] エラーメッセージから上限値を正しく抽出できる

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- llm-info probe --model gpt-4o-miniで正常ケースを検証
- llm-info probe --model invalidでエラーケースを検証
- llm-info probe --model gpt-4o --format tableで形式指定を検証

統合テスト:
- OpenAI APIクライアントのモックを使用
- エラーレスポンスのパース処理をテスト
- 二分探索ロジックの単体テスト

単体テスト:
- TokenEstimatorのビジネスロジック
- TestDataGeneratorのテストデータ生成
- BoundarySearcherの収束判定
- 各種エラーハンドリング
```

## 実装アプローチ

### Outside-Inアプローチ
1. **外側**: CLIコマンド定義と引数パース
2. **中間**: APIクライアントとレスポンス処理
3. **内側**: 探索アルゴリズムとテストデータ生成

### TDDサイクル
1. **Red**: 最低限のCLIとモックAPIで失敗テストを作成
2. **Green**: OpenAI API実際の呼び出しで成功させる
3. **Refactor**: アルゴリズムを改良し、速度を最適化

### リファクタリング重点項目
- 二分探索ロジックの抽象化
- エラーハンドリングの共通化
- 出力形式ロジックの分離

## 見積もり

**ストーリーポイント**: 21（約2スプリント）

内訳:
- 基本API連携: 8
- Context Window探索: 8
- Max Output Tokens探索: 5

## 技術的考慮事項

### 依存関係
- Go 1.21+
- OpenAI Go SDK
- 設定管理（既存のconfigパッケージ）

### テスタビリティ
- HTTPクライアントはinterfaceとして定義
- テスト用にモックサーバーを準備
- 確定的なテストデータ生成関数

### パフォーマンス
- 各APIコール間に1秒間隔を設ける（レート制限対策）
- 再試行は指数バックオフで実装
- 並列実行は行わない（レート制限回避）

## Definition of Done

- [ ] 全てのBDDシナリオが自動テストでパス
- [ ] テストカバレッジが90%以上
- [ ] レビュー済みのコードがmainブランチにマージ済み
- [ ] ユーザードキュメント（USAGE.md）更新済み
- [ ] ベンチマークテスト（性能測定）済み
- [ ] 実機でのE2Eテスト済み（OpenAI API使用）
- [ ] CI/CDでの自動テスト実行済み

## 分割計画

このPBIは以下のサブPBIに分割して実装する：

1. **PBI-2-0-002**: 技術調査 - OpenAI API基本連携
2. **PBI-2-0-003**: Context Window推定アルゴリズム実装
3. **PBI-2-0-004**: Max Output Tokens推定実装
4. **PBI-2-0-005**: 出力形式とUX改善
5. **PBI-2-0-006**: Needle In A Haystack機能
6. **PBI-2-0-007**: パフォーマンス最適化とエラーハンドリング

## 将来の拡張（v2.1以降）

- tiktokenによる事前計算オプション
- 複数モデルの一括測定
- 推定結果のキャッシュ機能
- Web UIでの可視化