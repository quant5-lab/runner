# E2E Test Suite

Centralized test runner with automatic test discovery and failure tracking.

## Architecture

```
e2e/
├── runner.mjs          # Centralized test orchestrator
├── run-all.sh          # Shell wrapper (delegates to runner.mjs)
├── tests/              # Individual test files
│   ├── test-built-in-variables.mjs  # Parametric tests for all built-in variables
│   ├── test-edge-cases.mjs          # Edge cases (first bar, gaps, values)
│   ├── test-indicators.mjs          # Technical indicators (ATR, ADX, DMI)
│   ├── test-function-vs-variable-scoping.mjs
│   ├── test-input-defval.mjs
│   ├── test-input-override.mjs
│   ├── test-multi-pane.mjs
│   ├── test-plot-color-variables.mjs
│   ├── test-plot-params.mjs
│   ├── test-reassignment.mjs
│   ├── test-security.mjs
│   ├── test-session-filtering.mjs
│   ├── test-strategy.mjs
│   ├── test-strategy-bearish.mjs
│   ├── test-strategy-bullish.mjs
│   ├── test-ta-functions.mjs
│   ├── test-timezone-session.mjs
│   └── test-tr.mjs              # TR bug regression tests (legacy)
├── fixtures/           # Test data and strategies
│   └── strategies/
│       ├── test-builtin-*.pine      # Built-in variable fixtures
│       ├── test-edge-*.pine         # Edge case fixtures
│       ├── test-tr-*.pine           # TR-specific fixtures (legacy)
│       └── ...
├── mocks/              # Mock providers
│   └── MockProvider.js
└── utils/              # Shared test utilities
    └── test-helpers.js
```

## Test Runner Features

- **Automatic test discovery**: Scans `tests/` directory for `.mjs` files
- **Failure tracking**: Counts passed/failed tests with detailed reporting
- **Timeout protection**: 60s timeout per test
- **Percentage metrics**: Shows pass/fail rates
- **Duration tracking**: Per-test and total suite timing
- **Exit code**: Returns non-zero on any failure

## Usage

```bash
# Run all e2e tests in Docker
pnpm e2e

# Run directly (requires environment setup)
node e2e/runner.mjs
```

## Test Coverage

### Current Tests (15)

1. **test-function-vs-variable-scoping.mjs** - Function/variable scoping rules
2. **test-input-defval.mjs** - Input default values
3. **test-input-override.mjs** - Input parameter overrides
4. **test-multi-pane.mjs** - Multi-pane chart rendering
5. **test-plot-color-variables.mjs** - Plot color variable handling
6. **test-plot-params.mjs** - Plot parameter validation
7. **test-reassignment.mjs** - Variable reassignment rules
8. **test-security.mjs** - Security function behavior
9. **test-session-filtering.mjs** - Session filtering logic
10. **test-strategy.mjs** - Basic strategy execution
11. **test-barmerge.mjs** - Barmerge function testing
12. **test-fixnan.mjs** - NaN handling and fixnan function
13. **test-function-vs-variable-scoping.mjs** - Variable scope rules
14. **test-input-defval.mjs** - Input default value handling
15. **test-input-override.mjs** - Input parameter overrides
16. **test-multi-pane.mjs** - Multi-pane indicator rendering
17. **test-plot-color-variables.mjs** - Plot color with variables
18. **test-plot-params.mjs** - Plot parameter validation
19. **test-reassignment.mjs** - Variable reassignment rules
20. **test-security.mjs** - Security function testing
21. **test-session-filtering.mjs** - Session filtering logic
22. **test-strategy.mjs** - Strategy execution
23. **test-strategy-bearish.mjs** - Bearish strategy patterns
24. **test-strategy-bullish.mjs** - Bullish strategy patterns
25. **test-ta-functions.mjs** - Technical analysis functions
26. **test-timezone-session.mjs** - Timezone/session handling
27. **test-built-in-variables.mjs** - Parametric tests for all built-in variables (6 scenarios)
28. **test-edge-cases.mjs** - Edge cases for all variables (3 scenarios)
29. **test-indicators.mjs** - Technical indicators (ATR, ADX, DMI) (3 scenarios)
30. **test-tr.mjs** - True Range bug regression (legacy, 11 scenarios)

### Built-in Variables Test Coverage (test-built-in-variables.mjs)

Parametric tests validating all 9 built-in variables (open, high, low, close, volume, hl2, hlc3, ohlc4, tr):

1. **Direct access** - All base variables accessible (5 variables)
2. **Derived calculation** - Derived variables match formula (4 variables)
3. **Variables in calculations** - SMA with each variable (4 variables)
4. **Variables in conditionals** - Signal generation (4 variables)
5. **Variables in function scope** - Scoping test (4 variables)
6. **Multiple simultaneous usages** - Same script, multiple variables

### Edge Cases Test Coverage (test-edge-cases.mjs)

Edge case validation applicable to all built-in variables:

1. **First bar behavior** - Variables without historical data
2. **Gap detection** - Variables with price discontinuities
3. **Edge value handling** - Zero/negative value handling

### Technical Indicators Test Coverage (test-indicators.mjs)

TR-dependent indicator validation:

1. **ATR calculation** - Average True Range with manual validation
2. **ADX/DMI indicators** - Directional movement indicators
3. **BB7 regression** - Original TR bug in strategy context

### TR Test Coverage (test-tr.mjs) - LEGACY

Comprehensive coverage for True Range variable bug fix (preserved for regression):

1. **Direct TR access** - Basic TR variable exposure
2. **TR in calculations** - SMA, EMA with TR
3. **ATR calculation** - Internal TR usage validation
4. **TR in conditional logic** - TR-based signal generation
5. **TR in strategy logic** - Entry/exit with TR
6. **TR with ADX/DMI** - Complex indicators using TR
7. **Edge case: First bar** - TR without previous close
8. **Edge case: Gaps** - TR with price discontinuities
9. **TR in function scope** - Scoping test
10. **Multiple TR usages** - Same script, multiple TR references
11. **Regression test** - BB7 ADX bug (original issue)

**Note**: test-tr.mjs is preserved for regression validation but has been superseded by the generalized test suite (test-built-in-variables.mjs, test-edge-cases.mjs, test-indicators.mjs).

All tests validate against manual calculations to ensure correctness.

## Output Format

```
═══════════════════════════════════════════════════════════
E2E Test Suite
═══════════════════════════════════════════════════════════

Discovered 6 tests

Running: test-input-defval.mjs
✅ PASS (2341ms)

Running: test-ta-functions.mjs
❌ FAIL (1523ms)
Error output:
AssertionError: Expected 10.5, got 10.6

═══════════════════════════════════════════════════════════
Test Summary
═══════════════════════════════════════════════════════════
Total:    6
Passed:   5 (83.3%)
Failed:   1 (16.7%)
Duration: 8.45s

Failed Tests:
  ❌ test-ta-functions.mjs (exit code: 1)

❌ SOME TESTS FAILED
```

## Adding New Tests

Create a new `.mjs` file in `tests/` directory:

```javascript
#!/usr/bin/env node
import { strict as assert } from 'assert';

console.log('Test: My New Feature');

/* Test logic here */
assert.strictEqual(actual, expected);

console.log('✅ PASS');
process.exit(0);
```

Test runner automatically discovers and executes it.

## Test Guidelines

- Exit with code 0 for success, non-zero for failure
- Use `console.log()` for test output
- Keep tests under 60s timeout
- Use deterministic data from MockProvider
- Include assertion context in error messages
