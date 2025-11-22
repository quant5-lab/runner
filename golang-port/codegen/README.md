# CodeGen Package

Transpiles PineScript AST to executable Go code with focus on modularity, testability, and extensibility.

## Quick Start

### Generate a Simple Moving Average (SMA)

```go
// 1. Create data accessor
accessor := CreateAccessGenerator("close")

// 2. Build indicator
builder := NewTAIndicatorBuilder("SMA", "sma20", 20, accessor, false)
builder.WithAccumulator(NewSumAccumulator())

// 3. Generate code
code := builder.Build()
```

### Generate Standard Deviation (STDEV)

```go
accessor := CreateAccessGenerator("close")

// Pass 1: Calculate mean
meanBuilder := NewTAIndicatorBuilder("MEAN", "mean", 20, accessor, false)
meanBuilder.WithAccumulator(NewSumAccumulator())
meanCode := meanBuilder.Build()

// Pass 2: Calculate variance
varianceBuilder := NewTAIndicatorBuilder("STDEV", "stdev20", 20, accessor, false)
varianceBuilder.WithAccumulator(NewVarianceAccumulator("mean"))
varianceCode := varianceBuilder.Build()

code := meanCode + "\n" + varianceCode
```

## Architecture

The package follows **SOLID principles** with modular, reusable components:

### Core Components

| Component | Purpose | Pattern |
|-----------|---------|---------|
| `TAIndicatorBuilder` | Constructs TA indicator code | Builder |
| `AccumulatorStrategy` | Defines accumulation logic | Strategy |
| `LoopGenerator` | Creates for-loop structures | - |
| `WarmupChecker` | Handles warmup periods | - |
| `CodeIndenter` | Manages indentation | - |
| `AccessGenerator` | Abstracts data access | Strategy |

### Design Patterns

- **Builder Pattern**: `TAIndicatorBuilder` for complex construction
- **Strategy Pattern**: `AccumulatorStrategy`, `AccessGenerator` for pluggable algorithms
- **Factory Pattern**: `CreateAccessGenerator()` for automatic type detection

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed design documentation.

## Available Components

### Accumulator Strategies

Implement custom accumulation logic for indicators:

```go
type AccumulatorStrategy interface {
    Initialize() string        // Variable declarations
    Accumulate(value string) string  // Loop body
    Finalize(period int) string      // Final calculation
    NeedsNaNGuard() bool             // Whether to check for NaN
}
```

**Built-in strategies**:
- `SumAccumulator`: Sum values (for SMA)
- `VarianceAccumulator`: Calculate variance (for STDEV)
- `EMAAccumulator`: Exponential weighting (for EMA)

### Access Generators

Abstract data source access:

```go
type AccessGenerator interface {
    GenerateLoopValueAccess(loopVar string) string
    GenerateInitialValueAccess(period int) string
}
```

**Built-in generators**:
- `SeriesVariableAccessGenerator`: `sma20Series.Get(offset)`
- `OHLCVFieldAccessGenerator`: `ctx.Data[ctx.BarIndex-offset].Close`

**Factory**:
```go
accessor := CreateAccessGenerator("close")  // Auto-detects OHLCV field
accessor := CreateAccessGenerator("sma20Series.Get(0)")  // Auto-detects Series variable
```

## Usage Examples

### Example 1: Simple SMA

```go
package main

import "github.com/borisquantlab/pinescript-go/codegen"

func generateSMA() string {
    accessor := codegen.CreateAccessGenerator("close")
    builder := codegen.NewTAIndicatorBuilder("SMA", "sma50", 50, accessor, false)
    builder.WithAccumulator(codegen.NewSumAccumulator())
    return builder.Build()
}
```

Output:
```go
/* Inline SMA(50) */
if ctx.BarIndex < 50-1 {
    sma50Series.Set(math.NaN())
} else {
    sum := 0.0
    hasNaN := false
    for j := 0; j < 50; j++ {
        val := ctx.Data[ctx.BarIndex-j].Close
        if math.IsNaN(val) {
            hasNaN = true
        }
        sum += val
    }
    if hasNaN {
        sma50Series.Set(math.NaN())
    } else {
        sma50Series.Set(sum / 50.0)
    }
}
```

### Example 2: Custom Accumulator (WMA)

```go
// Weighted Moving Average accumulator
type WMAAccumulator struct {
    period int
}

func NewWMAAccumulator(period int) *WMAAccumulator {
    return &WMAAccumulator{period: period}
}

func (w *WMAAccumulator) Initialize() string {
    return "weightedSum := 0.0\nweightSum := 0.0\nhasNaN := false"
}

func (w *WMAAccumulator) Accumulate(value string) string {
    return fmt.Sprintf(
        "weight := float64(%d - j)\nweightedSum += %s * weight\nweightSum += weight",
        w.period, value,
    )
}

func (w *WMAAccumulator) Finalize(period int) string {
    return "weightedSum / weightSum"
}

func (w *WMAAccumulator) NeedsNaNGuard() bool {
    return true
}

// Usage
func generateWMA() string {
    accessor := codegen.CreateAccessGenerator("close")
    builder := codegen.NewTAIndicatorBuilder("WMA", "wma20", 20, accessor, false)
    builder.WithAccumulator(NewWMAAccumulator(20))
    return builder.Build()
}
```

