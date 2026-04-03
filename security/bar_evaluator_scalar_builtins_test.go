package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// makeSimpleCallExpr builds a call expression with a bare identifier callee (not namespace-qualified).
func makeSimpleCallExpr(funcName string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee:    &ast.Identifier{Name: funcName},
		Arguments: args,
	}
}

// fptr is a convenience helper for providing an optional float64 replacement argument.
func fptr(v float64) *float64 { return &v }

func TestNz(t *testing.T) {
	// nz(value [, replacement=0]) replaces NaN with replacement (default 0).
	// Non-NaN values pass through unchanged regardless of any replacement argument.
	// ta.nz is a registered alias and must behave identically to nz.
	t.Run("non_nan_passes_through", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"})
		got, err := ev.EvaluateAtBar(call, makeCtxClose(5.0), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "nz(5)", got, 5.0, 1e-12)
	})

	t.Run("nan_replaced_by_default_zero", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"})
		got, err := ev.EvaluateAtBar(call, makeCtxClose(math.NaN()), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "nz(NaN)", got, 0.0, 1e-12)
	})

	t.Run("nan_replaced_by_explicit_value", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"}, lit(99.0))
		got, err := ev.EvaluateAtBar(call, makeCtxClose(math.NaN()), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "nz(NaN, 99)", got, 99.0, 1e-12)
	})

	t.Run("non_nan_ignores_replacement_arg", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"}, lit(99.0))
		got, err := ev.EvaluateAtBar(call, makeCtxClose(3.0), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "nz(3, 99)", got, 3.0, 1e-12)
	})

	t.Run("zero_is_not_nan", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"}, lit(99.0))
		got, err := ev.EvaluateAtBar(call, makeCtxClose(0.0), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "nz(0, 99)", got, 0.0, 1e-12)
	})

	t.Run("ta_nz_alias_replaces_nan", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeTACall1("nz", "close")
		got, err := ev.EvaluateAtBar(call, makeCtxClose(math.NaN()), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "ta.nz(NaN)", got, 0.0, 1e-12)
	})
}

func TestNa(t *testing.T) {
	// The identifier `na` evaluates to NaN (a typed-NaN literal in PineScript).
	// The function form na(x) returns 1.0 when x is NaN, 0.0 otherwise.
	// ta.na is a registered alias for the function form.
	t.Run("identifier_is_nan", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(&ast.Identifier{Name: "na"}, makeCtxClose(1.0), 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("identifier 'na' should evaluate to NaN, got %v", got)
		}
	})

	t.Run("function_form_non_nan_is_false", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("na", &ast.Identifier{Name: "close"})
		got, err := ev.EvaluateAtBar(call, makeCtxClose(5.0), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "na(5)", got, 0.0, 1e-12)
	})

	t.Run("function_form_nan_is_true", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeSimpleCallExpr("na", &ast.Identifier{Name: "close"})
		got, err := ev.EvaluateAtBar(call, makeCtxClose(math.NaN()), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "na(NaN)", got, 1.0, 1e-12)
	})

	t.Run("ta_na_alias_is_function_form", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeTACall1("na", "close")
		got, err := ev.EvaluateAtBar(call, makeCtxClose(math.NaN()), 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "ta.na(NaN)", got, 1.0, 1e-12)
	})
}

func TestInt(t *testing.T) {
	// int(value) truncates toward zero — equivalent to math.Trunc in Go.
	// This differs from floor for negative fractions: int(-3.9) = -3, floor(-3.9) = -4.
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"positive_fraction_truncates", 3.9, 3.0},
		{"negative_fraction_truncates_toward_zero", -3.9, -3.0},
		{"zero_unchanged", 0.0, 0.0},
		{"exact_integer_unchanged", 7.0, 7.0},
		{"negative_exact_integer", -4.0, -4.0},
		{"nan_propagates", math.NaN(), math.NaN()},
	}
	ev := NewStreamingBarEvaluator()
	call := makeSimpleCallExpr("int", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, c.name, got, c.want, 1e-12)
		})
	}
}

func TestFloat(t *testing.T) {
	// float(value) is an identity cast — the security evaluator already works in
	// float64, so the value is returned unchanged including NaN.
	cases := []struct {
		name  string
		input float64
	}{
		{"normal_value", 42.5},
		{"zero", 0.0},
		{"negative", -7.3},
		{"nan_propagates", math.NaN()},
	}
	ev := NewStreamingBarEvaluator()
	call := makeSimpleCallExpr("float", &ast.Identifier{Name: "close"})
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, c.name, got, c.input, 1e-15)
		})
	}
}

func TestScalarBuiltinsInsufficientArguments(t *testing.T) {
	// All scalar builtins must return an error when called with no arguments.
	ctx := makeCtxClose(1.0)
	cases := []struct {
		name string
		call *ast.CallExpression
	}{
		{"nz", makeSimpleCallExpr("nz")},
		{"na", makeSimpleCallExpr("na")},
		{"int", makeSimpleCallExpr("int")},
		{"float", makeSimpleCallExpr("float")},
	}
	for _, c := range cases {
		t.Run(c.name+"_no_args", func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			_, err := ev.EvaluateAtBar(c.call, ctx, 0)
			if err == nil {
				t.Errorf("%s with no args: expected error, got nil", c.name)
			}
		})
	}
}

func TestNzComposedInBinaryExpression(t *testing.T) {
	// nz should correctly participate as a sub-expression inside a larger expression
	// tree — verifying that NaN replacement and normal passthrough both compose
	// correctly with surrounding arithmetic.
	ev := NewStreamingBarEvaluator()
	nzCall := makeSimpleCallExpr("nz", &ast.Identifier{Name: "close"}, lit(5.0))
	addExpr := &ast.BinaryExpression{
		Operator: "+",
		Left:     nzCall,
		Right:    lit(1.0),
	}

	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"nan_replaced_then_added", math.NaN(), 6.0},   // nz(NaN, 5) + 1 = 6
		{"normal_value_passed_then_added", 10.0, 11.0}, // nz(10, 5) + 1 = 11
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(addExpr, makeCtxClose(c.input), 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, c.name, got, c.want, 1e-12)
		})
	}
}
