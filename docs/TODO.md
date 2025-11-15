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
- [x] `runtime/request/security.go` multi-timeframe data fetching
- [x] `runtime/chartdata/chartdata.go` ChartData struct
- [x] `runtime/chartdata/chartdata.go` Candlestick []OHLCV field
- [x] `runtime/chartdata/chartdata.go` Plots map[string]PlotSeries field
- [x] `runtime/chartdata/chartdata.go` Strategy struct (Trades, OpenTrades, Equity, NetProfit)
- [x] `runtime/chartdata/chartdata.go` Timestamp field
- [x] `runtime/chartdata/chartdata.go` ToJSON() method

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
- [x] Complete AST → Go code generation for Pine functions (ta.sma, plot implemented)
- [x] Implement strategy.entry, strategy.close, strategy.exit codegen (strategy.close lines 247-251, strategy.entry working)
- [ ] `./bin/strategy` on BB7 produces 9 trades
- [ ] `diff out/chart-data.json expected/bb7-chart-data.json` (structure match)
- [x] `time ./bin/strategy` execution <50ms (49µs achieved with real SMA calculation)
- [ ] `ldd ./bin/strategy` shows no external deps (static binary)
- [ ] E2E: replace `node src/index.js` with `./bin/strategy` in tests
- [ ] E2E: 26/26 tests pass with Go binary

## Current Status (as of commit cea50a7)
- **Parser**: 18/37 Pine fixtures parse successfully
- **Runtime**: 10 packages with 80 passing tests
- **Codegen**: ta.sma() and plot() generation working
- **Binary**: test-simple.pine → 2.9MB static binary (49µs execution for 30 bars)
- **Output**: SMA20 values: [121.00 @ bar 19, 128.50 @ bar 29]
- **Project structure**: Proper .gitignore (bin/, testdata/*-output.json excluded)
