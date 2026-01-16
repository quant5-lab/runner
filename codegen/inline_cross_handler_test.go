package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestCrossInlineHandler_ExpressionTypes(t *testing.T) {
	tests := []struct {
		name    string
		isUnder bool
		args    []ast.Expression
		wantErr bool
		checks  []string
	}{
		{
			name:    "two Identifiers",
			isUnder: false,
			args: []ast.Expression{
				&ast.Identifier{Name: "tenkanSen"},
				&ast.Identifier{Name: "kijunSen"},
			},
			checks: []string{
				"tenkanSenSeries.Get(0)",
				"kijunSenSeries.Get(0)",
				"curr1 > curr2",
				"prev1 <= prev2",
			},
		},
		{
			name:    "Identifier and MemberExpression",
			isUnder: false,
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "close"},
					Property: &ast.Literal{Value: 0},
					Computed: true,
				},
				&ast.Identifier{Name: "sma20"},
			},
			checks: []string{
				"bar.Close",
				"sma20Series.Get(0)",
			},
		},
		{
			name:    "two MemberExpressions",
			isUnder: false,
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "high"},
					Property: &ast.Literal{Value: 0},
					Computed: true,
				},
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "low"},
					Property: &ast.Literal{Value: 0},
					Computed: true,
				},
			},
			checks: []string{
				"bar.High",
				"bar.Low",
			},
		},
		{
			name:    "BinaryExpression in argument",
			isUnder: false,
			args: []ast.Expression{
				&ast.Identifier{Name: "close"},
				&ast.BinaryExpression{
					Operator: "*",
					Left:     &ast.Identifier{Name: "sma20"},
					Right:    &ast.Literal{Value: 1.02},
				},
			},
			checks: []string{
				"closeSeries.Get(0)",
				"sma20Series.GetCurrent()",
				"* 1.02",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CrossInlineHandler{isUnder: tt.isUnder}
			gen := newTestGeneratorWithPlotHandler()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: tt.args,
			}

			code, err := handler.GenerateInline(call, gen)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, check := range tt.checks {
				if !strings.Contains(code, check) {
					t.Errorf("missing %q in generated code:\n%s", check, code)
				}
			}
		})
	}
}

func TestCrossInlineHandler_OperatorLogic(t *testing.T) {
	tests := []struct {
		name                string
		isUnder             bool
		expectedCurrOp      string
		expectedPrevOp      string
		expectedDescription string
	}{
		{
			name:                "crossover operators",
			isUnder:             false,
			expectedCurrOp:      "curr1 > curr2",
			expectedPrevOp:      "prev1 <= prev2",
			expectedDescription: "crosses above",
		},
		{
			name:                "crossunder operators",
			isUnder:             true,
			expectedCurrOp:      "curr1 < curr2",
			expectedPrevOp:      "prev1 >= prev2",
			expectedDescription: "crosses below",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CrossInlineHandler{isUnder: tt.isUnder}
			gen := newTestGeneratorWithPlotHandler()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "ema_fast"},
					&ast.Identifier{Name: "ema_slow"},
				},
			}

			code, err := handler.GenerateInline(call, gen)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, tt.expectedCurrOp) {
				t.Errorf("missing current operator %q in code:\n%s", tt.expectedCurrOp, code)
			}

			if !strings.Contains(code, tt.expectedPrevOp) {
				t.Errorf("missing previous operator %q in code:\n%s", tt.expectedPrevOp, code)
			}

			if !strings.Contains(code, "ctx.BarIndex == 0") {
				t.Error("missing warmup check: ctx.BarIndex == 0")
			}

			if !strings.Contains(code, "ctx.BarIndex--") {
				t.Error("missing bar index decrement for previous value access")
			}

			if !strings.Contains(code, "ctx.BarIndex = prevBarIdx") {
				t.Error("missing bar index restoration")
			}
		})
	}
}

