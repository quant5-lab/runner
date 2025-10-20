# End-to-End Tests

This directory contains end-to-end tests for the BorisQuantLab Runner.

## Structure

```
e2e/
├── fixtures/
│   └── strategies/          # Pine Script test strategies
│       └── *.pine
├── tests/
│   └── *.mjs                # Test runner scripts
└── README.md
```

## Running Tests

### Individual Test

```bash
# From project root
docker compose run --rm runner node e2e/tests/test-reassignment-operator.mjs
```

### All E2E Tests

```bash
# From project root
docker compose run --rm runner sh -c "for test in e2e/tests/*.mjs; do node \$test || exit 1; done"
```

## Tests

### test-reassignment-operator.mjs

**Purpose**: Validates the fix for Pine Script `:=` reassignment operator with historical references

**Coverage**:

- Simple cumulative counters
- Step counters with different increments
- Conditional counters (BB strategy pattern)
- Running max/min (tracking highest/lowest values)
- Accumulators with conditional reset
- Multiple reassignments in sequence
- Trailing stop level patterns (BB v8 strategy)
- Session counters with `nz()` pattern
- Multi-historical references ([1], [2], [3])

**Expected Results**: All 10 tests should pass

**Strategy File**: `e2e/fixtures/strategies/test-reassignment-operator.pine`

## Adding New Tests

1. Create Pine Script strategy in `e2e/fixtures/strategies/`
2. Create test runner in `e2e/tests/`
3. Follow naming convention: `test-*.mjs` and `test-*.pine`
4. Update this README with test description

## Test Patterns

### Basic Structure

```javascript
import { createContainer } from '../../src/container.js';
import { createProviderChain, DEFAULTS } from '../../src/config.js';
import { readFile } from 'fs/promises';

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/your-strategy.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);

const result = await runner.runPineScriptStrategy(
  'BTCUSDT',
  '1h',
  10,
  jsCode,
  'your-strategy.pine',
);

// Validate results...
console.log(result.plots);
```

### Validation

- Extract plot data: `result.plots?.['Plot Name']?.data?.map(d => d.value)`
- Check for expected patterns
- Use `process.exit(0)` for pass, `process.exit(1)` for fail

## CI Integration

These tests are designed to run in Docker and can be integrated into CI pipelines:

```yaml
# Example GitHub Actions
- name: Run E2E Tests
  run: |
    docker compose run --rm runner sh -c "
      for test in e2e/tests/*.mjs; do
        echo \"Running \$test\"
        node \$test || exit 1
      done
    "
```
