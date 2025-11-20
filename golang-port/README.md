# PineScript Go Port

High-performance PineScript v5 parser, transpiler, and runtime written in Go.

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

## Architecture

- **Parser**: Custom PineScript v5 grammar using `participle/v2`
- **Transpiler**: AST → Go source code generation
- **Runtime**: Pure Go TA library (SMA/EMA/RSI/ATR/BBands/MACD/Stoch)
- **Execution**: <50ms for 500 bars (50x faster than Python)

## Next Steps

See project root README for complete documentation and Node.js integration.
