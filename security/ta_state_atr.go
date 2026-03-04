package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type ATRStateManager struct {
	cacheKey     string
	period       int
	trCalculator *TrueRangeCalculator
	buf          *series.Series
	prevClose    float64
	computed     int
	hasHistory   bool
}

func NewATRStateManager(cacheKey string, period int, capacity int) *ATRStateManager {
	return &ATRStateManager{
		cacheKey:     cacheKey,
		period:       period,
		trCalculator: NewTrueRangeCalculator(),
		buf:          series.NewSeries(max(capacity, 1)),
	}
}

func (s *ATRStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed >= len(secCtx.Data) {
			break
		}

		if s.computed > 0 {
			s.buf.Next()
		}

		isFirstBar := s.computed == 0 || !s.hasHistory
		trueRange := s.trCalculator.CalculateAtBar(secCtx.Data, s.computed, s.prevClose, isFirstBar, true)

		var atrValue float64
		switch {
		case s.computed == 0:
			atrValue = trueRange
		case s.computed < s.period:
			prevATR := s.buf.Get(1)
			atrValue = (prevATR*float64(s.computed) + trueRange) / float64(s.computed+1)
		default:
			alpha := 1.0 / float64(s.period)
			prevATR := s.buf.Get(1)
			atrValue = alpha*trueRange + (1-alpha)*prevATR
		}

		s.buf.Set(atrValue)
		s.prevClose = secCtx.Data[s.computed].Close
		s.hasHistory = true
		s.computed++
	}

	if barIdx < s.period-1 {
		return 0.0, nil
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}
