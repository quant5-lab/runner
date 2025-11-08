# TODO List - BorisQuantLab Runner

## High Priority 🔴

- [ ] **Strategy trade timestamp accuracy**
  - Current: trades use `Date.now()` for entryTime/exitTime (all same timestamp)
  - Need: Use actual bar timestamp from candlestick data
  - Impact: Trade timing analysis currently requires mapping via entryBar/exitBar indices

## Medium Priority 🟡

- [ ] **Common PineScript plot parameters (line width, etc.) must be configurable**
  - Most plot parameters currently not configurable
  - Need user control over visual properties (linewidth, transparency, style, etc.)
- [ ] **Strategy trade consistency and math correctness validation**
  - Trades executing but need deep validation of trade logic accuracy
  - Verify: entry/exit prices, position sizes, P&L calculations, stop-loss/take-profit levels
  - Current: Basic execution verified, detailed correctness unvalidated

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

- **Tests**: 554/554 unit + 18/18 E2E ✅ (100% pass rate, 202.90s duration)
- **Linting**: 0 errors ✅
- **Open Issues**: 0 critical ✅
```

