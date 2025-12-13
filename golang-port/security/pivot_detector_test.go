package security

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func TestPivotDetector_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name      string
		leftBars  int
		rightBars int
		dataLen   int
		testIdx   int
		expectNaN bool
		desc      string
	}{
		{"insufficient_left_bars", 5, 5, 20, 3, true, "barIdx < leftBars"},
		{"exact_left_boundary", 5, 5, 20, 7, false, "barIdx == leftBars with valid pivot at index 7"},
		{"insufficient_right_bars", 5, 5, 20, 16, true, "barIdx + rightBars >= dataLen"},
		{"exact_right_boundary", 5, 5, 20, 14, false, "barIdx + rightBars == dataLen - 1"},
		{"start_of_data", 2, 2, 10, 0, true, "cannot detect at start"},
		{"end_of_data", 2, 2, 10, 9, true, "cannot detect at end"},
		{"valid_middle", 3, 3, 15, 7, false, "sufficient bars both sides"},
	}

	data := make([]context.OHLCV, 20)
	for i := range data {
		data[i] = context.OHLCV{
			High: float64(100 + i*2),
			Low:  float64(90 + i*2),
		}
	}
	data[5].High = 200
	data[7].High = 210
	data[14].High = 190

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPivotDetector(tt.leftBars, tt.rightBars)
			result := detector.DetectHighAtBar(data[:tt.dataLen], "high", tt.testIdx)

			if tt.expectNaN && !math.IsNaN(result) {
				t.Errorf("%s: expected NaN, got %.2f", tt.desc, result)
			}
			if !tt.expectNaN && math.IsNaN(result) {
				t.Errorf("%s: expected valid pivot, got NaN", tt.desc)
			}
		})
	}
}

func TestPivotDetector_WindowSizeVariations(t *testing.T) {
	data := []context.OHLCV{
		{High: 100}, {High: 102}, {High: 104}, {High: 106}, {High: 108},
		{High: 110}, {High: 108}, {High: 106}, {High: 104}, {High: 102}, {High: 100},
	}

	tests := []struct {
		name      string
		leftBars  int
		rightBars int
		testIdx   int
		wantPivot bool
		desc      string
	}{
		{"symmetric_small", 2, 2, 5, true, "window [3,4,5,6,7]"},
		{"symmetric_large", 5, 5, 5, true, "window [0,1,2,3,4,5,6,7,8,9,10]"},
		{"asymmetric_left_heavy", 4, 2, 5, true, "window [1,2,3,4,5,6,7]"},
		{"asymmetric_right_heavy", 2, 4, 5, true, "window [3,4,5,6,7,8,9]"},
		{"single_bar_each_side", 1, 1, 5, true, "window [4,5,6]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPivotDetector(tt.leftBars, tt.rightBars)
			result := detector.DetectHighAtBar(data, "high", tt.testIdx)

			isPivot := !math.IsNaN(result)
			if isPivot != tt.wantPivot {
				t.Errorf("%s: wantPivot=%v, got isPivot=%v (value=%.2f)",
					tt.desc, tt.wantPivot, isPivot, result)
			}
			if tt.wantPivot && result != 110 {
				t.Errorf("%s: expected pivot value 110, got %.2f", tt.desc, result)
			}
		})
	}
}

