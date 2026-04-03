package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func newMinimalGenerator() *generator {
	gen := &generator{
		imports:          make(map[string]bool),
		variables:        make(map[string]string),
		varInits:         make(map[string]ast.Expression),
		constants:        make(map[string]interface{}),
		strategyConfig:   NewStrategyConfig(),
		constEvaluator:   validation.NewWarmupAnalyzer(),
		literalFormatter: NewLiteralFormatter(),
	}
	return gen
}

func TestDynamicPeriodTAGenerator_PeriodExpressionRendering(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name:     "identifier to series access",
			expr:     &ast.Identifier{Name: "myPeriod"},
			expected: "myPeriodSeries.Get(0)",
		},
		{
			name:     "numeric literal",
			expr:     &ast.Literal{Value: 20.0},
			expected: "20\n",
		},
		{
			name: "binary addition",
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "basePeriod"},
				Operator: "+",
				Right:    &ast.Literal{Value: 5.0},
			},
			expected: "(basePeriodSeries.Get(0) + 5\n)",
		},
		{
			name: "binary multiplication",
			expr: &ast.BinaryExpression{
				Left:     &ast.Literal{Value: 2.0},
				Operator: "*",
				Right:    &ast.Identifier{Name: "factor"},
			},
			expected: "(2\n * factorSeries.Get(0))",
		},
		{
			name: "nested binary expression",
			expr: &ast.BinaryExpression{
				Left: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "base"},
					Operator: "+",
					Right:    &ast.Literal{Value: 5.0},
				},
				Operator: "*",
				Right:    &ast.Literal{Value: 2.0},
			},
			expected: "((baseSeries.Get(0) + 5\n) * 2\n)",
		},
		{
			name: "ternary conditional simple",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "bar_index"},
					Operator: ">",
					Right:    &ast.Literal{Value: 100.0},
				},
				Consequent: &ast.Literal{Value: 20.0},
				Alternate:  &ast.Literal{Value: 10.0},
			},
			expected: "func() float64 { if (bar_indexSeries.Get(0) > 100\n) != 0 { return 20\n } else { return 10\n } }()",
		},
		{
			name: "ternary with identifier branches",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "condition"},
					Operator: "!=",
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Identifier{Name: "longPeriod"},
				Alternate:  &ast.Identifier{Name: "shortPeriod"},
			},
			expected: "func() float64 { if (conditionSeries.Get(0) != 0\n) != 0 { return longPeriodSeries.Get(0) } else { return shortPeriodSeries.Get(0) } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			result := dynGen.renderPeriodExpression(tt.expr)

			if result != tt.expected {
				t.Errorf("renderPeriodExpression()\ngot:  %q\nwant: %q", result, tt.expected)
			}
		})
	}
}

func TestDynamicPeriodTAGenerator_EmitterDispatch(t *testing.T) {
	supportedFunctions := []struct {
		name       string
		funcName   string
		sourceExpr ast.Expression
	}{
		{"ta.sma", "ta.sma", &ast.Identifier{Name: "close"}},
		{"ta.ema", "ta.ema", &ast.Identifier{Name: "close"}},
		{"ta.rsi", "ta.rsi", &ast.Identifier{Name: "close"}},
		{"ta.stdev", "ta.stdev", &ast.Identifier{Name: "close"}},
		{"ta.highest", "ta.highest", &ast.Identifier{Name: "high"}},
		{"ta.lowest", "ta.lowest", &ast.Identifier{Name: "low"}},
		{"ta.atr with nil source", "ta.atr", nil},
	}

	for _, tt := range supportedFunctions {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)
			periodResult := NewRuntimeDynamicPeriod(&ast.Identifier{Name: "period"})

			code, err := dynGen.Generate("testVar", tt.funcName, tt.sourceExpr, periodResult)
			if err != nil {
				t.Fatalf("Generate() unexpected error: %v", err)
			}
			if code == "" {
				t.Fatal("Generate() returned empty code")
			}
			if !strings.Contains(code, "period := int(") {
				t.Error("generated code should contain period conversion")
			}
			if !strings.Contains(code, "testVarSeries.Set(") {
				t.Error("generated code should write to testVarSeries")
			}
		})
	}

	t.Run("unknown function returns error", func(t *testing.T) {
		g := newMinimalGenerator()
		dynGen := NewDynamicPeriodTAGenerator(g)
		periodResult := NewRuntimeDynamicPeriod(&ast.Identifier{Name: "period"})

		_, err := dynGen.Generate("testVar", "ta.unknown", &ast.Identifier{Name: "close"}, periodResult)
		if err == nil {
			t.Fatal("expected error for unsupported function")
		}
		if !strings.Contains(err.Error(), "ta.unknown") {
			t.Errorf("error should reference function name, got: %v", err)
		}
	})
}

