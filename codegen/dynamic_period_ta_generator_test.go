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

func TestDynamicPeriodTAGenerator_TAFunctionSupport(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		sourceExpr   ast.Expression
		wantError    bool
		checkCode    func(string) bool
	}{
		{
			name:         "ta.sma generates sum loop",
			functionName: "ta.sma",
			sourceExpr:   &ast.Identifier{Name: "close"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "sum") &&
					strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.ema generates calculation",
			functionName: "ta.ema",
			sourceExpr:   &ast.Identifier{Name: "close"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.stdev generates calculation",
			functionName: "ta.stdev",
			sourceExpr:   &ast.Identifier{Name: "close"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.highest generates calculation",
			functionName: "ta.highest",
			sourceExpr:   &ast.Identifier{Name: "high"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.lowest generates calculation",
			functionName: "ta.lowest",
			sourceExpr:   &ast.Identifier{Name: "low"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.rsi generates calculation",
			functionName: "ta.rsi",
			sourceExpr:   &ast.Identifier{Name: "close"},
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "ta.atr generates calculation",
			functionName: "ta.atr",
			sourceExpr:   nil,
			checkCode: func(code string) bool {
				return strings.Contains(code, "period")
			},
		},
		{
			name:         "unknown function errors",
			functionName: "ta.unknown",
			sourceExpr:   &ast.Identifier{Name: "close"},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			periodResult := NewRuntimeDynamicPeriod(&ast.Identifier{Name: "period"})
			sourceExpr := tt.sourceExpr
			if sourceExpr == nil {
				sourceExpr = &ast.Identifier{Name: "close"}
			}

			code, err := dynGen.Generate("testVar", tt.functionName, sourceExpr, periodResult)

			if tt.wantError {
				if err == nil {
					t.Error("Generate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Generate() unexpected error: %v", err)
			}

			if code == "" {
				t.Error("Generate() returned empty code")
			}

			if tt.checkCode != nil && !tt.checkCode(code) {
				t.Errorf("Generate() code validation failed")
			}
		})
	}
}

func TestDynamicPeriodTAGenerator_CodeStructure(t *testing.T) {
	tests := []struct {
		name          string
		functionName  string
		varName       string
		periodExpr    ast.Expression
		sourceExpr    ast.Expression
		requiredParts []string
	}{
		{
			name:         "SMA contains all required elements",
			functionName: "ta.sma",
			varName:      "mySma",
			periodExpr:   &ast.Identifier{Name: "dynamicLen"},
			sourceExpr:   &ast.Identifier{Name: "close"},
			requiredParts: []string{
				"period := int(dynamicLenSeries.Get(0))",
				"if period <= 0 || ctx.BarIndex < period-1",
				"mySmaSeries.Set(math.NaN())",
				"sum := 0.0",
				"for j := 0; j < period; j++",
				"sum += closeSeries.Get(j)",
				"mySmaSeries.Set(sum / float64(period))",
			},
		},
		{
			name:         "STDEV validates warmup period",
			functionName: "ta.stdev",
			varName:      "myStdev",
			periodExpr:   &ast.Identifier{Name: "len"},
			sourceExpr:   &ast.Identifier{Name: "close"},
			requiredParts: []string{
				"period := int(lenSeries.Get(0))",
				"if period <= 0 || ctx.BarIndex < period-1",
				"myStdevSeries.Set(math.NaN())",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			periodResult := NewRuntimeDynamicPeriod(tt.periodExpr)
			code, err := dynGen.Generate(tt.varName, tt.functionName, tt.sourceExpr, periodResult)

			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			for _, part := range tt.requiredParts {
				if !strings.Contains(code, part) {
					t.Errorf("Generate() missing required element: %s", part)
				}
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
	tests := []struct {
		name           string
		sourceExpr     ast.Expression
		expectedAccess string
	}{
		{
			name:           "identifier becomes series access",
			sourceExpr:     &ast.Identifier{Name: "close"},
			expectedAccess: "closeSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newMinimalGenerator()
			dynGen := NewDynamicPeriodTAGenerator(g)

			periodResult := NewRuntimeDynamicPeriod(&ast.Identifier{Name: "period"})
			code, err := dynGen.Generate("testVar", "ta.sma", tt.sourceExpr, periodResult)

			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			if !strings.Contains(code, tt.expectedAccess) {
				t.Errorf("Generate() expected accessor %q not found in code", tt.expectedAccess)
			}
		})
	}
}
