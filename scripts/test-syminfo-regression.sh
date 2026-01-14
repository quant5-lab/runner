#!/bin/bash
# Regression Testing Script for syminfo.tickerid feature
# Usage: ./scripts/test-syminfo-regression.sh

set -e

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 syminfo.tickerid Regression Test Suite"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

FAILED=0
PASSED=0

# Navigate to project root
cd "$(dirname "$0")/.."

echo "📋 Test 1/6: Integration Tests"
echo "────────────────────────────────────────────────────────────"
if go test -v ./tests/test-integration -run Syminfo 2>&1 | tee /tmp/syminfo-test.log | grep -q "PASS"; then
    PASS_COUNT=$(grep -c "^--- PASS:" /tmp/syminfo-test.log || echo 0)
    echo "✅ PASS: $PASS_COUNT/6 integration tests passing"
    PASSED=$((PASSED + 1))
else
    echo "❌ FAIL: Integration tests failed"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "📋 Test 2/6: Basic syminfo.tickerid Build"
echo "────────────────────────────────────────────────────────────"
if make build-strategy STRATEGY=strategies/test-security.pine OUTPUT=test-regression-1 > /dev/null 2>&1; then
    echo "✅ PASS: test-security.pine compiled"
    PASSED=$((PASSED + 1))
else
    echo "❌ FAIL: test-security.pine compilation failed"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "📋 Test 3/6: Multiple Security Calls (DRY)"
echo "────────────────────────────────────────────────────────────"
if make build-strategy STRATEGY=testdata/e2e/test-multi-security.pine OUTPUT=test-regression-2 > /dev/null 2>&1; then
    echo "✅ PASS: test-multi-security.pine compiled"
    PASSED=$((PASSED + 1))
else
    echo "❌ FAIL: test-multi-security.pine compilation failed"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "📋 Test 4/6: Literal Symbol Regression"
echo "────────────────────────────────────────────────────────────"
if make build-strategy STRATEGY=testdata/e2e/test-literal-security.pine OUTPUT=test-regression-3 > /dev/null 2>&1; then
    echo "✅ PASS: test-literal-security.pine compiled (hardcoded symbols still work)"
    PASSED=$((PASSED + 1))
else
    echo "❌ FAIL: test-literal-security.pine compilation failed"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "📋 Test 5/6: Complex Expression"
echo "────────────────────────────────────────────────────────────"
if make build-strategy STRATEGY=testdata/e2e/test-complex-syminfo.pine OUTPUT=test-regression-4 > /dev/null 2>&1; then
    echo "✅ PASS: test-complex-syminfo.pine compiled"
    PASSED=$((PASSED + 1))
else
    echo "❌ FAIL: test-complex-syminfo.pine compilation failed"
    FAILED=$((FAILED + 1))
fi
echo ""

echo "📋 Test 6/6: Full Test Suite"
echo "────────────────────────────────────────────────────────────"
if go test ./... -timeout 30m > /tmp/syminfo-full-test.log 2>&1; then
    echo "✅ PASS: Full test suite passing (no regressions)"
    PASSED=$((PASSED + 1))
    cd ..
else
    echo "❌ FAIL: Full test suite has failures"
    echo "See /tmp/syminfo-full-test.log for details"
    FAILED=$((FAILED + 1))
    cd ..
fi
echo ""

# Cleanup
rm -f build/test-regression-*

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Regression Test Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  Passed: $PASSED/6"
echo "  Failed: $FAILED/6"
echo ""

if [ "$FAILED" -eq 0 ]; then
    echo "✅ SUCCESS: All regression tests passed"
    echo "🎯 syminfo.tickerid feature is stable"
    echo ""
    exit 0
else
    echo "❌ FAILURE: $FAILED regression test(s) failed"
    echo "⚠️  Feature stability compromised - investigate immediately"
    echo ""
    exit 1
fi
