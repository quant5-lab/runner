# Makefile for Runner - PineScript Go Port
# Centralized build automation following Go project conventions

.PHONY: help build test clean fmt vet bench coverage integration e2e cross-compile

# Project configuration
PROJECT_NAME := runner
BINARY_NAME := pine-gen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
GOLANG_PORT := golang-port
CMD_DIR := $(GOLANG_PORT)/cmd/pine-gen
BUILD_DIR := $(GOLANG_PORT)/build
DIST_DIR := $(GOLANG_PORT)/dist
COVERAGE_DIR := $(GOLANG_PORT)/coverage

# Go configuration
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.CommitHash=$(COMMIT_HASH)"
GOTEST := $(GO) test $(GOFLAGS)
GOBUILD := $(GO) build $(GOFLAGS) $(LDFLAGS)

# Test configuration
TEST_TIMEOUT := 30m
TEST_FLAGS := -race -timeout $(TEST_TIMEOUT)
BENCH_FLAGS := -benchmem -benchtime=3s

# Cross-compilation targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

fmt: ## Format Go code
	@echo "Formatting code..."
	@cd $(GOLANG_PORT) && gofmt -s -w .
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@cd $(GOLANG_PORT) && $(GO) vet ./...
	@echo "✓ Vet passed"

lint: ## Run linter
	@echo "Running linter..."
	@cd $(GOLANG_PORT) && $(GO) vet ./...
	@echo "✓ Lint passed"

##@ Build

build: ## Build pine-gen for current platform
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@cd $(GOLANG_PORT) && $(GOBUILD) -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/pine-gen
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-strategy: ## Build standalone strategy binary (usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name)
	@if [ -z "$(STRATEGY)" ]; then echo "Error: STRATEGY not set. Usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name"; exit 1; fi
	@if [ -z "$(OUTPUT)" ]; then echo "Error: OUTPUT not set. Usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name"; exit 1; fi
	@echo "Building strategy: $(STRATEGY) -> $(OUTPUT)"
	@$(MAKE) -s _build_strategy_internal STRATEGY=$(STRATEGY) OUTPUT=$(OUTPUT)

_build_strategy_internal:
	@mkdir -p $(BUILD_DIR)
	@echo "[1/3] Generating Go code from Pine Script..."
	@TEMP_FILE=$$(cd $(GOLANG_PORT) && $(GO) run ./cmd/pine-gen -input ../$(STRATEGY) -output $(BUILD_DIR)/$(OUTPUT) 2>&1 | grep "Generated:" | awk '{print $$2}'); \
	if [ -z "$$TEMP_FILE" ]; then echo "Failed to generate Go code"; exit 1; fi; \
	echo "[2/3] Compiling binary..."; \
	cd $(GOLANG_PORT) && $(GO) build -o ../$(BUILD_DIR)/$(OUTPUT) $$TEMP_FILE
	@echo "[3/3] Cleanup..."
	@echo "✓ Strategy compiled: $(BUILD_DIR)/$(OUTPUT)"

cross-compile: ## Build pine-gen for all platforms (strategy code generator)
	@echo "Cross-compiling pine-gen for distribution..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		$(MAKE) -s _cross_compile_platform \
		PLATFORM_OS=$(word 1,$(subst /, ,$(platform))) \
		PLATFORM_ARCH=$(word 2,$(subst /, ,$(platform))) ; \
	)
	@echo "✓ Cross-compilation complete: $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

_cross_compile_platform:
	@BINARY=$(DIST_DIR)/pine-gen-$(PLATFORM_OS)-$(PLATFORM_ARCH)$(if $(findstring windows,$(PLATFORM_OS)),.exe,); \
	echo "  Building $$BINARY..."; \
	cd $(GOLANG_PORT) && GOOS=$(PLATFORM_OS) GOARCH=$(PLATFORM_ARCH) \
	$(GOBUILD) -o ../$$BINARY ./cmd/pine-gen

##@ Testing

test: ## Run all tests
	@echo "Running tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) ./...
	@echo "✓ All tests passed"

test-parser: ## Run parser tests only
	@echo "Running parser tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) ./parser/...
	@echo "✓ Parser tests passed"

