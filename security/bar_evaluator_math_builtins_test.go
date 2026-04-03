package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// makeMathMemberExpr builds a math.<property> member expression for testing math.* constants.
func makeMathMemberExpr(property string) *ast.MemberExpression {
	return &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "math"},
		Property: &ast.Identifier{Name: property},
	}
}

func TestMathNamespaceConstants(t *testing.T) {
	ev := NewStreamingBarEvaluator()
	ctx := makeCtxClose(1.0)

	cases := []struct {
		property string
		want     float64
	}{
		{"pi", math.Pi},
		{"e", math.E},
		{"phi", 1.6180339887498948482},
		{"rphi", 0.6180339887498948482},
		{"huge", math.MaxFloat64},
		{"tiny", math.SmallestNonzeroFloat64},
	}
	for _, c := range cases {
		t.Run(c.property, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(makeMathMemberExpr(c.property), ctx, 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math."+c.property, got, c.want, 1e-15)
		})
	}
}

func TestMathLogarithms(t *testing.T) {
	// Both math.log (natural) and math.log10 (base-10) share the same domain rule:
	// inputs <= 0 produce NaN; NaN input also produces NaN.
	cases := []struct {
		name     string
		funcName string
		input    float64
		want     float64
	}{
		{"log_of_1_is_zero", "log", 1.0, 0.0},
		{"log_of_e_is_one", "log", math.E, 1.0},
		{"log_positive", "log", 100.0, math.Log(100.0)},
		{"log_zero_is_nan", "log", 0.0, math.NaN()},
		{"log_negative_is_nan", "log", -1.0, math.NaN()},
		{"log_nan_propagates", "log", math.NaN(), math.NaN()},
		{"log10_of_1_is_zero", "log10", 1.0, 0.0},
		{"log10_of_10_is_one", "log10", 10.0, 1.0},
		{"log10_of_100_is_two", "log10", 100.0, 2.0},
		{"log10_zero_is_nan", "log10", 0.0, math.NaN()},
		{"log10_negative_is_nan", "log10", -5.0, math.NaN()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := makeMathCall(c.funcName, &ast.Identifier{Name: "close"})
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math."+c.funcName, got, c.want, 1e-12)
		})
	}
}

func TestMathExp(t *testing.T) {
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"exp_of_zero_is_one", 0.0, 1.0},
		{"exp_of_one_is_e", 1.0, math.E},
		{"exp_of_neg_one", -1.0, 1.0 / math.E},
		{"exp_nan_propagates", math.NaN(), math.NaN()},
	}
	ev := NewStreamingBarEvaluator()
	call := makeMathCall("exp", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math.exp", got, c.want, 1e-12)
		})
	}
}

func TestMathSqrt(t *testing.T) {
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"sqrt_of_zero", 0.0, 0.0},
		{"sqrt_of_four", 4.0, 2.0},
		{"sqrt_of_nine", 9.0, 3.0},
		{"sqrt_negative_is_nan", -1.0, math.NaN()},
		{"sqrt_nan_propagates", math.NaN(), math.NaN()},
	}
	ev := NewStreamingBarEvaluator()
	call := makeMathCall("sqrt", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math.sqrt", got, c.want, 1e-12)
		})
	}
}

func TestMathPow(t *testing.T) {
	cases := []struct {
		name string
		base float64
		exp  float64
		want float64
	}{
		{"integer_exponent", 2.0, 10.0, 1024.0},
		{"cubing", 3.0, 3.0, 27.0},
		{"fractional_exponent_as_root", 4.0, 0.5, 2.0},
		{"any_base_to_zero_is_one", 7.0, 0.0, 1.0},
		{"zero_base_to_positive", 0.0, 3.0, 0.0},
		{"negative_integer_exponent", 2.0, -1.0, 0.5},
		{"negative_base_fractional_exp_is_nan", -1.0, 0.5, math.NaN()},
	}
	ctx := makeCtxClose(0) // base/exp supplied as literals; ctx not used for inputs
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := makeMathCall("pow", lit(c.base), lit(c.exp))
			got, err := ev.EvaluateAtBar(call, ctx, 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math.pow", got, c.want, 1e-10)
		})
	}
}

func TestMathRounding(t *testing.T) {
	// floor rounds toward -∞; ceil rounds toward +∞; round rounds half away from zero.
	cases := []struct {
		name     string
		funcName string
		input    float64
		want     float64
	}{
		{"floor_positive", "floor", 1.7, 1.0},
		{"floor_negative", "floor", -1.7, -2.0},
		{"floor_exact_integer", "floor", 3.0, 3.0},
		{"floor_nan_propagates", "floor", math.NaN(), math.NaN()},
		{"ceil_positive", "ceil", 1.7, 2.0},
		{"ceil_negative", "ceil", -1.7, -1.0},
		{"ceil_exact_integer", "ceil", 3.0, 3.0},
		{"ceil_nan_propagates", "ceil", math.NaN(), math.NaN()},
		{"round_half_up", "round", 1.5, 2.0},
		{"round_below_half", "round", 1.4, 1.0},
		{"round_negative_half", "round", -1.5, -2.0},
		{"round_negative_below_half", "round", -1.4, -1.0},
		{"round_nan_propagates", "round", math.NaN(), math.NaN()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := makeMathCall(c.funcName, &ast.Identifier{Name: "close"})
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math."+c.funcName, got, c.want, 1e-12)
		})
	}
}

