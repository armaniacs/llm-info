# プロダクトバックログアイテム: JSON形式での出力

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Backlog（v2.1以降）
**ランク**: 12

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## ユーザーストーリー

開発者として、探索結果をJSON形式で出力したい、なぜならCI/CDパイプラインやスクリプトで機械的に処理したいから

## ビジネス価値

- **自動化**: スクリプトやCI/CDでの自動処理が可能
- **データ連携**: 他ツールとのデータ連携が容易
- **機械可読性**: jqなどのツールで柔軟に処理可能
- **標準化**: 業界標準のフォーマットで結果を提供

## 機能概要

### --format フラグ
```bash
# JSON形式で出力
llm-info probe-context --model GLM-4.6 --format json

# 出力例:
{
  "model": "GLM-4.6",
  "estimated_max_context_tokens": 127000,
  "method_confidence": "high",
  "trials": 12,
  "duration_seconds": 45.3,
  "max_input_tokens_at_success": 126800,
  "timestamp": "2026-01-11T19:30:00Z"
}
```

### --out フラグとの組み合わせ
```bash
# JSONファイルに保存
llm-info probe-context --model GLM-4.6 \
    --format json \
    --out result.json
```

## 受け入れ基準（Draft）

- [ ] --format jsonで有効なJSON形式で出力される
- [ ] 全ての測定結果がJSONに含まれる
- [ ] --outオプションでファイルに保存できる
- [ ] jqなどのツールで処理可能な形式

## 見積もり

**ストーリーポイント**: 2（約4時間）

## v2.0で延期した理由

- テーブル形式で人間が確認できれば十分
- JSON出力なしでも基本的な価値を提供できる
- 早期リリースを優先

## v2.1での優先順位

中〜高（自動化ニーズが高い環境では重要）
