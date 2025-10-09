# PineTS Compatibility Analysis v2.0

## Executive Summary

**rolling-cagr.pine:** 2 namespaces need implementation (`timeframe.is*` checks + `barstate.isfirst`)

**bb-strategy dependencies:** 6 namespaces/functions missing
- `strategy.*` (entry/exit/position tracking)
- `syminfo.*` (tickerid for security calls)
- `barmerge.*` (lookahead constants)
- `fixnan()` (data cleaning)
- `time()` (session filtering)
- `timeframe.*` (period access)

---

## Evidence-Based Gap Analysis for Pine Script Execution

**Analysis Date:** 2025-10-09  
**PineTS Version:** Latest from sibling directory  
**Evidence Sources:** 12 searches across official docs, PineTS GitHub, local source code

---

## Executive Summary

```
┌─────────────────────────────────────────────────────────────┐
│  PINE SCRIPT → JAVASCRIPT EXECUTION PIPELINE                │
└─────────────────────────────────────────────────────────────┘

.pine file    →  Python Parser  →  JS Code (ESTree)  →  PineTS.run()
(v4/v5)          (pynescript)       (escodegen)           (Context)
                                                              │
                                                              ▼
                                                    ┌──────────────────┐
                                                    │ Runtime Context  │
                                                    │ + Namespaces     │
                                                    └──────────────────┘
```

---

## Implementation Status by Namespace

### ✅ **Technical Analysis (ta.*)** - FULLY IMPLEMENTED

**Evidence:** https://alaa-eddine.github.io/PineTS/api-coverage/ta.html

| Function | Status | Priority |
|----------|--------|----------|
| ta.ema() | ✅ Tested | HIGH |
| ta.sma() | ✅ Tested | HIGH |
| ta.rsi() | ✅ Tested | HIGH |
| ta.atr() | ✅ Tested | HIGH |
| ta.wma() | ✅ Tested | MEDIUM |
| ta.hma() | ✅ Tested | MEDIUM |
| ta.rma() | ✅ Tested | MEDIUM |
| ta.vwma() | ✅ Tested | MEDIUM |
| ta.change() | ✅ Tested | HIGH |
| ta.mom() | ✅ Tested | MEDIUM |
| ta.roc() | ✅ Tested | MEDIUM |
| ta.dev() | ✅ Tested | MEDIUM |
| ta.variance() | ✅ Tested | LOW |
| ta.highest() | ✅ Tested | HIGH |
| ta.lowest() | ✅ Tested | HIGH |
| ta.median() | ✅ Tested | LOW |
| ta.stdev() | ✅ Tested | HIGH |
| ta.crossover() | ✔️ Implemented | MEDIUM |
| ta.crossunder() | ✔️ Implemented | MEDIUM |
| ta.pivothigh() | ✔️ Implemented | HIGH |
| ta.pivotlow() | ✔️ Implemented | HIGH |
| ta.tema() | ✔️ Implemented | LOW |
| ta.linreg() | ✔️ Implemented | LOW |
| ta.tr() | ✔️ Implemented | HIGH |
| ta.supertrend() | ✔️ Implemented | LOW |

**Used in strategies:** bb-strategy-7-rus.pine (320 lines)

---

### ✅ **Input Functions (input.*)** - FULLY IMPLEMENTED

**Evidence:** https://alaa-eddine.github.io/PineTS/api-coverage/input.html + GitHub source

| Function | Status | Usage Count |
|----------|--------|-------------|
| input.int() | ✅ Tested | 15+ |
| input.float() | ✅ Tested | 20+ |
| input.bool() | ✅ Tested | 10+ |
| input.string() | ✅ Tested | 5+ |
| input.session() | ✅ Tested | 2+ |
| input.timeframe() | ✅ Tested | - |
| input.time() | ✅ Tested | - |
| input.price() | ✅ Tested | - |
| input.source() | ✅ Tested | - |
| input.symbol() | ✅ Tested | - |
| input.text_area() | ✅ Tested | - |
| input.enum() | ✅ Tested | - |
| input.color() | ✅ Tested | - |

**Evidence from parser.py:** Automatic type detection based on first argument value

```python
if isinstance(first_arg_py_value, bool):
    input_type_attr = 'bool'
elif isinstance(first_arg_py_value, float):
    input_type_attr = 'float'
elif isinstance(first_arg_py_value, int):
    input_type_attr = 'int'
```