func TestCrossInlineHandler_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []ast.Expression
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "no arguments",
			args:        []ast.Expression{},
			wantErr:     true,
			expectedMsg: "requires 2 arguments",
		},
		{
			name: "single argument",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
			},
			wantErr:     true,
			expectedMsg: "requires 2 arguments",
		},
		{
			name: "three arguments",
			args: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
				&ast.Identifier{Name: "c"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CrossInlineHandler{isUnder: false}
			gen := newTestGeneratorWithPlotHandler()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "crossover"},
				},
				Arguments: tt.args,
			}

			_, err := handler.GenerateInline(call, gen)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.expectedMsg != "" && !strings.Contains(err.Error(), tt.expectedMsg) {
					t.Errorf("expected error containing %q, got %q", tt.expectedMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCrossInlineHandler_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		handler  *CrossInlineHandler
		funcName string
		want     bool
	}{
		{"crossover handler with ta.crossover", NewCrossoverInlineHandler(), "ta.crossover", true},
		{"crossover handler with crossover", NewCrossoverInlineHandler(), "crossover", true},
		{"crossover handler with ta.crossunder", NewCrossoverInlineHandler(), "ta.crossunder", false},
		{"crossover handler with crossunder", NewCrossoverInlineHandler(), "crossunder", false},
		{"crossover handler with ta.sma", NewCrossoverInlineHandler(), "ta.sma", false},
		{"crossover handler with empty", NewCrossoverInlineHandler(), "", false},
		{"crossunder handler with ta.crossunder", NewCrossunderInlineHandler(), "ta.crossunder", true},
		{"crossunder handler with crossunder", NewCrossunderInlineHandler(), "crossunder", true},
		{"crossunder handler with ta.crossover", NewCrossunderInlineHandler(), "ta.crossover", false},
		{"crossunder handler with crossover", NewCrossunderInlineHandler(), "crossover", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.handler.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestCrossInlineHandler_IIFEStructure(t *testing.T) {
	handler := &CrossInlineHandler{isUnder: false}
	gen := newTestGeneratorWithPlotHandler()

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "crossover"},
		},
		Arguments: []ast.Expression{
			&ast.Identifier{Name: "a"},
			&ast.Identifier{Name: "b"},
		},
	}

	code, err := handler.GenerateInline(call, gen)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requiredPatterns := []string{
		"(func() bool {",
		"}())",
		"if ctx.BarIndex == 0 { return false }",
		"curr1 :=",
		"curr2 :=",
		"prevBarIdx := ctx.BarIndex",
		"ctx.BarIndex--",
		"prev1 :=",
		"prev2 :=",
		"ctx.BarIndex = prevBarIdx",
		"return ",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(code, pattern) {
			t.Errorf("missing required pattern %q in IIFE:\n%s", pattern, code)
		}
	}
}

func TestCrossInlineHandler_BothDirections(t *testing.T) {
	args := []ast.Expression{
		&ast.Identifier{Name: "fast"},
		&ast.Identifier{Name: "slow"},
	}

	tests := []struct {
		name    string
		isUnder bool
	}{
		{"crossover", false},
		{"crossunder", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &CrossInlineHandler{isUnder: tt.isUnder}
			gen := newTestGeneratorWithPlotHandler()

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: tt.name},
				},
				Arguments: args,
			}

			code, err := handler.GenerateInline(call, gen)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if code == "" {
				t.Error("expected non-empty code")
			}

			if !strings.Contains(code, "func() bool") {
				t.Error("expected boolean return type")
			}
		})
	}
}

func newTestGeneratorWithPlotHandler() *generator {
	gen := &generator{
		imports:        make(map[string]bool),
		variables:      make(map[string]string),
		strategyConfig: NewStrategyConfig(),
		taRegistry:     NewTAFunctionRegistry(),
		builtinHandler: NewBuiltinIdentifierHandler(),
	}
	gen.plotExprHandler = NewPlotExpressionHandler(gen)
	return gen
}
