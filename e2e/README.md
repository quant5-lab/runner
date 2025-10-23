# E2E Test Suite

Centralized test runner with automatic test discovery and failure tracking.

## Architecture

```
e2e/
├── runner.mjs          # Centralized test orchestrator
├── run-all.sh          # Shell wrapper (delegates to runner.mjs)
├── tests/              # Individual test files
│   ├── test-input-defval.mjs
│   ├── test-input-override.mjs
│   ├── test-plot-params.mjs
│   ├── test-reassignment.mjs
│   ├── test-security.mjs
│   └── test-ta-functions.mjs
├── fixtures/           # Test data and strategies
│   └── strategies/
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
