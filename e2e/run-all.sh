#!/bin/bash
# Run all E2E tests

set -e

echo "==================================="
echo "Running E2E Tests"
echo "==================================="
echo ""

FAILED=0
PASSED=0

for test in e2e/tests/*.mjs; do
  echo "Running: $test"
  if node "$test"; then
    PASSED=$((PASSED + 1))
    echo "✅ PASS"
  else
    FAILED=$((FAILED + 1))
    echo "❌ FAIL"
  fi
  echo ""
done

echo "==================================="
echo "E2E Test Results"
echo "==================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
  echo "✅ ALL TESTS PASSED"
  exit 0
else
  echo "❌ SOME TESTS FAILED"
  exit 1
fi
