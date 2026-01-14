# Security() Performance Analysis

## PERFORMANCE VIOLATIONS

### 1. ARRAY ALLOCATION IN HOT PATH
**Location**: `security/evaluator.go:28`
```go
values := make([]float64, len(secCtx.Data))
```
**Issue**: Allocates full array for EVERY identifier evaluation (close, open, high, low, volume)
**Impact**: O(n) allocation repeated 5x per security() call
**Evidence**: Each `evaluateIdentifier()` creates new slice instead of returning view

### 2. FULL ARRAY COPY IN TA FUNCTIONS
**Location**: `runtime/ta/ta.go:13,34,80`
```go
result := make([]float64, len(source))
```
**Issue**: Every TA function (Sma, Ema, Rma, Rsi) allocates full result array
**Impact**: O(n) allocation + O(n) iteration per TA calculation
**Pattern**: Batch processing instead of ForwardSeriesBuffer index math

### 3. PREFETCH EVALUATES ALL BARS UPFRONT
**Location**: `security/prefetcher.go:55-75`
```go
/* Evaluate all expressions for this symbol+timeframe */
for exprName, exprAST := range req.Expressions {
    values, err := EvaluateExpression(exprAST, secCtx)
```
**Issue**: Calculates ALL security bars before strategy runs
**Impact**: O(warmup + limit) computation even if strategy only needs recent bars
**Waste**: Computes 500 warmup bars that may never be accessed

### 4. NO SERIES REUSE BETWEEN SECURITY CALLS
**Location**: `security/cache.go:12-14`
```go
type CacheEntry struct {
    Context     *context.Context
    Expressions map[string][]float64  /* Pre-computed arrays */
}
```
**Issue**: Each expression stored as standalone array
**Impact**: Cannot share ta.sma(close, 20) between multiple security() calls
**Miss**: No deduplication of identical TA calculations across timeframes

---

## ALIGNMENT GAPS vs ForwardSeriesBuffer

### Series Pattern (Expected)
```
┌─────────────────────────────────────┐
│ ForwardSeriesBuffer                 │
│ - Fixed capacity pre-allocated     │
│ - Index math: buffer[cursor]       │
│ - Zero array mutations             │
│ - O(1) per-bar access              │
└─────────────────────────────────────┘
```

### Security Pattern (Actual)
```
┌─────────────────────────────────────┐
│ Batch Array Processing              │
│ - make() per evaluation             │
│ - Full array loops                  │
│ - Multiple allocations              │
│ - O(n) per-bar cost                 │
└─────────────────────────────────────┘
```

**Architecture Mismatch**: Main strategy uses forward-only Series, security() uses backward batch arrays

---

## DATAFETCHER ARCHITECTURE

### Current: File-Based JSON
**Location**: `datafetcher/file_fetcher.go:29-54`
```go
func (f *FileFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
    filename := fmt.Sprintf("%s/%s_%s.json", f.dataDir, symbol, timeframe)
    data, err := os.ReadFile(filename)
    var bars []context.OHLCV
    json.Unmarshal(data, &bars)
    if limit > 0 && limit < len(bars) {
        bars = bars[len(bars)-limit:]  // Array slice
    }
    return bars, nil
}
```

**Issues**:
- Reads ENTIRE file even if only need recent 100 bars
- Parses ALL JSON even if only accessing last bars
- Array slicing creates new backing array

### No Lazy/Streaming Fetch
- No support for incremental data loading
- No cursor-based pagination
- No pre-allocated buffer reuse across fetches

---

## RUNTIME FLOW VIOLATIONS

### Prefetch Phase (Pre-Bar-Loop)
```
AnalyzeAST() 
  → deduplicateCalls()
    → Fetch() [reads full JSON file]
      → EvaluateExpression() [allocates arrays, computes ALL bars]
        → Cache.Set() [stores pre-computed arrays]
```
**Problem**: Compute all upfront, store in memory

### Per-Bar Phase (Inside Bar Loop)
```
security() call
  → Cache lookup [O(1) map access]
    → Array indexing [values[barIndex]]
      → Series.Set() [stores single value]
```
**Problem**: Cached arrays hold ALL bars, only access 1 per iteration