func TestMathSign(t *testing.T) {
	// NaN comparisons are always false in Go, so sign(NaN) falls to the default
	// case and returns 0.0 — consistent with PineScript behavior.
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"positive_returns_one", 5.0, 1.0},
		{"negative_returns_neg_one", -3.0, -1.0},
		{"zero_returns_zero", 0.0, 0.0},
		{"nan_returns_zero", math.NaN(), 0.0},
	}
	ev := NewStreamingBarEvaluator()
	call := makeMathCall("sign", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math.sign", got, c.want, 1e-12)
		})
	}
}

func TestMathTrigonometry(t *testing.T) {
	// Covers all six trig functions. asin and acos are only defined on [-1,1];
	// inputs outside that range produce NaN. All functions propagate NaN input.
	cases := []struct {
		name     string
		funcName string
		input    float64
		want     float64
	}{
		// sin
		{"sin_zero", "sin", 0.0, 0.0},
		{"sin_half_pi", "sin", math.Pi / 2, 1.0},
		{"sin_neg_half_pi", "sin", -math.Pi / 2, -1.0},
		{"sin_nan_propagates", "sin", math.NaN(), math.NaN()},
		// cos
		{"cos_zero", "cos", 0.0, 1.0},
		{"cos_pi", "cos", math.Pi, -1.0},
		{"cos_nan_propagates", "cos", math.NaN(), math.NaN()},
		// tan
		{"tan_zero", "tan", 0.0, 0.0},
		{"tan_quarter_pi", "tan", math.Pi / 4, 1.0},
		{"tan_nan_propagates", "tan", math.NaN(), math.NaN()},
		// asin — domain [-1, 1]
		{"asin_zero", "asin", 0.0, 0.0},
		{"asin_one_is_half_pi", "asin", 1.0, math.Pi / 2},
		{"asin_neg_one_is_neg_half_pi", "asin", -1.0, -math.Pi / 2},
		{"asin_out_of_domain_is_nan", "asin", 2.0, math.NaN()},
		{"asin_nan_propagates", "asin", math.NaN(), math.NaN()},
		// acos — domain [-1, 1]
		{"acos_one_is_zero", "acos", 1.0, 0.0},
		{"acos_zero_is_half_pi", "acos", 0.0, math.Pi / 2},
		{"acos_neg_one_is_pi", "acos", -1.0, math.Pi},
		{"acos_out_of_domain_is_nan", "acos", 2.0, math.NaN()},
		{"acos_nan_propagates", "acos", math.NaN(), math.NaN()},
		// atan — domain is all real numbers
		{"atan_zero", "atan", 0.0, 0.0},
		{"atan_one_is_quarter_pi", "atan", 1.0, math.Pi / 4},
		{"atan_neg_one", "atan", -1.0, -math.Pi / 4},
		{"atan_nan_propagates", "atan", math.NaN(), math.NaN()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := makeMathCall(c.funcName, &ast.Identifier{Name: "close"})
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math."+c.funcName, got, c.want, 1e-12)
		})
	}
}

func TestMathAngleConversion(t *testing.T) {
	cases := []struct {
		name     string
		funcName string
		input    float64
		want     float64
	}{
		{"toradians_180_is_pi", "toradians", 180.0, math.Pi},
		{"toradians_90_is_half_pi", "toradians", 90.0, math.Pi / 2},
		{"toradians_zero", "toradians", 0.0, 0.0},
		{"toradians_nan_propagates", "toradians", math.NaN(), math.NaN()},
		{"todegrees_pi_is_180", "todegrees", math.Pi, 180.0},
		{"todegrees_half_pi_is_90", "todegrees", math.Pi / 2, 90.0},
		{"todegrees_zero", "todegrees", 0.0, 0.0},
		{"todegrees_nan_propagates", "todegrees", math.NaN(), math.NaN()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := makeMathCall(c.funcName, &ast.Identifier{Name: "close"})
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math."+c.funcName, got, c.want, 1e-12)
		})
	}
}

func TestMathRoundToMintick(t *testing.T) {
	// round_to_mintick is a passthrough in the security evaluator: tick size is
	// unavailable at security-evaluation time, so the value is returned unchanged.
	cases := []struct {
		name  string
		input float64
	}{
		{"positive_fraction", 1.23456},
		{"negative_value", -5.5},
		{"zero", 0.0},
		{"nan_propagates", math.NaN()},
	}
	ev := NewStreamingBarEvaluator()
	call := makeMathCall("round_to_mintick", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "math.round_to_mintick", got, c.input, 1e-15)
		})
	}
}

func TestMathInsufficientArguments(t *testing.T) {
	// Every unary math function must error when called with zero arguments.
	// math.pow requires two arguments and must error with only one.
	ctx := makeCtxClose(1.0)

	unaryFuncs := []string{
		"log", "log10", "exp", "sqrt", "round", "round_to_mintick",
		"floor", "ceil", "sign",
		"sin", "cos", "tan", "asin", "acos", "atan",
		"toradians", "todegrees",
	}
	for _, fn := range unaryFuncs {
		t.Run(fn+"_no_args", func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "math"},
					Property: &ast.Identifier{Name: fn},
				},
			}
			_, err := ev.EvaluateAtBar(call, ctx, 0)
			if err == nil {
				t.Errorf("math.%s with no args: expected error, got nil", fn)
			}
		})
	}

	t.Run("pow_one_arg", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		_, err := ev.EvaluateAtBar(makeMathCall("pow", lit(2.0)), ctx, 0)
		if err == nil {
			t.Error("math.pow with one arg: expected error, got nil")
		}
	})
}
