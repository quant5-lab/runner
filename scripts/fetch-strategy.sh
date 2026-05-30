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
METADATA_FILE="$TEMP_DIR/metadata.json"
node -e "
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('$SYMBOL', '$TIMEFRAME', $BARS)
      .then(result => {
        const fs = require('fs');
        fs.writeFileSync('$BINANCE_FILE', JSON.stringify(result.data, null, 2));
        fs.writeFileSync('$METADATA_FILE', JSON.stringify({ symbol: '$SYMBOL', timezone: result.timezone, provider: result.provider }, null, 2));
        console.log('✓ Fetched ' + result.data.length + ' bars from ' + result.provider + ' (timezone: ' + result.timezone + ')');
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
node scripts/convert-binance-to-standard.cjs "$BINANCE_FILE" "$DATA_FILE" "$METADATA_FILE" > /dev/null || {
    echo "❌ Failed to convert data format"
    exit 1
}

# Normalize timeframe for filename (D → 1D, W → 1W, M → 1M)
NORM_TIMEFRAME="$TIMEFRAME"
if [ "$TIMEFRAME" = "D" ]; then
    NORM_TIMEFRAME="1D"
elif [ "$TIMEFRAME" = "W" ]; then
    NORM_TIMEFRAME="1W"
elif [ "$TIMEFRAME" = "M" ]; then
    NORM_TIMEFRAME="1M"
fi

# Save to test data directory for future use
TESTDATA_DIR="tests/fixtures/ohlcv"
mkdir -p "$TESTDATA_DIR"
SAVED_FILE="${TESTDATA_DIR}/${SYMBOL}_${NORM_TIMEFRAME}.json"
cp "$DATA_FILE" "$SAVED_FILE"
echo "  Saved: $SAVED_FILE"

# Detect security() calls and fetch additional timeframes
echo "  Checking for security() calls..."
SECURITY_TFS=$(grep -o "security([^)]*)" "$STRATEGY" | grep -o "'[^']*'" | tr -d "'" | grep -v "^$" | sort -u || true)
for SEC_TF in $SECURITY_TFS; do
    # Skip if same as base timeframe
    if [ "$SEC_TF" = "$TIMEFRAME" ]; then
        continue
    fi
    
    # Normalize timeframe (D → 1D, W → 1W, M → 1M)
    NORM_TF="$SEC_TF"
    if [ "$SEC_TF" = "D" ]; then
        NORM_TF="1D"
    elif [ "$SEC_TF" = "W" ]; then
        NORM_TF="1W"
    elif [ "$SEC_TF" = "M" ]; then
        NORM_TF="1M"
    fi
    
    SEC_FILE="${TESTDATA_DIR}/${SYMBOL}_${NORM_TF}.json"
    
    # Check if cached file exists and is recent enough
    FETCH_NEEDED=false
    if [ ! -f "$SEC_FILE" ]; then
        FETCH_NEEDED=true
    else
        # Check file age: refetch if older than 1 day for intraday/daily, 7 days for weekly/monthly
        FILE_MOD_TIME=$(stat -c %Y "$SEC_FILE" 2>/dev/null || stat -f %m "$SEC_FILE" 2>/dev/null || echo 0)
        FILE_AGE_HOURS=$(( ($(date +%s) - FILE_MOD_TIME) / 3600 ))
        
        case "$NORM_TF" in
            *m|*h|D|1D)
                # Intraday/daily: refetch if older than 24 hours
                if [ "$FILE_AGE_HOURS" -gt 24 ]; then
                    FETCH_NEEDED=true
                    echo "  Cached $NORM_TF data is $FILE_AGE_HOURS hours old, refetching..."
                fi
                ;;
            W|1W|*W)
                # Weekly: refetch if older than 7 days (168 hours)
                if [ "$FILE_AGE_HOURS" -gt 168 ]; then
                    FETCH_NEEDED=true
                    echo "  Cached $NORM_TF data is $FILE_AGE_HOURS hours old, refetching..."
                fi
                ;;
            M|1M|*M)
                # Monthly: refetch if older than 30 days (720 hours)
                if [ "$FILE_AGE_HOURS" -gt 720 ]; then
                    FETCH_NEEDED=true
                    echo "  Cached $NORM_TF data is $FILE_AGE_HOURS hours old, refetching..."
                fi
                ;;
        esac
    fi
    
    if [ "$FETCH_NEEDED" = true ]; then
        # Calculate needed bars: base_bars * timeframe_ratio + 500 (conservative warmup)
        # For weekly base with 500 bars: 500 * 7 + 500 = 4000 daily bars needed
        SEC_BARS=$((BARS * 10 + 500))
        echo "  Fetching security timeframe: $NORM_TF (need ~$SEC_BARS bars for warmup)"
        SEC_TEMP="$TEMP_DIR/security_${NORM_TF}.json"
        SEC_STD="$TEMP_DIR/security_${NORM_TF}_std.json"
        
        node -e "
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('$SYMBOL', '$NORM_TF', $SEC_BARS)
      .then(result => {
        const fs = require('fs');
        fs.writeFileSync('$SEC_TEMP', JSON.stringify(result.data, null, 2));
        console.log('  ✓ Fetched ' + result.data.length + ' ' + '$NORM_TF' + ' bars');
      })
      .catch(err => {
        console.error('  Warning: Could not fetch $NORM_TF data:', err.message);
        process.exit(0);
      });
  });
});
        " || echo "  Warning: Failed to fetch $NORM_TF data"
        
        if [ -f "$SEC_TEMP" ]; then
            node scripts/convert-binance-to-standard.cjs "$SEC_TEMP" "$SEC_STD" > /dev/null 2>&1 || true
            if [ -f "$SEC_STD" ]; then
                cp "$SEC_STD" "$SEC_FILE"
                echo "  Saved: $SEC_FILE"
            fi
        fi
    fi
done

# Step 2: Build strategy binary
echo ""
echo "[2/4] 🔨 Building strategy binary..."
STRATEGY_NAME=$(basename "$STRATEGY" .pine)
OUTPUT_BINARY="/tmp/${STRATEGY_NAME}"

# Generate Go code from Pine Script
TEMP_GO=$(go run cmd/pine-gen/main.go -input "$STRATEGY" -output "$OUTPUT_BINARY" 2>&1 | grep "Generated:" | awk '{print $2}')
if [ -z "$TEMP_GO" ]; then
    echo "❌ Failed to generate Go code"
    exit 1
fi

# Compile binary (from root where go.mod exists)
go build -o "$OUTPUT_BINARY" "$TEMP_GO" > /dev/null 2>&1 || {
    echo "❌ Failed to compile binary"
    exit 1
}
echo "✓ Binary: $OUTPUT_BINARY"

# Step 3: Execute strategy
echo ""
echo "[3/4] ⚡ Executing strategy..."
mkdir -p out
"$OUTPUT_BINARY" \
    -symbol "$SYMBOL" \
    -timeframe "$TIMEFRAME" \
    -data "$DATA_FILE" \
    -datadir tests/fixtures/ohlcv \
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