---

### ✅ **Core Functions** - FULLY IMPLEMENTED

**Evidence:** Core.d.ts TypeScript definitions + dist analysis

| Function | Status | Implementation |
|----------|--------|----------------|
| na() | ✅ Tested | context.core.na() |
| nz() | ✅ Tested | context.core.nz() |
| plot() | ✅ Tested | context.core.plot() |
| indicator() | ✅ Tested | context.core.indicator() |
| color.* | ✅ Tested | context.core.color.* |

**Code Evidence from Core.d.ts:**
```typescript
export declare class Core {
    na(series: any): boolean;
    nz(series: any, replacement?: number): any;
    plot(series: any, title: string, options: PlotOptions): void;
    indicator(title: string, shorttitle?: string, options?: IndicatorOptions): void;
    color: { rgb, new, white, lime, green, red, maroon, black, gray, blue };
}
```

---

### ✅ **Request/Security Functions** - FULLY IMPLEMENTED

**Evidence:** PineRequest.d.ts + dist grep confirmation

| Function | Status | Usage |
|----------|--------|-------|
| request.security() | ✅ Tested | Multi-timeframe |
| security() | ✅ Tested | Legacy alias |

**Used extensively in bb-strategy-7-rus.pine:**
```pine
highUsePivot = security(syminfo.tickerid, "1D", fixnan(pivothigh(leftBars, rightBars)[1]))
sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
```

---

### ⚠️ **Syminfo Namespace** - PARTIALLY IMPLEMENTED

**Evidence:** 
- Official docs page: https://alaa-eddine.github.io/PineTS/api-coverage/syminfo.html
- TypeScript definitions exist: `/PineTS/dist/types/namespaces/Syminfo.d.ts`
- Dist grep: NO matches for `syminfo.` in runtime code

| Variable | Pine Script v5 | PineTS Docs | Runtime Status |
|----------|----------------|-------------|----------------|
| syminfo.ticker() | ✅ VERIFIED | ⚠️ Listed | ❓ Unknown |
| syminfo.prefix() | ✅ VERIFIED | ⚠️ Listed | ❓ Unknown |
| syminfo.tickerid | ✅ VERIFIED | ❌ Missing | ❌ Not Found |
| syminfo.type | ✅ VERIFIED | ❌ Missing | ❌ Not Found |
| syminfo.currency | ✅ VERIFIED | ❌ Missing | ❌ Not Found |
| syminfo.session | ✅ VERIFIED | ❌ Missing | ❌ Not Found |

**Critical Usage in strategies (5+ occurrences):**
```pine
security(syminfo.tickerid, "1D", ...)
```

**Required Implementation:**
- `syminfo.tickerid` returns current symbol ticker ID
- `syminfo.ticker` returns current symbol name
- Context enrichment in PineScriptStrategyRunner wrapper

---

### ❌ **Strategy Namespace** - NOT IMPLEMENTED

**Evidence:** https://alaa-eddine.github.io/PineTS/api-coverage/strategy.html (empty checkboxes)

| Category | Functions | Status |
|----------|-----------|--------|
| Declaration | strategy() | ❌ Stub only |
| Entry | strategy.entry() | ❌ Required |
| Exit | strategy.exit(), strategy.close(), strategy.close_all() | ❌ Required |
| Position Info | strategy.position_avg_price, strategy.position_size | ❌ Required |
| Account Info | strategy.equity, strategy.opentrades, strategy.closedtrades | ❌ Required |
| Constants | strategy.long, strategy.short, strategy.cash, strategy.commission.* | ❌ Required |

**Usage in bb-strategy-7-rus.pine (10+ occurrences):**
```pine
strategy(title="BB Strategy 7 rus", overlay=true, default_qty_type=strategy.cash, ...)
has_active_trade = not na(strategy.position_avg_price)
strategy.entry("BB entry", entry_type, when=entry_condition)
strategy.exit("BB exit", "BB entry", stop=stop_level, limit=smart_take_level)
strategy.close_all()
```

**Current Implementation Status:**
```javascript
// PineScriptStrategyRunner.js - STUB ONLY
const strategy = () => {}; // Does nothing
```

**60+ items require implementation** (functions, variables, constants)

---

### ❌ **Barmerge Constants** - NOT IMPLEMENTED

