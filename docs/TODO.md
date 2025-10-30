# TODO List - BorisQuantLab Runner

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

## High Priority 🔴

- [ ] **BB Strategy 7 - full execution validation**
  - ✅ dirmov() function scoping fixed
  - ⏳ End-to-end strategy execution with real data

## Medium Priority 🟡

- [ ] Remove sma_cache optimization from PineTS
- [ ] Fix null handling in averaging functions (treat as average propagation, not zero)
- [ ] Rework determineChartType() for multi-pane indicators (research Pine Script native approach)
- [ ] Design Y-axis scale configuration (priceScaleId mapping)
- [ ] Implement varip runtime persistence (Context.varipStorage, initVarIp/setVarIp)

## Low Priority 🟢

- [ ] Increase test coverage to 80%
- [ ] Increase test coverage to 95%
- [ ] Support blank candlestick mode (plots-only for capital growth modeling)
- [ ] Python unit tests for parser.py (90%+ coverage goal)
- [ ] Remove parser dead code ($.let.glb1_ wrapping, unused _rename_identifiers_in_ast)

---

## Current Status

- **Tests**: 515/515 unit + 10/10 E2E ✅
- **Linting**: 0 errors ✅
- **E2E Suite**: test-function-vs-variable-scoping, test-input-defval/override, test-plot-params, test-reassignment, test-security, test-strategy (bearish/bullish/base), test-ta-functions
- **Strategy Validation**: bb-strategy-7/8/9-rus, ema-strategy, daily-lines-simple, daily-lines, rolling-cagr, rolling-cagr-5-10yr ✅
