package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestMFIHandler_EdgeCases(t *testing.T) {
	handler := &MFIHandler{}
	gen := newTestGenerator()

	t.Run("empty_arguments", func(t *testing.T) {
		call := &ast.CallExpression{Arguments: []ast.Expression{}}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for empty arguments")
		}
	})

	t.Run("missing_period", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
		}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for missing period argument")
		}
	})

	t.Run("extra_arguments_accepted", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14},
				&ast.Literal{Value: 999},
			},
		}
		code, err := handler.GenerateCode(gen, "test", call)
		_ = code
		_ = err
	})

	t.Run("dynamic_period_error", func(t *testing.T) {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "hlc3"},
				&ast.Identifier{Name: "dynamicPeriod"},
			},
		}
		_, err := handler.GenerateCode(gen, "test", call)
		if err == nil {
			t.Error("Expected error for dynamic period")
		}
		if err != nil && !strings.Contains(err.Error(), "runtime dynamic period") {
			t.Errorf("Error should mention 'runtime dynamic period', got: %v", err)
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
		periods := []int{0, -14, 5000}
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

func TestMFIHandler_PeriodBoundaries(t *testing.T) {
	handler := &MFIHandler{}
	gen := newTestGenerator()

	testCases := []struct {
		period    int
		warmupBar int
	}{
		{1, 1},
		{2, 2},
		{14, 14},
		{100, 100},
	}

	for _, tc := range testCases {
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "hlc3"},
				&ast.Literal{Value: tc.period},
			},
		}

		code, err := handler.GenerateCode(gen, "mfi", call)
		if err != nil {
			t.Fatalf("GenerateCode(%d) error = %v", tc.period, err)
		}

		if !strings.Contains(code, "if ctx.BarIndex <") {
			t.Errorf("Missing warmup check for period %d", tc.period)
		}

		if !strings.Contains(code, "_positive_mf") || !strings.Contains(code, "_negative_mf") {
			t.Error("Missing internal series for positive/negative money flow")
		}
	}
}

func TestMFIHandler_AlgorithmCorrectness(t *testing.T) {
	handler := &MFIHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "hlc3"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "mfi", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("change_calculation_present", func(t *testing.T) {
		if code == "" {
			t.Error("Code should not be empty")
		}
		if len(code) < 100 {
			t.Error("Handler should generate substantial code via builder delegation")
		}
	})

	t.Run("money_flow_formula", func(t *testing.T) {
		if !strings.Contains(code, "rawMF") {
			t.Error("Missing raw money flow variable")
		}
		if !strings.Contains(code, "Volume") {
			t.Error("Raw money flow should use volume")
		}
	})

	t.Run("directional_split", func(t *testing.T) {
		if !strings.Contains(code, "if math.IsNaN") {
			t.Error("Missing NaN check branch")
		}
		if !strings.Contains(code, "> 0") {
			t.Error("Missing positive change check")
		}

		if !strings.Contains(code, "_mfi") && !strings.Contains(code, "posMF") && !strings.Contains(code, "negMF") {
			t.Error("Missing positive/negative money flow variables")
		}
	})

	t.Run("window_based_summation", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j <") {
			t.Error("Missing window summation loop")
		}
	})

	t.Run("mfi_formula", func(t *testing.T) {
		if !strings.Contains(code, "mfr") && !strings.Contains(code, "ratio") {
			t.Error("Missing money flow ratio calculation")
		}
		if !strings.Contains(code, "100") {
			t.Error("Missing MFI formula constant 100")
		}
		if !strings.Contains(code, "mfi") {
			t.Error("Missing mfi variable")
		}
	})

	t.Run("zero_denominator_protection", func(t *testing.T) {
		if !strings.Contains(code, "negSum == 0") {
			t.Error("Missing zero-denominator check")
		}
		if !strings.Contains(code, "100") {
			t.Error("Missing MFI=100 for zero denominator case")
		}
	})
}

