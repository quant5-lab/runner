package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

// streamingEMA advances its series cursor on every update() call,
// matching the main-loop ForwardSeriesBuffer contract.
type streamingEMA struct {
	buf         *series.Series
	multiplier  float64
	period      int
	count       int
	initialized bool
}

func newStreamingEMA(period, capacity int) *streamingEMA {
	return &streamingEMA{
		buf:        series.NewSeries(capacity),
		multiplier: 2.0 / float64(period+1),
		period:     period,
	}
}

func (e *streamingEMA) update(value float64) float64 {
	if e.initialized {
		e.buf.Next()
	}
	e.initialized = true

	if math.IsNaN(value) {
		e.buf.Set(math.NaN())
		return math.NaN()
	}

	e.count++

	var result float64
	switch {
	case e.count == 1:
		result = value
	case e.count <= e.period:
		prevEMA := e.buf.Get(1)
		result = (prevEMA*float64(e.count-1) + value) / float64(e.count)
	default:
		prevEMA := e.buf.Get(1)
		result = value*e.multiplier + prevEMA*(1-e.multiplier)
	}

	e.buf.Set(result)

	if e.count < e.period {
		return math.NaN()
	}
	return result
}

type TSIStateManager struct {
	cacheKey    string
	shortPeriod int
	longPeriod  int
	ema1Mom     *streamingEMA
	ema1Abs     *streamingEMA
	ema2Mom     *streamingEMA
	ema2Abs     *streamingEMA
	prevSource  float64
	buf         forwardBufferE
	evaluator   BarEvaluator
}

func NewTSIStateManager(cacheKey string, shortPeriod, longPeriod, capacity int, evaluator BarEvaluator) *TSIStateManager {
	return &TSIStateManager{
		cacheKey:    cacheKey,
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		ema1Mom:     newStreamingEMA(longPeriod, capacity),
		ema1Abs:     newStreamingEMA(longPeriod, capacity),
		ema2Mom:     newStreamingEMA(shortPeriod, capacity),
		ema2Abs:     newStreamingEMA(shortPeriod, capacity),
		buf:         newForwardBufferE(capacity),
		evaluator:   evaluator,
	}
}

func (s *TSIStateManager) ComputeAtBar(secCtx *context.Context, sourceExpr ast.Expression, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
		s.prevSource = 0
		s.ema1Mom = newStreamingEMA(s.longPeriod, len(secCtx.Data))
		s.ema1Abs = newStreamingEMA(s.longPeriod, len(secCtx.Data))
		s.ema2Mom = newStreamingEMA(s.shortPeriod, len(secCtx.Data))
		s.ema2Abs = newStreamingEMA(s.shortPeriod, len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		return s.tsiAtBar(secCtx, sourceExpr, bar)
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
}

func (s *TSIStateManager) tsiAtBar(secCtx *context.Context, sourceExpr ast.Expression, bar int) (float64, error) {
	v, err := s.evaluator.EvaluateAtBar(sourceExpr, secCtx, bar)
	if err != nil {
		return math.NaN(), err
	}
	var momentum float64
	if bar == 0 {
		momentum = math.NaN()
	} else {
		momentum = v - s.prevSource
	}
	s.prevSource = v
	momentumAbs := math.NaN()
	if !math.IsNaN(momentum) {
		momentumAbs = math.Abs(momentum)
	}
	ema1MomVal := s.ema1Mom.update(momentum)
	ema1AbsVal := s.ema1Abs.update(momentumAbs)
	ema2MomVal := s.ema2Mom.update(ema1MomVal)
	ema2AbsVal := s.ema2Abs.update(ema1AbsVal)
	warmup := s.longPeriod + s.shortPeriod - 1
	switch {
	case bar < warmup:
		return math.NaN(), nil
	case ema2AbsVal == 0.0:
		return 0.0, nil
	default:
		return 100.0 * ema2MomVal / ema2AbsVal, nil
	}
}
