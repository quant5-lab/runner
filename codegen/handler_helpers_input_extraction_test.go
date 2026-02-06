package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestTryExtractInputIntValue_DirectCalls(t *testing.T) {
	tests := []struct {
		name       string
		expr       ast.Expression
		wantPeriod int
	}{
		{
			name: "input.int with integer literal",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
				},
			},
			wantPeriod: 14,
		},
		{
			name: "input.int with float literal",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 20.0},
				},
			},
			wantPeriod: 20,
		},
		{
			name: "input.int with title argument",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
					&ast.Literal{Value: "Period"},
				},
			},
			wantPeriod: 14,
		},
		{
			name: "input() with integer default",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 10},
				},
			},
			wantPeriod: 10,
		},
		{
			name: "input.int with small period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 1},
				},
			},
			wantPeriod: 1,
		},
		{
			name: "input.int with large period",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 500},
				},
			},
			wantPeriod: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period := tryExtractInputIntValue(tt.expr)
			if period != tt.wantPeriod {
				t.Errorf("tryExtractInputIntValue() = %d, want %d", period, tt.wantPeriod)
			}
		})
	}
}

func TestTryExtractInputIntValue_NotApplicable(t *testing.T) {
	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			name: "non-call expression",
			expr: &ast.Identifier{Name: "period"},
		},
		{
			name: "non-input function",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "getPeriod"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
				},
			},
		},
		{
			name: "input.float (not input.int)",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "float"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14.0},
				},
			},
		},
		{
			name: "input.int with no arguments",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "input.int with non-literal first arg",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "defaultPeriod"},
				},
			},
		},
		{
			name: "input.int with string literal",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "14"},
				},
			},
		},
		{
			name: "input.int with zero (invalid period)",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 0},
				},
			},
		},
		{
			name: "input.int with negative (invalid period)",
			expr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: -5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			period := tryExtractInputIntValue(tt.expr)
			if period != 0 {
				t.Errorf("tryExtractInputIntValue() = %d, want 0 (not applicable)", period)
			}
		})
	}
}

func TestExtractSinglePeriodArgument_InputIntDirectCall(t *testing.T) {
	tests := []struct {
		name       string
		periodExpr ast.Expression
		wantPeriod int
		wantError  bool
	}{
		{
			name: "input.int(14)",
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
				},
			},
			wantPeriod: 14,
		},
		{
			name: "input.int(14, 'Period')",
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
					&ast.Literal{Value: "Period"},
				},
			},
			wantPeriod: 14,
		},
		{
			name: "input(20)",
			periodExpr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 20},
				},
			},
			wantPeriod: 20,
		},
		{
			name: "input.int with minval/maxval",
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
					&ast.Literal{Value: "Length"},
					&ast.Literal{Value: 1},   // minval
					&ast.Literal{Value: 100}, // maxval
				},
			},
			wantPeriod: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				constEvaluator: validation.NewWarmupAnalyzer(),
			}

			call := &ast.CallExpression{
				Arguments: []ast.Expression{tt.periodExpr},
			}

			period, err := extractSinglePeriodArgument(g, call, "ta.atr")

			if tt.wantError {
				if err == nil {
					t.Error("extractSinglePeriodArgument() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("extractSinglePeriodArgument() unexpected error = %v", err)
			}

			if period != tt.wantPeriod {
				t.Errorf("period = %d, want %d", period, tt.wantPeriod)
			}
		})
	}
}

