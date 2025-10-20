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
  - Fixed v3/v4 syntax detection to exclude ta.\* functions
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
  - COMPLETED: Fixed colors migration (color=yellow → color=color.yellow). Fixed v3/v4 syntax detection regex (exclude ta.\* functions). Moved test files to tests/fixtures/strategies/. Extracted v3/v4 detection to PineVersionMigrator.hasV3V4Syntax() method. Updated src/index.js to use method instead of inline regex. Single source of truth achieved. All 297 tests passing. 0 linting errors.

### Performance & Statistics

- [x] **Eliminate API flooding for security() calls**
  - COMPLETED: Created TickeridMigrator utility (15 tests). Integrated into PineVersionMigrator. Implemented duration-based limit calculation in PineTS.security() (\_calculateDurationBasedLimit, \_timeframeToMinutes). Fixed symbol resolution (tickerid→syminfo.tickerid). Result: API requests reduced from 79→3 for daily-lines.pine (3 security() calls). All 312 tests passing. Cache key format: SBER|D|1 (correct limit).

## High Priority 🔴

- [x] **Enhance E2E test coverage with deterministic data**
  - **Status**: COMPLETED ✅ (100% deterministic)
  - **Completed**:
    - ✅ Created test-input-defval.mjs with MockProvider
    - ✅ Created test-reassignment.mjs with MockProvider (8/8 tests passing)
    - ✅ Created test-security.mjs with MockProvider (4/4 structural tests passing)
    - ✅ Updated MockProvider to use current timestamps (fixes data freshness validation)
    - ✅ Removed all redundant live API tests
    - ✅ Simplified test naming (removed -deterministic suffix)
  - **Validation**:
    - ✅ Input tests: SMA(14)=17 values, SMA(20)=11 values, SMA(10)=21 values
    - ✅ Reassignment tests: All 8 patterns validated with exact value matching
    - ✅ Security tests: Validates security() executes without crashes, correct structure
  - **Final Test Suite**:
    - 3 deterministic tests (100% - zero API dependencies)
  - **Current Tests**: 3/3 E2E tests passing

- [x] **Understand security() strategy rerun pattern**
  - **Observation**: PineTS security() reruns entire strategy code in nested context via `await pineTS.run(this.context.pineTSCode)` at PineTS/dist/pinets.dev.es.js:1794
  - **Goal**: Examine design goals and benefits of this pattern vs expression-only evaluation
  - **Test**: tests/issues/security-empty-object.test.js documents current behavior
  - **COMPLETED**: Full analysis of design pattern where security() reruns entire strategy in nested context for expression evaluation

- [x] **Migrate to PineTS rev3 API**
  - **COMPLETED**: Migrated prefetchSecurityData() from array to code string signature
  - **Evidence**: Removed ~60 lines (parseSecurityCalls, deduplicatePrefetchData, manual parsing)
  - **Result**: Single line `await pineTS.prefetchSecurityData(wrappedCode)`
  - **Validation**: All 358 tests passing, daily-lines.pine executes successfully
  - **Documentation**: PINETS_REV3_MIGRATION.md

- [x] **Fix security() empty object bug in PineTS library**
  - **Status**: RESOLVED ✅
  - **Resolution**: Fixed in PineTS library - security() now correctly returns numeric values
  - **Validation**: daily-lines.pine strategy executes successfully with security() calls
  - **Result**: Security function fully functional with indicator expressions

- [x] **Implement downscaling for PineTS security()**
  - **Status**: COMPLETED ✅
  - **Solution**: Option B - downsample parameter added to security() function
  - **Strategies**: "first", "last", "high", "low", "avg", "mean"
  - **Default**: "last" (uses last confirmed lower TF bar, no lookahead)
  - **Validation**: daily-lines.pine executes on W timeframe (100 candles, 3829.75ms)
  - **Result**: Downscaling functional, all 351 tests passing

- [x] **Fix reassignment operator (:=) broken**
  - **Status**: COMPLETED ✅
  - **Fix**: AST transformation converts `:=` to `=` for non-declaration reassignments
  - **Validation**: 8/8 reassignment E2E tests passing, 475 unit tests passing

