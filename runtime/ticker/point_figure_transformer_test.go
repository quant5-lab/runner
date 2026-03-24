package ticker

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

// TestPointFigureTransformer_NoColumnsWithoutDirectionChange covers nil input and
// all-same-price input, both of which produce zero output columns.
func TestPointFigureTransformer_NoColumnsWithoutDirectionChange(t *testing.T) {
	tests := []struct {
		name   string
		prices []float64
	}{
		{"nil_input", nil},
		{"all_same_price", []float64{100, 100, 100}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
				Transform(pricesToBars(tt.prices))
			if len(result.Bars) != 0 {
				t.Errorf("expected 0 columns, got %d", len(result.Bars))
			}
		})
	}
}

// TestPointFigureTransformer_DevelopingColumnFlush verifies that a developing column
// is flushed as the final bar for both X (rising) and O (falling) directions.
func TestPointFigureTransformer_DevelopingColumnFlush(t *testing.T) {
	tests := []struct {
		name        string
		prices      []float64
		wantCols    int
		wantXColumn bool // true=X col (C>O), false=O col (C<O)
	}{
		{
			name:        "x_column_flushed",
			prices:      []float64{100, 120, 135},
			wantCols:    1,
			wantXColumn: true,
		},
		{
			name:        "o_column_flushed",
			prices:      []float64{100, 80, 65},
			wantCols:    1,
			wantXColumn: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
				Transform(pricesToBars(tt.prices))
			if len(result.Bars) != tt.wantCols {
				t.Fatalf("column count = %d, want %d", len(result.Bars), tt.wantCols)
			}
			col := result.Bars[0]
			if tt.wantXColumn && col.Close <= col.Open {
				t.Errorf("X column: Close=%v should be > Open=%v", col.Close, col.Open)
			}
			if !tt.wantXColumn && col.Close >= col.Open {
				t.Errorf("O column: Close=%v should be < Open=%v", col.Close, col.Open)
			}
		})
	}
}

// TestPointFigureTransformer_ColumnOHLC verifies the OHLC layout for both column types:
//
//	X column: O=colLow  C=colHigh H=C  L=O
//	O column: O=colHigh C=colLow  H=O  L=C
func TestPointFigureTransformer_ColumnOHLC(t *testing.T) {
	// X column: price rises from 100 to 135 — colLow becomes Open, colHigh becomes Close.
	xResult := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
		Transform(pricesToBars([]float64{100, 120, 135}))
	if len(xResult.Bars) < 1 {
		t.Fatalf("expected at least 1 X column")
	}
	xc := xResult.Bars[0]
	if xc.High != xc.Close {
		t.Errorf("X column High=%v, want Close=%v", xc.High, xc.Close)
	}
	if xc.Low != xc.Open {
		t.Errorf("X column Low=%v, want Open=%v", xc.Low, xc.Open)
	}

	// O column: price falls from 100 to 65 — colHigh becomes Open, colLow becomes Close.
	oResult := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
		Transform(pricesToBars([]float64{100, 80, 65}))
	if len(oResult.Bars) < 1 {
		t.Fatalf("expected at least 1 O column")
	}
	oc := oResult.Bars[0]
	if oc.High != oc.Open {
		t.Errorf("O column High=%v, want Open=%v", oc.High, oc.Open)
	}
	if oc.Low != oc.Close {
		t.Errorf("O column Low=%v, want Close=%v", oc.Low, oc.Close)
	}
}

// TestPointFigureTransformer_Reversal verifies that a reversal produces the correct
// number of columns and that column direction alternates correctly.
func TestPointFigureTransformer_Reversal(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		wantCols int
	}{
		{
			// X→O reversal: rise to 150, then drop 35 below column high.
			name:     "x_to_o_reversal",
			prices:   []float64{100, 150, 115},
			wantCols: 2,
		},
		{
			// O→X reversal: fall to 60, then rise 35 above column low.
			name:     "o_to_x_reversal",
			prices:   []float64{100, 60, 95},
			wantCols: 2,
		},
		{
			// Two reversals: X→O→X.
			name:     "two_reversals",
			prices:   []float64{100, 150, 115, 150},
			wantCols: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
				Transform(pricesToBars(tt.prices))
			if len(result.Bars) != tt.wantCols {
				t.Errorf("column count = %d, want %d", len(result.Bars), tt.wantCols)
			}
		})
	}
}

