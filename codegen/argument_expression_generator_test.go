package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func newTestArgumentExpressionGenerator(funcName string, paramIdx int) *ArgumentExpressionGenerator {
	gen := newTestGenerator()
	return NewArgumentExpressionGenerator(gen, funcName, paramIdx)
}

func newTestArrowArgumentExpressionGenerator(funcName string, paramIdx int) *ArgumentExpressionGenerator {
	gen := newTestGenerator()
	gen.inArrowFunctionBody = true
	return NewArgumentExpressionGenerator(gen, funcName, paramIdx)
}

func TestArgumentExpressionGenerator_ExpressionTypeDispatch(t *testing.T) {
	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
	}{
		{
			name:     "literal float",
			expr:     &ast.Literal{Value: 3.14},
			expected: "3.1",
		},
		{
			name:     "literal int",
			expr:     &ast.Literal{Value: 42},
			expected: "42.0",
		},
		{
			name:     "identifier user variable",
			expr:     &ast.Identifier{Name: "src"},
			expected: "srcSeries.GetCurrent()",
		},
		{
			name: "binary expression",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Literal{Value: 1.0},
			},
			expected: "(aSeries.GetCurrent() + 1.0)",
		},
		{
			name: "unary negation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.Literal{Value: 5.0},
				Prefix:   true,
			},
			expected: "-5.0",
		},
		{
			name: "unary not",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.Identifier{Name: "flag"},
				Prefix:   true,
			},
			expected: "func() float64 { if !(value.IsTrue(flagSeries.GetCurrent())) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "logical and",
			expr: &ast.LogicalExpression{
				Operator: "and",
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			},
			expected: "func() float64 { if ((value.IsTrue(aSeries.GetCurrent())) && (value.IsTrue(bSeries.GetCurrent()))) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "logical or",
			expr: &ast.LogicalExpression{
				Operator: "or",
				Left:     &ast.Identifier{Name: "x"},
				Right:    &ast.Identifier{Name: "y"},
			},
			expected: "func() float64 { if ((value.IsTrue(xSeries.GetCurrent())) || (value.IsTrue(ySeries.GetCurrent()))) { return 1.0 } else { return 0.0 } }()",
		},
		{
			name: "conditional ternary",
			expr: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Literal{Value: 0.0},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expected: "func() float64 { if (aSeries.GetCurrent() > 0.0) { return 1.0 } else { return 0.0 } }()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aeg := newTestArgumentExpressionGenerator("myFunc", 0)
			result, err := aeg.Generate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("got:\n  %s\nwant:\n  %s", result, tt.expected)
			}
		})
	}
}

