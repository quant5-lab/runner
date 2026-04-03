package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/runtime/context"
)

// volumeFormula computes one bar's contribution to a cumulative volume indicator.
// prevValue is either the seed (bar 0) or the prior bar's accumulated value.
type volumeFormula func(bar, prevBar context.OHLCV, prevValue float64, isFirstBar bool) float64

// volumeIndicatorState mirrors the TAStateManager contract without requiring a
// source expression — volume indicators are computed exclusively from OHLCV.
type volumeIndicatorState struct {
	buf     forwardBufferE
	seed    float64
	formula volumeFormula
}

func newVolumeIndicatorState(capacity int, seed float64, formula volumeFormula) *volumeIndicatorState {
	return &volumeIndicatorState{
		buf:     newForwardBufferE(capacity),
		seed:    seed,
		formula: formula,
	}
}

func (s *volumeIndicatorState) computeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	if err := s.buf.advanceTo(barIdx, func(bar int) (float64, error) {
		isFirst := bar == 0
		current := secCtx.Data[bar]
		var prevBar context.OHLCV
		if !isFirst {
			prevBar = secCtx.Data[bar-1]
		}
		prev := s.seed
		if !isFirst {
			if p := s.buf.prev(); !math.IsNaN(p) {
				prev = p
			}
		}
		return s.formula(current, prevBar, prev, isFirst), nil
	}); err != nil {
		return math.NaN(), err
	}
	return s.buf.at(barIdx), nil
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
