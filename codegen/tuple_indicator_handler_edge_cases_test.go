package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestTupleIndicatorHandler_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		varNames      []string
		call          *ast.CallExpression
		shouldError   bool
		errorContains string
	}{
		{
			name:     "Zero outputs requested",
			varNames: []string{},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "macd"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "output count mismatch",
		},
		{
			name:     "Single output for three-output indicator",
			varNames: []string{"macdLine"},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "macd"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "output count mismatch",
		},
		{
			name:     "Excessive outputs requested",
			varNames: []string{"m", "s", "h", "extra"},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "macd"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "output count mismatch",
		},
		{
			name:     "Two outputs for three-output indicator",
			varNames: []string{"macd", "signal"},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "bb"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "output count mismatch",
		},
		{
			name:     "Unregistered indicator function",
			varNames: []string{"out1", "out2"},
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "nonexistent"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "not registered as tuple indicator",
		},
		{
			name:     "Non-member-expression callee",
			varNames: []string{"out1", "out2"},
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "standalone"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			shouldError:   true,
			errorContains: "not registered as tuple indicator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTupleIndicatorHandler()
			g := newTestGenerator()
			g.indent = 1

			_, err := handler.GenerateTupleCode(g, tt.varNames, tt.call)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTupleIndicatorHandler_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name            string
		indicatorName   string
		varNames        []string
		mustContainCode []string
	}{
		{
			name:          "Stoch with 2 outputs",
			indicatorName: "stoch",
			varNames:      []string{"k", "d"},
			mustContainCode: []string{
				"ta.Stoch(",
				"kSeries.Set(",
				"dSeries.Set(",
			},
		},
		{
			name:          "MACD with 3 outputs",
			indicatorName: "macd",
			varNames:      []string{"macdLine", "signalLine", "histogram"},
			mustContainCode: []string{
				"ta.Macd(",
				"macdLineSeries.Set(",
				"signalLineSeries.Set(",
				"histogramSeries.Set(",
			},
		},
		{
			name:          "BB with 3 outputs",
			indicatorName: "bb",
			varNames:      []string{"middle", "upper", "lower"},
			mustContainCode: []string{
				"ta.BBands(",
				"middleSeries.Set(",
				"upperSeries.Set(",
				"lowerSeries.Set(",
			},
		},
		{
			name:          "Stoch with different variable names",
			indicatorName: "stoch",
			varNames:      []string{"slowK", "slowD"},
			mustContainCode: []string{
				"ta.Stoch(",
				"slowKSeries.Set(",
				"slowDSeries.Set(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTupleIndicatorHandler()
			g := newTestGenerator()
			g.indent = 1

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.indicatorName},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			}

			code, err := handler.GenerateTupleCode(g, tt.varNames, call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() error: %v", err)
			}

			for _, pattern := range tt.mustContainCode {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern %q in generated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestTupleIndicatorHandler_PeriodArgumentVariations(t *testing.T) {
	tests := []struct {
		name               string
		arguments          []ast.Expression
		mustContainPattern string
		description        string
	}{
		{
			name: "MACD with literal periods",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(12)},
				&ast.Literal{Value: float64(26)},
				&ast.Literal{Value: float64(9)},
			},
			mustContainPattern: "ta.Macd(",
			description:        "Standard MACD with literal period values",
		},
		{
			name: "MACD with identifier periods",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "fastPeriod"},
				&ast.Identifier{Name: "slowPeriod"},
				&ast.Identifier{Name: "signalPeriod"},
			},
			mustContainPattern: "ta.Macd(",
			description:        "MACD with variable period parameters",
		},
		{
			name: "BB with literal period and multiplier",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: float64(20)},
				&ast.Literal{Value: float64(2.0)},
			},
			mustContainPattern: "ta.BBands(",
			description:        "Bollinger Bands with literal parameters",
		},
		{
			name: "Stoch with multiple periods",
			arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "high"},
				&ast.Identifier{Name: "low"},
				&ast.Literal{Value: float64(14)},
				&ast.Literal{Value: float64(3)},
				&ast.Literal{Value: float64(3)},
			},
			mustContainPattern: "ta.Stoch(",
			description:        "Stochastic with multiple period parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewTupleIndicatorHandler()
			g := newTestGenerator()
			g.indent = 1

			var indicatorName string
			var varNames []string

			if strings.Contains(tt.name, "MACD") {
				indicatorName = "macd"
				varNames = []string{"macd", "signal", "hist"}
			} else if strings.Contains(tt.name, "BB") {
				indicatorName = "bb"
				varNames = []string{"middle", "upper", "lower"}
			} else if strings.Contains(tt.name, "Stoch") {
				indicatorName = "stoch"
				varNames = []string{"k", "d"}
			}

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: indicatorName},
				},
				Arguments: tt.arguments,
			}

			code, err := handler.GenerateTupleCode(g, varNames, call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() error: %v", err)
			}

			if !strings.Contains(code, tt.mustContainPattern) {
				t.Errorf("Missing pattern %q in generated code:\n%s", tt.mustContainPattern, code)
			}
		})
	}
}

