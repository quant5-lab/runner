package security

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

// TestExpressionKey_Stability verifies that identical AST structures produce
// the same key string and that structurally distinct expressions produce
// distinct keys — the two invariants required for correct TA cache keying.
func TestExpressionKey_Stability(t *testing.T) {
	close1 := &ast.Identifier{Name: "close"}
	close2 := &ast.Identifier{Name: "close"}

	if expressionKey(close1) != expressionKey(close2) {
		t.Errorf("identical Identifiers produced different keys: %q vs %q",
			expressionKey(close1), expressionKey(close2))
	}

	open := &ast.Identifier{Name: "open"}
	if expressionKey(close1) == expressionKey(open) {
		t.Errorf("different identifiers 'close' and 'open' produced the same key: %q",
			expressionKey(close1))
	}
}

func TestExpressionKey_NodeTypes(t *testing.T) {
	closeID := &ast.Identifier{Name: "close"}
	openID := &ast.Identifier{Name: "open"}
	numLit := &ast.Literal{Value: 14.0}

	cases := []struct {
		name string
		expr ast.Expression
		want string
	}{
		{"nil", nil, "nil"},
		{"identifier", closeID, "close"},
		{"literal_float", numLit, "14"},
		{"unary_neg", &ast.UnaryExpression{Operator: "-", Argument: closeID}, "(-close)"},
		{"binary_add", &ast.BinaryExpression{Left: closeID, Operator: "+", Right: openID}, "(close+open)"},
		{"binary_mul", &ast.BinaryExpression{Left: closeID, Operator: "*", Right: numLit}, "(close*14)"},
		{
			"member_property",
			&ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ta"},
				Property: &ast.Identifier{Name: "tr"},
			},
			"ta.tr",
		},
		{
			"member_subscript",
			&ast.MemberExpression{
				Object:   closeID,
				Property: numLit,
			},
			"close[14]",
		},
		{
			"call_one_arg",
			&ast.CallExpression{
				Callee:    &ast.Identifier{Name: "abs"},
				Arguments: []ast.Expression{closeID},
			},
			"abs(close)",
		},
		{
			"call_two_args",
			&ast.CallExpression{
				Callee:    &ast.Identifier{Name: "max"},
				Arguments: []ast.Expression{closeID, openID},
			},
			"max(close,open)",
		},
		{
			"conditional",
			&ast.ConditionalExpression{
				Test:       closeID,
				Consequent: openID,
				Alternate:  numLit,
			},
			"(close?open:14)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := expressionKey(c.expr)
			if got != c.want {
				t.Errorf("expressionKey = %q, want %q", got, c.want)
			}
		})
	}
}

func TestExpressionKey_NestedExpressionProducesUniqueKey(t *testing.T) {
	closeID := &ast.Identifier{Name: "close"}
	lit3 := &ast.Literal{Value: 3.0}
	lit5 := &ast.Literal{Value: 5.0}

	ema3 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{closeID, lit3},
	}
	ema5 := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
		Arguments: []ast.Expression{closeID, lit5},
	}

	key3 := expressionKey(ema3)
	key5 := expressionKey(ema5)

	if key3 == key5 {
		t.Errorf("ema(close,3) and ema(close,5) produced identical keys: %q", key3)
	}

	key3Again := expressionKey(ema3)
	if key3 != key3Again {
		t.Errorf("non-deterministic key for same expression: %q vs %q", key3, key3Again)
	}
}

func TestExpressionKey_SourceVsArgumentDistinction(t *testing.T) {
	closeID := &ast.Identifier{Name: "close"}
	openID := &ast.Identifier{Name: "open"}

	smaClose := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{closeID, &ast.Literal{Value: 5.0}},
	}
	smaOpen := &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "sma"},
		},
		Arguments: []ast.Expression{openID, &ast.Literal{Value: 5.0}},
	}

	if expressionKey(smaClose) == expressionKey(smaOpen) {
		t.Errorf("sma(close,5) and sma(open,5) share the same key: %q", expressionKey(smaClose))
	}
}
