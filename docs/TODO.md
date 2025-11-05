# TODO List - BorisQuantLab Runner

## Completed ✅

- [x] **Multi-pane chart architecture (unlimited dynamic panes)**
- [x] **PineTS `na()` function bug - returns array instead of scalar**
  - ✅ Bug identified: `na()` returned `[false, false, ..., true]` instead of scalar boolean
  - ✅ Caused `!na(variable)` conditions to fail in strategy entry logic
  - ✅ Fixed by PineTS team
  - ✅ Validated on GDYN 1h 500 bars: 73 trades executing correctly, 18 in October 2025
- [x] **Strategy trade data capture and output**
  - ✅ Strategy trades now captured from PineTS context
  - ✅ Trade data (entry/exit prices, P&L, direction) exported to chart-data.json
  - ✅ Trade summary logging added to runner output
  - ✅ Validated on GDYN 1h 500 bars: 73 trades captured, 18 in October 2025

## High Priority 🔴

- [ ] **BB Strategy 7 - Calculation bugs investigation**
  - ✅ dirmov() function scoping fixed
  - ✅ Transpilation successful
  - ✅ All variable transformations working
  - ✅ Timeframe validation working
  - ✅ bb-strategy-7-debug.pine cloned for dissection
  - ❌ Complex interrelated calculation bugs present
  - **Dissection checklist:**
    - [x] 1D S&R Detection (pivothigh/pivotlow + security()) - ✅ Works
    - [ ] Session/Time Filters -  Suspicious, time filter always 0
    - [x] SMAs (current + 1D via security()) - ✅ Works
    - [x] Bollinger Bands (bb_buy/bb_sell signals) - ✅ Works
    - [x] Stop Loss (fixed + trailing) - ✅ Works (trades executing with exits)
    - [ ] ADX/DMI (dirmov() → adx() → buy/sell signals) - ⚠️ SUSPICIOUS
    - [ ] Take Profit (fixed + smart S&R detection) - ⚠️ SUSPICIOUS (TP not locked on entry, S&R always at 0)
    - [x] Volatility Check (atr vs sl) - ✅ Works
    - [x] Potential Check (distance to targets) - ✅ Works
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

- **Tests**: 515/515 unit + 10/10 E2E ✅
- **Linting**: 0 errors ✅
- **E2E Suite**: test-function-vs-variable-scoping, test-input-defval/override, test-plot-params, test-reassignment, test-security, test-strategy (bearish/bullish/base), test-ta-functions
- **Strategy Validation**: bb-strategy-7/8/9-rus, ema-strategy, daily-lines-simple, daily-lines, rolling-cagr, rolling-cagr-5-10yr ✅
- **Strategy Execution**: bb7-dissect-sl.pine on GDYN 1h 500 bars - 73 trades, 18 in October 2025, Net P/L: $-0.83 ✅

