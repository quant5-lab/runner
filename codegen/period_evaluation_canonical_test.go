package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* TestEvaluatePeriodExpression_LiteralValidation tests period literal validation */
func TestEvaluatePeriodExpression_LiteralValidation(t *testing.T) {
	tests := []struct {
		name       string
		literal    interface{}
		wantKind   PeriodEvaluationKind
		wantValue  int
		wantFailed bool
	}{
		{name: "positive integer literal", literal: 14, wantKind: PeriodKindCompileTime, wantValue: 14},
		{name: "positive float literal", literal: 20.0, wantKind: PeriodKindCompileTime, wantValue: 20},
		{name: "float truncation", literal: 14.7, wantKind: PeriodKindCompileTime, wantValue: 14},
		{name: "float boundary", literal: 19.9, wantKind: PeriodKindCompileTime, wantValue: 19},
		{name: "minimum valid (1)", literal: 1, wantKind: PeriodKindCompileTime, wantValue: 1},
		{name: "large valid", literal: 500, wantKind: PeriodKindCompileTime, wantValue: 500},
		{name: "zero - invalid", literal: 0, wantKind: PeriodKindFailed, wantFailed: true},
		{name: "negative int - invalid", literal: -5, wantKind: PeriodKindFailed, wantFailed: true},
		{name: "negative float - invalid", literal: -10.5, wantKind: PeriodKindFailed, wantFailed: true},
		{name: "zero float - invalid", literal: 0.0, wantKind: PeriodKindFailed, wantFailed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{constEvaluator: validation.NewWarmupAnalyzer()}
			periodArg := &ast.Literal{Value: tt.literal}
			result := evaluatePeriodExpression(g, periodArg)

			if result.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", result.Kind, tt.wantKind)
			}
			if tt.wantFailed {
				if !result.IsFailed() {
					t.Errorf("Expected Failed result")
				}
				if result.FailureReason == "" {
					t.Error("Failed result missing FailureReason")
				}
			} else if result.StaticValue != tt.wantValue {
				t.Errorf("StaticValue = %d, want %d", result.StaticValue, tt.wantValue)
			}
		})
	}
}

/* TestEvaluatePeriodExpression_InputIntExtraction tests input.int() period extraction */
func TestEvaluatePeriodExpression_InputIntExtraction(t *testing.T) {
	tests := []struct {
		name      string
		callee    ast.Expression
		args      []ast.Expression
		wantKind  PeriodEvaluationKind
		wantValue int
	}{
		{
			name:      "input.int integer defval",
			callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
			args:      []ast.Expression{&ast.Literal{Value: 14}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 14,
		},
		{
			name:      "input.int float defval",
			callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
			args:      []ast.Expression{&ast.Literal{Value: 20.0}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 20,
		},
		{
			name:      "input.int with title",
			callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
			args:      []ast.Expression{&ast.Literal{Value: 50}, &ast.Literal{Value: "Period"}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 50,
		},
		{
			name:      "input() legacy",
			callee:    &ast.Identifier{Name: "input"},
			args:      []ast.Expression{&ast.Literal{Value: 10}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 10,
		},
		{
			name:     "input.int zero defval - skips",
			callee:   &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
			args:     []ast.Expression{&ast.Literal{Value: 0}},
			wantKind: PeriodKindRuntimeDynamic,
		},
		{
			name:     "input.int negative defval - skips",
			callee:   &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
			args:     []ast.Expression{&ast.Literal{Value: -5}},
			wantKind: PeriodKindRuntimeDynamic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{constEvaluator: validation.NewWarmupAnalyzer()}
			periodArg := &ast.CallExpression{Callee: tt.callee, Arguments: tt.args}
			result := evaluatePeriodExpression(g, periodArg)

			if result.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", result.Kind, tt.wantKind)
			}
			if tt.wantKind == PeriodKindCompileTime && result.StaticValue != tt.wantValue {
				t.Errorf("StaticValue = %d, want %d", result.StaticValue, tt.wantValue)
			}
		})
	}
}

/* TestEvaluatePeriodExpression_ConstantFolding tests constant folding */
func TestEvaluatePeriodExpression_ConstantFolding(t *testing.T) {
	tests := []struct {
		name      string
		expr      ast.Expression
		wantKind  PeriodEvaluationKind
		wantValue int
	}{
		{
			name:      "addition: 10 + 4",
			expr:      &ast.BinaryExpression{Left: &ast.Literal{Value: 10}, Operator: "+", Right: &ast.Literal{Value: 4}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 14,
		},
		{
			name:      "multiplication: 7 * 2",
			expr:      &ast.BinaryExpression{Left: &ast.Literal{Value: 7}, Operator: "*", Right: &ast.Literal{Value: 2}},
			wantKind:  PeriodKindCompileTime,
			wantValue: 14,
		},
		{
			name: "nested: (10 + 5) * 2",
			expr: &ast.BinaryExpression{
				Left:     &ast.BinaryExpression{Left: &ast.Literal{Value: 10}, Operator: "+", Right: &ast.Literal{Value: 5}},
				Operator: "*",
				Right:    &ast.Literal{Value: 2},
			},
			wantKind:  PeriodKindCompileTime,
			wantValue: 30,
		},
		{
			name:     "variable - not foldable",
			expr:     &ast.Identifier{Name: "dynamicVar"},
			wantKind: PeriodKindRuntimeDynamic,
		},
		{
			name:     "zero result - invalid",
			expr:     &ast.BinaryExpression{Left: &ast.Literal{Value: 0}, Operator: "/", Right: &ast.Literal{Value: 10}},
			wantKind: PeriodKindRuntimeDynamic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{constEvaluator: validation.NewWarmupAnalyzer()}
			result := evaluatePeriodExpression(g, tt.expr)

			if result.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", result.Kind, tt.wantKind)
			}
			if tt.wantKind == PeriodKindCompileTime && result.StaticValue != tt.wantValue {
				t.Errorf("StaticValue = %d, want %d", result.StaticValue, tt.wantValue)
			}
		})
	}
}

/* TestEvaluatePeriodExpression_RuntimeDynamic tests non-constant expressions */
func TestEvaluatePeriodExpression_RuntimeDynamic(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		wantKind PeriodEvaluationKind
	}{
		{
			name:     "ternary expression",
			expr:     &ast.ConditionalExpression{Test: &ast.Identifier{Name: "cond"}, Consequent: &ast.Literal{Value: 14}, Alternate: &ast.Literal{Value: 21}},
			wantKind: PeriodKindRuntimeDynamic,
		},
		{name: "series variable", expr: &ast.Identifier{Name: "mySeriesVar"}, wantKind: PeriodKindRuntimeDynamic},
		{
			name:     "member expression",
			expr:     &ast.MemberExpression{Object: &ast.Identifier{Name: "arr"}, Property: &ast.Literal{Value: 0}, Computed: true},
			wantKind: PeriodKindRuntimeDynamic,
		},
		{
			name: "ta function call",
			expr: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "ta"}, Property: &ast.Identifier{Name: "sma"}},
				Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, &ast.Literal{Value: 20}},
			},
			wantKind: PeriodKindRuntimeDynamic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{constEvaluator: validation.NewWarmupAnalyzer()}
			result := evaluatePeriodExpression(g, tt.expr)

			if result.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", result.Kind, tt.wantKind)
			}
			if result.IsRuntimeDynamic() && result.DynamicExpr == nil {
				t.Error("RuntimeDynamic missing DynamicExpr")
			}
		})
	}
}

