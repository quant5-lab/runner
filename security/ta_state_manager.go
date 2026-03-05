package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

// TAStateManager guarantees sequential state accumulation via a catch-up loop,
// allowing arbitrary-barIdx queries via historical look-back inside security().
type TAStateManager interface {
	ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error)
}

// ── SMA ────────────────────────────────────────────────────────────────────────

type SMAStateManager struct {
	cacheKey  string
	period    int
	buf       *series.Series
	computed  int
	evaluator BarEvaluator
}

func newSMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *SMAStateManager {
	return &SMAStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       series.NewSeries(max(capacity, 1)),
		evaluator: evaluator,
	}
}

func (s *SMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		if s.computed < s.period-1 {
			s.buf.Set(math.NaN())
		} else {
			sum := 0.0
			for i := 0; i < s.period; i++ {
				val, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, s.computed-s.period+1+i)
				if err != nil {
					return math.NaN(), err
				}
				sum += val
			}
			s.buf.Set(sum / float64(s.period))
		}

		s.computed++
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}

// ── EMA ────────────────────────────────────────────────────────────────────────

type EMAStateManager struct {
	cacheKey   string
	period     int
	buf        *series.Series
	multiplier float64
	computed   int
	evaluator  BarEvaluator
}

func newEMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *EMAStateManager {
	return &EMAStateManager{
		cacheKey:   cacheKey,
		period:     period,
		buf:        series.NewSeries(max(capacity, 1)),
		multiplier: 2.0 / float64(period+1),
		evaluator:  evaluator,
	}
}

func (s *EMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		sourceVal, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		var emaValue float64
		switch {
		case s.computed == 0:
			emaValue = sourceVal
		case s.computed < s.period:
			prevEMA := s.buf.Get(1)
			emaValue = (prevEMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		default:
			prevEMA := s.buf.Get(1)
			emaValue = (sourceVal * s.multiplier) + (prevEMA * (1 - s.multiplier))
		}

		s.buf.Set(emaValue)
		s.computed++
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}

// ── RMA ────────────────────────────────────────────────────────────────────────

type RMAStateManager struct {
	cacheKey  string
	period    int
	buf       *series.Series
	computed  int
	evaluator BarEvaluator
}

func newRMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *RMAStateManager {
	return &RMAStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       series.NewSeries(max(capacity, 1)),
		evaluator: evaluator,
	}
}

func (s *RMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		sourceVal, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		var rmaValue float64
		switch {
		case s.computed == 0:
			rmaValue = sourceVal
		case s.computed < s.period:
			prevRMA := s.buf.Get(1)
			rmaValue = (prevRMA*float64(s.computed) + sourceVal) / float64(s.computed+1)
		default:
			alpha := 1.0 / float64(s.period)
			prevRMA := s.buf.Get(1)
			rmaValue = alpha*sourceVal + (1-alpha)*prevRMA
		}

		s.buf.Set(rmaValue)
		s.computed++
	}

	if barIdx < s.period-1 {
		return math.NaN(), nil
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}

// ── RSI ────────────────────────────────────────────────────────────────────────

type RSIStateManager struct {
	cacheKey  string
	period    int
	gainBuf   *series.Series
	lossBuf   *series.Series
	computed  int
	evaluator BarEvaluator
}

func newRSIStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *RSIStateManager {
	return &RSIStateManager{
		cacheKey:  cacheKey,
		period:    period,
		gainBuf:   series.NewSeries(max(capacity, 1)),
		lossBuf:   series.NewSeries(max(capacity, 1)),
		evaluator: evaluator,
	}
}

func (s *RSIStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed < s.period {
			s.computed++
			continue
		}

		if s.computed > s.period {
			s.gainBuf.Next()
			s.lossBuf.Next()
		}

		prevSource, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, s.computed-1)
		if err != nil {
			return math.NaN(), err
		}

		currentSource, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		change := currentSource - prevSource
		gain := math.Max(change, 0)
		loss := math.Max(-change, 0)

		storageIdx := s.computed - s.period

		var avgGain, avgLoss float64
		switch {
		case storageIdx == 0:
			avgGain = gain
			avgLoss = loss
		case storageIdx < s.period:
			prevAvgGain := s.gainBuf.Get(1)
			prevAvgLoss := s.lossBuf.Get(1)
			avgGain = (prevAvgGain*float64(storageIdx) + gain) / float64(storageIdx+1)
			avgLoss = (prevAvgLoss*float64(storageIdx) + loss) / float64(storageIdx+1)
		default:
			alpha := 1.0 / float64(s.period)
			prevAvgGain := s.gainBuf.Get(1)
			prevAvgLoss := s.lossBuf.Get(1)
			avgGain = alpha*gain + (1-alpha)*prevAvgGain
			avgLoss = alpha*loss + (1-alpha)*prevAvgLoss
		}

		s.gainBuf.Set(avgGain)
		s.lossBuf.Set(avgLoss)
		s.computed++
	}

	if barIdx < s.period {
		return math.NaN(), nil
	}

	storageIdx := barIdx - s.period
	offset := s.gainBuf.Position() - storageIdx
	avgGain := s.gainBuf.Get(offset)
	avgLoss := s.lossBuf.Get(offset)

	if avgLoss == 0 {
		return 100.0, nil
	}

	rs := avgGain / avgLoss
	return 100.0 - (100.0 / (1.0 + rs)), nil
}

// ── Factory ────────────────────────────────────────────────────────────────────

func NewTAStateManager(cacheKey string, period int, capacity int, evaluator BarEvaluator) TAStateManager {
	switch {
	case contains(cacheKey, "sma"):
		return newSMAStateManager(cacheKey, period, capacity, evaluator)
	case contains(cacheKey, "ema"):
		return newEMAStateManager(cacheKey, period, capacity, evaluator)
	case contains(cacheKey, "rma"):
		return newRMAStateManager(cacheKey, period, capacity, evaluator)
	case contains(cacheKey, "rsi"):
		return newRSIStateManager(cacheKey, period, capacity, evaluator)
	case contains(cacheKey, "atr"):
		return NewATRStateManager(cacheKey, period, capacity)
	case contains(cacheKey, "stdev"):
		return NewSTDEVStateManager(cacheKey, period, capacity, evaluator)
	default:
		panic(fmt.Sprintf("unknown TA function in cache key: %s", cacheKey))
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
