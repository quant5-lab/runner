# PineScript Support Blockers

**Evidence-based list of ALL blockers preventing 100% arbitrary PineScript support**

## CODEGEN LIMITATIONS

### Inline Call Support
- ❌ `request.security()` inline in conditionals
  - File: `bb-strategy-8-rus.pine:284,285,351`
  - Pattern: `security(...) ? a : b` or `cond and security(...)`
  - Error: `codegen/inline_condition_handler_registry.go:37` - "unsupported inline function in condition"
  - Fix: Add `SecurityInlineHandler` to registry

- ❌ `ta.rsi()` inline generation not implemented
  - File: `codegen/generator.go:2933`
  - Error: "ta.rsi inline generation not yet implemented"
  - Impact: RSI cannot be used in inline expressions

### Function Support
- ❌ `strategy.exit()` not implemented
  - Status: Grep shows NO implementation
  - Impact: Stop-loss/take-profit exit logic unavailable

## PARSER LIMITATIONS

### Language Constructs
- ❌ Single-line arrow functions
  - Pattern: `func(x) => expression`
  - Status: 39/40 fixtures parse (97.5% success)
  - Workaround: Multi-line arrow functions work

- ❌ BB9 syntax error at line 342
  - File: `bb-strategy-9-rus.pine:342`
  - Error: "unexpected token ''"
  - Needs: Investigation of failing syntax

- ⚠️ `for` loops
  - Status: Parse✅ Generate✅ Compile✅ (literals only, not actual loops)
  - Evidence: `test-for-loop.pine` generates `sumVal := 0.0; sumVal = 50.0`
  - Impact: Loop logic not executed

- ❌ `while` loops
  - Status: Parse❌
  - Evidence: `test-while-loop.pine` → "binary expression should be used in condition context"
  - Impact: Cannot use while loops

- ✅ `var` declarations
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-var-decl.pine` successful

- ⚠️ `varip` declarations (UNTESTED)
  - Status: No evidence in grammar
  - Impact: Intra-bar mutable variables

## TYPE SYSTEM

### String Support
- ❌ String variable assignment not supported
  - File: `tests/test-integration/syminfo_tickerid_test.go:88,117`
  - Pattern: `ticker = syminfo.tickerid`
  - Impact: Variables holding string values fail

## RUNTIME DATA

### Security Context
- ❌ Multi-symbol `security()` calls
  - File: `test-security-multi-symbol.pine.skip`
  - Status: Parse✅ Generate✅ Compile✅ Execute❌
  - Issue: Requires OHLCV data for multiple symbols

- ❌ `syminfo.tickerid` dynamic file mapping
  - File: `test-security-same-tf.pine.skip`
  - Status: Parse✅ Generate✅ Compile✅ Execute❌
  - Issue: Data file mapping not implemented

## BUILT-IN FUNCTIONS

### Drawing Functions
- ✅ `label.new()`, `label.set_text()`
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-label.pine` successful

- ⚠️ `line.new()`, `line.set_*()`, `line.delete()` (UNTESTED)
- ⚠️ `box.new()`, `box.set_*()`, `box.delete()` (UNTESTED)
- ⚠️ `table.new()`, `table.set_*()`, `table.delete()` (UNTESTED)

### Alert Functions (CODEGEN TODO)
- ❌ `alert()` - Parse✅ Generate TODO comment
  - Evidence: `test-alert.pine` → `// alert() - TODO: implement`
- ❌ `alertcondition()` - Parse✅ Generate TODO comment
  - Evidence: `test-alert.pine` → `// alertcondition() - TODO: implement`

### Visual Functions
- ✅ `fill()`, `bgcolor()`, `hline()`
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-visual-funcs.pine` successful

### Array Functions
- ✅ `array.new_float()`, `array.push()`, `array.get()`, `array.size()`
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-array.pine` successful

### Map Functions
- ❌ `map.new<K,V>()` with generics
  - Status: Parse❌
  - Evidence: `test-map.pine` → "unexpected token ,"
  - Impact: Cannot use map collections with generic types

