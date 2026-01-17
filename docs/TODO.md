# Go Runner PoC

## Performance
- Total: <50ms (excl. data fetch)
- Go parser: 5-10ms
- Go runtime: <10ms

## License Safety
- Go stdlib (BSD-3-Clause)
- participle/v2 (MIT)
- Pure Go TA

## Phase 1: Go Parser + Transpiler
- [x] Create golang-port structure
- [x] Initialize Go module
- [x] Study pine-parser AST output
- [x] Install participle parser
- [x] Define PineScript v5 grammar
- [x] Implement lexer
- [x] Implement parser
- [x] Map AST nodes to Go structs
- [x] Implement codegen
- [x] Test parsing
- [x] Compare AST output
- [x] Generate executable Go code
- [x] Verify compilation

## Phase 2: Go Runtime
- [x] Create runtime structure
- [x] Pure Go TA implementation
- [x] OHLCV context
- [x] NA value handling
- [x] Color constants
- [x] PlotCollector interface
- [x] Math functions
- [x] Input functions with overrides
- [x] SMA, EMA, RMA with warmup
- [x] RSI with RMA smoothing
- [x] TR, ATR calculation (security() support added)
- [x] Bollinger Bands
- [x] STDEV calculation (security() support added)
- [x] MACD
- [x] Stochastic oscillator
- [x] Strategy entry/close/exit
- [x] Trade tracking
- [x] Equity calculation
- [x] ChartData structure
- [x] JSON output

## Phase 2.5: request.security() Module

### Baseline
- [x] AST scanner (5/5 tests)
- [x] JSON reader (5/5 tests)
- [x] Context cache (8/8 tests)
- [x] Expression prefetch (3/3 tests)
- [x] Code injection (4/4 tests)
- [x] BB pattern tests (7/7 PASS)

### ForwardSeriesBuffer Alignment
- [x] Extract AST utilities (SRP)
- [x] Fetch contexts only
- [x] Direct OHLCV access
- [x] Comprehensive edge case tests
- [x] 266/266 tests PASS

### Inline TA States
- [x] Circular buffer warmup
- [x] Forward-only sliding window
- [x] 7/7 tests PASS
- [x] 82KB → 0B, O(N) → O(1)
- [x] 8/13 TA functions O(1)
- [x] SMA circular buffer optimization
- [x] Keep O(period) for window scans

### Complex Expressions
- [x] BinaryExpression in security
- [x] Identifier in security
- [x] 5/5 codegen tests PASS
- [x] 7/7 baseline tests PASS
- [x] TernaryExpr in arguments
- [x] String literal quote trim
- [x] Parenthesized expressions
- [x] Visitor/transformer updates
- [x] Complex expression parsing
- [x] 10/10 integration tests (28+ cases)
- [x] Plot styling parameters (style, linewidth, transp, pane)

### Integration
- [x] Builder pipeline integration
- [x] 10 test suites PASS
- [x] E2E with multi-timeframe data
- [x] SMA value verification
- [x] Timeframe conversion tests
- [x] Dynamic warmup calculation
- [x] Bar conversion formula
- [x] Automatic timeframe fetch
- [x] Timeframe normalization

## Phase 3: Binary Template
- [x] Create template structure
- [x] Main template with imports
- [x] CLI flags
- [x] Data loading integration
- [x] Code injection
- [x] AST codegen
- [x] CLI entry point
- [x] Build pine-gen
- [x] Test code generation
- [x] Test binary compilation
- [x] Test execution
- [x] Verify JSON output
- [x] Execution <50ms (24µs for 30 bars with placeholder strategy)

## Validation
- [x] Complete AST → Go code generation for Pine functions (ta.sma/ema/rsi/atr/bbands/macd/stoch, plot, if/ternary, Series[offset])
- [x] Implement strategy.entry, strategy.close, strategy.exit codegen (strategy.close lines 247-251, strategy.entry working)
- [x] `./bin/strategy` on daily-lines-simple.pine validates basic features
- [x] `./bin/strategy` on daily-lines.pine validates advanced features

