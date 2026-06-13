package testutil

import (
	"math"
	"testing"
)

func TestMatchRelative(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		eps  float64
		want bool
	}{
		{"equal_zero", 0, 0, 1e-9, true},
		{"equal_positive", 5.0, 5.0, 1e-9, true},
		{"equal_negative", -7.0, -7.0, 1e-9, true},
		{"equal_large", 1e12, 1e12, 1e-9, true},
		{"equal_tiny", 1e-15, 1e-15, 1e-9, true},

		{"both_nan", math.NaN(), math.NaN(), 1e-9, true},
		{"nan_vs_zero", math.NaN(), 0, 1e-9, false},
		{"nan_vs_positive", math.NaN(), 1.0, 1e-9, false},
		{"positive_vs_nan", 1.0, math.NaN(), 1e-9, false},

		{"both_pos_inf", math.Inf(1), math.Inf(1), 1e-9, true},
		{"both_neg_inf", math.Inf(-1), math.Inf(-1), 1e-9, true},
		{"pos_inf_vs_neg_inf", math.Inf(1), math.Inf(-1), 1e-9, false},
		{"pos_inf_vs_finite", math.Inf(1), 1e300, 1e-9, false},

		{"zero_vs_one", 0, 1.0, 1e-9, false},
		{"zero_vs_tiny", 0, 1e-15, 1e-9, false},

		{"large_within_eps_5e-4", 1000.0, 1000.5, 5e-4, true},
		{"large_outside_eps_5e-4", 1000.0, 1001.0, 5e-4, false},

		{"well_within_boundary", 1.0, 1.0 + 5e-10, 1e-9, true},

		{"just_outside_boundary", 1.0, 1.0 + 2e-9, 1e-9, false},

		{"negative_within", -100.0, -100.0 + 100.0*1e-9*0.5, 1e-9, true},
		{"negative_outside", -100.0, -90.0, 1e-9, false},

		{"mixed_sign", 1.0, -1.0, 1e-9, false},

		{"asymmetric_scale", 1000.0, 1000.0 + 0.4, 5e-4, true},
		{"asymmetric_scale_fail", 1000.0, 1000.0 + 0.6, 5e-4, false},

		{"tiny_within", 1e-10, 1e-10 * (1 + 1e-9*0.5), 1e-9, true},
		{"tiny_outside", 1e-10, 2e-10, 1e-9, false},

		{"wide_eps_passes", 1.0, 1.5, 0.5, true},
		{"wide_eps_exact_boundary", 1.0, 2.0, 1.0, true},
		{"wide_eps_fails", 1.0, -2.0, 1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchRelative(tt.a, tt.b, tt.eps); got != tt.want {
				t.Errorf("matchRelative(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.eps, got, tt.want)
			}
			if got := matchRelative(tt.b, tt.a, tt.eps); got != tt.want {
				t.Errorf("matchRelative(%v, %v, %v) symmetry failed: got %v, want %v",
					tt.b, tt.a, tt.eps, got, tt.want)
			}
		})
	}
}

func TestMatchRelative_ScaleProportionality(t *testing.T) {
	const eps = 1e-6
	pairs := [][2]float64{
		{1.0, 1.0 + 5e-7},
		{100.0, 100.0 + 5e-5},
		{1e6, 1e6 + 0.5},
		{0.001, 0.001 + 5e-10},
	}
	for _, p := range pairs {
		if !matchRelative(p[0], p[1], eps) {
			t.Errorf("matchRelative(%v, %v, %v): same relative diff should pass at all scales",
				p[0], p[1], eps)
		}
	}
}

