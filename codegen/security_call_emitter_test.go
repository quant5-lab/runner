package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func TestSecurityCallEmitter_ExtractTimeframeCode_Literals(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "daily timeframe literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "1D",
			},
			expected: `"1D"`,
			wantErr:  false,
		},
		{
			name: "hourly timeframe literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "1h",
			},
			expected: `"1h"`,
			wantErr:  false,
		},
		{
			name: "5 minute timeframe literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "5m",
			},
			expected: `"5m"`,
			wantErr:  false,
		},
		{
			name: "weekly timeframe literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "1W",
			},
			expected: `"1W"`,
			wantErr:  false,
		},
		{
			name: "monthly timeframe literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "1M",
			},
			expected: `"1M"`,
			wantErr:  false,
		},
		{
			name: "empty string literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    "",
			},
			expected: `""`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := emitter.extractTimeframeCode(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimeframeCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("extractTimeframeCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSecurityCallEmitter_ExtractTimeframeCode_RuntimeResolution(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "timeframe.period MemberExpression",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "timeframe",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "period",
				},
			},
			expected: "ctx.Timeframe",
			wantErr:  false,
		},
		{
			name: "timeframe.multiplier MemberExpression - not runtime",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "timeframe",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "multiplier",
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "syminfo.tickerid MemberExpression - wrong namespace",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "syminfo",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "tickerid",
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "strategy.period MemberExpression - wrong namespace",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "strategy",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "period",
				},
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := emitter.extractTimeframeCode(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimeframeCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("extractTimeframeCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSecurityCallEmitter_ExtractTimeframeCode_InvalidExpressions(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	tests := []struct {
		name    string
		expr    ast.Expression
		wantErr bool
	}{
		{
			name: "Identifier not supported",
			expr: &ast.Identifier{
				NodeType: ast.TypeIdentifier,
				Name:     "myTimeframe",
			},
			wantErr: true,
		},
		{
			name: "CallExpression not supported",
			expr: &ast.CallExpression{
				NodeType: ast.TypeCallExpression,
				Callee: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "getTimeframe",
				},
				Arguments: []ast.Expression{},
			},
			wantErr: true,
		},
		{
			name: "BinaryExpression not supported",
			expr: &ast.BinaryExpression{
				NodeType: ast.TypeBinaryExpression,
				Left: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    "1",
				},
				Operator: "+",
				Right: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    "D",
				},
			},
			wantErr: true,
		},
		{
			name: "numeric literal not supported",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    60,
			},
			wantErr: true,
		},
		{
			name: "boolean literal not supported",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := emitter.extractTimeframeCode(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimeframeCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecurityCallEmitter_ExtractTimeframeCode_EdgeCases(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "MemberExpression with non-Identifier property",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "timeframe",
				},
				Property: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    "period",
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "MemberExpression with non-Identifier object",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.Literal{
					NodeType: ast.TypeLiteral,
					Value:    "timeframe",
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "period",
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "nested MemberExpression",
			expr: &ast.MemberExpression{
				NodeType: ast.TypeMemberExpression,
				Object: &ast.MemberExpression{
					NodeType: ast.TypeMemberExpression,
					Object: &ast.Identifier{
						NodeType: ast.TypeIdentifier,
						Name:     "ctx",
					},
					Property: &ast.Identifier{
						NodeType: ast.TypeIdentifier,
						Name:     "timeframe",
					},
				},
				Property: &ast.Identifier{
					NodeType: ast.TypeIdentifier,
					Name:     "period",
				},
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := emitter.extractTimeframeCode(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimeframeCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("extractTimeframeCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSecurityCallEmitter_ExtractTimeframeCode_QuotedStrings(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	tests := []struct {
		name     string
		expr     ast.Expression
		expected string
		wantErr  bool
	}{
		{
			name: "double quoted string in literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    `"1D"`,
			},
			expected: `"\"1D\""`,
			wantErr:  false,
		},
		{
			name: "single quoted string in literal",
			expr: &ast.Literal{
				NodeType: ast.TypeLiteral,
				Value:    `'1h'`,
			},
			expected: `"'1h'"`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := emitter.extractTimeframeCode(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimeframeCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("extractTimeframeCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSecurityCallEmitter_RuntimeTimeframeUnquoted(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	expr := &ast.MemberExpression{
		NodeType: ast.TypeMemberExpression,
		Object: &ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     "timeframe",
		},
		Property: &ast.Identifier{
			NodeType: ast.TypeIdentifier,
			Name:     "period",
		},
	}

	result, err := emitter.extractTimeframeCode(expr)
	if err != nil {
		t.Fatalf("extractTimeframeCode() failed: %v", err)
	}

	if result != "ctx.Timeframe" {
		t.Errorf("extractTimeframeCode() = %q, want unquoted %q", result, "ctx.Timeframe")
	}

	if result[0] == '"' || result[0] == '\'' {
		t.Errorf("extractTimeframeCode() returned quoted string %q, expected unquoted variable reference", result)
	}
}

func TestSecurityCallEmitter_LiteralTimeframeQuoted(t *testing.T) {
	emitter := &SecurityCallEmitter{}

	expr := &ast.Literal{
		NodeType: ast.TypeLiteral,
		Value:    "1D",
	}

	result, err := emitter.extractTimeframeCode(expr)
	if err != nil {
		t.Fatalf("extractTimeframeCode() failed: %v", err)
	}

	expected := `"1D"`
	if result != expected {
		t.Errorf("extractTimeframeCode() = %q, want quoted %q", result, expected)
	}

	if result[0] != '"' {
		t.Errorf("extractTimeframeCode() returned unquoted string %q, expected quoted string literal", result)
	}
}
