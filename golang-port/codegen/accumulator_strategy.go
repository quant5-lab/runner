package codegen

import "fmt"

// AccumulatorStrategy defines how values are accumulated during iteration
type AccumulatorStrategy interface {
	Initialize() string
	Accumulate(value string) string
	Finalize(period int) string
	NeedsNaNGuard() bool
}

// SumAccumulator sums values for average calculations
type SumAccumulator struct{}

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

// VarianceAccumulator calculates variance for standard deviation
type VarianceAccumulator struct {
	mean string
}

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
	return false
}

// EMAAccumulator applies exponential moving average calculation
type EMAAccumulator struct {
	alpha     string
	resultVar string
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