- [x] **Fix security() returning identical values**
  - **Status**: COMPLETED ✅
  - **Root Cause**: \_findSecContextIdx() offset bug + paramArray[0] undefined at beginning indices
  - **Fix**: (1) Changed return from `i+(lookahead?1:2)` to `i`, (2) Added fallback to cached.data.close[secIdx] when paramArray[secIdx] undefined
  - **Validation**: 3/3 security E2E tests passing, 8/8 reassignment tests passing, 475 unit tests passing
  - **Cleanup**: Deleted 9 debug test files, kept test-security.mjs and test-reassignment.mjs

## Medium Priority 🟡

- [x] Fix pagination issue : add tolerance to overlapping pages returned by MOEX provider
- [x] **Fix upscaling issue : Yahoo security() sparse points**
  - **Root Cause**: Yahoo hardcoded closeTime = openTime + 60000ms regardless of timeframe
  - **Fix**: Changed to closeTime = openTime + intervalMinutes _ 60 _ 1000 - 1
  - **Validation**: AMZN 15m + daily security() now shows stepped lines (229.6874992371 repeated)
  - **Result**: Universal timeframe support
- [x] **Fix upscaling issue : `security()` extra 500 candlesticks for upper timeframes**
  - **Status**: RESOLVED ✅
  - **Resolution**: Fixed in PineTS library via TimeframeCalculator.calculateAdjustedLimit()
  - **Implementation**: `if (targetTfMinutes > sourceTfMinutes) return baseLimit + UPSCALING_BUFFER`
  - **Result**: Upscaling now adds 500-bar buffer for cumulative studies (ta.sma(200), ta.ema(50))
  - **Validation**: PineTS line 1755-1763 confirms UPSCALING_BUFFER applied only when targetTfMinutes > sourceTfMinutes

- [x] **Fix PineTS downscaling sparse data issue**
  - **Status**: RESOLVED ✅
  - **Resolution**: Fixed in PineTS library via TimeframeCalculator.calculateAdjustedLimit()
  - **Implementation**: Duration-based calculation delegates to TimeframeCalculator static method
  - **Result**: Downscaling uses correct baseLimit calculation without unnecessary buffer

- [x] **Investigate provider limit vs requested bars mismatch**
  - **Status**: RESOLVED ✅
  - **Evidence**: MOEX pagination successfully returns 700 W bars (500 + 200)
  - **Validation**: docker exec runner node src/index.js SBER W 700 strategies/ema-strategy.pine
  - **Result**: Pagination working correctly, no limit mismatch exists

- [x] **Debug and fix rolling-cagr strategy issues**
  - **Status**: COMPLETED ✅
  - **PineTS Changes**: Applied (format, scale, timeframe helpers, security() fixes)
  - **Validation**: CHMF M 72 rolling-cagr.pine executes successfully
  - **Result**: 12 non-null CAGR values calculated (bars 61-72), range -3.04% to 10.63%
  - **Tests**: 477 unit + 4 e2e passing

- [x] **Fix adapter to pass all plot() parameters**
  - **Status**: COMPLETED ✅
  - **Implementation**: IIFE pattern passes all options through (avoids PineTS transformation)
  - **Coverage**: All 15 Pine v5 plot() parameters supported
  - **Tests**: 8 adapter unit tests + test-plot-params.mjs E2E test passing

- [x] **Update adapter tests for all plot() parameters**
  - **Status**: COMPLETED ✅
  - **Tests**: transp, histbase, offset, linewidth, style validated
  - **File**: tests/adapters/PinePlotAdapter.test.js (8 tests)

- [x] **Add E2E test for plot() parameters**
  - **Status**: COMPLETED ✅
  - **File**: e2e/tests/test-plot-params.mjs
  - **Validation**: Full pipeline (Pine → Parser → Adapter → PineTS → JSON) verified

- [x] **Update documentation for plot() parameter support**
  - **Status**: COMPLETED ✅
  - **File**: docs/PINETS_COMPATIBILITY_v3.md
  - **Content**: 15 parameters table with status, examples, architecture diagram

- [x] **Fix color test expectations after PineTS update**
  - **Status**: COMPLETED ✅
  - Updated test expectations: 'blue'→'#2962FF', 'red'→'#FF5252', 'green'→'#4CAF50', 'purple'→'#9C27B0'
  - Tests: 477/477 unit + 5/5 E2E passing