func TestPivotDetector_DataPatterns(t *testing.T) {
	tests := []struct {
		name      string
		data      []context.OHLCV
		leftBars  int
		rightBars int
		testIdx   int
		wantPivot bool
		wantValue float64
		desc      string
	}{
		{
			name: "flat_no_pivot",
			data: []context.OHLCV{
				{High: 100}, {High: 100}, {High: 100}, {High: 100}, {High: 100},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: false,
			desc:      "all values equal",
		},
		{
			name: "monotonic_increasing",
			data: []context.OHLCV{
				{High: 100}, {High: 101}, {High: 102}, {High: 103}, {High: 104},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: false,
			desc:      "strictly increasing",
		},
		{
			name: "monotonic_decreasing",
			data: []context.OHLCV{
				{High: 104}, {High: 103}, {High: 102}, {High: 101}, {High: 100},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: false,
			desc:      "strictly decreasing",
		},
		{
			name: "single_peak",
			data: []context.OHLCV{
				{High: 100}, {High: 105}, {High: 110}, {High: 105}, {High: 100},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: true,
			wantValue: 110,
			desc:      "clear single peak",
		},
		{
			name: "plateau_no_pivot",
			data: []context.OHLCV{
				{High: 100}, {High: 110}, {High: 110}, {High: 110}, {High: 100},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: false,
			desc:      "plateau at peak",
		},
		{
			name: "equal_neighbor_left",
			data: []context.OHLCV{
				{High: 110}, {High: 110}, {High: 120}, {High: 105}, {High: 100},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: true,
			wantValue: 120,
			desc:      "left neighbor equals center",
		},
		{
			name: "equal_neighbor_right",
			data: []context.OHLCV{
				{High: 100}, {High: 105}, {High: 120}, {High: 120}, {High: 110},
			},
			leftBars:  1,
			rightBars: 1,
			testIdx:   2,
			wantPivot: false,
			desc:      "right neighbor equals center",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewPivotDetector(tt.leftBars, tt.rightBars)
			result := detector.DetectHighAtBar(tt.data, "high", tt.testIdx)

			isPivot := !math.IsNaN(result)
			if isPivot != tt.wantPivot {
				t.Errorf("%s: wantPivot=%v, got isPivot=%v (value=%.2f)",
					tt.desc, tt.wantPivot, isPivot, result)
			}
			if tt.wantPivot && result != tt.wantValue {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.wantValue, result)
			}
		})
	}
}

func TestPivotDetector_MultiplePivotsInSeries(t *testing.T) {
	data := []context.OHLCV{
		{High: 100, Low: 90},
		{High: 105, Low: 85},
		{High: 110, Low: 80},
		{High: 105, Low: 85},
		{High: 100, Low: 90},
		{High: 105, Low: 85},
		{High: 112, Low: 82},
		{High: 105, Low: 85},
		{High: 100, Low: 90},
	}

	detector := NewPivotDetector(2, 2)

	t.Run("first_pivot_high", func(t *testing.T) {
		result := detector.DetectHighAtBar(data, "high", 2)
		if result != 110 {
			t.Errorf("expected first pivot 110, got %.2f", result)
		}
	})

	t.Run("second_pivot_high", func(t *testing.T) {
		result := detector.DetectHighAtBar(data, "high", 6)
		if result != 112 {
			t.Errorf("expected second pivot 112, got %.2f", result)
		}
	})

	t.Run("non_pivot_between", func(t *testing.T) {
		result := detector.DetectHighAtBar(data, "high", 4)
		if !math.IsNaN(result) {
			t.Errorf("expected NaN for non-pivot, got %.2f", result)
		}
	})

	t.Run("first_pivot_low", func(t *testing.T) {
		result := detector.DetectLowAtBar(data, "low", 2)
		if result != 80 {
			t.Errorf("expected first pivot low 80, got %.2f", result)
		}
	})

	t.Run("second_pivot_low", func(t *testing.T) {
		result := detector.DetectLowAtBar(data, "low", 6)
		if result != 82 {
			t.Errorf("expected second pivot low 82, got %.2f", result)
		}
	})
}

func TestPivotDetector_FieldSourceVariations(t *testing.T) {
	data := []context.OHLCV{
		{High: 105, Low: 95, Close: 100, Open: 98},
		{High: 110, Low: 90, Close: 105, Open: 103},
		{High: 115, Low: 85, Close: 110, Open: 108},
		{High: 110, Low: 90, Close: 105, Open: 103},
		{High: 105, Low: 95, Close: 100, Open: 98},
	}

	detector := NewPivotDetector(1, 1)

	tests := []struct {
		field     string
		wantValue float64
		desc      string
	}{
		{"high", 115, "pivot on high field"},
		{"close", 110, "pivot on close field"},
		{"open", 108, "pivot on open field"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := detector.DetectHighAtBar(data, tt.field, 2)
			if result != tt.wantValue {
				t.Errorf("%s: expected %.2f, got %.2f", tt.desc, tt.wantValue, result)
			}
		})
	}

	t.Run("low_field", func(t *testing.T) {
		result := detector.DetectLowAtBar(data, "low", 2)
		if result != 85 {
			t.Errorf("expected pivot low 85, got %.2f", result)
		}
	})

	t.Run("invalid_field", func(t *testing.T) {
		result := detector.DetectHighAtBar(data, "invalid_field", 2)
		if !math.IsNaN(result) {
			t.Errorf("expected NaN for invalid field, got %.2f", result)
		}
	})
}

func TestPivotEvaluator_IntegrationWithBarEvaluator(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100, Low: 90},
			{High: 105, Low: 85},
			{High: 110, Low: 80},
			{High: 105, Low: 85},
			{High: 100, Low: 90},
			{High: 105, Low: 85},
			{High: 107, Low: 83},
			{High: 106, Low: 84},
			{High: 101, Low: 89},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("pivothigh_via_evaluator", func(t *testing.T) {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		result, err := evaluator.EvaluateAtBar(call, ctx, 2)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected pivot high 110, got %.2f", result)
		}
	})

	t.Run("pivotlow_via_evaluator", func(t *testing.T) {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivotlow"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "low"},
				&ast.Literal{Value: float64(2)},
				&ast.Literal{Value: float64(2)},
			},
		}

		result, err := evaluator.EvaluateAtBar(call, ctx, 6)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 83 {
			t.Errorf("expected pivot low 83, got %.2f", result)
		}
	})

	t.Run("insufficient_arguments", func(t *testing.T) {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
			},
		}

		_, err := evaluator.EvaluateAtBar(call, ctx, 2)
		if err == nil {
			t.Error("expected error for insufficient arguments")
		}
	})

	t.Run("non_numeric_period", func(t *testing.T) {
		call := &ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "pivothigh"},
			},
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "high"},
				&ast.Identifier{Name: "invalid"},
				&ast.Literal{Value: float64(2)},
			},
		}

		_, err := evaluator.EvaluateAtBar(call, ctx, 2)
		if err == nil {
			t.Error("expected error for non-numeric period")
		}
	})
}

