package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// TAStateManager guarantees sequential state accumulation via a catch-up loop,
// allowing arbitrary-barIdx queries via historical look-back inside security().
type TAStateManager interface {
	ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error)
}

// ── SMA ───────────────────────────────────────────────────────────────────────

type SMAStateManager struct {
	cacheKey  string
	period    int
	buf       forwardBufferE
	evaluator BarEvaluator
}

func newSMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *SMAStateManager {
	return &SMAStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *SMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		if bar < s.period-1 {
			return math.NaN(), nil
		}
		sum := 0.0
		for i := 0; i < s.period; i++ {
			v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar-s.period+1+i)
			if err != nil {
				return math.NaN(), err
			}
			sum += v
		}
		return sum / float64(s.period), nil
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}

// ── EMA ───────────────────────────────────────────────────────────────────────

type EMAStateManager struct {
	cacheKey   string
	period     int
	multiplier float64
	buf        forwardBufferE
	evaluator  BarEvaluator
}

func newEMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *EMAStateManager {
	return &EMAStateManager{
		cacheKey:   cacheKey,
		period:     period,
		multiplier: 2.0 / float64(period+1),
		buf:        newForwardBufferE(capacity),
		evaluator:  evaluator,
	}
}

func (s *EMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar)
		if err != nil {
			return math.NaN(), err
		}
		switch {
		case bar == 0:
			return v, nil
		case bar < s.period:
			return (s.buf.prev()*float64(bar) + v) / float64(bar+1), nil
		default:
			return v*s.multiplier + s.buf.prev()*(1-s.multiplier), nil
		}
	}); err != nil {
		return math.NaN(), err
	}
	if barIdx < s.period-1 {
		return math.NaN(), nil
	}
	return s.buf.at(barIdx), nil
}

// ── RMA ───────────────────────────────────────────────────────────────────────

type RMAStateManager struct {
	cacheKey  string
	period    int
	buf       forwardBufferE
	evaluator BarEvaluator
}

func newRMAStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *RMAStateManager {
	return &RMAStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *RMAStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar)
		if err != nil {
			return math.NaN(), err
		}
		switch {
		case bar == 0:
			return v, nil
		case bar < s.period:
			return (s.buf.prev()*float64(bar) + v) / float64(bar+1), nil
		default:
			alpha := 1.0 / float64(s.period)
			return alpha*v + (1-alpha)*s.buf.prev(), nil
		}
	}); err != nil {
		return math.NaN(), err
	}
	if barIdx < s.period-1 {
		return math.NaN(), nil
	}
	return s.buf.at(barIdx), nil
}

// ── RSI ───────────────────────────────────────────────────────────────────────

type RSIStateManager struct {
	cacheKey    string
	period      int
	prevAvgGain float64
	prevAvgLoss float64
	buf         forwardBufferE
	evaluator   BarEvaluator
}

func newRSIStateManager(cacheKey string, period, capacity int, evaluator BarEvaluator) *RSIStateManager {
	return &RSIStateManager{
		cacheKey:  cacheKey,
		period:    period,
		buf:       newForwardBufferE(capacity),
		evaluator: evaluator,
	}
}

func (s *RSIStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
		s.prevAvgGain = 0
		s.prevAvgLoss = 0
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		return s.evalBar(secCtx, sourceExpr, bar)
	}); err != nil {
		return math.NaN(), err
	}
	if barIdx < s.period {
		return math.NaN(), nil
	}
	return s.buf.at(barIdx), nil
}

func (s *RSIStateManager) evalBar(secCtx *context.Context, sourceExpr ast.Expression, bar int) (float64, error) {
	if bar < s.period {
		return math.NaN(), nil
	}
	prev, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar-1)
	if err != nil {
		return math.NaN(), err
	}
	curr, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar)
	if err != nil {
		return math.NaN(), err
	}
	change := curr - prev
	gain := math.Max(change, 0)
	loss := math.Max(-change, 0)
	storageIdx := bar - s.period
	var avgGain, avgLoss float64
	switch {
	case storageIdx == 0:
		avgGain, avgLoss = gain, loss
	case storageIdx < s.period:
		avgGain = (s.prevAvgGain*float64(storageIdx) + gain) / float64(storageIdx+1)
		avgLoss = (s.prevAvgLoss*float64(storageIdx) + loss) / float64(storageIdx+1)
	default:
		alpha := 1.0 / float64(s.period)
		avgGain = alpha*gain + (1-alpha)*s.prevAvgGain
		avgLoss = alpha*loss + (1-alpha)*s.prevAvgLoss
	}
	s.prevAvgGain = avgGain
	s.prevAvgLoss = avgLoss
	if avgLoss == 0 {
		return 100.0, nil
	}
	rs := avgGain / avgLoss
	return 100.0 - (100.0 / (1.0 + rs)), nil
}

// ── Factory ───────────────────────────────────────────────────────────────────

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
