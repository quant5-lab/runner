package codegen

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

type mockSeriesExtractor struct {
	extractFunc  func(ast.Expression) string
	callCount    int
	lastExtrExpr ast.Expression
}

func (m *mockSeriesExtractor) ExtractSeriesExpression(expr ast.Expression) string {
	m.callCount++
	m.lastExtrExpr = expr
	if m.extractFunc != nil {
		return m.extractFunc(expr)
	}
	if lit, ok := expr.(*ast.Literal); ok {
		return fmt.Sprintf("%v", lit.Value)
	}
	if id, ok := expr.(*ast.Identifier); ok {
		return id.Name + "Series.GetCurrent()"
	}
	return "unknownExpr"
}

func newMockExtractor() *mockSeriesExtractor {
	return &mockSeriesExtractor{}
}

func TestMathFunctionRegistry_Lookup(t *testing.T) {
	registry := NewMathFunctionRegistry()

	tests := []struct {
		name         string
		funcName     string
		shouldFind   bool
		expectedPine string
		expectedGo   string
		minArgs      int
		maxArgs      int
	}{
		{"prefixed abs", "math.abs", true, "math.abs", "math.Abs", 1, 1},
		{"unprefixed abs", "abs", true, "math.abs", "math.Abs", 1, 1},
		{"prefixed sin", "math.sin", true, "math.sin", "math.Sin", 1, 1},
		{"unprefixed cos", "cos", true, "math.cos", "math.Cos", 1, 1},
		{"prefixed max", "math.max", true, "math.max", "math.Max", 2, 2},
		{"unprefixed min", "min", true, "math.min", "math.Min", 2, 2},
		{"prefixed pow", "math.pow", true, "math.pow", "math.Pow", 2, 2},
		{"unprefixed sqrt", "sqrt", true, "math.sqrt", "math.Sqrt", 1, 1},
		{"sign special", "math.sign", true, "math.sign", "", 1, 1},
		{"todegrees", "math.todegrees", true, "math.todegrees", "", 1, 1},
		{"toradians", "math.toradians", true, "math.toradians", "", 1, 1},
		{"avg variadic", "math.avg", true, "math.avg", "", 1, -1},
		{"random", "math.random", true, "math.random", "", 0, 3},
		{"round_to_mintick", "math.round_to_mintick", true, "math.round_to_mintick", "", 1, 1},
		{"unsupported", "math.unknown", false, "", "", 0, 0},
		{"ta function", "ta.sma", false, "", "", 0, 0},
		{"empty", "", false, "", "", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, found := registry.Lookup(tt.funcName)
			if found != tt.shouldFind {
				t.Errorf("Lookup(%q): found=%v, want %v", tt.funcName, found, tt.shouldFind)
			}
			if found {
				if spec.PineName != tt.expectedPine {
					t.Errorf("PineName: got %q, want %q", spec.PineName, tt.expectedPine)
				}
				if spec.GoFunc != tt.expectedGo {
					t.Errorf("GoFunc: got %q, want %q", spec.GoFunc, tt.expectedGo)
				}
				if spec.MinArgs != tt.minArgs {
					t.Errorf("MinArgs: got %d, want %d", spec.MinArgs, tt.minArgs)
				}
				if spec.MaxArgs != tt.maxArgs {
					t.Errorf("MaxArgs: got %d, want %d", spec.MaxArgs, tt.maxArgs)
				}
			}
		})
	}
}

func TestMathExpressionGenerator_CanHandle(t *testing.T) {
	tests := []struct {
		funcName string
		want     bool
	}{
		{"math.sin", true},
		{"math.abs", true},
		{"abs", true},
		{"max", true},
		{"math.sign", true},
		{"sign", true},
		{"math.todegrees", true},
		{"math.avg", true},
		{"math.random", true},
		{"math.round_to_mintick", true},
		{"math.unknown", false},
		{"ta.sma", false},
		{"", false},
	}

	extractor := newMockExtractor()
	gen := NewMathExpressionGenerator(extractor)

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			result := gen.CanHandle(tt.funcName)
			if result != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, result, tt.want)
			}
		})
	}
}

