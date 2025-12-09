package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type TAStateManager interface {
	ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error)
}

type SMAStateManager struct {
	cacheKey string
	period   int
	buffer   []float64
	computed int
}

type EMAStateManager struct {
	cacheKey   string
	period     int
	prevEMA    float64
	multiplier float64
	computed   int
}

type RMAStateManager struct {
	cacheKey string
	period   int
	prevRMA  float64
	computed int
}

type RSIStateManager struct {
	cacheKey string
	period   int
	rmaGain  *RMAStateManager
	rmaLoss  *RMAStateManager
	computed int
}

func NewTAStateManager(cacheKey string, period int, capacity int) TAStateManager {
	if contains(cacheKey, "sma") {
		return &SMAStateManager{
			cacheKey: cacheKey,
			period:   period,
			buffer:   make([]float64, period),
			computed: 0,
		}
	}

	if contains(cacheKey, "ema") {
		multiplier := 2.0 / float64(period+1)
		return &EMAStateManager{
			cacheKey:   cacheKey,
			period:     period,
			multiplier: multiplier,
			computed:   0,
		}
	}

	if contains(cacheKey, "rma") {
		return &RMAStateManager{
			cacheKey: cacheKey,
			period:   period,
			computed: 0,
		}
	}

	if contains(cacheKey, "rsi") {
		return &RSIStateManager{
			cacheKey: cacheKey,
			period:   period,
			rmaGain: &RMAStateManager{
				cacheKey: cacheKey + "_gain",
				period:   period,
				computed: 0,
			},
			rmaLoss: &RMAStateManager{
				cacheKey: cacheKey + "_loss",
				period:   period,
				computed: 0,
			},
			computed: 0,
		}
	}

	panic(fmt.Sprintf("unknown TA function in cache key: %s", cacheKey))
}

func (s *SMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return 0.0, err
		}

		idx := s.computed % s.period
		s.buffer[idx] = sourceVal
		s.computed++
	}

	if barIdx < s.period-1 {
		return 0.0, nil
	}

	sum := 0.0
	for i := 0; i < s.period; i++ {
		sum += s.buffer[i]
	}

	return sum / float64(s.period), nil
}

func (s *EMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return 0.0, err
		}

		if s.computed == 0 {
			s.prevEMA = sourceVal
		} else if s.computed < s.period {
			s.prevEMA = (s.prevEMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		} else {
			s.prevEMA = (sourceVal * s.multiplier) + (s.prevEMA * (1 - s.multiplier))
		}

		s.computed++
	}

	if barIdx < s.period-1 {
		return 0.0, nil
	}

	return s.prevEMA, nil
}

func (s *RMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return 0.0, err
		}

		if s.computed == 0 {
			s.prevRMA = sourceVal
		} else if s.computed < s.period {
			s.prevRMA = (s.prevRMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		} else {
			alpha := 1.0 / float64(s.period)
			s.prevRMA = alpha*sourceVal + (1-alpha)*s.prevRMA
		}

		s.computed++
	}

	if barIdx < s.period-1 {
		return 0.0, nil
	}

	return s.prevRMA, nil
}

func (s *RSIStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	if barIdx < s.period {
		return 0.0, nil
	}

	var prevSource float64
	if barIdx > 0 {
		val, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-1)
		if err != nil {
			return 0.0, err
		}
		prevSource = val
	}

	currentSource, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}

	change := currentSource - prevSource
	gain := 0.0
	loss := 0.0

	if change > 0 {
		gain = change
	} else {
		loss = -change
	}

	avgGain := s.rmaGain.prevRMA
	avgLoss := s.rmaLoss.prevRMA

	if s.computed == 0 {
		avgGain = gain
		avgLoss = loss
	} else if s.computed < s.period {
		avgGain = (avgGain*float64(s.computed) + gain) / float64(s.computed+1)
		avgLoss = (avgLoss*float64(s.computed) + loss) / float64(s.computed+1)
	} else {
		alpha := 1.0 / float64(s.period)
		avgGain = alpha*gain + (1-alpha)*avgGain
		avgLoss = alpha*loss + (1-alpha)*avgLoss
	}

	s.rmaGain.prevRMA = avgGain
	s.rmaLoss.prevRMA = avgLoss
	s.computed++

	if avgLoss == 0 {
		return 100.0, nil
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
