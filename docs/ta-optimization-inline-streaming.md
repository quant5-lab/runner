# TA Functions: Streaming State Optimization Analysis

## Summary

**Total TA Functions**: 13
**Streamable to O(1)**: 8 (62%)
**Require O(period) window scan**: 5 (38%)

---

## ✅ STREAMABLE TO O(1) (8 functions)

### 1. **SMA (Simple Moving Average)**
**Current**: O(period) - sum last N values each bar
**Streaming**: O(1) - circular buffer with running sum
```go
sum = sum - buffer[cursor] + newValue
buffer[cursor] = newValue
cursor = (cursor + 1) % period
```

### 2. **EMA (Exponential Moving Average)**
**Current**: O(period) warmup + O(1) after
**Streaming**: O(1) - recursive formula
```go
ema = alpha * newValue + (1 - alpha) * prevEma
```
**Already optimal after warmup**

### 3. **RMA (Relative Moving Average)**
**Current**: O(period) warmup + O(1) after
**Streaming**: O(1) - Wilder's smoothing
```go
rma = (prevRma * (period-1) + newValue) / period
```
**Already optimal after warmup**

### 4. **RSI (Relative Strength Index)**
**Current**: O(period) - uses RMA internally
**Streaming**: O(1) - RMA of gains/losses
```go
avgGain = rmaGain.Next(gain)
avgLoss = rmaLoss.Next(loss)
rsi = 100 - 100/(1 + avgGain/avgLoss)
```

### 5. **ATR (Average True Range)**
**Current**: O(period) - RMA of TR
**Streaming**: O(1) - TR is O(1), RMA is O(1)
```go
tr = max(high-low, abs(high-prevClose), abs(low-prevClose))
atr = rma.Next(tr)
```

### 6. **TR (True Range)**
**Current**: O(1) - already optimal
**Streaming**: O(1) - no state needed
```go
tr = max(high-low, abs(high-prevClose), abs(low-prevClose))
```
**No optimization needed - inherently O(1)**

### 7. **Change**
**Current**: O(1) - already optimal
**Streaming**: O(1) - no state needed
```go
change = source[i] - source[i-1]
```
**No optimization needed - inherently O(1)**

### 8. **MACD (Moving Average Convergence Divergence)**
**Current**: O(fastPeriod + slowPeriod + signalPeriod)
**Streaming**: O(1) - three EMA states
```go
fastEma = emaFast.Next(close)
slowEma = emaSlow.Next(close)
macd = fastEma - slowEma
signal = emaSignal.Next(macd)
histogram = macd - signal
```

---

## ❌ REQUIRE O(period) WINDOW SCAN (5 functions)

### 1. **Stdev (Standard Deviation)**
**Complexity**: O(period) - must scan window for variance
**Why**: Needs mean AND deviation from mean
```go
mean = sum(window) / period          // O(period)
variance = sum((x - mean)²) / period // O(period)
stdev = sqrt(variance)
```
**Cannot be O(1)**: Requires two-pass calculation (mean, then variance)

**Possible optimization**: Welford's online algorithm
- O(1) per bar for **rolling** variance
- But still requires window scan for **lookback** access
- Not applicable to security() context where we access arbitrary bars

### 2. **BBands (Bollinger Bands)**
**Complexity**: O(period) - uses SMA + Stdev
**Why**: Stdev inherently O(period)
```go
middle = sma(close, period)    // Can be O(1)
stdev = stdev(close, period)   // MUST be O(period)
upper = middle + k * stdev
lower = middle - k * stdev
```
**Cannot optimize Stdev component**

### 3. **Stoch (Stochastic Oscillator)**
**Complexity**: O(kPeriod) - find min/max in window
**Why**: Must scan window for highest high / lowest low
```go
highestHigh = max(high[i-kPeriod+1..i])  // O(kPeriod)
lowestLow = min(low[i-kPeriod+1..i])     // O(kPeriod)
k = 100 * (close - lowestLow) / (highestHigh - lowestLow)
```
**Cannot be O(1)**: No efficient online min/max for sliding window

**Advanced optimization**: Monotonic deque
- O(1) amortized per bar
- Complex implementation, memory overhead
- Not worth it for typical periods (14)

### 4. **Pivothigh**
**Complexity**: O(leftBars + rightBars) - scan neighborhood
**Why**: Requires future bars (lookahead)
```go
// Check if bar[i] is local maximum
for j in [-leftBars, +rightBars]:
    if source[i+j] > source[i]: not_pivot
```
**Cannot be O(1)**: Inherently requires neighborhood scan

