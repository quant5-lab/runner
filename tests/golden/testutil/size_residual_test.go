package testutil

import (
	"math"
	"testing"
)

func TestSizeRelDev(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"zero/zero", 0, 0, 0},
		{"positive/equal", 1.0, 1.0, 0},
		{"large/equal", 1e6, 1e6, 0},
		{"tiny/equal", 1e-12, 1e-12, 0},

		{"zero_vs_positive", 0, 1.0, 1.0},
		{"zero_vs_large", 0, 1e6, 1.0},

		{"denominator_is_larger", 1.0, 2.0, 0.5},

		{"btc_equity_drift", 0.95355932, 0.95451287, (0.95451287 - 0.95355932) / 0.95451287},

		// position sizes are always ≥ 0 in runner output; these cases cover defined
		// mathematical behavior of the formula when inputs are negative
		{"negative/equal", -3.0, -3.0, 0},
		{"negative_vs_positive_same_magnitude", -1.0, 1.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sizeRelDev(tt.a, tt.b)
			if math.IsNaN(got) || math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("sizeRelDev(%v, %v) = %.15e, want %.15e", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSizeRelDev_Symmetry(t *testing.T) {
	pairs := [][2]float64{
		{0, 0},
		{1.0, 1.0},
		{0, 5.0},
		{3.0, 7.0},
		{0.001, 0.001 + 0.001*5e-4},
		{1000.0, 1000.0 + 0.3},
		{0.95355932, 0.95451287},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := sizeRelDev(a, b)
		rev := sizeRelDev(b, a)
		if fwd != rev {
			t.Errorf("sizeRelDev not symmetric: (%v,%v)=%.15e != (%v,%v)=%.15e",
				a, b, fwd, b, a, rev)
		}
	}
}

func TestSizeRelDev_ScaleProportionality(t *testing.T) {
	// The same fractional difference produces the same relative deviation
	// regardless of the absolute magnitude of the position size.
	// sizeRelDev(base, base*(1+r)) = r/(1+r) for all base > 0.
	const relDiff = 3.8e-4
	wantApprox := relDiff / (1 + relDiff)
	scales := []float64{0.001, 0.01, 1.0, 10.0, 1000.0, 1e6}
	for _, base := range scales {
		a := base
		b := base * (1 + relDiff)
		got := sizeRelDev(a, b)
		if math.Abs(got-wantApprox) > 1e-12 {
			t.Errorf("scale=%v: got %.15e, want %.15e", base, got, wantApprox)
		}
	}
}

func tradeWithSize(size float64) Trade {
	tr := baselineTrade()
	tr.Size = size
	return tr
}

func TestSizeResidualOverSlice(t *testing.T) {
	tests := []struct {
		name string
		a, b []Trade
		want float64
	}{
		{"both_nil", nil, nil, 0},
		{"both_empty", []Trade{}, []Trade{}, 0},
		{"a_nil_b_nonempty", nil, []Trade{tradeWithSize(1.0)}, 0},
		{"a_nonempty_b_nil", []Trade{tradeWithSize(1.0)}, nil, 0},

		// equal sizes
		{"single_equal", []Trade{tradeWithSize(1.0)}, []Trade{tradeWithSize(1.0)}, 0},
		{"multiple_equal",
			[]Trade{tradeWithSize(1.0), tradeWithSize(2.0)},
			[]Trade{tradeWithSize(1.0), tradeWithSize(2.0)},
			0,
		},

		// surplus trades beyond the shorter list are not compared
		{"a_longer_surplus_ignored",
			[]Trade{tradeWithSize(1.0), tradeWithSize(0)},
			[]Trade{tradeWithSize(1.0)},
			0,
		},
		{"b_longer_surplus_ignored",
			[]Trade{tradeWithSize(1.0)},
			[]Trade{tradeWithSize(1.0), tradeWithSize(0)},
			0,
		},

		// single differing trade
		{"single_zero_vs_one",
			[]Trade{tradeWithSize(0)},
			[]Trade{tradeWithSize(1.0)},
			1.0,
		},
		{"single_partial_diff",
			[]Trade{tradeWithSize(1.0)},
			[]Trade{tradeWithSize(2.0)},
			0.5,
		},

		{"max_at_last",
			[]Trade{tradeWithSize(1.0), tradeWithSize(1.0), tradeWithSize(1.0)},
			[]Trade{tradeWithSize(1.001), tradeWithSize(1.01), tradeWithSize(2.0)},
			0.5,
		},
		{"max_at_middle",
			[]Trade{tradeWithSize(1.0), tradeWithSize(1.0), tradeWithSize(1.0)},
			[]Trade{tradeWithSize(1.001), tradeWithSize(2.0), tradeWithSize(1.001)},
			0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sizeResidualOverSlice(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("got %.15e, want %.15e", got, tt.want)
			}
		})
	}
}

func resultWithSizes(closed, open []float64) *StrategyResult {
	r := &StrategyResult{}
	for _, s := range closed {
		r.Trades = append(r.Trades, tradeWithSize(s))
	}
	for _, s := range open {
		r.OpenTrades = append(r.OpenTrades, tradeWithSize(s))
	}
	return r
}

func TestSizeResidual(t *testing.T) {
	tests := []struct {
		name string
		a, b *StrategyResult
		want float64
	}{
		{
			"both_empty_results",
			resultWithSizes(nil, nil),
			resultWithSizes(nil, nil),
			0,
		},
		{
			"equal_closed_trades",
			resultWithSizes([]float64{1.0, 2.0}, nil),
			resultWithSizes([]float64{1.0, 2.0}, nil),
			0,
		},
		{
			"equal_open_trades",
			resultWithSizes(nil, []float64{1.0, 2.0}),
			resultWithSizes(nil, []float64{1.0, 2.0}),
			0,
		},
		{
			"deviation_in_closed_only",
			resultWithSizes([]float64{1.0}, nil),
			resultWithSizes([]float64{2.0}, nil),
			0.5,
		},
		{
			"deviation_in_open_only",
			resultWithSizes(nil, []float64{1.0}),
			resultWithSizes(nil, []float64{2.0}),
			0.5,
		},
		{
			"closed_deviation_exceeds_open",
			resultWithSizes([]float64{1.0, 1.0}, []float64{1.0}),
			resultWithSizes([]float64{1.0, 2.0}, []float64{1.001}),
			0.5,
		},
		{
			"open_deviation_exceeds_closed",
			resultWithSizes([]float64{1.0}, []float64{1.0, 1.0}),
			resultWithSizes([]float64{1.001}, []float64{1.0, 2.0}),
			0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SizeResidual(tt.a, tt.b)
			if math.Abs(got-tt.want) > 1e-12 {
				t.Errorf("got %.15e, want %.15e", got, tt.want)
			}
		})
	}
}

