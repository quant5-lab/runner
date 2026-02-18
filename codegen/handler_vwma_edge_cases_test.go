package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestVWMAHandler_EdgeCases(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	t.Run("empty_arguments", func(t *testing.T) {
		call := &ast.CallExpression{Arguments: []ast.Expression{}}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for empty arguments")
		}
	})

	t.Run("single_argument_only", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for missing period argument")
		}
	})

	t.Run("extra_arguments_ignored", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
				&ast.Literal{Value: 999},
			},
		}
		code, err := handler.GenerateCode(gen, "test", call)
		if err != nil {
			t.Fatalf("Should accept extra arguments, got error: %v", err)
		}
		if code == "" {
			t.Error("Should generate code with extra arguments")
		}
	})

	t.Run("empty_variable_name", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}
		code, err := handler.GenerateCode(gen, "", call)
		if err != nil {
			t.Fatalf("Should handle empty varName, got error: %v", err)
		}
		if code == "" {
			t.Error("Should generate code even with empty varName")
		}
	})

	t.Run("boundary_periods", func(t *testing.T) {
		periods := []int{1, 2, 100, 5000}
		for _, period := range periods {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: period},
				},
			}
			code, err := handler.GenerateCode(gen, "test", call)
			if err != nil {
				t.Errorf("Period %d should generate code, got error: %v", period, err)
			}
			if code == "" {
				t.Errorf("Period %d generated empty code", period)
			}
		}
	})

	t.Run("zero_negative_periods", func(t *testing.T) {
		periods := []int{0, -1, -14}
		for _, period := range periods {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: period},
				},
			}
			_, _ = handler.GenerateCode(gen, "test", call)
		}
	})
}

func TestVWMAHandler_PeriodBoundaries(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	testCases := []struct {
		period    int
		warmupBar int
	}{
		{1, 0},
		{2, 1},
		{14, 13},
		{20, 19},
		{100, 99},
	}

	for _, tc := range testCases {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: tc.period},
			},
		}

		code, err := handler.GenerateCode(gen, "vwma", call)
		if err != nil {
			t.Fatalf("GenerateCode(%d) error = %v", tc.period, err)
		}

		if !strings.Contains(code, "if ctx.BarIndex <") {
			t.Errorf("Missing warmup check for period %d", tc.period)
		}
	}
}

func TestVWMAHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "vwma14", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("volume_weighted_accumulation", func(t *testing.T) {
		if !strings.Contains(code, "weightedSum") {
			t.Error("Missing weightedSum variable for weighted accumulation")
		}
		if !strings.Contains(code, "volumeSum") {
			t.Error("Missing volumeSum variable for volume accumulation")
		}
		if !strings.Contains(code, "Volume") {
			t.Error("Algorithm must use bar.Volume for weighting")
		}
	})

	t.Run("weighted_sum_formula", func(t *testing.T) {
		if !strings.Contains(code, "weightedSum +=") || !strings.Contains(code, "volumeSum +=") {
			t.Error("Missing accumulation operators for weighted sum and volume sum")
		}
		if !strings.Contains(code, "Volume") {
			t.Error("Missing volume access in weighted calculation")
		}
	})

	t.Run("vwma_calculation", func(t *testing.T) {
		if !strings.Contains(code, "weightedSum / volumeSum") {
			t.Error("Missing VWMA formula: weightedSum / volumeSum")
		}
	})

	t.Run("nan_handling", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Missing NaN validation")
		}
		if !strings.Contains(code, "hasNaN") {
			t.Error("Missing hasNaN tracking variable")
		}
	})

	t.Run("window_iteration", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j <") && !strings.Contains(code, "for i := 0; i <") {
			t.Error("Missing window iteration loop")
		}
	})

	t.Run("zero_volume_protection", func(t *testing.T) {
		if !strings.Contains(code, "volumeSum") {
			t.Error("Should track volume sum for zero-volume protection")
		}
	})
}

