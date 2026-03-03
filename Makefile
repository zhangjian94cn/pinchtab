.PHONY: setup install-hooks test lint fmt check build clean help

## Setup (run once after clone)
setup: install-hooks
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  make test    # Run tests"
	@echo "  make build   # Build pinchtab"
	@echo "  make fmt     # Format code"

## Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@./scripts/install-hooks.sh

## Run tests
test:
	go test ./...

## Run tests with coverage
test-cover:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not installed"; \
		echo "Install: brew install golangci-lint"; \
		echo "Or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## Format code
fmt:
	gofmt -w .
	@echo "✅ Code formatted"

## Check code (fmt + lint + test)
check: fmt lint test
	@echo "✅ All checks passed"

## Build pinchtab binary
build:
	go build -o pinchtab ./cmd/pinchtab

## Build for release
build-release:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o pinchtab ./cmd/pinchtab

## Clean build artifacts
clean:
	go clean
	rm -f pinchtab coverage.out coverage.html

## Run pinchtab
run: build
	./pinchtab

## Show help
help:
	@echo "Pinchtab Development Commands"
	@echo ""
	@echo "Setup:"
	@echo "  make setup          Setup development environment (run once)"
	@echo "  make install-hooks  Install git hooks"
	@echo ""
	@echo "Development:"
	@echo "  make fmt            Format code with gofmt"
	@echo "  make lint           Run golangci-lint"
	@echo "  make test           Run tests"
	@echo "  make test-cover     Run tests with coverage report"
	@echo "  make check          Run fmt + lint + test"
	@echo ""
	@echo "Build:"
	@echo "  make build          Build pinchtab binary"
	@echo "  make build-release  Build optimized release binary"
	@echo "  make run            Build and run pinchtab"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          Remove build artifacts"