**Evidence:** Dist grep returns NO matches for `barmerge`

| Constant | Pine Script v5 | Usage | Status |
|----------|----------------|-------|--------|
| barmerge.gaps_on | ✅ VERIFIED | Allow na gaps | ❌ Not Found |
| barmerge.gaps_off | ✅ VERIFIED | Fill gaps | ❌ Not Found |
| barmerge.lookahead_on | ✅ VERIFIED | Future leak | ❌ Not Found |
| barmerge.lookahead_off | ✅ VERIFIED | No future leak | ❌ Not Found |

**Critical Usage in security() calls:**
```pine
open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
```

**Required Implementation:**
```javascript
const barmerge = {
  gaps_on: 'gaps_on',
  gaps_off: 'gaps_off',
  lookahead_on: 'lookahead_on',
  lookahead_off: 'lookahead_off'
};
```

---

### ❌ **Other Missing Functions**

**Evidence:** https://alaa-eddine.github.io/PineTS/api-coverage/others.html (empty checkboxes)

| Function | Category | Status | Usage |
|----------|----------|--------|-------|
| fixnan() | Data cleaning | ❌ Not Implemented | HIGH |
| time() | Date/Time | ❌ Not Implemented | MEDIUM |
| timeframe.period | Runtime info | ❌ Not Implemented | MEDIUM |
| timeframe.ismonthly | Timeframe check | ❌ Not Implemented | LOW |
| timeframe.isdaily | Timeframe check | ❌ Not Implemented | LOW |
| timeframe.isweekly | Timeframe check | ❌ Not Implemented | LOW |
| barstate.isfirst | Bar state | ❌ Not Implemented | LOW |

**Critical Usage:**
```pine
// fixnan() - used in security() calls
highUsePivot = security(syminfo.tickerid, "1D", fixnan(pivothigh(...)))

// time() - session filtering
session_open = na(time(timeframe.period, trading_session)) ? false : true

// timeframe.* - rolling-cagr.pine
interval_multiplier = timeframe.ismonthly ? 12 : timeframe.isdaily ? 252 : ...

// barstate.isfirst - rolling-cagr.pine
varip first_value = barstate.isfirst ? src[0] : src
```

---

## Priority Implementation Matrix

```
CRITICAL (Blocks bb-strategy-7-rus.pine execution):
├─ 1. syminfo.tickerid → Used in 5+ security() calls
├─ 2. barmerge.lookahead_on → Required for security() calls
├─ 3. strategy.entry() → Core trading function
├─ 4. strategy.exit() → Stop loss/take profit management
├─ 5. strategy.position_avg_price → Position tracking
└─ 6. fixnan() → Data cleaning in pivothigh/pivotlow chains

HIGH (Enables full strategy execution):
├─ 7. strategy.close_all() → Session close logic
├─ 8. strategy.position_size → Position management
├─ 9. strategy.long / strategy.short → Direction constants
├─ 10. strategy.cash → Quantity type constant
└─ 11. time() → Session time filtering

MEDIUM (Enhanced functionality):
├─ 12. strategy.equity → Performance tracking
├─ 13. strategy.opentrades → Trade count
├─ 14. timeframe.period → Current timeframe access
└─ 15. barmerge.gaps_off → Data merge control

LOW (Advanced features):
├─ 16. strategy.risk.* functions → Risk management
├─ 17. timeframe.is* checks → Timeframe detection
├─ 18. barstate.* variables → Bar state info
└─ 19. strategy.closedtrades.* → Historical trade analytics
```

---

## Architectural Solution

