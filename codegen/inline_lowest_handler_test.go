package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestLowestInlineHandler_CanHandle(t *testing.T) {
	handler := NewLowestInlineHandler()

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.lowest", true},
		{"lowest", true},
		{"ta.highest", false},
		{"highest", false},
		{"ta.sma", false},
		{"low", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			if got := handler.CanHandle(tt.funcName); got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestLowestInlineHandler_GenerateInline_SingleArgument(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		arg         ast.Expression
		wantPattern []string
	}{
		{
			name:        "literal period",
			arg:         &ast.Literal{Value: 5.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 4", "lowest :=", "for j :=", "return lowest"},
		},
		{
			name:        "single period",
			arg:         &ast.Literal{Value: 1.0},
			wantPattern: []string{"func() float64", "lowest :=", "return lowest"},
		},
		{
			name:        "large period",
			arg:         &ast.Literal{Value: 100.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 99", "lowest :=", "for j := 99", "return lowest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "lowest"},
				Arguments: []ast.Expression{tt.arg},
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			for _, pattern := range tt.wantPattern {
				if !strings.Contains(result, pattern) {
					t.Errorf("result missing pattern %q\nGot: %s", pattern, result)
				}
			}
		})
	}
}

func TestLowestInlineHandler_GenerateInline_TwoArguments(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()
	g.variables["close"] = "series"

	tests := []struct {
		name        string
		sourceArg   ast.Expression
		periodArg   ast.Expression
		wantPattern []string
	}{
		{
			name:        "close source with literal period",
			sourceArg:   &ast.Identifier{Name: "close"},
			periodArg:   &ast.Literal{Value: 10.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 9", "lowest :=", "for j :=", "return lowest"},
		},
		{
			name:        "low builtin source",
			sourceArg:   &ast.Identifier{Name: "low"},
			periodArg:   &ast.Literal{Value: 5.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 4", "lowest :=", "return lowest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "ta.lowest"},
				Arguments: []ast.Expression{tt.sourceArg, tt.periodArg},
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			for _, pattern := range tt.wantPattern {
				if !strings.Contains(result, pattern) {
					t.Errorf("result missing pattern %q\nGot: %s", pattern, result)
				}
			}
		})
	}
}

func TestLowestInlineHandler_GenerateInline_InputVariable(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()

	analyzer := validation.NewWarmupAnalyzer()
	analyzer.AddConstant("barsBack", 20.0)
	g.constEvaluator = analyzer
	g.constants["barsBack"] = 20.0

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "lowest"},
		Arguments: []ast.Expression{&ast.Identifier{Name: "barsBack"}},
	}

	result, err := handler.GenerateInline(call, g)
	if err != nil {
		t.Fatalf("GenerateInline failed: %v", err)
	}

	wantPatterns := []string{"func() float64", "if ctx.BarIndex < 19", "lowest :=", "return lowest"}
	for _, pattern := range wantPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("result missing pattern %q\nGot: %s", pattern, result)
		}
	}
}

func TestLowestInlineHandler_GenerateInline_ErrorCases(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name    string
		args    []ast.Expression
		wantErr bool
	}{
		{
			name:    "no arguments",
			args:    []ast.Expression{},
			wantErr: true,
		},
		{
			name: "invalid period non-constant",
			args: []ast.Expression{
				&ast.BinaryExpression{
					Operator: "+",
					Left:     &ast.Identifier{Name: "x"},
					Right:    &ast.Identifier{Name: "y"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "lowest"},
				Arguments: tt.args,
			}

			_, err := handler.GenerateInline(call, g)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateInline error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLowestInlineHandler_GenerateInline_IIFEStructure(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "lowest"},
		Arguments: []ast.Expression{&ast.Literal{Value: 3.0}},
	}

	result, err := handler.GenerateInline(call, g)
	if err != nil {
		t.Fatalf("GenerateInline failed: %v", err)
	}

	if !strings.HasPrefix(result, "func() float64 {") {
		t.Errorf("result should start with IIFE declaration, got: %s", result[:30])
	}

	if !strings.HasSuffix(result, "}()") {
		t.Errorf("result should end with IIFE invocation, got: %s", result[len(result)-10:])
	}

	if !strings.Contains(result, "return lowest") {
		t.Error("result should return lowest variable")
	}
}

func TestLowestInlineHandler_GenerateInline_WarmupCheck(t *testing.T) {
	handler := NewLowestInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name       string
		period     float64
		wantWarmup int
	}{
		{"period 1 no warmup", 1.0, 0},
		{"period 2 warmup 1", 2.0, 1},
		{"period 10 warmup 9", 10.0, 9},
		{"period 50 warmup 49", 50.0, 49},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "lowest"},
				Arguments: []ast.Expression{&ast.Literal{Value: tt.period}},
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			if tt.wantWarmup > 0 {
				expectedCheck := "if ctx.BarIndex <"
				if !strings.Contains(result, expectedCheck) {
					t.Errorf("expected warmup check for period %v, got: %s", tt.period, result)
				}
			}
		})
	}
}
