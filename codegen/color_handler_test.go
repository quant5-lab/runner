package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestColorHandler_CanHandle(t *testing.T) {
	ch := NewColorHandler()

	handled := []string{"color.new", "color.rgb", "color.from_gradient", "color.r", "color.g", "color.b", "color.t"}
	for _, fn := range handled {
		if !ch.CanHandle(fn) {
			t.Errorf("CanHandle(%q) = false, want true", fn)
		}
	}

	notHandled := []string{"color", "color.unknown", "math.abs", "ta.sma", "plot"}
	for _, fn := range notHandled {
		if ch.CanHandle(fn) {
			t.Errorf("CanHandle(%q) = true, want false", fn)
		}
	}
}

func TestColorHandler_GenerateColorNew(t *testing.T) {
	ch := NewColorHandler()
	g := newTestGenerator()

	tests := []struct {
		name     string
		args     []ast.Expression
		contains []string
		wantErr  bool
	}{
		{
			name: "color.new with constant and literal transp",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
				&ast.Literal{Value: float64(50)},
			},
			contains: []string{"visual.PineColorNew", "#FF5252", "50"},
		},
		{
			name: "color.new with hex literal",
			args: []ast.Expression{
				&ast.Literal{Value: "#00BCD4"},
				&ast.Literal{Value: float64(0)},
			},
			contains: []string{"visual.PineColorNew", "#00BCD4"},
		},
		{
			name: "color.new with bare color identifier",
			args: []ast.Expression{
				&ast.Identifier{Name: "blue"},
				&ast.Literal{Value: float64(70)},
			},
			contains: []string{"visual.PineColorNew", "#2962FF"},
		},
		{
			name:    "color.new with no args errors",
			args:    []ast.Expression{},
			wantErr: true,
		},
		{
			name: "color.new with 1 arg defaults transp to 0",
			args: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "green"},
				},
			},
			contains: []string{"visual.PineColorNew", "#4CAF50", "0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := ch.GenerateColorCall("color.new", tt.args, g)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			for _, s := range tt.contains {
				if !strings.Contains(code, s) {
					t.Errorf("Generated code %q missing expected substring %q", code, s)
				}
			}
		})
	}
}

func TestColorHandler_GenerateColorRGB(t *testing.T) {
	ch := NewColorHandler()
	g := newTestGenerator()

	code, err := ch.GenerateColorCall("color.rgb", []ast.Expression{
		&ast.Literal{Value: float64(255)},
		&ast.Literal{Value: float64(128)},
		&ast.Literal{Value: float64(0)},
		&ast.Literal{Value: float64(30)},
	}, g)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(code, "visual.PineColorRGB") {
		t.Errorf("Expected visual.PineColorRGB in %q", code)
	}

	/* 3-arg form (transp defaults to 0) */
	code, err = ch.GenerateColorCall("color.rgb", []ast.Expression{
		&ast.Literal{Value: float64(255)},
		&ast.Literal{Value: float64(0)},
		&ast.Literal{Value: float64(0)},
	}, g)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(code, "0.0") {
		t.Errorf("Expected default transp 0.0 in %q", code)
	}
}

func TestColorHandler_GenerateColorComponents(t *testing.T) {
	ch := NewColorHandler()
	g := newTestGenerator()

	components := []struct {
		funcName string
		goFunc   string
	}{
		{"color.r", "visual.PineColorR"},
		{"color.g", "visual.PineColorG"},
		{"color.b", "visual.PineColorB"},
		{"color.t", "visual.PineColorT"},
	}

	for _, cc := range components {
		t.Run(cc.funcName, func(t *testing.T) {
			code, err := ch.GenerateColorCall(cc.funcName, []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
			}, g)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !strings.Contains(code, cc.goFunc) {
				t.Errorf("Expected %s in %q", cc.goFunc, code)
			}
			if !strings.Contains(code, "#FF5252") {
				t.Errorf("Expected resolved hex in %q", code)
			}
		})
	}
}

func TestColorHandler_GenerateColorFromGradient(t *testing.T) {
	ch := NewColorHandler()
	g := newTestGenerator()

	code, err := ch.GenerateColorCall("color.from_gradient", []ast.Expression{
		&ast.Identifier{Name: "rsi"},
		&ast.Literal{Value: float64(0)},
		&ast.Literal{Value: float64(100)},
		&ast.MemberExpression{
			Object:   &ast.Identifier{Name: "color"},
			Property: &ast.Identifier{Name: "red"},
		},
		&ast.MemberExpression{
			Object:   &ast.Identifier{Name: "color"},
			Property: &ast.Identifier{Name: "green"},
		},
	}, g)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(code, "visual.PineColorFromGradient") {
		t.Errorf("Expected visual.PineColorFromGradient in %q", code)
	}
	if !strings.Contains(code, "#FF5252") {
		t.Errorf("Expected resolved red hex in %q", code)
	}
	if !strings.Contains(code, "#4CAF50") {
		t.Errorf("Expected resolved green hex in %q", code)
	}
}

func TestColorHandler_NestedColorCall(t *testing.T) {
	ch := NewColorHandler()
	g := newTestGenerator()

	/* color.r(color.new(color.red, 50)) — nested call as argument */
	code, err := ch.GenerateColorCall("color.r", []ast.Expression{
		&ast.CallExpression{
			Callee: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "color"},
				Property: &ast.Identifier{Name: "new"},
			},
			Arguments: []ast.Expression{
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "color"},
					Property: &ast.Identifier{Name: "red"},
				},
				&ast.Literal{Value: float64(50)},
			},
		},
	}, g)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(code, "visual.PineColorR") {
		t.Errorf("Expected visual.PineColorR in %q", code)
	}
	if !strings.Contains(code, "visual.PineColorNew") {
		t.Errorf("Expected nested visual.PineColorNew in %q", code)
	}
}
