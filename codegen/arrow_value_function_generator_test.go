package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Validates value function generation across all expression contexts in arrow functions */

func TestArrowValueFunctionGenerator_CanHandle(t *testing.T) {
	gen := NewArrowValueFunctionGenerator(nil)

	tests := []struct {
		funcName string
		want     bool
	}{
		{"nz", true},
		{"na", true},
		{"fixnan", true},
		{"ta.fixnan", false}, // Only unprefixed in value context
		{"sma", false},
		{"ta.ema", false},
		{"max", false},
		{"abs", false},
		{"plot", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			got := gen.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Nz_HistoricalAccess(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name     string
		varName  string
		offset   float64
		contains []string
	}{
		{
			name:     "historical access offset 1",
			varName:  "value",
			offset:   1.0,
			contains: []string{"value.Nz(", "valueSeries.Get(int(1))", ", 0)"},
		},
		{
			name:     "historical access offset 5",
			varName:  "count",
			offset:   5.0,
			contains: []string{"value.Nz(", "countSeries.Get(int(5))", ", 0)"},
		},
		{
			name:     "current access offset 0",
			varName:  "price",
			offset:   0.0,
			contains: []string{"value.Nz(", "priceSeries.Get(int(0))", ", 0)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: tt.varName},
						Property: &ast.Literal{Value: tt.offset},
						Computed: true,
					},
				},
			}

			code, err := valueGen.Generate(call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			for _, substr := range tt.contains {
				if !strings.Contains(code, substr) {
					t.Errorf("Generate() missing %q in: %s", substr, code)
				}
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Nz_ParameterAccess(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	resolver.RegisterParameter("threshold")
	resolver.RegisterParameter("fallback")
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name         string
		arg          ast.Expression
		replacement  ast.Expression
		expectedArg  string
		expectedRepl string
	}{
		{
			name:         "parameter without replacement",
			arg:          &ast.Identifier{Name: "threshold"},
			replacement:  nil,
			expectedArg:  "threshold",
			expectedRepl: "0",
		},
		{
			name:         "parameter with literal replacement",
			arg:          &ast.Identifier{Name: "threshold"},
			replacement:  &ast.Literal{Value: float64(-1)},
			expectedArg:  "threshold",
			expectedRepl: "-1",
		},
		{
			name:         "parameter with parameter replacement",
			arg:          &ast.Identifier{Name: "threshold"},
			replacement:  &ast.Identifier{Name: "fallback"},
			expectedArg:  "threshold",
			expectedRepl: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []ast.Expression{tt.arg}
			if tt.replacement != nil {
				args = append(args, tt.replacement)
			}

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "nz"},
				Arguments: args,
			}

			code, err := valueGen.Generate(call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			expected := "value.Nz(" + tt.expectedArg + ", " + tt.expectedRepl + ")"
			if code != expected {
				t.Errorf("Generate() = %q, want %q", code, expected)
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Nz_BuiltinSeries(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name     string
		builtin  string
		expected string
	}{
		{"close", "close", "value.Nz(ctx.Data[ctx.BarIndex].Close, 0)"},
		{"open", "open", "value.Nz(ctx.Data[ctx.BarIndex].Open, 0)"},
		{"high", "high", "value.Nz(ctx.Data[ctx.BarIndex].High, 0)"},
		{"low", "low", "value.Nz(ctx.Data[ctx.BarIndex].Low, 0)"},
		{"volume", "volume", "value.Nz(ctx.Data[ctx.BarIndex].Volume, 0)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{&ast.Identifier{Name: tt.builtin}},
			}

			code, err := valueGen.Generate(call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			if code != tt.expected {
				t.Errorf("Generate() = %q, want %q", code, tt.expected)
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Nz_EdgeCases(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected string
	}{
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{},
			},
			expected: "0",
		},
		{
			name: "literal source",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{&ast.Literal{Value: float64(100)}},
			},
			expected: "value.Nz(100, 0)",
		},
		{
			name: "literal source and replacement",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "nz"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: float64(100)},
					&ast.Literal{Value: float64(-999)},
				},
			},
			expected: "value.Nz(100, -999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := valueGen.Generate(tt.call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			if code != tt.expected {
				t.Errorf("Generate() = %q, want %q", code, tt.expected)
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Na_AllContexts(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	resolver.RegisterParameter("param1")
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name     string
		arg      ast.Expression
		expected string
	}{
		{
			name:     "no argument",
			arg:      nil,
			expected: "true",
		},
		{
			name:     "local variable",
			arg:      &ast.Identifier{Name: "x"},
			expected: "math.IsNaN(x)",
		},
		{
			name:     "parameter",
			arg:      &ast.Identifier{Name: "param1"},
			expected: "math.IsNaN(param1)",
		},
		{
			name:     "builtin series",
			arg:      &ast.Identifier{Name: "close"},
			expected: "math.IsNaN(ctx.Data[ctx.BarIndex].Close)",
		},
		{
			name:     "literal",
			arg:      &ast.Literal{Value: float64(50)},
			expected: "math.IsNaN(50)",
		},
		{
			name: "historical access",
			arg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "value"},
				Property: &ast.Literal{Value: float64(1)},
				Computed: true,
			},
			expected: "math.IsNaN(valueSeries.Get(int(1)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var args []ast.Expression
			if tt.arg != nil {
				args = []ast.Expression{tt.arg}
			}

			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "na"},
				Arguments: args,
			}

			code, err := valueGen.Generate(call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			if code != tt.expected {
				t.Errorf("Generate() = %q, want %q", code, tt.expected)
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Fixnan_AllContexts(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	resolver.RegisterParameter("src")
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name             string
		arg              ast.Expression
		requiredPatterns []string
	}{
		{
			name: "parameter",
			arg:  &ast.Identifier{Name: "src"},
			requiredPatterns: []string{
				"func()",
				"float64",
				"val := src",
				"math.IsNaN(val)",
				"return 0.0",
				"return val",
			},
		},
		{
			name: "local variable",
			arg:  &ast.Identifier{Name: "x"},
			requiredPatterns: []string{
				"func()",
				"val := x",
				"math.IsNaN(val)",
			},
		},
		{
			name: "builtin series",
			arg:  &ast.Identifier{Name: "close"},
			requiredPatterns: []string{
				"val := ctx.Data[ctx.BarIndex].Close",
			},
		},
		{
			name: "historical access",
			arg: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "price"},
				Property: &ast.Literal{Value: float64(2)},
				Computed: true,
			},
			requiredPatterns: []string{
				"val := priceSeries.Get(int(2))",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "fixnan"},
				Arguments: []ast.Expression{tt.arg},
			}

			code, err := valueGen.Generate(call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			for _, pattern := range tt.requiredPatterns {
				if !strings.Contains(code, pattern) {
					t.Errorf("Generate() missing %q in: %s", pattern, code)
				}
			}
		})
	}
}

func TestArrowValueFunctionGenerator_Fixnan_NoArgument(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "fixnan"},
		Arguments: []ast.Expression{},
	}

	_, err := valueGen.Generate(call)
	if err == nil {
		t.Error("Generate() expected error for fixnan() with no arguments, got nil")
	}

	if !strings.Contains(err.Error(), "requires 1 argument") {
		t.Errorf("Generate() error = %v, want 'requires 1 argument'", err)
	}
}