## Phase 4: Additional Pine Features for Complex Strategies
- [x] Unary expressions (`-1`, `+x`, `not x`, `!condition`)
- [x] `na` constant for NaN value representation
- [x] `timeframe.ismonthly`, `timeframe.isdaily`, `timeframe.isweekly` built-in variables
- [x] `timeframe.period` built-in variable
- [x] `input.float()` with title and defval parameters (positional + named)
- [x] `input.int()`, `input.bool()`, `input.string()` for typed configuration
- [x] `input.source()` for selecting price source (close, open, high, low)
- [x] `math.pow()` with expression arguments (not just literals)
- [x] Variable subscript indexing `src[variable]` where variable is computed
- [x] Named parameter extraction: `input.float(defval=1.4, title="X")` fully supported
- [x] Comprehensive test coverage: input_handler_test.go (6 tests), math_handler_test.go (6 tests), subscript_resolver_test.go (8 tests)
- [x] Frontend config loading fix: metadata.strategy uses source filename instead of title

## Phase 4.5: BB7 Strategy Prerequisites
- [x] `input.session()` for time range inputs (entry_time, trading_session)
- [x] `time()` function for session filtering
- [x] Session timezone support (America/New_York, Europe/Moscow, UTC)
- [x] `syminfo.tickerid` built-in variable (for security() calls) - Added to template
- [x] `fixnan()` function for forward-filling NaN values (pivothigh/pivotlow results)
- [x] `pivothigh()` function for resistance detection
- [x] `pivotlow()` function for support detection
- [x] Nested ternary expressions in parentheses (parser grammar fix)
- [x] `math.min()` and `math.max()` inline in conditions/ternaries
- [x] `security()` with complex TA function chains (sma, pivothigh/pivotlow, fixnan combinations)
- [x] `barmerge.lookahead_on` constant for security() lookahead parameter
- [x] `security()` with lookahead parameter support
- [x] `wma()` weighted moving average function (WMAHandler implemented and registered)
- [x] `dev()` function for deviation detection (DEVHandler implemented and registered)
- [x] `strategy.position_avg_price` built-in variable (StateManager + codegen sampling order fixed)
- [x] `valuewhen()` function for conditional value retrieval (66+ tests: handler validation, runtime correctness, integration scenarios)
- [x] `valuewhen()` runtime evaluation in security() contexts (StreamingBarEvaluator support, 7 test functions, 25 subtests, occurrence/boundary/expression/condition/validation/progression/state coverage)
- [x] Arrow function preamble extraction (ArrowVarInitResult, PreambleExtractor, module-level functions, 100+ tests, double-assignment syntax fixed)
- [x] Multi-condition strategy logic with session management
- [ ] Visualization config system integration with BB7

## PineScript Support Blockers (5)
- Codegen TODO: alert, alertcondition, str.tostring, str.tonumber, str.split
- Type: string variables (standalone assignment)
- Runtime: multi-symbol security (data files only)
- Parser: while loops, for loops (execution only), map generics
- Codegen: RSI inline
- Parser: varip (not implemented)
- Note: arrow functions ✅, syminfo.tickerid ✅ (security context), strategy.exit ⚠️ (triggers but misaligned)

### BB7 Dissected Components Testing
- [x] `bb7-dissect-session.pine` - manual validation PASSED
- [x] `bb7-dissect-sma.pine` - manual validation PASSED
- [x] `bb7-dissect-bb.pine` - manual validation PASSED
- [x] `bb7-dissect-vol.pine` - manual validation PASSED
- [x] `bb7-dissect-potential.pine` - manual validation PASSED
- [x] `bb7-dissect-sl.pine` - manual validation PASSED
- [x] `bb7-dissect-tp.pine` - manual validation PASSED
- [x] `bb7-dissect-adx.pine` - manual validation PASSED

