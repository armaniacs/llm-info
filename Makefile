.PHONY: build test clean lint install run-example

BINARY_NAME=llm-info
BUILD_DIR=bin

# デフォルトターゲット
all: build

# ビルド
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/llm-info/*.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# テスト実行
test:
	@echo "Running tests..."
	go test -v ./...

# カバレッジ付きテスト
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# クリーンアップ
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# リンティング
lint:
	@echo "Running linter..."
	golangci-lint run

# インストール
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install cmd/llm-info/*.go
	@echo "Install complete"

# サンプル実行
run-example: build
	@echo "Running example..."
	$(BUILD_DIR)/$(BINARY_NAME) --url https://openrouter.ai/api

# ヘルプ表示
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linter"
	@echo "  install        - Install the binary"
	@echo "  run-example    - Run with example gateway"
	@echo "  help           - Show this help message"