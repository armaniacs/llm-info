# テスト修正計画

## 概要
現在失敗しているテストを修正するための計画。最終的にすべてのテストがパスすることを目指す。

## 問題の特定

### 統合テスト (test/integration)
- [ ] `TestLegacyConfigCompatibility` - YAMLパースエラー
- [ ] `TestConfigPriority/partial_CLI_override` - 設定の優先順位問題
- [ ] `TestConfigSourceInfo` - 設定ソース情報の問題
- [ ] `TestConfigValidation/invalid_URL` - URL検証が期待通り動作しない
- [ ] `TestEnvironmentVariableValidation/invalid_timeout` - タイムアウト検証問題
- [ ] `TestEnvironmentVariableValidation/invalid_output_format` - 出力形式検証問題

### E2Eテスト (test/e2e)
- [ ] `TestAdvancedDisplayFeatures/JSON_output_with_filter` - フラグ名不一致 (-output vs --format)
- [ ] `TestAdvancedDisplayFeatures/help_output` - ヘルプ出力の期待値不一致
- [ ] `TestAdvancedDisplayWorkflow/complete_workflow` - フラグ名不一致
- [ ] 複数のテストでバイナリが見つからない問題

## 優先順位と修正戦略

### Phase 1: 高優先度（機能に直接影響）

#### 1.1 フラグ名の不一致修正
- **影響**: E2Eテストの多くが失敗
- **問題**: テストが `-output` を期待しているが、実際のフラグは `--format`
- **修正ファイル**:
  - `test/e2e/advanced_display_test.go`
  - その他E2Eテストファイル
- **修正内容**: `-output` を `--format` に置換

#### 1.2 設定優先順位の修正
- **影響**: 設定システムの中核機能
- **問題**: CLI引数と環境変数のマージロジック
- **修正ファイル**:
  - `test/integration/env_priority_test.go`
  - `internal/config/manager.go`（必要であれば）
- **調査が必要**:
  - ResolveConfig メソッドの動作
  - 優先順位の実装ロジック

### Phase 2: 中優先度（検証関連）

#### 2.1 設定検証の修正
- **問題**: 検証メソッドがエラーを返すべきでない場合にエラーを返している
- **修正内容**:
  - URL検証ロジックの見直し
  - タイムアウトと出力形式の検証ロジック確認
- **影響ファイル**:
  - `test/integration/env_priority_test.go`
  - `internal/config/validator.go`

#### 2.2 レガー設定との互換性
- **問題**: YAMLパースエラー
- **修正内容**:
  - レガシー設定ファイルのフォーマット確認
  - パーサーのエラーハンドリング改善
- **影響ファイル**:
  - `test/integration/config_integration_test.go`
  - `internal/config/file.go`

### Phase 3: 低優先度（インフラ関連）

#### 3.1 E2Eテストのバイナリパス問題
- **問題**: テストが `../../llm-info` を期待している
- **解決策案**:
  - 案1: テスト前にバイナリをビルドして適切な場所に配置
  - 案2: テストをインストール済みバイナリを使用するように変更
  - 案3: Makefileにテスト用ターゲットを追加

## 実行計画

### ステップ1: 準備
- [ ] テスト環境のセットアップ確認
- [ ] 現在のテスト結果を保存（`make test > test_results_before.txt`）

### ステップ2: Phase 1の修正
- [ ] フラグ名の不一致を修正
- [ ] 設定優先順位問題を調査
- [ ] 設定優先順位の修正を実装
- [ ] Phase 1のテストがすべてパスすることを確認

### ステップ3: Phase 2の修正
- [ ] 設定検証ロジックを修正
- [ ] レガー設定互換性を修正
- [ ] Phase 2のテストがすべてパスすることを確認

### ステップ4: Phase 3の修正
- [ ] E2Eテストのバイナリパス問題を解決
- [ ] すべてのE2Eテストが実行できることを確認

### ステップ5: 最終検証
- [ ] すべてのテストスイットを実行
- [ ] テスト結果をドキュメント化
- [ ] リグレッションテストを実行

## 注意事項

1. **バージョン互換性**: 修正が既存のAPIや動作に影響を与えないか確認
2. **ドキュメント更新**: 仕様変更があった場合はドキュメントを更新
3. **コミット粒度**: 各フェーズでコミットを作成し、変更を追跡可能にする
4. **テストカバレッジ**: 修正によるカバレッジの変化を確認

## 成功の定義

- すべての統合テストがパス
- すべてのE2Eテストが実行可能でパス
- テストカバレッジが現状を維持または向上
- 既存機能が破壊されていないこと

## 期間見積

- Phase 1: 30-45分
- Phase 2: 30-45分
- Phase 3: 15-30分
- 合計: 1.5-2時間

---
*作成日: 2026-01-10*
*更新日: 2026-01-10*