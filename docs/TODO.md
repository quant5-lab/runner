# Golang Port PoC

## Current Performance (Measured)
- Total: 2792ms
- Python parser: 2108ms (75%)
- Runtime execution: 18ms (0.6%)
- Data fetch: 662ms (23.7%)

## Target Performance
- Total: <50ms (excl. data fetch)
- Go parser: 5-10ms
- Go runtime: <10ms

## License Safety
- Current: pynescript v0.2.0 (LGPL 3.0 - VIRAL if embedded)
- Current: escodegen v2.1.0 (BSD-2-Clause - safe)
- Current: pinets local (unknown - assume MIT)
- Target: Go stdlib only (BSD-3-Clause - safe)
- Target: github.com/alecthomas/participle/v2 (MIT - safe)
- Target: Pure Go TA implementation (no external dependencies)

## Phase 1: Go Parser + Transpiler (8 weeks)
- [x] `mkdir -p golang-port/{lexer,parser,codegen,ast}`
- [x] `go mod init github.com/borisquantlab/pinescript-go`
- [x] Study `services/pine-parser/parser.py` lines 1-795 AST output
- [x] Install `github.com/alecthomas/participle/v2` (MIT license)
- [x] Define PineScript v5 grammar in `parser/grammar.go`
- [x] Implement lexer using participle.Lexer
- [x] Implement parser using participle.Parser
- [x] Map pynescript AST nodes to Go structs in `ast/nodes.go`
- [x] Implement `codegen/generator.go` AST → Go source
- [x] Test parse `strategies/test-simple.pine` → AST
- [x] Compare AST output vs `services/pine-parser/parser.py`
- [x] Generate Go code matching PineTS execution semantics
- [x] Test generated code compiles with `go build`

## Phase 2: Go Runtime (12 weeks)
- [x] `mkdir -p golang-port/runtime/{context,core,math,input,ta,strategy,request}`
- [x] Pure Go TA implementation (no external library - PineTS compatible)
- [x] `runtime/context/context.go` OHLCV structs, bar_index, time
- [x] `runtime/value/na.go` na, nz(), fixnan() (SOLID: separated from visual)
- [x] `runtime/visual/color.go` color constants as hex strings (PineTS compatible)
- [x] `runtime/output/plot.go` PlotCollector interface (SOLID: testable, mockable)
- [x] `runtime/math/math.go` abs(), max(), min(), pow(), sqrt(), floor(), ceil(), round(), log(), exp(), sum(), avg()
- [x] `runtime/input/input.go` Int(), Float(), String(), Bool() with title-based overrides
- [x] `runtime/ta/ta.go` Sma, Ema, Rma with NaN warmup period
- [x] `runtime/ta/ta.go` Rsi using Rma smoothing (PineTS semantics)
- [x] `runtime/ta/ta.go` Tr, Atr with correct high-low-close calculation
- [x] `runtime/ta/ta.go` BBands (upper, middle, lower bands)
- [x] `runtime/ta/ta.go` Macd (macd, signal, histogram with NaN-aware EMA)
- [x] `runtime/ta/ta.go` Stoch (%K, %D oscillator)
- [x] `runtime/strategy/entry.go` Entry(), Close(), Exit()
- [x] `runtime/strategy/trades.go` trade tracking slice
- [x] `runtime/strategy/equity.go` equity calculation
- [x] `runtime/chartdata/chartdata.go` ChartData struct
- [x] `runtime/chartdata/chartdata.go` Candlestick []OHLCV field
- [x] `runtime/chartdata/chartdata.go` Plots map[string]PlotSeries field
- [x] `runtime/chartdata/chartdata.go` Strategy struct (Trades, OpenTrades, Equity, NetProfit)
- [x] `runtime/chartdata/chartdata.go` Timestamp field
- [x] `runtime/chartdata/chartdata.go` ToJSON() method

## Phase 2.5: request.security() Module - Series Alignment (6 weeks)

### Baseline (Working - Array-Based)
- [x] `security/analyzer.go` AST scanner for security() calls (5/5 tests)
- [x] `datafetcher/file_fetcher.go` Local JSON reader (5/5 tests)
- [x] `security/cache.go` Context + expression array storage (8/8 tests)
- [x] `security/evaluator.go` Batch array evaluation (6/6 tests)
- [x] `security/prefetcher.go` Upfront expression computation (3/3 tests)
- [x] `codegen/security_inject.go` Array lookup code generation (4/4 tests)
- [x] BB pattern tests: 7/7 PASS (close, sma, ema, open with identifiers + named TA)

### Phase 2.5.1: Context-Only Cache (O(1) Per-Bar Access)
- [x] `security/cache.go` Remove Expressions map[string][]float64, keep Context only
- [x] `security/evaluator.go` Remove EvaluateExpression batch processing
- [x] `security/prefetcher.go` Remove expression evaluation loop, fetch contexts only
- [x] `codegen/generator.go` Replace array lookup with secCtx.Data[barIndex].Close direct access
- [x] Test: 7/7 BB pattern tests PASS (baseline preserved)
- [x] Benchmark: evaluateIdentifier 40KB → 0B allocation proof

### Phase 2.5.2: Inline TA Series States (O(1) Streaming)
- [x] `codegen/generator.go` generateInlineTA use circular buffer for SMA/EMA warmup
- [x] Replace ctx.Data backward loops with forward-only sliding window
- [x] Test: 7/7 BB pattern tests PASS (TA calculations correct)
- [x] Benchmark: ta.Sma 82KB → 0B, O(N) → O(1) proof
- [ ] Streaming state optimization: 8/13 TA functions support O(1) per-bar (SMA, EMA, RMA, RSI, ATR, TR, Change, MACD)
- [ ] Priority: SMA circular buffer (50-200x speedup for period>100), RSI/ATR composition
- [ ] Keep O(period): Stdev, BBands, Stoch, Pivothigh, Pivotlow (inherent window scan)

