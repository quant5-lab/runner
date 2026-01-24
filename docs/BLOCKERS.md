| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **2** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **3** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **4** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **5** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **6** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **7** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **8** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **9** | **Codegen** | `input.float` treated as Series | VALID | Generates `input_floatSeries.GetCurrent()` instead of constant | keltner-squeeze.pine, adx-di-strategy.pine |
| **10** | **Codegen** | Crossover literal integer type | VALID | `ta.crossover(rsi, 30)` generates `float64 > int` mismatch | keltner-squeeze.pine |
| **11** | **Codegen** | Unimplemented TA in ternary | VALID | `ta.alma`, `ta.linreg` etc. not in registry | - |
| **12** | **Codegen** | Boolean function return type | VALID | `myBool() => rsi > 70` returns bool but typed as float64 | - |
| **14** | **Runtime** | `strategy.exit()` execution | VALID | Trade timing misalignment vs TradingView | supertrend.pine |
| **15** | **Runtime** | `bar_index` historical access | VALID | Requires bar_indexSeries generation | test-bar-index-*.pine |
| **16** | **Runtime** | Float modulo operator | VALID | Go codegen needs math.Mod() | test-bar-index-modulo.pine |
