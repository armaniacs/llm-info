# E2E Tests

このディレクトリにはllm-infoのエンドツーエンドテストが含まれています。

## 実行方法

### 開発環境

```bash
# 環境のセットアップ
make test-e2e-setup

# E2Eテストの実行
make test-e2e

# クリーンアップ
make test-e2e-clean
```

### 環境変数の設定

```bash
export LLM_INFO_BIN_PATH="./test/bin/llm-info"
export LLM_INFO_TEST_API_URL="http://localhost:8080"
export LLM_INFO_TEST_API_KEY="test-key"
go test ./test/e2e -v
```

### CI/CD

GitHub Actionsが自動で実行します。手動で実行する場合：

```bash
# すべてのテストを実行
make test-e2e-all
```

## モックサーバーのセットアップ

E2Eテストの中にはローカルでモックサーバーが必要なものがあります。

```bash
# テスト用のモックサーバーを起動
make run-mock-server &
MOCK_PID=$!

# テスト実行
make test-e2e

# クリーンアップ
kill $MOCK_PID
```

## トラブルシューティング

### 「binary not found」エラー

```bash
# バイナリをビルド
make test-e2e-setup

# または
go build -o test/bin/llm-info cmd/llm-info/*.go
```

### 「connection refused」エラー

テストがモックサーバーを要求しています。サーバーを起動するか、該当テストをスキップしてください。

## 前提条件

- Go 1.21+
- Make
- Docker（オプション：コンテナでテストする場合）