func TestVWMAHandler_SourceTypes(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	testCases := []struct {
		name   string
		source ast.Expression
	}{
		{"close", &ast.Identifier{Name: "close"}},
		{"open", &ast.Identifier{Name: "open"}},
		{"high", &ast.Identifier{Name: "high"}},
		{"low", &ast.Identifier{Name: "low"}},
		{"hl2", &ast.Identifier{Name: "hl2"}},
		{"hlc3", &ast.Identifier{Name: "hlc3"}},
		{"ohlc4", &ast.Identifier{Name: "ohlc4"}},
		{
			"complex_expression",
			&ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "high"},
				Operator: "+",
				Right:    &ast.Identifier{Name: "low"},
			},
		},
		{
			"nested_ta_call",
			&ast.CallExpression{
				Callee:    &ast.Identifier{Name: "ta.sma"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 10}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{tc.source, &ast.Literal{Value: 14}},
			}

			code, err := handler.GenerateCode(gen, "vwma", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if code == "" {
				t.Error("Generated code should not be empty")
			}

			if !strings.Contains(code, "weightedSum") {
				t.Error("All sources should produce volume-weighted calculation")
			}
		})
	}
}

func TestVWMAHandler_PeriodExpressions(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	testCases := []struct {
		name   string
		period ast.Expression
		valid  bool
	}{
		{"literal_constant", &ast.Literal{Value: 14}, true},
		{"small_period", &ast.Literal{Value: 1}, true},
		{"large_period", &ast.Literal{Value: 200}, true},
		{"dynamic_period", &ast.Identifier{Name: "myPeriod"}, false},
		{
			"expression_period",
			&ast.BinaryExpression{
				Left:     &ast.Literal{Value: 10},
				Operator: "+",
				Right:    &ast.Literal{Value: 4},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					tc.period,
				},
			}

			code, err := handler.GenerateCode(gen, "vwma", call)
			if tc.valid {
				if err != nil {
					t.Errorf("Should accept period, got error: %v", err)
				}
				if code == "" {
					t.Error("Should generate code for valid period")
				}
			} else {
				if err == nil {
					t.Error("Should reject invalid period")
				}
			}
		})
	}
}

func TestVWMAHandler_CodeStructureInvariants(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "vwma14", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("proper_indentation", func(t *testing.T) {
		lines := strings.Split(code, "\n")
		hasIndentation := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
				hasIndentation = true
				break
			}
		}
		if !hasIndentation {
			t.Error("Code should have proper indentation")
		}
	})

	t.Run("balanced_braces", func(t *testing.T) {
		openCount := strings.Count(code, "{")
		closeCount := strings.Count(code, "}")
		if openCount != closeCount {
			t.Errorf("Unbalanced braces: %d open, %d close", openCount, closeCount)
		}
	})

	t.Run("series_buffer_usage", func(t *testing.T) {
		if !strings.Contains(code, "Series.Set") || !strings.Contains(code, "Series.Get") {
			t.Error("Should use ForwardSeriesBuffer pattern (Set/Get)")
		}
	})

	t.Run("warmup_guard", func(t *testing.T) {
		if !strings.Contains(code, "if ctx.BarIndex <") {
			t.Error("Missing warmup period guard")
		}
		if !strings.Contains(code, "Series.Set(math.NaN())") {
			t.Error("Warmup period should set NaN via Series.Set")
		}
	})

	t.Run("volume_access_pattern", func(t *testing.T) {
		if !strings.Contains(code, "bar.Volume") && !strings.Contains(code, ".Volume") {
			t.Error("Should access volume from bar data")
		}
	})
}

