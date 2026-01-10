package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// HelpProvider はヘルプ機能を提供する
type HelpProvider struct {
	version string
}

// NewHelpProvider は新しいヘルププロバイダーを作成する
func NewHelpProvider(version string) *HelpProvider {
	return &HelpProvider{version: version}
}

// ShowGeneralHelp は一般的なヘルプを表示する
func (hp *HelpProvider) ShowGeneralHelp() {
	fmt.Printf(`llm-info - LLMゲートウェイ情報可視化ツール (バージョン: %s)

使用方法:
  llm-info [flags]

フラグ:
`, hp.version)

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "  --url string\tゲートウェイのURL")
	fmt.Fprintln(w, "  --api-key string\tAPIキー")
	fmt.Fprintln(w, "  --gateway string\t使用するゲートウェイ名")
	fmt.Fprintln(w, "  --timeout duration\tリクエストタイムアウト (デフォルト: 10s)")
	fmt.Fprintln(w, "  --format string\t出力形式 (table|json) (デフォルト: table)")
	fmt.Fprintln(w, "  --filter string\tフィルタ条件")
	fmt.Fprintln(w, "  --sort string\tソート条件")
	fmt.Fprintln(w, "  --columns string\t表示するカラム (カンマ区切り)")
	fmt.Fprintln(w, "  --config string\t設定ファイルパス")
	fmt.Fprintln(w, "  --verbose\t詳細なログを表示")
	fmt.Fprintln(w, "  --help\tヘルプを表示")
	fmt.Fprintln(w, "  --version\tバージョンを表示")
	fmt.Fprintln(w, "  --init-config\t設定ファイルのテンプレートを作成")
	fmt.Fprintln(w, "  --check-config\t設定ファイルを検証")
	fmt.Fprintln(w, "  --list-gateways\t登録済みゲートウェイを一覧表示")
	w.Flush()

	fmt.Println(`
使用例:
  # 基本使用
  llm-info --url https://api.example.com --api-key your-key
  
  # 設定ファイルを使用
  llm-info --gateway production
  
  # フィルタリングとソート
  llm-info --filter "gpt" --sort "tokens"
  
  # JSON出力
  llm-info --format json
  
詳細なヘルプ:
  llm-info --help filter    # フィルタ構文のヘルプ
  llm-info --help sort      # ソートオプションのヘルプ
  llm-info --help config    # 設定ファイルのヘルプ
  llm-info --help examples  # 使用例のヘルプ
  llm-info --help errors    # エラーメッセージのヘルプ
`)
}

// ShowFilterHelp はフィルタ構文のヘルプを表示する
func (hp *HelpProvider) ShowFilterHelp() {
	fmt.Println(`フィルタ構文ヘルプ

基本構文:
  --filter "条件1,条件2,..."

使用可能な条件:
  name:パターン          モデル名でフィルタ（部分一致）
  exclude:パターン       モデル名で除外（部分一致）
  tokens>数値           最大トークン数が指定値より大きい
  tokens<数値           最大トークン数が指定値より小さい
  cost>数値             入力コストが指定値より大きい
  cost<数値             入力コストが指定値より小さい
  mode:値               モードでフィルタ（chat/completion）

使用例:
  llm-info --filter "gpt"                           # GPTモデルのみ
  llm-info --filter "name:gpt,tokens>1000"          # GPTでトークン数>1000
  llm-info --filter "exclude:beta,cost<0.01"        # ベータ版除外でコスト<0.01
  llm-info --filter "mode:chat,tokens>4000"         # チャットモードでトークン数>4000

ヒント:
  - 条件はカンマ(,)で区切って複数指定できます
  - 条件はAND条件で結合されます
  - 大文字小文字は区別されません
  - ワイルドカード(*)は使用できません
`)
}

// ShowSortHelp はソートオプションのヘルプを表示する
func (hp *HelpProvider) ShowSortHelp() {
	fmt.Println(`ソートオプションヘルプ

基本構文:
  --sort "フィールド"        # 昇順
  --sort "-フィールド"       # 降順

使用可能なフィールド:
  name, model              モデル名
  tokens, max_tokens       最大トークン数
  cost, input_cost         入力コスト
  mode                     モード

使用例:
  llm-info --sort "name"           # 名前の昇順
  llm-info --sort "-tokens"        # トークン数の降順
  llm-info --sort "cost"           # コストの昇順

ヒント:
  - マイナス(-)を付けると降順になります
  - デフォルトは昇順です
  - 複数のフィールドでのソートはできません
`)
}

// ShowConfigHelp は設定ファイルのヘルプを表示する
func (hp *HelpProvider) ShowConfigHelp() {
	fmt.Println(`設定ファイルヘルプ

設定ファイルの場所:
  ~/.config/llm-info/llm-info.yaml

設定ファイル形式:
  gateways:
    - name: "production"
      url: "https://api.example.com"
      api_key: "your-api-key"
      timeout: "10s"
    - name: "development"
      url: "https://dev-api.example.com"
      api_key: "dev-api-key"
      timeout: "5s"
  
  default_gateway: "production"
  
  global:
    timeout: "10s"
    output_format: "table"
    sort_by: "name"

環境変数:
  LLM_INFO_URL           デフォルトのゲートウェイURL
  LLM_INFO_API_KEY       デフォルトのAPIキー
  LLM_INFO_CONFIG_PATH   設定ファイルのパス
  LLM_INFO_DEBUG         デバッグモードを有効にする

コマンド:
  llm-info --init-config     # 設定ファイルのテンプレートを作成
  llm-info --check-config    # 設定ファイルを検証
  llm-info --list-gateways   # 登録済みゲートウェイを一覧表示

優先順位:
  1. コマンドライン引数
  2. 環境変数
  3. 設定ファイル
  4. デフォルト値
`)
}

