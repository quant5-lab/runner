# E2E Test Fixtures

Pine Script test fixtures used by Go integration tests.

## Structure

```
e2e/
└── fixtures/
    └── strategies/     # Pine Script test cases
        ├── test-*.pine           # Active test fixtures
        └── test-*.pine.skip      # Pending implementation
```

## Usage

Go tests reference these fixtures:

```go
// tests/test-integration/integration_test.go
strategyPath := "../../e2e/fixtures/strategies/test-strategy.pine"
```

## Test Coverage

- Built-in variables (bar_index, close, high, etc.)
- Technical indicators (ATR, SMA, RSI, etc.)
- Strategy functions (entry, exit, close)
- Edge cases (first bar, NaN handling, etc.)
- Security/multi-timeframe patterns

**Note:** Legacy Node.js e2e test runner removed. All tests now in Go test suite.
