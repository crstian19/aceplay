# Makefile for Aceplay

# Variables
BINARY_NAME := aceplay
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Directories
BUILD_DIR := build
DIST_DIR := dist
CMD_DIR := cmd

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64

.PHONY: all build clean test coverage install uninstall deps tidy fmt vet lint release help

# Default target
all: test build

## build: Compile binary for current platform
build:
	@echo "🔨 Compiling $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✅ Binary created: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Compile for all platforms
build-all: clean
	@echo "🔨 Compiling for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		OUTPUT=$(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		echo "  → $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GOBUILD) $(LDFLAGS) -o $$OUTPUT ./$(CMD_DIR); \
	done
	@echo "✅ Binaries created in $(DIST_DIR)/"

## clean: Clean generated files
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	$(GOCLEAN)
	@echo "✅ Clean completed"

## test: Run tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v -race ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## test-short: Run short tests
test-short:
	@echo "🧪 Running short tests..."
	$(GOTEST) -v -short ./...

## deps: Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## tidy: Clean and organize dependencies
tidy:
	@echo "🧹 Organizing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

## fmt: Format code
fmt:
	@echo "📝 Formatting code..."
	$(GOCMD) fmt ./...

## vet: Analyze code with go vet
vet:
	@echo "🔍 Analyzing code..."
	$(GOCMD) vet ./...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

## install: Install binary in the system
install: build
	@echo "📥 Installing $(BINARY_NAME)..."
	@install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)/usr/bin/$(BINARY_NAME)
	@install -Dm644 aceplay.desktop $(DESTDIR)/usr/share/applications/aceplay.desktop 2>/dev/null || true
	@echo "✅ Installation completed"

## uninstall: Uninstall binary from the system
uninstall:
	@echo "🗑️  Uninstalling $(BINARY_NAME)..."
	@rm -f $(DESTDIR)/usr/bin/$(BINARY_NAME)
	@rm -f $(DESTDIR)/usr/share/applications/aceplay.desktop
	@echo "✅ Uninstallation completed"

## release: Create GitHub release (requires gh CLI)
release: build-all
	@echo "🚀 Creating release $(VERSION)..."
	@if command -v gh >/dev/null 2>&1; then \
		gh release create $(VERSION) $(DIST_DIR)/* --title "$(BINARY_NAME) $(VERSION)" --notes "Release $(VERSION)"; \
	else \
		echo "⚠️  GitHub CLI (gh) not installed"; \
	fi

## run: Run program (for development)
run:
	@echo "▶️  Running $(BINARY_NAME)..."
	$(GOCMD) run $(LDFLAGS) ./$(CMD_DIR)

## run-live: Run with hot-reload (requires air)
run-live:
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
	fi

## check: Run all checks
check: fmt vet lint test
	@echo "✅ All checks passed"

## ci: Prepare for CI
ci: deps fmt vet test

## help: Show this help
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

# Rule to detect changes in Go files and recompile automatically
watch:
	@find . -name "*.go" | entr -r make build

# Local development installation
.PHONY: dev-install
dev-install:
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) ./$(CMD_DIR)
