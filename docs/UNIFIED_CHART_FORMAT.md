# Unified Chart Data Format

## Overview
Single-file JSON format combining all chart visualization data, metadata, and configuration. Replaces the previous 2-file design (chart-config.json + chart-data.json).

## Design Principles
- **Single Source of Truth**: All chart data in one file
- **Self-Describing**: Metadata embedded with data
- **Extensible**: Easy to add new fields without breaking compatibility
- **UI-Agnostic**: Separates data from presentation hints
- **Backward Compatible**: Maintains essential fields from legacy format

## Full Schema

```json
{
  "metadata": {
    "symbol": "BTCUSDT",
    "timeframe": "1h",
    "strategy": "SMA Crossover Strategy",
    "title": "SMA Crossover Strategy - BTCUSDT",
    "timestamp": "2025-11-16T00:14:28+03:00"
  },
  
  "candlestick": [
    {
      "time": 1700000000,
      "open": 100.0,
      "high": 105.0,
      "low": 95.0,
      "close": 102.0,
      "volume": 1000.0
    }
  ],
  
  "indicators": {
    "sma20": {
      "title": "SMA 20",
      "pane": "main",
      "style": {
        "color": "#2196F3",
        "lineWidth": 2
      },
      "data": [
        {"time": 1700000000, "value": 121.0}
      ]
    },
    "rsi": {
      "title": "RSI 14",
      "pane": "indicator",
      "style": {
        "color": "#FF9800",
        "lineWidth": 2
      },
      "data": [
        {"time": 1700000000, "value": 65.5}
      ]
    }
  },
  
  "strategy": {
    "trades": [
      {
        "entryId": "long1",
        "entryPrice": 100.0,
        "entryBar": 10,
        "entryTime": 1700036000,
        "exitPrice": 110.0,
        "exitBar": 15,
        "exitTime": 1700054000,
        "size": 10,
        "profit": 100.0,
        "direction": "long"
      }
    ],
    "openTrades": [
      {
        "entryId": "long2",
        "entryPrice": 140.0,
        "entryBar": 20,
        "entryTime": 1700072000,
        "size": 1,
        "direction": "long"
      }
    ],
    "equity": 10100.0,
    "netProfit": 100.0
  },
  
  "ui": {
    "panes": {
      "main": {
        "height": 400,
        "fixed": true
      },
      "indicator": {
        "height": 200,
        "fixed": false
      }
    }
  }
}
```

## Field Specifications

### metadata
Root-level metadata about the chart and strategy.

- **symbol** (string, required): Trading pair symbol (e.g., "BTCUSDT", "AAPL")
- **timeframe** (string, required): Bar interval (e.g., "1m", "5m", "1h", "1D")
- **strategy** (string, optional): Strategy name from Pine Script `indicator()` or `strategy()` call
- **title** (string, required): Display title (format: "{strategy} - {symbol}" or just symbol if no strategy)
- **timestamp** (string, required): ISO 8601 timestamp of chart generation

### candlestick
Array of OHLCV bars. Core price data for candlestick chart.

Each element:
- **time** (int64, required): Unix timestamp in seconds
- **open** (float64, required): Opening price
- **high** (float64, required): Highest price
- **low** (float64, required): Lowest price
- **close** (float64, required): Closing price
- **volume** (float64, required): Trading volume

### indicators
Map of indicator series. Key = indicator identifier (e.g., "sma20", "rsi14").

Each indicator:
- **title** (string, required): Display name for legend
- **pane** (string, required): Chart pane ("main" overlays candlesticks, "indicator" separate pane)
- **style** (object, required): Visual styling
  - **color** (string): Hex color code (e.g., "#2196F3")
  - **lineWidth** (int): Line thickness (1-5)
- **data** (array, required): Time-series data points
  - **time** (int64): Unix timestamp
  - **value** (float64): Indicator value

### strategy
Strategy execution results. Optional - only present if strategy() or strategy.entry() used.

- **trades** (array): Closed trades
  - **entryId** (string): Trade identifier from strategy.entry()
  - **entryPrice** (float64): Entry execution price
  - **entryBar** (int): Entry bar index
  - **entryTime** (int64): Entry Unix timestamp
  - **exitPrice** (float64): Exit execution price
  - **exitBar** (int): Exit bar index
  - **exitTime** (int64): Exit Unix timestamp
  - **size** (float64): Position size (shares/contracts)
  - **profit** (float64): Realized P&L
  - **direction** (string): "long" or "short"

