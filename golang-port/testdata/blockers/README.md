# PineScript Blocker Test Files

Minimal test cases for documented PineScript support blockers.

## Purpose

Automated regression safety for 13 documented blockers preventing 100% PineScript support.

## Test Coverage

| File | Feature | Parse | Generate | Compile | Status |
|------|---------|-------|----------|---------|--------|
| test-for-loop.pine | for loops | ✅ | ✅ | ❌ | Generates literals, not loop execution |
| test-while-loop.pine | while loops | ❌ | ❌ | ❌ | Parse error: "binary expression in condition" |
| test-var-decl.pine | var declarations | ✅ | ✅ | ✅ | Working |
| test-label.pine | label.new/set_text | ✅ | ✅ | ✅ | Working |
| test-array.pine | array functions | ✅ | ✅ | ✅ | Working |
| test-strategy-exit.pine | strategy.exit | ✅ | ✅ | ✅ | Working |
| test-operators.pine | bitwise operators | ❌ | ❌ | ❌ | Lexer error: "invalid input" |
| test-ta-missing.pine | CCI/WMA/VWAP | ✅ | ✅ | ✅ | Working |
| test-color-funcs.pine | color.rgb/new | ✅ | ✅ | ✅ | Working |
| test-visual-funcs.pine | fill/bgcolor/hline | ✅ | ✅ | ✅ | Working |
| test-alert.pine | alert/alertcondition | ✅ | ✅ TODO | ✅ | Codegen TODO comments |
| test-string-funcs.pine | str.tostring/tonumber | ✅ | ✅ TODO | ✅ | Codegen TODO comments |
| test-map.pine | map generics | ❌ | ❌ | ❌ | Parse error: "unexpected token" |

## Test Execution

```bash
cd /home/coder/proj/quant5-lab/runner/golang-port
go test -v ./tests/test-integration -run TestPineScriptBlockers
```

## Documentation Sync

Tests synchronized with `/docs/BLOCKERS.md` via `TestBlockerDocumentation`.

## Maintenance

- Add new blocker: Create .pine file → Add test case → Update BLOCKERS.md
- Fix blocker: Update test expectations → Verify pass → Update docs
- Remove blocker: Change expectCompile → Verify pass → Update BLOCKERS.md summary
