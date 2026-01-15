#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../.."

generate_data() {
    echo "Generating test data files..."
    
    go run tests/golden/cmd/gendata/main.go -symbol AAPL -timeframe 1h -bars 5500 -output tests/golden/fixtures/data/AAPL-1h.json
    go run tests/golden/cmd/gendata/main.go -symbol AAPL -timeframe D -bars 252 -output tests/golden/fixtures/data/AAPL-D.json
    go run tests/golden/cmd/gendata/main.go -symbol AAPL -timeframe W -bars 52 -output tests/golden/fixtures/data/AAPL-W.json
    go run tests/golden/cmd/gendata/main.go -symbol AAPL -timeframe M -bars 120 -output tests/golden/fixtures/data/AAPL-M.json
    go run tests/golden/cmd/gendata/main.go -symbol AAPL_1D -timeframe D -bars 252 -output tests/golden/fixtures/data/AAPL_1D.json
    
    go run tests/golden/cmd/gendata/main.go -symbol BTCUSDT -timeframe 1h -bars 5500 -output tests/golden/fixtures/data/BTCUSDT-1h.json
    go run tests/golden/cmd/gendata/main.go -symbol BTCUSDT -timeframe M -bars 120 -output tests/golden/fixtures/data/BTCUSDT-M.json
    go run tests/golden/cmd/gendata/main.go -symbol BTCUSDT_1D -timeframe D -bars 365 -output tests/golden/fixtures/data/BTCUSDT_1D.json
    
    go run tests/golden/cmd/gendata/main.go -symbol SBERP -timeframe 1h -bars 5500 -output tests/golden/fixtures/data/SBERP-1h.json
    go run tests/golden/cmd/gendata/main.go -symbol SBERP -timeframe M -bars 120 -output tests/golden/fixtures/data/SBERP-M.json
    go run tests/golden/cmd/gendata/main.go -symbol SBERP_1D -timeframe D -bars 252 -output tests/golden/fixtures/data/SBERP_1D.json
    
    echo "Data generation complete"
}

update_golden() {
    echo "Updating golden files..."
    go test -v ./tests/golden/... -update-golden
    echo "Golden files updated"
}

run_tests() {
    echo "Running golden file regression tests..."
    go test -v ./tests/golden/...
}

case "${1:-}" in
    generate)
        generate_data
        ;;
    update)
        update_golden
        ;;
    test)
        run_tests
        ;;
    all)
        generate_data
        update_golden
        run_tests
        ;;
    *)
        echo "Usage: $0 {generate|update|test|all}"
        echo ""
        echo "Commands:"
        echo "  generate  - Generate synthetic test data"
        echo "  update    - Update golden files with current results"
        echo "  test      - Run regression tests"
        echo "  all       - Generate data, update golden files, and test"
        exit 1
        ;;
esac