func TestArgumentExpressionGenerator_UnsupportedTypeReturnsError(t *testing.T) {
	aeg := newTestArgumentExpressionGenerator("myFunc", 0)

	_, err := aeg.Generate(&ast.ArrowFunctionExpression{})
	if err == nil {
		t.Fatal("expected error for unsupported expression type")
	}
	if !strings.Contains(err.Error(), "unsupported argument expression type") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestArgumentExpressionGenerator_NestedComposition(t *testing.T) {
	tests := []struct {
		name           string
		expr           ast.Expression
		expectContains []string
	}{
		{
			name: "unary negation inside binary multiplication",
			expr: &ast.BinaryExpression{
				Operator: "*",
				Left:     &ast.Identifier{Name: "x"},
				Right: &ast.UnaryExpression{
					Operator: "-",
					Argument: &ast.Literal{Value: 0.5},
					Prefix:   true,
				},
			},
			expectContains: []string{"xSeries.GetCurrent()", "*", "-0.5"},
		},
		{
			name: "logical inside conditional test",
			expr: &ast.ConditionalExpression{
				Test: &ast.LogicalExpression{
					Operator: "and",
					Left: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "a"},
						Right:    &ast.Literal{Value: 0.0},
					},
					Right: &ast.BinaryExpression{
						Operator: "<",
						Left:     &ast.Identifier{Name: "b"},
						Right:    &ast.Literal{Value: 100.0},
					},
				},
				Consequent: &ast.Literal{Value: 1.0},
				Alternate:  &ast.Literal{Value: 0.0},
			},
			expectContains: []string{"if", "&&", "return 1.0", "return 0.0"},
		},
		{
			name: "conditional inside binary addition",
			expr: &ast.BinaryExpression{
				Operator: "+",
				Left:     &ast.Identifier{Name: "base"},
				Right: &ast.ConditionalExpression{
					Test: &ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "x"},
						Right:    &ast.Literal{Value: 0.0},
					},
					Consequent: &ast.Literal{Value: 10.0},
					Alternate:  &ast.Literal{Value: 0.0},
				},
			},
			expectContains: []string{"baseSeries.GetCurrent()", "+", "func() float64"},
		},
		{
			name: "not wrapping logical expression",
			expr: &ast.UnaryExpression{
				Operator: "not",
				Argument: &ast.LogicalExpression{
					Operator: "or",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				},
				Prefix: true,
			},
			expectContains: []string{"!", "||"},
		},
		{
			name: "binary inside unary negation",
			expr: &ast.UnaryExpression{
				Operator: "-",
				Argument: &ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "a"},
					Right:    &ast.Identifier{Name: "b"},
				},
				Prefix: true,
			},
			expectContains: []string{"-(aSeries.GetCurrent() + bSeries.GetCurrent())"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aeg := newTestArgumentExpressionGenerator("myFunc", 0)
			result, err := aeg.Generate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, sub := range tt.expectContains {
				if !strings.Contains(result, sub) {
					t.Errorf("result %q missing expected substring %q", result, sub)
				}
			}
		})
	}
}

func TestArgumentExpressionGenerator_UDFCompoundArguments(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "unary negation in UDF argument",
			pine: `
//@version=5
indicator("Test")
negate(v) =>
    -v
result = negate(-close)
`,
			mustContain: []string{
				"negate(arrowCtx_negate_1,",
			},
		},
		{
			name: "binary with negated literal coefficient",
			pine: `
//@version=5
indicator("Test")
scale(v) =>
    v * 100
result = scale(close * -0.5)
`,
			mustContain: []string{
				"scale(arrowCtx_scale_1,",
			},
		},
		{
			name: "conditional ternary as UDF argument",
			pine: `
//@version=5
indicator("Test")
pick(v) =>
    v * 2
result = pick(close > open ? high : low)
`,
			mustContain: []string{
				"pick(arrowCtx_pick_1,",
				"func() float64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestArgumentExpressionGenerator_IfForStatementDispatch(t *testing.T) {
	t.Run("IfStatement produces IIFE returning float64", func(t *testing.T) {
		aeg := newTestArgumentExpressionGenerator("plot", 0)
		ifStmt := &ast.IfStatement{
			Test: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "close"},
				Right:    &ast.Literal{Value: 100.0},
			},
			Consequent: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Literal{Value: 1.0},
				},
			},
			Alternate: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Literal{Value: 0.0},
				},
			},
		}

		result, err := aeg.Generate(ifStmt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sub := range []string{"func() float64", "if"} {
			if !strings.Contains(result, sub) {
				t.Errorf("result %q missing expected substring %q", result, sub)
			}
		}
	})

	t.Run("ForStatement produces IIFE returning float64", func(t *testing.T) {
		aeg := newTestArgumentExpressionGenerator("plot", 0)
		forStmt := &ast.ForStatement{
			Counter: "i",
			From:    &ast.Literal{Value: 0},
			To:      &ast.Literal{Value: 10},
			Body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Identifier{Name: "i"},
				},
			},
		}

		result, err := aeg.Generate(forStmt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sub := range []string{"func() float64", "for"} {
			if !strings.Contains(result, sub) {
				t.Errorf("result %q missing expected substring %q", result, sub)
			}
		}
	})
}