// TestPointFigureTransformer_HLSource verifies that source="hl" uses bar.High for
// X column advances and bar.Low for O column advances (instead of Close for both).
// Note: the undecided-direction transition always uses bar.Close regardless of source,
// so direction must be established first via a Close change.
func TestPointFigureTransformer_HLSource(t *testing.T) {
	// bar[1] establishes xCol direction via Close change; bar[2] extends the X column
	// to High=130 with hl source while Close stays at 102. Low=105 stays above the
	// reversal threshold (130 − 3×10 = 100) so no premature reversal fires.
	bars := []context.OHLCV{
		{Close: 100, High: 100, Low: 100},
		{Close: 102, High: 102, Low: 100},
		{Close: 102, High: 130, Low: 105},
	}
	hlResult := NewPointFigureTransformer("hl", PointFigStyleATR, 10, 3).Transform(bars)
	closeResult := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).Transform(bars)

	if len(hlResult.Bars) == 0 {
		t.Errorf("hl source: expected a flushed X column, got 0 columns")
	}
	if len(closeResult.Bars) == 0 {
		t.Fatalf("close source: expected a flushed X column, got 0 columns")
	}

	hlClose := hlResult.Bars[len(hlResult.Bars)-1].Close
	closeClose := closeResult.Bars[len(closeResult.Bars)-1].Close
	if hlClose <= closeClose {
		t.Errorf("hl source colExtreme=%v should exceed close source colExtreme=%v (High=130 vs Close=102)", hlClose, closeClose)
	}
}

// TestPointFigureTransformer_MappingPreFirstColumn verifies that source bars before
// any column forms are retroactively mapped to column 0 after -1 normalization.
func TestPointFigureTransformer_MappingPreFirstColumn(t *testing.T) {
	// Bars 0 and 1 are pre-column (undecided/developing); bar 2 triggers a reversal
	// completing the first column, making bars 0 and 1 point to column 0.
	// boxSize=10, reversal=3: rise to 150 (2 bars), drop to 115 (triggers X→O reversal).
	prices := []float64{100, 150, 115}
	result := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3).
		Transform(pricesToBars(prices))

	if len(result.Bars) < 2 {
		t.Fatalf("expected at least 2 columns (1 completed + 1 developing), got %d", len(result.Bars))
	}
	for i, idx := range result.MainToSynthetic {
		if idx < 0 {
			t.Errorf("MainToSynthetic[%d]=%d, want >= 0 (all -1 should be normalized to 0)", i, idx)
		}
	}
}

// TestPointFigureTransformer_MappingInvariants verifies structural mapping guarantees
// across multiple price sequences.
func TestPointFigureTransformer_MappingInvariants(t *testing.T) {
	tests := []struct {
		name     string
		prices   []float64
		boxSize  float64
		reversal float64
	}{
		{"x_developing", linspace(100, 15, 12), 10, 3},
		{"o_developing", linspace(200, -15, 12), 10, 3},
		{"multi_reversal", oscillate(100, 50, -40, 20), 10, 3},
		{"tight_reversal", oscillate(100, 20, -15, 20), 5, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, tt.boxSize, tt.reversal).
				Transform(pricesToBars(tt.prices))
			assertMappingLength(t, result, len(tt.prices))
			assertMappingNonDecreasing(t, result)
			if len(result.Bars) > 0 {
				assertMappingValidIndices(t, result)
				assertOHLCInvariants(t, result.Bars)
			}
		})
	}
}

func TestPointFigureTransformer_Type(t *testing.T) {
	tr := NewPointFigureTransformer(PointFigDefaultSource, PointFigStyleATR, 10, 3)
	if tr.Type() != ModifierPointFig {
		t.Errorf("Type() = %q, want %q", tr.Type(), ModifierPointFig)
	}
}
