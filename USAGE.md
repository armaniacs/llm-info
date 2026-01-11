# llm-info 使用ガイド

llm-infoはLiteLLM互換およびOpenAI標準互換のゲートウェイから利用可能なモデル情報を取得して表示するCLIツールです。シンプルで使いやすいインターフェースで、LLMゲートウェイのモデル情報を素早く確認できます。

プロジェクトの概要や開発情報については、[README.md](README.md)を参照してください。

## クイックスタート

最速で使い始めるための手順：

```bash
# ビルド
make build

# 実行
./bin/llm-info --url https://openrouter.ai/api
```

## インストール

### 前提条件

- Go 1.21以上
- Git
- Make（オプション、手動ビルドも可能）

### ソースからビルド

```bash
# リポジトリをクローン
git clone https://github.com/armaniacs/llm-info.git
cd llm-info

# ビルド
make build
```

### Goインストール

```bash
go install github.com/armaniacs/llm-info/cmd/llm-info@latest
```

### バイナリのインストール

```bash
make install
```

## 基本的な使い方

### 最もシンプルな使用例

```bash
llm-info --url https://openrouter.ai/api
```

出力例：
```
❯ ./bin/llm-info |head
⚠️  LiteLLM endpoint failed, falling back to OpenAI standard endpoint: failed to decode JSON response: invalid character '<' looking for beginning of value. Response preview: <!DOCTYPE html><html lang="en"><head><meta charSet="utf-8"/><meta name="viewport" content="width=device-width, initial-scale=1, minimum-scale=1"/><link rel="stylesheet" href="/_next/static/css/cda3af1...
MODEL NAME                                                     MAX TOKENS  MODE  INPUT COST
-------------------------------------------------------------  ----------  ----  ----------
ai21/jamba-large-1.7                                           0           chat  0.000000
ai21/jamba-mini-1.7                                            0           chat  0.000000
aion-labs/aion-1.0                                             0           chat  0.000000
aion-labs/aion-1.0-mini                                        0           chat  0.000000
aion-labs/aion-rp-llama-3.1-8b                                 0           chat  0.000000
alfredpros/codellama-7b-instruct-solidity                      0           chat  0.000000
alibaba/tongyi-deepresearch-30b-a3b                            0           chat  0.000000
```

### APIキーを使用する場合

```bash
llm-info --url https://gateway.example.com/v1 --api-key your-api-key
```

APIキーはゲートウェイプロバイダから取得してください。セキュリティのため、APIキーは環境変数で管理することをお勧めします。

### 環境変数を使用する場合

```bash
export LLM_INFO_URL=https://gateway.example.com/v1
export LLM_INFO_API_KEY=your-api-key
llm-info
```

環境変数を使用すると、コマンドライン引数を省略できます。環境変数はコマンドライン引数より優先度が低くなります。

### 設定ファイルを使用する場合

```bash
# 設定ファイルを作成
mkdir -p ~/.config/llm-info
cp configs/example.yaml ~/.config/llm-info/llm-info.yaml

# 設定ファイルを使用して実行
llm-info --gateway development
```

設定ファイルを使用すると、複数のゲートウェイを事前に登録しておくことができます。

### JSON形式で出力する場合

```bash
llm-info --url https://gateway.example.com/v1 --format json
```

JSON出力例：
```json
[
  {
    "name": "gpt-4",
    "max_tokens": 8192,
    "mode": "chat",
    "input_cost": 0.00003
  },
  {
    "name": "claude-3-opus",
    "max_tokens": 200000,
    "mode": "chat",
    "input_cost": 0.000015
  }
]
```

### タイムアウトのカスタマイズ

```bash
llm-info --url https://gateway.example.com/v1 --timeout 30s
```

タイムアウト値は`10s`、`1m`などの形式で指定できます。デフォルトは10秒です。

### モデルのソート

```bash
# 最大トークン数でソート
llm-info --url https://gateway.example.com/v1 --sort max_tokens

# 降順でソート
llm-info --url https://gateway.example.com/v1 --sort "-max_tokens"

# 複数条件でソート
llm-info --url https://gateway.example.com/v1 --sort "mode,max_tokens"
```

