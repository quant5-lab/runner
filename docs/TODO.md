


For running use docker. Docker already running, just run start 

docker compose exec -e DEBUG=true runner node src/index.js SBER 10m 1000 strategies/test-security.pine

source code is volume mapped, and you must examine source code locally in this workspace. PineTS is located in sibling dir to this workspace. Also source code is available as a `pinets.dev.es.js` bundle attached

# TODO List - BorisQuantLab Runner

## Completed ✅

### Core Compatibility & Migration
- [x] Fix syminfo.ticker → syminfo.tickerid
  - Renamed syminfo.ticker to syminfo.tickerid for Pine Script v5 compatibility
  
- [x] Add security() compatibility wrapper
  - Added security() wrapper function that delegates to request.security() for v3/v4 compatibility
  
- [x] Add sma() compatibility wrapper
  - Added sma() wrapper function that delegates to ta.sma() for v3/v4 compatibility
  
- [x] Add indicator/strategy/study stubs
  - Added indicator(), strategy(), study() stub functions for Pine Script version compatibility

- [x] **Implement just-in-time Pine v3/v4→v5 migration**
  - COMPLETED: Created PineVersionMigrator class with 100+ v3/v4→v5 function mappings (study→indicator, sma→ta.sma, security→request.security, color constants, etc)
  - Integrated into pipeline (src/index.js)
  - Fixed PineScriptTranspiler.detectVersion() to return actual version
  - Added indicator/strategy stubs to runtime
  - Validated migration: v3→v5 ✅, v4→v5 ✅, v5 pass-through ✅
  - Fixed color migration (color=yellow → color=color.yellow)
  - Fixed v3/v4 syntax detection to exclude ta.* functions
  - Created 37 unit tests covering all migration scenarios
  - All 297 tests passing

### Timeframe System Refactoring
- [x] Update TimeframeError class with supportedTimeframes
  - TimeframeError(timeframe, symbol, providerName, supportedTimeframes) - error message MUST include list of supported timeframes
  
- [x] Update MoexProvider with supportedTimeframes list
  - Add supportedTimeframes property and pass to TimeframeError constructor. List: 1m, 10m, 1h, 1d, 1w, 1M
  
- [x] Update BinanceProvider with supportedTimeframes list
  - Add supportedTimeframes property and pass to TimeframeError constructor. List: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
  
- [x] Update YahooFinanceProvider with supportedTimeframes list
  - Add supportedTimeframes property and pass to TimeframeError constructor. List: 1m, 2m, 5m, 15m, 30m, 1h, 90m, 1d, 1wk, 1mo
  
- [x] Extract SUPPORTED_TIMEFRAMES to shared constant (DRY)
  - Extracted SUPPORTED_TIMEFRAMES constant to timeframeParser.js as single source of truth. All 3 providers now import and use shared constant. Eliminated hardcoded timeframe arrays. ARCHITECTURE: DRY/SOLID/KISS principles enforced
  
- [x] TECH DEBT: Migrate to unified timeframe format
  - PRIORITY: Update SUPPORTED_TIMEFRAMES to unified format: M (monthly), W (weekly), D (daily), 1h/2h/4h (hourly), 1m/5m/15m (minute). Eliminate inconsistent formats: 1wk→W, 1mo→M, 1d→D, 1w→W, 1M→M. Single source of truth for timeframe representation
  
- [x] TECH DEBT: Ensure conversion functions SOLID/DRY/KISS
  - COMPLETED: Refactored timeframe architecture. TimeframeParser = parsing (SRP), TimeframeConverter = format conversions (SRP). Eliminated duplication by moving conversions to TimeframeConverter. Fixed fromPineTS() return type bug (string not number). Fixed redundant error messages. All 254 tests passing. E2E validated.
  
- [x] TECH DEBT: Update provider timeframe mappings
  - COMPLETED: Verified all 3 providers accept unified D/W/M format. E2E tests confirm: BSPB D (100 candles), BSPB W (100 candles), BSPB M (100 candles) all execute successfully. Integration tests passing (14/14). TimeframeConverter mappings validated. All 254 tests passing.
  
- [x] TECH DEBT: Validate E2E with unified format
  - COMPLETED: E2E validation complete. BTCUSDT D (100 candles), BTCUSDT W (50 candles), BTCUSDT M (30 candles) all execute successfully via Binance provider. Provider chain correctly falls through MOEX→Binance. Error messages show unified format. All 254 tests passing.

### Test Infrastructure
- [x] Update ProviderManager.test.js - TimeframeError with 3 mocked providers
  - Unit tests for ProviderManager error handling: Mock 3 providers, test TimeframeError stops chain, [] continues chain, data returns success. Verify error message includes supported timeframes list. DONE: All 7 new tests passing
  