func TestTupleIndicatorHandler_MultipleIndicatorsConcurrent(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	call1 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "macd"},
		},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}
	varNames1 := []string{"macd1", "signal1", "hist1"}

	call2 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "stoch"},
		},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}
	varNames2 := []string{"k", "d"}

	code1, err1 := handler.GenerateTupleCode(g, varNames1, call1)
	if err1 != nil {
		t.Fatalf("First indicator generation failed: %v", err1)
	}

	code2, err2 := handler.GenerateTupleCode(g, varNames2, call2)
	if err2 != nil {
		t.Fatalf("Second indicator generation failed: %v", err2)
	}

	if !strings.Contains(code1, "macd1Series.Set(") {
		t.Error("First indicator code missing macd1 series storage")
	}

	if !strings.Contains(code2, "kSeries.Set(") {
		t.Error("Second indicator code missing k series storage")
	}

	if strings.Contains(code1, "kSeries") || strings.Contains(code1, "dSeries") {
		t.Error("First indicator code contaminated with second indicator variables")
	}

	if strings.Contains(code2, "macd1Series") || strings.Contains(code2, "signal1Series") {
		t.Error("Second indicator code contaminated with first indicator variables")
	}
}

func TestTupleIndicatorHandler_CodeStructureValidation(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "macd"},
		},
		Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
	}
	varNames := []string{"macd", "signal", "hist"}

	code, err := handler.GenerateTupleCode(g, varNames, call)
	if err != nil {
		t.Fatalf("GenerateTupleCode() error: %v", err)
	}

	requiredStructure := []string{
		"sourceWindow := make([]float64",
		"for j := 0; j <",
		"sourceWindow[j] =",
		"ta.Macd(",
		"macdSeries.Set(",
		"signalSeries.Set(",
		"histSeries.Set(",
	}

	for _, pattern := range requiredStructure {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing required code structure: %q\nGenerated:\n%s", pattern, code)
		}
	}

	if strings.Count(code, "sourceWindow := make(") != 1 {
		t.Error("Window extraction should occur exactly once")
	}

	if strings.Count(code, "ta.Macd(") != 1 {
		t.Error("Runtime call should occur exactly once")
	}
}

func TestTupleIndicatorHandler_VariableNamingCollision(t *testing.T) {
	handler := NewTupleIndicatorHandler()
	g := newTestGenerator()
	g.indent = 1

	tests := []struct {
		name     string
		varNames []string
	}{
		{
			name:     "Reserved keywords",
			varNames: []string{"if", "for", "var"},
		},
		{
			name:     "Similar names",
			varNames: []string{"macd", "macd1", "macd2"},
		},
		{
			name:     "Underscores",
			varNames: []string{"my_macd", "my_signal", "my_hist"},
		},
		{
			name:     "Camel case",
			varNames: []string{"fastMACD", "slowSignal", "histValue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "macd"},
				},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			}

			code, err := handler.GenerateTupleCode(g, tt.varNames, call)
			if err != nil {
				t.Fatalf("GenerateTupleCode() failed for %s: %v", tt.name, err)
			}

			for _, varName := range tt.varNames {
				if !strings.Contains(code, varName+"Series") {
					t.Errorf("Expected variable %sSeries in generated code", varName)
				}
			}
		})
	}
}
