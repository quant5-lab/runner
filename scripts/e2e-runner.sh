#!/bin/bash
# E2E Test Runner for Pine strategies
# Centralized orchestrator for all Pine script validation

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TESTDATA_FIXTURES_DIR="$PROJECT_ROOT/testdata/fixtures"
TESTDATA_E2E_DIR="$PROJECT_ROOT/testdata/e2e"
STRATEGIES_DIR="$PROJECT_ROOT/strategies"
BUILD_DIR="$PROJECT_ROOT/build"
DATA_DIR="$PROJECT_ROOT/tests/fixtures/ohlcv"
OUTPUT_DIR="$PROJECT_ROOT/out"

# Test tracking
TOTAL=0
PASSED=0
FAILED=0
SKIPPED=0
FAILED_TESTS=()
SKIPPED_TESTS=()

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 E2E Test Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Ensure build directory exists
mkdir -p "$BUILD_DIR"
mkdir -p "$OUTPUT_DIR"

# Build pine-gen if not exists
if [ ! -f "$BUILD_DIR/pine-gen" ]; then
    echo "📦 Building pine-gen..."
    cd "$PROJECT_ROOT" && make build > /dev/null 2>&1
    echo "✅ pine-gen built"
    echo ""
fi

# Discover testdata/fixtures/*.pine files
if [ -d "$TESTDATA_FIXTURES_DIR" ]; then
    FIXTURES_FILES=$(find "$TESTDATA_FIXTURES_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
    FIXTURES_COUNT=$(echo "$FIXTURES_FILES" | grep -c . || echo 0)
else
    FIXTURES_FILES=""
    FIXTURES_COUNT=0
fi

# Discover testdata/e2e/*.pine files
if [ -d "$TESTDATA_E2E_DIR" ]; then
    E2E_FILES=$(find "$TESTDATA_E2E_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
    E2E_COUNT=$(echo "$E2E_FILES" | grep -c . || echo 0)
else
    E2E_FILES=""
    E2E_COUNT=0
fi

# Discover strategies/*.pine files
if [ -d "$STRATEGIES_DIR" ]; then
    STRATEGY_FILES=$(find "$STRATEGIES_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
    STRATEGY_COUNT=$(echo "$STRATEGY_FILES" | grep -c . || echo 0)
else
    STRATEGY_FILES=""
    STRATEGY_COUNT=0
fi

TOTAL=$((FIXTURES_COUNT + E2E_COUNT + STRATEGY_COUNT))

echo "📋 Discovered $TOTAL test files:"
echo "   - testdata/fixtures/*.pine: $FIXTURES_COUNT unit test fixtures"
echo "   - testdata/e2e/*.pine: $E2E_COUNT e2e test strategies"
echo "   - strategies/*.pine: $STRATEGY_COUNT production strategies"
echo ""

# Test function
run_test() {
    local PINE_FILE="$1"
    local TEST_NAME=$(basename "$PINE_FILE" .pine)
    local OUTPUT_BINARY="$BUILD_DIR/e2e-$TEST_NAME"
    local SKIP_FILE="${PINE_FILE}.skip"
    
    echo "────────────────────────────────────────────────────────────"
    echo "Running: $TEST_NAME"
    echo "────────────────────────────────────────────────────────────"
    
    # Check for skip file
    if [ -f "$SKIP_FILE" ]; then
        SKIP_REASON=$(head -1 "$SKIP_FILE")
        echo "⏭️  SKIP: $SKIP_REASON"
        echo ""
        SKIPPED=$((SKIPPED + 1))
        SKIPPED_TESTS+=("$TEST_NAME: $SKIP_REASON")
        return 0
    fi
    
    # Build strategy
    if ! make -C "$PROJECT_ROOT" -s build-strategy \
        STRATEGY="$PINE_FILE" \
        OUTPUT="$OUTPUT_BINARY" > /tmp/e2e-build-$TEST_NAME.log 2>&1; then
        echo "❌ BUILD FAILED"
        echo ""
        FAILED=$((FAILED + 1))
        FAILED_TESTS+=("$TEST_NAME (build)")
        return 1
    fi
    
    # Find suitable data file, fetch if none exists
    DATA_FILE=""
    if [ -f "$DATA_DIR/BTCUSDT_1h.json" ]; then
        DATA_FILE="$DATA_DIR/BTCUSDT_1h.json"
    elif [ -f "$DATA_DIR/BTCUSDT_1D.json" ]; then
        DATA_FILE="$DATA_DIR/BTCUSDT_1D.json"
    else
        # Use first available data file
        DATA_FILE=$(find "$DATA_DIR" -name "*.json" -type f | head -1)
    fi
    
    if [ -z "$DATA_FILE" ]; then
        # No data files exist - fetch default test data
        echo "📡 No cached data found, fetching BTCUSDT 1h (500 bars)..."
        mkdir -p "$DATA_DIR"
        
        # Fetch data using Node.js providers
        TEMP_DIR=$(mktemp -d)
        trap "rm -rf $TEMP_DIR" RETURN
        
        BINANCE_FILE="$TEMP_DIR/binance.json"
        METADATA_FILE="$TEMP_DIR/metadata.json"
        STANDARD_FILE="$DATA_DIR/BTCUSDT_1h.json"
        
        # Node.js fetch command
        if ! node -e "
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('BTCUSDT', '1h', 500)
      .then(result => {
        const fs = require('fs');
        fs.writeFileSync('$BINANCE_FILE', JSON.stringify(result.data, null, 2));
        fs.writeFileSync('$METADATA_FILE', JSON.stringify({ timezone: result.timezone, provider: result.provider }, null, 2));
        console.log('✓ Fetched ' + result.data.length + ' bars from ' + result.provider);
      })
      .catch(err => {
        console.error('Error:', err.message);
        process.exit(1);
      });
  });
});" > /tmp/e2e-fetch-$TEST_NAME.log 2>&1; then
            # If fetch fails, skip test with explanation
            echo "⚠️  SKIP: Failed to fetch test data (network issue or provider unavailable)"
            echo ""
            SKIPPED=$((SKIPPED + 1))
            SKIPPED_TESTS+=("$TEST_NAME: Network data fetch failed")
            return 0
        fi
        
        # Convert to standard format
        if ! node scripts/convert-binance-to-standard.cjs "$BINANCE_FILE" "$STANDARD_FILE" "$METADATA_FILE" > /dev/null 2>&1; then
            echo "⚠️  SKIP: Failed to convert data format"
            echo ""
            SKIPPED=$((SKIPPED + 1))
            SKIPPED_TESTS+=("$TEST_NAME: Data conversion failed")
            return 0
        fi
        
        DATA_FILE="$STANDARD_FILE"
        echo "✓ Data fetched and cached: $DATA_FILE"
    fi
    
    # Determine symbol and timeframe from data file
    SYMBOL=$(basename "$DATA_FILE" | sed 's/_[^_]*\.json//')
    TIMEFRAME=$(basename "$DATA_FILE" .json | sed 's/.*_//')
    
    # Detect security() calls and fetch additional timeframes for the SAME symbol
    SECURITY_TFS=$(grep -o "security([^)]*)" "$PINE_FILE" | grep -oE "\"[^\"]+\"|'[^']+'" | tr -d "\"'" | grep -E "^(1h|1D|1W|1M|D|W|M)$" | sort -u || true)
    for SEC_TF in $SECURITY_TFS; do
        # Normalize timeframe
        NORM_TF="$SEC_TF"
        [ "$SEC_TF" = "D" ] && NORM_TF="1D"
        [ "$SEC_TF" = "W" ] && NORM_TF="1W"
        [ "$SEC_TF" = "M" ] && NORM_TF="1M"
        
        SEC_FILE="$DATA_DIR/${SYMBOL}_${NORM_TF}.json"
        
        # Skip if already exists (avoid re-downloading)
        if [ -f "$SEC_FILE" ]; then
            echo "  ✓ Using cached: $SEC_FILE"
            continue
        fi
        
        # Fetch additional timeframe for the same symbol
        echo "  📡 Fetching security() timeframe: $SYMBOL $SEC_TF..."
        
        TEMP_DIR=$(mktemp -d)
        trap "rm -rf $TEMP_DIR" RETURN
        
        BINANCE_FILE="$TEMP_DIR/binance.json"
        METADATA_FILE="$TEMP_DIR/metadata.json"
        
        if node -e "
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('$SYMBOL', '$SEC_TF', 500)
      .then(result => {
        const fs = require('fs');
        fs.writeFileSync('$BINANCE_FILE', JSON.stringify(result.data, null, 2));
        fs.writeFileSync('$METADATA_FILE', JSON.stringify({ timezone: result.timezone, provider: result.provider }, null, 2));
      })
      .catch(err => {
        console.error('Error:', err.message);
        process.exit(1);
      });
  });
});" > /dev/null 2>&1; then
            node scripts/convert-binance-to-standard.cjs "$BINANCE_FILE" "$SEC_FILE" "$METADATA_FILE" > /dev/null 2>&1
            echo "  ✓ Fetched and cached: $SEC_FILE"
        else
            echo "  ⚠️  Failed to fetch $SYMBOL $SEC_TF (will skip if strategy errors)"
        fi
    done
    
    # Execute strategy
    
    if ! "$OUTPUT_BINARY" \
        -symbol "$SYMBOL" \
        -timeframe "$TIMEFRAME" \
        -data "$DATA_FILE" \
        -datadir "$DATA_DIR" \
        -output "$OUTPUT_DIR/e2e-$TEST_NAME-output.json" > /tmp/e2e-run-$TEST_NAME.log 2>&1; then
        echo "❌ EXECUTION FAILED"
        echo ""
        FAILED=$((FAILED + 1))
        FAILED_TESTS+=("$TEST_NAME (execution)")
        return 1
    fi
    
    echo "✅ PASS"
    echo ""
    PASSED=$((PASSED + 1))
    
    # Cleanup binary
    rm -f "$OUTPUT_BINARY"
    return 0
}

