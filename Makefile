# Makefile for PineScript Go Port
# Centralized build automation following Go project conventions

.PHONY: help build test clean install lint fmt vet bench coverage parser codegen integration e2e docker run dev release cross-compile

# Project configuration
PROJECT_NAME := pinescript-go
BINARY_NAME := pinescript-runner
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Directories
GOLANG_PORT := golang-port
CMD_DIR := $(GOLANG_PORT)/cmd/pinescript-go
BUILD_DIR := build
DIST_DIR := dist
COVERAGE_DIR := coverage

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

install: ## Install development dependencies
	@echo "Installing development tools..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✓ Development tools installed"

fmt: ## Format Go code
	@echo "Formatting code..."
	@cd $(GOLANG_PORT) && gofmt -s -w .
	@cd $(GOLANG_PORT) && goimports -w .
	@echo "✓ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@cd $(GOLANG_PORT) && $(GO) vet ./...
	@echo "✓ Vet passed"

lint: ## Run linter
	@echo "Running linter..."
	@cd $(GOLANG_PORT) && golangci-lint run --timeout 5m
	@echo "✓ Lint passed"

security: ## Run security scanner
	@echo "Running security scan..."
	@cd $(GOLANG_PORT) && gosec -quiet ./...
	@echo "✓ Security scan passed"

##@ Build

build: ## Build binary for current platform
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@cd $(GOLANG_PORT) && $(GOBUILD) -o ../$(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-strategy: ## Build standalone strategy binary (usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name)
	@if [ -z "$(STRATEGY)" ]; then echo "Error: STRATEGY not set. Usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name"; exit 1; fi
	@if [ -z "$(OUTPUT)" ]; then echo "Error: OUTPUT not set. Usage: make build-strategy STRATEGY=path/to/strategy.pine OUTPUT=runner-name"; exit 1; fi
	@echo "Building strategy: $(STRATEGY) -> $(OUTPUT)"
	@$(MAKE) -s _build_strategy_internal STRATEGY=$(STRATEGY) OUTPUT=$(OUTPUT)

_build_strategy_internal:
	@mkdir -p $(BUILD_DIR)
	@echo "[1/4] Parsing Pine Script..."
	@cd $(GOLANG_PORT) && $(GO) run $(CMD_DIR) ../$(STRATEGY) > /tmp/ast_output.json
	@echo "[2/4] Generating Go code..."
	@cd $(GOLANG_PORT) && $(GO) run -tags=build_strategy ./internal/build_strategy /tmp/ast_output.json $(BUILD_DIR)/$(OUTPUT)
	@echo "[3/4] Compiling binary..."
	@cd $(BUILD_DIR) && $(GOBUILD) -o $(OUTPUT) .
	@echo "[4/4] Cleanup..."
	@rm -f /tmp/ast_output.json
	@echo "✓ Strategy compiled: $(BUILD_DIR)/$(OUTPUT)"

cross-compile: ## Build for all platforms
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		$(MAKE) -s _cross_compile_platform \
		PLATFORM_OS=$(word 1,$(subst /, ,$(platform))) \
		PLATFORM_ARCH=$(word 2,$(subst /, ,$(platform))) ; \
	)
	@echo "✓ Cross-compilation complete"
	@ls -lh $(DIST_DIR)/

_cross_compile_platform:
	@BINARY=$(DIST_DIR)/$(BINARY_NAME)-$(PLATFORM_OS)-$(PLATFORM_ARCH)$(if $(findstring windows,$(PLATFORM_OS)),.exe,); \
	echo "Building $$BINARY..."; \
	cd $(GOLANG_PORT) && GOOS=$(PLATFORM_OS) GOARCH=$(PLATFORM_ARCH) \
	$(GOBUILD) -o ../$$BINARY $(CMD_DIR)

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

ci: check bench ## Run CI pipeline (all checks + benchmarks)
	@echo "✓ CI pipeline completed"

verify-series: ## Verify Series implementation correctness
	@echo "Verifying Series implementation..."
	@cd $(GOLANG_PORT) && $(GOTEST) -v -run TestSeries ./codegen/...
	@cd $(GOLANG_PORT) && $(GOTEST) -v ./runtime/series/...
	@echo "✓ Series verification passed"

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

##@ Docker

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(PROJECT_NAME):$(VERSION) -t $(PROJECT_NAME):latest .
	@echo "✓ Docker image built: $(PROJECT_NAME):$(VERSION)"

docker-run: docker-build ## Build and run in Docker
	@echo "Running in Docker..."
	@docker run --rm -it $(PROJECT_NAME):latest

docker-test: ## Run tests in Docker
	@echo "Running tests in Docker..."
	@docker run --rm $(PROJECT_NAME):latest make test

##@ Release

release: clean check cross-compile ## Build release binaries for all platforms
	@echo "Creating release artifacts..."
	@mkdir -p $(DIST_DIR)/release
	@cd $(DIST_DIR) && for f in $(BINARY_NAME)-*; do \
		if [ -f "$$f" ]; then \
			tar czf release/$${f}.tar.gz $$f; \
			echo "Created release/$${f}.tar.gz"; \
		fi \
	done
	@cd $(DIST_DIR)/release && shasum -a 256 *.tar.gz > checksums.txt
	@echo "✓ Release artifacts created in $(DIST_DIR)/release/"
	@echo ""
	@echo "Release $(VERSION) ready for distribution"
	@cat $(DIST_DIR)/release/checksums.txt

tag: ## Create git tag (usage: make tag VERSION=v1.0.0)
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION not set. Usage: make tag VERSION=v1.0.0"; exit 1; fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "✓ Tag $(VERSION) created and pushed"

##@ Development Workflow

dev: ## Development mode with auto-rebuild
	@echo "Starting development mode..."
	@cd $(GOLANG_PORT) && $(GO) run $(CMD_DIR) $(ARGS)

run: build ## Build and run binary
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

run-strategy: ## Run strategy and generate chart-data.json (usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json)
	@if [ -z "$(STRATEGY)" ]; then echo "Error: STRATEGY not set. Usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json"; exit 1; fi
	@if [ -z "$(DATA)" ]; then echo "Error: DATA not set. Usage: make run-strategy STRATEGY=path/to/strategy.pine DATA=path/to/data.json"; exit 1; fi
	@echo "Running strategy: $(STRATEGY)"
	@mkdir -p out
	@TEMP_FILE=$$(cd $(GOLANG_PORT) && $(GO) run cmd/pinescript-builder/main.go \
		-input ../$(STRATEGY) \
		-output /tmp/pinescript-strategy 2>&1 | grep "Generated:" | awk '{print $$2}'); \
	cd $(GOLANG_PORT) && $(GO) build -o /tmp/pinescript-strategy $$TEMP_FILE
	@/tmp/pinescript-strategy -symbol TEST -data $(DATA) -output out/chart-data.json
	@echo "✓ Strategy executed: out/chart-data.json"
	@ls -lh out/chart-data.json

serve: ## Serve ./out directory with Python HTTP server on port 8000
	@echo "Starting web server on http://localhost:8000"
	@echo "Chart data available at: http://localhost:8000/chart-data.json"
	@echo "Press Ctrl+C to stop server"
	@cd out && python3 -m http.server 8000

test-manual: run-strategy serve ## Run strategy and start web server for manual testing

watch: ## Watch for changes and run tests (requires entr)
	@echo "Watching for changes..."
	@find $(GOLANG_PORT) -name "*.go" | entr -c make test

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

install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@echo '#!/bin/sh\nmake fmt vet' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✓ Git hooks installed"