func TestNumericTolerancePolicy_FieldBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		base    float64
		delta   float64
		matcher func(float64, float64) bool
		want    bool
	}{
		{name: "price/exact", base: 87164.54, delta: 0, matcher: matchPrice, want: true},
		{name: "price/within", base: 87164.54, delta: 87164.54 * priceRelEps * 0.5, matcher: matchPrice, want: true},
		{name: "price/outside", base: 87164.54, delta: 87164.54 * priceRelEps * 2, matcher: matchPrice, want: false},

		{name: "financial/exact", base: 1.2028817, delta: 0, matcher: matchFinancial, want: true},
		{name: "financial/within", base: 1.2028817, delta: 1.2028817 * financialRelEps * 0.5, matcher: matchFinancial, want: true},
		{name: "financial/outside", base: 1.2028817, delta: 1.2028817 * financialRelEps * 2, matcher: matchFinancial, want: false},
		{name: "financial/negative-within", base: -10.0, delta: -10.0 * financialRelEps * 0.5, matcher: matchFinancial, want: true},
		{name: "financial/negative-outside", base: -10.0, delta: -10.0 * financialRelEps * 2, matcher: matchFinancial, want: false},
		// Sub-floor: the $1.00 denominator floor switches semantics from
		// relative-to-magnitude to relative-to-floor for values below $1.
		// Within = |delta|/floor < eps; outside = |delta|/floor > eps.
		{name: "financial/sub-floor-within", base: 0.0009307, delta: financialRelEps * 0.5, matcher: matchFinancial, want: true},
		{name: "financial/sub-floor-outside", base: 0.0009307, delta: financialRelEps * 2, matcher: matchFinancial, want: false},

		{name: "plot/exact", base: 65.4321, delta: 0, matcher: matchPlot, want: true},
		{name: "plot/within", base: 65.4321, delta: plotAbsEps * 0.5, matcher: matchPlot, want: true},
		{name: "plot/outside", base: 65.4321, delta: plotAbsEps * 2, matcher: matchPlot, want: false},
		{name: "plot/negative-within", base: -30.12, delta: plotAbsEps * 0.5, matcher: matchPlot, want: true},
		{name: "plot/negative-outside", base: -30.12, delta: plotAbsEps * 2, matcher: matchPlot, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.matcher(tt.base, tt.base+tt.delta); got != tt.want {
				t.Errorf("matcher(%v, %v) = %v, want %v", tt.base, tt.base+tt.delta, got, tt.want)
			}
		})
	}
}

func TestNumericTolerancePolicy_NaNContract(t *testing.T) {
	matchers := map[string]func(float64, float64) bool{
		"price":     matchPrice,
		"financial": matchFinancial,
		"plot":      matchPlot,
	}

	for name, matcher := range matchers {
		t.Run(name+"/both_nan", func(t *testing.T) {
			if !matcher(math.NaN(), math.NaN()) {
				t.Fatal("expected matching NaN values to pass")
			}
		})
		t.Run(name+"/nan_mismatch", func(t *testing.T) {
			if matcher(math.NaN(), 1.0) || matcher(1.0, math.NaN()) {
				t.Fatal("expected one-sided NaN to fail")
			}
		})
	}
}

func TestMatcherStrictOrdering(t *testing.T) {
	tests := []struct {
		name    string
		a, b    float64
		tighter func(float64, float64) bool
		looser  func(float64, float64) bool
	}{
		{
			name:    "price_rejects_what_financial_accepts",
			a:       100.0,
			b:       100.0 + 100.0*priceRelEps*10,
			tighter: matchPrice,
			looser:  matchFinancial,
		},
		{
			name:    "plot_rejects_absolute_delta_that_financial_accepts",
			a:       65.4321,
			b:       65.4321 + plotAbsEps*10,
			tighter: matchPlot,
			looser:  matchFinancial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tighter(tt.a, tt.b) {
				t.Errorf("tighter(%v, %v): expected reject, got accept", tt.a, tt.b)
			}
			if !tt.looser(tt.a, tt.b) {
				t.Errorf("looser(%v, %v): expected accept, got reject", tt.a, tt.b)
			}
		})
	}
}

func TestMatchPlot_InputEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want bool
	}{
		{"zero_vs_zero", 0, 0, true},
		{"zero_vs_within", 0, plotAbsEps * 0.5, true},
		{"zero_vs_outside", 0, plotAbsEps * 2, false},
		{"large_scale_within", 1e6, 1e6 + plotAbsEps*0.5, true},
		{"large_scale_outside", 1e6, 1e6 + plotAbsEps*2, false},
		{"both_pos_inf", math.Inf(1), math.Inf(1), true},
		{"both_neg_inf", math.Inf(-1), math.Inf(-1), true},
		{"pos_inf_vs_neg_inf", math.Inf(1), math.Inf(-1), false},
		{"inf_vs_large_finite", math.Inf(1), 1e15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPlot(tt.a, tt.b); got != tt.want {
				t.Errorf("matchPlot(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			if got := matchPlot(tt.b, tt.a); got != tt.want {
				t.Errorf("matchPlot(%v, %v) symmetry = %v, want %v", tt.b, tt.a, got, tt.want)
			}
		})
	}
}

func TestMatchPlot_AbsoluteScaleInvariance(t *testing.T) {
	tests := []struct {
		name  string
		scale float64
		delta float64
		want  bool
	}{
		{"tiny/within", 1e-3, plotAbsEps * 0.5, true},
		{"tiny/outside", 1e-3, plotAbsEps * 2, false},
		{"unit/within", 1.0, plotAbsEps * 0.5, true},
		{"unit/outside", 1.0, plotAbsEps * 2, false},
		{"price_like/within", 100.0, plotAbsEps * 0.5, true},
		{"price_like/outside", 100.0, plotAbsEps * 2, false},
		{"large/within", 1e6, plotAbsEps * 0.5, true},
		{"large/outside", 1e6, plotAbsEps * 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := tt.scale, tt.scale+tt.delta
			if got := matchPlot(a, b); got != tt.want {
				t.Errorf("matchPlot(%v, %v) = %v, want %v", a, b, got, tt.want)
			}
		})
	}
}

