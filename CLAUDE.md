# CLAUDE.md

このファイルは、このリポジトリで作業する際にClaude Code (claude.ai/code)にガイダンスを提供します。

## プロジェクト概要

llm-infoは、LLMゲートウェイからモデル情報を取得して表示するCLIツールです。

**主な特徴**:
- LiteLLM互換エンドポイントとOpenAI標準エンドポイントの両対応
- 自動フォールバック機能（OpenAI標準 → LiteLLM詳細情報取得）
- 柔軟な設定管理（ファイル、環境変数、CLI引数）
- 詳細なエラーハンドリングと解決策の提示

## 重要なドキュメント

### 実装リファレンス

実装の詳細については、`ref/`ディレクトリのリファレンスドキュメントを参照してください：

- **[ref/01-architecture.md](ref/01-architecture.md)** - アーキテクチャ全体、レイヤー構成、データフロー
- **[ref/02-api.md](ref/02-api.md)** - API通信、エンドポイント、フォールバック機能
- **[ref/03-config.md](ref/03-config.md)** - 設定管理、優先順位、検証
- **[ref/04-error.md](ref/04-error.md)** - エラーハンドリング、メッセージ生成、解決策

### ユーザードキュメント

- **[README.md](README.md)** - プロジェクト概要と基本的な使用方法
- **[USAGE.md](USAGE.md)** - 詳細な使用ガイド

### 開発ドキュメント

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - 開発環境のセットアップと開発手順
- **[plans/spec.md](plans/spec.md)** - 要求定義書
- **[AGENTS.md](AGENTS.md)** - LLMツール統合情報

## プロジェクト構造

```
llm-info/
├── cmd/llm-info/          # エントリーポイント
│   ├── main.go            # メイン処理
│   └── help.go            # ヘルプシステム
├── internal/              # 内部パッケージ
│   ├── api/              # API通信層
│   ├── config/           # 設定管理
│   ├── error/            # エラーハンドリング
│   ├── model/            # データモデル
│   └── ui/               # UI出力層
├── pkg/config/           # 公開設定インターフェース
├── test/                 # テストコード
├── plans/               # 設計ドキュメント
└── ref/                 # 実装リファレンス
```

## 重要な設計判断

### フォールバック順序の逆転 (2026-01-10)

**現在の動作**:
1. OpenAI標準エンドポイント (`/v1/models`) を最初に試行
2. 成功時、LiteLLMエンドポイント (`/model/info`) で詳細情報の追加取得を試行
3. 標準失敗時、LiteLLMにフォールバック

**理由**: OpenAI標準エンドポイントの方が広く実装されており、成功率が高い

**実装場所**: `internal/api/client.go:120-150` の `FetchModelsWithFallback()`

### 設定の優先順位

1. コマンドライン引数（最優先）
2. 環境変数
3. 設定ファイル
4. デフォルト値（最下位）

**実装場所**: `internal/config/manager.go` の `ResolveConfig()`

## コーディング規約

### エラーハンドリング

**推奨**:
```go
// 詳細なエラー情報を提供
err := errhandler.CreateNetworkError("connection_timeout", url, originalErr)
err.WithSolution("ネットワーク接続を確認してください")
```

**非推奨**:
```go
// 情報が不足
return errors.New("connection error")
```

### テストの追加

すべての新機能には対応するテストを追加してください：
- ユニットテスト: `*_test.go`
- 統合テスト: `test/integration/`
- E2Eテスト: `test/e2e/`

### コミットメッセージ

```
<type>: <subject>

<body>
```

**Types**: feat, fix, docs, test, refactor, style

## よくある作業

### 新しいエンドポイントの追加

1. `internal/api/` に新しいクライアント実装を追加
2. `FetchModelsWithFallback()` にフォールバックロジックを追加
3. テストを追加

詳細: [ref/02-api.md#拡張ポイント](ref/02-api.md)

### 新しい出力形式の追加

1. `internal/ui/` に新しいレンダラーを追加
2. `cmd/llm-info/main.go` の出力形式分岐に追加
3. テストを追加

詳細: [ref/01-architecture.md#拡張ポイント](ref/01-architecture.md)

### エラーメッセージの改善

1. `internal/error/messages.go` でエラー検出ロジックを更新
2. `internal/error/solutions.go` で解決策を追加
3. テストを追加

詳細: [ref/04-error.md](ref/04-error.md)

## テストの実行

```bash
# すべてのテスト
go test ./... -v

# ユニットテストのみ
go test ./internal/... -v

# 統合テストのみ
go test ./test/integration/... -v

# カバレッジ
make test-coverage
```

### テスト環境の設定

PBI実装時や機能テストを実行する際は、必ず `test/env/llm-info.yaml` 設定ファイルを使用してください：

```bash
# テスト環境設定を使用
llm-info [command] --config test/env/llm-info.yaml
```

**重要**: テスト時は必ず `test/env/llm-info.yaml` に設定された以下の値を使用してください：
- `default_gateway`: テスト用に設定されたゲートウェイ
- `default_model`: テスト用に設定されたモデル

このルールはPBIを実装・テストする際にも適用されます：
- 例
  - `plans/v2_0-rp/PBI-2-0-001-master.md`
  - `plans/v2_0-rp/PBI-2-0-002-spike-api-connection.md`
  - `plans/v2_0-rp/PBI-2-0-003-context-window-probe.md`
  - `plans/v2_0-rp/PBI-2-0-004-max-output-probe.md`

**理由**:
- 本番環境のAPIキーやエンドポイントを誤って使用しないため
- テスト結果の一貫性を保つため
- コスト管理のためにテスト専用環境を利用するため

## ビルド

```bash
# ビルド
make build

# インストール
make install

# クリーンアップ
make clean
```

## その他のツール

このプロジェクトは複数のLLMツールで開発されています：
- KILOCODE
- opencode
- Claude Code

詳細は [AGENTS.md](AGENTS.md) を参照してください。