func TestDynamicPeriodTAGenerator_ScopeIsolation(t *testing.T) {
	g := newMinimalGenerator()
	dynGen := NewDynamicPeriodTAGenerator(g)

	periodResult := NewRuntimeDynamicPeriod(&ast.Identifier{Name: "dynamicLen"})
	code, err := dynGen.Generate("mySma", "ta.sma", &ast.Identifier{Name: "close"}, periodResult)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	t.Run("opens_scope_block", func(t *testing.T) {
		if !strings.Contains(code, "{\n") {
			t.Error("generated code should open a scope block")
		}
	})

	t.Run("closes_scope_block", func(t *testing.T) {
		trimmed := strings.TrimSpace(code)
		if !strings.HasSuffix(trimmed, "}") {
			t.Error("generated code should close the scope block")
		}
	})

	t.Run("period_declaration_inside_scope", func(t *testing.T) {
		openIdx := strings.Index(code, "{\n")
		periodIdx := strings.Index(code, "period := int(")
		if periodIdx <= openIdx {
			t.Error("period declaration should be inside the scope block")
		}
	})
}

func TestDynamicPeriodTAGenerator_PeriodExpressionIntegration(t *testing.T) {
	tests := []struct {
		name         string
		periodExpr   ast.Expression
		wantContains string
	}{
		{
			name:         "identifier period",
			periodExpr:   &ast.Identifier{Name: "dynamicLen"},
			wantContains: "period := int(dynamicLenSeries.Get(0))",
		},
		{
			name: "binary expression period",
			periodExpr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "base"},
				Operator: "+",
				Right:    &ast.Literal{Value: 5.0},
			},
			wantContains: "period := int((baseSeries.Get(0) + 5\n))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			periodResult := NewRuntimeDynamicPeriod(tt.periodExpr)
			code, err := dynGen.Generate("testVar", "ta.sma", &ast.Identifier{Name: "close"}, periodResult)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if !strings.Contains(code, tt.wantContains) {
				t.Errorf("expected %q in generated code", tt.wantContains)
			}
		})
	}
}

func TestDynamicPeriodTAGenerator_PeriodKindHandling(t *testing.T) {
	tests := []struct {
		name         string
		periodResult PeriodEvaluationResult
		wantCode     bool
		wantError    bool
	}{
		{
			name:         "runtime dynamic generates code",
			periodResult: NewRuntimeDynamicPeriod(&ast.Identifier{Name: "period"}),
			wantCode:     true,
		},
		{
			name:         "compile time constant returns empty",
			periodResult: NewCompileTimeConstantPeriod(20),
			wantCode:     false,
		},
		{
			name:         "failed period returns empty",
			periodResult: NewFailedPeriodEvaluation("test error"),
			wantCode:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			sourceExpr := &ast.Identifier{Name: "close"}
			code, err := dynGen.Generate("testVar", "ta.sma", sourceExpr, tt.periodResult)

			if tt.wantError {
				if err == nil {
					t.Error("Generate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Generate() unexpected error: %v", err)
			}

			if tt.wantCode && code == "" {
				t.Error("Generate() expected code, got empty string")
			}

			if !tt.wantCode && code != "" {
				t.Errorf("Generate() expected empty, got: %s", code)
			}
		})
	}
}

func TestDynamicPeriodTAGenerator_SourceAccessorExtraction(t *testing.T) {
	g := newMinimalGenerator()
	dynGen := NewDynamicPeriodTAGenerator(g)

	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		expectedAccess string
	}{
		{
			name:           "identifier maps to series",
			sourceExpr:     &ast.Identifier{Name: "close"},
			expectedAccess: "closeSeries",
		},
		{
			name:           "custom identifier maps to series",
			sourceExpr:     &ast.Identifier{Name: "myPrice"},
			expectedAccess: "myPriceSeries",
		},
		{
			name: "member expression concatenates",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "close"},
			},
			expectedAccess: "barcloseSeries",
		},
		{
			name:           "nil source returns empty string",
			sourceExpr:     nil,
			expectedAccess: "",
		},
		{
			name:           "unknown expression type defaults to closeSeries",
			sourceExpr:     &ast.Literal{Value: 42.0},
			expectedAccess: "closeSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dynGen.extractSourceAccessor(tt.sourceExpr)
			if result != tt.expectedAccess {
				t.Errorf("extractSourceAccessor() = %q, want %q", result, tt.expectedAccess)
			}
		})
	}
}
