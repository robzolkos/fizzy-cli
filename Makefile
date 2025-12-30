.PHONY: test test-unit test-e2e test-go test-file test-run build clean tidy help

BINARY := $(CURDIR)/bin/fizzy
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

# Test configuration (set these or export as environment variables)
# export FIZZY_TEST_TOKEN=your-token
# export FIZZY_TEST_ACCOUNT=your-account

help:
	@echo "Fizzy CLI"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the CLI"
	@echo "  make test-unit    Run unit tests (no API required)"
	@echo "  make test-e2e     Run e2e tests (requires API credentials)"
	@echo "  make test         Alias for test-e2e"
	@echo "  make test-file    Run a specific e2e test file"
	@echo "  make test-run     Run a specific e2e test by name"
	@echo "  make clean        Remove build artifacts"
	@echo "  make tidy         Tidy dependencies"
	@echo ""
	@echo "Environment variables (required for e2e tests):"
	@echo "  FIZZY_TEST_TOKEN   API token"
	@echo "  FIZZY_TEST_ACCOUNT Account slug"
	@echo "  FIZZY_TEST_API_URL API base URL (default: https://app.fizzy.do)"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test-unit"
	@echo "  export FIZZY_TEST_TOKEN=your-token"
	@echo "  export FIZZY_TEST_ACCOUNT=your-account"
	@echo "  make test-e2e"

# Build CLI
build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/fizzy

# Run unit tests (no API required)
test-unit:
	go test -v ./internal/...

# Run e2e tests (requires API credentials)
test-e2e: build
	@if [ -z "$$FIZZY_TEST_TOKEN" ]; then echo "Error: FIZZY_TEST_TOKEN not set"; exit 1; fi
	@if [ -z "$$FIZZY_TEST_ACCOUNT" ]; then echo "Error: FIZZY_TEST_ACCOUNT not set"; exit 1; fi
	FIZZY_TEST_BINARY=$(BINARY) go test -v ./e2e/tests/...

# Alias for test-e2e
test: test-e2e
test-go: test-e2e

# Run a single test file (e.g., make test-file FILE=board)
test-file: build
	@if [ -z "$(FILE)" ]; then echo "Usage: make test-file FILE=board"; exit 1; fi
	@if [ -z "$$FIZZY_TEST_TOKEN" ]; then echo "Error: FIZZY_TEST_TOKEN not set"; exit 1; fi
	@if [ -z "$$FIZZY_TEST_ACCOUNT" ]; then echo "Error: FIZZY_TEST_ACCOUNT not set"; exit 1; fi
	FIZZY_TEST_BINARY=$(BINARY) go test -v ./e2e/tests/$(FILE)_test.go

# Run a single test by name (e.g., make test-run NAME=TestBoardCRUD)
test-run: build
	@if [ -z "$(NAME)" ]; then echo "Usage: make test-run NAME=TestBoardCRUD"; exit 1; fi
	@if [ -z "$$FIZZY_TEST_TOKEN" ]; then echo "Error: FIZZY_TEST_TOKEN not set"; exit 1; fi
	@if [ -z "$$FIZZY_TEST_ACCOUNT" ]; then echo "Error: FIZZY_TEST_ACCOUNT not set"; exit 1; fi
	FIZZY_TEST_BINARY=$(BINARY) go test -v -run $(NAME) ./e2e/tests/...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Tidy dependencies
tidy:
	go mod tidy
