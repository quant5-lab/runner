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

			expectedFunctionCall := "FindBarIndexByTimestamp"
			if tt.expectedLookahead {
				expectedFunctionCall = "FindBarIndexByTimestampWithLookahead"
			}

			if !strings.Contains(code, expectedFunctionCall) {
				t.Errorf("Expected %s in generated code, got:\n%s", expectedFunctionCall, code)
			}
		})
	}
}
