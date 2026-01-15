# Golden File Regression Tests

Frozen-data regression testing using **REAL historical market data**.

## Commands

```bash
make test-golden         # Run regression tests
make test-golden-update  # Update baselines
```

## Workflow

```bash
make test-golden              # Verify baseline
# make changes
make test-golden              # Detect regressions
make test-golden-update       # Update if intentional
```