func TestPivotDetector_MemberExpressionSupport(t *testing.T) {
	ctx := &context.Context{
		Data: []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110},
			{High: 108}, {High: 103}, {High: 102},
			{High: 104}, {High: 107}, {High: 106}, {High: 101},
		},
	}

	evaluator := NewStreamingBarEvaluator()

	t.Run("subscript_offset_1", func(t *testing.T) {
		memberExpr := &ast.MemberExpression{
			Object: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: float64(2)},
					&ast.Literal{Value: float64(2)},
				},
			},
			Property: &ast.Literal{Value: float64(1)},
		}

		result, err := evaluator.EvaluateAtBar(memberExpr, ctx, 3)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected pivot[1] = 110 at bar 3, got %.2f", result)
		}
	})

	t.Run("subscript_offset_2", func(t *testing.T) {
		memberExpr := &ast.MemberExpression{
			Object: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: float64(2)},
					&ast.Literal{Value: float64(2)},
				},
			},
			Property: &ast.Literal{Value: float64(2)},
		}

		result, err := evaluator.EvaluateAtBar(memberExpr, ctx, 4)
		if err != nil {
			t.Fatalf("EvaluateAtBar failed: %v", err)
		}
		if result != 110 {
			t.Errorf("expected pivot[2] = 110 at bar 4, got %.2f", result)
		}
	})

	t.Run("subscript_out_of_bounds", func(t *testing.T) {
		memberExpr := &ast.MemberExpression{
			Object: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "pivothigh"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: float64(2)},
					&ast.Literal{Value: float64(2)},
				},
			},
			Property: &ast.Literal{Value: float64(10)},
		}

		_, err := evaluator.EvaluateAtBar(memberExpr, ctx, 5)
		if err == nil {
			t.Error("expected error for out-of-bounds subscript")
		}
	})
}

func TestPivotDetector_EmptyAndSmallDatasets(t *testing.T) {
	detector := NewPivotDetector(2, 2)

	tests := []struct {
		name    string
		data    []context.OHLCV
		testIdx int
		desc    string
	}{
		{"empty_data", []context.OHLCV{}, 0, "zero length array"},
		{"single_bar", []context.OHLCV{{High: 100}}, 0, "one bar only"},
		{"two_bars", []context.OHLCV{{High: 100}, {High: 110}}, 1, "two bars only"},
		{"exact_window_size", []context.OHLCV{
			{High: 100}, {High: 105}, {High: 110}, {High: 105}, {High: 100},
		}, 2, "exactly leftBars + 1 + rightBars"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.DetectHighAtBar(tt.data, "high", tt.testIdx)
			if len(tt.data) < 5 && !math.IsNaN(result) {
				t.Errorf("%s: expected NaN for insufficient data, got %.2f", tt.desc, result)
			}
			if len(tt.data) == 5 && tt.testIdx == 2 && result != 110 {
				t.Errorf("%s: expected pivot 110, got %.2f", tt.desc, result)
			}
		})
	}
}
