## Evidence Table: Missing Pine Script v5 Features

| Namespace/Feature        | Official Pine v5 Docs | PineTS Implementation          | Usage in Strategies              | Priority |
| ------------------------ | --------------------- | ------------------------------ | -------------------------------- | -------- |
| `format.percent`         | ✅ const string       | ✅ Context:2870                | rolling-cagr.pine:3              | ✅ DONE  |
| `format.price`           | ✅ const string       | ✅ Context:2871                | None (but standard)              | ✅ DONE  |
| `format.volume`          | ✅ const string       | ✅ Context:2872                | None (but standard)              | ✅ DONE  |
| `format.inherit`         | ✅ const string       | ✅ Context:2873                | None (but standard)              | ✅ DONE  |
| `format.mintick`         | ✅ const string       | ✅ Context:2874                | None (but standard)              | ✅ DONE  |
| `scale.right`            | ✅ const scale_type   | ✅ Context:2877                | rolling-cagr.pine:3              | ✅ DONE  |
| `scale.left`             | ✅ const scale_type   | ✅ Context:2878                | None (but standard)              | ✅ DONE  |
| `scale.none`             | ✅ const scale_type   | ✅ Context:2879                | None (but standard)              | ✅ DONE  |
| `timeframe.ismonthly`    | ✅ simple bool        | ✅ Context:2882+helper         | rolling-cagr.pine:13             | ✅ DONE  |
| `timeframe.isdaily`      | ✅ simple bool        | ✅ Context:2883+helper         | rolling-cagr.pine:13             | ✅ DONE  |
| `timeframe.isweekly`     | ✅ simple bool        | ✅ Context:2884+helper         | rolling-cagr.pine:13             | ✅ DONE  |
| `timeframe.isticks`      | ✅ simple bool        | ✅ Context:2885                | None                             | ✅ DONE  |
| `timeframe.isminutes`    | ✅ simple bool        | ✅ Context:2886+helper         | None                             | ✅ DONE  |
| `timeframe.isseconds`    | ✅ simple bool        | ✅ Context:2887                | None                             | ✅ DONE  |
| `barstate.isfirst`       | ✅ series bool        | ✅ Context:2890                | rolling-cagr.pine:10 (commented) | ✅ DONE  |
| `syminfo.tickerid`       | ✅ simple string      | ✅ Context:2868                | bb-strategy-7:5+ times           | ✅ DONE  |
| `input.source()`         | ✅ function           | ✅ PineTS:1632 + Parser Fix    | rolling-cagr.pine:9              | ✅ DONE  |
| `input.int()`            | ✅ function           | ✅ PineTS + Generic Parser Fix | test-input-int.pine              | ✅ DONE  |
| `input.float()`          | ✅ function           | ✅ PineTS + Generic Parser Fix | test-input-float.pine            | ✅ DONE  |
| `input.bool()`           | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.string()`         | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.color()`          | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.time()`           | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.symbol()`         | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.session()`        | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `input.timeframe()`      | ✅ function           | ✅ PineTS + Generic Parser Fix | (covered by fix)                 | ✅ DONE  |
| `barmerge.lookahead_on`  | ✅ const              | ❌ Not Found                   | bb-strategy-7:3 times            | CRITICAL |
| `barmerge.lookahead_off` | ✅ const              | ❌ Not Found                   | None                             | MEDIUM   |
| `fixnan()`               | ✅ series function    | ❌ Not Found                   | bb-strategy-7:5+ times           | CRITICAL |
| `strategy.*` (60+ items) | ✅ namespace          | ❌ Not Found                   | bb-strategy-7/8/9                | CRITICAL |

---

## ASCII Architecture Diagram

````
┌─────────────────────────────────────────────────────────────────┐
│ PineTS Runtime Injection Architecture                           │
└─────────────────────────────────────────────────────────────────┘

IMPLEMENTATION COMPLETE (rolling-cagr.pine ✅ WORKS):

  Rolling-CAGR.pine
       │
       └──> PineParser (Python) ──> ESTree AST ──> escodegen ──> jsCode
                 │                                                     │
                 │ FIX: Generic input.*(defval=X) → input.*(X, {})    │
                 │ Functions: source, int, float, bool, string,       │
                 │           color, time, symbol, session, timeframe  │
                 │ Commits: b6350ab "Fix input.source defval"         │
                 │          [NEW] "Extend to all input.* functions"   │
                 │                                                     │
                 └─────────────────────────────────────────────────────┘
                                                                      │
                                                                      ▼
                               ┌────────────────────────────────────────┐
                               │ PineTS Context (Modified)              │
                               │ File: PineTS/dist/pinets.dev.es.js     │
                               │                                        │
                               │ constructor() {                        │
                               │   this.syminfo = {                     │
                               │     tickerid, ticker                   │
                               │   };                                   │
                               │   this.format = {                      │
                               │     percent, price, volume,            │
                               │     inherit, mintick                   │
                               │   };                                   │
                               │   this.scale = {                       │
                               │     right, left, none                  │
                               │   };                                   │
                               │   this.timeframe = {                   │
                               │     ismonthly, isdaily, isweekly,      │
                               │     isticks, isminutes, isseconds      │
                               │   };                                   │
                               │   this.barstate = {                    │
                               │     isfirst                            │
                               │   };                                   │
                               │ }                                      │
                               │ _isMonthly(tf) {...}                   │
                               │ _isDaily(tf) {...}                     │
                               │ _isWeekly(tf) {...}                    │
                               │ _isMinutes(tf) {...}                   │
                               │                                        │
                               │ input.source(value, {opts}) {          │
                               │   return Array.isArray(value) ?        │
                               │     value[0] : value;                  │
                               │ }                                      │
                               └────────────────┬───────────────────────┘
                                                │
                                                ▼
                        ┌───────────────────────────────────────────┐
                        │ PineScriptStrategyRunner Wrapper         │
                        │                                           │
                        │ wrappedCode = `(context) => {             │
                        │   const format = context.format;          │
                        │   const scale = context.scale;            │
                        │   const timeframe = context.timeframe;    │
                        │   const barstate = context.barstate;      │
                        │   const input = context.input;            │
                        │   ${jsCode}                               │
                        │ }`                                        │
                        └───────────────┬───────────────────────────┘
                                        │
                                        ▼
                          PineTS.run(wrappedCode) ✅ SUCCESS
                                        │
                                        ▼
                                  Returns plots

                                  Bar 1-12: null (insufficient history)
                                  Bar 13: -11.43% CAGR
                                  Bar 24: -12.42% CAGR

Test Evidence: docker compose exec runner node src/index.js CHMF M 24 strategies/rolling-cagr.pine
Result: 24 candles, 12 null plots (bars 1-12), 12 CAGR values (bars 13-24) - EXIT CODE 0

Additional Test Evidence (Generic Fix Validation):
1. test-input-int.pine: input.int(title="X", defval=20) → input.int(20, {title:"X"}) ✅
2. test-input-float.pine: input.float(defval=2.5, title="Y") → input.float(2.5, {title:"Y"}) ✅
3. Regression test: rolling-cagr.pine still produces 12 CAGR values after generic refactoring ✅

Parser Implementation (services/pine-parser/parser.py:332-341):
```python
INPUT_DEFVAL_FUNCTIONS = {'source', 'int', 'float', 'bool', 'string',
                           'color', 'time', 'symbol', 'session', 'timeframe'}
is_input_with_defval = (isinstance(node.func, Attribute) and
                         isinstance(node.func.value, Name) and
                         node.func.value.id == 'input' and
                         node.func.attr in INPUT_DEFVAL_FUNCTIONS)
````

Coverage: 10 input.\* functions now handle defval parameter correctly (was 1, now 10)