### モデルのフィルタリング

```bash
# 名前でフィルタリング
llm-info --url https://gateway.example.com/v1 --filter "name:gpt"

# トークン数でフィルタリング
llm-info --url https://gateway.example.com/v1 --filter "tokens>100000"

# 複数条件でフィルタリング
llm-info --url https://gateway.example.com/v1 --filter "name:gpt,tokens>1000,mode:chat"
```

### 表示列のカスタマイズ

```bash
# 特定の列のみ表示
llm-info --url https://gateway.example.com/v1 --columns "name,max_tokens"

# 列の順序を指定
llm-info --url https://gateway.example.com/v1 --columns "max_tokens,name,mode"
```

### 設定ファイルテンプレートの作成

```bash
llm-info --init-config
```

これにより、`~/.config/llm-info/llm-info.yaml` に設定ファイルのテンプレートが作成されます。

### 設定ファイルの検証

```bash
llm-info --check-config
```

設定ファイルが有効かどうかを検証します。

### 設定済みゲートウェイの一覧表示

```bash
llm-info --list-gateways
```

設定ファイルに登録されているゲートウェイの一覧を表示します。

### 設定ソースの確認

```bash
llm-info --show-sources
```

現在の設定がどのソース（コマンドライン、環境変数、設定ファイル）から読み込まれたかを表示します。

### トピック別ヘルプの表示

```bash
llm-info --help-topic filter
llm-info --help-topic sort
llm-info --help-topic config
llm-info --help-topic examples
llm-info --help-topic errors
```

特定のトピックに関する詳細なヘルプを表示します。

### ヘルプの表示

```bash
llm-info --help
```

### バージョン情報の表示

```bash
llm-info --version
```

## モデル制約値の探索

llm-info v2.0では、実際のAPI動作をテストしてモデルの制約値を探索する機能が追加されました。

### Context Windowの探索

モデルが受け入れ可能な最大コンテキストトークン数を探索します。

```bash
# 基本的な使用方法
llm-info probe-context --model gpt-4o

# カスタムゲートウェイを使用
llm-info probe-context --model claude-3-opus --gateway production

# 詳細なログを表示
llm-info probe-context --model gpt-4o --verbose

# タイムアウトを延長
llm-info probe-context --model gpt-4o --timeout 60s

# 実行計画のみ表示（API呼び出しなし）
llm-info probe-context --model gpt-4o --dry-run
```

出力例：
```
Context Window Probe Results
============================
Model:                 GLM-4.6
Estimated Context:     127,000 tokens
Method Confidence:     high
Trials:                12
Duration:              45.3s
Max Input at Success:  126,800 tokens

Status: ✓ Success
```

verboseモードの場合、探索履歴も表示されます。
```
Search History:
------------------------------------------------------------
Trial    Tokens          Result       Message
------------------------------------------------------------
1        1,000           ✓            Success
2        10,000          ✗            Context length exceeded
3        127,000         ✗            Context length exceeded
4        126,800         ✓            Success at boundary
```

### Max Output Tokensの探索

モデルが生成可能な最大出力トークン数を探索します。

```bash
# 基本的な使用方法
llm-info probe-max-output --model gpt-4o

# カスタムゲートウェイを使用
llm-info probe-max-output --model claude-3-opus --url https://api.example.com

# 詳細なログを表示
llm-info probe-max-output --model gpt-4o --verbose --api-key your-key

# 実行計画のみ表示
llm-info probe-max-output --model gpt-4o --dry-run
```

出力例：
```
Max Output Tokens Probe Results
================================
Model:                 GLM-4.6
Max Output Tokens:     16,384
Evidence:              validation_error
Trials:                8
Duration:              30.2s
Max Successfully Gen:  8,192 tokens

Status: ✓ Success
```

### 探索コマンドのオプション