### Example 3: Building Step by Step

```go
builder := NewTAIndicatorBuilder("SMA", "sma20", 20, accessor, false)
builder.WithAccumulator(NewSumAccumulator())

// Build each component separately
header := builder.BuildHeader()
warmup := builder.BuildWarmupCheck()
init := builder.BuildInitialization()
loop := builder.BuildLoop()
finalization := builder.BuildFinalization()

// Or build all at once
code := builder.Build()
```

## Testing

Comprehensive test coverage with 40+ tests:

```bash
# Run all codegen tests
go test ./codegen -v

# Run specific test suite
go test ./codegen -run TestTAIndicatorBuilder -v

# Run with coverage
go test ./codegen -cover
```

### Test Files

- `series_accessor_test.go`: AccessGenerator implementations (24 tests)
- `ta_components_test.go`: Accumulators and WarmupChecker
- `loop_generator_test.go`: LoopGenerator
- `ta_indicator_builder_test.go`: TAIndicatorBuilder integration (9 tests)

## Extending the Package

### Adding a New Indicator

1. **Determine accumulation logic**
2. **Create or reuse accumulator**
3. **Use builder**

Example - RSI (Relative Strength Index):

```go
// RSI needs custom accumulation
type RSIAccumulator struct {
    period int
}

func (r *RSIAccumulator) Initialize() string {
    return "gainSum := 0.0\nlossSum := 0.0"
}

func (r *RSIAccumulator) Accumulate(value string) string {
    return fmt.Sprintf(`
        change := %s - prevValue
        if change > 0 {
            gainSum += change
        } else {
            lossSum += math.Abs(change)
        }
        prevValue = %s
    `, value, value)
}

func (r *RSIAccumulator) Finalize(period int) string {
    return fmt.Sprintf("100 - (100 / (1 + (gainSum/%d.0) / (lossSum/%d.0)))", period, period)
}

func (r *RSIAccumulator) NeedsNaNGuard() bool {
    return true
}
```

### Adding New Build Steps

Extend `TAIndicatorBuilder`:

```go
func (b *TAIndicatorBuilder) BuildValidation() string {
    return b.indenter.Line("if period < 1 { return error }")
}

// Use in custom build workflow
code := builder.BuildHeader()
code += builder.BuildValidation()  // New step
code += builder.BuildWarmupCheck()
// ...
```

## API Reference

### TAIndicatorBuilder

```go
// Constructor
func NewTAIndicatorBuilder(
    name string,      // Indicator name
    varName string,   // Output variable
    period int,       // Lookback period
    accessor AccessGenerator,  // Data source
    needsNaN bool,    // Add NaN checking
) *TAIndicatorBuilder

// Methods
func (b *TAIndicatorBuilder) WithAccumulator(acc AccumulatorStrategy) *TAIndicatorBuilder
func (b *TAIndicatorBuilder) Build() string
func (b *TAIndicatorBuilder) BuildHeader() string
func (b *TAIndicatorBuilder) BuildWarmupCheck() string
func (b *TAIndicatorBuilder) BuildInitialization() string
func (b *TAIndicatorBuilder) BuildLoop() string
func (b *TAIndicatorBuilder) BuildFinalization() string
```

### AccumulatorStrategy Interface

```go
type AccumulatorStrategy interface {
    Initialize() string               // Code before loop
    Accumulate(value string) string   // Code inside loop
    Finalize(period int) string       // Final expression
    NeedsNaNGuard() bool              // Add NaN checking?
}
```

### Factory Functions

```go
// Create appropriate accessor based on expression
func CreateAccessGenerator(expr string) AccessGenerator

// Create built-in accumulators
func NewSumAccumulator() *SumAccumulator
func NewEMAAccumulator(period int) *EMAAccumulator
func NewVarianceAccumulator(mean string) *VarianceAccumulator

// Create utilities
func NewLoopGenerator(period int, accessor AccessGenerator, needsNaN bool) *LoopGenerator
func NewWarmupChecker(period int) *WarmupChecker
func NewCodeIndenter() *CodeIndenter
```

## Best Practices

### ✅ Do

- Use interfaces for flexibility
- Keep components small and focused
- Write tests first (TDD)
- Document public APIs
- Follow SOLID principles

### ❌ Don't

- Hardcode indentation (use `CodeIndenter`)
- Mix responsibilities in one component
- Skip testing edge cases
- Create tight coupling between components
- Duplicate code generation logic

## Performance Considerations

- Components are lightweight (no heavy allocations)
- String building uses efficient concatenation
- Builders can be reused for multiple indicators
- Factory pattern avoids duplicate type detection

## Contributing

When adding new components:

1. Follow existing patterns (Builder, Strategy)
2. Write comprehensive tests
3. Add godoc comments with examples
4. Update ARCHITECTURE.md if adding new patterns
5. Ensure all tests pass: `go test ./... -v`

## Resources

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed design documentation
- [Go Design Patterns](https://refactoring.guru/design-patterns/go) - Pattern reference
- [SOLID Principles](https://dave.cheney.net/2016/08/20/solid-go-design) - SOLID in Go

## License

Part of the PineScript-Go transpiler project.
