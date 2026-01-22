| # | Category | Blocker | Verdict | Evidence |
|---|----------|---------|---------|----------|
| **1** | **Codegen** | `ta.rsi()` inline generation | ✅ **FIXED** | Composite indicator architecture with internal series `.Next()` calls. Universal context support (TopLevel + Arrow). Tests: 21/21 integration + 4/4 golden pass (12 fixtures + 4 golden strategy tests). |
| **2** | **Codegen** | `bar_index` Series generation | ✅ **FIXED** | Commit 72f73aa "improve `bar_index`". Tested: codegen + compile ✅ |
| **3** | **Parser** | `while` loops | ✅ **VALID** | Parse error: "binary expression should be used in condition context" |
| **4** | **Parser** | `for` loops execution | ✅ **VALID** | Parses but generates literals only: `sumVal = 50.0` instead of loop logic |
| **5** | **Parser** | `varip` declarations | ✅ **VALID** | Not implemented. No matches in codegen/*.go or grammar.go |
| **6** | **Parser** | `map.new<K,V>()` generics | ✅ **VALID** | Parse error: "unexpected token ," on generic syntax |
| **7** | **Type System** | String variable assignment | ✅ **FIXED** | Commit bbf317f "Add syminfo.tickerid string variable support". Tested: `ticker = syminfo.tickerid` compiles ✅ |
| **8** | **Codegen** | `alert()` function | ✅ **VALID** | Generates TODO comment only. Not implemented |
| **9** | **Codegen** | `alertcondition()` function | ✅ **VALID** | Generates TODO comment only. Not implemented |
| **10** | **Codegen** | String functions (`str.*`) | ✅ **VALID** | `str.tostring()`, `str.tonumber()`, `str.split()` generate TODO comments |
| **11** | **Runtime** | Multi-symbol security() | ✅ **VALID** | Parse✅ Generate✅ Compile✅ Execute❌. Requires data files for multiple symbols |
| **12** | **Parser** | `input()` type parameter syntax | ✅ **VALID** | Parse❌: "unexpected token ',' at line 5:54". Blocks adx-di-strategy.pine |
| **14** | **Codegen** | TA member expression arguments | ✅ **FIXED** | Derived price builtins (hl2, hlc3, ohlc4, hlcc4) supported in TA function arguments. DerivedPriceAccessor generates inline calculations with offset support. Tested: `ta.ema(hl2, 10)`, `ta.ema(hl2[1], 10)`, `ta.ema(hl2 + hlc3, 10)` compile ✅ |
| **15** | **Codegen** | `ta.crossover()` CallExpression requirement | ✅ **FIXED** | Crossover/crossunder now supports variable arguments (Identifier, MemberExpression, etc.) |
| **16** | **Codegen** | `ta.crossover()` incorrect previous bar access | ✅ **FIXED** | IIFE now uses .Get(1) for previous bar access. Crossover detection working correctly. |
| **17** | **Codegen** | Tuple destructuring syntax | ✅ **FIXED** | Universal tuple indicator architecture implemented. MACD, BB, Stoch supported via data-driven registry. |
| **18** | **Codegen** | `ta.crossover()/crossunder()` arbitrary expression support | ✅ **FIXED** | Inline IIFE pattern enables crossover/crossunder with arbitrary PineScript expressions in any context. Recursive stateful indicator detection. Comprehensive test coverage: 35+ behavioral tests (ArgumentTypes, StatefulDetection, WindowFunctionInlining, EdgeCases, LiteralTypes, LogicConditions). |
| **19** | **Codegen** | MemberExpression namespace support | ✅ **FIXED** | Commits 41fed2f, bbf317f. Tested: `syminfo.*`, `strategy.*` compile ✅ |
| **20** | **Codegen** | Array/map functions | ✅ **VALID** | `array.new_float()`, `array.push()`, `array.get()`, `map.*` generate TODO comments |
| **21** | **Codegen** | User-defined functions with `=>` syntax | ✅ **FIXED** | Universal ForwardSeriesBuffer paradigm implemented. ALL variables get Series storage with scope-aware loop modification tracking. Arrow functions compile and execute correctly. |
| **22** | **Codegen** | Binary expression operator precedence in arrow functions | ✅ **FIXED** | BinaryExpressionFormatter with precedence-aware parenthesization. Recursively formats binary expressions without mutual recursion. Test verification: `(a - b) / b * c` generates correct `((a - b) / b * c)` not `((a - b) / (b * c))`. Momentum cascade strategy: 27 trades ✅ |
| **23** | **Codegen** | Go reserved word collision | ✅ **FIXED** | AST transformation preprocessor `IdentifierSanitizer` renames Pine identifiers conflicting with Go reserved words (e.g., `len` → `len_`, `type` → `type_`, `map` → `map_`). Preserves Pine built-ins (`close`, `open`, `high`, `low`, `volume`). Integrated in cmd/pine-gen/main.go before warmup analysis. Tested: `len = input.int(14)` generates `const len_ = 14` and compiles successfully. |
| **24** | **Codegen** | Arrow function return value in binary expressions | ✅ **FIXED** | `extractSeriesExpression` now detects user-defined functions and generates proper calls with ArrowContext. `bband(length, 2)` generates `bband(arrowCtx_bband_1, length, 2)` not `bbandSeries.GetCurrent()`. |
| **25** | **Codegen** | Color constants treated as Series | ✅ **VALID** | Parse✅ Generate✅ Compile❌. `color.lime`, `color.white` generate `colorSeries.Get(0)` instead of color constants. Blocks keltner-squeeze.pine |
| **26** | **Codegen** | input.float treated as Series | ✅ **VALID** | Parse✅ Generate✅ Compile❌. `input.float(1.2, ...)` generates `input_floatSeries.GetCurrent()` instead of constant. Blocks keltner-squeeze.pine |
| **27** | **Codegen** | Crossover literal integer type conversion | ✅ **VALID** | Parse✅ Generate✅ Compile❌. `ta.crossover(rsi, 30)` generates type mismatch: `float64 > int` and `float64 <= int`. Needs float64() cast. Blocks keltner-squeeze.pine |