| オプション | 説明 |
|-----------|------|
| `--model` | 対象モデルID（必須） |
| `--url` | LLMゲートウェイのベースURL |
| `--api-key` | 認証用APIキー |
| `--gateway` | 設定ファイルのゲートウェイ名 |
| `--timeout` | リクエストタイムアウト（デフォルト: 30s） |
| `--config` | 設定ファイルのパス |
| `--verbose` | 詳細な探索履歴を表示 |
| `--dry-run` | 実行計画の表示のみ（API呼び出しなし） |
| `--help` | コマンド固有のヘルプを表示 |

## 探索機能の活用例

### 1. 新しいモデルの制約値調査

新しいLLMモデルを使用する前に、その実際の制約値を把握できます。

```bash
# モデルのコンテキスト制限を確認
llm-info probe-context --model new-model-2024 --verbose

# 出力生成の制限を確認
llm-info probe-max-output --model new-model-2024 --verbose
```

### 2. ゲートウェイの動作検証

ゲートウェイがドキュメント通りに動作しているか検証できます。

```bash
# 開発環境ゲートウェイのテスト
llm-info probe-context --model gpt-4o --gateway development --verbose

# 本番環境ゲートウェイのテスト
llm-info probe-context --model gpt-4o --gateway production --verbose
```

### 3. CI/CDでのリグレッション検出

CI/CDパイプラインでモデル制約値の変更を検出できます。

```yaml
# GitHub Actionsの例
- name: Check model constraints
  run: |
    llm-info probe-context --model gpt-4o --gateway staging > context.txt
    llm-info probe-max-output --model gpt-4o --gateway staging > output.txt
    # 期待値との比較処理を追加
```

## 出力の見方

### テーブル列の説明

- **Model Name**: モデルの識別子
- **Max Tokens**: モデルの最大トークン数
- **Mode**: モデルのモード（chat, completion, etc.）
- **Input Cost**: 入力トークンあたりのコスト

### 動的列制御

llm-infoは利用可能なデータに応じて表示列を自動調整します。すべての列が常に表示されるわけではなく、ゲートウェイから提供された情報のみが表示されます。

## 設定ファイルの詳細

### 設定ファイルの場所

デフォルトの設定ファイルの場所は `~/.config/llm-info/llm-info.yaml` です。`--config` オプションで別の場所を指定することもできます。

### 設定ファイルの構造

```yaml
# ゲートウェイ設定
gateways:
  # 本番環境ゲートウェイ
  - name: "production"
    url: "https://api.example.com"
    api_key: "your-production-api-key"
    timeout: "10s"
    description: "本番環境ゲートウェイ"
  
  # 開発環境ゲートウェイ
  - name: "development"
    url: "https://dev-api.example.com"
    api_key: "your-development-api-key"
    timeout: "5s"
    description: "開発環境ゲートウェイ"
  
  # ステージング用ゲートウェイ
  - name: "staging"
    url: "https://staging-api.example.com"
    api_key: "staging-api-key"
    timeout: "15s"
    description: "ステージング環境ゲートウェイ"
  
  # ローカルゲートウェイ
  - name: "local"
    url: "http://localhost:8000"
    api_key: ""
    timeout: "5s"
    description: "ローカル開発環境"

# デフォルトで使用するゲートウェイ名
default_gateway: "production"

# グローバル設定
global:
  # デフォルトのタイムアウト（各ゲートウェイで指定されていない場合に使用）
  timeout: "10s"
  
  # デフォルトの出力形式 (table, json)
  output_format: "table"
  
  # デフォルトのソート項目 (name, max_tokens, mode, input_cost)
  sort_by: "name"
  
  # デフォルトの表示列
  columns: "name,max_tokens,mode,input_cost"
  
  # 詳細ログを有効にするかどうか
  verbose: false
```

### 設定の優先順位

設定は以下の優先順位で適用されます：

1. コマンドライン引数（最優先）
2. 環境変数
3. 設定ファイル
4. デフォルト値

## よくあるユースケース

### 複数ゲートウェイの情報を取得

```bash
# 設定ファイルを使用
llm-info --gateway development
llm-info --gateway staging

# 直接URLを指定
llm-info --url https://gateway1.example.com/v1
llm-info --url https://gateway2.example.com/v1
```

