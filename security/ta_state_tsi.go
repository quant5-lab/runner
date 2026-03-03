package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/* TSIStateManager streams TSI via double-smoothed EMA chains: ema2/ema2Abs * 100 */
type TSIStateManager struct {
	cacheKey    string
	shortPeriod int
	longPeriod  int
	ema1Mom     *emaState
	ema1Abs     *emaState
	ema2Mom     *emaState
	ema2Abs     *emaState
	prevSource  float64
	computed    int
}

type emaState struct {
	prevEMA    float64
	multiplier float64
	period     int
	count      int
}

func newEMAState(period int) *emaState {
	return &emaState{
		multiplier: 2.0 / float64(period+1),
		period:     period,
	}
}

/* SMA-seeded EMA — matches EMAStateManager; NaN until seeded. */
func (e *emaState) update(val float64) float64 {
	if math.IsNaN(val) {
		return math.NaN()
	}
	e.count++
	if e.count == 1 {
		e.prevEMA = val
	} else if e.count <= e.period {
		e.prevEMA = (e.prevEMA*float64(e.count-1) + val) / float64(e.count)
	} else {
		e.prevEMA = val*e.multiplier + e.prevEMA*(1-e.multiplier)
	}
	if e.count < e.period {
		return math.NaN()
	}
	return e.prevEMA
}

func NewTSIStateManager(cacheKey string, shortPeriod, longPeriod int) *TSIStateManager {
	return &TSIStateManager{
		cacheKey:    cacheKey,
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
		ema1Mom:     newEMAState(longPeriod),
		ema1Abs:     newEMAState(longPeriod),
		ema2Mom:     newEMAState(shortPeriod),
		ema2Abs:     newEMAState(shortPeriod),
		computed:    0,
	}
}

func (s *TSIStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		sourceVal, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		var mom float64
		if s.computed == 0 {
			mom = math.NaN()
		} else {
			mom = sourceVal - s.prevSource
		}
		s.prevSource = sourceVal

		var momAbs float64
		if math.IsNaN(mom) {
			momAbs = math.NaN()
		} else {
			momAbs = math.Abs(mom)
		}

		e1m := s.ema1Mom.update(mom)
		e1a := s.ema1Abs.update(momAbs)
		s.ema2Mom.update(e1m)
		s.ema2Abs.update(e1a)

		s.computed++
	}

	warmup := s.longPeriod + s.shortPeriod - 1
	if barIdx < warmup {
		return math.NaN(), nil
	}

	e2a := s.ema2Abs.prevEMA
	if e2a == 0.0 {
		return 0.0, nil
	}

	return 100.0 * s.ema2Mom.prevEMA / e2a, nil
}