func TestArrowValueFunctionGenerator_UnsupportedFunction(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "unsupported"},
		Arguments: []ast.Expression{},
	}

	_, err := valueGen.Generate(call)
	if err == nil {
		t.Error("Generate() expected error for unsupported function, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported value function") {
		t.Errorf("Generate() error = %v, want 'unsupported value function'", err)
	}
}

func TestArrowValueFunctionGenerator_ErrorPropagation(t *testing.T) {
	g := newTestGenerator()
	resolver := NewArrowSeriesAccessResolver()
	exprGen := NewArrowExpressionGeneratorImpl(g, resolver)
	valueGen := NewArrowValueFunctionGenerator(exprGen)

	tests := []struct {
		name          string
		funcName      string
		args          []ast.Expression
		errorContains string
	}{
		{
			name:     "nz with invalid argument type",
			funcName: "nz",
			args: []ast.Expression{
				&ast.ObjectExpression{}, // Unsupported expression type
			},
			errorContains: "nz() argument failed",
		},
		{
			name:     "na with invalid argument type",
			funcName: "na",
			args: []ast.Expression{
				&ast.ObjectExpression{},
			},
			errorContains: "na() argument failed",
		},
		{
			name:     "fixnan with invalid argument type",
			funcName: "fixnan",
			args: []ast.Expression{
				&ast.ObjectExpression{},
			},
			errorContains: "fixnan() source failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: tt.funcName},
				Arguments: tt.args,
			}

			_, err := valueGen.Generate(call)
			if err == nil {
				t.Errorf("Generate() expected error containing %q, got nil", tt.errorContains)
				return
			}

			if !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("Generate() error = %v, want to contain %q", err, tt.errorContains)
			}
		})
	}
}