---

## CONCRETE VIOLATIONS

### V1: evaluateIdentifier() - OHLCV Extraction
```go
func evaluateIdentifier(id *ast.Identifier, secCtx *context.Context) ([]float64, error) {
    values := make([]float64, len(secCtx.Data))  // ⚠️ ALLOCATION
    switch id.Name {
    case "close":
        for i, bar := range secCtx.Data {         // ⚠️ FULL ITERATION
            values[i] = bar.Close
        }
    }
    return values, nil  // ⚠️ RETURN FULL ARRAY
}
```
**Fix**: Return Series interface with lazy index math

### V2: ta.Sma() - Moving Average
```go
func Sma(source []float64, period int) []float64 {
    result := make([]float64, len(source))  // ⚠️ ALLOCATION
    for i := range result {                 // ⚠️ FULL ITERATION
        if i < period-1 {
            result[i] = math.NaN()
            continue
        }
        sum := 0.0
        for j := 0; j < period; j++ {       // ⚠️ NESTED ITERATION
            sum += source[i-j]
        }
        result[i] = sum / float64(period)
    }
    return result  // ⚠️ RETURN FULL ARRAY
}
```
**Fix**: Streaming SMA with circular buffer, O(1) per bar

### V3: Prefetcher - Upfront Evaluation
```go
/* Evaluate all expressions for this symbol+timeframe */
for exprName, exprAST := range req.Expressions {
    values, err := EvaluateExpression(exprAST, secCtx)  // ⚠️ COMPUTE ALL BARS
    err = p.cache.SetExpression(symbol, timeframe, exprName, values)
}
```
**Fix**: Lazy evaluation - compute only when bar accessed

---

## PROPOSAL: ForwardSeriesBuffer Alignment

### Architecture
```
Prefetch Phase:
  - Fetch OHLCV → Store as raw context.Data (no arrays)
  - NO expression evaluation upfront
  - Cache holds contexts, NOT pre-computed values

Runtime Phase (per-bar):
  - security() call → lookup context
  - Evaluate expression for CURRENT bar only
  - Use index math on context.Data[barIndex]
  - Store result in Series
```

### Code Changes

#### 1. Remove Array Allocations
```go
// evaluator.go - BEFORE
values := make([]float64, len(secCtx.Data))
for i, bar := range secCtx.Data {
    values[i] = bar.Close
}

// evaluator.go - AFTER
func evaluateIdentifierAtIndex(id *ast.Identifier, secCtx *context.Context, idx int) (float64, error) {
    if idx >= len(secCtx.Data) { return math.NaN(), nil }
    bar := secCtx.Data[idx]
    switch id.Name {
    case "close": return bar.Close, nil
    case "open":  return bar.Open, nil
    }
}
```

#### 2. Lazy TA Evaluation
```go
// ta/ta.go - BEFORE (batch)
func Sma(source []float64, period int) []float64 {
    result := make([]float64, len(source))
    for i := range result { /* compute all */ }
    return result
}

// ta/ta.go - AFTER (streaming)
type SmaState struct {
    buffer []float64
    cursor int
    sum    float64
}
func (s *SmaState) Next(value float64) float64 {
    // O(1) circular buffer update
}
```

#### 3. Cache Refactor
```go
// cache.go - BEFORE
type CacheEntry struct {
    Context     *context.Context
    Expressions map[string][]float64  // Pre-computed arrays
}

// cache.go - AFTER
type CacheEntry struct {
    Context   *context.Context
    TAStates  map[string]interface{}  // Stateful TA calculators
}
```

---

## EVIDENCE GAPS

### Need Runtime Profiling
- Memory allocation hotspots (pprof)
- CPU time per function (benchmark)
- Cache hit/miss rates

### Need Benchmarks
- 1h→1D downsampling with 500 warmup
- Multiple security() calls with shared expressions
- Memory usage: batch arrays vs Series

### Need Load Testing
- 10+ security() calls in single strategy
- Large datasets (10k+ bars)
- Multiple timeframes (1m, 5m, 15m, 1h, 1D)
