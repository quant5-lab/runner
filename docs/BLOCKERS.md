| # | Category | Blocker | Status | Evidence | Blocks |
|---|----------|---------|--------|----------|--------|
| **1** | **Codegen** | String functions (`str.*`) | VALID | `str.tostring()`, `str.tonumber()`, `str.split()` not implemented | - |
| **2** | **Parser** | `while` loops | VALID | Not implemented in grammar/codegen | - |
| **3** | **Parser** | `varip` declarations | VALID | Not implemented. No matches in codegen/*.go | - |
| **4** | **Parser** | `map.new<K,V>()` generics | VALID | Parse error: "unexpected token ," on generic syntax | - |
| **5** | **Parser** | `switch` expression | VALID | Parse error: "unexpected token =>" at line 31 | pivot-reversal.pine |
| **6** | **Parser** | For-loop assignment outside var/variable | ✅ FIXED | Parse error: "unexpected token =" at `gx = for i = 1 to per-1` resolved. Added ForExpr/IfExpr to grammar as expression alternatives. ForStatement/IfStatement implement Expression interface. Codegen generates Go IIFE pattern: `result := (func() float64 { var __result float64; for i := ...; return __result }())` | emperor-ma.pine |
| **7** | **Codegen** | Array/map functions | VALID | `array.new_float()`, `array.push()`, `array.get()`, `map.*` not implemented | pivot-reversal.pine |
| **8** | **Codegen** | Drawing objects (`line.*`, `label.*`) | VALID | Not implemented | pivot-reversal.pine |
| **9** | **Codegen** | `alert()` function | VALID | Not implemented | - |
| **10** | **Codegen** | `alertcondition()` function | VALID | Not implemented | - |
| **11** | **Codegen** | `input.*` missing handlers | VALID | `input.color`, `input.time`, `input.timeframe`, `input.symbol` not implemented | pivot-reversal.pine |
| **12** | **Codegen** | `ticker.*` functions incomplete | VALID | `ticker.modify()` returns ctx.Symbol (no modification), `ticker.new()`/`ticker.inherit()` do string concat only | - |
| **13** | **Codegen** | Non-HA chart types return identity | VALID | Renko, Kagi, LineBreak, PointFig transformers return IdentityTransformer (stubs) | - |
| **14** | **Codegen** | Color constants (`blue`, `silver`, `green`, `red`, etc.) | ✅ FIXED | All 17 TradingView color constants (bare identifiers and color.* member expressions) now resolve to hex strings | test.pine |
| **15** | **Codegen** | Boolean comparison in ternary generates float64 vs bool | ✅ FIXED | mismatched types float64 and untyped bool | test.pine |
| **16** | **Codegen** | Incomplete math namespace functions | VALID | Missing: sin, cos, tan, asin, acos, atan, log10, avg, sign, random, todegrees, toradians, round_to_mintick. Bug: `math.sum` incorrectly implemented as SMA (divides by length). See `tests/math_function_edge_cases/` | test.pine |
| **17** | **Codegen** | Dynamic period expressions in TA functions | VALID | **TESTED EXHAUSTIVELY.** ALL non-literal periods fail with "period must be literal": variables (`varPeriod`), binary expressions (`7*2`), call expressions (`getPeriod()`, `input.int()`). Affects: `ta.sma`, `ta.ema`, `ta.rsi`, `ta.atr`, `ta.stdev`, `ta.highest`, `ta.lowest`. **EXCEPTIONS:** `ta.supertrend` and `ta.bb` multiplier accept dynamic values. Source args work. Only integer literals supported. See `tests/dynamic-period-edge-cases/` | test.pine |
| **18** | **Parser** | Multiple variable declaration with comma | VALID | Parse error: unexpected token "," at `src = close,` | test.pine |
| **19** | **Parser** | Tuple destructuring assignment | VALID | Parse error: unexpected token "," at `[t08, s08] = security(...)` | test.pine |
| **20** | **Codegen** | Inline function composition in plot() | VALID | **TESTED EXHAUSTIVELY.** Three distinct failure modes: (A) `plot(nz(...))`, `plot(fixnan(...))` fail with "unsupported inline function in plot" - works when assigned first; (B) `plot(ta.sma() + ta.ema())` - binary expressions with multiple TA calls fail with "unsupported inline function in condition"; (C) `ta.sma(close, input.int())` - dynamic period fails with "period must be literal". **WORKING:** nested TA (`plot(ta.sma(ta.ema()))`), nested math (`plot(math.max(math.avg()))`), ternary in plot, single TA in plot, custom func composition. See `tests/composition-edge-cases/` | - |
| **21** | **Lexer** | Incomplete number literal formats | FIXED | Fixed: leading decimal (`.5`), trailing decimal (`1.`), scientific notation (`6.02e23`). Remaining: underscore separators (`1_000_000`). See `tests/number_format_edge_cases/` | test.pine |
