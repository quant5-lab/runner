package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestMathHandler_Comprehensive(t *testing.T) {
	mh := NewMathHandler()
	g := &generator{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}

	allFunctions := []struct {
		name string
		args []ast.Expression
	}{
		{"abs", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"sqrt", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"floor", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"ceil", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"round", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"log", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"log10", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"exp", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"sin", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"cos", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"tan", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"asin", []ast.Expression{&ast.Literal{Value: 0.5}}},
		{"acos", []ast.Expression{&ast.Literal{Value: 0.5}}},
		{"atan", []ast.Expression{&ast.Literal{Value: 0.5}}},
		{"sign", []ast.Expression{&ast.Literal{Value: 5.0}}},
		{"todegrees", []ast.Expression{&ast.Literal{Value: 3.14159}}},
		{"toradians", []ast.Expression{&ast.Literal{Value: 180.0}}},
		{"round_to_mintick", []ast.Expression{&ast.Literal{Value: 1.2345}}},
		{"max", []ast.Expression{&ast.Literal{Value: 2.0}, &ast.Literal{Value: 3.0}}},
		{"min", []ast.Expression{&ast.Literal{Value: 2.0}, &ast.Literal{Value: 3.0}}},
		{"pow", []ast.Expression{&ast.Literal{Value: 2.0}, &ast.Literal{Value: 3.0}}},
		{"avg", []ast.Expression{&ast.Literal{Value: 1.0}, &ast.Literal{Value: 2.0}}},
		{"random", []ast.Expression{}},
	}

	for _, tc := range allFunctions {
		t.Run(tc.name, func(t *testing.T) {
			unprefixed := tc.name
			prefixed := "math." + tc.name

			if !mh.CanHandle(unprefixed) {
				t.Errorf("CanHandle(%q) = false", unprefixed)
			}
			if !mh.CanHandle(prefixed) {
				t.Errorf("CanHandle(%q) = false", prefixed)
			}

			code1, err1 := mh.GenerateMathCall(unprefixed, tc.args, g)
			code2, err2 := mh.GenerateMathCall(prefixed, tc.args, g)

			if err1 != nil {
				t.Errorf("unprefixed %q failed: %v", unprefixed, err1)
			}
			if err2 != nil {
				t.Errorf("prefixed %q failed: %v", prefixed, err2)
			}
			if code1 != code2 {
				t.Errorf("mismatch: %q=%s, %q=%s", unprefixed, code1, prefixed, code2)
			}
		})
	}
}
