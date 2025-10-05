# PineTS Trading Analysis & Visualization

A Node.js application for technical analysis using PineTS with real-time chart visualization.

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

## Available Scripts

### Core Commands
- `pnpm start` - Generate trading analysis data
- `pnpm run serve` - Start HTTP server for chart visualization  
- `pnpm run visualize` - Build data + start server (recommended)

### Development
- `pnpm run dev` - Run analysis with file watching
- `pnpm run dev:watch` - Run analysis + server with file watching
- `pnpm run build` - Generate chart data and configuration
- `pnpm run stop` - Stop HTTP server

## Configuration

Edit the `TRADING_CONFIG` in `index.js` to change:
- Symbol (BTCUSDT, ETHUSDT, etc.)
- Timeframe (1m, 5m, 1h, D, W, M)
- Strategy parameters
- Visual appearance

## Architecture

- **index.js** - Single source of truth for all configuration
- **chart.html** - Data-agnostic visualization interface  
- **chart-config.json** - Auto-generated UI configuration
- **chart-data.json** - Auto-generated market data

## Dependencies

- **Node.js 18+** - Runtime environment
- **PineTS** - Technical analysis engine
- **http-server** - Local development server
- **concurrently** - Multi-script runner