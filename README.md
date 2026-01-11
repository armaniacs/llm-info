# llm-info

LLMゲートウェイからモデル情報を取得して表示するCLIツールです。
OpenAI標準の`/v1/models`エンドポイントをサポートしています。

LiteLLM互換の`/model/info`エンドポイントについては、実際のテストを行っていません。

## 機能

- LiteLLM互換のゲートウェイからモデル情報を取得
- OpenAI標準互換のゲートウェイからモデル情報を取得
- 自動フォールバック機能（LiteLLMエンドポイント失敗時にOpenAI標準エンドポイントを試行）
- モデル情報を整形されたテーブル形式で表示
- JSON形式での出力に対応
- 動的な列制御（利用可能なデータに応じて表示列を調整）
- 設定ファイルによるゲートウェイ管理（YAML形式）
- 環境変数による設定
- APIキー認証に対応
- タイムアウト設定可能
- 詳細なエラーメッセージと解決策の提示
- **v2.0新機能**: モデル制約値の探索機能
  - Context Window探索（最大入力トークン数）
  - Max Output Tokens探索（最大出力トークン数）
  - 見やすいテーブル形式での結果表示

## インストール

### ソースからビルド

```bash
# リポジトリをクローン
git clone https://github.com/armaniacs/llm-info.git
cd llm-info

# ビルド
make build

# インストール（オプション）
make install
```

### Goインストール

```bash
go install github.com/armaniacs/llm-info/cmd/llm-info@latest
```

**インストール先について**:
- `GOBIN` 環境変数が設定されている場合: `$GOBIN/llm-info`
- `GOBIN` が未設定の場合: `$HOME/go/bin/llm-info`

`go env GOBIN` でインストール先を確認できます。PATHを適切に設定してください。

## 使用方法

詳細な使用方法については、[USAGE.md](USAGE.md)を参照してください。

### 基本的な使い方

```bash
llm-info --url https://gateway.example.com/v1
```

### APIキーを使用する場合

```bash
llm-info --url https://gateway.example.com/v1 --api-key your-api-key
```

### 設定ファイルを使用する場合

```bash
llm-info --config ~/.config/llm-info/llm-info.yaml --gateway development
```

### JSON形式で出力する場合

```bash
llm-info --url https://gateway.example.com/v1 --output json
```

### 環境変数を使用する場合

```bash
export LLM_INFO_URL=https://gateway.example.com/v1
export LLM_INFO_API_KEY=your-api-key
llm-info
```

### タイムアウトを指定する場合

```bash
llm-info --url https://gateway.example.com/v1 --timeout 30s
```

### v2.0: モデル制約値の探索

```bash
# Context Windowの探索
llm-info probe-context --model gpt-4o

# Max Output Tokensの探索
llm-info probe-max-output --model gpt-4o

# 詳細な探索履歴を表示
llm-info probe-context --model gpt-4o --verbose

# カスタムゲートウェイで探索
llm-info probe-max-output --model claude-3-opus --gateway production
```

### ヘルプを表示

```bash
llm-info --help
```

## コマンドラインオプション

| オプション | 説明 | 必須 | デフォルト値 |
|-----------|------|------|-------------|
| `--url` | LLMゲートウェイのベースURL | いいえ¹ | - |
| `--api-key` | 認証用APIキー | いいえ | - |
| `--timeout` | リクエストタイムアウト | いいえ | 10s |
| `--config` | 設定ファイルのパス | いいえ | ~/.config/llm-info/llm-info.yaml |
| `--gateway` | 使用するゲートウェイ名 | いいえ | default |
| `--output` | 出力形式 (table, json) | いいえ | table |
| `--help` | ヘルプメッセージを表示 | いいえ | - |
| `--version` | バージョン情報を表示 | いいえ | - |

¹ `--url` は設定ファイルまたは環境変数で指定されていない場合に必須です。

## 出力例

```
Fetching model information from https://gateway.example.com/v1...
Found 3 models:

Model Name      Max Tokens    Mode      Input Cost
gpt-4           8192          chat      0.000030
claude-3-opus   200000        chat      0.000015
gemini-1.5-pro  1000000       chat      0.000000
```

## 設定ファイル

設定ファイルを使用すると、複数のゲートウェイを事前に登録しておくことができます。設定ファイルはYAML形式で、デフォルトでは `~/.config/llm-info/llm-info.yaml` に配置します。

### 設定ファイルの例

```yaml
gateways:
  - name: "default"
    url: "https://api.example.com/v1"
    api_key: "your-api-key-here"
    timeout: "10s"
  
  - name: "development"
    url: "https://dev-api.example.com/v1"
    api_key: "dev-api-key"
    timeout: "5s"

default_gateway: "default"

common:
  timeout: "10s"
  output:
    format: "table"
    table:
      always_show: ["name"]
      show_if_available: ["max_tokens", "mode", "input_cost"]
```

## 環境変数

以下の環境変数を使用して設定を指定できます：

| 環境変数 | 説明 |
|----------|------|
| `LLM_INFO_URL` | LLMゲートウェイのベースURL |
| `LLM_INFO_API_KEY` | 認証に使用するAPIキー |
| `LLM_INFO_TIMEOUT` | リクエストタイムアウト (例: 10s, 1m) |
| `LLM_INFO_DEFAULT_GATEWAY` | デフォルトゲートウェイ名 |
| `LLM_INFO_OUTPUT_FORMAT` | 出力形式 (table, json) |
| `LLM_INFO_SORT_BY` | ソート項目 (name, max_tokens, mode) |
| `LLM_INFO_FILTER` | フィルタ条件 |
| `LLM_INFO_CONFIG_PATH` | 設定ファイルのパス |
| `LLM_INFO_LOG_LEVEL` | ログレベル |
| `LLM_INFO_USER_AGENT` | ユーザーエージェント |

