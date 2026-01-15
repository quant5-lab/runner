# Golden File Regression Tests

Frozen-data regression testing for strategies

## Commands

### Via Makefile (Recommended)

```bash
make test-golden              # Run tests
make test-golden-update       # Update baselines
```

### Via Script

```bash
cd tests/golden
./golden.sh generate          # Generate frozen data
./golden.sh update            # Generate golden baselines
./golden.sh test              # Run tests
./golden.sh all               # All steps
```

### Direct

```bash
go test ./tests/golden/...                    # Run tests
go test ./tests/golden/... -update-golden     # Update baselines
```

## Regression Testing Workflow

```bash
# 1. Before changes: Verify baseline
make test-golden  # PASS: 15/15

# 2. Make code changes
vim security/security.go

# 3. Detect regressions
make test-golden  # FAIL: Trade count mismatch

# 4. Fix and validate
make test-golden  # PASS: 15/15

# 5. If behavior changed intentionally:
make test-golden-update
git diff tests/golden/fixtures/expected/
git commit -m "Update golden baselines after architecture change"
```
