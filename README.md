# PineTS Multi-Provider Trading Analysis

![Coverage](https://img.shields.io/badge/coverage-94.7%25-brightgreen)

A unified Node.js application for technical analysis across multiple exchanges using PineTS with dynamic provider fallback and real-time chart visualization.

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

# Generate analysis and start visualization server
pnpm start
```

Visit: http://localhost:8080/chart.html

## Configuration Parameters

### Local Development

```bash
# Symbol configuration
SYMBOL=AAPL pnpm start          # Apple stock (Yahoo Finance)
SYMBOL=BTCUSDT pnpm start       # Bitcoin (Binance)
SYMBOL=SBER pnpm start          # Sberbank (MOEX)

# Historical data length (number of candlesticks)
BARS=50 pnpm start              # Get 50 candles
BARS=200 pnpm start             # Get 200 candles (default: 100)

# Timeframe configuration
TIMEFRAME=1h pnpm start         # 1-hour candles
TIMEFRAME=D pnpm start          # Daily candles (default)

# Combined configuration
SYMBOL=AAPL BARS=150 TIMEFRAME=D pnpm start
```

### Docker Usage

```bash
# Start/restart runner with env vars
SYMBOL=BTCUSDT TIMEFRAME=D pnpm docker:start

# Run tests in Docker
pnpm docker:test

# Access running container
docker-compose exec app pnpm test
docker-compose exec app pnpm start
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

- `pnpm test` - Run tests with automatic network monitoring (if tcpdump available)
- `pnpm test:ui` - Run tests with interactive UI
- `pnpm start` - Run tests (prestart), then generate analysis and start HTTP server
- `pnpm coverage` - Generate test coverage report
- `pnpm lint` - Lint code
- `pnpm format` - Format and fix code

### Docker Commands

- `pnpm docker:test` - Run tests in Docker with network isolation
- `pnpm docker:start` - Start runner in Docker (pass env vars: `SYMBOL=BTCUSDT pnpm docker:start`)

## Dynamic Provider Fallback

The system automatically tries providers in order:

1. **MOEX** (for Russian stocks)
2. **Binance** (for crypto pairs)
3. **Yahoo Finance** (for US stocks)

If a provider doesn't have data for the requested symbol, it automatically falls back to the next provider.

## Technical Analysis

The application implements an EMA Crossover Strategy with:

- **EMA9** - 9-period Exponential Moving Average (blue line)
- **EMA18** - 18-period Exponential Moving Average (red line)
- **BullSignal** - Bullish signal when EMA9 > EMA18 (green line)

## Architecture

- **index.js** - Multi-provider system with SOLID architecture
- **providers/** - Custom provider implementations (MOEX, Yahoo Finance)
- **chart.html** - Data-agnostic TradingView-style visualization
- **chart-config.json** - Auto-generated UI configuration (gitignored)
- **chart-data.json** - Auto-generated market data (gitignored)

## Dependencies

- **Node.js 18+** - Runtime environment
- **PineTS** - Technical analysis engine with Binance provider
- **Custom Providers** - MOEX and Yahoo Finance integrations