### 5. **Pivotlow**
**Complexity**: O(leftBars + rightBars) - scan neighborhood
**Why**: Requires future bars (lookahead)
```go
// Check if bar[i] is local minimum
for j in [-leftBars, +rightBars]:
    if source[i+j] < source[i]: not_pivot
```
**Cannot be O(1)**: Inherently requires neighborhood scan

---

## OPTIMIZATION IMPACT ANALYSIS

### High Impact (Worth Optimizing)

**SMA** - Most common, large windows (50, 200)
```
Current:  SMA(200) = 200 ops/bar
Streaming: SMA(200) = 1 op/bar
Speedup: 200x
```

**BBands** - Partial optimization
```
Current:  SMA(20) + Stdev(20) = 20 + 20 = 40 ops/bar
Streaming: SMA(20) + Stdev(20) = 1 + 20 = 21 ops/bar
Speedup: 1.9x (only SMA optimized)
```

### Low Impact (Already Fast)

**EMA, RMA** - Already O(1) after warmup
**TR, Change** - Already O(1) always

### Medium Impact

**ATR, RSI, MACD** - Composition of O(1) components
```
ATR: TR O(1) + RMA O(1) = O(1) total
RSI: Change O(1) + RMA O(1) = O(1) total
MACD: 3x EMA O(1) = O(1) total
```

---

## PRACTICAL RECOMMENDATION

### Priority 1: Optimize SMA
**Why**: Most used, largest periods, simple implementation
**Impact**: 50-200x speedup for typical periods

### Priority 2: Optimize RSI/ATR
**Why**: Common indicators, composition benefit
**Impact**: 10-50x speedup

### Priority 3: Don't optimize Stdev/Stoch/Pivots
**Why**: Inherently O(period), small periods (<30), infrequent use
**Impact**: Not worth complexity

---

## CONCLUSION

**8 out of 13 functions (62%) can benefit from streaming O(1) optimization**

**However**, current O(period) inline loops are **acceptable** for typical use:
- SMA(20): 20 operations per bar (fast)
- SMA(200): 200 operations per bar (still reasonable)
- Cost bounded by period, not dataset size

**Streaming optimization worthwhile for**:
- Strategies with many security() calls using large-period SMAs
- Real-time applications where per-bar latency matters
- When SMA period > 100

**Not urgent** for typical backtesting workloads where O(20-50) per bar is negligible.

-----

# TA Inline Loop Performance Analysis

## O(N) Complexity Clarification

**N = window period** (NOT total bars)

### Current Implementation Cost Model

```
Per-bar cost = O(period)
Total cost = O(period × total_bars)

Examples:
  SMA(20) × 5000 bars  = 100,000 iterations   ✅ Acceptable
  SMA(200) × 5000 bars = 1,000,000 iterations ⚠️ Noticeable
  SMA(20) × 50000 bars = 1,000,000 iterations ⚠️ Scaling issue
```

### Streaming State Cost Model

```
Per-bar cost = O(1)
Total cost = O(total_bars)

Examples:
  SMA(20) × 5000 bars  = 5,000 operations   ✅ Optimal
  SMA(200) × 5000 bars = 5,000 operations   ✅ Optimal
  SMA(20) × 50000 bars = 50,000 operations  ✅ Scales linearly
```

---

## Benchmark Estimates (Apple M1)

### Small Window: SMA(20)

**Current (Inline Loop)**:
- Per-bar: 20 iterations × 1.5ns = 30ns
- 5000 bars: 30ns × 5000 = 150μs
- **Status**: Negligible overhead ✅

**Streaming State**:
- Per-bar: 2 operations × 1.5ns = 3ns
- 5000 bars: 3ns × 5000 = 15μs
- **Improvement**: 10x faster (but already fast)

### Medium Window: SMA(50)

**Current (Inline Loop)**:
- Per-bar: 50 iterations × 1.5ns = 75ns
- 5000 bars: 75ns × 5000 = 375μs
- **Status**: Acceptable ✅

**Streaming State**:
- Per-bar: 3ns (constant)
- 5000 bars: 15μs
- **Improvement**: 25x faster

### Large Window: SMA(200)

**Current (Inline Loop)**:
- Per-bar: 200 iterations × 1.5ns = 300ns
- 5000 bars: 300ns × 5000 = 1.5ms
- **Status**: Starting to be noticeable ⚠️

