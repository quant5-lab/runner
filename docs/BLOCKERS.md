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
| **9** | Arrow function delegation bypasses `ArrowSeriesAccessResolver` | Non-TA calls delegate to `callRouter.RouteCall` → `extractSeriesExpression()` → unconditional `%sSeries.GetCurrent()`. Resolver correct but bypassed via 50+ call sites. | `arrow_expression_generator_impl.go:125` → `generator.go:2843` |
| **10** | Dual identifier resolution paths | `generateConditionExpression` omits hl2/hlc3/ohlc4/hlcc4. `extractSeriesExpression` handles them via `TryResolveIdentifier`. Two inconsistent paths. | `generator.go:1414` vs `builtin_identifier_registry.go` |
| **11** | `security()` unavailable in arrow functions | Main-body-only construct, not in `sharedTASignatures`, not reachable from arrow call routing | 0 hits in `shared_ta_signatures.go` |
| **12** | ~~Input type inference fails for named-arg-only calls~~ | ~~`inferInputTypeFromLiteral` checks `Arguments[0]` as Literal/Identifier only. Named-arg calls have ObjectExpression → fails.~~ | FIXED: `input_type_resolver.go` |
| **13** | String functions (`str.*`) | Zero handlers. `str.tostring()`, `str.tonumber()`, `str.format()`, etc. not implemented | 0 hits in codegen |
| **14** | Array data structure (`array.*`) | Zero handlers. `array.new_float()`, `array.push()`, `array.get()`, etc. No runtime array type. | 0 hits in codegen |
| **15** | Map data structure (`map.*`) | Zero handlers. Doubly blocked with #3 (generic syntax). | 0 hits in codegen |
| **16** | Drawing objects (`label.*`, `line.*`, `box.*`, `table.*`) | Zero handlers. No runtime drawing model. | 0 hits in codegen |
| **17** | `alert()`/`alertcondition()` | Zero handlers. No runtime alert model. | 0 hits in codegen |
| **18** | `input.*` missing type handlers | `input.color`, `input.time`, `input.timeframe`, `input.symbol` not implemented | `input_handler.go` switch cases |
| **19** | `ticker.*` semantically incomplete | `ticker.modify()` returns `ctx.Symbol` (no-op), `ticker.new()`/`ticker.inherit()` do string concat only | `call_handler_ticker.go:186` |
| **20** | Non-HA chart type transformers return identity | Renko, Kagi, LineBreak, PointFigure → `IdentityTransformer` (passthrough). Only HeikinAshi has real transform. | `bar_transformer.go:52` |
