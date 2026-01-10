.PHONY: build test clean lint install uninstall run-example help test-e2e test-e2e-setup test-e2e-clean test-e2e-all

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
install:
	@echo "Installing $(BINARY_NAME)..."
	@echo "Note: Will be installed to:"
	@if [ -n "$$(go env GOBIN)" ]; then \
		echo "  $$GOBIN/$(BINARY_NAME)"; \
	else \
		echo "  $$(go env GOPATH)/bin/$(BINARY_NAME)"; \
	fi
	go install ./cmd/llm-info
	@echo "Install complete"

# アンインストール
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@if [ -f "$$(go env GOBIN)/$(BINARY_NAME)" ]; then \
		rm "$$(go env GOBIN)/$(BINARY_NAME)"; \
		echo "Removed: $$(go env GOBIN)/$(BINARY_NAME)"; \
	elif [ -f "$$(go env GOPATH)/bin/$(BINARY_NAME)" ]; then \
		rm "$$(go env GOPATH)/bin/$(BINARY_NAME)"; \
		echo "Removed: $$(go env GOPATH)/bin/$(BINARY_NAME)"; \
	else \
		echo "$(BINARY_NAME) not found in installation directory"; \
	fi
	@echo "Uninstall complete"

# サンプル実行
run-example: build
	@echo "Running example..."
	$(BUILD_DIR)/$(BINARY_NAME) --url https://openrouter.ai/api

# E2Eテストの準備
test-e2e-setup:
	@echo "Setting up E2E test environment..."
	@mkdir -p test/bin
	@go build -o test/bin/llm-info cmd/llm-info/*.go
	@echo "E2E binary ready at test/bin/llm-info"

# E2Eテストの実行
test-e2e: test-e2e-setup
	@echo "Running E2E tests..."
	@LLM_INFO_BIN_PATH=$(PWD)/test/bin/llm-info go test ./test/e2e -v
	@echo "E2E tests completed"

# E2Eテストのクリーンアップ
test-e2e-clean:
	@echo "Cleaning E2E test environment..."
	@rm -rf test/bin/
	@echo "E2E test environment cleaned"

# すべてのテストを含むE2Eテスト
test-e2e-all: test test-e2e
	@echo "All tests (unit + integration + e2e) completed"

# ヘルプ表示
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linter"
	@echo "  install        - Install the binary to GOBIN or GOPATH/bin"
	@echo "  uninstall      - Remove the installed binary"
	@echo "  run-example    - Run with example gateway"
	@echo "  test-e2e-setup - Setup E2E test environment"
	@echo "  test-e2e       - Run E2E tests"
	@echo "  test-e2e-clean - Clean E2E test environment"
	@echo "  test-e2e-all   - Run all tests including E2E"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Installation destination:"
	@if [ -n "$$(go env GOBIN)" ]; then \
		echo "  GOBIN is set: $$GOBIN"; \
	else \
		echo "  GOBIN not set, using: $$(go env GOPATH)/bin"; \
	fi