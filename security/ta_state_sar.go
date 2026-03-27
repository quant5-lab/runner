package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// SARStateManager ignores sourceExpr — uses High and Low from secCtx.Data directly.
type SARStateManager struct {
	cacheKey  string
	start     float64
	inc       float64
	maxAF     float64
	isUptrend bool
	sar       float64
	ep        float64
	af        float64
	buf       forwardBufferE
}

func NewSARStateManager(cacheKey string, start, inc, maxAF float64, capacity int) *SARStateManager {
	return &SARStateManager{
		cacheKey: cacheKey,
		start:    start,
		inc:      inc,
		maxAF:    maxAF,
		buf:      newForwardBufferE(capacity),
	}
}

func (s *SARStateManager) ComputeAtBar(secCtx *context.Context, _ ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
		s.isUptrend = false
		s.sar = 0
		s.ep = 0
		s.af = 0
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		if bar == 0 {
			return math.NaN(), nil
		}
		if bar == 1 {
			s.initFromBar0(secCtx.Data)
		}
		result := s.projectSAR(secCtx.Data, bar)
		s.sar = result
		return result, nil
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}

func (s *SARStateManager) initFromBar0(data []context.OHLCV) {
	high0 := data[0].High
	low0 := data[0].Low
	high1 := data[1].High
	s.isUptrend = high1 >= high0
	s.af = s.start
	if s.isUptrend {
		s.sar = low0
		s.ep = high0
	} else {
		s.sar = high0
		s.ep = low0
	}
}

func (s *SARStateManager) projectSAR(data []context.OHLCV, i int) float64 {
	high := data[i].High
	low := data[i].Low
	highPrev := data[i-1].High
	lowPrev := data[i-1].Low
	projected := s.sar + s.af*(s.ep-s.sar)
	if s.isUptrend {
		if i >= 2 && projected > data[i-2].Low {
			projected = data[i-2].Low
		}
		if projected > lowPrev {
			projected = lowPrev
		}
		if low < projected {
			s.isUptrend = false
			projected = s.ep
			s.ep = low
			s.af = s.start
		} else if high > s.ep {
			s.ep = high
			s.af = math.Min(s.af+s.inc, s.maxAF)
		}
	} else {
		if i >= 2 && projected < data[i-2].High {
			projected = data[i-2].High
		}
		if projected < highPrev {
			projected = highPrev
		}
		if high > projected {
			s.isUptrend = true
			projected = s.ep
			s.ep = high
			s.af = s.start
		} else if low < s.ep {
			s.ep = low
			s.af = math.Min(s.af+s.inc, s.maxAF)
		}
	}
	return projected
}
