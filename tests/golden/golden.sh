#!/bin/bash
# Golden File Regression Testing
# Uses FROZEN REAL market data - DO NOT regenerate
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../.."

update_golden() {
    echo "Updating golden reference files..."
    go test -v ./tests/golden/... -update-golden
    echo "✓ Golden references updated"
}

run_tests() {
    echo "Running golden file regression tests..."
    go test -v ./tests/golden/...
}

verify_data() {
    echo "Verifying frozen market data integrity..."
    echo ""
    for f in tests/golden/fixtures/data/*-1h.json; do
        name=$(basename "$f")
        bars=$(jq '.bars | length' "$f" 2>/dev/null || echo "ERROR")
        first_close=$(jq '.bars[0].close' "$f" 2>/dev/null || echo "N/A")
        printf "  %-20s %5s bars  first_close=%.2f\n" "$name" "$bars" "$first_close"
    done
    echo ""
    echo "✓ Data verification complete"
}

case "${1:-}" in
    update)
        update_golden
        ;;
    test)
        run_tests
        ;;
    verify)
        verify_data
        ;;
    *)
        echo "Golden File Regression Testing"
        echo "==============================="
        echo "Uses FROZEN REAL market data from repository"
        echo ""
        echo "Usage: $0 {update|test|verify}"
        echo ""
        echo "Commands:"
        echo "  test    - Run regression tests against frozen data"
        echo "  update  - Update golden reference files (after intentional changes)"
        echo "  verify  - Verify frozen market data integrity"
        echo ""
        echo "NOTE: Market data is FROZEN in repository. Do not regenerate."
        echo "      Data source: Real historical OHLCV from external provider"
        exit 1
        ;;
esac