func TestArgumentExpressionGenerator_ScopeAwareIdentifierResolution(t *testing.T) {
	type testCase struct {
		name           string
		scope          AccessScope
		identifier     string
		paramType      FunctionParameterType
		expectContains []string
		mustNotContain []string
	}

	tests := []testCase{
		/* BarLoopScope — OHLCV builtins value params */
		{"bar-loop close value", BarLoopScope, "close", ParamTypeScalar, []string{"closeSeries.Get(0)"}, nil},
		{"bar-loop open value", BarLoopScope, "open", ParamTypeScalar, []string{"openSeries.Get(0)"}, nil},
		{"bar-loop high value", BarLoopScope, "high", ParamTypeScalar, []string{"highSeries.Get(0)"}, nil},
		{"bar-loop low value", BarLoopScope, "low", ParamTypeScalar, []string{"lowSeries.Get(0)"}, nil},
		{"bar-loop volume value", BarLoopScope, "volume", ParamTypeScalar, []string{"volumeSeries.Get(0)"}, nil},

		/* BarLoopScope — OHLCV builtins series params */
		{"bar-loop close series", BarLoopScope, "close", ParamTypeSeries, []string{"closeSeries"}, []string{"closeSeries.Get(0)", "ctx.LookupSeries"}},
		{"bar-loop open series", BarLoopScope, "open", ParamTypeSeries, []string{"openSeries"}, []string{"openSeries.Get(0)", "ctx.LookupSeries"}},
		{"bar-loop volume series", BarLoopScope, "volume", ParamTypeSeries, []string{"volumeSeries"}, []string{"volumeSeries.Get(0)", "ctx.LookupSeries"}},

		/* BarLoopScope — other builtins */
		{"bar-loop bar_index", BarLoopScope, "bar_index", ParamTypeScalar, []string{"float64(i)"}, []string{"ctx.BarIndex"}},
		{"bar-loop time", BarLoopScope, "time", ParamTypeScalar, []string{"float64(bar.Time * 1000)"}, nil},

		/* SecurityScope — OHLCV builtins value params */
		{"security close value", SecurityScope, "close", ParamTypeScalar, []string{"closeSeries.GetCurrent()"}, []string{"bar.Close", "ctx.Data[ctx.BarIndex]"}},
		{"security open value", SecurityScope, "open", ParamTypeScalar, []string{"openSeries.GetCurrent()"}, []string{"bar.Open", "ctx.Data[ctx.BarIndex]"}},
		{"security high value", SecurityScope, "high", ParamTypeScalar, []string{"highSeries.GetCurrent()"}, []string{"bar.High", "ctx.Data[ctx.BarIndex]"}},

		/* SecurityScope — OHLCV builtins series params */
		{"security close series", SecurityScope, "close", ParamTypeSeries, []string{"closeSeries"}, []string{"closeSeries.GetCurrent()", "ctx.LookupSeries"}},
		{"security open series", SecurityScope, "open", ParamTypeSeries, []string{"openSeries"}, []string{"openSeries.GetCurrent()", "ctx.LookupSeries"}},

		/* SecurityScope — other builtins */
		{"security bar_index", SecurityScope, "bar_index", ParamTypeScalar, []string{"float64(ctx.BarIndex)"}, []string{"float64(i)"}},
		{"security time", SecurityScope, "time", ParamTypeScalar, []string{"timeSeries.GetCurrent()"}, []string{"bar.Time", "ctx.Data[ctx.BarIndex]"}},

		/* ArrowScope — OHLCV builtins value params */
		{"arrow close value", ArrowScope, "close", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].Close"}, []string{"closeSeries.Get(0)", "bar.Close"}},
		{"arrow open value", ArrowScope, "open", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].Open"}, []string{"openSeries.Get(0)", "bar.Open"}},
		{"arrow high value", ArrowScope, "high", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].High"}, []string{"highSeries.Get(0)", "bar.High"}},
		{"arrow low value", ArrowScope, "low", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].Low"}, []string{"lowSeries.Get(0)", "bar.Low"}},
		{"arrow volume value", ArrowScope, "volume", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].Volume"}, []string{"volumeSeries.Get(0)", "bar.Volume"}},

		/* ArrowScope — OHLCV builtins series params */
		{"arrow close series", ArrowScope, "close", ParamTypeSeries, []string{"ctx.LookupSeries", "*series.Series", "closeSeries"}, []string{"bar.Close"}},
		{"arrow open series", ArrowScope, "open", ParamTypeSeries, []string{"ctx.LookupSeries", "*series.Series", "openSeries"}, []string{"bar.Open"}},
		{"arrow volume series", ArrowScope, "volume", ParamTypeSeries, []string{"ctx.LookupSeries", "*series.Series", "volumeSeries"}, []string{"bar.Volume"}},

		/* ArrowScope — other builtins */
		{"arrow bar_index", ArrowScope, "bar_index", ParamTypeScalar, []string{"float64(ctx.BarIndex)"}, []string{"float64(i)"}},
		{"arrow time", ArrowScope, "time", ParamTypeScalar, []string{"ctx.Data[ctx.BarIndex].Time"}, []string{"bar.Time", "timeSeries"}},
		{"arrow last_bar_index", ArrowScope, "last_bar_index", ParamTypeScalar, []string{"float64(len(ctx.Data) - 1)"}, []string{"last_bar_index\""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var aeg *ArgumentExpressionGenerator
			switch tt.scope {
			case BarLoopScope:
				aeg = newTestArgumentExpressionGenerator("testFunc", 0)
			case SecurityScope:
				gen := newTestGenerator()
				gen.inSecurityContext = true
				aeg = NewArgumentExpressionGenerator(gen, "testFunc", 0)
			case ArrowScope:
				aeg = newTestArrowArgumentExpressionGenerator("testFunc", 0)
			}

			if tt.paramType == ParamTypeSeries {
				aeg.signatureRegistry.Register("testFunc", []FunctionParameterType{ParamTypeSeries}, "float64")
			}

			result, err := aeg.Generate(&ast.Identifier{Name: tt.identifier})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("result %q missing expected substring %q", result, expected)
				}
			}

			for _, forbidden := range tt.mustNotContain {
				if strings.Contains(result, forbidden) {
					t.Errorf("result %q must not contain %q (scope leak)", result, forbidden)
				}
			}
		})
	}
}

