# Makefile for srv Go library

# Variables
GO_VERSION := 1.22.0
MODULE_NAME := github.com/cfichtmueller/srv
BUILD_DIR := build

# Go commands
GO := go
GOFMT := gofmt
GOLINT := golint
GOVET := go vet
GOTEST := go test
GOCOVER := go test -cover

# Default target
.PHONY: all
all: clean fmt vet lint test build

# Build the library
.PHONY: build
build:
	@echo "Building library..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -v ./...

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...


# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run golint (requires golint to be installed)
.PHONY: lint
lint:
	@echo "Running golint..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) ./...; \
	else \
		echo "golint not installed. Install with: go install golang.org/x/lint/golint@latest"; \
	fi

# Run go mod tidy
.PHONY: mod-tidy
mod-tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

# Build examples
.PHONY: build-examples
build-examples:
	@echo "Building examples..."
	@for example in examples/*/; do \
		if [ -f "$$example/main.go" ]; then \
			echo "Building $$example"; \
			$(GO) build -o $(BUILD_DIR)/$$(basename $$example) $$example/main.go; \
		fi; \
	done

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

# Check if Go version matches requirement
.PHONY: check-go-version
check-go-version:
	@echo "Checking Go version..."
	@if ! $(GO) version | grep -q "go$(GO_VERSION)"; then \
		echo "Warning: Go version $(GO_VERSION) is required"; \
		echo "Current version: $$($(GO) version)"; \
	fi


# Full CI pipeline
.PHONY: ci
ci: check-go-version fmt vet lint test build build-examples

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all              - Run clean, fmt, vet, lint, test, and build"
	@echo "  build            - Build the library"
	@echo "  test             - Run tests"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  lint             - Run golint"
	@echo "  mod-tidy         - Run go mod tidy"
	@echo "  build-examples   - Build all examples"
	@echo "  clean            - Clean build artifacts"
	@echo "  check-go-version - Check Go version"
	@echo "  ci               - Run full CI pipeline"
	@echo "  help             - Show this help message"
