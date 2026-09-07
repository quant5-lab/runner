package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type STDEVStateManager struct {
	cacheKey  string
	period    int
	buf       forwardBufferE
	evaluator BarEvaluator
}

func NewSTDEVStateManager(cacheKey string, period int, capacity int, evaluator BarEvaluator) *STDEVStateManager {
	return &STDEVStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *STDEVStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		if bar < s.period-1 {
			return math.NaN(), nil
		}
		return s.stdevAtBar(secCtx, sourceExpr, bar)
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}

func (s *STDEVStateManager) stdevAtBar(secCtx *context.Context, sourceExpr ast.Expression, bar int) (float64, error) {
	mean, err := s.windowMean(secCtx, sourceExpr, bar)
	if err != nil {
		return math.NaN(), err
	}
	variance, err := s.windowVariance(secCtx, sourceExpr, bar, mean)
	if err != nil {
		return math.NaN(), err
	}
	return math.Sqrt(variance), nil
}

func (s *STDEVStateManager) windowMean(secCtx *context.Context, sourceExpr ast.Expression, bar int) (float64, error) {
	sum := 0.0
	for i := 0; i < s.period; i++ {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar-s.period+1+i)
		if err != nil {
			return 0, err
		}
		sum += v
	}
	return sum / float64(s.period), nil
}

func (s *STDEVStateManager) windowVariance(secCtx *context.Context, sourceExpr ast.Expression, bar int, mean float64) (float64, error) {
	variance := 0.0
	for i := 0; i < s.period; i++ {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar-s.period+1+i)
		if err != nil {
			return 0, err
		}
		d := tvStdevDiff(v, mean)
		variance += d * d
	}
	return variance / float64(s.period), nil
}

/* tvStdevDiff mirrors runtime/ta.tvStdevDiff: applies Pine ta.stdev's
 * per-deviation epsilon compensation to avoid float drift in marginal cases. */
func tvStdevDiff(value, mean float64) float64 {
	diff := value - mean
	if math.Abs(diff) <= 1e-10 {
		return 0
	}
	if math.Abs(diff) <= 1e-4 {
		return 1e-5
	}
	return diff
}