```
┌────────────────────────────────────────────────────────────┐
│  HYBRID EXECUTION ARCHITECTURE                             │
└────────────────────────────────────────────────────────────┘

Layer 1: PineTS Core (WORKING)
├─ context.data → {close, open, high, low, volume, hl2, hlc3, ohlc4}
├─ context.ta → Technical analysis functions (25+ implemented)
├─ context.core → {plot, na, nz, color, indicator}
├─ context.input → Input functions (13+ types)
└─ context.request → {security} for multi-timeframe

Layer 2: Context Wrapper (REQUIRED)
├─ Syminfo Injection
│   └─ syminfo = {tickerid: symbol, ticker: symbol, ...}
│
├─ Barmerge Constants
│   └─ barmerge = {gaps_on, gaps_off, lookahead_on, lookahead_off}
│
├─ Timeframe Namespace
│   └─ timeframe = {period: currentTF, ismonthly, isdaily, isweekly}
│
├─ Barstate Namespace
│   └─ barstate = {isfirst, islast, isconfirmed}
│
└─ Utility Functions
    └─ fixnan = (src) => src.map(v => isNaN(v) ? lastValid : v)

Layer 3: Strategy State Manager (REQUIRED)
├─ Position Tracking
│   ├─ strategy.position_avg_price
│   ├─ strategy.position_size
│   └─ strategy.position_entry_name
│
├─ Order Management
│   ├─ strategy.entry(id, direction, qty, when)
│   ├─ strategy.exit(id, from, stop, limit)
│   ├─ strategy.close(id)
│   └─ strategy.close_all()
│
├─ Account State
│   ├─ strategy.equity
│   ├─ strategy.opentrades
│   └─ strategy.closedtrades
│
└─ Constants
    ├─ strategy.long / strategy.short
    ├─ strategy.cash / strategy.fixed / strategy.percent_of_equity
    └─ strategy.commission.percent
```

---

## Current Wrapper Implementation

**File:** `src/classes/PineScriptStrategyRunner.js`

```javascript
async executeTranspiledStrategy(jsCode, data) {
  const pineTS = new PineTS(data);
  
  const wrappedCode = `(context) => {
    /* ✅ WORKING: Data series */
    const { close, open, high, low, volume } = context.data;
    
    /* ✅ WORKING: Technical Analysis */
    const ta = context.ta;
    
    /* ✅ WORKING: Request functions */
    const request = context.request;
    const security = request.security.bind(request);
    
    /* ✅ WORKING: Core functions */
    const { plot, color } = context.core;
    
    /* ❌ MISSING: Built-in variables */
    const tickerid = context.tickerId; // Exists but not as syminfo.tickerid
    
    /* ⚠️ STUB ONLY: Strategy functions */
    const indicator = () => {};
    const strategy = () => {};
    const study = indicator;
    
    /* ❌ MISSING: All strategy.* namespace */
    /* ❌ MISSING: All barmerge.* constants */
    /* ❌ MISSING: fixnan() function */
    /* ❌ MISSING: time() function */
    
    ${jsCode}
    
    return { plots: context.plots };
  }`;
  
  await pineTS.run(wrappedCode);
  return { plots: [] };
}
```

---

## Evidence Summary

**Searches Performed:** 12+

1. ✅ PineTS language coverage page
2. ✅ PineTS API coverage index
3. ✅ Strategy namespace documentation
4. ✅ Syminfo namespace documentation
5. ✅ Technical Analysis namespace documentation
6. ✅ Others namespace (fixnan, time, etc.)
7. ✅ PineTS TypeScript type definitions listing
8. ✅ Core.d.ts namespace definitions
9. ✅ Context class implementation (lines 2728-2850)
10. ✅ Dist grep for fixnan/barmerge/syminfo (0 matches)
11. ✅ Dist grep for strategy implementations (0 matches)
12. ✅ Local .pine files for usage patterns (4 strategies)

**Evidence Sources:**
- Official PineTS documentation: https://alaa-eddine.github.io/PineTS/
- PineTS TypeScript definitions: `/PineTS/dist/types/namespaces/`
- PineTS runtime distribution: `/PineTS/dist/pinets.dev.es.js`
- Local strategy files: `/strategies/*.pine` (7 files, 320+ lines)
- Project transpiler: `services/pine-parser/parser.py`
- Wrapper implementation: `src/classes/PineScriptStrategyRunner.js`

---

## Missing Features in Our Strategy Files

**Strategy File Analysis:**
- Total: 1,034 lines across 7 files
- Complex strategies: bb-strategy-7-rus (276), bb-strategy-8-rus (343), bb-strategy-9-rus (367)
- Simple indicators: ema-strategy (13), daily-lines (8), rolling-cagr (24), test (3)

### bb-strategy-7-rus.pine (276 lines) - CRITICAL