func TestMathExpressionGenerator_UnaryFunctions(t *testing.T) {
	unaryFunctions := []struct {
		pineName string
		goFunc   string
	}{
		{"math.abs", "math.Abs"},
		{"math.acos", "math.Acos"},
		{"math.asin", "math.Asin"},
		{"math.atan", "math.Atan"},
		{"math.ceil", "math.Ceil"},
		{"math.cos", "math.Cos"},
		{"math.exp", "math.Exp"},
		{"math.floor", "math.Floor"},
		{"math.log", "math.Log"},
		{"math.log10", "math.Log10"},
		{"math.round", "math.Round"},
		{"math.sin", "math.Sin"},
		{"math.sqrt", "math.Sqrt"},
		{"math.tan", "math.Tan"},
	}

	for _, fn := range unaryFunctions {
		t.Run(fn.pineName, func(t *testing.T) {
			extractor := newMockExtractor()
			gen := NewMathExpressionGenerator(extractor)

			arg := &ast.Identifier{Name: "x"}
			result, err := gen.GenerateExpression(fn.pineName, []ast.Expression{arg})
			if err != nil {
				t.Fatalf("GenerateExpression failed: %v", err)
			}

			expectedPrefix := fn.goFunc + "("
			if !strings.HasPrefix(result, expectedPrefix) {
				t.Errorf("Expected prefix %q, got %q", expectedPrefix, result)
			}

			if !strings.Contains(result, "xSeries.GetCurrent()") {
				t.Errorf("Expected series access, got %q", result)
			}

			if extractor.callCount != 1 {
				t.Errorf("Expected 1 extractor call, got %d", extractor.callCount)
			}
		})
	}
}

func TestMathExpressionGenerator_BinaryFunctions(t *testing.T) {
	binaryFunctions := []struct {
		pineName string
		goFunc   string
	}{
		{"math.max", "math.Max"},
		{"math.min", "math.Min"},
		{"math.pow", "math.Pow"},
	}

	for _, fn := range binaryFunctions {
		t.Run(fn.pineName, func(t *testing.T) {
			extractor := newMockExtractor()
			gen := NewMathExpressionGenerator(extractor)

			args := []ast.Expression{
				&ast.Identifier{Name: "x"},
				&ast.Identifier{Name: "y"},
			}
			result, err := gen.GenerateExpression(fn.pineName, args)
			if err != nil {
				t.Fatalf("GenerateExpression failed: %v", err)
			}

			expectedPrefix := fn.goFunc + "("
			if !strings.HasPrefix(result, expectedPrefix) {
				t.Errorf("Expected prefix %q, got %q", expectedPrefix, result)
			}

			if !strings.Contains(result, "xSeries.GetCurrent()") {
				t.Errorf("Expected x series access, got %q", result)
			}
			if !strings.Contains(result, "ySeries.GetCurrent()") {
				t.Errorf("Expected y series access, got %q", result)
			}

			if extractor.callCount != 2 {
				t.Errorf("Expected 2 extractor calls, got %d", extractor.callCount)
			}
		})
	}
}

func TestMathExpressionGenerator_SpecialFunctions(t *testing.T) {
	tests := []struct {
		name            string
		funcName        string
		args            []ast.Expression
		expectedContent []string
	}{
		{
			name:     "sign",
			funcName: "math.sign",
			args:     []ast.Expression{&ast.Identifier{Name: "x"}},
			expectedContent: []string{
				"xSeries.GetCurrent()",
				"if",
				"> 0",
				"return 1",
				"return -1",
				"return 0",
			},
		},
		{
			name:     "todegrees",
			funcName: "math.todegrees",
			args:     []ast.Expression{&ast.Literal{Value: math.Pi}},
			expectedContent: []string{
				"* 180",
				"/ math.Pi",
			},
		},
		{
			name:     "toradians",
			funcName: "math.toradians",
			args:     []ast.Expression{&ast.Literal{Value: 180.0}},
			expectedContent: []string{
				"* math.Pi",
				"/ 180",
			},
		},
		{
			name:     "avg with 3 args",
			funcName: "math.avg",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
				&ast.Identifier{Name: "c"},
			},
			expectedContent: []string{
				"aSeries.GetCurrent()",
				"bSeries.GetCurrent()",
				"cSeries.GetCurrent()",
				"/ 3",
			},
		},
		{
			name:            "random no args",
			funcName:        "math.random",
			args:            []ast.Expression{},
			expectedContent: []string{"rand.Float64()"},
		},
		{
			name:     "random with min/max",
			funcName: "math.random",
			args: []ast.Expression{
				&ast.Literal{Value: 0.0},
				&ast.Literal{Value: 100.0},
			},
			expectedContent: []string{
				"rand.Float64()",
				"*",
			},
		},
		{
			name:     "round_to_mintick",
			funcName: "math.round_to_mintick",
			args:     []ast.Expression{&ast.Identifier{Name: "price"}},
			expectedContent: []string{
				"priceSeries.GetCurrent()",
				"ctx.Mintick",
				"math.Round",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := newMockExtractor()
			gen := NewMathExpressionGenerator(extractor)

			result, err := gen.GenerateExpression(tt.funcName, tt.args)
			if err != nil {
				t.Fatalf("GenerateExpression failed: %v", err)
			}

			for _, content := range tt.expectedContent {
				if !strings.Contains(result, content) {
					t.Errorf("Expected result to contain %q, got: %s", content, result)
				}
			}
		})
	}
}

