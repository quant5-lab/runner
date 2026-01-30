package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSecurityInlineHandler_CanHandle(t *testing.T) {
	handler := NewSecurityInlineHandler()

	tests := []struct {
		funcName string
		want     bool
	}{
		{"request.security", true},
		{"security", true},
		{"ta.sma", false},
		{"ta.security", false},
		{"request", false},
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

func TestSecurityInlineHandler_GenerateInline_ArgumentValidation(t *testing.T) {
	handler := NewSecurityInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name     string
		args     []ast.Expression
		wantIIFE bool
		wantNaN  bool
	}{
		{
			name:     "no arguments",
			args:     []ast.Expression{},
			wantIIFE: true,
			wantNaN:  true,
		},
		{
			name: "one argument only",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
			},
			wantIIFE: true,
			wantNaN:  true,
		},
		{
			name: "two arguments only",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
			},
			wantIIFE: true,
			wantNaN:  true,
		},
		{
			name: "invalid symbol type",
			args: []ast.Expression{
				&ast.Literal{Value: 123},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
			},
			wantIIFE: true,
			wantNaN:  true,
		},
		{
			name: "invalid timeframe type",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: 123},
				&ast.Identifier{Name: "close"},
			},
			wantIIFE: true,
			wantNaN:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "security"},
				Arguments: tt.args,
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			if tt.wantIIFE && !strings.Contains(result, "(func() float64 {") {
				t.Error("expected IIFE wrapper")
			}

			if tt.wantNaN && !strings.Contains(result, "math.NaN()") {
				t.Error("expected NaN return for invalid arguments")
			}
		})
	}
}

func TestSecurityInlineHandler_GenerateInline_OHLCVFields(t *testing.T) {
	handler := NewSecurityInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		field        string
		expectAccess string
	}{
		{"close", "secCtx.Data[secBarIdx].Close"},
		{"open", "secCtx.Data[secBarIdx].Open"},
		{"high", "secCtx.Data[secBarIdx].High"},
		{"low", "secCtx.Data[secBarIdx].Low"},
		{"volume", "secCtx.Data[secBarIdx].Volume"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "syminfo"},
						Property: &ast.Identifier{Name: "tickerid"},
					},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: tt.field},
				},
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			if !strings.Contains(result, "(func() float64 {") {
				t.Error("expected IIFE wrapper")
			}
			if !strings.Contains(result, "secCtx, secFound := securityContexts[secKey]") {
				t.Error("expected cache lookup")
			}
			if !strings.Contains(result, tt.expectAccess) {
				t.Errorf("expected %q, got result:\n%s", tt.expectAccess, result)
			}
			if !g.hasSecurityCalls {
				t.Error("expected hasSecurityCalls flag to be set")
			}
		})
	}
}

func TestSecurityInlineHandler_GenerateInline_ComplexExpressions(t *testing.T) {
	handler := NewSecurityInlineHandler()
	g := newTestGenerator()

	tests := []struct {
		name        string
		expression  ast.Expression
		mustContain []string
	}{
		{
			name: "TA call - sma",
			expression: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			mustContain: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.CallExpression",
				"secCtx, secBarIdx",
			},
		},
		{
			name: "Binary expression - comparison",
			expression: &ast.BinaryExpression{
				Operator: ">",
				Left:     &ast.Identifier{Name: "low"},
				Right:    &ast.Identifier{Name: "bb_upperBB"},
			},
			mustContain: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.BinaryExpression",
				"Operator: \">\"",
			},
		},
		{
			name: "Conditional expression - ternary",
			expression: &ast.ConditionalExpression{
				Test: &ast.BinaryExpression{
					Operator: ">",
					Left:     &ast.Identifier{Name: "close"},
					Right:    &ast.Identifier{Name: "open"},
				},
				Consequent: &ast.Identifier{Name: "close"},
				Alternate:  &ast.Identifier{Name: "open"},
			},
			mustContain: []string{
				"secBarEvaluator.EvaluateAtBar",
				"&ast.ConditionalExpression",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					tt.expression,
				},
			}

			result, err := handler.GenerateInline(call, g)
			if err != nil {
				t.Fatalf("GenerateInline failed: %v", err)
			}

			for _, substr := range tt.mustContain {
				if !strings.Contains(result, substr) {
					t.Errorf("expected substring %q in result:\n%s", substr, result)
				}
			}
		})
	}
}

