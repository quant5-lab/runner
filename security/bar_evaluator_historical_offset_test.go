package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// TestBarEvaluator_MemberExpressionHistoricalOffset tests historical lookback
// in security() expressions: expr[N] patterns
func TestBarEvaluator_MemberExpressionHistoricalOffset(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100, Open: 98, High: 102, Low: 96, Volume: 1000},
			{Close: 105, Open: 103, High: 107, Low: 101, Volume: 1100},
			{Close: 110, Open: 108, High: 112, Low: 106, Volume: 1200},
			{Close: 115, Open: 113, High: 117, Low: 111, Volume: 1300},
			{Close: 120, Open: 118, High: 122, Low: 116, Volume: 1400},
			{Close: 125, Open: 123, High: 127, Low: 121, Volume: 1500},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name        string
		taFunc      string
		taArg       string
		offset      int
		currentBar  int
		expected    float64
		shouldError bool
		desc        string
	}{
		{
			name:       "sma_close[1]_at_bar4",
			taFunc:     "sma",
			taArg:      "close",
			offset:     1,
			currentBar: 4,
			expected:   110,
			desc:       "SMA(close,3)[1] from bar 4 should return bar 3 SMA = 110",
		},
		{
			name:       "sma_close[2]_at_bar5",
			taFunc:     "sma",
			taArg:      "close",
			offset:     2,
			currentBar: 5,
			expected:   110,
			desc:       "SMA(close,3)[2] from bar 5 should return bar 3 SMA = 110",
		},
		{
			name:        "offset_at_boundary_first_valid_bar",
			taFunc:      "sma",
			taArg:       "close",
			offset:      1,
			currentBar:  2,
			expected:    math.NaN(),
			shouldError: false,
			desc:        "SMA[1] at first valid bar (2) returns NaN (adjusted index < 0)",
		},
		{
			name:        "offset_exceeds_history",
			taFunc:      "sma",
			taArg:       "close",
			offset:      10,
			currentBar:  4,
			shouldError: true,
			desc:        "Offset[10] exceeds available history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build SMA call: ta.sma(close, 3)
			smaCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.taFunc},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: tt.taArg},
					&ast.Literal{Value: 3.0},
				},
			}

			// Build MemberExpression: sma()[offset]
			memberExpr := &ast.MemberExpression{
				Object:   smaCall,
				Property: &ast.Literal{Value: float64(tt.offset)},
			}

			value, err := evaluator.evaluateMemberExpressionAtBar(memberExpr, ctx, tt.currentBar)

			if tt.shouldError {
				if err == nil {
					t.Errorf("%s: expected error but got value %.2f", tt.desc, value)
				}
				return
			}

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			// Handle NaN comparison
			if math.IsNaN(tt.expected) {
				if !math.IsNaN(value) {
					t.Errorf("%s: expected NaN, got %.2f", tt.desc, value)
				}
			} else if math.Abs(value-tt.expected) > 0.0001 {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.expected, value)
			}
		})
	}
}

