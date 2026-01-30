package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestLinregHandler_TAFunctionHandlerInterface validates LinregHandler implements TAFunctionHandler correctly */
func TestLinregHandler_TAFunctionHandlerInterface(t *testing.T) {
	handler := &LinregHandler{}

	t.Run("can_handle_ta_dot_linreg", func(t *testing.T) {
		if !handler.CanHandle("ta.linreg") {
			t.Error("LinregHandler should handle 'ta.linreg'")
		}
	})

	t.Run("can_handle_linreg", func(t *testing.T) {
		if !handler.CanHandle("linreg") {
			t.Error("LinregHandler should handle 'linreg' (Pine v4 syntax)")
		}
	})

	t.Run("cannot_handle_other_functions", func(t *testing.T) {
		invalidFuncs := []string{"ta.sma", "ta.ema", "ta.rsi", "ta.highest", "ta.lowest", "linregslope", ""}
		for _, fn := range invalidFuncs {
			if handler.CanHandle(fn) {
				t.Errorf("LinregHandler should not handle '%s'", fn)
			}
		}
	})
}

/* TestLinregHandler_CodeGenerationDispatch validates handler generates code via generator */
func TestLinregHandler_CodeGenerationDispatch(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
			&ast.Literal{Value: 0.0},
		},
	}

	code, err := handler.GenerateCode(gen, "myLinreg", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if code == "" {
		t.Fatal("GenerateCode() returned empty string")
	}

	t.Run("contains_least_squares_calculation", func(t *testing.T) {
		patterns := []string{"sumX", "sumY", "sumXY", "sumX2", "slope", "intercept"}
		for _, pattern := range patterns {
			if !strings.Contains(code, pattern) {
				t.Errorf("Generated code missing least squares variable: %s", pattern)
			}
		}
	})

	t.Run("contains_series_storage", func(t *testing.T) {
		if !strings.Contains(code, "Series.Set") {
			t.Error("Generated code missing Series.Set() for final linreg value")
		}
	})

	t.Run("contains_warmup_check", func(t *testing.T) {
		if !strings.Contains(code, "ctx.BarIndex") {
			t.Error("Generated code missing warmup check (ctx.BarIndex)")
		}
	})

	t.Run("contains_iife_structure", func(t *testing.T) {
		if !strings.Contains(code, "func() float64") {
			t.Error("Generated code missing IIFE structure")
		}
	})
}

/* TestLinregHandler_IntegrationWithTARegistry validates TAFunctionRegistry routes to LinregHandler */
func TestLinregHandler_IntegrationWithTARegistry(t *testing.T) {
	registry := NewTAFunctionRegistry()

	t.Run("registry_has_linreg_handler", func(t *testing.T) {
		handler := registry.FindHandler("ta.linreg")
		if handler == nil {
			t.Fatal("TAFunctionRegistry missing Linreg handler")
		}

		if _, ok := handler.(*LinregHandler); !ok {
			t.Errorf("Registry returned wrong handler type: %T", handler)
		}
	})

	t.Run("registry_supports_linreg", func(t *testing.T) {
		if !registry.IsSupported("ta.linreg") {
			t.Error("TAFunctionRegistry should support 'ta.linreg'")
		}

		if !registry.IsSupported("linreg") {
			t.Error("TAFunctionRegistry should support 'linreg'")
		}
	})

	t.Run("registry_generates_code", func(t *testing.T) {
		gen := newTestGenerator()
		gen.variables["close"] = "series"

		call := &ast.CallExpression{
			Arguments: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 14.0},
				&ast.Literal{Value: 0.0},
			},
		}

		code, err := registry.GenerateInlineTA(gen, "testLinreg", "ta.linreg", call)
		if err != nil {
			t.Fatalf("GenerateInlineTA() error = %v", err)
		}

		if code == "" {
			t.Error("Registry dispatched to handler but got empty code")
		}

		if !strings.Contains(code, "slope") || !strings.Contains(code, "intercept") {
			t.Error("Registry-generated code missing linear regression logic")
		}
	})
}

/* TestLinregHandler_ArgumentValidation validates proper argument handling */
func TestLinregHandler_ArgumentValidation(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no_arguments",
			args:    []ast.Expression{},
			wantErr: true,
			errMsg:  "3 arguments",
		},
		{
			name: "single_argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
			},
			wantErr: true,
			errMsg:  "3 arguments",
		},
		{
			name: "two_arguments_missing_offset",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
			},
			wantErr: true,
			errMsg:  "3 arguments",
		},
		{
			name: "valid_three_arguments",
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: 20.0},
				&ast.Literal{Value: 0.0},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: tt.args,
			}

			_, err := handler.GenerateCode(gen, "test", call)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

/* TestLinregHandler_SourceExpressions validates handling of different source types */
func TestLinregHandler_SourceExpressions(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"
	gen.variables["myPrice"] = "series"

	tests := []struct {
		name       string
		sourceExpr ast.Expression
		period     float64
		wantErr    bool
	}{
		{
			name:       "identifier_source",
			sourceExpr: &ast.Identifier{Name: "close"},
			period:     14.0,
			wantErr:    false,
		},
		{
			name: "binary_expression_source",
			sourceExpr: &ast.BinaryExpression{
				Operator: "-",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 5.0},
			},
			period:  10.0,
			wantErr: false,
		},
		{
			name: "member_expression_source",
			sourceExpr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "bar"},
				Property: &ast.Identifier{Name: "close"},
			},
			period:  20.0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					tt.sourceExpr,
					&ast.Literal{Value: tt.period},
					&ast.Literal{Value: 0.0},
				},
			}

			code, err := handler.GenerateCode(gen, "testVar", call)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCode() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && code != "" {
				if !strings.Contains(code, "Series.Set") {
					t.Error("Generated code should contain Series.Set()")
				}
			}
		})
	}
}

