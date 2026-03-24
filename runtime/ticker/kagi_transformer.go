package ticker

import "github.com/quant5-lab/runner/runtime/context"

const KagiDefaultReversal = 1.0

// KagiTransformer converts standard OHLCV bars into Kagi line segments.
//
// Algorithm (no future peeking):
//   - In an up-trend: extend when close rises; reverse when close drops
//     >= reversalAmount below the current segment extreme.
//   - In a down-trend: extend when close falls; reverse when close rises
//     >= reversalAmount above the current segment extreme.
//
// Each completed reversal emits one segment bar:
//
//	up:   O=segStart  C=segHigh  H=C  L=O
//	down: O=segStart  C=segLow   H=O  L=C
//
// The ongoing developing segment is flushed as the final bar so indicators
// computed against Kagi data receive a value on every source bar.
type KagiTransformer struct {
	reversalAmount float64
}

func NewKagiTransformer(reversalAmount float64) *KagiTransformer {
	return &KagiTransformer{reversalAmount: reversalAmount}
}

func (t *KagiTransformer) Type() ModifierType { return ModifierKagi }

func (t *KagiTransformer) Transform(bars []context.OHLCV) TransformResult {
	if len(bars) == 0 {
		return TransformResult{}
	}

	segments := make([]context.OHLCV, 0, len(bars)/5+1)
	mapping := make([]int, len(bars))

	const up, down, undecided = 1, -1, 0
	direction := undecided
	segStart := bars[0].Close
	extreme := bars[0].Close
	segTime := bars[0].Time
	developingStart := 0 // first source-bar index of the current developing segment
	lastClosed := -1     // index of the last segment appended to segments

	appendSegment := func(barTime int64, volume float64) {
		var o, c, h, l float64
		if direction == up {
			o, c = segStart, extreme
			h, l = c, o
		} else {
			o, c = segStart, extreme
			h, l = o, c
		}
		segments = append(segments, context.OHLCV{
			Time: barTime, Open: o, High: h, Low: l, Close: c, Volume: volume,
		})
		lastClosed = len(segments) - 1
	}

	for i, bar := range bars {
		switch direction {
		case undecided:
			if bar.Close > segStart {
				direction = up
				extreme = bar.Close
			} else if bar.Close < segStart {
				direction = down
				extreme = bar.Close
			}

		case up:
			if bar.Close >= extreme {
				extreme = bar.Close
			} else if bar.Close <= extreme-t.reversalAmount {
				appendSegment(segTime, bar.Volume)
				segStart = extreme
				extreme = bar.Close
				segTime = bar.Time
				direction = down
				developingStart = i + 1
			}

		case down:
			if bar.Close <= extreme {
				extreme = bar.Close
			} else if bar.Close >= extreme+t.reversalAmount {
				appendSegment(segTime, bar.Volume)
				segStart = extreme
				extreme = bar.Close
				segTime = bar.Time
				direction = up
				developingStart = i + 1
			}
		}

		mapping[i] = lastClosed
	}

	if direction == undecided {
		// All bars had identical closes — no segments to emit.
		return TransformResult{Bars: segments, MainToSynthetic: mapping}
	}

	// Flush the developing (unflushed) segment.
	appendSegment(segTime, 0)

	// Bars that belong to the developing segment now point to the flushed segment.
	for i := developingStart; i < len(mapping); i++ {
		mapping[i] = lastClosed
	}

	// Bars before the first completed reversal (mapping == -1) point to the first segment.
	for i := range mapping {
		if mapping[i] == -1 {
			mapping[i] = 0
		}
	}

	return TransformResult{Bars: segments, MainToSynthetic: mapping}
}
