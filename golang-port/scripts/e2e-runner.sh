#!/bin/bash
# E2E Test Runner for golang-port Pine strategies
# Centralized orchestrator for all Pine script validation

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
TESTDATA_FIXTURES_DIR="$PROJECT_ROOT/golang-port/testdata/fixtures"
TESTDATA_E2E_DIR="$PROJECT_ROOT/golang-port/testdata/e2e"
STRATEGIES_DIR="$PROJECT_ROOT/strategies"
BUILD_DIR="$PROJECT_ROOT/golang-port/build"
DATA_DIR="$PROJECT_ROOT/golang-port/testdata/ohlcv"
OUTPUT_DIR="$PROJECT_ROOT/out"

# Test tracking
TOTAL=0
PASSED=0
FAILED=0
FAILED_TESTS=()

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 golang-port E2E Test Suite"
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
FIXTURES_FILES=$(find "$TESTDATA_FIXTURES_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
FIXTURES_COUNT=$(echo "$FIXTURES_FILES" | grep -c . || echo 0)

# Discover testdata/e2e/*.pine files
E2E_FILES=$(find "$TESTDATA_E2E_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
E2E_COUNT=$(echo "$E2E_FILES" | grep -c . || echo 0)

# Discover strategies/*.pine files
STRATEGY_FILES=$(find "$STRATEGIES_DIR" -maxdepth 1 -name "*.pine" -type f 2>/dev/null | sort)
STRATEGY_COUNT=$(echo "$STRATEGY_FILES" | grep -c . || echo 0)

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
    
    echo "────────────────────────────────────────────────────────────"
    echo "Running: $TEST_NAME"
    echo "────────────────────────────────────────────────────────────"
    
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
    
    # Find suitable data file
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
        echo "⚠️  SKIP: No data files in $DATA_DIR"
        echo ""
        return 0
    fi
    
    # Execute strategy
    SYMBOL=$(basename "$DATA_FILE" | sed 's/_[^_]*\.json//')
    TIMEFRAME=$(basename "$DATA_FILE" .json | sed 's/.*_//')
    
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

# Run testdata fixtures
if [ $TESTDATA_COUNT -gt 0 ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📂 Testing testdata/*.pine fixtures"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    while IFS= read -r PINE_FILE; do
        [ -z "$PINE_FILE" ] && continue
        run_test "$PINE_FILE"
    done <<< "$TESTDATA_FILES"
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
echo "  Total:  $TOTAL"
echo "  Passed: $PASSED ($((PASSED * 100 / TOTAL))%)"
echo "  Failed: $FAILED ($((FAILED * 100 / TOTAL))%)"
echo ""

if [ $FAILED -gt 0 ]; then
    echo "Failed tests:"
    for TEST in "${FAILED_TESTS[@]}"; do
        echo "  ❌ $TEST"
    done
    echo ""
    echo "❌ E2E SUITE FAILED"
    exit 1
else
    echo "✅ SUCCESS: All E2E tests passed"
    exit 0
fi