### スクリプトでの使用

```bash
#!/bin/bash
# 環境変数を使用
export LLM_INFO_URL="https://gateway.example.com/v1"
export LLM_INFO_API_KEY="your-api-key"
export LLM_INFO_OUTPUT_FORMAT="json"

# JSON出力を変数に格納
models=$(llm-info)
echo "Available models: $models"
```

### CI/CDパイプラインでの使用

```bash
#!/bin/bash
# 設定ファイルを環境に応じて切り替え
if [ "$ENVIRONMENT" = "production" ]; then
    llm-info --gateway production
else
    llm-info --gateway development
fi
```

### JSON出力を他のツールと連携

```bash
# jqを使用してモデル名のみ抽出
llm-info --url https://api.example.com/v1 --output json | jq -r '.[].name'

# モデル数をカウント
llm-info --url https://api.example.com/v1 --output json | jq 'length'
```

## トラブルシューティング

### 接続エラー

**症状**: `connection refused` や `no such host` などのエラー

**原因**: 
- URLが間違っている
- ゲートウェイが停止している
- ネットワーク接続の問題

**解決策**:
- URLが正しいことを確認
- ゲートウェイが稼働していることを確認
- ネットワーク接続を確認

### 認証エラー

**症状**: `401 Unauthorized` エラー

**原因**: 
- APIキーが間違っている
- APIキーに必要な権限がない

**解決策**:
- APIキーを再確認
- APIキーにモデル情報へのアクセス権限があることを確認

### タイムアウトエラー

**症状**: `timeout` エラー

**原因**: 
- ネットワーク遅延
- ゲートウェイの応答遅延

**解決策**:
- `--timeout`フラグで時間を延長
- ネットワーク接続を確認

### エンドポイント未対応

**症状**: `404 Not Found` エラー

**原因**:
- ゲートウェイがLiteLLM互換またはOpenAI標準互換でない
- URLが間違っている

**解決策**:
- LiteLLM互換またはOpenAI標準互換のゲートウェイを使用
- URLが正しいことを確認
- 自動フォールバック機能により、LiteLLMエンドポイントが利用できない場合はOpenAI標準エンドポイントが試行されます

### レート制限

**症状**: `429 Too Many Requests` エラー

**原因**: 
- API呼び出し回数が制限を超えた

**解決策**:
- しばらく待ってから再試行
- APIプランの確認

## 開発者向け情報

### 開発環境のセットアップ

```bash
git clone https://github.com/armaniacs/llm-info.git
cd llm-info
go mod download
```

### テストの実行

```bash
make test              # 単体テスト
make test-coverage     # カバレッジ付きテスト
```

### ビルドターゲット

- `make build`: バイナリビルド
- `make test`: テスト実行
- `make test-coverage`: カバレッジ付きテスト
- `make lint`: コード品質チェック
- `make clean`: クリーンアップ
- `make install`: バイナリをインストール
- `make run-example`: サンプルゲートウェイで実行
- `make help`: ヘルプ表示

### 貢献方法

1. Issue報告: バグや機能要求をIssueで報告
2. プルリクエスト: コードの改善や新機能を提案
3. 詳細は[README.md](README.md)と[CONTRIBUTING.md](CONTRIBUTING.md)を参照

## 詳細情報へのリンク

- [README.md](README.md): 包括的なプロジェクト情報
- [API仕様](plan/spec.md): LiteLLM互換エンドポイント仕様
- [開発ガイド](plans/mvp-development-guide.md): 開発者向け詳細情報
- [ロードマップ](plans/llm-info-roadmap.md): 将来の機能計画

## よくある質問

### Q: どのゲートウェイで動作しますか？

A: LiteLLM互換の`/model/info`エンドポイントまたはOpenAI標準互換の`/v1/models`エンドポイントを実装しているゲートウェイで動作します。ツールはまずLiteLLMエンドポイントを試行し、失敗した場合は自動的にOpenAI標準エンドポイントにフォールバックします。

### Q: 設定ファイルは使用できますか？

