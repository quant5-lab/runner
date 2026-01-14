package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSecurityCallEmitter_LookaheadParameter(t *testing.T) {
	tests := []struct {
		name              string
		arguments         []ast.Expression
		expectedLookahead bool
	}{
		{
			name: "lookahead=true literal",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: true},
			},
			expectedLookahead: true,
		},
		{
			name: "lookahead=false literal",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.Literal{Value: false},
			},
			expectedLookahead: false,
		},
		{
			name: "lookahead=barmerge.lookahead_on constant",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "barmerge"},
					Property: &ast.Identifier{Name: "lookahead_on"},
					Computed: false,
				},
			},
			expectedLookahead: true,
		},
		{
			name: "lookahead=barmerge.lookahead_off constant",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.MemberExpression{
					Object:   &ast.Identifier{Name: "barmerge"},
					Property: &ast.Identifier{Name: "lookahead_off"},
					Computed: false,
				},
			},
			expectedLookahead: false,
		},
		{
			name: "named parameter lookahead=true",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
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
			expectedLookahead: true,
		},
		{
			name: "named parameter lookahead=barmerge.lookahead_on",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key: &ast.Identifier{Name: "lookahead"},
							Value: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "barmerge"},
								Property: &ast.Identifier{Name: "lookahead_on"},
								Computed: false,
							},
						},
					},
				},
			},
			expectedLookahead: true,
		},
		{
			name: "named parameter lookahead=barmerge.lookahead_off",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
				&ast.ObjectExpression{
					Properties: []ast.Property{
						{
							Key: &ast.Identifier{Name: "lookahead"},
							Value: &ast.MemberExpression{
								Object:   &ast.Identifier{Name: "barmerge"},
								Property: &ast.Identifier{Name: "lookahead_off"},
								Computed: false,
							},
						},
					},
				},
			},
			expectedLookahead: false,
		},
		{
			name: "no lookahead parameter defaults to false",
			arguments: []ast.Expression{
				&ast.Literal{Value: "BTCUSDT"},
				&ast.Literal{Value: "1h"},
				&ast.Identifier{Name: "close"},
			},
			expectedLookahead: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &generator{}
			emitter := NewSecurityCallEmitter(gen)

			callExpr := &ast.CallExpression{
				Callee:    &ast.MemberExpression{Object: &ast.Identifier{Name: "request"}, Property: &ast.Identifier{Name: "security"}},
				Arguments: tt.arguments,
			}

			code, err := emitter.EmitSecurityCall("testVar", callExpr)
			if err != nil {
				t.Fatalf("EmitSecurityCall failed: %v", err)
			}

			if !strings.Contains(code, "secLookahead := ") {
				t.Errorf("Expected secLookahead variable declaration in generated code, got:\n%s", code)
			}

			expectedInitValue := "secLookahead := false"
			if tt.expectedLookahead {
				expectedInitValue = "secLookahead := true"
			}
			if !strings.Contains(code, expectedInitValue) {
				t.Errorf("Expected %s in generated code, got:\n%s", expectedInitValue, code)
			}

			if !strings.Contains(code, "if \"1h\" == ctx.Timeframe {") {
				t.Errorf("Expected runtime same-timeframe detection in generated code, got:\n%s", code)
			}

			if !strings.Contains(code, "securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)") {
				t.Errorf("Expected FindDailyBarIndex call with secLookahead variable in generated code, got:\n%s", code)
			}
		})
	}
}