test-codegen: ## Run codegen tests only
	@echo "Running codegen tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) ./codegen/...
	@echo "✓ Codegen tests passed"

test-runtime: ## Run runtime tests only
	@echo "Running runtime tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) ./runtime/...
	@echo "✓ Runtime tests passed"

test-series: ## Run Series tests only
	@echo "Running Series tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) -v ./runtime/series/...
	@echo "✓ Series tests passed"

integration: ## Run integration tests
	@echo "Running integration tests..."
	@cd $(GOLANG_PORT) && $(GOTEST) $(TEST_FLAGS) -tags=integration ./tests/integration/...
	@echo "✓ Integration tests passed"

e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@cd e2e && ./run-all.sh
	@echo "✓ E2E tests passed"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@cd $(GOLANG_PORT) && $(GO) test $(BENCH_FLAGS) -bench=. ./...

bench-series: ## Benchmark Series performance
	@echo "Benchmarking Series..."
	@cd $(GOLANG_PORT) && $(GO) test $(BENCH_FLAGS) -bench=. ./runtime/series/
	@echo ""
	@echo "Performance targets:"
	@echo "  Series.Get():    < 10ns/op"
	@echo "  Series.Set():    < 5ns/op"
	@echo "  Series.Next():   < 3ns/op"

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@mkdir -p $(COVERAGE_DIR)
	@cd $(GOLANG_PORT) && $(GO) test -coverprofile=../$(COVERAGE_DIR)/coverage.out ./...
	@cd $(GOLANG_PORT) && $(GO) tool cover -html=../$(COVERAGE_DIR)/coverage.out -o ../$(COVERAGE_DIR)/coverage.html
	@cd $(GOLANG_PORT) && $(GO) tool cover -func=../$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "✓ Coverage report: $(COVERAGE_DIR)/coverage.html"

coverage-show: coverage ## Generate and open coverage report
	@open $(COVERAGE_DIR)/coverage.html

##@ Verification

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
	@echo "✓ All checks passed"

##@ Cleanup

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@cd $(GOLANG_PORT) && $(GO) clean -cache -testcache
	@find . -name "*.test" -type f -delete
	@find . -name "*.out" -type f -delete
	@echo "✓ Cleaned"

clean-all: clean ## Remove all generated files including dependencies
	@echo "Removing all generated files..."
	@cd $(GOLANG_PORT) && $(GO) clean -modcache
	@echo "✓ Deep cleaned"

##@ Development Workflow

run-strategy: ## Run strategy with pre-generated data file (usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json)
	@if [ -z "$(STRATEGY)" ]; then echo "Error: STRATEGY not set. Usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json"; exit 1; fi
	@if [ -z "$(DATA)" ]; then echo "Error: DATA not set. Usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json"; exit 1; fi
	@echo "Running strategy: $(STRATEGY)"
	@mkdir -p out
	@TEMP_FILE=$$(cd $(GOLANG_PORT) && $(GO) run cmd/pine-gen/main.go \
		-input ../$(STRATEGY) \
		-output /tmp/pinescript-strategy 2>&1 | grep "Generated:" | awk '{print $$2}'); \
	cd $(GOLANG_PORT) && $(GO) build -o /tmp/pinescript-strategy $$TEMP_FILE
	@SYMBOL=$$(basename $(DATA) | sed 's/_[^_]*\.json//'); \
	TIMEFRAME=$$(basename $(DATA) .json | sed 's/.*_//'); \
	/tmp/pinescript-strategy -symbol $$SYMBOL -timeframe $$TIMEFRAME -data $(DATA) -datadir golang-port/testdata/ohlcv -output out/chart-data.json
	@echo "✓ Strategy executed: out/chart-data.json"
	@ls -lh out/chart-data.json

