# Pine Script Trading Analysis Runner

![Coverage](https://img.shields.io/badge/coverage-86.6%25-brightgreen)

Node.js application for Pine Script strategy transpilation and execution across multiple exchanges with dynamic provider fallback and real-time chart visualization.e Script Trading Analysis Runner

![Coverage](https://img.shields.io/badge/coverage-80.8%25-brightgreen)

Node.js application for Pine Script strategy transpilation and execution across multiple exchanges with dynamic provider fallback and real-time chart visualization.

## Supported Exchanges

- **MOEX** - Russian stock exchange (free API)
- **Binance** - Cryptocurrency exchange (native PineTS provider)
- **Yahoo Finance** - US stocks NYSE/NASDAQ/AMEX (free API)

## Quick Start

```bash
# Install pnpm globally (if not already installed)
npm install -g pnpm

# Install dependencies
pnpm install

# Run tests before starting
pnpm test

# Run E2E tests
docker compose run --rm runner sh e2e/run-all.sh

# Run Pine Script strategy analysis
pnpm start AAPL 1h 100 strategies/test.pine

# Run without Pine Script (default EMA strategy)
pnpm start
```

Visit: http://localhost:8080/chart.html

## Configuration Parameters

### Local Development

```bash
# Default EMA strategy
pnpm start                                    # AAPL, Daily, 100 bars

# With Pine Script strategy
pnpm start AAPL 1h 100 strategies/test.pine  # Symbol, Timeframe, Bars, Strategy

# Symbol configuration
pnpm start BTCUSDT                            # Bitcoin (Binance)
pnpm start SBER                               # Sberbank (MOEX)

# Historical data length (number of candlesticks)
pnpm start AAPL 1h 50                         # 50 candles
pnpm start AAPL D 200                         # 200 candles

# Timeframe configuration
pnpm start AAPL 1h                            # 1-hour candles
pnpm start AAPL D                             # Daily candles
```

### Docker Usage

```bash
# Start runner container
docker-compose up -d

# Run Pine Script strategy
docker-compose exec runner pnpm start AAPL 1h 100 strategies/bb-strategy-7-rus.pine

# Run tests
docker-compose exec runner pnpm test

# Format code
docker-compose exec runner pnpm format

# Show transpiled JavaScript code
docker-compose exec -e DEBUG=true runner pnpm start AAPL 1h 100 strategies/test.pine

# Access running container shell
docker-compose exec runner sh
```

### Supported Symbols by Provider

**MOEX (Russian Stocks):**

- `SBER` - Sberbank
- `GAZP` - Gazprom
- `LKOH` - Lukoil
- `YNDX` - Yandex

**Binance (Cryptocurrency):**

- `BTCUSDT` - Bitcoin
- `ETHUSDT` - Ethereum
- `ADAUSDT` - Cardano
- `SOLUSDT` - Solana

**Yahoo Finance (US Stocks):**

- `AAPL` - Apple
- `GOOGL` - Google
- `MSFT` - Microsoft
- `TSLA` - Tesla

## Available Scripts

### Local Development

- `pnpm test` - Run tests with automatic network monitoring
- `pnpm test:ui` - Run tests with interactive UI
- `pnpm start [SYMBOL] [TIMEFRAME] [BARS] [STRATEGY]` - Run strategy analysis
- `pnpm coverage` - Generate test coverage report
- `pnpm lint` - Lint code
- `pnpm format` - Format and fix code

### Docker Commands

- `docker-compose up -d` - Start runner container
- `docker-compose exec runner pnpm test` - Run tests in Docker
- `docker-compose exec runner pnpm start` - Run analysis in Docker
- `docker-compose exec runner pnpm format` - Format code in Docker

## Dynamic Provider Fallback

The system automatically tries providers in order:

1. **MOEX** (for Russian stocks)
2. **Binance** (for crypto pairs)
3. **Yahoo Finance** (for US stocks)

If a provider doesn't have data for the requested symbol, it automatically falls back to the next provider.

## Pine Script Strategy Execution

The application transpiles and executes Pine Script strategies:

1. **Transpilation**: Pine Script → JavaScript via pynescript + custom AST visitor
2. **Execution**: JavaScript runs in sandboxed context with market data
3. **Visualization**: Results rendered in TradingView-style chart

### Example Strategies

- `strategies/test.pine` - Simple indicator test
- `strategies/bb-strategy-7-rus.pine` - BB + ADX strategy (320 lines)
- `strategies/bb-strategy-8-rus.pine` - Pyramiding strategy (295 lines)
- `strategies/bb-strategy-9-rus.pine` - Partial close strategy (316 lines)

### Transpilation Architecture

- **Parser Service**: Python service using pynescript library
- **AST Converter**: Custom visitor pattern (PyneToJsAstConverter)
- **Code Generation**: ESTree JavaScript AST → escodegen
- **Execution**: Function constructor with Pine Script API stubs

## Technical Analysis

### Default EMA Strategy

- **EMA9** - 9-period Exponential Moving Average (blue line)
- **EMA18** - 18-period Exponential Moving Average (red line)
- **BullSignal** - Bullish signal when EMA9 > EMA18 (green line)

### Pine Script Strategies

Custom strategies with full Pine Script v3/v4/v5 support (auto-migration) including:

- Built-in functions: `indicator()`, `strategy()`, `plot()`
- Built-in variables: `close`, `open`, `high`, `low`, `volume`
- Technical indicators: Bollinger Bands, ADX, SMA, RSI
- Array destructuring: `[ADX, up, down] = adx()`

## Architecture

- **index.js** - Entry point with Pine Script integration
- **src/classes/** - SOLID architecture with DI container
- **src/providers/** - Exchange integrations (MOEX, Yahoo Finance)
- **src/pine/** - Pine Script transpiler (Node.js → Python bridge)
- **services/pine-parser/** - Python parser service (pynescript + escodegen)
- **strategies/** - Pine Script strategy files (.pine)
- **out/** - Generated output files (chart-data.json, chart-config.json)
- **chart.html** - TradingView-style visualization

## Environment Variables

- `DEBUG=true` - Show verbose output

## Dependencies

- **Node.js 18+** - Runtime environment
- **Python 3.12+** - Pine Script parser service
- **pynescript 0.2.0** - Pine Script AST parser
- **escodegen** - JavaScript code generation
- **PineTS** - Default EMA strategy engine
- **Custom Providers** - MOEX and Yahoo Finance integrations
