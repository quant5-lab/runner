# Runner - PineScript Go Port

High-performance PineScript v5 parser, transpiler, and runtime written in Go for Quant 5 Lab.

## Tooling

- **pine-inspect**: AST parser/debugger (outputs JSON AST for inspection)
- **pine-gen**: Code generator (transpiles .pine → Go source)
- **Strategy binaries**: Standalone executables (compiled per-strategy)

## Quick Start

### Testing Commands

```bash
# Fetch live data and run strategy
make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=1h BARS=500 STRATEGY=strategies/daily-lines.pine

# Fetch + run + start web server (combined workflow)
make serve-strategy SYMBOL=AAPL TIMEFRAME=1D BARS=200 STRATEGY=strategies/test-simple.pine

# Run with pre-generated data file (deterministic, CI-friendly)
make run-strategy STRATEGY=strategies/daily-lines.pine DATA=golang-port/testdata/ohlcv/BTCUSDT_1h.json
```

### Build Commands

```bash
# Build any .pine strategy to standalone binary
make build-strategy STRATEGY=strategies/your-strategy.pine OUTPUT=your-runner
```

## Command Reference

| Command | Purpose | Usage |
|---------|---------|-------|
| `fetch-strategy` | Fetch live data and run strategy | `SYMBOL=X TIMEFRAME=Y BARS=Z STRATEGY=file.pine` |
| `serve-strategy` | Fetch + run + serve results | `SYMBOL=X TIMEFRAME=Y BARS=Z STRATEGY=file.pine` |
| `run-strategy` | Run with pre-generated data file | `STRATEGY=file.pine DATA=data.json` |
| `build-strategy` | Build strategy to standalone binary | `STRATEGY=file.pine OUTPUT=binary-name` |

## Examples

### Testing with Live Data
```bash
# Crypto (Binance)
make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=1h BARS=500 STRATEGY=strategies/daily-lines.pine

# US Stocks (Yahoo Finance)
make fetch-strategy SYMBOL=GOOGL TIMEFRAME=1D BARS=250 STRATEGY=strategies/rolling-cagr.pine

# Russian Stocks (MOEX)
make fetch-strategy SYMBOL=SBER TIMEFRAME=1h BARS=500 STRATEGY=strategies/ema-strategy.pine
```

### Testing with Pre-generated Data
```bash
# Reproducible test (no network)
make run-strategy \
  STRATEGY=strategies/test-simple.pine \
  DATA=testdata/ohlcv/BTCUSDT_1h.json
```

### Building Standalone Binaries
```bash
# Build custom strategy
make build-strategy \
  STRATEGY=strategies/bb-strategy-7-rus.pine \
  OUTPUT=bb-runner

# Execute binary
./build/bb-runner -symbol BTCUSDT -data testdata/BTCUSDT_1h.json -output out/chart-data.json
```

## Makefile Command Examples for Manual Testing

### Basic Commands

```bash
# Display all available commands
make help

# Format code
make fmt

# Run static analysis
make vet

# Run all checks
make check
```

### Build Commands

```bash
# Build pine-gen for current platform
make build

# Build a specific strategy
make build-strategy STRATEGY=strategies/test-simple.pine OUTPUT=test-runner
make build-strategy STRATEGY=strategies/ema-strategy.pine OUTPUT=ema-runner
make build-strategy STRATEGY=strategies/bb-strategy-7-rus.pine OUTPUT=bb7-runner

# Cross-compile for all platforms
make cross-compile
```

### Testing Commands

```bash
# Run all tests
make test

# Run specific test suites
make test-parser      # Parser tests only
make test-codegen     # Code generation tests
make test-runtime     # Runtime tests
make test-series      # Series buffer tests
make integration      # Integration tests
make e2e             # End-to-end tests

# Run benchmarks
make bench
make bench-series

# Generate coverage report
make coverage
make coverage-show   # Opens in browser
```

### Development Workflow

```bash
# Run strategy with existing data
make run-strategy \
  STRATEGY=strategies/daily-lines.pine \
  DATA=golang-port/testdata/ohlcv/BTCUSDT_1h.json

# Fetch live data and run strategy
make fetch-strategy \
  SYMBOL=BTCUSDT \
  TIMEFRAME=1h \
  BARS=500 \
  STRATEGY=strategies/daily-lines.pine

# More examples:
make fetch-strategy SYMBOL=ETHUSDT TIMEFRAME=1D BARS=200 STRATEGY=strategies/ema-strategy.pine
make fetch-strategy SYMBOL=BTCUSDT TIMEFRAME=15m BARS=1000 STRATEGY=strategies/test-simple.pine

# Start web server to view results
make serve  # Opens http://localhost:8000

# Fetch, run, and serve in one command
make serve-strategy \
  SYMBOL=BTCUSDT \
  TIMEFRAME=1h \
  BARS=500 \
  STRATEGY=strategies/daily-lines.pine
```

### Maintenance Commands

```bash
# Clean build artifacts
make clean

# Deep clean (including Go cache)
make clean-all

# Update dependencies
make mod-tidy
make mod-update

# Install pre-commit hooks
make install-hooks
```

### Complete Testing Workflow

```bash
# 1. Format and verify
make fmt
make vet

# 2. Run all tests
make test

# 3. Run integration tests
make integration

# 4. Check benchmarks
make bench-series

# 5. Build a strategy and test it
make build-strategy STRATEGY=strategies/test-simple.pine OUTPUT=test-runner
./golang-port/build/test-runner \
  -symbol BTCUSDT \
  -timeframe 1h \
  -data golang-port/testdata/ohlcv/BTCUSDT_1h.json \
  -output out/test-result.json

# 6. View results
cat out/test-result.json | jq '.strategy.equity'

# 7. Test cross-compilation
make cross-compile

# 8. Generate coverage report
make coverage

# 9. Full verification
make check
```

### Advanced Testing

```bash
# Verbose test output
cd golang-port
go test -v ./tests/integration/

# Test specific function
cd golang-port
go test -v ./tests/integration -run TestSecurity

# Check for race conditions
cd golang-port
go test -race -count=10 ./...

# Benchmark specific package
cd golang-port
go test -bench=. -benchmem -benchtime=5s ./runtime/series/

# Memory profiling
cd golang-port
go test -memprofile=mem.prof -bench=. ./runtime/series/
go tool pprof mem.prof
```

### Quick Commands

```bash
# Build and test everything
make all

# Quick iteration
make quick  # fmt + test

# Complete verification
make check  # fmt + vet + lint + test
```
