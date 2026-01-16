package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

func TestHighestInlineHandler_CanHandle(t *testing.T) {
	handler := NewHighestInlineHandler()

	tests := []struct {
		funcName string
		want     bool
	}{
		{"ta.highest", true},
		{"highest", true},
		{"ta.lowest", false},
		{"lowest", false},
		{"ta.sma", false},
		{"high", false},
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

func TestHighestInlineHandler_GenerateInline_SingleArgument(t *testing.T) {
	handler := NewHighestInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		arg         ast.Expression
		wantPattern []string
	}{
		{
			name:        "literal period",
			arg:         &ast.Literal{Value: 5.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 4", "highest :=", "for j :=", "return highest"},
		},
		{
			name:        "single period",
			arg:         &ast.Literal{Value: 1.0},
			wantPattern: []string{"func() float64", "highest :=", "return highest"},
		},
		{
			name:        "large period",
			arg:         &ast.Literal{Value: 100.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 99", "highest :=", "for j := 99", "return highest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "highest"},
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

func TestHighestInlineHandler_GenerateInline_TwoArguments(t *testing.T) {
	handler := NewHighestInlineHandler()
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
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 9", "highest :=", "for j :=", "return highest"},
		},
		{
			name:        "high builtin source",
			sourceArg:   &ast.Identifier{Name: "high"},
			periodArg:   &ast.Literal{Value: 5.0},
			wantPattern: []string{"func() float64", "if ctx.BarIndex < 4", "highest :=", "return highest"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "ta.highest"},
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

func TestHighestInlineHandler_GenerateInline_InputVariable(t *testing.T) {
	handler := NewHighestInlineHandler()
	g := newTestGenerator()

	analyzer := validation.NewWarmupAnalyzer()
	analyzer.AddConstant("barsBack", 20.0)
	g.constEvaluator = analyzer
	g.constants["barsBack"] = 20.0

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "highest"},
		Arguments: []ast.Expression{&ast.Identifier{Name: "barsBack"}},
	}

	result, err := handler.GenerateInline(call, g)
	if err != nil {
		t.Fatalf("GenerateInline failed: %v", err)
	}

	wantPatterns := []string{"func() float64", "if ctx.BarIndex < 19", "highest :=", "return highest"}
	for _, pattern := range wantPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("result missing pattern %q\nGot: %s", pattern, result)
		}
	}
}

func TestHighestInlineHandler_GenerateInline_ErrorCases(t *testing.T) {
	handler := NewHighestInlineHandler()
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
				Callee:    &ast.Identifier{Name: "highest"},
				Arguments: tt.args,
			}

			_, err := handler.GenerateInline(call, g)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateInline error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHighestInlineHandler_GenerateInline_IIFEStructure(t *testing.T) {
	handler := NewHighestInlineHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "highest"},
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

	if !strings.Contains(result, "return highest") {
		t.Error("result should return highest variable")
	}
}

func TestHighestInlineHandler_GenerateInline_WarmupCheck(t *testing.T) {
	handler := NewHighestInlineHandler()
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
				Callee:    &ast.Identifier{Name: "highest"},
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

func TestHighestLowestSymmetry(t *testing.T) {
	lowestHandler := NewLowestInlineHandler()
	highestHandler := NewHighestInlineHandler()
	g := newTestGenerator()

	period := &ast.Literal{Value: 10.0}

	lowestCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "lowest"},
		Arguments: []ast.Expression{period},
	}

	highestCall := &ast.CallExpression{
		Callee:    &ast.Identifier{Name: "highest"},
		Arguments: []ast.Expression{period},
	}

	lowestResult, lowestErr := lowestHandler.GenerateInline(lowestCall, g)
	highestResult, highestErr := highestHandler.GenerateInline(highestCall, g)

	if lowestErr != nil {
		t.Fatalf("lowest GenerateInline failed: %v", lowestErr)
	}
	if highestErr != nil {
		t.Fatalf("highest GenerateInline failed: %v", highestErr)
	}

	sharedPatterns := []string{
		"func() float64",
		"if ctx.BarIndex < 9",
		"for j := 9",
		"}()",
	}

	for _, pattern := range sharedPatterns {
		if !strings.Contains(lowestResult, pattern) {
			t.Errorf("lowest missing shared pattern %q", pattern)
		}
		if !strings.Contains(highestResult, pattern) {
			t.Errorf("highest missing shared pattern %q", pattern)
		}
	}

	if strings.Contains(lowestResult, "highest") {
		t.Error("lowest result should not contain 'highest' variable name")
	}
	if strings.Contains(highestResult, "lowest") {
		t.Error("highest result should not contain 'lowest' variable name")
	}
}
