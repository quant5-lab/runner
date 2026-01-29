| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | `bar_index` historical access | VALID | Generates `value.Nz(, 0)` with missing first argument | test-bar-index-*.pine |
| **2** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **3** | **Codegen** | Arrow function self-reference | RESOLVED | `nz(_direction[1])` - implemented via ArrowValueFunctionGenerator | zigzag-pa.pine |
| **4** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **5** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **6** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **7** | **Codegen** | Unimplemented TA functions | VALID | `ta.linreg` not in TAFunctionRegistry (only 20 handlers exist) | keltner-squeeze.pine |
| **8** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **9** | **Parser** | For-loop assignment outside var/variable | VALID | Parse error: "unexpected token =" at `gx = for i = 1 to per-1` | emperor-ma.pine |
| **10** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **11** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **12** | **Codegen** | `heikinashi()` function | VALID | Generates TODO comment, function not implemented | utbot-quantnomad.pine |
| **13** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **14** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **15** | **Codegen** | `security()` package import missing | VALID | Generates `security.BarEvaluator` but doesn't import security package | utbot-quantnomad.pine |
| **16** | **Codegen** | `input.color` not implemented | VALID | No handler in input_handler.go | pivot-reversal.pine |
| **17** | **Codegen** | `input.time` not implemented | VALID | No handler in input_handler.go | - |
| **18** | **Codegen** | `input.timeframe` not implemented | VALID | No handler in input_handler.go | - |
| **19** | **Codegen** | `input.symbol` not implemented | VALID | No handler in input_handler.go | - |
| **20** | **Runtime** | Arrow function SMA/STDEV missing bounds check | VALID | `ctx.Data[ctx.BarIndex-j]` crashes when BarIndex < length | keltner-squeeze.pine |
| **21** | **Codegen** | `ObjectExpression` in expression generator | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ObjectExpression" | - |
| **22** | **Codegen** | `ArrowFunctionExpression` inline | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ArrowFunctionExpression" | - |
| **23** | **Codegen** | Arrow accessor literal argument | RESOLVED | `highest(2)` now compiles via TAFunctionSignatureRegistry + ArrowTACallSignatureResolver | zigzag-pa.pine |
| **24** | **Codegen** | Arrow TA: implicit OHLC builtins (ta.tr, ta.atr) | VALID | ta.tr() 0-arg, ta.atr(length) 1-arg use implicit high/low/close, need dedicated IIFE generators | - |
| **25** | **Codegen** | Arrow TA: dual-period functions (ta.pivothigh, ta.pivotlow) | VALID | Dual-period interface implemented, no strategy uses pivot in arrow context | - |
| **26** | **Codegen** | Arrow TA: multi-output tuples (ta.bb, ta.macd, ta.stoch) | VALID | Return tuples, need TupleIndicatorRegistry routing in arrow context | - |


