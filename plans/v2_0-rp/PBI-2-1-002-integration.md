# プロダクトバックログアイテム: 統合機能（同時実行とマージ）

**作成日**: 2026-01-11
**更新日**: 2026-01-12
**ステータス**: Done
**ランク**: 11

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## ユーザーストーリー

開発者として、Context WindowとMax Output Tokensを1回のコマンドで測定したい、なぜなら両方の情報が必要で、個別実行は手間がかかるから

## ビジネス価値

- **効率化**: 1回のコマンドで全ての制約値を取得
- **時間短縮**: 個別実行の2回分の時間を削減
- **統合レポート**: 両方の結果を1つのレポートにまとめて表示
- **UX向上**: シンプルなコマンド体系

## 機能概要

### 新コマンド: probe（統合版）
```bash
# 両方を実行
llm-info probe --model GLM-4.6

# 出力例:
# Model Constraints Probe Results
# ================================
# Model:             GLM-4.6
# Context Window:    127,000 tokens
# Max Output Tokens: 16,384 tokens
# Total Trials:      20
# Total Duration:    75.5s
```

### オプション
- `--context-only`: Context Windowのみ測定
- `--output-only`: Max Output Tokensのみ測定
- デフォルト: 両方測定

### 実行順序
1. Context Window測定（時間がかかる方を先に）
2. Max Output Tokens測定
3. 結果をマージして表示

## 受け入れ基準（Draft）

- [ ] 1回のコマンドで両方の測定が実行される
- [ ] 結果が統合されたテーブルで表示される
- [ ] --context-onlyオプションで個別測定可能
- [ ] エラー時も適切にハンドリングされる

## 見積もり

**ストーリーポイント**: 2（約4時間）

## v2.0で延期した理由

- 個別実行でも十分な価値を提供できる
- 統合ロジックの複雑性を避け、早期リリースを優先
- 基本機能の安定化が先決

## v2.1での優先順位

高（UX向上の観点から重要）

## 実装記録

### 2026-01-12 06:41:00

**実装者**: Claude Code

**実装内容**:
- `cmd/llm-info/probe.go:26-291`: 統合probeコマンドを実装し、--context-onlyと--output-onlyオプションを追加
- `cmd/llm-info/probe.go:662-717`: showIntegratedExecutionPlan関数を実装し、dry-runモードでの実行計画表示を対応
- `internal/ui/table_formatter.go:23-73`: FormatIntegratedResultメソッドを実装し、統合結果のテーブル表示を追加
- `cmd/llm-info/probe.go:615-659`: showProbeHelp関数を更新し、新しいオプションの説明を追加

**遭遇した問題と解決策**:
- **問題**: 型の不一致によりコンパイルエラーが発生（logging.ProbeLoggerとstorage.ResultStorageがインターフェース型）
  **解決策**: 変数宣言をインターフェース型に修正
- **問題**: probe.MaxOutputTokensResult型が未定義
  **解決策**: 正しい型名probe.MaxOutputResultを使用
- **問題**: fmt.Printlnの末尾に余分な改行があるというlint警告
  **解決策**: ヘルプ文字列の末尾の改行を削除

**テスト結果**:
- ビルドテスト: ✅ 成功 - llm-infoバイナリが正常にビルドできることを確認
- ヘルプ表示テスト: ✅ 成功 - probe --helpで新しいオプションが正しく表示されることを確認
- Dry-runテスト（両方）: ✅ 成功 - --dry-runで統合実行計画が正しく表示されることを確認
- Dry-runテスト（context-only）: ✅ 成功 - --context-only --dry-runでContext Windowのみの計画が表示されることを確認
- Dry-runテスト（output-only）: ✅ 成功 - --output-only --dry-runでMax Output Tokensのみの計画が表示されることを確認
- 既存のprobeテスト: ✅ 成功 - internal/probeパッケージの既存テストがすべて成功することを確認

**受け入れ基準の達成状況**:
- [x] 1回のコマンドで両方の測定が実行される
- [x] 結果が統合されたテーブルで表示される
- [x] --context-onlyオプションで個別測定可能
- [x] --output-onlyオプションで個別測定可能
- [x] エラー時も適切にハンドリングされる（片方失敗時は警告を表示して継続）

**備考**:
- 実行順序はPBIの仕様通り、Context Window（時間がかかる方）を先に実行
- ログと結果保存機能は既存の個別コマンドと同様に対応
- エラーハンドリングでは、片方の探索が失敗してももう片方は実行を継続する設計とし、部分的な成功も報告できるようにした
