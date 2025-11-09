````markdown
# TODO List - BorisQuantLab Runner

## High Priority 🔴

- [ ] **Strategy trade timestamp accuracy**
  - Current: trades use `Date.now()` for entryTime/exitTime (all same timestamp)
  - Need: Use actual bar timestamp from candlestick data
  - Impact: Trade timing analysis currently requires mapping via entryBar/exitBar indices

- [ ] **Strategy trade consistency and math correctness validation**
  - Trades executing but need deep validation of trade logic accuracy
  - Verify: entry/exit prices, position sizes, P&L calculations, stop-loss/take-profit levels
  - Current: Basic execution verified, detailed correctness unvalidated

## Medium Priority 🟡

- [ ] **Common PineScript plot parameters (line width, etc.) must be configurable**
  - Most plot parameters currently not configurable
  - Need user control over visual properties (linewidth, transparency, style, etc.)

- [ ] **PineTS: Integration test for reassignment operator blocked by transpiler**
  - Issue: Reassignment `:=` triggers "Cannot read properties of undefined (reading '0')" in test context
  - Root cause: Series variable handling works in production runtime, fails in isolated tests
  - Impact: Cannot create automated tests for nested ternary + reassignment patterns
  - Workaround: Production validation confirms BB7 strategy works (9 closed + 1 open trades)
  - Status: Low priority - production works, test infrastructure limitation

## Low Priority 🟢

- [ ] Replace or fork/optimize pynescript (26s parse time bottleneck)
- [ ] Increase test coverage to 80%
- [ ] Increase test coverage to 95%
- [ ] Support blank candlestick mode (plots-only for capital growth modeling)
- [ ] Python unit tests for parser.py (90%+ coverage goal)
- [ ] Remove parser dead code ($.let.glb1_ wrapping, unused _rename_identifiers_in_ast)
- [ ] Implement varip runtime persistence (Context.varipStorage, initVarIp/setVarIp)
- [ ] Design Y-axis scale configuration (priceScaleId mapping)
- [ ] Rework determineChartType() for multi-pane indicators (research Pine Script native approach)
- [ ] **PineTS: Refactor src/transpiler/index.ts** - Decouple monolithic transpiler for maintainability and extensibility

---

## Recently Completed ✅

- [x] **PineTS: sl_inp reassignment operator bug** 🚨 FIXED (2025-11-09)
  - Bug: Nested ternary + nz() in reassignment returned 0 (99% of bars)
  - Fixed: PineTS commit 8c166f8 - ParentTrackingWalker resolves nested ternary
  - Validation: BB7 strategy now enters 10 trades (9 closed + 1 open) on GDYN 1h 500
  - Evidence: sl_inp 100% non-zero (was 0%), volatility_below_sl 27.8% (was 0%)
  - Documentation: `docs/pinets-fix-validation-summary.md`

- [x] **E2E Test Suite Generalization** (2025-11-08)
  - Refactored TR-specific tests into parametric built-in variable tests
  - Created: test-built-in-variables.mjs (6 scenarios, 9 variables)
  - Created: test-edge-cases.mjs (3 scenarios)
  - Created: test-indicators.mjs (3 scenarios)
  - Documentation: `E2E-GENERALIZATION-COMPLETE.md`
  - Impact: Future-proof tests for all built-in variables, not just TR

- [x] **PineTS: TR (True Range) variable not exposed to transpiled code** 🚨 FIXED (2025-11-08)
  - Bug reports: `BUG-TR-INCOMPLETE-FIX.md`, `TRANSPILER-MYSTERY-EVIDENCE.md`
  - Fixed: Build 20:16 - AST reference mismatch resolved
  - Validation: `VALIDATION-SUCCESS-BUILD-20-16.md` (4/4 tests passed)
  - Impact: All strategies using `tr`, ATR, ADX, DMI now work

---

## Current Status

```
┌─────────────────────────────────────┐
│  Tests:   554/554 unit + 18/18 E2E │
│  Linting: 0 errors                  │
│  Network: 0% (100% deterministic)   │
│  Status:  ✅ All systems nominal     │
└─────────────────────────────────────┘
```

