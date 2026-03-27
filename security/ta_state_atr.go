package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type ATRStateManager struct {
	cacheKey     string
	period       int
	trCalculator *TrueRangeCalculator
	prevClose    float64
	hasHistory   bool
	buf          forwardBufferE
}

func NewATRStateManager(cacheKey string, period int, capacity int) *ATRStateManager {
	return &ATRStateManager{
		cacheKey:     cacheKey,
		period:       period,
		trCalculator: NewTrueRangeCalculator(),
		buf:          newForwardBufferE(capacity),
	}
}

func (s *ATRStateManager) ComputeAtBar(secCtx *context.Context, _ ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
		s.prevClose = 0
		s.hasHistory = false
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		if bar >= len(secCtx.Data) {
			return math.NaN(), nil
		}
		return s.atrAtBar(secCtx, bar), nil
	}); err != nil {
		return math.NaN(), err
	}
	if barIdx < s.period-1 {
		return math.NaN(), nil
	}
	return s.buf.at(barIdx), nil
}

func (s *ATRStateManager) atrAtBar(secCtx *context.Context, bar int) float64 {
	isFirstBar := bar == 0 || !s.hasHistory
	tr := s.trCalculator.CalculateAtBar(secCtx.Data, bar, s.prevClose, isFirstBar, true)
	var atr float64
	switch {
	case bar == 0:
		atr = tr
	case bar < s.period:
		atr = (s.buf.prev()*float64(bar) + tr) / float64(bar+1)
	default:
		alpha := 1.0 / float64(s.period)
		atr = alpha*tr + (1-alpha)*s.buf.prev()
	}
	s.prevClose = secCtx.Data[bar].Close
	s.hasHistory = true
	return atr
}
