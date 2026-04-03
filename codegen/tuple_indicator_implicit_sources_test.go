package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTupleIndicatorHandler_ImplicitSources(t *testing.T) {
	tests := []struct {
		name              string
		funcName          string
		args              []ast.Expression
		varNames          []string
		wantSeriesExtract bool
		wantArrayCount    int
		wantPeriodCount   int
	}{
		{
			name:              "dmi_with_two_periods",
			funcName:          "dmi",
			args:              []ast.Expression{&ast.Literal{Value: float64(14)}, &ast.Literal{Value: float64(13)}},
			varNames:          []string{"plusDI", "minusDI", "adx"},
			wantSeriesExtract: true,
			wantArrayCount:    3, // high, low, close
			wantPeriodCount:   2,
		},
		{
			name:              "ta_dmi_with_identifier_periods",
			funcName:          "ta.dmi",
			args:              []ast.Expression{&ast.Identifier{Name: "len"}, &ast.Identifier{Name: "lensig"}},
			varNames:          []string{"di_plus", "di_minus", "adx_value"},
			wantSeriesExtract: true,
			wantArrayCount:    3,
			wantPeriodCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			gen.constants = map[string]interface{}{
				"len":    14.0,
				"lensig": 13.0,
			}

			handler := NewTupleIndicatorHandler()
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			code, err := handler.GenerateTupleCode(gen, tt.varNames, call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() error = %v", err)
			}

			if code == "" {
				t.Fatal("GenerateTupleCode() returned empty code")
			}

			if tt.wantSeriesExtract {
				if !strings.Contains(code, "highWindow") {
					t.Error("code should extract highWindow")
				}
				if !strings.Contains(code, "lowWindow") {
					t.Error("code should extract lowWindow")
				}
				if !strings.Contains(code, "closeWindow") {
					t.Error("code should extract closeWindow")
				}
			}

			if strings.Contains(code, "ta.Dmi(") {
				expectedArgs := tt.wantArrayCount + tt.wantPeriodCount
				commaCount := strings.Count(code[strings.Index(code, "ta.Dmi("):], ",")
				if commaCount != expectedArgs-1 {
					t.Errorf("ta.Dmi call has %d commas (want %d for %d args)", commaCount, expectedArgs-1, expectedArgs)
				}
			}

			for _, varName := range tt.varNames {
				if !strings.Contains(code, varName+"Series.Set(") {
					t.Errorf("code should set %sSeries", varName)
				}
			}
		})
	}
}

func TestTupleIndicatorCodeGenerator_ImplicitArrayExtraction(t *testing.T) {
	tests := []struct {
		name      string
		sources   []string
		wantLoops int
	}{
		{
			name:      "three_sources_dmi",
			sources:   []string{"high", "low", "close"},
			wantLoops: 3,
		},
		{
			name:      "single_source",
			sources:   []string{"ohlc4"},
			wantLoops: 1,
		},
		{
			name:      "two_sources",
			sources:   []string{"high", "low"},
			wantLoops: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTupleIndicatorCodeGenerator()
			indenter := func() string { return "\t" }

			code := gen.generateImplicitArrayExtraction(tt.sources, indenter)

			for _, source := range tt.sources {
				windowVar := source + "Window"
				if !strings.Contains(code, windowVar) {
					t.Errorf("code should declare %s", windowVar)
				}
			}

			loopCount := strings.Count(code, "for j := 0")
			if loopCount != tt.wantLoops {
				t.Errorf("loop count = %d, want %d", loopCount, tt.wantLoops)
			}

			for _, source := range tt.sources {
				seriesVar := source + "Series"
				if !strings.Contains(code, seriesVar+".Get(") {
					t.Errorf("code should access %s.Get()", seriesVar)
				}
			}
		})
	}
}

func TestTupleIndicatorCodeGenerator_ImplicitSourcesRuntimeCall(t *testing.T) {
	tests := []struct {
		name           string
		spec           *TupleIndicatorSpec
		periods        []int
		wantArrayArgs  int
		wantPeriodArgs int
	}{
		{
			name: "dmi_three_arrays_two_periods",
			spec: &TupleIndicatorSpec{
				FunctionName:    "ta.dmi",
				RuntimeFunction: "ta.Dmi",
				OutputCount:     3,
				ImplicitSources: []string{"high", "low", "close"},
			},
			periods:        []int{14, 13},
			wantArrayArgs:  3,
			wantPeriodArgs: 2,
		},
		{
			name: "hypothetical_single_array",
			spec: &TupleIndicatorSpec{
				FunctionName:    "ta.example",
				RuntimeFunction: "ta.Example",
				OutputCount:     2,
				ImplicitSources: []string{"close"},
			},
			periods:        []int{20},
			wantArrayArgs:  1,
			wantPeriodArgs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTupleIndicatorCodeGenerator()
			indenter := func() string { return "\t" }

			params := &TupleIndicatorArguments{
				Periods: tt.periods,
			}

			code := gen.generateImplicitSourcesRuntimeCall(tt.spec, params, indenter)

			if !strings.Contains(code, tt.spec.RuntimeFunction+"(") {
				t.Errorf("code should call %s", tt.spec.RuntimeFunction)
			}

			for _, source := range tt.spec.ImplicitSources {
				windowVar := source + "Window"
				if !strings.Contains(code, windowVar) {
					t.Errorf("code should pass %s as argument", windowVar)
				}
			}

			for _, period := range tt.periods {
				periodStr := string(rune(period + '0'))
				if period < 10 && !strings.Contains(code, periodStr) {
					t.Errorf("code should pass period %d", period)
				}
			}

			commaCount := strings.Count(code[:strings.Index(code, ":=")], ",")
			if commaCount != tt.spec.OutputCount-1 {
				t.Errorf("output variable count = %d, want %d", commaCount+1, tt.spec.OutputCount)
			}
		})
	}
}