fetch-strategy: ## Fetch live data and run strategy (usage: make fetch-strategy SYMBOL=GDYN TIMEFRAME=1D BARS=500 STRATEGY=strategies/daily-lines.pine)
	@if [ -z "$(SYMBOL)" ] || [ -z "$(STRATEGY)" ]; then \
		echo "Usage: make fetch-strategy SYMBOL=<symbol> TIMEFRAME=<tf> BARS=<n> STRATEGY=<file>"; \
		echo ""; \
		echo "Examples:"; \
		echo "  make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=1h BARS=500 STRATEGY=strategies/daily-lines.pine"; \
		echo "  make fetch-strategy SYMBOL=AAPL TIMEFRAME=1D BARS=200 STRATEGY=strategies/test-simple.pine"; \
		echo ""; \
		exit 1; \
	fi
	@./scripts/fetch-strategy.sh $(SYMBOL) $(TIMEFRAME) $(BARS) $(STRATEGY)

serve: ## Serve ./out directory with Python HTTP server on port 8000
	@echo "Starting web server on http://localhost:8000"
	@echo "Chart data available at: http://localhost:8000/chart-data.json"
	@echo "Press Ctrl+C to stop server"
	@cd out && python3 -m http.server 8000

serve-strategy: fetch-strategy serve ## Fetch live data, run strategy, and start web server

##@ Information

version: ## Show version information
	@echo "Version:     $(VERSION)"
	@echo "Build Time:  $(BUILD_TIME)"
	@echo "Commit:      $(COMMIT_HASH)"
	@echo "Go Version:  $(shell $(GO) version)"

deps: ## Show dependencies
	@echo "Project dependencies:"
	@cd $(GOLANG_PORT) && $(GO) list -m all

mod-tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@cd $(GOLANG_PORT) && $(GO) mod tidy
	@cd $(GOLANG_PORT) && $(GO) mod verify
	@echo "✓ Dependencies tidied"

mod-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@cd $(GOLANG_PORT) && $(GO) get -u ./...
	@$(MAKE) mod-tidy
	@echo "✓ Dependencies updated"

##@ Quick Commands

all: clean build test ## Clean, build, and test everything

quick: fmt test ## Quick check (format + test)

install-hooks: ## Install/update git pre-commit hook
	@echo "Installing pre-commit hook..."
	@printf '#!/bin/sh\n\nset -e\n\necho "🔍 Running pre-commit checks..."\n\n' > .git/hooks/pre-commit
	@printf '# SourceTree compatibility: Find go binary in common locations\n' >> .git/hooks/pre-commit
	@printf 'if ! command -v go >/dev/null 2>&1; then\n' >> .git/hooks/pre-commit
	@printf '    if [ -x "/usr/local/go/bin/go" ]; then\n' >> .git/hooks/pre-commit
	@printf '        export PATH="/usr/local/go/bin:$$PATH"\n' >> .git/hooks/pre-commit
	@printf '    elif [ -x "$$HOME/go/bin/go" ]; then\n' >> .git/hooks/pre-commit
	@printf '        export PATH="$$HOME/go/bin:$$PATH"\n' >> .git/hooks/pre-commit
	@printf '    elif [ -x "/opt/homebrew/bin/go" ]; then\n' >> .git/hooks/pre-commit
	@printf '        export PATH="/opt/homebrew/bin:$$PATH"\n' >> .git/hooks/pre-commit
	@printf '    else\n' >> .git/hooks/pre-commit
	@printf '        echo "Error: go binary not found. Please install Go or add it to PATH."\n' >> .git/hooks/pre-commit
	@printf '        exit 1\n' >> .git/hooks/pre-commit
	@printf '    fi\nfi\n\n' >> .git/hooks/pre-commit
	@printf '# Format\necho "  [1/3] Formatting Go code..."\n' >> .git/hooks/pre-commit
	@printf 'cd golang-port && gofmt -s -w . && cd .. || exit 1\n\n' >> .git/hooks/pre-commit
	@printf '# Lint\necho "  [2/3] Running linter..."\n' >> .git/hooks/pre-commit
	@printf 'cd golang-port && go vet ./... && cd .. || exit 1\n\n' >> .git/hooks/pre-commit
	@printf '# Test\necho "  [3/3] Running tests..."\n' >> .git/hooks/pre-commit
	@printf 'cd golang-port && go test ./... -timeout 30m || exit 1\n\n' >> .git/hooks/pre-commit
	@printf 'echo "✅ Pre-commit checks passed!"\nexit 0\n' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✓ Pre-commit hook installed"

