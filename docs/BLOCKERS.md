| # | Blocker | Scope | Evidence |
|---|---------|-------|----------|
| **1** | `while` loops | Keyword in lexer/grammar, but no parser rule, no AST node, no codegen | 0 hits for `WhileStatement` across parser/ast/codegen |
| **2** | `varip` declarations | Absent from all layers — lexer, parser, AST, codegen, runtime | 0 hits for `varip` anywhere |
| **3** | `map.new<K,V>()` generics | `<`/`>` lexed as comparison operators, no generic type syntax in grammar | Parse error: `unexpected token ","` on `map.new<string, float>()` |
| **4** | Switch inline case results | `SwitchCase` requires `Indent Body+ Dedent`, no inline `cond => expr` form | `grammar.go` SwitchCase; all tests use multi-line form only |
| **5** | `for...in` iteration | Only `for i = from to end` form exists, no `for element in array` syntax | `grammar.go` ForStatement has no `in` alternative |
| **6** | User-defined types (`type`) | No grammar, AST, or codegen for Pine v5 `type` declarations / UDTs | 0 hits |
| **7** | Methods (`method`) | No grammar, AST, or codegen for Pine v5 `method` declarations | 0 hits |
| **8** | Library `import`/`export` | No grammar, AST, or codegen for Pine library system | 0 hits |
| **9** | ~~`security()` unavailable in arrow functions~~ | ✅ Resolved: detector, call generator, expression walker, hoister bridge, runtime context bridge | `arrow_security_detector.go`, `arrow_security_call_generator.go`, `expression_walker.go` |
| **10** | String functions (`str.*`) | Zero handlers. `str.tostring()`, `str.tonumber()`, `str.format()`, etc. not implemented | 0 hits in codegen |
| **11** | Array data structure (`array.*`) | Zero handlers. `array.new_float()`, `array.push()`, `array.get()`, etc. No runtime array type. | 0 hits in codegen |
| **12** | Map data structure (`map.*`) | Zero handlers. Doubly blocked with #3 (generic syntax). | 0 hits in codegen |
| **13** | Drawing objects (`label.*`, `line.*`, `box.*`, `table.*`) | Zero handlers. No runtime drawing model. | 0 hits in codegen |
| **14** | `alert()`/`alertcondition()` | Zero handlers. No runtime alert model. | 0 hits in codegen |
| **15** | `input.*` missing type handlers | `input.color`, `input.time`, `input.timeframe`, `input.symbol` not implemented | `input_handler.go` switch cases |
| **16** | `ticker.*` semantically incomplete | `ticker.modify()` returns `ctx.Symbol` (no-op), `ticker.new()`/`ticker.inherit()` do string concat only | `call_handler_ticker.go:186` |
| **17** | Non-HA chart type transformers return identity | Renko, Kagi, LineBreak, PointFigure → `IdentityTransformer` (passthrough). Only HeikinAshi has real transform. | `bar_transformer.go:52` |
| **18** | `ArgumentExpressionGenerator` incomplete | Handles 5 of 9+ expression types; missing `Unary`, `Conditional`, `Logical`, `Object` | `argument_expression_generator.go:47` |
