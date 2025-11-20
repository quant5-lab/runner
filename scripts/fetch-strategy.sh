#!/bin/bash
# Fetch live data and run strategy for development/testing
# Usage: ./scripts/fetch-strategy.sh SYMBOL TIMEFRAME BARS STRATEGY_FILE
#   Example: ./scripts/fetch-strategy.sh BTCUSDT 1h 500 strategies/daily-lines.pine

set -e

SYMBOL="${1:-}"
TIMEFRAME="${2:-1h}"
BARS="${3:-500}"
STRATEGY="${4:-}"

if [ -z "$SYMBOL" ] || [ -z "$STRATEGY" ]; then
    echo "Usage: $0 SYMBOL TIMEFRAME BARS STRATEGY_FILE"
    echo ""
    echo "Examples:"
    echo "  $0 BTCUSDT 1h 500 strategies/daily-lines.pine"
    echo "  $0 AAPL 1D 200 strategies/test-simple.pine"
    echo "  $0 SBER 1h 500 strategies/rolling-cagr.pine"
    echo "  $0 GDYN 1h 500 strategies/test-simple.pine"
    echo ""
    echo "Supported symbols:"
    echo "  - Crypto: BTCUSDT, ETHUSDT, etc. (Binance)"
    echo "  - US Stocks: AAPL, GOOGL, MSFT, GDYN, etc. (Yahoo Finance)"
    echo "  - Russian Stocks: SBER, GAZP, etc. (MOEX)"
    exit 1
fi

if [ ! -f "$STRATEGY" ]; then
    echo "Error: Strategy file not found: $STRATEGY"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 Running Strategy"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Symbol:     $SYMBOL"
echo "Timeframe:  $TIMEFRAME"
echo "Bars:       $BARS"
echo "Strategy:   $STRATEGY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Create temp directory for data
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

DATA_FILE="$TEMP_DIR/data.json"

# Step 1: Fetch data using Node.js (existing providers)
echo ""
echo "[1/4] 📡 Fetching market data..."
BINANCE_FILE="$TEMP_DIR/binance.json"
node -e "
import('./src/container.js').then(({ createContainer }) => {
  import('./src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.getMarketData('$SYMBOL', '$TIMEFRAME', $BARS)
      .then(bars => {
        const fs = require('fs');
        fs.writeFileSync('$BINANCE_FILE', JSON.stringify(bars, null, 2));
        console.log('✓ Fetched ' + bars.length + ' bars');
      })
      .catch(err => {
        console.error('Error fetching data:', err.message);
        process.exit(1);
      });
  });
});
" || {
    echo "❌ Failed to fetch data"
    exit 1
}

# Convert Binance format to standard OHLCV format
echo "  Converting to standard format..."
node scripts/convert-binance-to-standard.cjs "$BINANCE_FILE" "$DATA_FILE" > /dev/null || {
    echo "❌ Failed to convert data format"
    exit 1
}

# Save to test data directory for future use
TESTDATA_DIR="golang-port/testdata/ohlcv"
mkdir -p "$TESTDATA_DIR"
SAVED_FILE="${TESTDATA_DIR}/${SYMBOL}_${TIMEFRAME}.json"
cp "$DATA_FILE" "$SAVED_FILE"
echo "  Saved: $SAVED_FILE"

# Step 2: Build strategy binary
echo ""
echo "[2/4] 🔨 Building strategy binary..."
STRATEGY_NAME=$(basename "$STRATEGY" .pine)
OUTPUT_BINARY="/tmp/${STRATEGY_NAME}"

# Run builder with output flag
cd golang-port && go run cmd/pinescript-builder/main.go -input ../"$STRATEGY" -output "$OUTPUT_BINARY" > /dev/null 2>&1 || {
    echo "❌ Failed to build strategy"
    exit 1
}
cd ..
echo "✓ Binary: $OUTPUT_BINARY"

# Step 3: Execute strategy
echo ""
echo "[3/4] ⚡ Executing strategy..."
mkdir -p out
"$OUTPUT_BINARY" \
    -symbol "$SYMBOL" \
    -timeframe "$TIMEFRAME" \
    -data "$DATA_FILE" \
    -output out/chart-data.json || {
    echo "❌ Failed to execute strategy"
    exit 1
}
echo "✓ Output: out/chart-data.json"

# Step 4: Show results
echo ""
echo "[4/4] 📊 Results:"
cat out/chart-data.json | grep -E '"closedTrades"|"equity"|"netProfit"' | head -5 || true

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Strategy execution complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next steps:"
echo "  1. View chart: make serve"
echo "  2. Open browser: http://localhost:8000"
echo ""
