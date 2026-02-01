package ticker

import "github.com/quant5-lab/runner/runtime/context"

/* BarTransformer converts standard OHLCV bars to modified chart types */
type BarTransformer interface {
	Transform(bars []context.OHLCV) []context.OHLCV
	Type() ModifierType
}

/* IdentityTransformer passes bars through unchanged */
type IdentityTransformer struct{}

func (t *IdentityTransformer) Transform(bars []context.OHLCV) []context.OHLCV {
	return bars
}

func (t *IdentityTransformer) Type() ModifierType {
	return ""
}

/* HeikinAshiTransformer converts standard bars to Heikin Ashi */
type HeikinAshiTransformer struct{}

func (t *HeikinAshiTransformer) Transform(bars []context.OHLCV) []context.OHLCV {
	if len(bars) == 0 {
		return bars
	}

	result := make([]context.OHLCV, len(bars))
	var prevHaOpen, prevHaClose float64

	for i, bar := range bars {
		haBar := CalculateHeikinAshiBar(bar, context.OHLCV{}, prevHaOpen, prevHaClose)
		result[i] = haBar
		prevHaOpen = haBar.Open
		prevHaClose = haBar.Close
	}

	return result
}

func (t *HeikinAshiTransformer) Type() ModifierType {
	return ModifierHeikinAshi
}

/* NewTransformer creates appropriate transformer for modifier type */
func NewTransformer(modifierType ModifierType) BarTransformer {
	switch modifierType {
	case ModifierHeikinAshi:
		return &HeikinAshiTransformer{}
	default:
		return &IdentityTransformer{}
	}
}
