package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type STDEVStateManager struct {
	cacheKey string
	period   int
	storage  TASeriesStorage
	computed int
}

func NewSTDEVStateManager(cacheKey string, period int, capacity int) *STDEVStateManager {
	return &STDEVStateManager{
		cacheKey: cacheKey,
		period:   period,
		storage:  NewSeriesStorage(capacity),
		computed: 0,
	}
}

func (s *STDEVStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed < s.period-1 {
			s.storage.Set(s.computed, math.NaN())
			s.computed++
			continue
		}

		mean, err := s.calculateMeanForWindow(secCtx, sourceID, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		variance, err := s.calculateVarianceForWindow(secCtx, sourceID, s.computed, mean)
		if err != nil {
			return math.NaN(), err
		}

		stdevValue := math.Sqrt(variance)
		s.storage.Set(s.computed, stdevValue)
		s.computed++
	}

	return s.storage.Get(barIdx), nil
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
