package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestArrowSecurityCallGenerator_CanHandle(t *testing.T) {
	tests := []struct {
		name     string
		call     *ast.CallExpression
		expected bool
	}{
		{
			name: "request_dot_security",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
			},
			expected: true,
		},
		{
			name: "legacy_security",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "security"},
			},
			expected: true,
		},
		{
			name: "ta_sma",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "ta"},
					Property: &ast.Identifier{Name: "sma"},
				},
			},
			expected: false,
		},
		{
			name: "plain_identifier",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "myFunc"},
			},
			expected: false,
		},
		{
			name: "request_dot_other",
			call: &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "seed"},
				},
			},
			expected: false,
		},
	}

	gen := &ArrowSecurityCallGenerator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.CanHandle(tt.call)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestArrowSecurityCallGenerator_Generate_OHLCVFields(t *testing.T) {
	fields := []string{"close", "open", "high", "low", "volume"}

	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			g := newTestGenerator()
			arrowSecGen := NewArrowSecurityCallGenerator(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: field},
				},
			}

			code, err := arrowSecGen.Generate(call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expectedField := strings.Title(field)
			if field == "volume" {
				expectedField = "Volume"
			}
			if !strings.Contains(code, "secCtx.Data[secBarIdx]."+expectedField) {
				t.Errorf("expected OHLCV field access for %s in output.\nGot:\n%s", field, code)
			}
		})
	}
}

func TestArrowSecurityCallGenerator_Generate_DerivedPriceFields(t *testing.T) {
	derivedFields := map[string]string{
		"ohlc4": "/ 4",
		"hlc3":  "/ 3",
		"hl2":   "/ 2",
		"hlcc4": "/ 4",
	}

	for field, divisor := range derivedFields {
		t.Run(field, func(t *testing.T) {
			g := newTestGenerator()
			arrowSecGen := NewArrowSecurityCallGenerator(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: []ast.Expression{
					&ast.Literal{Value: "BTCUSDT"},
					&ast.Literal{Value: "1D"},
					&ast.Identifier{Name: field},
				},
			}

			code, err := arrowSecGen.Generate(call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, divisor) {
				t.Errorf("expected derived price formula with %q for %s in output.\nGot:\n%s", divisor, field, code)
			}
			if !strings.Contains(code, "secCtx.Data[secBarIdx]") {
				t.Errorf("expected bar data access for %s.\nGot:\n%s", field, code)
			}
		})
	}
}

func TestArrowSecurityCallGenerator_Generate_ScopeIsolation(t *testing.T) {
	g := newTestGenerator()
	arrowSecGen := NewArrowSecurityCallGenerator(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
			&ast.Identifier{Name: "close"},
		},
	}

	code, err := arrowSecGen.Generate(call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requiredPatterns := []string{
		"arrowCtx.SecurityContexts[secKey]",
		"arrowCtx.SecurityBarMappers[secKey]",
		"secBarMapper.FindDailyBarIndex",
		"func() float64",
	}
	for _, p := range requiredPatterns {
		if !strings.Contains(code, p) {
			t.Errorf("missing required pattern %q.\nGot:\n%s", p, code)
		}
	}

	trimmed := strings.ReplaceAll(code, "arrowCtx.SecurityContexts", "REPLACED")
	trimmed = strings.ReplaceAll(trimmed, "arrowCtx.SecurityBarMappers", "REPLACED")
	if strings.Contains(trimmed, "securityContexts[") || strings.Contains(trimmed, "securityBarMappers[") {
		t.Errorf("IIFE should NOT reference direct scope variables.\nGot:\n%s", code)
	}
}

func TestArrowSecurityCallGenerator_Generate_InsufficientArgs(t *testing.T) {
	tests := []struct {
		name string
		args []ast.Expression
	}{
		{"zero_args", []ast.Expression{}},
		{"one_arg", []ast.Expression{&ast.Literal{Value: "BTCUSDT"}}},
		{"two_args", []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			arrowSecGen := NewArrowSecurityCallGenerator(g)

			call := &ast.CallExpression{
				Callee: &ast.MemberExpression{
					Object:   &ast.Identifier{Name: "request"},
					Property: &ast.Identifier{Name: "security"},
				},
				Arguments: tt.args,
			}

			code, err := arrowSecGen.Generate(call)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(code, "math.NaN()") {
				t.Errorf("expected NaN fallback for insufficient args, got %q", code)
			}
		})
	}
}

func TestArrowSecurityCallGenerator_Generate_SetsSecurityFlag(t *testing.T) {
	g := newTestGenerator()
	arrowSecGen := NewArrowSecurityCallGenerator(g)

	if g.hasSecurityCalls {
		t.Fatal("hasSecurityCalls should be false before generation")
	}

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
			&ast.Identifier{Name: "close"},
		},
	}

	_, err := arrowSecGen.Generate(call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !g.hasSecurityCalls {
		t.Error("hasSecurityCalls should be true after generation")
	}
}

func TestArrowSecurityCallGenerator_Generate_NaNGuards(t *testing.T) {
	g := newTestGenerator()
	arrowSecGen := NewArrowSecurityCallGenerator(g)

	call := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "request"},
			Property: &ast.Identifier{Name: "security"},
		},
		Arguments: []ast.Expression{
			&ast.Literal{Value: "BTCUSDT"},
			&ast.Literal{Value: "1D"},
			&ast.Identifier{Name: "close"},
		},
	}

	code, err := arrowSecGen.Generate(call)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	/* Must guard against missing context and missing mapper */
	guards := []string{
		"!secFound",
		"return math.NaN()",
		"!mapperFound",
		"secBarIdx < 0",
	}
	for _, guard := range guards {
		if !strings.Contains(code, guard) {
			t.Errorf("missing NaN guard %q.\nGot:\n%s", guard, code)
		}
	}
}