func TestMathExpressionGenerator_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name      string
		funcName  string
		argCount  int
		wantError bool
	}{
		{"abs valid 1 arg", "math.abs", 1, false},
		{"abs invalid 0 args", "math.abs", 0, true},
		{"abs invalid 2 args", "math.abs", 2, true},
		{"max valid 2 args", "math.max", 2, false},
		{"max invalid 1 arg", "math.max", 1, true},
		{"avg valid 1 arg", "math.avg", 1, false},
		{"avg valid 10 args", "math.avg", 10, false},
		{"random valid 0 args", "math.random", 0, false},
		{"random valid 2 args", "math.random", 2, false},
		{"random invalid 4 args", "math.random", 4, true},
		{"sin invalid 0 args", "math.sin", 0, true},
		{"sin invalid 2 args", "math.sin", 2, true},
		{"pow invalid 1 arg", "math.pow", 1, true},
		{"pow invalid 3 args", "math.pow", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := newMockExtractor()
			gen := NewMathExpressionGenerator(extractor)

			args := make([]ast.Expression, tt.argCount)
			for i := 0; i < tt.argCount; i++ {
				args[i] = &ast.Literal{Value: float64(i)}
			}

			_, err := gen.GenerateExpression(tt.funcName, args)
			if (err != nil) != tt.wantError {
				t.Errorf("wantError=%v, got err=%v", tt.wantError, err)
			}
		})
	}
}

func TestMathExpressionGenerator_EdgeCases(t *testing.T) {
	t.Run("empty function name", func(t *testing.T) {
		extractor := newMockExtractor()
		gen := NewMathExpressionGenerator(extractor)

		_, err := gen.GenerateExpression("", []ast.Expression{})
		if err == nil {
			t.Error("Expected error for empty function name")
		}
	})

	t.Run("unsupported function", func(t *testing.T) {
		extractor := newMockExtractor()
		gen := NewMathExpressionGenerator(extractor)

		_, err := gen.GenerateExpression("math.nonexistent", []ast.Expression{&ast.Literal{Value: 1.0}})
		if err == nil {
			t.Error("Expected error for unsupported function")
		}
	})

	t.Run("NaN argument", func(t *testing.T) {
		extractor := newMockExtractor()
		gen := NewMathExpressionGenerator(extractor)

		args := []ast.Expression{&ast.Literal{Value: math.NaN()}}
		result, err := gen.GenerateExpression("math.abs", args)
		if err != nil {
			t.Errorf("Should handle NaN gracefully, got error: %v", err)
		}
		if !strings.Contains(result, "NaN") {
			t.Errorf("Expected NaN in result, got: %s", result)
		}
	})

	t.Run("Inf argument", func(t *testing.T) {
		extractor := newMockExtractor()
		gen := NewMathExpressionGenerator(extractor)

		args := []ast.Expression{&ast.Literal{Value: math.Inf(1)}}
		_, err := gen.GenerateExpression("math.abs", args)
		if err != nil {
			t.Errorf("Should handle Inf gracefully, got error: %v", err)
		}
	})

	t.Run("unprefixed function name normalization", func(t *testing.T) {
		extractor := newMockExtractor()
		gen := NewMathExpressionGenerator(extractor)

		arg := &ast.Identifier{Name: "x"}
		resultPrefixed, err1 := gen.GenerateExpression("math.abs", []ast.Expression{arg})
		resultUnprefixed, err2 := gen.GenerateExpression("abs", []ast.Expression{arg})

		if err1 != nil || err2 != nil {
			t.Fatalf("Both prefixed and unprefixed should work: err1=%v, err2=%v", err1, err2)
		}

		if resultPrefixed != resultUnprefixed {
			t.Errorf("Prefixed and unprefixed should generate same code:\nprefixed:   %q\nunprefixed: %q", resultPrefixed, resultUnprefixed)
		}
	})
}

func TestMathExpressionGenerator_Integration_WithSeriesAccessConverter(t *testing.T) {
	st := NewSymbolTable()
	st.Register("close", VariableTypeSeries)
	st.Register("open", VariableTypeSeries)

	conv := NewSeriesAccessConverter(st, "0", nil)

	tests := []struct {
		name            string
		expr            ast.Expression
		expectedContent []string
	}{
		{
			name: "abs with series identifier",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "math.abs"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			expectedContent: []string{"math.Abs", "close"},
		},
		{
			name: "max with two series",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "math.max"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "open"},
				},
			},
			expectedContent: []string{"math.Max", "close", "open"},
		},
		{
			name: "unprefixed abs",
			expr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "abs"},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}},
			},
			expectedContent: []string{"math.Abs", "close"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.ConvertExpression(tt.expr)
			if err != nil {
				t.Fatalf("ConvertExpression failed: %v", err)
			}

			for _, content := range tt.expectedContent {
				if !strings.Contains(result, content) {
					t.Errorf("Expected result to contain %q, got: %s", content, result)
				}
			}
		})
	}
}