func TestVWMAHandler_IntegrationWithRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("registry_has_vwma_handler", func(t *testing.T) {
		handler := registry.FindHandler("ta.vwma")
		if handler == nil {
			t.Fatal("TAFunctionRegistry missing VWMA handler")
		}

		if _, ok := handler.(*VWMAHandler); !ok {
			t.Errorf("Registry returned wrong handler type: %T", handler)
		}
	})

	t.Run("registry_supports_both_forms", func(t *testing.T) {
		if !registry.IsSupported("ta.vwma") {
			t.Error("TAFunctionRegistry should support 'ta.vwma'")
		}

		if !registry.IsSupported("vwma") {
			t.Error("TAFunctionRegistry should support 'vwma'")
		}
	})

	t.Run("registry_generates_code", func(t *testing.T) {
		gen := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20},
			},
		}

		code, err := registry.GenerateInlineTA(gen, "vwma20", "ta.vwma", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("GenerateInlineTA() returned empty string")
		}

		if !strings.Contains(code, "weightedSum") {
			t.Error("Registry-generated code should contain VWMA algorithm")
		}
	})

	t.Run("bare_alias_generates_identical_code", func(t *testing.T) {
		gen1 := createTestGenerator()
		gen2 := createTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
			},
		}

		code1, err1 := registry.GenerateInlineTA(gen1, "test", "ta.vwma", call)
		code2, err2 := registry.GenerateInlineTA(gen2, "test", "vwma", call)

		if err1 != nil || err2 != nil {
			t.Fatalf("Errors: %v, %v", err1, err2)
		}

		if code1 != code2 {
			t.Error("Namespaced and bare forms should generate identical code")
		}
	})
}

func TestVWMAHandler_ComparisonWithWMA(t *testing.T) {
	vwmaHandler := &VWMAHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	vwmaCode, err := vwmaHandler.GenerateCode(gen, "indicator", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("vwma_uses_volume_weighting", func(t *testing.T) {
		if !strings.Contains(vwmaCode, "Volume") {
			t.Error("VWMA must use volume for weighting (unlike simple WMA)")
		}
	})

	t.Run("vwma_has_dual_accumulation", func(t *testing.T) {
		if !strings.Contains(vwmaCode, "weightedSum") {
			t.Error("VWMA must accumulate weighted values")
		}
		if !strings.Contains(vwmaCode, "volumeSum") {
			t.Error("VWMA must accumulate volume separately")
		}
	})

	t.Run("vwma_division_formula", func(t *testing.T) {
		if !strings.Contains(vwmaCode, "/ volumeSum") {
			t.Error("VWMA must divide weighted sum by volume sum")
		}
	})
}

func TestVWMAHandler_NaNPropagation(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "vwma", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("nan_detection_in_window", func(t *testing.T) {
		if !strings.Contains(code, "math.IsNaN") {
			t.Error("Must check for NaN values in window")
		}
	})

	t.Run("nan_flag_tracking", func(t *testing.T) {
		if !strings.Contains(code, "hasNaN") {
			t.Error("Should track NaN presence across window")
		}
	})

	t.Run("nan_result_on_flag", func(t *testing.T) {
		if !strings.Contains(code, "if hasNaN") {
			t.Error("Should check hasNaN flag before returning result")
		}
		if !strings.Contains(code, "math.NaN()") {
			t.Error("Should return NaN when flag is set")
		}
	})

	t.Run("skip_nan_in_accumulation", func(t *testing.T) {
		hasNaNCheck := strings.Contains(code, "math.IsNaN")
		hasConditionalAccumulation := strings.Contains(code, "if") || strings.Contains(code, "else")

		if !hasNaNCheck || !hasConditionalAccumulation {
			t.Error("Should skip NaN values in accumulation loop")
		}
	})
}

func TestVWMAHandler_ZeroVolumeBehavior(t *testing.T) {
	handler := &VWMAHandler{}
	gen := createTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10},
		},
	}

	code, err := handler.GenerateCode(gen, "vwma", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("accumulates_volume", func(t *testing.T) {
		if !strings.Contains(code, "volumeSum") {
			t.Error("Must accumulate volume to handle zero-volume case")
		}
	})

	t.Run("division_by_volumesum", func(t *testing.T) {
		if !strings.Contains(code, "/ volumeSum") {
			t.Error("Division by volumeSum provides implicit zero-volume handling")
		}
	})
}