### 設定の優先順位

設定は以下の優先順位で適用されます（上にあるほど優先度が高い）：

1. コマンドライン引数
2. 環境変数
3. 設定ファイル
4. デフォルト値

### 環境変数の使用例

```bash
# 基本設定
export LLM_INFO_URL="https://api.example.com/v1"
export LLM_INFO_API_KEY="your-api-key"
export LLM_INFO_TIMEOUT="15s"

# 出力設定
export LLM_INFO_OUTPUT_FORMAT="json"
export LLM_INFO_SORT_BY="max_tokens"
export LLM_INFO_FILTER="gpt"

# 実行
llm-info
```

### CI/CDパイプラインでの使用例

```yaml
# GitHub Actions
- name: List available models
  env:
    LLM_INFO_URL: ${{ secrets.LLM_GATEWAY_URL }}
    LLM_INFO_API_KEY: ${{ secrets.LLM_API_KEY }}
    LLM_INFO_OUTPUT_FORMAT: json
  run: |
    llm-info > models.json
```

### Dockerコンテナでの使用例

```dockerfile
FROM alpine:latest
RUN apk add --no-cache llm-info
ENV LLM_INFO_URL="https://api.example.com"
ENV LLM_INFO_API_KEY=""
CMD ["llm-info"]
```

## 対応しているゲートウェイ

このツールは以下のエンドポイントを実装しているゲートウェイで動作します：

- LiteLLM互換の`/model/info`エンドポイント
- OpenAI標準互換の`/v1/models`エンドポイント

### 自動フォールバック機能

ツールはまずLiteLLMの`/model/info`エンドポイントを試行し、失敗した場合は自動的にOpenAI標準の`/v1/models`エンドポイントにフォールバックします。これにより、より多くのゲートウェイでシームレスに動作します。

### テスト済みゲートウェイ

- https://openrouter.ai/api (OpenAPI互換)


## 開発

### 前提条件

- Go 1.21+
- Make

### 開発環境のセットアップ

```bash
# 依存関係をダウンロード
go mod download

# テスト実行
make test

# カバレッジ付きテスト
make test-coverage

# リンティング
make lint
```

### プロジェクト構造

```
llm-info/
├── cmd/
│   └── llm-info/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── api/
│   │   ├── client.go            # APIクライアント
│   │   ├── response.go          # APIレスポンスモデル
│   │   ├── endpoints.go         # エンドポイント管理
│   │   └── standard_client.go   # OpenAI標準クライアント
│   ├── config/
│   │   ├── manager.go           # 設定マネージャー
│   │   ├── env.go               # 環境変数設定
│   │   └── config.go            # 設定構造体
│   ├── model/
│   │   └── model.go             # データモデル
│   ├── ui/
│   │   ├── table.go             # テーブル表示
│   │   └── json.go              # JSON出力
│   └── error/
│       └── handler.go           # エラーハンドリング
├── pkg/
│   └── config/
│       └── config.go            # 設定構造体
├── configs/
│   └── example.yaml             # 設定ファイル例
├── test/
│   └── integration/             # 統合テスト
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### ビルドターゲット

```bash
make build          # バイナリをビルド
make test           # テストを実行
make test-coverage  # カバレッジ付きテスト
make clean          # ビルド成果物をクリーンアップ
make lint           # リンティングを実行
make install        # バイナリをインストール
make run-example    # サンプルゲートウェイで実行
make help           # 利用可能なターゲットを表示
```

## テスト

### 単体テスト

```bash
make test
```

### 統合テスト

```bash
go test -v ./test/integration/...
```

### カバレッジ

```bash
make test-coverage
```

カバレッジレポートは `coverage.html` に生成されます。

## トラブルシューティング

より詳細なトラブルシューティング情報については、[USAGE.md](USAGE.md#トラブルシューティング)を参照してください。

### 接続エラー

- `--url` が正しいことを確認してください
- ゲートウェイが稼働していることを確認してください
- ネットワーク接続を確認してください

### 認証エラー

- `--api-key` が正しいことを確認してください
- APIキーに必要な権限があることを確認してください

### タイムアウトエラー

- `--timeout` を増やしてみてください
- ゲートウェイの応答時間を確認してください

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 貢献

バグ報告や機能要求はIssueを通じてお願いします。プルリクエストも歓迎します。

## リリース

### v2.0.0 (Latest)

- [x] **モデル制約値探索機能**
  - [x] Context Window探索（probe-contextコマンド）
  - [x] Max Output Tokens探索（probe-max-outputコマンド）
  - [x] テーブル形式での結果出力
  - [x] 詳細な探索履歴表示（verboseモード）
  - [x] 数値の3桁区切り表示
  - [x] ステータスインジケーター（✓/✗）

### v1.1.0 (Phase 2)

- [x] 設定ファイル対応（YAML形式）
- [x] 環境変数による設定
- [x] JSON出力形式
- [x] 詳細なエラーハンドリング
- [x] 複数ゲートウェイ対応
- [x] 高いテストカバレッジ（80%以上）

### v1.0.0 (MVP)

- [x] LiteLLM互換の`/model/info`エンドポイントからモデル情報を取得
- [x] モデル情報をテーブル形式で表示
- [x] コマンドライン引数でベースURLとAPIキーを指定
- [x] タイムアウト処理
- [x] シンプルで分かりやすいエラーメッセージ
- [x] Go言語のシングルバイナリとしてビルド
- [x] 基本的なテストカバレッジ

## 将来の機能

- 追加出力形式（CSV）
- モデル比較機能
- コスト計算ツール
- キャッシュ機能
- JSON出力形式での探索結果
- 探索結果のキャッシュと比較