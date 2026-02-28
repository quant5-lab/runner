package security

import (
	"fmt"
	"math"

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
	storage    TASeriesStorage
	multiplier float64
	computed   int
}

type RMAStateManager struct {
	cacheKey string
	period   int
	storage  TASeriesStorage
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
			storage:    NewSeriesStorage(capacity),
			multiplier: multiplier,
			computed:   0,
		}
	}

	if contains(cacheKey, "rma") {
		return &RMAStateManager{
			cacheKey: cacheKey,
			period:   period,
			storage:  NewSeriesStorage(capacity),
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
				storage:  NewSeriesStorage(capacity),
				computed: 0,
			},
			rmaLoss: &RMAStateManager{
				cacheKey: cacheKey + "_loss",
				period:   period,
				storage:  NewSeriesStorage(capacity),
				computed: 0,
			},
			computed: 0,
		}
	}

	if contains(cacheKey, "atr") {
		return NewATRStateManager(cacheKey, period, capacity)
	}

	if contains(cacheKey, "stdev") {
		return NewSTDEVStateManager(cacheKey, period)
	}

	panic(fmt.Sprintf("unknown TA function in cache key: %s", cacheKey))
}

func (s *SMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	/* Fill buffer up to requested bar */
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		idx := s.computed % s.period
		s.buffer[idx] = sourceVal
		s.computed++
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	/* Compute SMA using the last `period` bars ending at barIdx */
	sum := 0.0
	for i := 0; i < s.period; i++ {
		barOffset := barIdx - s.period + 1 + i
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, barOffset)
		if err != nil {
			return math.NaN(), err
		}
		sum += sourceVal
	}

	return sum / float64(s.period), nil
}

func (s *EMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		var emaValue float64
		if s.computed == 0 {
			emaValue = sourceVal
		} else if s.computed < s.period {
			prevEMA := s.storage.Get(s.computed - 1)
			emaValue = (prevEMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		} else {
			prevEMA := s.storage.Get(s.computed - 1)
			emaValue = (sourceVal * s.multiplier) + (prevEMA * (1 - s.multiplier))
		}

		s.storage.Set(s.computed, emaValue)
		s.computed++
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	return s.storage.Get(barIdx), nil
}

func (s *RMAStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		var rmaValue float64
		if s.computed == 0 {
			rmaValue = sourceVal
		} else if s.computed < s.period {
			prevRMA := s.storage.Get(s.computed - 1)
			rmaValue = (prevRMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		} else {
			alpha := 1.0 / float64(s.period)
			prevRMA := s.storage.Get(s.computed - 1)
			rmaValue = alpha*sourceVal + (1-alpha)*prevRMA
		}

		s.storage.Set(s.computed, rmaValue)
		s.computed++
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	return s.storage.Get(barIdx), nil
}

func (s *RSIStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	if barIdx < s.period {
		return math.NaN(), nil
	}

	var prevSource float64
	if barIdx > 0 {
		val, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx-1)
		if err != nil {
			return math.NaN(), err
		}
		prevSource = val
	}

	currentSource, err := evaluateOHLCVAtBar(sourceID, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}

	change := currentSource - prevSource
	gain := 0.0
	loss := 0.0

	if change > 0 {
		gain = change
	} else {
		loss = -change
	}

	var avgGain, avgLoss float64
	if s.computed == 0 {
		avgGain = gain
		avgLoss = loss
	} else if s.computed < s.period {
		prevAvgGain := s.rmaGain.storage.Get(s.computed - 1)
		prevAvgLoss := s.rmaLoss.storage.Get(s.computed - 1)
		avgGain = (prevAvgGain*float64(s.computed) + gain) / float64(s.computed+1)
		avgLoss = (prevAvgLoss*float64(s.computed) + loss) / float64(s.computed+1)
	} else {
		alpha := 1.0 / float64(s.period)
		prevAvgGain := s.rmaGain.storage.Get(s.computed - 1)
		prevAvgLoss := s.rmaLoss.storage.Get(s.computed - 1)
		avgGain = alpha*gain + (1-alpha)*prevAvgGain
		avgLoss = alpha*loss + (1-alpha)*prevAvgLoss
	}

	s.rmaGain.storage.Set(s.computed, avgGain)
	s.rmaLoss.storage.Set(s.computed, avgLoss)
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
