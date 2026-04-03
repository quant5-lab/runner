package ticker

import (
	"math"

	"github.com/quant5-lab/runner/runtime/context"
)

const (
	RenkoStyleATR           = "ATR"
	RenkoStyleTraditional   = "Traditional"
	RenkoStylePercentageLTP = "PercentageLTP"
	RenkoDefaultBoxSize     = 14.0
)

// RenkoTransformer converts standard OHLCV bars into Renko bricks.
//
// Price-level algorithm (no future peeking):
//   - An up brick forms when close >= lastLevel + boxSize.
//   - A down brick forms when close <= lastLevel − boxSize.
//   - Multiple bricks may form in one source bar when price moves several box sizes.
//
// OHLC per brick:
//
//	up:   O=lastLevel  C=lastLevel+boxSize  H=max(C,barHigh)  L=min(O,barLow)
//	down: O=lastLevel  C=lastLevel−boxSize  H=max(O,barHigh)  L=min(C,barLow)
type RenkoTransformer struct {
	style   string
	boxSize float64
}

func NewRenkoTransformer(style string, boxSize float64) *RenkoTransformer {
	return &RenkoTransformer{style: style, boxSize: boxSize}
}

func (t *RenkoTransformer) Type() ModifierType { return ModifierRenko }

func (t *RenkoTransformer) Transform(bars []context.OHLCV) TransformResult {
	if len(bars) == 0 {
		return TransformResult{}
	}

	bricks := make([]context.OHLCV, 0, len(bars)/4+1)
	mapping := make([]int, len(bars))
	lastLevel := bars[0].Close
	lastSyntheticIdx := -1

	for i, bar := range bars {
		for bar.Close >= lastLevel+t.boxSize {
			open, close := lastLevel, lastLevel+t.boxSize
			bricks = append(bricks, context.OHLCV{
				Time:   bar.Time,
				Open:   open,
				High:   math.Max(close, bar.High),
				Low:    math.Min(open, bar.Low),
				Close:  close,
				Volume: bar.Volume,
			})
			lastLevel = close
			lastSyntheticIdx = len(bricks) - 1
		}
		for bar.Close <= lastLevel-t.boxSize {
			open, close := lastLevel, lastLevel-t.boxSize
			bricks = append(bricks, context.OHLCV{
				Time:   bar.Time,
				Open:   open,
				High:   math.Max(open, bar.High),
				Low:    math.Min(close, bar.Low),
				Close:  close,
				Volume: bar.Volume,
			})
			lastLevel = close
			lastSyntheticIdx = len(bricks) - 1
		}
		mapping[i] = lastSyntheticIdx
	}

	return TransformResult{Bars: bricks, MainToSynthetic: mapping}
}