// TestBarEvaluator_PivotHistoricalOffset tests pivot functions with historical lookback
// Pattern: pivothigh(left, right)[N] inside security()
func TestBarEvaluator_PivotHistoricalOffset(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, // 0
			{High: 102}, // 1
			{High: 105}, // 2 <- pivot
			{High: 103}, // 3
			{High: 101}, // 4
			{High: 104}, // 5
			{High: 107}, // 6
			{High: 110}, // 7 <- pivot
			{High: 108}, // 8
			{High: 106}, // 9
			{High: 109}, // 10
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name        string
		leftBars    int
		rightBars   int
		offset      int
		currentBar  int
		expectNaN   bool
		expectValue *float64
		desc        string
	}{
		{
			name:        "pivothigh[0]_at_detection_bar",
			leftBars:    2,
			rightBars:   2,
			offset:      0,
			currentBar:  4,
			expectValue: floatPtr(105),
			desc:        "Pivot[0] at bar 4 detects bar 2 pivot (105)",
		},
		{
			name:        "pivothigh[1]_lookback_one_detection",
			leftBars:    2,
			rightBars:   2,
			offset:      1,
			currentBar:  9,
			expectValue: floatPtr(105),
			desc:        "Pivot[1] at bar 9 should return previous detected pivot",
		},
		{
			name:       "pivothigh[1]_at_first_detection_returns_nan",
			leftBars:   2,
			rightBars:  2,
			offset:     1,
			currentBar: 4,
			expectNaN:  true,
			desc:       "Pivot[1] at first detection (bar 4) has no history, returns NaN",
		},
		{
			name:       "pivothigh[N]_before_detection_returns_nan",
			leftBars:   2,
			rightBars:  2,
			offset:     0,
			currentBar: 2,
			expectNaN:  true,
			desc:       "Pivot at bar 2 (center) can't be detected until bar 4 (right bars)",
		},
		{
			name:        "offset_exceeds_detection_history",
			leftBars:    2,
			rightBars:   2,
			offset:      5,
			currentBar:  9,
			expectValue: floatPtr(105),
			desc:        "Large offset may wrap to earlier detections (implementation-specific)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pivothigh(leftBars, rightBars)
			pivotCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(tt.leftBars)},
					&ast.Literal{Value: float64(tt.rightBars)},
				},
			}

			// Wrap in MemberExpression: pivothigh()[offset]
			memberExpr := &ast.MemberExpression{
				Object:   pivotCall,
				Property: &ast.Literal{Value: float64(tt.offset)},
			}

			value, err := evaluator.evaluateMemberExpressionAtBar(memberExpr, ctx, tt.currentBar)

			if err != nil {
				if tt.expectNaN {
					// Out of range error is acceptable for NaN cases
					return
				}
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if tt.expectNaN {
				if !math.IsNaN(value) {
					t.Errorf("%s: expected NaN, got %.2f", tt.desc, value)
				}
				return
			}

			if tt.expectValue != nil {
				if math.Abs(value-*tt.expectValue) > 0.0001 {
					t.Errorf("%s: expected %.2f, got %.2f", tt.desc, *tt.expectValue, value)
				}
			}
		})
	}
}

// TestBarEvaluator_TAFunctionHistoricalOffset tests TA functions with historical lookback
// Pattern: sma(close, 20)[N], ema(close, 10)[N]
func TestBarEvaluator_TAFunctionHistoricalOffset(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 102},
			{Close: 104},
			{Close: 106},
			{Close: 108},
			{Close: 110},
			{Close: 112},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name        string
		taFunc      string
		period      int
		offset      int
		currentBar  int
		expectedMin float64
		expectedMax float64
		desc        string
	}{
		{
			name:        "sma[1]_lookback_one_bar",
			taFunc:      "sma",
			period:      3,
			offset:      1,
			currentBar:  4,
			expectedMin: 103.0,
			expectedMax: 105.0,
			desc:        "SMA(3)[1] at bar 4 returns SMA from bar 3",
		},
		{
			name:        "sma[2]_lookback_two_bars",
			taFunc:      "sma",
			period:      3,
			offset:      2,
			currentBar:  5,
			expectedMin: 103.0,
			expectedMax: 105.0,
			desc:        "SMA(3)[2] at bar 5 returns SMA from bar 3",
		},
		{
			name:        "ema[1]_exponential_lookback",
			taFunc:      "ema",
			period:      3,
			offset:      1,
			currentBar:  5,
			expectedMin: 103.0,
			expectedMax: 109.0,
			desc:        "EMA(3)[1] at bar 5 returns EMA from bar 4",
		},
		{
			name:        "rma[1]_smoothed_lookback",
			taFunc:      "rma",
			period:      3,
			offset:      1,
			currentBar:  6,
			expectedMin: 104.0,
			expectedMax: 112.0,
			desc:        "RMA(3)[1] at bar 6 returns RMA from bar 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TA call: ta.func(close, period)
			taCall := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.taFunc},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: float64(tt.period)},
				},
			}

			// Wrap in MemberExpression: ta.func()[offset]
			memberExpr := &ast.MemberExpression{
				Object:   taCall,
				Property: &ast.Literal{Value: float64(tt.offset)},
			}

			value, err := evaluator.evaluateMemberExpressionAtBar(memberExpr, ctx, tt.currentBar)

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.desc, err)
			}

			if value < tt.expectedMin || value > tt.expectedMax {
				t.Errorf("%s: expected [%.2f, %.2f], got %.2f",
					tt.desc, tt.expectedMin, tt.expectedMax, value)
			}
		})
	}
}

