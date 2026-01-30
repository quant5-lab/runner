| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | `bar_index` historical access | VALID | Generates `prevBarIndexSeries.Set()` with missing argument | test-bar-index-*.pine |
| **2** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **3** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **4** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **5** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **6** | **Codegen** | Unimplemented TA functions | VALID | `ta.linreg` generates TODO comment, not implemented | keltner-squeeze.pine |
| **7** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **8** | **Parser** | For-loop assignment outside var/variable | VALID | Parse error: "unexpected token =" at `gx = for i = 1 to per-1` | emperor-ma.pine |
| **9** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **10** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **11** | **Codegen** | `heikinashi()` function | VALID | Generates TODO comment, function not implemented | utbot-quantnomad.pine |
| **12** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **13** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **14** | **Codegen** | `input.color` not implemented | VALID | No handler in input_handler.go | pivot-reversal.pine |
| **15** | **Codegen** | `input.time` not implemented | VALID | No handler in input_handler.go | - |
| **16** | **Codegen** | `input.timeframe` not implemented | VALID | No handler in input_handler.go | - |
| **17** | **Codegen** | `input.symbol` not implemented | VALID | No handler in input_handler.go | - |
| **18** | **Codegen** | `ObjectExpression` in expression generator | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ObjectExpression" | - |
| **19** | **Codegen** | `ArrowFunctionExpression` inline | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ArrowFunctionExpression" | - |
| **20** | **Codegen** | Arrow TA: ta.atr (1-arg implicit OHLC) | VALID | "unknown function requires exactly 2 arguments, got 1" - needs 1-arg pattern in ArrowTACallSignatureResolver | - |
| **21** | **Codegen** | Arrow TA: ta.pivothigh/ta.pivotlow (3-arg) | VALID | "unknown function requires exactly 2 arguments, got 3" - needs 3-arg pattern in ArrowTACallSignatureResolver | - |
