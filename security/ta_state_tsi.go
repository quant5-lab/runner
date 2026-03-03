package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type TSIStateManager struct {
	cacheKey     string
	shortPeriod  int
	longPeriod   int
	ema1Mom      *streamingEMAWithStorage
	ema1Abs      *streamingEMAWithStorage
	ema2Mom      *streamingEMAWithStorage
	ema2Abs      *streamingEMAWithStorage
	tsiResults   TASeriesStorage
	sourceBuffer TASeriesStorage
	computed     int
}

type streamingEMAWithStorage struct {
	storage    TASeriesStorage
	multiplier float64
	period     int
	count      int
}

func newStreamingEMAWithStorage(period int, capacity int) *streamingEMAWithStorage {
	return &streamingEMAWithStorage{
		storage:    NewSeriesStorage(capacity),
		multiplier: 2.0 / float64(period+1),
		period:     period,
		count:      0,
	}
}

func (e *streamingEMAWithStorage) updateAndStore(barIdx int, inputValue float64) float64 {
	if math.IsNaN(inputValue) {
		result := math.NaN()
		e.storage.Set(barIdx, result)
		return result
	}

	e.count++
	var emaValue float64

	if e.count == 1 {
		emaValue = inputValue
	} else if e.count <= e.period {
		prevEMA := e.storage.Get(barIdx - 1)
		emaValue = (prevEMA*float64(e.count-1) + inputValue) / float64(e.count)
	} else {
		prevEMA := e.storage.Get(barIdx - 1)
		emaValue = inputValue*e.multiplier + prevEMA*(1-e.multiplier)
	}

	e.storage.Set(barIdx, emaValue)

	if e.count < e.period {
		return math.NaN()
	}
	return emaValue
}

func NewTSIStateManager(cacheKey string, shortPeriod, longPeriod int) *TSIStateManager {
	return &TSIStateManager{
		cacheKey:     cacheKey,
		shortPeriod:  shortPeriod,
		longPeriod:   longPeriod,
		ema1Mom:      newStreamingEMAWithStorage(longPeriod, 5000),
		ema1Abs:      newStreamingEMAWithStorage(longPeriod, 5000),
		ema2Mom:      newStreamingEMAWithStorage(shortPeriod, 5000),
		ema2Abs:      newStreamingEMAWithStorage(shortPeriod, 5000),
		tsiResults:   NewSeriesStorage(5000),
		sourceBuffer: NewSeriesStorage(5000),
		computed:     0,
	}
}

func (s *TSIStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}
		s.sourceBuffer.Set(s.computed, sourceVal)

		var momentum float64
		if s.computed == 0 {
			momentum = math.NaN()
		} else {
			prevSource := s.sourceBuffer.Get(s.computed - 1)
			momentum = sourceVal - prevSource
		}

		momentumAbs := math.NaN()
		if !math.IsNaN(momentum) {
			momentumAbs = math.Abs(momentum)
		}

		ema1MomValue := s.ema1Mom.updateAndStore(s.computed, momentum)
		ema1AbsValue := s.ema1Abs.updateAndStore(s.computed, momentumAbs)
		ema2MomValue := s.ema2Mom.updateAndStore(s.computed, ema1MomValue)
		ema2AbsValue := s.ema2Abs.updateAndStore(s.computed, ema1AbsValue)

		warmup := s.longPeriod + s.shortPeriod - 1
		var tsiValue float64
		if s.computed < warmup {
			tsiValue = math.NaN()
		} else if ema2AbsValue == 0.0 {
			tsiValue = 0.0
		} else {
			tsiValue = 100.0 * ema2MomValue / ema2AbsValue
		}

		s.tsiResults.Set(s.computed, tsiValue)
		s.computed++
	}

	return s.tsiResults.Get(barIdx), nil
}
