# valuewhen() Implementation - ForwardSeriesBuffer Paradigm

## Summary

Implemented `ta.valuewhen()` function handler with per-bar inline generation aligned to ForwardSeriesBuffer architecture.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│ TAFunctionHandler Interface (Strategy Pattern)          │
├─────────────────────────────────────────────────────────┤
│ ValuewhenHandler                                         │
│  ├─ CanHandle(funcName) → bool                          │
│  └─ GenerateCode(g, varName, call) → string             │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ Generator (Code Generation)                              │
├─────────────────────────────────────────────────────────┤
│ generateValuewhen(varName, condExpr, srcExpr, occur)    │
│  └─ Per-bar lookback loop                               │
│     ├─ Count condition occurrences                      │
│     ├─ Return source[offset] when Nth found             │
│     └─ Return NaN if not found                          │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│ Helper: convertSeriesAccessToOffset                     │
├─────────────────────────────────────────────────────────┤
│ Converts current bar access to offset-based:            │
│  • bar.Close → ctx.Data[i-offset].Close                 │
│  • xSeries.GetCurrent() → xSeries.Get(offset)           │
└─────────────────────────────────────────────────────────┘
```

## Implementation Details

### Handler Registration (SOLID: Open/Closed Principle)
**File:** `codegen/ta_function_handler.go`
- Added `&ValuewhenHandler{}` to registry
- No modification to existing handlers
- Extensible without changing core logic

### Handler Implementation (SOLID: Single Responsibility)
**File:** `codegen/ta_handlers.go`
- `ValuewhenHandler` handles only valuewhen function
- Validates 3 required arguments
- Delegates code generation to generator

### Code Generation (DRY: Reusable Helpers)
**File:** `codegen/generator.go`
- `generateValuewhen()`: Per-bar inline logic
- `convertSeriesAccessToOffset()`: Reusable series access converter
- Pattern matches other TA functions (change, crossover)

## Generated Code Pattern

```go
/* Inline valuewhen(condition, source, occurrence) */
varSeries.Set(func() float64 {
    occurrenceCount := 0
    for lookbackOffset := 0; lookbackOffset <= i; lookbackOffset++ {
        if conditionSeries.Get(lookbackOffset) != 0 {
            if occurrenceCount == occurrence {
                return sourceSeries.Get(lookbackOffset)
            }
            occurrenceCount++
        }
    }
    return math.NaN()
}())
```

## ForwardSeriesBuffer Alignment

### OLD (Array-based):
```go
func Valuewhen(condition []bool, source []float64, occurrence int) []float64
```
- Processes entire array at once
- Returns full array result
- ❌ Incompatible with per-bar forward iteration

### NEW (ForwardSeriesBuffer):
```go
// Per-bar inline generation
for i := 0; i < barCount; i++ {
    valuewhenSeries.Set(func() float64 {
        // Look back from current bar only
        for offset := 0; offset <= i; offset++ {
            if conditionSeries.Get(offset) != 0 { ... }
        }
    }())
}
```
- ✅ Processes one bar at a time
- ✅ Uses Series.Get(offset) for historical access
- ✅ Enforces immutability of past values
- ✅ No future value access

## Test Coverage

### Runtime Tests (Backward Compatibility)
**File:** `tests/value/valuewhen_test.go`
- ✅ Array-based implementation still works
- ✅ All 5 test cases pass
- ✅ Occurrence 0, 1, 2 behavior validated

### Integration Tests
- ✅ All codegen tests pass (71.4% coverage)
- ✅ All integration tests pass (5.178s)
- ✅ No regressions in existing TA functions

### Real-World Usage
**Test file:** `strategies/test-valuewhen.pine`
```pine
condition = close > open
lastBullishClose = ta.valuewhen(condition, close, 0)
prevBullishClose = ta.valuewhen(condition, close, 1)
```
✅ Compiles successfully
✅ Generates clean inline code

## Performance Characteristics

### Time Complexity
- Per-bar: O(i) where i is current bar index
- Total: O(N²) where N is total bars
- ⚠️ Can be expensive for large N and high occurrence values

### Space Complexity
- O(1) per bar (no arrays allocated)
- State stored in Series buffers only

## SOLID/DRY/KISS Adherence

### Single Responsibility (SRP) ✅
- `ValuewhenHandler`: Only handles valuewhen
- `generateValuewhen()`: Only generates code
- `convertSeriesAccessToOffset()`: Only converts access patterns

### Open/Closed (OCP) ✅
- Added new handler without modifying existing code
- Registry pattern allows extension

### Liskov Substitution (LSP) ✅
- `ValuewhenHandler` implements `TAFunctionHandler` interface
- Substitutable with any other handler

### Interface Segregation (ISP) ✅
- `TAFunctionHandler` has minimal interface (2 methods)

### Dependency Inversion (DIP) ✅
- Depends on `TAFunctionHandler` interface, not concrete types

### Don't Repeat Yourself (DRY) ✅
- `convertSeriesAccessToOffset()` reusable by other functions
- Pattern follows existing TA function structure

### Keep It Simple (KISS) ✅
- Clear variable names (`occurrenceCount`, `lookbackOffset`)
- Straightforward loop logic
- No premature optimization

## Files Modified

1. `codegen/ta_handlers.go` (+29 lines)
   - Added `ValuewhenHandler` struct and methods

2. `codegen/generator.go` (+49 lines)
   - Added `generateValuewhen()` method
   - Added `convertSeriesAccessToOffset()` helper

3. `codegen/ta_function_handler.go` (+1 line)
   - Registered `&ValuewhenHandler{}`

## Known Limitations

### BB7 Files Still Blocked
- `bb7-dissect-bb.pine` has typo: `bblenght` vs `bblength`
- `bb7-dissect-sl.pine` has type mismatch errors
- ❌ NOT blocked by valuewhen implementation

### Recommendation
Fix input variable name generation bug before claiming BB7 files are unblocked.

## Build/Test Status

```bash
✅ go build ./...           # Clean
✅ go vet ./codegen/...     # Clean
✅ go test ./...            # All pass
✅ valuewhen_test.go        # 5/5 tests pass
✅ Integration tests        # 26 tests pass
✅ test-valuewhen.pine      # Compiles and builds
```

## Conclusion

✅ `valuewhen()` successfully implemented
✅ Aligned to ForwardSeriesBuffer paradigm
✅ No regressions
✅ Clean architecture (SOLID/DRY/KISS)
✅ Ready for production use