/* TestLinregHandler_PeriodExpressions validates handling of different period types */
func TestLinregHandler_PeriodExpressions(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"
	gen.constants["myPeriod"] = 20.0
	gen.constEvaluator.AddConstant("myPeriod", 20.0)

	tests := []struct {
		name        string
		periodExpr  ast.Expression
		wantPattern string
	}{
		{
			name:        "literal_period",
			periodExpr:  &ast.Literal{Value: 14.0},
			wantPattern: "for j := 0; j < 14",
		},
		{
			name:        "small_period",
			periodExpr:  &ast.Literal{Value: 5.0},
			wantPattern: "for j := 0; j < 5",
		},
		{
			name:        "large_period",
			periodExpr:  &ast.Literal{Value: 100.0},
			wantPattern: "for j := 0; j < 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					tt.periodExpr,
					&ast.Literal{Value: 0.0},
				},
			}

			code, err := handler.GenerateCode(gen, "testVar", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, tt.wantPattern) {
				t.Errorf("Generated code missing pattern %q\nGot: %s", tt.wantPattern, code)
			}
		})
	}
}

/* TestLinregHandler_WarmupBehavior validates warmup period calculation */
func TestLinregHandler_WarmupBehavior(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"

	tests := []struct {
		name           string
		period         float64
		wantWarmupEdge int
	}{
		{
			name:           "period_5_warmup_4",
			period:         5.0,
			wantWarmupEdge: 4,
		},
		{
			name:           "period_14_warmup_13",
			period:         14.0,
			wantWarmupEdge: 13,
		},
		{
			name:           "period_20_warmup_19",
			period:         20.0,
			wantWarmupEdge: 19,
		},
		{
			name:           "period_1_warmup_0",
			period:         1.0,
			wantWarmupEdge: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: tt.period},
					&ast.Literal{Value: 0.0},
				},
			}

			code, err := handler.GenerateCode(gen, "testVar", call)
			if err != nil {
				t.Fatalf("GenerateCode() error = %v", err)
			}

			if !strings.Contains(code, "ctx.BarIndex") {
				t.Error("Generated code missing warmup check")
			}
		})
	}
}

/* TestLinregHandler_IIFEStructure validates IIFE code generation pattern */
func TestLinregHandler_IIFEStructure(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 10.0},
			&ast.Literal{Value: 0.0},
		},
	}

	code, err := handler.GenerateCode(gen, "testVar", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	t.Run("has_iife_wrapper", func(t *testing.T) {
		if !strings.Contains(code, "func() float64") {
			t.Error("Code should have IIFE wrapper: func() float64")
		}
		if !strings.Contains(code, "}()") {
			t.Error("Code should have IIFE invocation: }()")
		}
	})

	t.Run("has_return_statement", func(t *testing.T) {
		if !strings.Contains(code, "return") {
			t.Error("IIFE should contain return statement")
		}
	})

	t.Run("assigns_to_series", func(t *testing.T) {
		if !strings.Contains(code, "testVarSeries.Set(") {
			t.Error("Code should assign result to testVarSeries.Set()")
		}
	})
}

/* TestLinregHandler_LeastSquaresAlgorithm validates core algorithm components */
func TestLinregHandler_LeastSquaresAlgorithm(t *testing.T) {
	handler := &LinregHandler{}
	gen := newTestGenerator()
	gen.variables["close"] = "series"

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 20.0},
			&ast.Literal{Value: 0.0},
		},
	}

	code, err := handler.GenerateCode(gen, "testVar", call)
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	requiredComponents := []struct {
		component string
		desc      string
	}{
		{"n := float64(", "sample size variable"},
		{"sumX := 0.0", "sum of x coordinates"},
		{"sumY := 0.0", "sum of y coordinates"},
		{"sumXY := 0.0", "sum of x*y products"},
		{"sumX2 := 0.0", "sum of x squared"},
		{"for j := 0; j <", "iteration over window"},
		{"x := float64(j)", "x coordinate calculation"},
		{"sumX += x", "accumulate x"},
		{"sumY += y", "accumulate y"},
		{"sumXY += x * y", "accumulate x*y"},
		{"sumX2 += x * x", "accumulate x squared"},
		{"slope :=", "slope calculation"},
		{"intercept :=", "intercept calculation"},
		{"(n * sumXY - sumX * sumY)", "slope numerator"},
		{"(n * sumX2 - sumX * sumX)", "slope denominator"},
		{"intercept + slope *", "linreg formula"},
	}

	for _, rc := range requiredComponents {
		t.Run(rc.desc, func(t *testing.T) {
			if !strings.Contains(code, rc.component) {
				t.Errorf("Generated code missing %s: %q", rc.desc, rc.component)
			}
		})
	}
}

/* TestLinregHandler_CodeUniqueness ensures generated code is deterministic */
func TestLinregHandler_CodeUniqueness(t *testing.T) {
	handler := &LinregHandler{}

	gen1 := newTestGenerator()
	gen1.variables["close"] = "series"

	gen2 := newTestGenerator()
	gen2.variables["close"] = "series"

	call := &ast.CallExpression{
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "close"},
			&ast.Literal{Value: 14.0},
			&ast.Literal{Value: 0.0},
		},
	}

	code1, err1 := handler.GenerateCode(gen1, "myLinreg", call)
	if err1 != nil {
		t.Fatalf("First GenerateCode() error = %v", err1)
	}

	code2, err2 := handler.GenerateCode(gen2, "myLinreg", call)
	if err2 != nil {
		t.Fatalf("Second GenerateCode() error = %v", err2)
	}

	if code1 != code2 {
		t.Error("GenerateCode() should produce identical output for same inputs")
	}
}
