package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestPeriodEvaluator_Evaluate_Literals(t *testing.T) {
	warmup := validation.NewWarmupAnalyzer()
	evaluator := NewPeriodEvaluator(warmup)

	tests := []struct {
		name         string
		literal      *ast.Literal
		expectedKind PeriodEvaluationKind
		expectedVal  int
		expectFailed bool
	}{
		{
			name:         "positive int literal",
			literal:      &ast.Literal{Value: 14},
			expectedKind: PeriodKindCompileTime,
			expectedVal:  14,
		},
		{
			name:         "positive float64 literal",
			literal:      &ast.Literal{Value: 20.0},
			expectedKind: PeriodKindCompileTime,
			expectedVal:  20,
		},
		{
			name:         "zero literal fails",
			literal:      &ast.Literal{Value: 0},
			expectedKind: PeriodKindFailed,
			expectFailed: true,
		},
		{
			name:         "negative literal fails",
			literal:      &ast.Literal{Value: -5},
			expectedKind: PeriodKindFailed,
			expectFailed: true,
		},
		{
			name:         "non-numeric literal fails",
			literal:      &ast.Literal{Value: "invalid"},
			expectedKind: PeriodKindFailed,
			expectFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.literal)

			if result.Kind != tt.expectedKind {
				t.Errorf("Kind = %d, want %d", result.Kind, tt.expectedKind)
			}

			if !tt.expectFailed && result.StaticValue != tt.expectedVal {
				t.Errorf("StaticValue = %d, want %d", result.StaticValue, tt.expectedVal)
			}

			if tt.expectFailed && result.FailureReason == "" {
				t.Error("Expected failure reason but got empty string")
			}
		})
	}
}

func TestPeriodEvaluator_Evaluate_RuntimeExpressions(t *testing.T) {
	warmup := validation.NewWarmupAnalyzer()
	evaluator := NewPeriodEvaluator(warmup)

	tests := []struct {
		name         string
		expr         ast.Expression
		expectedKind PeriodEvaluationKind
	}{
		{
			name:         "identifier that cannot be evaluated",
			expr:         &ast.Identifier{Name: "unknownVar"},
			expectedKind: PeriodKindRuntimeDynamic,
		},
		{
			name: "conditional expression",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "bar_index"},
					Operator: ">",
					Right:    &ast.Literal{Value: 100.0},
				},
				Consequent: &ast.Literal{Value: 20.0},
				Alternate:  &ast.Literal{Value: 10.0},
			},
			expectedKind: PeriodKindRuntimeDynamic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(tt.expr)

			if result.Kind != tt.expectedKind {
				t.Errorf("Kind = %d, want %d", result.Kind, tt.expectedKind)
			}

			if result.Kind == PeriodKindRuntimeDynamic && result.DynamicExpr == nil {
				t.Error("Expected DynamicExpr to be set for runtime dynamic period")
			}
		})
	}
}