func TestTupleIndicatorArgumentExtractor_NumericConstantDetection(t *testing.T) {
	tests := []struct {
		name        string
		args        []ast.Expression
		constants   map[string]interface{}
		wantSource  bool
		wantPeriods []int
	}{
		{
			name: "both_literal_periods",
			args: []ast.Expression{
				&ast.Literal{Value: float64(14)},
				&ast.Literal{Value: float64(13)},
			},
			constants:   map[string]interface{}{},
			wantSource:  false,
			wantPeriods: []int{14, 13},
		},
		{
			name: "both_identifier_periods_in_constants",
			args: []ast.Expression{
				&ast.Identifier{Name: "diLen"},
				&ast.Identifier{Name: "adxSmooth"},
			},
			constants: map[string]interface{}{
				"diLen":     14.0,
				"adxSmooth": 13.0,
			},
			wantSource:  false,
			wantPeriods: []int{14, 13},
		},
		{
			name: "mixed_literal_and_constant",
			args: []ast.Expression{
				&ast.Literal{Value: float64(20)},
				&ast.Identifier{Name: "smoothing"},
			},
			constants: map[string]interface{}{
				"smoothing": 10.0,
			},
			wantSource:  false,
			wantPeriods: []int{20, 10},
		},
		{
			name: "source_then_period",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(10)},
			},
			constants:   map[string]interface{}{},
			wantSource:  true,
			wantPeriods: []int{10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewTupleIndicatorArgumentExtractor()
			call := &ast.CallExpression{Arguments: tt.args}

			sourceExtractor := func(expr ast.Expression) string {
				if id, ok := expr.(*ast.Identifier); ok {
					return id.Name
				}
				return ""
			}

			result, err := extractor.Extract(call, sourceExtractor, tt.constants)
			if err != nil {
				t.Fatalf("Extract() error = %v", err)
			}

			hasSource := result.SourceExpr != ""
			if hasSource != tt.wantSource {
				t.Errorf("has source = %v, want %v", hasSource, tt.wantSource)
			}

			if len(result.Periods) != len(tt.wantPeriods) {
				t.Errorf("period count = %d, want %d", len(result.Periods), len(tt.wantPeriods))
			}

			for i, wantPeriod := range tt.wantPeriods {
				if i < len(result.Periods) && result.Periods[i] != wantPeriod {
					t.Errorf("period[%d] = %d, want %d", i, result.Periods[i], wantPeriod)
				}
			}
		})
	}
}

func TestTupleIndicatorRegistry_ImplicitSourcesPropagation(t *testing.T) {
	registry := NewTupleIndicatorRegistry()

	spec := registry.Lookup("ta.dmi")
	if spec == nil {
		t.Fatal("ta.dmi not registered")
	}

	if len(spec.ImplicitSources) != 3 {
		t.Errorf("ta.dmi ImplicitSources length = %d, want 3", len(spec.ImplicitSources))
	}

	expectedSources := []string{"high", "low", "close"}
	for i, expected := range expectedSources {
		if i >= len(spec.ImplicitSources) || spec.ImplicitSources[i] != expected {
			t.Errorf("ta.dmi ImplicitSources[%d] = %q, want %q", i,
				func() string {
					if i < len(spec.ImplicitSources) {
						return spec.ImplicitSources[i]
					}
					return ""
				}(),
				expected)
		}
	}

	bareSpec := registry.Lookup("dmi")
	if bareSpec == nil {
		t.Fatal("dmi (bare) not registered")
	}

	if len(bareSpec.ImplicitSources) != len(spec.ImplicitSources) {
		t.Errorf("bare dmi ImplicitSources length = %d, want %d", len(bareSpec.ImplicitSources), len(spec.ImplicitSources))
	}

	for i := range spec.ImplicitSources {
		if i >= len(bareSpec.ImplicitSources) || bareSpec.ImplicitSources[i] != spec.ImplicitSources[i] {
			t.Errorf("bare dmi should inherit ImplicitSources from ta.dmi")
		}
	}
}

func TestTupleIndicatorCodeGenerator_WarmupPeriodCalculation(t *testing.T) {
	tests := []struct {
		name       string
		periods    []int
		wantWarmup int
	}{
		{
			name:       "dmi_14_13",
			periods:    []int{14, 13},
			wantWarmup: 13, // max(14, 13) - 1
		},
		{
			name:       "single_period_5",
			periods:    []int{5},
			wantWarmup: 4,
		},
		{
			name:       "three_periods",
			periods:    []int{10, 20, 15},
			wantWarmup: 19, // max(10, 20, 15) - 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewTupleIndicatorCodeGenerator()
			params := &TupleIndicatorArguments{Periods: tt.periods}

			warmup := gen.calculateWarmupPeriod(params)

			if warmup != tt.wantWarmup {
				t.Errorf("warmup period = %d, want %d", warmup, tt.wantWarmup)
			}
		})
	}
}
