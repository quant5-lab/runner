package codegen

import "fmt"

// AccumulatorStrategy defines how values are accumulated during iteration over a lookback period.
//
// This interface implements the Strategy pattern, allowing different accumulation algorithms
// to be plugged into the TA indicator builder without modifying its code.
//
// Implementing a new strategy:
//
//	type MyAccumulator struct {
//	    // state fields
//	}
//
//	func (m *MyAccumulator) Initialize() string {
//	    return "myVar := 0.0"  // Variable declarations
//	}
//
//	func (m *MyAccumulator) Accumulate(value string) string {
//	    return fmt.Sprintf("myVar += transform(%s)", value)  // Loop body
//	}
//
//	func (m *MyAccumulator) Finalize(period int) string {
//	    return "myVar / float64(count)"  // Final calculation
//	}
//
//	func (m *MyAccumulator) NeedsNaNGuard() bool {
//	    return true  // Whether to check for NaN in input values
//	}
//
// The builder will generate:
//
//	if ctx.BarIndex < period-1 {
//	    seriesVar.Set(math.NaN())  // Warmup period
//	} else {
//	    myVar := 0.0               // Initialize()
//	    hasNaN := false            // Added if NeedsNaNGuard() == true
//	    for j := 0; j < period; j++ {
//	        val := accessor.Get(j)
//	        if math.IsNaN(val) {   // Added if NeedsNaNGuard() == true
//	            hasNaN = true
//	        }
//	        myVar += transform(val)  // Accumulate(value)
//	    }
//	    if hasNaN {                 // Added if NeedsNaNGuard() == true
//	        seriesVar.Set(math.NaN())
//	    } else {
//	        seriesVar.Set(myVar / float64(count))  // Finalize(period)
//	    }
//	}
type AccumulatorStrategy interface {
	// Initialize returns code for variable declarations before the loop
	Initialize() string

	// Accumulate returns code for the loop body that processes each value
	Accumulate(value string) string

	// Finalize returns the final calculation expression after the loop
	Finalize(period int) string

	// NeedsNaNGuard indicates whether NaN checking should be added
	NeedsNaNGuard() bool
}

// SumAccumulator accumulates values by summing them, used for SMA calculations.
//
// Generates code that sums all values in the lookback period and divides by the period:
//
//	sum := 0.0
//	hasNaN := false
//	for j := 0; j < period; j++ {
//	    val := data.Get(j)
//	    if math.IsNaN(val) { hasNaN = true }
//	    sum += val
//	}
//	result = sum / period
type SumAccumulator struct{}

// NewSumAccumulator creates a sum accumulator for SMA-style calculations.
func NewSumAccumulator() *SumAccumulator {
	return &SumAccumulator{}
}

func (s *SumAccumulator) Initialize() string {
	return "sum := 0.0\nhasNaN := false"
}

func (s *SumAccumulator) Accumulate(value string) string {
	return fmt.Sprintf("sum += %s", value)
}

func (s *SumAccumulator) Finalize(period int) string {
	return fmt.Sprintf("sum / %d.0", period)
}

func (s *SumAccumulator) NeedsNaNGuard() bool {
	return true
}

// VarianceAccumulator calculates variance for standard deviation (STDEV).
//
// This accumulator requires a pre-calculated mean value. It computes:
//
//	variance = Σ(value - mean)² / period
//
// Usage (two-pass STDEV calculation):
//
//	// Pass 1: Calculate mean
//	meanBuilder := NewTAIndicatorBuilder("STDEV_MEAN", "stdev20", 20, accessor, false)
//	meanBuilder.WithAccumulator(NewSumAccumulator())
//	meanCode := meanBuilder.Build()
//
//	// Pass 2: Calculate variance
//	varianceBuilder := NewTAIndicatorBuilder("STDEV", "stdev20", 20, accessor, false)
//	varianceBuilder.WithAccumulator(NewVarianceAccumulator("mean"))
//	varianceCode := varianceBuilder.Build()
type VarianceAccumulator struct {
	mean string // Variable name containing the pre-calculated mean
}

// NewVarianceAccumulator creates a variance accumulator for STDEV calculations.
//
// Parameters:
//   - mean: Variable name containing the pre-calculated mean value
func NewVarianceAccumulator(mean string) *VarianceAccumulator {
	return &VarianceAccumulator{mean: mean}
}

func (v *VarianceAccumulator) Initialize() string {
	return "variance := 0.0"
}

func (v *VarianceAccumulator) Accumulate(value string) string {
	return fmt.Sprintf("diff := %s - %s\nvariance += diff * diff", value, v.mean)
}

func (v *VarianceAccumulator) Finalize(period int) string {
	return fmt.Sprintf("variance /= %d.0", period)
}

func (v *VarianceAccumulator) NeedsNaNGuard() bool {
	return false // Mean calculation already filtered NaN values
}

// EMAAccumulator applies exponential moving average weighting.
//
// EMA formula: EMA = α * current + (1 - α) * previous_EMA
// where α = 2 / (period + 1)
//
// Unlike SMA, EMA gives more weight to recent values and requires
// special initialization handling for the first value.
type EMAAccumulator struct {
	alpha     string // Smoothing factor expression
	resultVar string // Variable name for EMA result
}

func NewEMAAccumulator(period int) *EMAAccumulator {
	return &EMAAccumulator{
		alpha:     fmt.Sprintf("2.0 / float64(%d+1)", period),
		resultVar: "ema",
	}
}

func (e *EMAAccumulator) Initialize() string {
	return fmt.Sprintf("alpha := %s", e.alpha)
}

func (e *EMAAccumulator) Accumulate(value string) string {
	return fmt.Sprintf("%s = alpha*%s + (1-alpha)*%s", e.resultVar, value, e.resultVar)
}

func (e *EMAAccumulator) Finalize(period int) string {
	return ""
}

func (e *EMAAccumulator) NeedsNaNGuard() bool {
	return true
}

func (e *EMAAccumulator) GetResultVariable() string {
	return e.resultVar
}
