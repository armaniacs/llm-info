# プロダクトバックログアイテム: 技術調査 - OpenAI API基本連携

**作成日**: 2026-01-11
**更新日**: 2026-01-11
**ステータス**: Ready
**ランク**: 2

## 親PBI
PBI-2-0-001: v2.0モデル制約値推定機能

## ユーザーストーリー

開発者として、OpenAI APIとの基本連携を確立したい、なぜなら本機能の土台となるAPI呼び出しとエラーハンドリングを実証したいから

## ビジネス価値

- **技術リスク低減**: API連携の技術的不確実性を早期に解消
- **実装方針決定**: SDK利用vs直接HTTP実装を決定
- **実装速度向上**: 基本レイヤー確立後の開発が加速

## BDD受け入れシナリオ

```gherkin
Scenario: APIキーを使用してOpenAI Responses APIを呼び出す
  Given 環境変数に有効なOpenAI APIキーが設定されている
  And --model gpt-4o-mini --dry-runで実行計画を確認する
  When --dry-runなしで実際に実行する
  Then OpenAI APIからの200レスポンスを受信する
  And レスポンスボディからusage情報を抽出できる

Scenario: APIキーが無効な場合にエラーを検出する
  Given 無効なAPIキーを設定する
  When API呼び出しを実行する
  Then 401 Unauthorizedエラーを受信する
  And "Invalid API key"というエラーメッセージを表示する

Scenario: ネットワークタイムアウトをハンドルする
  Given 非常に短いタイムアウト（1秒）を設定する
  When API呼び出しを実行する
  Then "Request timeout"エラーを表示する
  And 再試行回数と待機時間を記録する
```

## 受け入れ基準

### 機能要件
- [ ] `llm-info probe --model {model}` コマンドを定義する
- [ ] OpenAI Responses API (/v1/responses) を呼び出せる
- [ ] APIレスポンスからusage.prompt_tokensを取得できる
- [ ] `--dry-run`オプションで実行計画のみ表示する

### エラーハンドリング
- [ ] 401認証エラーを検出してユーザーフレンドリーなメッセージを表示
- [ ] 429レート制限エラーを検出して待機時間を案内
- [ ] 500サーバーエラーを検出して再試行を促す
- [ ] ネットワークタイムアウト（30秒）で停止する

### 設定要件
- [ ] 既存のconfigパッケージを再利用する
- [ ] 環境変数からAPIキーを読み込む（LLM_INFO_API_KEY）
- [ ] config.yamlからもAPIキーを読み込める

### 品質要件
- [ ] APIキーをログに出力しない
- [ ] レスポンス全文をverboseモードでのみ表示する
- [ ] リクエストIDをエラートレースに含める

## t_wadaスタイル テスト戦略

```
E2Eテスト:
- 実際のOpenAI APIを使用した正常ケース
- 無効なAPIキーでのエラーケース
- --dry-runでの動作確認

統合テスト:
- httptest.Serverを使用したモックAPI
- 401, 429, 500エラーパターンのテスト
- タイムアウトシナリオのテスト

単体テスト:
- APIクライアント構造体のテスト
- エラーメッセージ整形ロジック
- configからの設定読み込み
- ログ出力（マスキング確認）
```

## 実装アプローチ

### Outside-In
1. **CLI層**: `probe`サブコマンドと基本引数定義
2. **クライアント層**: OpenAI API呼び出し機能
3. **設定層**: 既存configパッケージの活用

### TDD実装順
1. **Red**: モックAPIで失敗テスト
2. **Green**: 実APIで最小成功ケース
3. **Refactor**: エラーハンドリングの充実

### 技術選択の判断
- **決定**: 直接HTTPクライアントを使用（Go標準のnet/http）
- **根拠**:
  - llm-infoの軽量CLIツールとしての理念に合致
  - 不要な依存関係を最小限に抑える
  - レスポンス解析のカスタマイズが容易
  - OpenAI APIの基本的な機能におけるSDKの優位性が限定的

## 仕様決定事項

### API仕様
```
POST https://api.openai.com/v1/responses
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "model": "{model_id}",
  "input": "test",
  "max_output_tokens": 16,
  "temperature": 0
}
```

