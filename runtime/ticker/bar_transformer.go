package ticker

import "github.com/quant5-lab/runner/runtime/context"

// BarTransformer converts a slice of standard OHLCV bars into a synthetic chart
// representation and returns both the synthetic bars and a per-source-bar synthetic
// index mapping so the security evaluator can locate the correct synthetic bar for
// any main chart bar.
type BarTransformer interface {
	Transform(bars []context.OHLCV) TransformResult
	Type() ModifierType
}

// IdentityTransformer passes bars through unchanged with a 1:1 mapping.
type IdentityTransformer struct{}

func (t *IdentityTransformer) Transform(bars []context.OHLCV) TransformResult {
	return TransformResult{
		Bars:            bars,
		MainToSynthetic: identityMapping(len(bars)),
	}
}

func (t *IdentityTransformer) Type() ModifierType { return "" }

// HeikinAshiTransformer converts standard bars to Heikin-Ashi.
// Bar count is preserved 1:1 so the mapping is identity.
type HeikinAshiTransformer struct{}

func (t *HeikinAshiTransformer) Transform(bars []context.OHLCV) TransformResult {
	if len(bars) == 0 {
		return TransformResult{}
	}

	result := make([]context.OHLCV, len(bars))
	var prevHaOpen, prevHaClose float64

	for i, bar := range bars {
		haBar := CalculateHeikinAshiBar(bar, context.OHLCV{}, prevHaOpen, prevHaClose)
		result[i] = haBar
		prevHaOpen = haBar.Open
		prevHaClose = haBar.Close
	}

	return TransformResult{
		Bars:            result,
		MainToSynthetic: identityMapping(len(bars)),
	}
}

func (t *HeikinAshiTransformer) Type() ModifierType { return ModifierHeikinAshi }

// NewTransformer returns a BarTransformer for the given modifier type using default
// parameters.  Callers with known literal parameters should use the typed constructors
// (NewRenkoTransformer, NewKagiTransformer, …) directly for accurate results.
func NewTransformer(modifierType ModifierType) BarTransformer {
	switch modifierType {
	case ModifierHeikinAshi:
		return &HeikinAshiTransformer{}
	case ModifierRenko:
		return NewRenkoTransformer(RenkoStyleATR, RenkoDefaultBoxSize)
	case ModifierKagi:
		return NewKagiTransformer(KagiDefaultReversal)
	case ModifierLineBreak:
		return NewLineBreakTransformer(LineBreakDefaultLines)
	case ModifierPointFig:
		return NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, PointFigDefaultBoxSize, PointFigDefaultReversal)
	default:
		return &IdentityTransformer{}
	}
}
