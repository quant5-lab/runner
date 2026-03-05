package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func makeTACall1(funcName, source string) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: []ast.Expression{&ast.Identifier{Name: source}},
	}
}

func makeTACall2Expr(funcName string, a1, a2 ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: []ast.Expression{a1, a2},
	}
}

func makeMathCall(funcName string, args ...ast.Expression) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "math"},
			Property: &ast.Identifier{Name: funcName},
		},
		Arguments: args,
	}
}

func makeCtxClose(closes ...float64) *context.Context {
	data := make([]context.OHLCV, len(closes))
	for i, c := range closes {
		data[i] = context.OHLCV{Close: c, Open: c, High: c, Low: c, Volume: 1000}
	}
	return &context.Context{Data: data}
}

func assertFloat64(t *testing.T, label string, got, want, tol float64) {
	t.Helper()
	if math.IsNaN(want) {
		if !math.IsNaN(got) {
			t.Errorf("%s: expected NaN, got %f", label, got)
		}
		return
	}
	if math.IsNaN(got) {
		t.Errorf("%s: expected %f, got NaN", label, want)
		return
	}
	if math.Abs(got-want) > tol {
		t.Errorf("%s: expected %f, got %f", label, want, got)
	}
}

func lit(v float64) *ast.Literal { return &ast.Literal{Value: v} }

func TestStreamingBarEvaluator_Change(t *testing.T) {
	ctx := makeCtxClose(10, 12, 15, 11, 14)

	t.Run("default_length", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeTACall1("change", "close")
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"warmup", 0, math.NaN()},
			{"positive_change", 1, 2},
			{"larger_positive", 2, 3},
			{"negative_change", 3, -4},
			{"recovery", 4, 3},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("explicit_length_extends_warmup", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("change", "close", 2)
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"warmup_bar0", 0, math.NaN()},
			{"warmup_bar1", 1, math.NaN()},
			{"bar2", 2, 5},
			{"bar3", 3, -1},
			{"bar4", 4, -1},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})
}

func TestStreamingBarEvaluator_Mom(t *testing.T) {
	ctx := makeCtxClose(10, 12, 15, 11, 14)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("mom", "close", 2)
	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, 5},
		{"bar3", 3, -1},
		{"bar4", 4, -1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "value", got, c.want, 1e-9)
		})
	}
}

func TestStreamingBarEvaluator_Roc(t *testing.T) {
	t.Run("warmup_and_correctness", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 15, 25, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("roc", "close", 2)
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"warmup_bar0", 0, math.NaN()},
			{"warmup_bar1", 1, math.NaN()},
			{"bar2", 2, 50.0},
			{"bar3", 3, 25.0},
			{"bar4", 4, 100.0 * 5 / 15},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 0.01)
			})
		}
	})

	t.Run("zero_previous_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(0, 10)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("roc", "close", 1)
		got, err := ev.EvaluateAtBar(call, ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN when previous=0, got %f", got)
		}
	})
}

