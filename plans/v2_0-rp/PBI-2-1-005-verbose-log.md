# プロダクトバックログアイテム: 詳細なverboseログ出力

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Backlog（v2.1以降）
**ランク**: 14

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## ユーザーストーリー

開発者として、探索の各試行をリアルタイムで確認したい、なぜなら長時間の探索中に進捗を把握し、問題をデバッグしたいから

## ビジネス価値

- **可視性**: 探索の進捗状況をリアルタイムで把握
- **デバッグ**: 問題発生時に詳細ログで原因を特定
- **安心感**: 長時間の探索でも進行中であることを確認
- **学習**: 探索アルゴリズムの動作を理解

## 機能概要

### 現在の --verbose の拡張

```bash
llm-info probe-context --model GLM-4.6 --verbose
```

### 出力内容

#### リアルタイム進捗表示
```
[INFO] Starting exponential search...
[PROGRESS] Trial 1/40: Testing with 4,096 tokens...
[SUCCESS] Trial 1: Success (prompt_tokens=4,050)
[PROGRESS] Trial 2/40: Testing with 8,192 tokens...
[SUCCESS] Trial 2: Success (prompt_tokens=8,100)
...
[INFO] Exponential search completed. Upper bound: 128,000 tokens
[INFO] Starting binary search...
[PROGRESS] Trial 15/40: Testing with 96,000 tokens...
[SUCCESS] Trial 15: Success (prompt_tokens=95,800)
[INFO] Converged! Final estimate: 127,000 tokens (±128)
```

#### API通信の詳細
- リクエストURL
- リクエストボディ（センシティブ情報はマスク）
- レスポンスステータス
- レスポンスボディの要約

#### 探索戦略の説明
- 次の試行値の計算根拠
- 収束判定のロジック
- エラー時の retry 戦略

## 受け入れ基準（Draft）

- [ ] 各試行がリアルタイムで表示される
- [ ] 進捗率（X/40）が表示される
- [ ] APIリクエスト/レスポンスの詳細が表示される
- [ ] センシティブ情報（APIキー）がマスクされる
- [ ] エラー時のスタックトレースが表示される

## 見積もり

**ストーリーポイント**: 2（約4時間）

## v2.0で延期した理由

- 基本的なverbose出力（探索履歴テーブル）で十分
- リアルタイム表示の実装が複雑
- 早期リリースを優先

## v2.1での優先順位

低〜中（デバッグ用途では有用だが必須ではない）
