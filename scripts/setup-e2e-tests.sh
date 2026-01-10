#!/bin/bash
set -e

echo "Setting up E2E test environment..."

# テスト用ディレクトリの作成
mkdir -p test/bin
mkdir -p test/logs
mkdir -p test/configs
mkdir -p test/tmp

# バイナリのビルド
echo "Building llm-info binary for E2E tests..."
go build -o test/bin/llm-info cmd/llm-info/*.go

# テスト用設定の準備
cat > test/configs/test.yaml << EOF
gateways:
  - name: "test-gateway"
    url: "http://localhost:8080"
    api_key: "test-key"
    timeout: "10s"

default_gateway: "test-gateway"

global:
  timeout: "10s"
  output_format: "table"
EOF

# 権限の設定
chmod +x test/bin/llm-info

echo "E2E test environment ready!"
echo "Binary: test/bin/llm-info"
echo "Config: test/configs/test.yaml"