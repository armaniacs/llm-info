# プロダクトバックログアイテム: API使用量計算とコスト警告

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Backlog（v2.1以降）
**ランク**: 15

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## ユーザーストーリー

開発者として、探索にかかるAPI使用料金を事前に知りたい、なぜなら予算超過を防ぎ、コストを管理したいから

## ビジネス価値

- **コスト管理**: 探索前に概算コストを把握
- **予算管理**: $0.05超過時に警告
- **透明性**: 実際のAPI使用量とコストを可視化
- **意思決定**: コストを考慮した探索範囲の決定

## 機能概要

### 事前コスト見積もり

```bash
llm-info probe-context --model GLM-4.6 --dry-run

# 出力例:
Estimated API Usage:
  Expected trials:  20-30
  Expected tokens:  500,000-800,000
  Estimated cost:   $0.025-$0.040 USD

Continue? [y/N]:
```

### 実行中のコスト表示

```bash
llm-info probe-context --model GLM-4.6 --show-cost

# 出力に追加:
API Usage Summary:
  Total trials:      25
  Total tokens:      650,000 (prompt: 640,000, completion: 10,000)
  Estimated cost:    $0.032 USD

⚠️  Warning: Cost exceeds $0.05 threshold
```

### 料金計算ロジック
- モデルごとの料金レートを保持
- 使用トークン数から自動計算
- LiteLLMゲートウェイの料金体系に対応

## 受け入れ基準（Draft）

- [ ] --dry-runで概算コストが表示される
- [ ] 実行後に実際のコストが表示される
- [ ] $0.05超過時に警告が表示される
- [ ] 料金レートが設定可能（config.yaml）
- [ ] 複数のモデル料金に対応

## 見積もり

**ストーリーポイント**: 3（約6時間）

## v2.0で延期した理由

- 初期ユーザーは手動で管理可能
- 料金レートの管理が複雑
- コスト計算なしでも基本機能は動作

## v2.1での優先順位

中（企業利用やコスト意識の高い環境では重要）

## 技術的考慮事項

### 料金レートの管理
- 設定ファイルでモデルごとに定義
- LiteLLMのモデルコスト情報を参照
- 定期的な更新が必要

### 計算式
```
Cost = (prompt_tokens × prompt_price_per_1k / 1000) +
       (completion_tokens × completion_price_per_1k / 1000)
```

### レート例（2026-01時点）
- gpt-4o: $0.0025/1K prompt, $0.01/1K completion
- gpt-4o-mini: $0.00015/1K prompt, $0.0006/1K completion