**Missing Features:**
```
STRATEGY NAMESPACE (10 occurrences):
├─ Line 2: strategy(title=..., default_qty_type=strategy.cash, commission_type=strategy.commission.percent, ...)
├─ Line 126: has_active_trade = not na(strategy.position_avg_price)
├─ Line 127: position_avg_price_or_close = has_active_trade ? strategy.position_avg_price : close
├─ Line 247: entry_type = sma_bullish ? strategy.long : strategy.short
├─ Line 260: strategy.entry("BB entry", entry_type, when=entry_condition)
├─ Line 261: strategy.exit("BB exit", "BB entry", stop=stop_level, limit=smart_take_level)
└─ Lines 265, 268, 272: strategy.close_all()

SYMINFO NAMESPACE (10 occurrences):
├─ Line 34: highUsePivot = security(syminfo.tickerid, "1D", fixnan(pivothigh(...)))
├─ Line 35: lowUsePivot = security(syminfo.tickerid, "1D", fixnan(pivotlow(...)))
├─ Line 53: sma_1d_20 = security(syminfo.tickerid, 'D', sma(close, 20))
├─ Line 54: sma_1d_50 = security(syminfo.tickerid, 'D', sma(close, 50))
├─ Line 55: sma_1d_200 = security(syminfo.tickerid, 'D', sma(close, 200))
├─ Line 123: open_1d = security(syminfo.tickerid, "D", open, lookahead=barmerge.lookahead_on)
└─ Line 124: atr_1d = security(syminfo.tickerid, "1D", atr(14))

BARMERGE CONSTANTS (1 occurrence):
└─ Line 123: lookahead=barmerge.lookahead_on

FIXNAN FUNCTION (6 occurrences):
├─ Lines 34, 35: fixnan(pivothigh/pivotlow(...))
├─ Lines 98, 99: fixnan(100 * rma(...))
└─ Lines 191, 194: fixnan(sr_h1/sr_l1)

TIME() FUNCTION (2 occurrences):
├─ Line 42: session_open = na(time(timeframe.period, trading_session))
└─ Line 45: is_entry_time = na(time(timeframe.period, entry_time))

TIMEFRAME.PERIOD (2 occurrences):
└─ Lines 42, 45: time(timeframe.period, ...)

INPUT FUNCTIONS (25+ occurrences - ALL WORKING):
└─ Lines 5-30: input.float(), input.bool(), input.session(), input.integer()
```

### bb-strategy-8-rus.pine (343 lines) - CRITICAL

**Missing Features (similar to bb-strategy-7):**
```
STRATEGY NAMESPACE (6 occurrences):
├─ Line 2: strategy(title=..., pyramiding=999, ...)
├─ Line 127: strategy.position_avg_price
├─ Line 305: strategy.long / strategy.short
├─ Line 318: strategy.entry("BB entry", entry_type, when=entry_condition)
└─ Line 339: strategy.close_all()

SYMINFO NAMESPACE (18 occurrences):
└─ Lines 32-33, 51-53, 121-122, 262-263, 280-281, 284-285, 288-291: syminfo.tickerid

BARMERGE CONSTANTS (1 occurrence):
└─ Line 121: barmerge.lookahead_on

FIXNAN FUNCTION (6 occurrences):
└─ Lines 32, 33, 96, 97, 201, 204

TIME() FUNCTION (2 occurrences):
└─ Lines 40, 43: time(timeframe.period, ...)
```

### bb-strategy-9-rus.pine (367 lines) - CRITICAL

**Missing Features (similar pattern):**
```
STRATEGY NAMESPACE (10+ occurrences)
SYMINFO NAMESPACE (20+ occurrences)
BARMERGE CONSTANTS (1 occurrence)
FIXNAN FUNCTION (6 occurrences)
TIME() FUNCTION (3 occurrences including line 351)
```

### rolling-cagr.pine (24 lines) - MEDIUM

**Missing Features:**
```
TIMEFRAME NAMESPACE (3 occurrences):
└─ Line 14: timeframe.ismonthly ? 12 : timeframe.isdaily ? 252 : timeframe.isweekly ? 52

BARSTATE NAMESPACE (1 occurrence):
└─ Line 11: varip first_value = barstate.isfirst ? src[0] : src

INPUT FUNCTIONS (2 occurrences - WORKING):
└─ Lines 6, 10: input.float(), input.source()
```

### daily-lines.pine (8 lines) - LOW

**Missing Features:**
```
SYMINFO NAMESPACE (implicit):
└─ Lines 2-5: security(tickerid, 'D', sma(...))
Note: Uses legacy 'tickerid' variable instead of syminfo.tickerid
```

