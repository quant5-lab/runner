package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// SARStateManager computes the Parabolic Stop-and-Reverse indicator bar by bar.
// Ignores sourceID — uses High and Low from secCtx.Data directly.
type SARStateManager struct {
	cacheKey  string
	start     float64
	inc       float64
	maxAF     float64
	storage   TASeriesStorage
	computed  int
	isUptrend bool
	sar       float64
	ep        float64
	af        float64
}

func NewSARStateManager(cacheKey string, start, inc, maxAF float64, capacity int) *SARStateManager {
	return &SARStateManager{
		cacheKey: cacheKey,
		start:    start,
		inc:      inc,
		maxAF:    maxAF,
		storage:  NewSeriesStorage(capacity),
		computed: 0,
	}
}

func (s *SARStateManager) ComputeAtBar(secCtx *context.Context, _ *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed == 0 {
			s.storage.Set(0, math.NaN())
			s.computed++
			continue
		}

		if s.computed == 1 {
			high0 := secCtx.Data[0].High
			low0 := secCtx.Data[0].Low
			high1 := secCtx.Data[1].High

			s.isUptrend = high1 >= high0
			s.af = s.start
			if s.isUptrend {
				s.sar = low0
				s.ep = high0
			} else {
				s.sar = high0
				s.ep = low0
			}
			s.storage.Set(0, s.sar)
		}

		i := s.computed
		high := secCtx.Data[i].High
		low := secCtx.Data[i].Low
		highPrev := secCtx.Data[i-1].High
		lowPrev := secCtx.Data[i-1].Low

		projectedSAR := s.sar + s.af*(s.ep-s.sar)

		if s.isUptrend {
			if i >= 2 && projectedSAR > secCtx.Data[i-2].Low {
				projectedSAR = secCtx.Data[i-2].Low
			}
			if projectedSAR > lowPrev {
				projectedSAR = lowPrev
			}
			if low < projectedSAR {
				s.isUptrend = false
				projectedSAR = s.ep
				s.ep = low
				s.af = s.start
			} else {
				if high > s.ep {
					s.ep = high
					s.af = math.Min(s.af+s.inc, s.maxAF)
				}
			}
		} else {
			if i >= 2 && projectedSAR < secCtx.Data[i-2].High {
				projectedSAR = secCtx.Data[i-2].High
			}
			if projectedSAR < highPrev {
				projectedSAR = highPrev
			}
			if high > projectedSAR {
				s.isUptrend = true
				projectedSAR = s.ep
				s.ep = high
				s.af = s.start
			} else {
				if low < s.ep {
					s.ep = low
					s.af = math.Min(s.af+s.inc, s.maxAF)
				}
			}
		}

		s.sar = projectedSAR
		s.storage.Set(s.computed, s.sar)
		s.computed++
	}

	return s.storage.Get(barIdx), nil
}
