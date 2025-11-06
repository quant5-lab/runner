# TODO List - BorisQuantLab Runner

## Completed ✅

- [x] **Multi-pane chart architecture (unlimited dynamic panes)**
- [x] **PineTS `na()` function bug - returns array instead of scalar**
- [x] **Strategy trade data capture and output**
- [x] **PineTS timezone-aware session filtering**
  - ✅ time() function now uses exchange timezone (was ignoring timezone parameter)
  - ✅ SYMBOL_TIMEZONES registry provides timezone mapping (GDYN → America/New_York)
  - ✅ SessionMask with bitmask optimization for O(1) session checks
  - ✅ E2E test added: test-timezone-session.mjs
- [x] **MockProvider milliseconds convention**
  - ✅ Fixed openTime to return milliseconds (was returning seconds)
  - ✅ Now matches real providers (YahooFinance, MOEX, Binance)

## High Priority 🔴

- [ ] **BB Strategy 7 - Session time filter BROKEN**
  - ✅ Timezone fix verified (GDYN uses America/New_York)
  - ✅ Session filtering mechanism working correctly (E2E test passes)
  - ❌ Session detection completely broken in BB7 strategy: all 500 bars marked as IN
  - ❌ Session plots show all 1s (should show 0s for pre-market/after-hours)
  - **Root cause**: time() function with session parameter returns non-na for ALL bars
  - **Impact**: HIGH - Strategy cannot filter by session, trades executing outside intended hours
  - **Test case**: BB7 session=0950-1645 on 1m GDYN data shows 500/500 IN (impossible)
  - **Dissection checklist**:
    - [ ] Verify time() transpiled code in BB7 output
    - [ ] Compare BB7 transpilation vs E2E test transpilation
    - [ ] Check if session input parameter handled differently than hardcoded session
    - [ ] Validate timezone parameter passed correctly to time() in BB7
    - [ ] Test minimal BB7 subset with only session detection
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

- **Tests**: 554/554 unit + 13/13 E2E ✅ (100% pass rate)
- **Linting**: 0 errors ✅
- **E2E Suite**: test-timezone-session, test-function-vs-variable-scoping, test-input-defval/override, test-multi-pane, test-plot-color-variables, test-plot-params, test-reassignment, test-security, test-strategy (bearish/bullish/base), test-ta-functions
- **Time Units**: Milliseconds convention enforced (PineTS + all providers)


```

