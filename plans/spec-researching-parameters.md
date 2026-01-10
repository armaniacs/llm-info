# OpenAI Model Limits Probe / Go 実装仕様書

2026-01-10 16:55 時点では、実装にはとりかからない。v2.0以降の将来仕様である。

## 目的

本ツールは、OpenAI API に登録されている任意のモデルについて、公式ドキュメントに明示されていない、または信頼できない以下の制約値を **API の実挙動から推定**することを目的とする。

1. 最大コンテキスト長（context window）

   * 入力トークンと出力トークンの合計上限
2. 最大出力トークン数（max output tokens）

   * 単一レスポンスで生成可能な最大トークン数

推定結果は再現可能かつ機械可読な形式で出力し、設計・運用・CI 検証等に利用可能とする。

参考:

* OpenAI Responses API Reference
  [https://platform.openai.com/docs/api-reference/responses](https://platform.openai.com/docs/api-reference/responses)
* Migrate to the Responses API
  [https://platform.openai.com/docs/guides/migrate-to-responses](https://platform.openai.com/docs/guides/migrate-to-responses)
* Controlling the length of model responses
  [https://help.openai.com/en/articles/5072518-controlling-the-length-of-openai-model-responses](https://help.openai.com/en/articles/5072518-controlling-the-length-of-openai-model-responses)

### LM実効コンテキスト容量の概算検証

1. 目的 対象LLMの有効なコンテキストウィンドウの境界値を、4K / 8K / 16K / 32K / 64K / 96K / 128K の段階的なステップで特定する。

2. 検証手法（Needle In A Haystack 簡易版）

テストデータ: 意味を持たない、あるいは一般的な長文（青空文庫や公文書など）の末尾に、特定の「検証用キーワード」を挿入したデータセットを作成する。

判定基準: キーワードを正確に抽出・回答できた場合、そのトークン長を「有効範囲内」とみなす。

絞り込み手順:

まず、べき乗（4K, 8K, 16K...）で段階的に検証する。

モデルが回答不能、または誤回答した直前の「成功した値」と「失敗した値」の中間値（例：64Kで成功、128Kで失敗した場合は96K）を最終確認ステップとして実行する。

#### ステップ,トークン数（目安）,判定の役割
Step 1-5,4K / 8K / 16K / 32K / 64K,基礎性能の確認
Step 6,128K,上限閾値の確認
Step 7 (補間),96K,64K-128K間の解像度向上（二分法）

#### 効率的な運用のアドバイス
トークン計算のゆとり: 日本語の場合、1文字が1トークン以上になることが多いため、文字数で指定する場合は「想定トークン数 × 0.7〜0.8」程度の文字数でテストデータを用意すると安全です。

ロスト・イン・ザ・ミドル現象: LLMは「文末」よりも「文の中央」にある情報を忘れやすい性質があります。より厳密に測る場合は、キーワードを文末ではなく、テキスト全体の80%付近に埋め込むと、実用的な限界値が見えやすくなります。

* **温度設定（Temperature）:** 推論の揺らぎを抑えるため、`Temperature = 0` または可能な限り低い値に設定して実行する。
* **パスの定義:** 「情報の埋め込み位置」によって結果が変わるため、**「文末から1,000トークン手前」**など、埋め込み位置を固定する（Needle In A Haystackテストの標準的な手法）。


3. テストデータ生成の仕様

「大雑把に」とはいえ、手動でテキストを用意するのは現実的ではありません。以下のロジックでスクリプト（Python等）を用いて生成することを仕様に盛り込みます。

#### テストデータの構造

1. **Prefix（冒頭）:** 「以下の内容を記憶してください。」等の定型句。
2. **Body（本文）:** 無関係な文章の繰り返し。
* 例：著作権フリーの小説（青空文庫）、ニュース記事、または「これはテスト用のダミー文章です。」というフレーズの1,000回繰り返し。


3. **Target（埋め込み情報）:** 判定用のユニークな情報。
* 例：「[日付]のラッキーカラーは[色]です。」


4. **Suffix（末尾指示）:** 「上記のデータにおいて、ラッキーカラーは何と記載されていましたか？」という問い。

#### トークン数制御の目安（日本語の場合）

正確なトークナイザー（tiktoken等）を使わない場合の、大まかな文字数計算式です。

* **1K (1,024 tokens) ≈ 約 700〜800文字**
* **96K (98,304 tokens) ≈ 約 70,000〜80,000文字**

5. 実機検証用ログシート（仕様書添付用）

検証結果を記録し、二分法で「一段階」絞り込むためのチェック表です。

| 試行順 | 設定サイズ | 成功(○) / 失敗(×) | 次のステップ |
| --- | --- | --- | --- |
| 1 | 64K |  | 成功なら次へ |
| 2 | 128K |  | 失敗なら間を検証 |
| **3** | **96K** |  | **ここで境界を確定** |

> **判定ルール:** > * 64K(○) かつ 128K(×) の場合 → **96K** をテスト。
> * 96K(○) ならば、「実効容量は **96K〜128K** の間」と結論付ける。
> * 96K(×) ならば、「実効容量は **64K〜96K** の間」と結論付ける。

---

## スコープ

### 対象

* API: OpenAI Responses API
* 実装言語: Go
* 実行形態: CLI ツール
* 対象モデル: `model` パラメータで指定された任意のモデル ID

### 非スコープ

* モデル品質評価（精度・推論性能など）
* コスト最適化アルゴリズム
* 会話状態（previous_response_id）を使った multi-turn 探索
* プロンプトエンジニアリングの最適化

---

## 前提・設計方針

* ローカルで正確な tokenization（tiktoken 等）を再現しない

  * API が返す **エラーメッセージ・status・reason** を一次情報とする
* 制約値は「固定値」ではなく「推定値」として扱う
* 実装は **課金爆発を防ぐ安全制御**を必須とする
* OpenAI 公式 Go SDK を第一候補としつつ、HTTP 実装に差し替え可能な抽象を用意する

---

## 用語定義

| 用語                | 定義                            |
| ----------------- | ----------------------------- |
| context window    | 入力 + 出力 + 内部オーバーヘッドを含む総トークン上限 |
| max output tokens | 単一レスポンスで生成可能な最大トークン数          |
| probe             | 上限推定のための一連の API 呼び出し          |
| trial             | 1 回の API 呼び出し                 |

---

## 機能要件

### 共通要件

#### CLI 引数

`--model` 以外の引数の未指定時は configから推定する     

| 引数             | 必須  | 説明                                |
| -------------- | --- | --------------------------------- |
| `--model`      | Yes | 対象モデル ID                          |
| `--api-key`    | No  |        |
| `--base-url`   | No  |  |
| `--timeout`    | No  | HTTP タイムアウト（秒、default 30）         |
| `--max-trials` | No  | 試行回数上限（default 40）                |
| `--out`        | No  | 結果 JSON の出力先                      |
| `--verbose`    | No  | 詳細ログ出力                            |
| `--dry-run`    | No  | API 呼び出しせず探索計画のみ表示                |

#### 出力

* JSON 形式
* 標準出力および任意のファイル出力
* 各 trial の要約情報を含む（省略オプション可）

---

### context window 推定

#### 方針

* 入力サイズを段階的に増やし、**context 超過エラー**を発生させる
* エラーメッセージに含まれる最大 context 長を抽出
* 抽出できない場合は観測値ベースで近似

#### 要件

* 出力トークンは最小（例: `max_output_tokens = 16`）に固定
* 入力は token 増加が単調な生成方法を使用

  * 例: `"x "` の反復
* 探索は以下の 2 フェーズ

##### フェーズ A: 指数探索

1. 初期入力規模 `N0`（例: 4096）
2. 成功する限り `N = 2N`
3. context 超過エラー発生で停止

##### フェーズ B: 観測値主導の二分探索

* 成功時・失敗時に API が返す token 数を利用
* ローカル token 計測には依存しない
* 収束条件:

  * 差分 ≤ 128 tokens または `max-trials` 到達

#### 出力項目

```json
{
  "estimated_max_context_tokens": 128000,
  "max_input_tokens_at_success": 127600,
  "method_confidence": "high|medium|low"
}
```

---

### max output tokens 推定

#### 方針

以下のいずれかを観測して上限を推定する。

1. `max_output_tokens` 過大指定時のバリデーションエラー
2. `status: "incomplete"` かつ
   `incomplete_details.reason == "max_output_tokens"`

#### 探索方法

* 初期値例: 512
* 指数的に増加: 512 → 1024 → 2048 → ...
* 境界検出後、二分探索で収束

#### 出力項目

```json
{
  "estimated_max_output_tokens": 16384,
  "evidence": "validation_error|max_output_incomplete",
  "observed_incomplete_reason": "max_output_tokens"
}
```

---

## API 呼び出し仕様

### エンドポイント

```
POST /v1/responses
```

### 認証

```
Authorization: Bearer {OPENAI_API_KEY}
```

### 共通パラメータ

| パラメータ               | 用途           |
| ------------------- | ------------ |
| `model`             | 対象モデル        |
| `input`             | 生成入力         |
| `max_output_tokens` | 探索変数または固定最小値 |
| `temperature`       | 0 固定         |

---

## エラー解析仕様

### 対象エラー

* HTTP 400 系
* `invalid_request_error`
* context 超過、出力上限超過

### 抽出対象（正規表現例）

* `maximum context length is (\d+) tokens`
* `your .* resulted in (\d+) tokens`

### incomplete 解析

* `response.status == "incomplete"`
* `incomplete_details.reason == "max_output_tokens"`

---

## 出力フォーマット（JSON）

```json
{
  "model": "gpt-4o-mini",
  "run_id": "2026-01-10T12:34:56Z",
  "context_window": { ... },
  "max_output_tokens": { ... },
  "trials": [ ... ],
  "cost_controls": {
    "max_trials": 40,
    "timeout_seconds": 30
  }
}
```

注意:

* 入力本文は保存しない
* サイズ・ハッシュのみ保持

---

## 非機能要件

### 安全性

* 試行回数・最大入力サイズに上限を設ける
* タイムアウト必須

### 再現性

* プロンプトテンプレート固定
* temperature = 0
* 乱数不使用

### 可観測性

* verbose モードで trial 単位ログ
* エラー raw データ保持

---

## 受入基準

* context 超過エラーから最大 context 長を抽出できる
* max output 上限に起因する incomplete を観測できる
* API キー未設定時に即時失敗する
* 入力本文がログ・出力に残らない

---

## 補足

* context window と max output tokens は **モデル仕様ではなく挙動仕様**として扱う
* 推定不能な場合は null + 理由を必ず出力
* CI 用に低コストプロファイルを別途定義することを推奨

