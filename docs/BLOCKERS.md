| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | `input.float` treated as Series | VALID | Generates `input_floatSeries.GetCurrent()` instead of constant | keltner-squeeze.pine, adx-di-strategy.pine |
| **2** | **Codegen** | `bar_index` historical access | VALID | Generates `value.Nz(, 0)` with missing first argument | test-bar-index-*.pine |
| **3** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **4** | **Codegen** | Arrow function self-reference | VALID | `nz(_direction[1])` in init expression - unhandled call in arrow context | zigzag-pa.pine |
| **5** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **6** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **7** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **8** | **Codegen** | Unimplemented TA functions | VALID | `ta.linreg` not in TAFunctionRegistry (only 20 handlers exist) | keltner-squeeze.pine |
| **9** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **10** | **Parser** | For-loop assignment outside var/variable | VALID | Parse error: "unexpected token =" at `gx = for i = 1 to per-1` | emperor-ma.pine |
| **11** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **12** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **13** | **Codegen** | `input.string` not implemented | VALID | Generates `ma1TypeSeries` but never declares variable | adx-di-strategy.pine |
| **14** | **Codegen** | `heikinashi()` function | VALID | Generates TODO comment, function not implemented | utbot-quantnomad.pine |
| **15** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **16** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **17** | **Runtime** | `strategy.exit()` execution | VALID | Trade timing/price is not very accurate vs reference | supertrend.pine |
| **18** | **Codegen** | `security()` package import missing | VALID | Generates `security.BarEvaluator` but doesn't import security package | utbot-quantnomad.pine |
| **19** | **Codegen** | `input.color` not implemented | VALID | No handler in input_handler.go | pivot-reversal.pine |
| **20** | **Codegen** | `input.time` not implemented | VALID | No handler in input_handler.go | - |
| **21** | **Codegen** | `input.timeframe` not implemented | VALID | No handler in input_handler.go | - |
| **22** | **Codegen** | `input.symbol` not implemented | VALID | No handler in input_handler.go | - |