func TestNumericTolerancePolicy_ConstantInvariants(t *testing.T) {
	invariants := []struct {
		name string
		hold bool
	}{
		{
			"price tolerance is stricter than the measured sizing residual bound",
			priceRelEps < MeasuredSizeResidualBound,
		},
		{
			"measured sizing residual bound is within the financial tolerance",
			MeasuredSizeResidualBound < financialRelEps,
		},
		{
			"financial tolerance provides configured headroom above the sizing residual",
			financialRelEps/MeasuredSizeResidualBound >= financialToleranceSafetyMultiple,
		},
		{
			"financial tolerance is not vacuously permissive",
			financialRelEps < 1.0,
		},
	}

	for _, inv := range invariants {
		t.Run(inv.name, func(t *testing.T) {
			if !inv.hold {
				t.Errorf("violated: priceRelEps=%.2e MeasuredSizeResidualBound=%.2e financialRelEps=%.2e",
					priceRelEps, MeasuredSizeResidualBound, financialRelEps)
			}
		})
	}
}

func TestMatchFinancial_DenominatorFloor(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want bool
	}{
		{"sub-floor/pos/within", 0.001, 0.001 + financialRelEps*0.5, true},
		{"sub-floor/pos/outside", 0.001, 0.001 + financialRelEps*2, false},
		{"sub-floor/neg/within", -0.001, -0.001 + financialRelEps*0.5, true},
		{"sub-floor/neg/outside", -0.001, -0.001 + financialRelEps*2, false},
		{"sub-floor/zero-base/within", 0, financialRelEps * 0.5, true},
		{"sub-floor/zero-base/outside", 0, financialRelEps * 2, false},
		{"sub-floor/near-boundary/within", 0.999, 0.999 + financialRelEps*0.5, true},
		{"at-floor/within", financialDenominatorFloor, financialDenominatorFloor + financialRelEps*0.5, true},
		{"at-floor/outside", financialDenominatorFloor, financialDenominatorFloor + financialRelEps*2, false},
		{"above-floor/within", 100.0, 100.0 + 100.0*financialRelEps*0.5, true},
		{"above-floor/outside", 100.0, 100.0 + 100.0*financialRelEps*2, false},
		{"straddling/both-values-diverge", 0.5, 1.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchFinancial(tt.a, tt.b); got != tt.want {
				t.Errorf("matchFinancial(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMatchFinancial_Symmetry(t *testing.T) {
	pairs := [][2]float64{
		{0.001, 0.001 + financialRelEps*0.5},
		{0.001, 0.001 + financialRelEps*2},
		{-0.001, -0.001 + financialRelEps*0.5},
		{0, financialRelEps * 0.5},
		{financialDenominatorFloor, financialDenominatorFloor + financialRelEps*0.5},
		{100.0, 100.0 + 100.0*financialRelEps*0.5},
		{100.0, 100.0 + 100.0*financialRelEps*2},
		{0.5, 1.5},
		{0.0, 0.0},
		{math.NaN(), math.NaN()},
	}

	for _, p := range pairs {
		a, b := p[0], p[1]
		if matchFinancial(a, b) != matchFinancial(b, a) {
			t.Errorf("matchFinancial not symmetric: matchFinancial(%v,%v)=%v != matchFinancial(%v,%v)=%v",
				a, b, matchFinancial(a, b), b, a, matchFinancial(b, a))
		}
	}
}

func TestNumericTolerancePolicy_FloorInvariant(t *testing.T) {
	invariants := []struct {
		name string
		hold bool
	}{
		{
			"financialDenominatorFloor is positive so scale is always well-defined",
			financialDenominatorFloor > 0,
		},
		{
			"financialRelEps is strictly greater than MeasuredSizeResidualBound",
			financialRelEps > MeasuredSizeResidualBound,
		},
	}

	for _, inv := range invariants {
		t.Run(inv.name, func(t *testing.T) {
			if !inv.hold {
				t.Errorf("violated: financialRelEps=%.2e MeasuredSizeResidualBound=%.2e financialDenominatorFloor=%.2f",
					financialRelEps, MeasuredSizeResidualBound, financialDenominatorFloor)
			}
		})
	}
}
