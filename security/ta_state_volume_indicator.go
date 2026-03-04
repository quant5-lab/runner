package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

// volumeFormula computes one bar's contribution to a cumulative volume indicator.
// prevValue is the seed on the very first bar for indicators that seed at 0;
// for nvi/pvi it is pre-seeded to 1000 by the factory.
type volumeFormula func(bar, prevBar context.OHLCV, prevValue float64, isFirstBar bool) float64

// volumeIndicatorState mirrors the TAStateManager contract without requiring a
// source Identifier — volume indicators are computed exclusively from OHLCV.
type volumeIndicatorState struct {
	buf      *series.Series
	computed int
	seed     float64
	formula  volumeFormula
}

func newVolumeIndicatorState(capacity int, seed float64, formula volumeFormula) *volumeIndicatorState {
	return &volumeIndicatorState{
		buf:     series.NewSeries(capacity),
		seed:    seed,
		formula: formula,
	}
}

func (s *volumeIndicatorState) computeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		bar := secCtx.Data[s.computed]
		isFirstBar := s.computed == 0

		var prevBar context.OHLCV
		if !isFirstBar {
			prevBar = secCtx.Data[s.computed-1]
		}

		prev := s.seed
		if !isFirstBar {
			prev = s.buf.Get(1)
			if math.IsNaN(prev) {
				prev = s.seed
			}
		}

		s.buf.Set(s.formula(bar, prevBar, prev, isFirstBar))
		s.computed++
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}

var volumeIndicatorFactories = map[string]func(capacity int) *volumeIndicatorState{
	"obv":     newOBVState,
	"accdist": newAccdistState,
	"pvt":     newPVTState,
	"iii":     newIIIState,
	"wvad":    newWVADState,
	"nvi":     newNVIState,
	"pvi":     newPVIState,
	"wad":     newWADState,
}

func newVolumeState(propName string, capacity int) (*volumeIndicatorState, error) {
	factory, ok := volumeIndicatorFactories[propName]
	if !ok {
		return nil, fmt.Errorf("unknown volume indicator: ta.%s", propName)
	}
	return factory(capacity), nil
}

func newOBVState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, prevBar context.OHLCV, prev float64, isFirstBar bool) float64 {
		if isFirstBar {
			return 0
		}
		if bar.Close > prevBar.Close {
			return prev + bar.Volume
		}
		if bar.Close < prevBar.Close {
			return prev - bar.Volume
		}
		return prev
	})
}

func newAccdistState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, _ context.OHLCV, prev float64, _ bool) float64 {
		hl := bar.High - bar.Low
		if hl == 0 {
			return prev
		}
		clv := ((bar.Close - bar.Low) - (bar.High - bar.Close)) / hl
		return prev + clv*bar.Volume
	})
}

func newPVTState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, prevBar context.OHLCV, prev float64, isFirstBar bool) float64 {
		if isFirstBar || prevBar.Close == 0 {
			return prev
		}
		return prev + (bar.Close-prevBar.Close)/prevBar.Close*bar.Volume
	})
}

func newIIIState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, _ context.OHLCV, prev float64, _ bool) float64 {
		hl := bar.High - bar.Low
		if hl == 0 || bar.Volume == 0 {
			return prev
		}
		return prev + (2*bar.Close-bar.High-bar.Low)/(hl*bar.Volume)
	})
}

func newWVADState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, _ context.OHLCV, prev float64, _ bool) float64 {
		hl := bar.High - bar.Low
		if hl == 0 {
			return prev
		}
		return prev + (bar.Close-bar.Open)/hl*bar.Volume
	})
}

func newNVIState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 1000, func(bar, prevBar context.OHLCV, prev float64, isFirstBar bool) float64 {
		if isFirstBar {
			return 1000
		}
		if bar.Volume < prevBar.Volume && prevBar.Close != 0 {
			return prev * (1 + (bar.Close-prevBar.Close)/prevBar.Close)
		}
		return prev
	})
}

func newPVIState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 1000, func(bar, prevBar context.OHLCV, prev float64, isFirstBar bool) float64 {
		if isFirstBar {
			return 1000
		}
		if bar.Volume > prevBar.Volume && prevBar.Close != 0 {
			return prev * (1 + (bar.Close-prevBar.Close)/prevBar.Close)
		}
		return prev
	})
}

func newWADState(capacity int) *volumeIndicatorState {
	return newVolumeIndicatorState(capacity, 0, func(bar, prevBar context.OHLCV, prev float64, isFirstBar bool) float64 {
		if isFirstBar {
			return 0
		}
		trueHigh := math.Max(bar.High, prevBar.Close)
		trueLow := math.Min(bar.Low, prevBar.Close)
		if bar.Close > prevBar.Close {
			return prev + bar.Close - trueLow
		}
		if bar.Close < prevBar.Close {
			return prev + bar.Close - trueHigh
		}
		return prev
	})
}
