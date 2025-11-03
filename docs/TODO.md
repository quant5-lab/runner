# TODO List - BorisQuantLab Runner

## High Priority 🔴

- [ ] **/anti-delusion-v2.1**
  - In the index.html create ability to add unlimited number of indicator panes by specifying any arbitrary name in `pane` property. These plots are to be grouped by name and separate indicator pane are to be rendered for each group (pane name).
  - Equity of bb7 must go into this separate pane.

## Completed ✅

- [x] Pine v3/v4→v5 migration (100+ function mappings, 37 tests)
- [x] Unified timeframe format (D/W/M, TimeframeParser/Converter refactor)
- [x] E2E test suite reorganization (centralized runner, timeout protection)
- [x] Plot adapter refactored (PinePlotAdapter module, 6 tests)
- [x] ESLint compliance (0 errors)
- [x] API flooding fix (79→3 requests via TickeridMigrator)
- [x] Parameter shadowing fix (_param_rename_stack, 11 tests)
- [x] Chart alignment fix (lineSeriesAdapter refactored to pure functions)
- [x] E2E deterministic tests (MockProvider, 100% coverage)
- [x] PineTS rev3 API migration (prefetchSecurityData)
- [x] security() downscaling (6 strategies: first/last/high/low/avg/mean)
- [x] Reassignment operator (:=) AST transformation
- [x] security() identical values bug (offset + fallback fix)
- [x] Provider pagination (MOEX 700W bars)
- [x] Rolling CAGR strategy (5Y/10Y support)
- [x] Plot parameters (all 15 Pine v5 params, test-plot-params.mjs)
- [x] Input overrides CLI (--settings parameter)
- [x] Color hex format tests (PineTS compatibility)
- [x] Strategy namespace (strategy() → strategy.call() transpiler)
- [x] ATR risk management (80% ATR14 SL, 5:1 RR, locked levels)
- [x] **Function vs Variable scoping bug (bb-strategy-7-rus.pine)**
  - User-defined functions incorrectly wrapped as $.let.glb1_*
  - Parser fix: track const vs let declarations in ScopeChain
  - Functions stay bare, variables wrapped for PineTS Context
  - 4 strategies validated + new E2E test
- [x] **Chart Y-axis auto-scaling bug with SMA warm-up periods**
  - **Fixed**: Changed anchor point `value: 0` → `value: NaN` in lineSeriesAdapter
  - NaN prevents auto-scale inclusion (Lightweight Charts official pattern)
  - Charts now scale to actual data range (min..max) instead of 0..max
- [x] **PineTS sma_cache optimization removed**
  - Cache removed from TechnicalAnalysis.ts sma() method
  - Direct calculation: `sma(reversedSource, period)` without caching
- [x] **Null handling in averaging functions (PineTS)**
  - **Fixed**: If ANY value in window is NaN/null/undefined, result is NaN
  - Matches Pine Script v5 behavior: NaN propagation, not zero substitution
  - Applied to: ta.sma and other averaging functions

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
    - [x] Session/Time Filters - ✅ Works
    - [x] SMAs (current + 1D via security()) - ✅ Works
    - [x] Bollinger Bands (bb_buy/bb_sell signals) - ✅ Works
    - [x] ADX/DMI (dirmov() → adx() → buy/sell signals) - ⚠️ SUSPICIOUS
    - [x] Stop Loss (fixed + trailing) - ⚠️ SUSPICIOUS (never enters trades)
    - [x] Take Profit (fixed + smart S&R detection) - ⚠️ SUSPICIOUS (TP not locked on entry, S&R always at 0)
    - [x] Volatility Check (atr vs sl) - ✅ Works
    - [x] Potential Check (distance to targets) - ✅ Works
  - **All mechanisms dissected - Ready for pair debugging to isolate calculation bugs**

## Medium Priority 🟡

- [ ] **Common PineScript plot parameters (line width, etc.) must be configurable**
  - Most plot parameters currently not configurable
  - Need user control over visual properties (linewidth, transparency, style, etc.)
- [ ] **Strategy trade consistency and math correctness unvalidated**
  - **Tech Debt**: No strict deterministic tests asserting correctness for each trade
  - Need deep validation: entry/exit prices, position sizes, P&L calculations, stop-loss/take-profit levels
  - Current E2E tests verify execution completes, but don't validate trade logic accuracy

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