func TestArgumentExpressionGenerator_ParameterTypeSpecialization(t *testing.T) {
	tests := []struct {
		name       string
		scope      AccessScope
		identifier string
		valueCode  string
		seriesCode string
	}{
		{
			name:       "bar-loop close: value Get(0) vs series bare",
			scope:      BarLoopScope,
			identifier: "close",
			valueCode:  "closeSeries.Get(0)",
			seriesCode: "closeSeries",
		},
		{
			name:       "security open: value GetCurrent vs series bare",
			scope:      SecurityScope,
			identifier: "open",
			valueCode:  "openSeries.GetCurrent()",
			seriesCode: "openSeries",
		},
		{
			name:       "arrow high: value ctx.Data vs series LookupSeries",
			scope:      ArrowScope,
			identifier: "high",
			valueCode:  "ctx.Data[ctx.BarIndex].High",
			seriesCode: "ctx.LookupSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var genValue, genSeries *ArgumentExpressionGenerator

			switch tt.scope {
			case BarLoopScope:
				genValue = newTestArgumentExpressionGenerator("funcValue", 0)
				genSeries = newTestArgumentExpressionGenerator("funcSeries", 0)
			case SecurityScope:
				gen := newTestGenerator()
				gen.inSecurityContext = true
				genValue = NewArgumentExpressionGenerator(gen, "funcValue", 0)
				genSeries = NewArgumentExpressionGenerator(gen, "funcSeries", 0)
			case ArrowScope:
				genValue = newTestArrowArgumentExpressionGenerator("funcValue", 0)
				genSeries = newTestArrowArgumentExpressionGenerator("funcSeries", 0)
			}

			genSeries.signatureRegistry.Register("funcSeries", []FunctionParameterType{ParamTypeSeries}, "float64")

			resultValue, err := genValue.Generate(&ast.Identifier{Name: tt.identifier})
			if err != nil {
				t.Fatalf("value param: unexpected error: %v", err)
			}

			resultSeries, err := genSeries.Generate(&ast.Identifier{Name: tt.identifier})
			if err != nil {
				t.Fatalf("series param: unexpected error: %v", err)
			}

			if !strings.Contains(resultValue, tt.valueCode) {
				t.Errorf("value param result %q missing expected %q", resultValue, tt.valueCode)
			}

			if !strings.Contains(resultSeries, tt.seriesCode) {
				t.Errorf("series param result %q missing expected %q", resultSeries, tt.seriesCode)
			}

			if resultValue == resultSeries {
				t.Errorf("value and series param generated identical code (no specialization): %q", resultValue)
			}
		})
	}
}