- ⚠️ `matrix.new_*()`, matrix operations (UNTESTED)

### String Functions (CODEGEN TODO)
- ❌ `str.tostring()` - Parse✅ Generate TODO comment
- ❌ `str.tonumber()` - Parse✅ Generate TODO comment  
- ❌ `str.split()` - Parse✅ Generate TODO comment
  - Evidence: `test-string-funcs.pine` → `// str.* - TODO: implement`

### Color Functions
- ✅ Hex colors work: `#ff0000`
- ✅ `color.red`, `color.blue`, etc. (constants)
- ✅ `color.rgb()`, `color.new()`
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-color-funcs.pine` successful

## STRATEGY FUNCTIONS

### Implemented
- ✅ `strategy.entry()`
- ✅ `strategy.close()`
- ✅ `strategy.close_all()`

### Not Implemented
- ✅ `strategy.exit()` 
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-strategy-exit.pine` successful
- ⚠️ `strategy.order()` (UNTESTED)
- ⚠️ `strategy.cancel()` (UNTESTED)
- ⚠️ `strategy.cancel_all()` (UNTESTED)

## TA FUNCTIONS

### Implemented (13)
- ✅ Atr, BBands, Change, Ema, Macd, Pivothigh, Pivotlow
- ✅ Rma, Rsi, Sma, Stdev, Stoch, Tr

### Not Implemented (Common ones)
- ✅ CCI (Commodity Channel Index)
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-ta-missing.pine` successful
- ✅ WMA (Weighted Moving Average)
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-ta-missing.pine` successful
- ✅ VWAP (Volume Weighted Average Price)
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-ta-missing.pine` successful
- ⚠️ OBV (On Balance Volume) (UNTESTED)
- ⚠️ SAR (Parabolic SAR) (UNTESTED)
- ⚠️ ADX (Average Directional Index) - **USER-DEFINED WORKS**
- ⚠️ HMA (Hull Moving Average) (UNTESTED)
- ⚠️ Supertrend (UNTESTED)
- ⚠️ Ichimoku components (UNTESTED)

## OPERATORS

### Supported
- ✅ Arithmetic: `+`, `-`, `*`, `/`
- ✅ Comparison: `>`, `<`, `>=`, `<=`, `==`, `!=`
- ✅ Logical: `and`, `or`, `not`
- ✅ Ternary: `? :`
- ✅ Assignment: `=`, `:=`
- ✅ Modulo: `%`
  - Status: Parse✅ Generate✅ Compile✅
  - Evidence: `test-operators.pine` successful

### Not Supported
- ❌ Bitwise: `&`, `|`, `^`, `~`, `<<`, `>>`
  - Status: Parse❌
  - Evidence: `test-operators.pine` → "lexer: invalid input text"
  - Impact: Cannot use bitwise operations

- ⚠️ Null coalescing: `??` (UNTESTED)

## LEGEND
- ✅ Verified working (evidence in code/tests)
- ❌ Verified NOT working (documented blocker)
- ⚠️ UNVERIFIED (no evidence either way)

## SUMMARY
- **Documented Blockers:** 13
  - Codegen: security inline, RSI inline
  - Parser: arrow functions, BB9 line 342, while loops, for loops (literals only), map generics, bitwise operators
  - Codegen TODO: alert, alertcondition, str.tostring, str.tonumber, str.split
  - Type: string variables
  - Runtime: multi-symbol security, syminfo.tickerid mapping
- **Verified Working:** 25+ features
  - var declarations, labels, arrays, strategy.exit, colors, visuals, TA (CCI/WMA/VWAP), operators (arithmetic/logical/modulo)
- **Untested:** 10+ features
  - varip, line/box/table drawing, matrix functions, strategy.order/cancel, OBV/SAR/HMA/Supertrend/Ichimoku, null coalescing

**CONCLUSION:** 13 blocking issues prevent 100% arbitrary PineScript support. Most core features work.
