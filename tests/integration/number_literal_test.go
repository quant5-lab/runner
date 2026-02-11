package integration

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Number literal end-to-end value correctness - validates Parse→Codegen→Compile→Execute preserves exact numeric values */

func TestNumberLiteralFormats(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Number Literal Formats", overlay=false)
int_pos = 42
int_zero = 0
float_std = 3.14
float_zero = 0.0
leading_half = .5
leading_small = .123
trailing_one = 1.
trailing_hundred = 100.
sci_large = 6.02e23
sci_small = 1.6e-19
sci_int = 3e8
sci_upper = 1E10
sci_neg_exp = 2.5E-5
sci_zero_exp = 5e0
sci_pos_sign = 1e+6
plot(int_pos, "int_pos")
plot(int_zero, "int_zero")
plot(float_std, "float_std")
plot(float_zero, "float_zero")
plot(leading_half, "leading_half")
plot(leading_small, "leading_small")
plot(trailing_one, "trailing_one")
plot(trailing_hundred, "trailing_hundred")
plot(sci_large, "sci_large")
plot(sci_small, "sci_small")
plot(sci_int, "sci_int")
plot(sci_upper, "sci_upper")
plot(sci_neg_exp, "sci_neg_exp")
plot(sci_zero_exp, "sci_zero_exp")
plot(sci_pos_sign, "sci_pos_sign")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "number-literal-formats", pineScript)

	tests := []struct {
		name   string
		expect float64
	}{
		{"int_pos", 42},
		{"int_zero", 0},
		{"float_std", 3.14},
		{"float_zero", 0.0},
		{"leading_half", 0.5},
		{"leading_small", 0.123},
		{"trailing_one", 1.0},
		{"trailing_hundred", 100.0},
		{"sci_large", 6.02e23},
		{"sci_small", 1.6e-19},
		{"sci_int", 3e8},
		{"sci_upper", 1e10},
		{"sci_neg_exp", 2.5e-5},
		{"sci_zero_exp", 5.0},
		{"sci_pos_sign", 1e6},
	}

	for _, tt := range tests {
		values := exec.ExtractPlotValues(t, output, tt.name)
		if len(values) == 0 {
			t.Errorf("%s: no values", tt.name)
			continue
		}

		got := values[0]
		if !floatEqual(got, tt.expect) {
			t.Errorf("%s: got %e, want %e", tt.name, got, tt.expect)
		}
	}

	t.Log("✅ All number literal formats validated")
}

func TestNumberLiteralExpressions(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Number Literal Expressions", overlay=false)
mixed_add = 1e3 + .5 + 10.
mixed_mult = 2e2 * .5 * 3.
sci_div = 6.02e23 / 1e20
boundary_sub = 100. - .5
complex = (1e2 + 50.) * .5
plot(mixed_add, "mixed_add")
plot(mixed_mult, "mixed_mult")
plot(sci_div, "sci_div")
plot(boundary_sub, "boundary_sub")
plot(complex, "complex")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "number-literal-expressions", pineScript)

	tests := []struct {
		name   string
		expect float64
	}{
		{"mixed_add", 1e3 + 0.5 + 10.0},
		{"mixed_mult", 2e2 * 0.5 * 3.0},
		{"sci_div", 6.02e23 / 1e20},
		{"boundary_sub", 100.0 - 0.5},
		{"complex", (1e2 + 50.0) * 0.5},
	}

	for _, tt := range tests {
		values := exec.ExtractPlotValues(t, output, tt.name)
		if len(values) == 0 {
			t.Errorf("%s: no values", tt.name)
			continue
		}

		got := values[0]
		if !floatEqual(got, tt.expect) {
			t.Errorf("%s: got %f, want %f", tt.name, got, tt.expect)
		}
	}

	t.Log("✅ Number literal expressions validated")
}

func TestNumberLiteralBoundaries(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Number Literal Boundaries", overlay=false)
very_large = 1e100
very_small = 1e-100
zero_variants_pos = 0.0
zero_variants_sci = 0e0
near_zero = 1e-15
plot(very_large, "very_large")
plot(very_small, "very_small")
plot(zero_variants_pos, "zero_variants_pos")
plot(zero_variants_sci, "zero_variants_sci")
plot(near_zero, "near_zero")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "number-literal-boundaries", pineScript)

	tests := []struct {
		name   string
		expect float64
	}{
		{"very_large", 1e100},
		{"very_small", 1e-100},
		{"zero_variants_pos", 0.0},
		{"zero_variants_sci", 0.0},
		{"near_zero", 1e-15},
	}

	for _, tt := range tests {
		values := exec.ExtractPlotValues(t, output, tt.name)
		if len(values) == 0 {
			t.Errorf("%s: no values", tt.name)
			continue
		}

		got := values[0]
		if !floatEqual(got, tt.expect) {
			t.Errorf("%s: got %e, want %e", tt.name, got, tt.expect)
		}
	}

	t.Log("✅ Number literal boundary values validated")
}

func floatEqual(a, b float64) bool {
	if b == 0 {
		return math.Abs(a) < 1e-15
	}
	return math.Abs((a-b)/b) < 1e-9
}