/* TestEvaluatePeriodExpression_DelegationConsistency tests wrapper delegation */
func TestEvaluatePeriodExpression_DelegationConsistency(t *testing.T) {
	testCases := []struct {
		name      string
		periodArg ast.Expression
		wantKind  PeriodEvaluationKind
		wantValue int
	}{
		{name: "literal 14", periodArg: &ast.Literal{Value: 14}, wantKind: PeriodKindCompileTime, wantValue: 14},
		{name: "literal zero", periodArg: &ast.Literal{Value: 0}, wantKind: PeriodKindFailed},
		{
			name: "input.int(20)",
			periodArg: &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "input"}, Property: &ast.Identifier{Name: "int"}},
				Arguments: []ast.Expression{&ast.Literal{Value: 20}},
			},
			wantKind:  PeriodKindCompileTime,
			wantValue: 20,
		},
		{name: "variable (dynamic)", periodArg: &ast.Identifier{Name: "dynamicPeriod"}, wantKind: PeriodKindRuntimeDynamic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := &generator{constEvaluator: validation.NewWarmupAnalyzer()}

			directResult := evaluatePeriodExpression(g, tc.periodArg)
			extractor := NewTAArgumentExtractor(g)
			extractorResult := extractor.extractPeriodResult(tc.periodArg, "test")
			singleCall := &ast.CallExpression{Arguments: []ast.Expression{tc.periodArg}}
			singleResult := extractSinglePeriodWithDynamic(g, singleCall, "test")
			twoArgCall := &ast.CallExpression{Arguments: []ast.Expression{&ast.Identifier{Name: "close"}, tc.periodArg}}
			_, twoArgResult := extractTAArgumentsWithDynamic(g, twoArgCall, "test")

			results := []PeriodEvaluationResult{directResult, extractorResult, singleResult, twoArgResult}
			for i, r := range results {
				if r.Kind != tc.wantKind {
					t.Errorf("Path %d: Kind = %v, want %v", i, r.Kind, tc.wantKind)
				}
				if tc.wantKind == PeriodKindCompileTime && r.StaticValue != tc.wantValue {
					t.Errorf("Path %d: StaticValue = %d, want %d", i, r.StaticValue, tc.wantValue)
				}
			}
		})
	}
}
