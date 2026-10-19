package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestVoidBuiltinHandler_CanHandle(t *testing.T) {
	h := NewVoidBuiltinHandler()

	tests := []struct {
		funcName string
		want     bool
	}{
		{"alert", true},
		{"alertcondition", true},
		{"ta.sma", false},
		{"strategy.entry", false},
		{"plot", false},
		{"math.abs", false},
		{"str.lower", false},
		{"", false},
		{"Alert", false},           // case-sensitive
		{"ALERTCONDITION", false},  // case-sensitive
		{"alert_custom", false},    // prefix must not match
		{"alertcondition2", false}, // suffix must not match
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			got := h.CanHandle(tt.funcName)
			if got != tt.want {
				t.Errorf("CanHandle(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestVoidBuiltinHandler_GenerateCode(t *testing.T) {
	h := NewVoidBuiltinHandler()
	g := newTestGenerator()

	tests := []struct {
		name string
		call *ast.CallExpression
	}{
		{
			name: "alert with message and freq member expression",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "alert"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "Long signal"},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "alert"},
						Property: &ast.Identifier{Name: "freq_once_per_bar_close"},
					},
				},
			},
		},
		{
			name: "alert with message only",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "alert"},
				Arguments: []ast.Expression{&ast.Literal{Value: "signal text"}},
			},
		},
		{
			name: "alert with no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "alert"},
				Arguments: []ast.Expression{},
			},
		},
		{
			name: "alert with complex string-concatenation argument",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "alert"},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Operator: "+",
						Left:     &ast.Identifier{Name: "syminfo.tickerid"},
						Right:    &ast.Literal{Value: " signal"},
					},
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "alert"},
						Property: &ast.Identifier{Name: "freq_once_per_bar_close"},
					},
				},
			},
		},
		{
			name: "alertcondition with boolean condition title and message",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "alertcondition"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: true},
					&ast.Literal{Value: "Cross Alert"},
					&ast.Literal{Value: "Moving Avg Crossing!"},
				},
			},
		},
		{
			name: "alertcondition with binary expression condition",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "alertcondition"},
				Arguments: []ast.Expression{
					&ast.BinaryExpression{
						Operator: ">",
						Left:     &ast.Identifier{Name: "close"},
						Right:    &ast.Identifier{Name: "open"},
					},
				},
			},
		},
		{
			name: "alertcondition with no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "alertcondition"},
				Arguments: []ast.Expression{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := h.GenerateCode(g, tt.call)
			if err != nil {
				t.Errorf("GenerateCode() unexpected error: %v", err)
			}
			if code != "" {
				t.Errorf("GenerateCode() = %q, want empty string", code)
			}
		})
	}
}

func TestVoidBuiltinHandler_GeneratorStateUnchanged(t *testing.T) {
	h := NewVoidBuiltinHandler()

	calls := []*ast.CallExpression{
		{Callee: &ast.Identifier{Name: "alert"}, Arguments: []ast.Expression{&ast.Literal{Value: "msg"}}},
		{Callee: &ast.Identifier{Name: "alertcondition"}, Arguments: []ast.Expression{&ast.Literal{Value: true}}},
	}

	for _, call := range calls {
		t.Run(extractCallFunctionName(call), func(t *testing.T) {
			g := newTestGenerator()
			before := len(g.featureGaps)

			_, _ = h.GenerateCode(g, call)

			if len(g.featureGaps) != before {
				t.Errorf("GenerateCode() modified featureGaps: before=%d after=%d gaps=%v",
					before, len(g.featureGaps), g.featureGaps)
			}
		})
	}
}
