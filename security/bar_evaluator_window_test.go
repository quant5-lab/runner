package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func makeWPRCall(period float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "wpr"},
		},
		Arguments: []ast.Expression{lit(period)},
	}
}

func makeLinregCall(source string, length, offset float64) *ast.CallExpression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "linreg"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: source},
			lit(length),
			lit(offset),
		},
	}
}

func makeCtxFull(bars []context.OHLCV) *context.Context {
	return &context.Context{Data: bars}
}

func TestStreamingBarEvaluator_Highest(t *testing.T) {
	ctx := makeCtxClose(10, 20, 30, 25, 15)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("highest", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, 30},
		{"bar3", 3, 30},
		{"bar4", 4, 30},
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

func TestStreamingBarEvaluator_Lowest(t *testing.T) {
	ctx := makeCtxClose(10, 20, 30, 25, 15)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("lowest", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, 10},
		{"bar3", 3, 20},
		{"bar4", 4, 15},
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

func TestStreamingBarEvaluator_Sum(t *testing.T) {
	ctx := makeCtxClose(10, 20, 30, 40, 50)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("sum", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, 60},
		{"bar3", 3, 90},
		{"bar4", 4, 120},
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

func TestStreamingBarEvaluator_Range(t *testing.T) {
	ctx := makeCtxClose(10, 30, 20, 50, 40)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("range", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, 20},
		{"bar3", 3, 30},
		{"bar4", 4, 30},
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

func TestStreamingBarEvaluator_Dev(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("dev", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("flat_source_zero_dev", func(t *testing.T) {
		ctx := makeCtxClose(100, 100, 100)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("dev", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})

	t.Run("correctness", func(t *testing.T) {
		/* [10,20,30]: mean=20, MAD=(10+0+10)/3 = 20/3 */
		ctx := makeCtxClose(10, 20, 30)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("dev", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 20.0/3.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_Variance(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("variance", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("biased_population", func(t *testing.T) {
		/* [10,20,30]: mean=20, var=200/3 */
		ctx := makeCtxClose(10, 20, 30)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("variance", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 200.0/3.0, 1e-9)
	})

	t.Run("unbiased_sample", func(t *testing.T) {
		/* [10,20,30]: mean=20, var=200/2=100 */
		ctx := makeCtxClose(10, 20, 30)
		ev := NewStreamingBarEvaluator()
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "variance"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				lit(3),
				&ast.Literal{Value: false},
			},
		}
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 100.0, 1e-9)
	})

	t.Run("unbiased_length_one_returns_zero", func(t *testing.T) {
		ctx := makeCtxClose(42)
		ev := NewStreamingBarEvaluator()
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "variance"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				lit(1),
				&ast.Literal{Value: false},
			},
		}
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_Median(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("median", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("odd_length_middle_element", func(t *testing.T) {
		ctx := makeCtxClose(30, 10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("median", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 20.0, 1e-9)
	})

	t.Run("even_length_averages_two_middle", func(t *testing.T) {
		ctx := makeCtxClose(10, 40, 20, 30)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("median", "close", 4)
		got, err := ev.EvaluateAtBar(call, ctx, 3)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 25.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_Mode(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mode", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("clear_mode", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 10, 30, 10)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mode", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 10.0, 1e-9)
	})

	t.Run("tie_break_selects_smallest", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 30, 10, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mode", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 4)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 10.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_CMO(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 15, 25, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("cmo", "close", 4)
		for _, barIdx := range []int{0, 1, 2, 3} {
			got, err := ev.EvaluateAtBar(call, ctx, barIdx)
			if err != nil {
				t.Fatalf("bar %d: %v", barIdx, err)
			}
			if !math.IsNaN(got) {
				t.Errorf("bar %d: expected NaN during warmup, got %f", barIdx, got)
			}
		}
	})

	t.Run("gains_and_losses", func(t *testing.T) {
		/* close=[10,20,15,25,20], length=4: sumGain=20, sumLoss=10, CMO=100*(20-10)/30 */
		ctx := makeCtxClose(10, 20, 15, 25, 20)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("cmo", "close", 4)
		got, err := ev.EvaluateAtBar(call, ctx, 4)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 100.0*10.0/30.0, 1e-9)
	})

	t.Run("flat_source_zero", func(t *testing.T) {
		ctx := makeCtxClose(100, 100, 100, 100, 100)
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("cmo", "close", 4)
		got, err := ev.EvaluateAtBar(call, ctx, 4)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_WPR(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{High: 105, Low: 95, Close: 102},
			{High: 107, Low: 97, Close: 104},
		})
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeWPRCall(3), ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("correctness", func(t *testing.T) {
		/* period=2, bar=2: hi=109, lo=97, close=106 → (109-106)/(109-97)*-100 = -25 */
		ctx := makeCtxFull([]context.OHLCV{
			{High: 105, Low: 95, Close: 102},
			{High: 107, Low: 97, Close: 104},
			{High: 109, Low: 99, Close: 106},
		})
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeWPRCall(2), ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, -25.0, 1e-9)
	})

	t.Run("hi_equals_lo_returns_zero", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{High: 100, Low: 100, Close: 100},
			{High: 100, Low: 100, Close: 100},
		})
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeWPRCall(2), ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 0.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_MFI(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{High: 100, Low: 100, Close: 100, Volume: 1000},
			{High: 200, Low: 200, Close: 200, Volume: 1000},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mfi", "hlc3", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("correctness", func(t *testing.T) {
		/* length=2, bar=2: posFlow=200000, negFlow=100000 → MFI=100-100/3=200/3 */
		ctx := makeCtxFull([]context.OHLCV{
			{High: 100, Low: 100, Close: 100, Volume: 1000},
			{High: 200, Low: 200, Close: 200, Volume: 1000},
			{High: 100, Low: 100, Close: 100, Volume: 1000},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mfi", "hlc3", 2)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 200.0/3.0, 0.001)
	})

	t.Run("all_positive_flow_returns_hundred", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{High: 100, Low: 100, Close: 100, Volume: 1000},
			{High: 200, Low: 200, Close: 200, Volume: 1000},
			{High: 300, Low: 300, Close: 300, Volume: 1000},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("mfi", "hlc3", 2)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 100.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_VWMA(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{Close: 10, Volume: 1},
			{Close: 20, Volume: 2},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("vwma", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("volume_weighted", func(t *testing.T) {
		/* (10*1+20*2+30*3)/(1+2+3) = 140/6 */
		ctx := makeCtxFull([]context.OHLCV{
			{Close: 10, Volume: 1},
			{Close: 20, Volume: 2},
			{Close: 30, Volume: 3},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("vwma", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 140.0/6.0, 1e-9)
	})

	t.Run("equal_volumes_is_simple_mean", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{
			{Close: 10, Volume: 1000},
			{Close: 20, Volume: 1000},
			{Close: 30, Volume: 1000},
		})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("vwma", "close", 3)
		got, err := ev.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 20.0, 1e-9)
	})

	t.Run("zero_volume_returns_nan", func(t *testing.T) {
		ctx := makeCtxFull([]context.OHLCV{{Close: 10, Volume: 0}})
		ev := NewStreamingBarEvaluator()
		call := createTACallExpression("vwma", "close", 1)
		got, err := ev.EvaluateAtBar(call, ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN for zero total volume, got %f", got)
		}
	})
}

func TestStreamingBarEvaluator_Linreg(t *testing.T) {
	t.Run("warmup_returns_nan", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 30)
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeLinregCall("close", 3, 0), ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !math.IsNaN(got) {
			t.Errorf("expected NaN during warmup, got %f", got)
		}
	})

	t.Run("perfect_linear_fit", func(t *testing.T) {
		ctx := makeCtxClose(10, 20, 30, 40, 50)
		cases := []struct {
			name   string
			barIdx int
			offset float64
			want   float64
		}{
			{"bar2_offset0", 2, 0, 30},
			{"bar4_offset0", 4, 0, 50},
			{"bar4_offset1", 4, 1, 40},
			{"bar4_offset2", 4, 2, 30},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				ev := NewStreamingBarEvaluator()
				got, err := ev.EvaluateAtBar(makeLinregCall("close", 3, c.offset), ctx, c.barIdx)
				if err != nil {
					t.Fatal(err)
				}
				assertFloat64(t, "value", got, c.want, 1e-9)
			})
		}
	})

	t.Run("length_one_returns_single_value", func(t *testing.T) {
		ctx := makeCtxClose(42)
		ev := NewStreamingBarEvaluator()
		got, err := ev.EvaluateAtBar(makeLinregCall("close", 1, 0), ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, 42.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_HighestBars(t *testing.T) {
	ctx := makeCtxClose(10, 30, 20, 25, 15)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("highestbars", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, -1},
		{"bar3", 3, -2},
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

	t.Run("tie_uses_oldest_bar", func(t *testing.T) {
		ctx2 := makeCtxClose(30, 30, 20)
		ev2 := NewStreamingBarEvaluator()
		got, err := ev2.EvaluateAtBar(createTACallExpression("highestbars", "close", 3), ctx2, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, -2.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_LowestBars(t *testing.T) {
	ctx := makeCtxClose(10, 30, 20, 25, 15)
	ev := NewStreamingBarEvaluator()
	call := createTACallExpression("lowestbars", "close", 3)

	cases := []struct {
		name   string
		barIdx int
		want   float64
	}{
		{"warmup_bar0", 0, math.NaN()},
		{"warmup_bar1", 1, math.NaN()},
		{"bar2", 2, -2},
		{"bar3", 3, -1},
		{"bar4", 4, 0},
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

	t.Run("tie_uses_oldest_bar", func(t *testing.T) {
		ctx2 := makeCtxClose(10, 20, 10)
		ev2 := NewStreamingBarEvaluator()
		got, err := ev2.EvaluateAtBar(createTACallExpression("lowestbars", "close", 3), ctx2, 2)
		if err != nil {
			t.Fatal(err)
		}
		assertFloat64(t, "value", got, -2.0, 1e-9)
	})
}

func TestStreamingBarEvaluator_WindowInsufficientArguments(t *testing.T) {
	ctx := makeCtxClose(10)

	zeroArgTA := func(funcName string) *ast.CallExpression {
		return &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: funcName},
			},
		}
	}

	cases := []struct {
		name string
		call *ast.CallExpression
	}{
		{"highest", zeroArgTA("highest")},
		{"lowest", zeroArgTA("lowest")},
		{"sum", zeroArgTA("sum")},
		{"range", zeroArgTA("range")},
		{"dev", zeroArgTA("dev")},
		{"variance", zeroArgTA("variance")},
		{"median", zeroArgTA("median")},
		{"mode", zeroArgTA("mode")},
		{"cmo", zeroArgTA("cmo")},
		{"wpr", zeroArgTA("wpr")},
		{"mfi", zeroArgTA("mfi")},
		{"vwma", zeroArgTA("vwma")},
		{"linreg", zeroArgTA("linreg")},
		{"highestbars", zeroArgTA("highestbars")},
		{"lowestbars", zeroArgTA("lowestbars")},
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
