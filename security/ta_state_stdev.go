package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// STDEVStateManager computes population standard deviation over rolling window.
// Uses two-pass algorithm: calculate mean, then variance from squared deviations.
type STDEVStateManager struct {
	cacheKey string
	period   int
	buffer   []float64
	computed int
}

// NewSTDEVStateManager creates manager for standard deviation calculation.
func NewSTDEVStateManager(cacheKey string, period int) *STDEVStateManager {
	return &STDEVStateManager{
		cacheKey: cacheKey,
		period:   period,
		buffer:   make([]float64, period),
		computed: 0,
	}
}

// ComputeAtBar calculates population standard deviation for bars ending at barIdx.
// Returns NaN during warmup period (first period-1 bars).
// Algorithm: sqrt(sum((x - mean)^2) / N) where N is period.
func (s *STDEVStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	if err := s.warmupBufferUpTo(secCtx, sourceID, barIdx); err != nil {
		return math.NaN(), err
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	mean, err := s.calculateMeanForWindow(secCtx, sourceID, barIdx)
	if err != nil {
		return math.NaN(), err
	}

	variance, err := s.calculateVarianceForWindow(secCtx, sourceID, barIdx, mean)
	if err != nil {
		return math.NaN(), err
	}

	return math.Sqrt(variance), nil
}

func (s *STDEVStateManager) warmupBufferUpTo(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) error {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return err
		}

		idx := s.computed % s.period
		s.buffer[idx] = sourceVal
		s.computed++
	}
	return nil
}

func (s *STDEVStateManager) calculateMeanForWindow(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	sum := 0.0
	for i := 0; i < s.period; i++ {
		barOffset := barIdx - s.period + 1 + i
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, barOffset)
		if err != nil {
			return 0, err
		}
		sum += sourceVal
	}
	return sum / float64(s.period), nil
}

func (s *STDEVStateManager) calculateVarianceForWindow(secCtx *context.Context, sourceID *ast.Identifier, barIdx int, mean float64) (float64, error) {
	variance := 0.0
	for i := 0; i < s.period; i++ {
		barOffset := barIdx - s.period + 1 + i
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, barOffset)
		if err != nil {
			return 0, err
		}
		deviation := sourceVal - mean
		variance += deviation * deviation
	}
	return variance / float64(s.period), nil
}