- [x] **Implement --settings CLI using PineTS native inputOverrides**
  - **Status**: COMPLETED ✅
  - Usage: `node src/index.js CHMF M 72 strategies/rolling-cagr.pine --settings='{"Rolling CAGR Year N":3}'`
  - E2E test: test-input-override.mjs

- [x] **Remove dead code and consolidate color constants**
  - **Status**: COMPLETED ✅
  - Removed: PINE_COLORS mapping, resolvePineColor(), hexToRgba(), entire chartColors.js file
  - Moved: CHART_COLORS to src/config.js with other shared constants
  - Tests: 477/477 unit + 5/5 E2E passing

- [x] **Design command-line input parameters for rolling CAGR**
  - **Status**: COMPLETED ✅ (implemented with --settings)

- [ ] **Design extension for BB strategies v7, 8, 9**
  - Analyze bb-strategy-7-rus.pine, bb-strategy-8-rus.pine, bb-strategy-9-rus.pine requirements
  - Design and plan necessary code extensions: new indicators, signal logic, parameter handling, strategy-specific features

## Low Priority 🟢

- [ ] **Rework determineChartType() to support multi-pane indicators**
  - **Status**: NOT STARTED - requires Pine Script multi-pane research
  - **Current**: Hardcoded logic in ConfigurationBuilder.determineChartType(key)
  - **Issue**: Uses string matching (CAGR, EMA, SMA, MA) to determine chart assignment
  - **Goal**: Align with Pine Script's native multi-pane indicator approach
  - **Research Needed**:
    - How TradingView Pine Script v5 handles multi-pane indicators
    - Pine Script syntax for specifying plot pane assignment
    - Whether to use display parameter, overlay setting, or other mechanism
  - **Implementation**:
    - Parse pane assignment from Pine Script indicator/plot metadata
    - Remove hardcoded string matching logic
    - Support TradingView-compatible pane configuration
  - **Example**: rolling-cagr-5-10yr.pine should declare all plots in same pane via Pine Script

- [ ] **Design Y-axis scale configuration for plots**
  - **Status**: NOT STARTED - design phase required
  - **Goal**: Each plot should be able to specify its Y-axis scale settings
  - **Considerations**:
    - How to configure in Pine Script (TradingView-compatible approach)
    - Mapping to lightweight-charts priceScaleId
    - Default behavior when not specified
    - Multi-pane vs single-pane layouts
  - **Example Use Case**: rolling-cagr-5-10yr.pine with 4 plots needing independent scales
  - **Next**: Research TradingView Pine Script approach for multi-scale indicators

- [ ] **Implement varip runtime persistence in PineTS**
  - **Status**: LOW PRIORITY - not needed for current strategies
  - **Required**: Context.varipStorage Map, initVarIp/setVarIp methods, parser AST transformation
  - **Note**: May be needed for future strategies with intra-bar state requirements

- [ ] **Increase test coverage to 80%**
  - Add unit tests for uncovered code paths
  - Focus on error handling, edge cases, provider chain logic

- [ ] **Increase test coverage to 95%**
  - Comprehensive test coverage including integration tests, edge cases, error scenarios across all modules

---

## Current Status

- **Total Tests**: 477/477 passing ✅
- **Linting**: 0 errors ✅
- **E2E Tests**: 5/5 passing ✅
  - test-input-defval.mjs: Input parameter defaults ✅
  - test-input-override.mjs: Input parameter overrides ✅
  - test-plot-params.mjs: Plot parameters ✅
  - test-reassignment.mjs: Reassignment operator ✅
  - test-security.mjs: Security function ✅
- **PineTS Integration**: Format/scale/timeframe context complete ✅
- **Plot Parameters**: All 15 Pine v5 parameters supported ✅
- **Color Tests**: Fixed for PineTS hex format (blue→#2196F3, red→#F23645, etc) ✅
- **rolling-cagr**: Working (requires 5-year history for 5-year CAGR) ✅
- **Input Overrides**: CLI --settings parameter implemented and tested ✅
- **Next Task**: Design extension for BB strategies v7, 8, 9 🎯
- **Race Condition Fix**: Duplicate API calls eliminated ✅
- **Universal Utilities**: deduplicate() with keyGetter pattern ✅
- **API Stats**: Tab-separated ASCII format ✅