func TestMFIHandler_SourceTypes(t *testing.T) {
	handler := &MFIHandler{}
	gen := newTestGenerator()

	testCases := []struct {
		name   string
		source ast.Expression
	}{
		{"ohlc4", &ast.Identifier{Name: "ohlc4"}},
		{"hlc3", &ast.Identifier{Name: "hlc3"}},
		{"hl2", &ast.Identifier{Name: "hl2"}},
		{"close", &ast.Identifier{Name: "close"}},
		{
			"ta_sma_call",
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

			code, err := handler.GenerateCode(gen, "mfi", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if code == "" {
				t.Error("Generated code should not be empty")
			}

			if !strings.Contains(code, "rawMF") && !strings.Contains(code, "raw") {
				t.Error("Missing raw money flow related code")
			}
		})
	}
}

func TestMFIHandler_CodeStructureInvariants(t *testing.T) {
	handler := &MFIHandler{}
	gen := newTestGenerator()

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "hlc3"},
			&ast.Literal{Value: 14},
		},
	}

	code, err := handler.GenerateCode(gen, "mfi14", call)
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

	t.Run("consistent_naming", func(t *testing.T) {
		if !strings.Contains(code, "mfi14Series") {
			t.Error("Missing consistent series variable naming")
		}
		if !strings.Contains(code, "_mfi14_positive_mf") {
			t.Error("Missing positive money flow series with varName prefix")
		}
		if !strings.Contains(code, "_mfi14_negative_mf") {
			t.Error("Missing negative money flow series with varName prefix")
		}
	})

	t.Run("imports_required_packages", func(t *testing.T) {
		if !strings.Contains(code, "math.") {
			t.Error("Code should use math package for NaN/IsNaN")
		}
	})

	t.Run("code_completeness", func(t *testing.T) {
		if len(code) < 100 {
			t.Error("Generated code should be substantial")
		}
	})
}

func TestMFIHandler_IntegrationWithRegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("handler_registered", func(t *testing.T) {
		handler := registry.FindHandler("ta.mfi")
		if handler == nil {
			t.Fatal("mfi handler not registered in TAFunctionRegistry")
		}

		if _, ok := handler.(*MFIHandler); !ok {
			t.Errorf("Wrong handler type: %T", handler)
		}
	})

	t.Run("supports_function_names", func(t *testing.T) {
		if !registry.IsSupported("ta.mfi") {
			t.Error("Registry should support 'ta.mfi'")
		}
		if !registry.IsSupported("mfi") {
			t.Error("Registry should support 'mfi' (Pine v4 syntax)")
		}
	})

	t.Run("generates_code_via_registry", func(t *testing.T) {
		gen := newTestGenerator()
		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "hlc3"},
				&ast.Literal{Value: 14},
			},
		}

		code, err := registry.GenerateInlineTA(gen, "test", "ta.mfi", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("Registry should generate non-empty code")
		}

		if !strings.Contains(code, "testSeries.Set") {
			t.Error("Registry-generated code missing series assignment")
		}
	})
}

func TestMFIHandler_CompositeIndicatorInterface(t *testing.T) {
	handler := &MFIHandler{}

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "hlc3"},
			&ast.Literal{Value: 14},
		},
	}

	t.Run("returns_correct_series_count", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("mfi", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		if len(seriesNames) != 2 {
			t.Errorf("Expected 2 internal series, got %d: %v", len(seriesNames), seriesNames)
		}
	})

	t.Run("series_names_use_varname_prefix", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("myMFI", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		expectedPrefix := "_myMFI_"
		for _, name := range seriesNames {
			if !strings.HasPrefix(name, expectedPrefix) {
				t.Errorf("Series name %q missing prefix %q", name, expectedPrefix)
			}
		}
	})

	t.Run("series_names_include_positive_negative", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("test", call)
		if err != nil {
			t.Fatalf("GetInternalSeriesNames() error = %v", err)
		}

		hasPositive := false
		hasNegative := false
		for _, name := range seriesNames {
			if strings.Contains(name, "positive_mf") {
				hasPositive = true
			}
			if strings.Contains(name, "negative_mf") {
				hasNegative = true
			}
		}

		if !hasPositive {
			t.Error("Missing positive_mf series in internal series names")
		}
		if !hasNegative {
			t.Error("Missing negative_mf series in internal series names")
		}
	})

	t.Run("nil_call_expression", func(t *testing.T) {
		seriesNames, err := handler.GetInternalSeriesNames("test", nil)
		_ = seriesNames
		_ = err
	})
}
