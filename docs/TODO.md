# TODO List - BorisQuantLab Runner

## High Priority 🔴

- [ ] **PineTS: TR (True Range) variable not exposed to transpiled code** 🚨 CRITICAL
  - Bug report: `/Users/boris/proj/internal/borisquantlab/PineTS/BUG-TR-INCOMPLETE-FIX.md`
  - Status: TR calculated correctly, added to BUILT_IN_DATA_VARIABLES, but transpiled code throws `ReferenceError: tr is not defined`
  - Impact: All strategies using `tr`, ATR, ADX, DMI fail
  - Handoff: PineTS team to fix variable exposure in transpiler
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

## Current Status

- **Tests**: 554/554 unit + 14/14 E2E ✅ (100% pass rate)
- **Linting**: 0 errors ✅
- **Open Issues**: 1 critical (TR variable exposure in PineTS transpiler)
```