A: はい、YAML形式の設定ファイルに対応しています。`~/.config/llm-info/llm-info.yaml` に配置するか、`--config` オプションで指定してください。

### Q: 出力をJSON形式で取得できますか？

A: はい、`--output json` オプションでJSON形式での出力に対応しています。

### Q: 環境変数は使用できますか？

A: はい、`LLM_INFO_URL`、`LLM_INFO_API_KEY` などの環境変数に対応しています。

### Q: プロキシ環境下で使用できますか？

A: 現在はプロキシ対応していません。将来の機能で対応予定です。

## 探索結果の理解

### ステータス表示

- `✓ Success`: 探索が正常完了
- `✗ Failed`: 探索が失敗

### 確信度（Method Confidence）

- `high`: エラーメッセージから正確な値を取得
- `medium`: 二分探索で境界を特定
- `low`: 上限が見つからず、推定値

### エビデンス（Evidence - Max Output）

- `validation_error`: バリデーションエラーメッセージから検出
- `max_output_incomplete`: 出力が途中で途切れた場合に検出
- `success`: 設定された値で正常に生成完了

## コマンドリファレンス

### 基本構文

```bash
llm-info [オプション]
llm-info probe-context --model <MODEL_ID> [オプション]
llm-info probe-max-output --model <MODEL_ID> [オプション]
```

### オプション

| オプション | 説明 | 必須 | デフォルト値 |
|-----------|------|------|-------------|
| `--url` | LLMゲートウェイのベースURL | いいえ¹ | - |
| `--api-key` | 認証用APIキー | いいえ | - |
| `--timeout` | リクエストタイムアウト | いいえ | 10s |
| `--config` | 設定ファイルのパス | いいえ | ~/.config/llm-info/llm-info.yaml |
| `--gateway` | 使用するゲートウェイ名 | いいえ | default |
| `--format` | 出力形式 (table, json) | いいえ | table |
| `--sort` | ソート項目 (name, max_tokens, mode, input_cost) | いいえ | name |
| `--filter` | フィルタ条件 (例: 'name:gpt,tokens>1000,mode:chat') | いいえ | - |
| `--columns` | 表示列 (例: 'name,max_tokens') | いいえ | すべて |
| `--verbose` | 詳細ログを表示 | いいえ | false |
| `--init-config` | 設定ファイルテンプレートを作成 | いいえ | - |
| `--check-config` | 設定ファイルを検証 | いいえ | - |
| `--list-gateways` | 設定済みゲートウェイを一覧表示 | いいえ | - |
| `--show-sources` | 設定ソース情報を表示 | いいえ | - |
| `--help-topic` | トピック別ヘルプを表示 | いいえ | - |
| `--help` | ヘルプメッセージを表示 | いいえ | - |
| `--version` | バージョン情報を表示 | いいえ | - |

¹ `--url` は設定ファイルまたは環境変数で指定されていない場合に必須です。

### 環境変数

| 環境変数 | 説明 | デフォルト値 |
|----------|------|-------------|
| `LLM_INFO_URL` | LLMゲートウェイのベースURL | - |
| `LLM_INFO_API_KEY` | 認証に使用するAPIキー | - |
| `LLM_INFO_TIMEOUT` | リクエストタイムアウト (例: 10s, 1m) | 10s |
| `LLM_INFO_DEFAULT_GATEWAY` | デフォルトゲートウェイ名 | default |
| `LLM_INFO_OUTPUT_FORMAT` | 出力形式 (table, json) | table |
| `LLM_INFO_SORT_BY` | ソート項目 (name, max_tokens, mode, input_cost) | name |
| `LLM_INFO_FILTER` | フィルタ条件 | - |
| `LLM_INFO_COLUMNS` | 表示列 | - |
| `LLM_INFO_CONFIG_PATH` | 設定ファイルのパス | ~/.config/llm-info/llm-info.yaml |
| `LLM_INFO_VERBOSE` | 詳細ログを有効にする | false |
| `LLM_INFO_DEBUG` | デバッグモードを有効にする | false |
| `LLM_INFO_USER_AGENT` | ユーザーエージェント | llm-info/1.0.0 |