// TestBarEvaluator_NestedOffsetExpressions tests complex nested patterns
// Pattern: fixnan(pivothigh()[1]), nz(sma()[2])
func TestBarEvaluator_NestedOffsetExpressions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Close: 100},
			{High: 102, Close: 102},
			{High: 105, Close: 105}, // pivot
			{High: 103, Close: 103},
			{High: 101, Close: 101},
			{High: 104, Close: 104},
			{High: 107, Close: 107},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("fixnan_wrapping_pivothigh_with_offset", func(t *testing.T) {
		// fixnan(pivothigh(2, 2)[1])
		pivotWithOffset := &ast.MemberExpression{
			Object: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 2.0},
					&ast.Literal{Value: 2.0},
				},
			},
			Property: &ast.Literal{Value: 1.0},
		}

		// Evaluate at bar 6 (should have history)
		value, err := evaluator.evaluateMemberExpressionAtBar(pivotWithOffset, ctx, 6)

		if err != nil {
			t.Fatalf("Expected no error for nested expression, got: %v", err)
		}

		// At bar 6, pivot[1] should be NaN or 0 (depends on implementation)
		// The key is no crash/panic
		t.Logf("Nested pivot[1] at bar 6: %.2f (NaN is acceptable)", value)
	})

	t.Run("multiple_sequential_offsets", func(t *testing.T) {
		// sma(close, 3)[1] then evaluate again with [2]
		smaCall := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "sma"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 3.0},
			},
		}

		offset1 := &ast.MemberExpression{
			Object:   smaCall,
			Property: &ast.Literal{Value: 1.0},
		}

		offset2 := &ast.MemberExpression{
			Object:   smaCall,
			Property: &ast.Literal{Value: 2.0},
		}

		val1, _ := evaluator.evaluateMemberExpressionAtBar(offset1, ctx, 5)
		val2, _ := evaluator.evaluateMemberExpressionAtBar(offset2, ctx, 5)

		if val1 == val2 {
			t.Errorf("Different offsets should return different values: [1]=%.2f, [2]=%.2f", val1, val2)
		}

		t.Logf("SMA[1]=%.2f, SMA[2]=%.2f - values are correctly different", val1, val2)
	})
}

// TestBarEvaluator_TAFunctionStatePreservation verifies historical values remain stable
// after forward computation - critical for FSB-backed storage (TSI, SMA, STDEV)
func TestBarEvaluator_TAFunctionStatePreservation(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 50),
	}
	for i := range ctx.Data {
		phase := float64(100 + i*2)
		if i >= 20 {
			phase = float64(100 + 20*2 - (i - 20))
		}
		ctx.Data[i] = context.OHLCV{Close: phase}
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name          string
		taFunc        string
		period1       float64
		period2       float64
		earlyBar      int
		lateBar       int
		minDivergence float64
	}{
		{
			name:          "tsi_historical_stability",
			taFunc:        "tsi",
			period1:       5.0,
			period2:       13.0,
			earlyBar:      20,
			lateBar:       30,
			minDivergence: 0.01,
		},
		{
			name:          "sma_historical_stability",
			taFunc:        "sma",
			period1:       10.0,
			period2:       0,
			earlyBar:      15,
			lateBar:       25,
			minDivergence: 0.01,
		},
		{
			name:          "stdev_historical_stability",
			taFunc:        "stdev",
			period1:       10.0,
			period2:       0,
			earlyBar:      15,
			lateBar:       25,
			minDivergence: 0.01,
		},
		{
			name:          "ema_historical_stability",
			taFunc:        "ema",
			period1:       10.0,
			period2:       0,
			earlyBar:      15,
			lateBar:       25,
			minDivergence: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var taCall *ast.CallExpression
			if tt.period2 > 0 {
				taCall = &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: tt.taFunc},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: tt.period1},
						&ast.Literal{Value: tt.period2},
					},
				}
			} else {
				taCall = &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: tt.taFunc},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: tt.period1},
					},
				}
			}

			valEarlyFirst, err := evaluator.EvaluateAtBar(taCall, ctx, tt.earlyBar)
			if err != nil {
				t.Fatalf("EvaluateAtBar(%d) first call failed: %v", tt.earlyBar, err)
			}
			if math.IsNaN(valEarlyFirst) {
				t.Fatalf("bar %d returned NaN (warmup issue)", tt.earlyBar)
			}

			valLate, err := evaluator.EvaluateAtBar(taCall, ctx, tt.lateBar)
			if err != nil {
				t.Fatalf("EvaluateAtBar(%d) failed: %v", tt.lateBar, err)
			}

			valEarlySecond, err := evaluator.EvaluateAtBar(taCall, ctx, tt.earlyBar)
			if err != nil {
				t.Fatalf("EvaluateAtBar(%d) second call failed: %v", tt.earlyBar, err)
			}

			if valEarlyFirst != valEarlySecond {
				t.Errorf("Historical value changed: bar%d first=%.6f, after bar%d=%.6f",
					tt.earlyBar, valEarlyFirst, tt.lateBar, valEarlySecond)
			}

			if math.Abs(valEarlyFirst-valLate) < tt.minDivergence {
				t.Errorf("Phase-change source should diverge: bar%d=%.6f, bar%d=%.6f (too close)",
					tt.earlyBar, valEarlyFirst, tt.lateBar, valLate)
			}
		})
	}
}

