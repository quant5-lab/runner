| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **2** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **3** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **4** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **5** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **6** | **Parser** | For-loop assignment outside var/variable | VALID | Parse error: "unexpected token =" at `gx = for i = 1 to per-1` | emperor-ma.pine |
| **7** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **8** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **9** | **Codegen** | `heikinashi()` function | RESOLVED | Implemented in staged changes: `call_handler_ticker.go`, `runtime/ticker/` | - |
| **10** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **11** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **12** | **Codegen** | `input.color` not implemented | VALID | No handler in input_handler.go | pivot-reversal.pine |
| **13** | **Codegen** | `input.time` not implemented | VALID | No handler in input_handler.go | - |
| **14** | **Codegen** | `input.timeframe` not implemented | VALID | No handler in input_handler.go | - |
| **15** | **Codegen** | `input.symbol` not implemented | VALID | No handler in input_handler.go | - |
| **16** | **Codegen** | `ObjectExpression` in expression generator | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ObjectExpression" | - |
| **17** | **Codegen** | `ArrowFunctionExpression` inline | VALID | `generateExpression` returns error: "unsupported expression type: *ast.ArrowFunctionExpression" | - |
| **18** | **Codegen** | Arrow TA: ta.atr (1-arg implicit OHLC) | VALID | "unknown function requires exactly 2 arguments, got 1" - needs 1-arg pattern in ArrowTACallSignatureResolver | - |
| **19** | **Codegen** | Arrow TA: ta.pivothigh/ta.pivotlow (3-arg) | VALID | "unknown function requires exactly 2 arguments, got 3" - needs 3-arg pattern in ArrowTACallSignatureResolver | - |
| **20** | **Codegen** | `ticker.*` functions incomplete | VALID | `ticker.modify()` returns ctx.Symbol (no modification), `ticker.new()`/`ticker.inherit()` do string concat only | - |
| **21** | **Codegen** | Non-HA chart types return identity | VALID | Renko, Kagi, LineBreak, PointFig transformers return IdentityTransformer (stubs) | - |
| **22** | **Codegen** | `iff()` function | VALID | Generates TODO comment, function not implemented | utbot-quantnomad.pine |
