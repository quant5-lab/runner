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
	sourceBuf   *series.Series
	resultBuf   *series.Series
	computed    int
}

func NewTSIStateManager(cacheKey string, shortPeriod, longPeriod, capacity int) *TSIStateManager {
	return &TSIStateManager{
		cacheKey:    cacheKey,
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		ema1Mom:     newStreamingEMA(longPeriod, capacity),
		ema1Abs:     newStreamingEMA(longPeriod, capacity),
		ema2Mom:     newStreamingEMA(shortPeriod, capacity),
		ema2Abs:     newStreamingEMA(shortPeriod, capacity),
		sourceBuf:   series.NewSeries(capacity),
		resultBuf:   series.NewSeries(capacity),
	}
}

func (s *TSIStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.sourceBuf.Next()
			s.resultBuf.Next()
		}

		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		s.sourceBuf.Set(sourceVal)

		var momentum float64
		if s.computed == 0 {
			momentum = math.NaN()
		} else {
			momentum = sourceVal - s.sourceBuf.Get(1)
		}

		momentumAbs := math.NaN()
		if !math.IsNaN(momentum) {
			momentumAbs = math.Abs(momentum)
		}

		ema1MomValue := s.ema1Mom.update(momentum)
		ema1AbsValue := s.ema1Abs.update(momentumAbs)
		ema2MomValue := s.ema2Mom.update(ema1MomValue)
		ema2AbsValue := s.ema2Abs.update(ema1AbsValue)

		warmup := s.longPeriod + s.shortPeriod - 1

		var tsiValue float64
		switch {
		case s.computed < warmup:
			tsiValue = math.NaN()
		case ema2AbsValue == 0.0:
			tsiValue = 0.0
		default:
			tsiValue = 100.0 * ema2MomValue / ema2AbsValue
		}

		s.resultBuf.Set(tsiValue)
		s.computed++
	}

	return s.resultBuf.Get(s.resultBuf.Position() - barIdx), nil
}