func TestStreamingBarEvaluator_Crossover(t *testing.T) {
	closeID := &ast.Identifier{Name: "close"}
	openID := &ast.Identifier{Name: "open"}

	cases := []struct {
		name string
		data []context.OHLCV
		bar  int
		want float64
	}{
		{
			"bar0_no_history",
			[]context.OHLCV{{Close: 98, Open: 100}},
			0, 0.0,
		},
		{
			"already_above_no_cross",
			[]context.OHLCV{{Close: 103, Open: 100}, {Close: 105, Open: 100}},
			1, 0.0,
		},
		{
			"true_crossover",
			[]context.OHLCV{{Close: 95, Open: 100}, {Close: 103, Open: 100}},
			1, 1.0,
		},
		{
			"touch_then_cross_counts",
			[]context.OHLCV{{Close: 100, Open: 100}, {Close: 103, Open: 100}},
			1, 1.0,
		},
		{
			"reaches_equal_not_crossover",
			[]context.OHLCV{{Close: 95, Open: 100}, {Close: 100, Open: 100}},
			1, 0.0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := &context.Context{Data: c.data}
			ev := NewStreamingBarEvaluator()
			call := makeTACall2Expr("crossover", closeID, openID)
			got, err := ev.EvaluateAtBar(call, ctx, c.bar)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestStreamingBarEvaluator_Crossunder(t *testing.T) {
	closeID := &ast.Identifier{Name: "close"}
	openID := &ast.Identifier{Name: "open"}

	cases := []struct {
		name string
		data []context.OHLCV
		bar  int
		want float64
	}{
		{
			"bar0_no_history",
			[]context.OHLCV{{Close: 103, Open: 100}},
			0, 0.0,
		},
		{
			"already_below_no_cross",
			[]context.OHLCV{{Close: 97, Open: 100}, {Close: 95, Open: 100}},
			1, 0.0,
		},
		{
			"true_crossunder",
			[]context.OHLCV{{Close: 105, Open: 100}, {Close: 97, Open: 100}},
			1, 1.0,
		},
		{
			"touch_then_cross_counts",
			[]context.OHLCV{{Close: 100, Open: 100}, {Close: 97, Open: 100}},
			1, 1.0,
		},
		{
			"reaches_equal_not_crossunder",
			[]context.OHLCV{{Close: 105, Open: 100}, {Close: 100, Open: 100}},
			1, 0.0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := &context.Context{Data: c.data}
			ev := NewStreamingBarEvaluator()
			call := makeTACall2Expr("crossunder", closeID, openID)
			got, err := ev.EvaluateAtBar(call, ctx, c.bar)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestStreamingBarEvaluator_Cross(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 98, Open: 100},
			{Close: 103, Open: 100},
			{Close: 97, Open: 100},
			{Close: 95, Open: 100},
		},
	}
	closeID := &ast.Identifier{Name: "close"}
	openID := &ast.Identifier{Name: "open"}
	ev := NewStreamingBarEvaluator()
	call := makeTACall2Expr("cross", closeID, openID)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"bar0_no_history", 0, 0.0},
		{"crossover_detected", 1, 1.0},
		{"crossunder_detected", 2, 1.0},
		{"no_cross", 3, 0.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestStreamingBarEvaluator_Falling(t *testing.T) {
	cases := []struct {
		name   string
		data   []float64
		length float64
		barIdx int
		want   float64
	}{
		{"warmup_returns_false", []float64{50, 40, 30}, 3, 0, 0.0},
		{"monotone_falling", []float64{50, 40, 30, 20, 10}, 3, 3, 1.0},
		{"flat_breaks_sequence", []float64{30, 20, 20, 10}, 2, 3, 0.0},
		{"rising_breaks_sequence", []float64{10, 20, 30, 20, 10}, 3, 4, 0.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := makeCtxClose(c.data...)
			ev := NewStreamingBarEvaluator()
			call := createTACallExpression("falling", "close", c.length)
			got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestStreamingBarEvaluator_Rising(t *testing.T) {
	cases := []struct {
		name   string
		data   []float64
		length float64
		barIdx int
		want   float64
	}{
		{"warmup_returns_false", []float64{10, 20, 30}, 3, 0, 0.0},
		{"monotone_rising", []float64{10, 20, 30, 40, 50}, 3, 3, 1.0},
		{"flat_breaks_sequence", []float64{10, 20, 20, 30}, 2, 3, 0.0},
		{"falling_breaks_sequence", []float64{50, 40, 30, 40, 50}, 3, 4, 0.0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := makeCtxClose(c.data...)
			ev := NewStreamingBarEvaluator()
			call := createTACallExpression("rising", "close", c.length)
			got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("expected %v, got %v", c.want, got)
			}
		})
	}
}

func TestStreamingBarEvaluator_BarsSince(t *testing.T) {
	cond := &ast.BinaryExpression{
		Left:     &ast.Identifier{Name: "close"},
		Operator: ">",
		Right:    lit(100),
	}
	makeCall := func(funcName string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: funcName},
			},
			Arguments: []ast.Expression{cond},
		}
	}

	t.Run("sequence", func(t *testing.T) {
		ctx := makeCtxClose(98, 103, 100, 102, 99)
		ev := NewStreamingBarEvaluator()
		call := makeCall("barssince")
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"never_true_yet", 0, math.NaN()},
			{"found_at_current", 1, 0.0},
			{"found_one_bar_ago", 2, 1.0},
			{"found_again_at_current", 3, 0.0},
			{"found_one_bar_ago_again", 4, 1.0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("never_true_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(95, 97, 93)
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeCall("barssince"), ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN when condition never true, got %f", got)
		}
	})

	t.Run("barsince_alias", func(t *testing.T) {
		ctx := makeCtxClose(95, 105)
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeCall("barsince"), ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})

	t.Run("first_bar_true_returns_zero", func(t *testing.T) {
		ctx := makeCtxClose(105)
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeCall("barssince"), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})

	t.Run("condition_always_true_counter_always_zero", func(t *testing.T) {
		ctx := makeCtxClose(105, 110, 115)
		ev := NewStreamingBarEvaluator()
		call := makeCall("barssince")
		for barIdx := 0; barIdx <= 2; barIdx++ {
			got, err := ev.EvaluateAtBar(call, ctx, barIdx)
			if err != nil {
				t.Fatalf("bar %d: %v", barIdx, err)
			}
			assertFloat64(t, "value", got, 0.0, 1e-9)
		}
	})

	t.Run("counter_grows_until_reset", func(t *testing.T) {
		ctx := makeCtxClose(105, 99, 99, 99, 99, 110)
		ev := NewStreamingBarEvaluator()
		call := makeCall("barssince")
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"true_resets_to_zero", 0, 0.0},
			{"one_bar_ago", 1, 1.0},
			{"two_bars_ago", 2, 2.0},
			{"three_bars_ago", 3, 3.0},
			{"four_bars_ago", 4, 4.0},
			{"true_resets_again", 5, 0.0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("nan_condition_treated_as_false", func(t *testing.T) {
		ctx := makeCtxClose(10, 10, 20)
		ev := NewStreamingBarEvaluator()
		changeCall := createTACallExpression("change", "close", 1)
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "barssince"},
			},
			Arguments: []ast.Expression{changeCall},
		}
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"nan_cond_no_prior", 0, math.NaN()},
			{"false_cond_no_valid_prior", 1, math.NaN()},
			{"true_cond_resets", 2, 0.0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("idempotent_repeated_bar_access", func(t *testing.T) {
		ctx := makeCtxClose(95, 105, 99)
		ev := NewStreamingBarEvaluator()
		call := makeCall("barssince")
		got1, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		got2, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "first", got1, 1.0, 1e-9)
		assertFloat64(t, "second", got2, 1.0, 1e-9)
	})

	t.Run("stateful_condition_sub_expression", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 15, 25)
		ev := NewStreamingBarEvaluator()
		changeCall := createTACallExpression("change", "close", 1)
		condExpr := &ast.BinaryExpression{
			Left:     changeCall,
			Operator: ">",
			Right:    lit(0),
		}
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "barssince"},
			},
			Arguments: []ast.Expression{condExpr},
		}
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"warmup_propagates", 0, math.NaN()},
			{"change_positive_resets", 1, 0.0},
			{"change_negative_increments", 2, 1.0},
			{"change_positive_resets_again", 3, 0.0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})
}