### 環境変数の詳細

#### 基本設定

- **LLM_INFO_URL**: LLMゲートウェイのベースURLを指定します。コマンドラインの`--url`オプションに相当します。
- **LLM_INFO_API_KEY**: 認証に使用するAPIキーを指定します。コマンドラインの`--api-key`オプションに相当します。
- **LLM_INFO_TIMEOUT**: リクエストのタイムアウト時間を指定します。`10s`、`1m`、`30s`などの形式で指定できます。
- **LLM_INFO_DEFAULT_GATEWAY**: 設定ファイルから使用するデフォルトのゲートウェイ名を指定します。

#### 表示設定

- **LLM_INFO_OUTPUT_FORMAT**: 出力形式を指定します。`table`または`json`を指定できます。
- **LLM_INFO_SORT_BY**: モデルのソート項目を指定します。`name`、`max_tokens`、`mode`、`input_cost`などを指定できます。
- **LLM_INFO_FILTER**: モデルのフィルタ条件を指定します。`name:gpt,tokens>1000,mode:chat`のような形式で指定できます。
- **LLM_INFO_COLUMNS**: 表示する列を指定します。`name,max_tokens,mode`のようにカンマ区切りで指定できます。

#### 詳細設定

- **LLM_INFO_CONFIG_PATH**: 設定ファイルのパスを指定します。
- **LLM_INFO_VERBOSE**: 詳細ログを有効にする場合は`true`を指定します。
- **LLM_INFO_DEBUG**: デバッグモードを有効にする場合は`true`を指定します。
- **LLM_INFO_USER_AGENT**: HTTPリクエストのUser-Agentヘッダーを指定します。

### 環境変数の使用例

#### 基本使用例

```bash
# 環境変数でゲートウェイを設定
export LLM_INFO_URL="https://api.example.com"
export LLM_INFO_API_KEY="your-api-key"
llm-info

# 出力形式を指定
export LLM_INFO_OUTPUT_FORMAT="json"
llm-info

# ソートとフィルタを指定
export LLM_INFO_SORT_BY="max_tokens"
export LLM_INFO_FILTER="gpt"
llm-info
```

#### CI/CDパイプラインでの使用例

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

#### Dockerコンテナでの使用例

```dockerfile
FROM alpine:latest
RUN apk add --no-cache llm-info
ENV LLM_INFO_URL="https://api.example.com"
ENV LLM_INFO_API_KEY=""
CMD ["llm-info"]
```

#### 設定ソースの確認

```bash
# 環境変数を設定
export LLM_INFO_URL="https://api.example.com"
export LLM_INFO_API_KEY="your-api-key"

# 設定ソース情報を表示
llm-info --show-sources
```

出力例：
```
Configuration sources:
  gateway: environment variable
  output_format: default
  sort_by: default
```

### 使用例

```bash
# 基本使用
llm-info --url https://api.example.com/v1

# APIキーを使用
llm-info --url https://api.example.com/v1 --api-key sk-1234567890

# 設定ファイルを使用
llm-info --gateway development

# 環境変数を使用
export LLM_INFO_URL=https://api.example.com/v1
export LLM_INFO_API_KEY=sk-1234567890
llm-info

# JSON出力
llm-info --url https://api.example.com/v1 --format json

# ソートとフィルタリング
llm-info --url https://api.example.com/v1 --sort max_tokens --filter "name:gpt"

# 表示列のカスタマイズ
llm-info --url https://api.example.com/v1 --columns "name,max_tokens"

# タイムアウトを30秒に設定
llm-info --url https://api.example.com/v1 --timeout 30s

# 設定ファイルテンプレートの作成
llm-info --init-config

# 設定ファイルの検証
llm-info --check-config

# 設定済みゲートウェイの一覧表示
llm-info --list-gateways

# 設定ソースの確認
llm-info --show-sources

# トピック別ヘルプの表示
llm-info --help-topic filter

# ヘルプ表示
llm-info --help

# バージョン表示
llm-info --version
```

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルを参照してください。