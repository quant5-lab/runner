package ticker

import "github.com/quant5-lab/runner/runtime/context"

const (
	PointFigDefaultSource    = "close"
	PointFigStyleATR         = "ATR"
	PointFigStyleTraditional = "Traditional"
	PointFigDefaultBoxSize   = 14.0
	PointFigDefaultReversal  = 3.0
)

// PointFigureTransformer converts standard OHLCV bars into Point & Figure columns.
//
// Algorithm (no future peeking):
//   - X column (rising): a new X box forms for each boxSize advance above the last X level.
//   - O column (falling): a new O box forms for each boxSize decline below the last O level.
//   - Column reversal: in an X column, when price drops >= reversal*boxSize from the
//     column high → close X column, open O column; vice versa for O columns.
//
// Source "close" uses the bar's close price; "hl" uses high for X advances and
// low for O advances.  Each completed or developing column is one OHLC bar:
//
//	X column: O=colLow   C=colHigh  H=C  L=O
//	O column: O=colHigh  C=colLow   H=O  L=C
type PointFigureTransformer struct {
	source   string
	style    string
	boxSize  float64
	reversal float64
}

func NewPointFigureTransformer(source, style string, boxSize, reversal float64) *PointFigureTransformer {
	return &PointFigureTransformer{
		source:   source,
		style:    style,
		boxSize:  boxSize,
		reversal: reversal,
	}
}

func (t *PointFigureTransformer) Type() ModifierType { return ModifierPointFig }

func (t *PointFigureTransformer) Transform(bars []context.OHLCV) TransformResult {
	if len(bars) == 0 {
		return TransformResult{}
	}

	columns := make([]context.OHLCV, 0, len(bars)/4+1)
	mapping := make([]int, len(bars))

	const xCol, oCol, undecided = 1, -1, 0
	direction := undecided
	colStart := bars[0].Close
	colExtreme := bars[0].Close
	colTime := bars[0].Time
	developingStart := 0
	lastClosed := -1

	priceFor := func(bar context.OHLCV, rising bool) float64 {
		if t.source == "hl" {
			if rising {
				return bar.High
			}
			return bar.Low
		}
		return bar.Close
	}

	appendColumn := func(barTime int64, volume float64) {
		var o, c, h, l float64
		if direction == xCol {
			o, c = colStart, colExtreme
			h, l = c, o
		} else {
			o, c = colStart, colExtreme
			h, l = o, c
		}
		columns = append(columns, context.OHLCV{
			Time: barTime, Open: o, High: h, Low: l, Close: c, Volume: volume,
		})
		lastClosed = len(columns) - 1
	}

	for i, bar := range bars {
		switch direction {
		case undecided:
			p := bar.Close
			if p > colStart {
				direction = xCol
				colExtreme = p
			} else if p < colStart {
				direction = oCol
				colExtreme = p
			}

		case xCol:
			advance := priceFor(bar, true)
			retreat := priceFor(bar, false)

			// Extend X column upward.
			if advance > colExtreme {
				colExtreme = advance
			}

			// Check for reversal: price drops reversal*boxSize from column high.
			if retreat <= colExtreme-t.reversal*t.boxSize {
				appendColumn(colTime, bar.Volume)
				colStart = colExtreme
				colExtreme = retreat
				colTime = bar.Time
				direction = oCol
				developingStart = i + 1
			}

		case oCol:
			advance := priceFor(bar, false)
			retreat := priceFor(bar, true)

			// Extend O column downward.
			if advance < colExtreme {
				colExtreme = advance
			}

			// Check for reversal: price rises reversal*boxSize from column low.
			if retreat >= colExtreme+t.reversal*t.boxSize {
				appendColumn(colTime, bar.Volume)
				colStart = colExtreme
				colExtreme = retreat
				colTime = bar.Time
				direction = xCol
				developingStart = i + 1
			}
		}

		mapping[i] = lastClosed
	}

	if direction == undecided {
		return TransformResult{Bars: columns, MainToSynthetic: mapping}
	}

	// Flush the developing column.
	appendColumn(colTime, 0)

	for i := developingStart; i < len(mapping); i++ {
		mapping[i] = lastClosed
	}

	for i := range mapping {
		if mapping[i] == -1 {
			mapping[i] = 0
		}
	}

	return TransformResult{Bars: columns, MainToSynthetic: mapping}
}