func TestStreamingBarEvaluator_Cum(t *testing.T) {
	ctx := makeCtxClose(10, 20, 30)
	ev := NewStreamingBarEvaluator()
	call := makeTACall1("cum", "close")

	t.Run("accumulation", func(t *testing.T) {
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"bar0", 0, 10},
			{"bar1", 1, 30},
			{"bar2", 2, 60},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := ev.EvaluateAtBar(call, ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("idempotent_repeated_bar", func(t *testing.T) {
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 60.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_MathMax(t *testing.T) {
	ctx := makeCtxClose(10, 5, 15)

	t.Run("two_args", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("max", lit(5), lit(3)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 5, 1e-9)
	})

	t.Run("three_args", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("max", lit(1), lit(9), lit(4)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 9, 1e-9)
	})

	t.Run("single_arg", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("max", lit(7)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 7, 1e-9)
	})

	t.Run("nan_propagates", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, _ := ev.EvaluateAtBar(makeMathCall("max", lit(5), &ast.Literal{Value: math.NaN()}), ctx, 0)
		if !math.IsNaN(got) {
			t.Errorf("expected NaN, got %f", got)
		}
	})

	t.Run("composed_expression", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		call := makeMathCall("max", makeTACall1("change", "close"), lit(0))
		cases := []struct {
			name   string
			barIdx int
			want   float64
		}{
			{"change_warmup_propagates", 0, math.NaN()},
			{"clamps_decrease", 1, 0.0},
			{"passes_increase", 2, 10.0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, _ := ev.EvaluateAtBar(call, ctx, c.barIdx)
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})
}

func TestStreamingBarEvaluator_MathMin(t *testing.T) {
	ctx := makeCtxClose(10)

	t.Run("two_args", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("min", lit(7), lit(3)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 3, 1e-9)
	})

	t.Run("three_args", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("min", lit(7), lit(3), lit(9)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 3, 1e-9)
	})

	t.Run("single_arg", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeMathCall("min", lit(5)), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 5, 1e-9)
	})

	t.Run("nan_propagates", func(t *testing.T) {
		ev := NewStreamingBarEvaluator()
		got, _ := ev.EvaluateAtBar(makeMathCall("min", lit(5), &ast.Literal{Value: math.NaN()}), ctx, 0)
		if !math.IsNaN(got) {
			t.Errorf("expected NaN, got %f", got)
		}
	})
}