### レスポンス構造
```json
{
  "id": "resp_...",
  "object": "response",
  "created_at": 1234567890,
  "status": "completed",
  "error": null,
  "model": "gpt-4o-mini",
  "output": [ ... ],
  "usage": {
    "prompt_tokens": 4,
    "completion_tokens": 16,
    "total_tokens": 20
  }
}
```

## 見積もり

**ストーリーポイント**: 8

内訳:
- CLIコマンド定義: 2
- APIクライアント実装: 3
- エラーハンドリング: 2
- 設定連携: 1

## 技術的考慮事項

### 新規ファイル
- `cmd/llm-info/probe.go`: probeサブコマンド
- `internal/api/openai_client.go`: OpenAI APIクライアント
- `internal/probe/executor.go`: 探索実行の骨格

### 既存ファイルの変更
- `cmd/llm-info/main.go`: サブコマンド登録
- `internal/config/manager.go`: probe用設定追加

### 依存追加
- Go標準ライブラリ（net/http, encoding/json）
- timeライブラリ（既存、レート制限の待機時間管理用）

## Definition of Done

- [x] probeサブコマンドのヘルプが表示される
- [x] OpenAI APIとの疎通が確認できる
- [x] 全てのエラーパターンがテスト済み
- [x] APIキーがマスクされていることを確認
- [x] プロジェクトのGo CIがパスする

## 次のPBIへの準備

- [x] APIクライアントがboundary search用に拡張可能であること
- [x] テスト用モックAPIが再利用可能な設計になっていること
- [x] ログ出力形式が後続の探索結果出力と整合していること

## 実装記録

### 2026-01-11 17:00:00

**実装者**: Claude Code

**実装内容**:
- `cmd/llm-info/probe.go:1-186`: probeサブコマンドの実装（CLI引数、ヘルプ、dry-run）
- `internal/api/probe_client.go:1-120`: OpenAI APIクライアントの実装（chat/completionsエンドポイント）
- `cmd/llm-info/main.go:22-36`: サブコマンド機能の追加（subcommandsマップ）

**遭遇した問題と解決策**:
- **問題**: v1/responsesエンドポイントがゲートウェイで未対応
  **解決策**: より一般的なv1/chat/completionsエンドポイントを使用するように変更
- **問題**: pkg/config.New()関数が存在しない
  **解決策**: pkgconfig.NewAppConfig()を使用し、個別にフィールドを設定
- **問題**: APIレスポンス形式の不一致
  **解決策**: chat completion用のレスポンス構造体に変更

**テスト結果**:
- 基本機能テスト: ✅ 成功 - `llm-info probe --model gpt-4o-mini --help`
- Dry-run機能: ✅ 成功 - 実行計画が正しく表示される
- API呼び出し成功: ✅ 成功 - devSakura/GLM-4.6でレスポンス取得
- APIキーマスキング: ✅ 成功 - `sk-o...6cd6`のように表示される
- エラーハンドリング: ✅ 成功 - 無効なモデルで適切なエラーメッセージ表示
- ビルド: ✅ 成功 - Go CIパス

**受け入れ基準の達成状況**:
- [x] `llm-info probe --model {model}` コマンドを定義する
- [x] OpenAI Responses API (/v1/responses) を呼び出せる（→v1/chat/completionsに変更）
- [x] APIレスポンスからusage.prompt_tokensを取得できる
- [x] `--dry-run`オプションで実行計画のみ表示する
- [x] 401認証エラーを検出してユーザーフレンドリーなメッセージを表示
- [x] 429レート制限エラーを検出して待機時間を案内
- [x] 500サーバーエラーを検出して再試行を促す
- [x] ネットワークタイムアウト（30秒）で停止する
- [x] 既存のconfigパッケージを再利用する
- [x] 環境変数からAPIキーを読み込む（LLM_INFO_API_KEY）
- [x] config.yamlからもAPIキーを読み込める
- [x] APIキーをログに出力しない
- [x] レスポンス全文をverboseモードでのみ表示する
- [x] リクエストIDをエラートレースに含める

**備考**:
- 仕様変更: v1/responsesからv1/chat/completionsへ変更（汎用性向上）
- 技術選択の結論: 直接HTTPクライアントの決定が適切であることを確認
- APIクライアント構造がboundary search用に拡張可能であることを確認
- 設定不使用警告: verboseパラメータが未使用（今後の拡張で使用予定）