### Phase 2.5.3: Complex Expressions (Parser Enhancement)
- [x] `codegen/generator.go` BinaryExpression in security context (generateBinaryExpressionInSecurityContext)
- [x] `codegen/generator.go` Identifier in security context (ctx.Data[ctx.BarIndex].Close)
- [x] Test suite: 5/5 comprehensive security codegen tests PASS (BinaryExpression, ConditionalExpression, ATR, STDEV, ContextIsolation)
- [x] Baseline: 7/7 BB pattern tests PASS (regression safety validated)
- [x] `parser/grammar.go` Argument → TernaryExpr (supports all operators in function arguments)
- [x] `parser/converter.go` String literal quote trimming fixed (both " and ')
- [x] `parser/grammar.go` Parenthesized expressions added to Factor (supports complex precedence)
- [x] `preprocessor/*.go` Updated visitor and ta_namespace transformer for TernaryExpr arguments
- [x] Test: ta.sma(close,20) + ta.ema(close,10) parses successfully ✅
- [x] Test: (high - low) / close * 100 parses successfully ✅
- [x] Test: 10/10 security integration test suites PASS (28+ test cases, 100% success) ✅

### Integration & Validation
- [x] Integrate InjectSecurityCode into builder pipeline (complete)
- [x] All security() test suites PASS (10 suites, 28+ cases)
- [ ] E2E: daily-lines.pine with BTCUSDT_1h.json + BTCUSDT_1D.json
- [ ] Verify: SMA values correct daily averages, not zeros
- [ ] Test: Downsampling (1h → 1D), Same timeframe (1D → 1D), Upsampling error (1D → 1h)

## Phase 3: Binary Template (4 weeks)
- [x] `mkdir -p golang-port/template`
- [x] `template/main.go.tmpl` package main + imports
- [x] `template/main.go.tmpl` flag.String("symbol", "", "")
- [x] `template/main.go.tmpl` flag.String("timeframe", "", "")
- [x] `template/main.go.tmpl` flag.String("data", "", "")
- [x] `template/main.go.tmpl` flag.String("output", "", "")
- [x] `template/main.go.tmpl` context.LoadData() integration
- [x] `codegen/inject.go` insert generated strategy code into template
- [x] `codegen/generator.go` AST → Go code generation (placeholder)
- [x] `cmd/pinescript-builder/main.go` CLI entry point
- [x] `go build -o bin/pinescript-builder cmd/pinescript-builder/main.go`
- [x] Test `./bin/pinescript-builder -input test-simple.pine -output bin/strategy`
- [x] Test `go build -o bin/test-simple-runner /tmp/pine_strategy_temp.go`
- [x] Test `./bin/test-simple-runner -symbol TEST -data sample-bars.json -output output.json`
- [x] Verify JSON output with candlestick/plots/strategy/timestamp
- [x] Execution <50ms (24µs for 30 bars with placeholder strategy)

## Validation
- [x] Complete AST → Go code generation for Pine functions (ta.sma/ema/rsi/atr/bbands/macd/stoch, plot, if/ternary, Series[offset])
- [x] Implement strategy.entry, strategy.close, strategy.exit codegen (strategy.close lines 247-251, strategy.entry working)
- [x] `./bin/strategy` on daily-lines-simple.pine validates basic features
- [x] `./bin/strategy` on daily-lines.pine validates advanced features
- [ ] `./bin/strategy` on rolling-cagr.pine validates calculation accuracy
- [ ] `./bin/strategy` on rolling-cagr-5-10yr.pine validates long-term calculations
- [ ] `./bin/strategy` on BB7 produces 9 trades
- [ ] `./bin/strategy` on BB8 produces expected trades
- [ ] `./bin/strategy` on BB9 produces expected trades
- [ ] `diff out/chart-data.json expected/bb7-chart-data.json` (structure match)
- [x] `time ./bin/strategy` execution <50ms (49µs achieved with real SMA calculation)
- [ ] `ldd ./bin/strategy` shows no external deps (static binary)
- [ ] E2E: replace `node src/index.js` with `./bin/strategy` in tests
- [ ] E2E: 26/26 tests pass with Go binary

## Current Status
- **Parser**: 18/37 Pine fixtures parse successfully
- **Runtime**: 14 packages (codegen, parser, chartdata, context, input, math, output, request, series, strategy, ta, value, visual, integration)
- **Codegen**: ForwardSeriesBuffer paradigm (ALL variables → Series storage, cursor-based, forward-only, immutable history, O(1) advance)
- **TA Functions**: ta.sma/ema/rma/rsi/atr/bbands/macd/stoch/crossover/crossunder/stdev/change/pivothigh/pivotlow, valuewhen (runtime library pre-calculation)
- **Strategy**: entry/close/close_all, if statements, ternary operators, Series historical access (var[offset])
- **Binary**: test-simple.pine → 2.9MB static binary (49µs execution for 30 bars)
- **Output**: Unified chart format (metadata + candlestick + indicators + strategy + ui sections)
- **Documentation**: UNIFIED_CHART_FORMAT.md, STRATEGY_RUNTIME_ARCHITECTURE.md, MANUAL_TESTING.md, data-fetching.md
- **Project structure**: Proper .gitignore (bin/, testdata/*-output.json excluded)
- **Test Suite**: 101+ tests (preprocessor: 21, chartdata: 16, builder: 18, codegen: 8, integration, runtime, datafetcher: 5, security: 27, security_inject: 4)
- **security() Module**: Complete disk-based prefetch architecture (31/31 tests) - analyzer, file_fetcher, cache, evaluator, prefetcher, codegen injection - ready for builder integration