**Streaming State**:
- Per-bar: 3ns (constant)
- 5000 bars: 15μs
- **Improvement**: 100x faster

### Very Large Window: SMA(500)

**Current (Inline Loop)**:
- Per-bar: 500 iterations × 1.5ns = 750ns
- 5000 bars: 750ns × 5000 = 3.75ms
- **Status**: Measurable impact ⚠️⚠️

**Streaming State**:
- Per-bar: 3ns (constant)
- 5000 bars: 15μs
- **Improvement**: 250x faster

---

## Real-World Strategy Impact

### Typical BB Strategy (SMA 20-50)
```pine
bb_basis = security(symbol, "1D", ta.sma(close, 46))  // Medium window
bb_dev = security(symbol, "1D", ta.stdev(close, 46))  // Medium window
```

**Current inline**: ~400μs per strategy run ✅ Acceptable
**Streaming**: ~30μs per strategy run ✅ Excellent
**Verdict**: Current implementation is **production-ready**

### Heavy TA Strategy (Multiple large windows)
```pine
sma200 = security(symbol, "1D", ta.sma(close, 200))   // Large window
sma500 = security(symbol, "1W", ta.sma(close, 500))   // Very large window
bb_basis = security(symbol, "1D", ta.sma(close, 50))  // Medium window
```

**Current inline**: ~5ms per strategy run ⚠️ Noticeable
**Streaming**: ~50μs per strategy run ✅ Excellent
**Verdict**: Streaming states **recommended for optimization**

---

## When Inline Loops Are Acceptable

✅ **Good enough for**:
- Short/medium windows (period ≤ 50)
- Single security() call per strategy
- Moderate dataset sizes (≤ 10k bars)
- Typical BB strategies

⚠️ **Consider streaming states for**:
- Large windows (period > 100)
- Multiple security() calls with TA
- Large datasets (> 20k bars)
- Performance-critical backtesting

---

## Scaling Analysis

### Dataset Size Impact

**Current (Inline Loop)**:
```
Time ∝ period × num_bars

1k bars:  SMA(200) = 200k iterations = 0.3ms   ✅
5k bars:  SMA(200) = 1M iterations = 1.5ms     ✅
10k bars: SMA(200) = 2M iterations = 3ms       ⚠️
50k bars: SMA(200) = 10M iterations = 15ms     ⚠️⚠️
```

**Streaming State**:
```
Time ∝ num_bars (period independent)

1k bars:  SMA(any) = 1k operations = 3μs   ✅
5k bars:  SMA(any) = 5k operations = 15μs  ✅
10k bars: SMA(any) = 10k operations = 30μs ✅
50k bars: SMA(any) = 50k operations = 150μs ✅
```

### Window Size Impact

**Current (Inline Loop)**:
```
Time ∝ period (for fixed num_bars)

SMA(20):  Linear with period × 5000 bars = 150μs   ✅
SMA(50):  Linear with period × 5000 bars = 375μs   ✅
SMA(200): Linear with period × 5000 bars = 1.5ms   ⚠️
SMA(500): Linear with period × 5000 bars = 3.75ms  ⚠️⚠️
```

**Streaming State**:
```
Time = constant (for any period)

SMA(20):  15μs   ✅
SMA(50):  15μs   ✅
SMA(200): 15μs   ✅
SMA(500): 15μs   ✅
```

---

## Conclusion

### Current Implementation Status

**Is O(N) where N = window period** (NOT total bars)
- ✅ Acceptable for typical use cases (period ≤ 50, dataset ≤ 10k)
- ⚠️ Noticeable overhead for large windows (period > 100)
- ⚠️⚠️ Scaling issues with both large windows AND large datasets

### Streaming State Would Provide

**True O(1) per-bar cost**
- ✅ Period-independent performance
- ✅ Linear scaling with dataset size only
- ✅ 10-250x speedup for large windows
- ✅ Completes ForwardSeriesBuffer alignment

### Recommendation

**Production deployment**: Current inline loops are **sufficient** for:
- BB strategies (typical periods 20-50)
- Standard backtesting (5-10k bars)
- Single-timeframe security() calls

**Optimization priority**: Implement streaming states when:
- Using large periods (SMA(200)+)
- Multiple security() calls with TA
- Large-scale backtesting (50k+ bars)
- Performance becomes measurable bottleneck