func TestArgumentExpressionGenerator_UserVariableResolution(t *testing.T) {
	tests := []struct {
		name       string
		scope      AccessScope
		varName    string
		paramType  FunctionParameterType
		wantValue  string
		wantSeries string
	}{
		{
			name:       "bar-loop user variable",
			scope:      BarLoopScope,
			varName:    "myVar",
			paramType:  ParamTypeScalar,
			wantValue:  "myVarSeries.GetCurrent()",
			wantSeries: "myVarSeries",
		},
		{
			name:       "security user variable",
			scope:      SecurityScope,
			varName:    "src",
			paramType:  ParamTypeScalar,
			wantValue:  "srcSeries.GetCurrent()",
			wantSeries: "srcSeries",
		},
		{
			name:       "arrow user variable",
			scope:      ArrowScope,
			varName:    "data",
			paramType:  ParamTypeScalar,
			wantValue:  "dataSeries.GetCurrent()",
			wantSeries: "dataSeries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var aeg *ArgumentExpressionGenerator
			switch tt.scope {
			case BarLoopScope:
				aeg = newTestArgumentExpressionGenerator("testFunc", 0)
			case SecurityScope:
				gen := newTestGenerator()
				gen.inSecurityContext = true
				aeg = NewArgumentExpressionGenerator(gen, "testFunc", 0)
			case ArrowScope:
				aeg = newTestArrowArgumentExpressionGenerator("testFunc", 0)
			}

			if tt.paramType == ParamTypeSeries {
				aeg.signatureRegistry.Register("testFunc", []FunctionParameterType{ParamTypeSeries}, "float64")
			}

			result, err := aeg.Generate(&ast.Identifier{Name: tt.varName})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedCode := tt.wantValue
			if tt.paramType == ParamTypeSeries {
				expectedCode = tt.wantSeries
			}

			if !strings.Contains(result, expectedCode) {
				t.Errorf("result %q missing expected %q", result, expectedCode)
			}
		})
	}
}

func TestArgumentExpressionGenerator_ConstantIdentifiers(t *testing.T) {
	tests := []struct {
		name       string
		scope      AccessScope
		constant   string
		wantCode   string
		setupConst func(*generator)
	}{
		{
			name:     "bar-loop input.source constant",
			scope:    BarLoopScope,
			constant: "src",
			wantCode: "srcSeries.GetCurrent()",
			setupConst: func(g *generator) {
				g.constants["src"] = "input.source"
			},
		},
		{
			name:     "arrow input.source constant",
			scope:    ArrowScope,
			constant: "mySource",
			wantCode: "mySourceSeries.GetCurrent()",
			setupConst: func(g *generator) {
				g.constants["mySource"] = "input.source"
			},
		},
		{
			name:     "security regular constant",
			scope:    SecurityScope,
			constant: "multiplier",
			wantCode: "multiplier",
			setupConst: func(g *generator) {
				g.constants["multiplier"] = 2.5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			if tt.scope == SecurityScope {
				gen.inSecurityContext = true
			} else if tt.scope == ArrowScope {
				gen.inArrowFunctionBody = true
			}
			tt.setupConst(gen)

			aeg := NewArgumentExpressionGenerator(gen, "testFunc", 0)
			result, err := aeg.Generate(&ast.Identifier{Name: tt.constant})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(result, tt.wantCode) {
				t.Errorf("result %q missing expected %q", result, tt.wantCode)
			}
		})
	}
}
