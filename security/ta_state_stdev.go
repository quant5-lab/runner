package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type STDEVStateManager struct {
	cacheKey  string
	period    int
	buf       *series.Series
	computed  int
	evaluator BarEvaluator
}

func NewSTDEVStateManager(cacheKey string, period int, capacity int, evaluator BarEvaluator) *STDEVStateManager {
	return &STDEVStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       series.NewSeries(max(capacity, 1)),
		evaluator: evaluator,
	}
}

func (s *STDEVStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		if s.computed < s.period-1 {
			s.buf.Set(math.NaN())
		} else {
			mean, err := s.windowMean(secCtx, sourceExpr, s.computed)
			if err != nil {
				return math.NaN(), err
			}

			variance, err := s.windowVariance(secCtx, sourceExpr, s.computed, mean)
			if err != nil {
				return math.NaN(), err
			}

			s.buf.Set(math.Sqrt(variance))
		}

		s.computed++
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}

func (s *STDEVStateManager) windowMean(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	sum := 0.0
	for i := 0; i < s.period; i++ {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, barIdx-s.period+1+i)
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return sum / float64(s.period), nil
}

func (s *STDEVStateManager) windowVariance(secCtx *context.Context, sourceExpr ast.Expression, barIdx int, mean float64) (float64, error) {
	variance := 0.0
	for i := 0; i < s.period; i++ {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, barIdx-s.period+1+i)
		if err != nil {
			return 0, err
		}
		d := v - mean
		variance += d * d
	}
	return variance / float64(s.period), nil
}