func TestSecurityInlineHandler_GenerateOHLCVAccess(t *testing.T) {
	handler := NewSecurityInlineHandler()

	tests := []struct {
		field    string
		expected string
	}{
		{"close", "\t\treturn secCtx.Data[secBarIdx].Close\n"},
		{"open", "\t\treturn secCtx.Data[secBarIdx].Open\n"},
		{"high", "\t\treturn secCtx.Data[secBarIdx].High\n"},
		{"low", "\t\treturn secCtx.Data[secBarIdx].Low\n"},
		{"volume", "\t\treturn secCtx.Data[secBarIdx].Volume\n"},
		{"invalid", "\t\treturn math.NaN()\n"},
		{"", "\t\treturn math.NaN()\n"},
		{"Close", "\t\treturn math.NaN()\n"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := handler.generateOHLCVAccess(tt.field)
			if result != tt.expected {
				t.Errorf("generateOHLCVAccess(%q) = %q, want %q", tt.field, result, tt.expected)
			}
		})
	}
}

func TestSecurityInlineHandler_IIFEStructure(t *testing.T) {
	handler := NewSecurityInlineHandler()
	g := newTestGenerator()

	call := &ast.CallExpression{
		Callee: &ast.Identifier{Name: "security"},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
			&ast.Identifier{Name: "close"},
		},
	}

	result, err := handler.GenerateInline(call, g)
	if err != nil {
		t.Fatalf("GenerateInline failed: %v", err)
	}

	requiredElements := []string{
		"(func() float64 {",
		"}())",
		"secCtx, secFound := securityContexts[secKey]",
		"if !secFound { return math.NaN() }",
		"securityBarMapper, mapperFound := securityBarMappers[secKey]",
		"if !mapperFound { return math.NaN() }",
		"secLookahead :=",
		"secBarIdx := securityBarMapper.FindDailyBarIndex",
		"if secBarIdx < 0 { return math.NaN() }",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(result, elem) {
			t.Errorf("IIFE missing required element: %q", elem)
		}
	}

	/* Check for either constant key or fmt.Sprintf pattern */
	hasConstantKey := strings.Contains(result, `secKey := "`)
	hasSprintfKey := strings.Contains(result, "secKey := fmt.Sprintf(")
	if !hasConstantKey && !hasSprintfKey {
		t.Error("IIFE missing secKey assignment (expected constant string or fmt.Sprintf)")
	}

	if strings.HasPrefix(result, "(func() float64 {") && strings.HasSuffix(result, "}())") {
		// Valid IIFE structure
	} else {
		t.Error("IIFE structure invalid: should start with '(func() float64 {' and end with '}())'")
	}
}

func TestSecurityInlineHandler_ExtractLookahead(t *testing.T) {
	handler := NewSecurityInlineHandler()

	tests := []struct {
		name     string
		args     []ast.Expression
		expected bool
	}{
		{
			name: "no fourth argument",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
			},
			expected: false,
		},
		{
			name: "boolean literal true",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: true},
			},
			expected: true,
		},
		{
			name: "boolean literal false",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: false},
			},
			expected: false,
		},
		{
			name: "object with lookahead true",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "lookahead"},
							Value: &ast.Literal{Value: true},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "object with lookahead false",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "lookahead"},
							Value: &ast.Literal{Value: false},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "object without lookahead",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key:   &ast.Identifier{Name: "other"},
							Value: &ast.Literal{Value: true},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "identifier - cannot resolve",
			args: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1D"},
				&ast.Identifier{Name: "close"},
				&ast.Identifier{Name: "lookahead_var"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractLookahead(tt.args)
			if result != tt.expected {
				t.Errorf("extractLookahead() = %v, want %v", result, tt.expected)
			}
		})
	}
}
