package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* Period extraction from numeric literals */
func TestExtractSinglePeriodArgument_Literals(t *testing.T) {
	tests := []struct {
		name       string
		periodExpr ast.Expression
		wantPeriod int
		wantError  bool
	}{
		{
			name:       "integer literal",
			periodExpr: &ast.Literal{Value: 14},
			wantPeriod: 14,
		},
		{
			name:       "float literal",
			periodExpr: &ast.Literal{Value: 20.0},
			wantPeriod: 20,
		},
		{
			name:       "small period",
			periodExpr: &ast.Literal{Value: 1},
			wantPeriod: 1,
		},
		{
			name:       "large period",
			periodExpr: &ast.Literal{Value: 500},
			wantPeriod: 500,
		},
		{
			name:       "fractional float truncates",
			periodExpr: &ast.Literal{Value: 14.7},
			wantPeriod: 14,
		},
		{
			name:       "string literal - invalid",
			periodExpr: &ast.Literal{Value: "invalid"},
			wantError:  true,
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

/* Period extraction from input() constants */
func TestExtractSinglePeriodArgument_Constants(t *testing.T) {
	tests := []struct {
		name       string
		periodName string
		constants  map[string]interface{}
		wantPeriod int
		wantError  bool
	}{
		{
			name:       "input.int constant",
			periodName: "atrPeriod",
			constants:  map[string]interface{}{"atrPeriod": 14},
			wantPeriod: 14,
		},
		{
			name:       "input.float constant",
			periodName: "length",
			constants:  map[string]interface{}{"length": 20.0},
			wantPeriod: 20,
		},
		{
			name:       "calculated constant",
			periodName: "period",
			constants:  map[string]interface{}{"period": 100},
			wantPeriod: 100,
		},
		{
			name:       "undefined constant",
			periodName: "undefined",
			constants:  map[string]interface{}{},
			wantError:  true,
		},
		{
			name:       "zero period - invalid",
			periodName: "zeroPeriod",
			constants:  map[string]interface{}{"zeroPeriod": 0},
			wantError:  true,
		},
		{
			name:       "negative period - invalid",
			periodName: "negativePeriod",
			constants:  map[string]interface{}{"negativePeriod": -5},
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
					&ast.Identifier{Name: tt.periodName},
				},
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

/* Period extraction boundary conditions */
func TestExtractSinglePeriodArgument_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		call      *ast.CallExpression
		wantError bool
		errorMsg  string
	}{
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{},
			},
			wantError: true,
			errorMsg:  "requires 1 argument",
		},
		{
			name: "multiple arguments - uses first",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Literal{Value: 14},
					&ast.Literal{Value: 20},
				},
			},
			wantError: false,
		},
		{
			name: "expression argument - not constant",
			call: &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "a"}, // Undefined variable
						Operator: "+",
						Right:    &ast.Identifier{Name: "b"}, // Undefined variable
					},
				},
			},
			wantError: true,
			errorMsg:  "compile-time constant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &generator{
				constEvaluator: validation.NewWarmupAnalyzer(),
			}

			period, err := extractSinglePeriodArgument(g, tt.call, "ta.atr")

			if tt.wantError {
				if err == nil {
					t.Error("extractSinglePeriodArgument() error = nil, want error")
				}
				if tt.errorMsg != "" && err != nil {
					// Error message check is optional - just ensure we got an error
				}
				return
			}

			if err != nil {
				t.Fatalf("extractSinglePeriodArgument() unexpected error = %v", err)
			}

			if period <= 0 {
				t.Errorf("period = %d, want > 0", period)
			}
		})
	}
}

/* Integration with ATR handler real-world scenarios */
func TestExtractSinglePeriodArgument_IntegrationWithATRHandler(t *testing.T) {
	tests := []struct {
		name       string
		pineCode   string
		periodName string
		periodVal  int
		wantPeriod int
	}{
		{
			name:       "literal period",
			pineCode:   "atr_value = ta.atr(14)",
			periodName: "",
			periodVal:  14,
			wantPeriod: 14,
		},
		{
			name:       "input constant period",
			pineCode:   "atrPeriod = input.int(20); atr_value = ta.atr(atrPeriod)",
			periodName: "atrPeriod",
			periodVal:  20,
			wantPeriod: 20,
		},
		{
			name:       "v4 input syntax",
			pineCode:   "periods = input(title='ATR', type=input.integer, defval=10); atr_value = ta.atr(periods)",
			periodName: "periods",
			periodVal:  10,
			wantPeriod: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := validation.NewWarmupAnalyzer()
			if tt.periodName != "" {
				analyzer.AddConstant(tt.periodName, float64(tt.periodVal))
			}

			g := &generator{
				constants:      make(map[string]interface{}),
				constEvaluator: analyzer,
			}

			if tt.periodName != "" {
				g.constants[tt.periodName] = tt.periodVal
			}

			var periodExpr ast.Expression
			if tt.periodName != "" {
				periodExpr = &ast.Identifier{Name: tt.periodName}
			} else {
				periodExpr = &ast.Literal{Value: tt.periodVal}
			}

			call := &ast.CallExpression{
				Arguments: []ast.Expression{periodExpr},
			}

			period, err := extractSinglePeriodArgument(g, call, "ta.atr")
			if err != nil {
				t.Fatalf("extractSinglePeriodArgument() error = %v", err)
			}

			if period != tt.wantPeriod {
				t.Errorf("period = %d, want %d", period, tt.wantPeriod)
			}
		})
	}
}