### ema-strategy.pine (13 lines) - WORKING

**Status:** ✅ No missing features - uses only ta.ema() and plot()

### test.pine (3 lines) - WORKING

**Status:** ✅ Minimal indicator - no advanced features

---

## Missing Feature Summary by Priority

### CRITICAL (Blocks 3 main strategies: 986 lines)

| Feature | Occurrences | Files Affected |
|---------|-------------|----------------|
| **strategy.*** | 26+ | bb-strategy-7/8/9 |
| **syminfo.tickerid** | 48+ | bb-strategy-7/8/9, daily-lines |
| **barmerge.lookahead_on** | 3 | bb-strategy-7/8/9 |
| **fixnan()** | 18 | bb-strategy-7/8/9 |

### HIGH (Enables session filtering: 986 lines)

| Feature | Occurrences | Files Affected |
|---------|-------------|----------------|
| **time()** | 7 | bb-strategy-7/8/9 |
| **timeframe.period** | 7 | bb-strategy-7/8/9 |

### MEDIUM (Enables rolling-cagr: 24 lines)

| Feature | Occurrences | Files Affected |
|---------|-------------|----------------|
| **timeframe.ismonthly** | 1 | rolling-cagr |
| **timeframe.isdaily** | 1 | rolling-cagr |
| **timeframe.isweekly** | 1 | rolling-cagr |
| **barstate.isfirst** | 1 | rolling-cagr |

### Implementation Impact

```
WITHOUT IMPLEMENTATIONS:
├─ bb-strategy-7-rus.pine → BLOCKED (276 lines)
├─ bb-strategy-8-rus.pine → BLOCKED (343 lines)
├─ bb-strategy-9-rus.pine → BLOCKED (367 lines)
├─ rolling-cagr.pine → BLOCKED (24 lines)
├─ daily-lines.pine → PARTIAL (needs tickerid)
├─ ema-strategy.pine → ✅ WORKS (13 lines)
└─ test.pine → ✅ WORKS (3 lines)

WITH CRITICAL IMPLEMENTATIONS:
├─ bb-strategy-7/8/9 → ENABLED (986 lines)
├─ rolling-cagr → ENABLED with timeframe/barstate (24 lines)
├─ daily-lines → ENABLED (8 lines)
└─ Total enabled: 1,018 lines (98.5%)
```

---

## Conclusion

**✅ FULLY WORKING:**
- Technical Analysis namespace (ta.*) - 25+ functions
- Input namespace (input.*) - 13+ types  
- Core functions (na, nz, plot, indicator, color.*)
- Request/Security (request.security, security)
- Data series (close, open, high, low, volume, hl2, hlc3, ohlc4)

**⚠️ PARTIALLY IMPLEMENTED:**
- Syminfo namespace - TypeScript definitions exist, runtime unclear
- Only 2 functions documented: syminfo.ticker(), syminfo.prefix()
- Missing critical syminfo.tickerid used in all security() calls

**❌ NOT IMPLEMENTED:**
- Strategy namespace (60+ functions/variables/constants) - CRITICAL
- Barmerge constants (4 constants) - CRITICAL
- fixnan() function - HIGH priority
- time() function - MEDIUM priority
- timeframe.* namespace - MEDIUM priority
- barstate.* namespace - LOW priority

**REAL-WORLD IMPACT:**
- 986 lines (95.4%) of strategy code BLOCKED without implementations
- 3 main strategies completely non-functional
- Only 2 simple indicators (16 lines) work out-of-the-box

**ARCHITECTURE RECOMMENDATION:**
Implement 3-layer hybrid approach:
1. Use PineTS for ta.*, input.*, core.*, request.* (WORKING)
2. Inject context wrappers for syminfo, barmerge, timeframe, barstate
3. Build strategy state manager for strategy.* namespace (60+ items)

**ESTIMATED IMPLEMENTATION PRIORITY:**
1. Phase 1 (CRITICAL): syminfo.tickerid, barmerge.*, fixnan() - Unlocks 986 lines
2. Phase 2 (HIGH): strategy.entry/exit/close, position tracking - Enables trading
3. Phase 3 (MEDIUM): time(), timeframe.* - Enables session filtering + rolling-cagr
4. Phase 4 (LOW): barstate.*, advanced strategy analytics - Polish features
