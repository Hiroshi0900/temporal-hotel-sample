.PHONY: test lint clean build run

# テスト実行
test:
	go test -v ./...

# テスト実行（カバレッジ付き）
test-coverage:
	go test -v -cover ./...

# リント実行
lint:
	golangci-lint run

# 依存関係のダウンロード
deps:
	go mod download
	go mod tidy

# ビルド
build:
	go build -o bin/server cmd/server/main.go

# 実行
run:
	go run cmd/server/main.go

# クリーンアップ
clean:
	rm -rf bin/

# すべてのチェック（テスト + リント）
check: test lint

# 開発用の依存関係インストール
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# ヘルプ
help:
	@echo "Available targets:"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linter"
	@echo "  deps          - Download dependencies"
	@echo "  build         - Build application"
	@echo "  run           - Run application"
	@echo "  clean         - Clean build artifacts"
	@echo "  check         - Run tests and lint"
	@echo "  install-tools - Install development tools"
