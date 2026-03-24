package ticker

import "github.com/quant5-lab/runner/runtime/context"

const LineBreakDefaultLines = 3

// LineBreakTransformer converts standard OHLCV bars into N-line break bars.
//
// Algorithm (no future peeking):
//   - A new up line forms when close > highest close of the last N lines.
//   - A new down line forms when close < lowest close of the last N lines.
//   - The first source bar seeds the first line unconditionally.
//
// OHLC per line:
//
//	up:   O=prevClose  C=barClose  H=C  L=O
//	down: O=prevClose  C=barClose  H=O  L=C
type LineBreakTransformer struct {
	numberOfLines int
}

func NewLineBreakTransformer(numberOfLines int) *LineBreakTransformer {
	n := numberOfLines
	if n < 1 {
		n = LineBreakDefaultLines
	}
	return &LineBreakTransformer{numberOfLines: n}
}

func (t *LineBreakTransformer) Type() ModifierType { return ModifierLineBreak }

func (t *LineBreakTransformer) Transform(bars []context.OHLCV) TransformResult {
	if len(bars) == 0 {
		return TransformResult{}
	}

	lines := make([]context.OHLCV, 0, len(bars))
	mapping := make([]int, len(bars))

	// Seed the first line from bar 0.
	seedClose := bars[0].Close
	lines = append(lines, context.OHLCV{
		Time:   bars[0].Time,
		Open:   bars[0].Open,
		High:   max(bars[0].Open, seedClose),
		Low:    min(bars[0].Open, seedClose),
		Close:  seedClose,
		Volume: bars[0].Volume,
	})
	mapping[0] = 0

	// lineCloses tracks the last N line closes for break-level determination.
	lineCloses := []float64{seedClose}
	prevClose := seedClose

	for i := 1; i < len(bars); i++ {
		bar := bars[i]
		high := sliceMax(lineCloses)
		low := sliceMin(lineCloses)

		switch {
		case bar.Close > high:
			open, close := prevClose, bar.Close
			lines = append(lines, context.OHLCV{
				Time: bar.Time, Open: open, High: close, Low: open, Close: close, Volume: bar.Volume,
			})
			lineCloses = cappedAppend(lineCloses, close, t.numberOfLines)
			prevClose = close

		case bar.Close < low:
			open, close := prevClose, bar.Close
			lines = append(lines, context.OHLCV{
				Time: bar.Time, Open: open, High: open, Low: close, Close: close, Volume: bar.Volume,
			})
			lineCloses = cappedAppend(lineCloses, close, t.numberOfLines)
			prevClose = close
		}

		mapping[i] = len(lines) - 1
	}

	return TransformResult{Bars: lines, MainToSynthetic: mapping}
}

func sliceMax(s []float64) float64 {
	v := s[0]
	for _, x := range s[1:] {
		if x > v {
			v = x
		}
	}
	return v
}

func sliceMin(s []float64) float64 {
	v := s[0]
	for _, x := range s[1:] {
		if x < v {
			v = x
		}
	}
	return v
}

// cappedAppend appends v and retains only the last n elements.
func cappedAppend(s []float64, v float64, n int) []float64 {
	s = append(s, v)
	if len(s) > n {
		s = s[len(s)-n:]
	}
	return s
}