# Run fixtures
if [ $FIXTURES_COUNT -gt 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📂 Testing testdata/fixtures/*.pine (unit test fixtures)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    while IFS= read -r PINE_FILE; do
        [ -z "$PINE_FILE" ] && continue
        run_test "$PINE_FILE"
    done <<< "$FIXTURES_FILES"
fi

# Run e2e test strategies
if [ $E2E_COUNT -gt 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📂 Testing testdata/e2e/*.pine (e2e test strategies)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    while IFS= read -r PINE_FILE; do
        [ -z "$PINE_FILE" ] && continue
        run_test "$PINE_FILE"
    done <<< "$E2E_FILES"
fi

# Run strategy files
if [ $STRATEGY_COUNT -gt 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📂 Testing strategies/*.pine (production strategies)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    while IFS= read -r PINE_FILE; do
        [ -z "$PINE_FILE" ] && continue
        run_test "$PINE_FILE"
    done <<< "$STRATEGY_FILES"
fi

# Cleanup temp files
rm -f /tmp/e2e-*.log
rm -f "$OUTPUT_DIR"/e2e-*-output.json

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 E2E Test Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Total:   $TOTAL"
echo "  Passed:  $PASSED"
echo "  Skipped: $SKIPPED"
echo "  Failed:  $FAILED"
echo ""

if [ $SKIPPED -gt 0 ]; then
    echo "Skipped tests (not yet implemented):"
    for TEST in "${SKIPPED_TESTS[@]}"; do
        echo "  ⏭️  $TEST"
    done
    echo ""
fi

if [ $FAILED -gt 0 ]; then
    echo "Failed tests:"
    for TEST in "${FAILED_TESTS[@]}"; do
        echo "  ❌ $TEST"
    done
    echo ""
    echo "❌ E2E SUITE FAILED"
    exit 1
else
    TESTABLE=$((TOTAL - SKIPPED))
    if [ $TESTABLE -gt 0 ]; then
        PASS_RATE=$((PASSED * 100 / TESTABLE))
        echo "✅ SUCCESS: All testable E2E tests passed ($PASSED/$TESTABLE = $PASS_RATE%)"
    else
        echo "✅ SUCCESS: All tests passed"
    fi
    exit 0
fi
