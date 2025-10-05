# PineTS Multi-Provider Trading Analysis

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

# Generate analysis and start visualization server
pnpm run visualize
```

Visit: http://localhost:8080/chart.html

## Configuration Parameters

You can customize the analysis using environment variables:

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

### Core Commands

- `pnpm start` - Generate trading analysis data with current parameters
- `pnpm run serve` - Start HTTP server for chart visualization
- `pnpm run visualize` - Build data + start server (recommended)

### Development

- `pnpm run dev` - Run analysis with file watching
- `pnpm run dev:watch` - Run analysis + server with file watching
- `pnpm run build` - Generate chart data and configuration
- `pnpm run stop` - Stop HTTP server

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
