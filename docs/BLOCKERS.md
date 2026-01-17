| # | Category | Blocker | Verdict | Evidence |
|---|----------|---------|---------|----------|
| **1** | **Codegen** | `ta.rsi()` inline generation | ✅ **VALID** | Error: "ta.rsi inline generation not yet implemented" in codegen/generator.go:2933 |
| **2** | **Codegen** | `bar_index` Series generation | ✅ **VALID** | Compilation error: `undefined: bar_indexSeries`. Codegen doesn't create Series var for bar_index |
| **3** | **Parser** | `while` loops | ✅ **VALID** | Parse error: "binary expression should be used in condition context" |
| **4** | **Parser** | `for` loops execution | ✅ **VALID** | Parses but generates literals only: `sumVal = 50.0` instead of loop logic |
| **5** | **Parser** | `varip` declarations | ✅ **VALID** | Not implemented. No matches in codegen/*.go or grammar.go |
| **6** | **Parser** | `map.new<K,V>()` generics | ✅ **VALID** | Parse error: "unexpected token ," on generic syntax |
| **7** | **Type System** | String variable assignment | ✅ **VALID** | Fails on `ticker = syminfo.tickerid` pattern. Series storage doesn't support strings |
| **8** | **Codegen** | `alert()` function | ✅ **VALID** | Generates TODO comment only. Not implemented |
| **9** | **Codegen** | `alertcondition()` function | ✅ **VALID** | Generates TODO comment only. Not implemented |
| **10** | **Codegen** | String functions (`str.*`) | ✅ **VALID** | `str.tostring()`, `str.tonumber()`, `str.split()` generate TODO comments |
| **11** | **Runtime** | Multi-symbol security() | ✅ **VALID** | Parse✅ Generate✅ Compile✅ Execute❌. Requires data files for multiple symbols |
| **12** | **Parser** | `input()` type parameter syntax | ✅ **VALID** | Parse❌: "unexpected token ',' at line 5:54". Blocks adx-di-strategy.pine |
| **13** | **Codegen** | `lowest()`/`highest()` in conditions | ✅ **FIXED** | Added inline handlers. supertrend.pine: Parse✅ Generate✅ Compile✅ Execute✅ |
| **14** | **Codegen** | TA member expression arguments | ✅ **VALID** | Parse✅ Generate❌: "unsupported member expression in TA call". Blocks keltner-squeeze.pine |
| **15** | **Codegen** | `ta.crossover()` CallExpression requirement | ✅ **FIXED** | Crossover/crossunder now supports variable arguments (Identifier, MemberExpression, etc.) |
| **16** | **Codegen** | `ta.crossover()` incorrect previous bar access | ✅ **FIXED** | IIFE now uses .Get(1) for previous bar access. Crossover detection working correctly. |
| **17** | **Codegen** | Tuple destructuring syntax | ✅ **FIXED** | Universal tuple indicator architecture implemented. MACD, BB, Stoch supported via data-driven registry. |