func TestSizeResidual_GlobalMaxAcrossAllTrades(t *testing.T) {
	// Verifies that SizeResidual returns the global maximum across both the
	// closed-trade and open-trade slices, and that the maximum is found even
	// when it is not at index 0 of either slice.
	a := resultWithSizes(
		[]float64{1.0, 1.0, 1.0},
		[]float64{1.0, 1.0},
	)
	b := resultWithSizes(
		[]float64{1.001, 1.01, 1.0}, // closed devs: ~1e-3, ~1e-2, 0
		[]float64{1.0, 2.0},         // open devs:   0, 0.5
	)
	got := SizeResidual(a, b)
	want := sizeRelDev(1.0, 2.0) // open[1] = 0.5 is the global max
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("got %.15e, want %.15e — open[1] should be the global max", got, want)
	}
}

func TestSizeResidual_SurplusTradesDoNotInfluenceResult(t *testing.T) {
	// Surplus trades in b beyond the length of a must not be compared.
	// If they were included, b's extra zero-size trade would produce a 1.0 deviation
	// and mask the actual max from the overlapping prefix.
	a := resultWithSizes([]float64{1.0}, nil)
	b := resultWithSizes([]float64{1.001, 999.0}, nil) // index 1 is surplus — must be ignored
	got := SizeResidual(a, b)
	want := sizeRelDev(1.0, 1.001)
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("got %.15e, want %.15e — surplus trade must not be compared", got, want)
	}
}