func TestExtractTAArgumentsAST_InputIntDirectCall(t *testing.T) {
	tests := []struct {
		name       string
		sourceExpr ast.Expression
		periodExpr ast.Expression
		wantPeriod int
		wantError  bool
	}{
		{
			name:       "ta.sma(close, input.int(14))",
			sourceExpr: &ast.Identifier{Name: "close"},
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
				},
			},
			wantPeriod: 14,
		},
		{
			name:       "ta.sma(close, input.int(14, 'Period'))",
			sourceExpr: &ast.Identifier{Name: "close"},
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
					&ast.Literal{Value: "Period"},
				},
			},
			wantPeriod: 14,
		},
		{
			name:       "ta.ema(high, input(20))",
			sourceExpr: &ast.Identifier{Name: "high"},
			periodExpr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "input"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 20},
				},
			},
			wantPeriod: 20,
		},
		{
			name: "ta.stdev(close, input.int(100))",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "close"},
				Property: &ast.Identifier{Name: "value"},
			},
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 100},
				},
			},
			wantPeriod: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				constEvaluator: validation.NewWarmupAnalyzer(),
			}

			call := &ast.CallExpression{
				Arguments: []ast.Expression{tt.sourceExpr, tt.periodExpr},
			}

			sourceAST, period, err := extractTAArgumentsAST(g, call, "ta.sma")

			if tt.wantError {
				if err == nil {
					t.Error("extractTAArgumentsAST() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("extractTAArgumentsAST() unexpected error = %v", err)
			}

			if period != tt.wantPeriod {
				t.Errorf("period = %d, want %d", period, tt.wantPeriod)
			}

			if sourceAST == nil {
				t.Error("sourceAST = nil, want non-nil")
			}
		})
	}
}

func TestExtractTAArgumentsAST_EvaluationChain(t *testing.T) {
	tests := []struct {
		name       string
		periodExpr ast.Expression
		constants  map[string]interface{}
		wantPeriod int
		wantError  bool
	}{
		{
			name:       "literal - fast path",
			periodExpr: &ast.Literal{Value: 14},
			wantPeriod: 14,
		},
		{
			name: "input.int() - second priority",
			periodExpr: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "input"},
					Property: &ast.Identifier{Name: "int"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: 20},
				},
			},
			wantPeriod: 20,
		},
		{
			name:       "const variable - third priority",
			periodExpr: &ast.Identifier{Name: "period"},
			constants:  map[string]interface{}{"period": 30},
			wantPeriod: 30,
		},
		{
			name: "binary expression - constEval",
			periodExpr: &ast.BinaryExpression{
				Left:     &ast.Literal{Value: float64(7)},
				Operator: "*",
				Right:    &ast.Literal{Value: float64(2)},
			},
			wantPeriod: 14,
		},
		{
			name: "custom function - fails (not compile-time constant)",
			periodExpr: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "getPeriod"},
				Arguments: []ast.Expression{},
			},
			wantError: true,
		},
		{
			name:       "undefined variable - fails",
			periodExpr: &ast.Identifier{Name: "undefined"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := validation.NewWarmupAnalyzer()
			for k, v := range tt.constants {
				if iv, ok := v.(int); ok {
					analyzer.AddConstant(k, float64(iv))
				} else if fv, ok := v.(float64); ok {
					analyzer.AddConstant(k, fv)
				}
			}

			g := &generator{
				constants:      tt.constants,
				constEvaluator: analyzer,
			}

			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					tt.periodExpr,
				},
			}

			_, period, err := extractTAArgumentsAST(g, call, "ta.sma")

			if tt.wantError {
				if err == nil {
					t.Error("extractTAArgumentsAST() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("extractTAArgumentsAST() unexpected error = %v", err)
			}

			if period != tt.wantPeriod {
				t.Errorf("period = %d, want %d", period, tt.wantPeriod)
			}
		})
	}
}

func TestExtractTAArgumentsAST_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		call      *ast.CallExpression
		wantError bool
		errorMsg  string
	}{
		{
			name: "insufficient arguments",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			wantError: true,
			errorMsg:  "at least 2 arguments",
		},
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			wantError: true,
			errorMsg:  "at least 2 arguments",
		},
		{
			name: "extra arguments - ignored",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 14},
					&ast.Literal{Value: "extra"},
				},
			},
			wantError: false,
		},
		{
			name: "zero period from literal",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 0},
				},
			},
			wantError: true,
			errorMsg:  "period",
		},
		{
			name: "negative period from literal",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: -5},
				},
			},
			wantError: true,
			errorMsg:  "period",
		},
		{
			name: "string literal as period",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: "14"},
				},
			},
			wantError: true,
			errorMsg:  "numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				constEvaluator: validation.NewWarmupAnalyzer(),
			}

			_, period, err := extractTAArgumentsAST(g, tt.call, "ta.sma")

			if tt.wantError {
				if err == nil {
					t.Error("extractTAArgumentsAST() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("extractTAArgumentsAST() unexpected error = %v", err)
			}

			if period <= 0 {
				t.Errorf("period = %d, want > 0", period)
			}
		})
	}
}