// ShowExamplesHelp は使用例のヘルプを表示する
func (hp *HelpProvider) ShowExamplesHelp() {
	fmt.Println(`使用例ヘルプ

基本使用例:
  # 直接指定
  llm-info --url https://api.openai.com --api-key sk-xxx
  
  # 設定ファイル使用
  llm-info --gateway production
  
  # 環境変数使用
  export LLM_INFO_URL="https://api.example.com"
  export LLM_INFO_API_KEY="your-key"
  llm-info

フィルタリング例:
  # GPTモデルのみ
  llm-info --filter "gpt"
  
  # 高トークン数モデル
  llm-info --filter "tokens>32000"
  
  # 安価なモデル
  llm-info --filter "cost<0.001"
  
  # 複合条件
  llm-info --filter "name:gpt-4,tokens>8000"

ソート例:
  # トークン数の降順
  llm-info --sort "-tokens"
  
  # コストの昇順
  llm-info --sort "cost"

出力形式例:
  # JSON出力
  llm-info --format json
  
  # 特定カラムのみ
  llm-info --columns "name,tokens"
  
  # スクリプトでの使用
  llm-info --format json | jq '.models[] | select(.max_tokens > 10000)'

CI/CDでの使用例:
  # GitHub Actions
  - name: List models
    env:
      LLM_INFO_URL: ${{ secrets.API_URL }}
      LLM_INFO_API_KEY: ${{ secrets.API_KEY }}
    run: llm-info --format json > models.json
  
  # Docker
  docker run --rm \
    -e LLM_INFO_URL="https://api.example.com" \
    -e LLM_INFO_API_KEY="your-key" \
    llm-info:latest

トラブルシューティング:
  # 詳細なログを表示
  llm-info --verbose
  
  # 設定ファイルを検証
  llm-info --check-config
  
  # タイムアウトを延長
  llm-info --timeout 30s
`)
}

// ShowErrorsHelp はエラーメッセージのヘルプを表示する
func (hp *HelpProvider) ShowErrorsHelp() {
	fmt.Println(`エラーメッセージヘルプ

エラーの種類:
  1. ネットワークエラー
     - 接続タイムアウト
     - DNS解決失敗
     - TLS証明書エラー
     - 接続拒否

  2. APIエラー
     - 認証失敗 (401)
     - 認可失敗 (403)
     - エンドポイント不在 (404)
     - レート制限 (429)
     - サーバーエラー (5xx)

  3. 設定エラー
     - 設定ファイル不在
     - 設定形式エラー
     - 必須項目不在

  4. ユーザーエラー
     - 不正な引数
     - 不正なフィルタ構文
     - 不正なソート条件

一般的な解決策:
  - ネットワーク接続を確認する
  - APIキーを確認する
  - 設定ファイルを検証する
  - ヘルプを参照する: llm-info --help

詳細なヘルプ:
  - ネットワークエラー: https://github.com/armaniacs/llm-info/wiki/network-errors
  - APIエラー: https://github.com/armaniacs/llm-info/wiki/api-errors
  - 設定エラー: https://github.com/armaniacs/llm-info/wiki/config-errors
  - トラブルシューティング: https://github.com/armaniacs/llm-info/wiki/troubleshooting

デバッグ方法:
  # 詳細なログを表示
  llm-info --verbose
  
  # 設定ファイルを検証
  llm-info --check-config
  
  # バージョン情報を表示
  llm-info --version
`)
}

// ShowTopicHelp はトピック別のヘルプを表示する
func (hp *HelpProvider) ShowTopicHelp(topic string) {
	switch strings.ToLower(topic) {
	case "filter":
		hp.ShowFilterHelp()
	case "sort":
		hp.ShowSortHelp()
	case "config":
		hp.ShowConfigHelp()
	case "examples":
		hp.ShowExamplesHelp()
	case "errors":
		hp.ShowErrorsHelp()
	default:
		fmt.Printf("不明なトピック: %s\n", topic)
		fmt.Println("利用可能なトピック: filter, sort, config, examples, errors")
		fmt.Println("詳細なヘルプについては、以下のコマンドを実行してください:")
		fmt.Println("  llm-info --help")
	}
}

// ShowVersion はバージョン情報を表示する
func (hp *HelpProvider) ShowVersion() {
	fmt.Printf("llm-info version %s\n", hp.version)
	fmt.Println("Copyright (c) 2024 llm-info contributors")
	fmt.Println("License: MIT")
	fmt.Println("Repository: https://github.com/armaniacs/llm-info")
}

// ShowConfigTemplate は設定ファイルのテンプレートを表示する
func (hp *HelpProvider) ShowConfigTemplate() {
	fmt.Println(`# llm-info 設定ファイルテンプレート
# このファイルを ~/.config/llm-info/llm-info.yaml に保存してください

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
  
  # テスト環境ゲートウェイ
  - name: "test"
    url: "https://test-api.example.com"
    api_key: "your-test-api-key"
    timeout: "10s"
    description: "テスト環境ゲートウェイ"

# デフォルトゲートウェイ
default_gateway: "production"

# グローバル設定
global:
  # デフォルトのタイムアウト
  timeout: "10s"
  
  # デフォルトの出力形式 (table|json)
  output_format: "table"
  
  # デフォルトのソートフィールド
  sort_by: "name"
  
  # デフォルトで表示するカラム
  columns: "name,tokens,cost,mode"
  
  # 詳細ログを有効にする (true|false)
  verbose: false

# 環境変数の設定例:
# export LLM_INFO_URL="https://api.example.com"
# export LLM_INFO_API_KEY="your-api-key"
# export LLM_INFO_CONFIG_PATH="/path/to/config.yaml"
# export LLM_INFO_DEBUG="true"
`)
}