func TestStreamingBarEvaluator_MathAbs(t *testing.T) {
	ctx := makeCtxClose(10)
	ev := NewStreamingBarEvaluator()

	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"negative", -7.5, 7.5},
		{"positive", 7.5, 7.5},
		{"zero", 0.0, 0.0},
		{"nan_passthrough", math.NaN(), math.NaN()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			call := makeMathCall("abs", &ast.Literal{Value: c.input})
			got, err := ev.EvaluateAtBar(call, ctx, 0)
			if err != nil {
				t.Fatal(err)
			}
			assertFloat64(t, "value", got, c.want, 1e-9)
		})
	}
}

func TestStreamingBarEvaluator_FormulaInsufficientArguments(t *testing.T) {
	ctx := makeCtxClose(10)

	zeroArgTA := func(funcName string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: funcName},
			},
		}
	}
	zeroArgMath := func(funcName string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "math"},
				Property: &ast.Identifier{Name: funcName},
			},
		}
	}

	cases := []struct {
		name string
		call *ast.CallExpression
	}{
		{"change", zeroArgTA("change")},
		{"crossover", zeroArgTA("crossover")},
		{"crossunder", zeroArgTA("crossunder")},
		{"cross", zeroArgTA("cross")},
		{"barssince", zeroArgTA("barssince")},
		{"cum", zeroArgTA("cum")},
		{"math.max", zeroArgMath("max")},
		{"math.min", zeroArgMath("min")},
		{"math.abs", zeroArgMath("abs")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := NewStreamingBarEvaluator()
			_, err := ev.EvaluateAtBar(c.call, ctx, 0)
			if err == nil {
				t.Errorf("expected error for zero arguments to %s", c.name)
			}
		})
	}
}