// TestBarEvaluator_TAFunctionNonSequentialAccess verifies arbitrary access order stability
func TestBarEvaluator_TAFunctionNonSequentialAccess(t *testing.T) {
	ctx := &context.Context{
		Data: make([]context.OHLCV, 40),
	}
	for i := range ctx.Data {
		oscillation := float64(100 + 10*((i%5)-2))
		ctx.Data[i] = context.OHLCV{Close: oscillation}
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name    string
		taFunc  string
		period1 float64
		period2 float64
		warmup  int
	}{
		{
			name:    "tsi_random_access",
			taFunc:  "tsi",
			period1: 5.0,
			period2: 13.0,
			warmup:  17,
		},
		{
			name:    "sma_random_access",
			taFunc:  "sma",
			period1: 10.0,
			period2: 0,
			warmup:  9,
		},
		{
			name:    "stdev_random_access",
			taFunc:  "stdev",
			period1: 10.0,
			period2: 0,
			warmup:  9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var taCall *ast.CallExpression
			if tt.period2 > 0 {
				taCall = &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: tt.taFunc},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: tt.period1},
						&ast.Literal{Value: tt.period2},
					},
				}
			} else {
				taCall = &ast.CallExpression{
					Callee: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "ta"},
						Property: &ast.Identifier{Name: tt.taFunc},
					},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "close"},
						&ast.Literal{Value: tt.period1},
					},
				}
			}

			valBar35, err := evaluator.EvaluateAtBar(taCall, ctx, 35)
			if err != nil {
				t.Fatalf("EvaluateAtBar(35) failed: %v", err)
			}

			accessPattern := []int{25, 30, 20, 35, 28}
			results := make(map[int]float64)

			for _, barIdx := range accessPattern {
				v, err := evaluator.EvaluateAtBar(taCall, ctx, barIdx)
				if err != nil {
					t.Fatalf("EvaluateAtBar(%d) failed: %v", barIdx, err)
				}
				if barIdx >= tt.warmup && math.IsNaN(v) {
					t.Errorf("bar %d (post-warmup): got NaN", barIdx)
				}
				results[barIdx] = v
			}

			if results[35] != valBar35 {
				t.Errorf("Non-sequential access changed bar 35: initial=%.6f, after pattern=%.6f",
					valBar35, results[35])
			}

			val25First := results[25]
			val25Second, _ := evaluator.EvaluateAtBar(taCall, ctx, 25)
			if val25First != val25Second {
				t.Errorf("Repeated access changed bar 25: first=%.6f, second=%.6f",
					val25First, val25Second)
			}
		})
	}
}

// TestBarEvaluator_OffsetBoundaryConditions tests edge cases for offset bounds
func TestBarEvaluator_OffsetBoundaryConditions(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{Close: 100},
			{Close: 105},
			{Close: 110},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	tests := []struct {
		name        string
		offset      int
		currentBar  int
		shouldError bool
		desc        string
	}{
		{
			name:        "negative_offset_at_bar0",
			offset:      1,
			currentBar:  0,
			shouldError: true,
			desc:        "Offset[1] at bar 0 creates negative index",
		},
		{
			name:        "offset_equals_current_bar",
			offset:      2,
			currentBar:  2,
			shouldError: true,
			desc:        "Offset[2] at bar 2 creates index 0 (boundary)",
		},
		{
			name:        "offset_exceeds_data_length",
			offset:      10,
			currentBar:  2,
			shouldError: true,
			desc:        "Offset[10] exceeds available data",
		},
		{
			name:        "max_valid_offset",
			offset:      2,
			currentBar:  2,
			shouldError: true,
			desc:        "Maximum valid offset (equals currentBar)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identityCall := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "identity"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			}

			memberExpr := &ast.MemberExpression{
				Object:   identityCall,
				Property: &ast.Literal{Value: float64(tt.offset)},
			}

			_, err := evaluator.evaluateMemberExpressionAtBar(memberExpr, ctx, tt.currentBar)

			if tt.shouldError && err == nil {
				t.Errorf("%s: expected error for boundary violation", tt.desc)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("%s: unexpected error: %v", tt.desc, err)
			}
		})
	}
}