- [x] Fix TimeframeError message format in all tests
  - DONE: Updated test expectations across timeframeParser.test.js (7 fixes), timeframeIntegration.test.js (2 fixes), MoexProvider.test.js (1 fix). Changed from 'X does not support' to 'Timeframe X not supported'. Reduced total failures from 15 to 6
  
- [x] Fix BinanceProvider.test.js mock imports and error expectations
  - COMPLETED: (1) Added SUPPORTED_TIMEFRAMES export to vi.mock() for timeframeParser module. (2) Updated 'should propagate errors' test to expect empty array [] instead of thrown error (matches new error handling behavior where non-TimeframeError exceptions return []). All 249 tests passing
  
- [x] Fix PineScriptStrategyRunner.test.js expectations
  - COMPLETED: Fixed 2 test failures - (1) Updated test to expect 'const { plot: corePlot, color }' instead of 'const { plot, color }', (2) Updated test to expect 'function indicator()' instead of 'const indicator = ()'. Tests now match actual implementation
  
- [x] Fix timeframeParser.test.js parseToMinutes test
  - COMPLETED: Fixed 1 test failure - Test now expects 1w to return 10080 (valid weekly timeframe) instead of 1440 (invalid fallback to daily). Code correctly parses 1w as weekly format

### Code Quality
- [x] Refactor and extract plot() adapter
  - COMPLETED: Extracted plotAdapterSource to PinePlotAdapter.js module. Created 6 unit tests. Injected via ${plotAdapterSource} template string into PineScriptStrategyRunner. Fixed inline expression pattern (no const opts variable) to prevent PineTS $.const accumulation bug. Removed debug logs from PineDataSourceAdapter. All 260 tests passing. E2E validated: BTCUSDT W 1000 candles, 8685ms execution.
  
- [x] Format and satisfy linter, ensure tests not broken
  - COMPLETED: Ran eslint locally. Fixed 35 auto-fixable errors (trailing spaces, missing commas). Fixed 8 manual errors: removed unused 'instance' variables (2x), added TimeframeError import to YahooFinanceProvider, removed unused TimeframeError import from timeframeParser, fixed duplicate closeTime key in test, added eslint-disable comment for Function constructor. 0 linting errors. All 260 tests passing.

- [x] **TODO 23 FINAL CLEANUP**
  - COMPLETED: Fixed colors migration (color=yellow → color=color.yellow). Fixed v3/v4 syntax detection regex (exclude ta.* functions). Moved test files to tests/fixtures/strategies/. Extracted v3/v4 detection to PineVersionMigrator.hasV3V4Syntax() method. Updated src/index.js to use method instead of inline regex. Single source of truth achieved. All 297 tests passing. 0 linting errors.

## In Progress 🔄

- [ ] **Debug and fix daily-lines strategy issues**
  - Test daily-lines.pine strategy on multiple timeframes (1m, 5m, 15m, 1h, 4h, 1d, 1w)
  - Identify and fix any calculation errors, plot issues, or runtime failures
  - Focus on security() parameter passing bug: PineTS transpiler wraps arguments with .param() calls, causing tuples ['BSPB', 'p5'] instead of raw values
  - Current status: Migration working, colors fixed, but parameter passing issue remains

## High Priority 🔴

- [ ] **Implement remaining logic of security() method**
  - Complete implementation of security() method functionality beyond basic wrapper
  - Handle context resolution, data fetching for requested symbol/timeframe combinations
  - Fix parameter passing issue (tuples vs raw values)

- [ ] **Extend security() for higher and lower timeframes**
  - Adjust PineTS source code in sibling directory to support both higher and lower timeframe requests
  - Commit PineTS to repository, rebuild PineTS distribution
  - Current limitation: only higher timeframes supported

## Medium Priority 🟡

- [ ] **Debug and fix rolling-cagr strategy issues**
  - Test rolling-cagr.pine strategy on multiple timeframes
  - Identify and fix any calculation errors, CAGR computation issues, or runtime failures

- [ ] **Design extension for BB strategies v7, 8, 9**
  - Analyze bb-strategy-7-rus.pine, bb-strategy-8-rus.pine, bb-strategy-9-rus.pine requirements
  - Design and plan necessary code extensions: new indicators, signal logic, parameter handling, strategy-specific features

## Low Priority 🟢

- [ ] **Increase test coverage to 80%**
  - Add unit tests for uncovered code paths
  - Focus on error handling, edge cases, provider chain logic

- [ ] **Increase test coverage to 95%**
  - Comprehensive test coverage including integration tests, edge cases, error scenarios across all modules

---

## Current Status
- **Total Tests**: 297/297 passing ✅
- **Linting**: 0 errors ✅
- **Migration System**: Fully functional (v3/v4/v5 support) ✅
- **Next Focus**: Complete TODO 23 cleanup (move test files + extract v3/v4 detection)