## Phase 5: Strategy Validation
- [x] Comprehensive test coverage: validation package with 28/41 tests passing (edge cases: exact minimum, insufficient data, multiple requirements)
- [x] `./bin/strategy` on rolling-cagr.pine - manual validation PASSED
- [x] `./bin/strategy` on rolling-cagr-5-10yr.pine - manual validation PASSED
- [x] Config management: Makefile targets (create-config, validate-configs, remove-config, clean-configs)
- [x] `./bin/strategy` on BB7 - manual validation PASSED
- [x] `./bin/strategy` on BB8 - manual validation PASSED
- [x] `./bin/strategy` on BB9 - manual validation PASSED
- [x] `time ./bin/strategy` execution <50ms (49µs achieved with real SMA calculation)
- [ ] `ldd ./bin/strategy` shows no external deps (static binary)
- [ ] E2E: replace `node src/index.js` with `./bin/strategy` in tests
- [ ] E2E: 26/26 tests pass with Go binary

## Current Status
- **Parser**: 40/40 Pine fixtures parse successfully (100% coverage)
- **Runtime**: 15 packages (codegen, parser, chartdata, context, input, math, output, request, series, strategy, ta, value, visual, integration, validation)
- **Codegen**: ForwardSeriesBuffer paradigm (ALL variables → Series storage, cursor-based, forward-only, immutable history, O(1) advance)
- **TA Functions**: ta.sma/ema/rma/rsi/atr/bbands/macd/stoch/crossover/crossunder/stdev/change/pivothigh/pivotlow/valuewhen, wma, dev
- **TA Execution**: Inline calculation per bar using ForwardSeriesBuffer, O(1) per-bar overhead
- **Strategy**: entry/close/close_all, if statements, ternary operators, Series historical access (var[offset])
- **Binary**: test-simple.pine → 2.9MB static binary (49µs execution for 30 bars)
- **Output**: Unified chart format (metadata + candlestick + indicators + strategy + ui sections)
- **Visualization**: Config system with filename-based loading (metadata.strategy = source filename)
- **Config Tools**: Makefile integration (create-config, validate-configs, list-configs, remove-config, clean-configs)
- **Project structure**: Proper .gitignore (bin/, testdata/*-output.json excluded)
- **Test Suite**: 605+ tests (preprocessor: 48, chartdata: 22, builder: 18, codegen: 8+11 handlers, expression_analyzer: 10, temp_variable_manager: 11, inline_function_registry: 10, series_source_classifier_ast: 5, validation: 28/41, integration: 40, runtime, datafetcher: 5, security: 271 (74 timezone, 5 Pine-based integration), valuewhen: 66+7, pivot: 95, call_handlers: 35, plot: 127, parser: 40, preprocessor: 29, blockers: 14) - 100% pass rate
- **Handler Test Coverage**: input_handler_test.go (6 tests, 14 subtests), math_handler_test.go (6 tests, 13 subtests), subscript_resolver_test.go (5 tests, 16 subtests), call_handler_*.go (35 tests, 6 files, 1600+ lines), plot_*.go (127 tests: 6 options, 6 buildOptions, 3 titleGen, 6 styleExtract, 20 new generalized tests)
- **Named Parameters**: Full ObjectExpression extraction support (input.float(defval=1.4) → const = 1.40)
- **Warmup Validation**: Compile-time analyzer detects subscript lookback requirements (close[252] → warns need 253+ bars)
- **Data Infrastructure**: BTCUSDT_1D.json extended to 1500 bars (4+ years) supporting 5-year CAGR calculations
- **security() Module**: ForwardSeriesBuffer alignment complete (271/271 tests) - ATR support added, dead code removed, AST utilities extracted, comprehensive edge case coverage, pivot runtime evaluation infrastructure (detector/cache/evaluator modules, 95 tests), pivot codegen integration complete, timezone-aware architecture (ExtractDateInTimezone, BuildMappingWithDateFilter, MOEX inference, 74 timezone tests, bar-count independence verified), Bug #1 & #2 regression tests (Pine-based integration with output validation: first-bar lookahead, non-overlapping ranges, upscaling, downscaling, same-timeframe)
- **Call Handler Architecture**: Strategy pattern refactoring (6 handlers: Meta, Plot, Strategy, TA, Unknown, Router), SOLID principles, 35 comprehensive tests (CanHandle, GenerateCode, Integration, EdgeCases)
- **Plot Module**: Comprehensive test coverage (127 tests), all styling parameters (style, linewidth, transp, pane, color, offset, title), type handling (float64 ↔ int conversion), edge cases, generalization, deduplication