- **openTrades** (array): Currently open trades (same fields as closed except no exit data)

- **equity** (float64): Current account equity (initial capital + realized P&L + unrealized P&L)
- **netProfit** (float64): Total realized profit/loss

### ui
UI configuration hints. Provides default visualization settings.

- **panes** (object): Pane layout configuration
  - **main** (object): Primary chart pane
    - **height** (int): Pixel height
    - **fixed** (bool): Whether height is locked
  - **indicator** (object): Secondary indicator pane
    - **height** (int): Pixel height
    - **fixed** (bool): Whether height is locked

## Extensibility

### Adding New Fields
1. **Metadata**: Add new root-level fields for chart-wide properties
2. **Indicators**: Add custom properties to indicator objects (e.g., `"smoothing": "sma"`)
3. **UI**: Add new pane types or configuration options
4. **Strategy**: Add performance metrics (e.g., `"sharpeRatio"`, `"maxDrawdown"`)

### Backward Compatibility
- Required fields must always be present
- Optional fields can be omitted
- Unknown fields are ignored by consumers
- Always validate JSON schema before production use

## Migration from Legacy Format

### Old Format (2 files)
**chart-config.json**:
```json
{
  "ui": {...},
  "dataSource": {...},
  "seriesConfig": {...}
}
```

**chart-data.json**:
```json
{
  "candlestick": [...],
  "plots": {...},
  "timestamp": "..."
}
```

### New Format (1 file)
All data merged into single `chart-data.json` with structured sections:
- `metadata` replaces `ui.title`, `ui.symbol`, etc.
- `indicators` replaces `plots` with richer metadata
- `ui` provides layout hints (optional)
- `candlestick` remains unchanged
- `strategy` unchanged

## Code Generation

The Go runtime automatically generates this format:

```go
// In generated strategy code
cd := chartdata.NewChartData(ctx, symbol, timeframe, strategyName)
cd.AddPlots(plotCollector)
cd.AddStrategy(strat, currentPrice)
jsonBytes, _ := cd.ToJSON()
```

Default styling applied automatically:
- 6 color rotation: Blue, Green, Orange, Red, Purple, Cyan
- Line width: 2px
- Pane assignment: "main" by default

## Visualization Example

Lightweight-charts v4.1.1 integration:

```javascript
fetch('chart-data.json')
  .then(r => r.json())
  .then(data => {
    // Title from metadata
    document.title = data.metadata.title;
    
    // Candlestick data
    chart.addCandlestickSeries().setData(data.candlestick);
    
    // Indicators
    for (const [key, indicator] of Object.entries(data.indicators)) {
      const series = chart.addLineSeries({
        color: indicator.style.color,
        lineWidth: indicator.style.lineWidth,
        title: indicator.title
      });
      series.setData(indicator.data);
    }
    
    // Strategy trades as markers
    const markers = data.strategy.trades.map(t => ({
      time: t.entryTime,
      position: 'belowBar',
      color: t.direction === 'long' ? '#26a69a' : '#ef5350',
      shape: t.direction === 'long' ? 'arrowUp' : 'arrowDown',
      text: `${t.direction} @${t.entryPrice}`
    }));
  });
```

## Performance

- **File Size**: ~14KB for 30 bars + 2 indicators + strategy (vs 2 separate files ~16KB total)
- **Parse Time**: Single JSON.parse() vs 2 separate fetches
- **Network**: 1 HTTP request vs 2
- **Caching**: Single cache entry vs coordination between 2 files

## Validation

Example JSON Schema validation (subset):

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["metadata", "candlestick", "indicators", "ui"],
  "properties": {
    "metadata": {
      "type": "object",
      "required": ["symbol", "timeframe", "title", "timestamp"]
    },
    "candlestick": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["time", "open", "high", "low", "close", "volume"]
      }
    }
  }
}
```

## References

- **Implementation**: `golang-port/runtime/chartdata/chartdata.go`
- **Tests**: `golang-port/runtime/chartdata/chartdata_test.go`
- **Template**: `golang-port/template/main.go.tmpl`
- **Manual Testing**: `MANUAL_TESTING.md`
