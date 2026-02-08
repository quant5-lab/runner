| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **2** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **3** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **4** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **5** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **7** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **8** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **9** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **10** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **11** | **Codegen** | `input.*` missing handlers | VALID | `input.color`, `input.time`, `input.timeframe`, `input.symbol` not implemented | pivot-reversal.pine |
| **12** | **Codegen** | `ticker.*` functions incomplete | VALID | `ticker.modify()` returns ctx.Symbol (no modification), `ticker.new()`/`ticker.inherit()` do string concat only | - |
| **13** | **Codegen** | Non-HA chart types return identity | VALID | Renko, Kagi, LineBreak, PointFig transformers return IdentityTransformer (stubs) | - |
| **19** | **Parser** | Tuple destructuring assignment | VALID | Parse error: unexpected token "," at `[t08, s08] = security(...)` | test.pine |
| **20** | **Codegen** | Inline function composition in plot() | FIXED | (A) `plot(nz(...))` ✅ via InlineExpressionScanner pre-hoisting. (B) `plot(ta.sma() + ta.ema())` ✅ via content-based hash. (C) `plot(fixnan(...))` ✅ via cross-bar state hoisting. (D) `plot(nz(fixnan(...)))` ✅ via bottom-up dependency ordering. | - |
| **21** | **Lexer** | Incomplete number literal formats | FIXED | Fixed: leading decimal (`.5`), trailing decimal (`1.`), scientific notation (`6.02e23`). Remaining: underscore separators (`1_000_000`). See `tests/number_format_edge_cases/` | test.pine |
| **22** | **Codegen** | Unprefixed `pow` function | RESOLVED | All unprefixed math functions (pow, asin, abs, sqrt, etc.) compile via registry-based MathHandler. 45 integration tests pass. emperor-ma.pine now blocked by `nz()` as UDF argument in arrow body (#23) | emperor-ma.pine |
| **23** | **Codegen** | Value functions as UDF arguments in arrow body — dual dispatch gap | VALID | `nz()`, `na()`, unprefixed `crossover()`/`crossunder()` fail as UDF args in arrow bodies. Details in `emperor-ma.pine.skip` | emperor-ma.pine |
