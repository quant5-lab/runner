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
- Target: github.com/markcheno/go-talib (MIT - safe)

## Phase 1: Go Parser + Transpiler (8 weeks)
- [ ] `mkdir -p golang-port/{lexer,parser,codegen,ast}`
- [ ] `go mod init github.com/borisquantlab/pinescript-go`
- [ ] Study `services/pine-parser/parser.py` lines 1-795 AST output
- [ ] Install `github.com/alecthomas/participle/v2` (MIT license)
- [ ] Define PineScript v5 grammar in `parser/grammar.go`
- [ ] Implement lexer using participle.Lexer
- [ ] Implement parser using participle.Parser
- [ ] Map pynescript AST nodes to Go structs in `ast/nodes.go`
- [ ] Implement `codegen/generator.go` AST → Go source
- [ ] Test parse `strategies/test-simple.pine` → AST
- [ ] Compare AST output vs `services/pine-parser/parser.py`
- [ ] Generate Go code matching PineTS execution semantics
- [ ] Test generated code compiles with `go build`

## Phase 2: Go Runtime (12 weeks)
- [ ] `mkdir -p golang-port/runtime/{context,core,math,input,ta,strategy,request}`
- [ ] Install `github.com/markcheno/go-talib` (MIT license)
- [ ] `runtime/context/context.go` OHLCV structs, bar_index, time
- [ ] `runtime/core/core.go` plot(), color, na, nz(), fixnan()
- [ ] `runtime/math/math.go` abs(), max(), min() wrappers
- [ ] `runtime/input/input.go` Int(), Float(), String() with JSON overrides
- [ ] `runtime/ta/sma.go` using go-talib.Sma()
- [ ] `runtime/ta/ema.go` using go-talib.Ema()
- [ ] `runtime/ta/rsi.go` using go-talib.Rsi()
- [ ] `runtime/ta/atr.go` using go-talib.Atr()
- [ ] `runtime/ta/bbands.go` using go-talib.BBands()
- [ ] `runtime/ta/macd.go` using go-talib.Macd()
- [ ] `runtime/ta/stoch.go` using go-talib.Stoch()
- [ ] `runtime/strategy/entry.go` Entry(), Close(), Exit()
- [ ] `runtime/strategy/trades.go` trade tracking slice
- [ ] `runtime/strategy/equity.go` equity calculation
- [ ] `runtime/request/security.go` multi-timeframe data fetching
- [ ] `runtime/output/chart.go` ChartData struct
- [ ] `runtime/output/chart.go` Candlestick []OHLCV field
- [ ] `runtime/output/chart.go` Plots []Plot field
- [ ] `runtime/output/chart.go` Strategy struct (Trades, OpenTrades, Equity, NetProfit)
- [ ] `runtime/output/chart.go` Timestamp time.Time field
- [ ] `runtime/output/json.go` json.Marshal(chartData)

## Phase 3: Binary Template (4 weeks)
- [ ] `mkdir -p golang-port/template`
- [ ] `template/main.go.tmpl` package main + imports
- [ ] `template/main.go.tmpl` flag.String("symbol", "", "")
- [ ] `template/main.go.tmpl` flag.String("timeframe", "", "")
- [ ] `template/main.go.tmpl` flag.Int("bars", 0, "")
- [ ] `template/main.go.tmpl` flag.String("strategy", "", "")
- [ ] `template/main.go.tmpl` context.LoadData() integration
- [ ] `codegen/inject.go` insert generated strategy code into template
- [ ] `cmd/pinescript-go/main.go` CLI entry point
- [ ] `go build -o bin/strategy cmd/pinescript-go/main.go`
- [ ] Test `./bin/strategy -symbol=SBER -timeframe=1h -bars=100 -strategy=test-simple.pine`
- [ ] Write JSON to stdout using json.NewEncoder(os.Stdout)
- [ ] Write JSON to `out/chart-data.json` using os.WriteFile()

## Validation
- [ ] `./bin/strategy` on BB7 produces 9 trades
- [ ] `diff out/chart-data.json expected/bb7-chart-data.json` (structure match)
- [ ] `time ./bin/strategy` execution <50ms (excl. data fetch)
- [ ] `ldd ./bin/strategy` shows no external deps (static binary)
- [ ] E2E: replace `node src/index.js` with `./bin/strategy` in tests
- [ ] E2E: 26/26 tests pass with Go binary
