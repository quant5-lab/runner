package parser

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* Inline case body tests: behavior unique to the inline body form */

func TestSwitchInlineCase_MixedInlineAndMultiLine(t *testing.T) {
	source := `x = switch mode
    1 => simpleVal
    2 =>
        complex = calc()
        complex
    => defaultVal`

	program := parseSwitchSource(t, source)
	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt := varDecl.Declarations[0].Init.(*ast.IfStatement)

	if len(ifStmt.Consequent) != 1 {
		t.Fatalf("inline case should have 1 consequent node, got %d", len(ifStmt.Consequent))
	}

	nested := ifStmt.Alternate[0].(*ast.IfStatement)
	if len(nested.Consequent) != 2 {
		t.Fatalf("multi-line case should have 2 consequent nodes, got %d", len(nested.Consequent))
	}

	lastIf := getLastIfInChain(ifStmt)
	if len(lastIf.Alternate) != 1 {
		t.Fatalf("expected default alternate, got %d", len(lastIf.Alternate))
	}
}

func TestSwitchInlineCase_ExpressionVariety(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		checkNode func(t *testing.T, node ast.Node)
	}{
		{
			name: "member access",
			source: `x = switch style
    "none" => extend.none
    "left" => extend.left`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				member, ok := expr.(*ast.MemberExpression)
				if !ok {
					t.Fatalf("expected MemberExpression, got %T", expr)
				}
				obj := member.Object.(*ast.Identifier)
				if obj.Name != "extend" {
					t.Fatalf("expected object 'extend', got %q", obj.Name)
				}
			},
		},
		{
			name: "function call",
			source: `x = switch maType
    "EMA" => ta.ema(close, 10)
    "SMA" => ta.sma(close, 10)`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				call, ok := expr.(*ast.CallExpression)
				if !ok {
					t.Fatalf("expected CallExpression, got %T", expr)
				}
				if len(call.Arguments) != 2 {
					t.Fatalf("expected 2 arguments, got %d", len(call.Arguments))
				}
			},
		},
		{
			name: "arithmetic",
			source: `x = switch mode
    1 => close * 2
    2 => open + high`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				bin, ok := expr.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("expected BinaryExpression, got %T", expr)
				}
				if bin.Operator != "*" {
					t.Fatalf("expected '*' operator, got %q", bin.Operator)
				}
			},
		},
		{
			name: "ternary",
			source: `x = switch mode
    1 => close > open ? 1 : 0
    => 0`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				if _, ok := expr.(*ast.ConditionalExpression); !ok {
					t.Fatalf("expected ConditionalExpression, got %T", expr)
				}
			},
		},
		{
			name: "unary",
			source: `x = switch mode
    1 => -close
    2 => not flag`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				unary, ok := expr.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("expected UnaryExpression, got %T", expr)
				}
				if unary.Operator != "-" {
					t.Fatalf("expected '-' operator, got %q", unary.Operator)
				}
			},
		},
		{
			name: "subscript access",
			source: `x = switch mode
    1 => close[1]
    => close[0]`,
			checkNode: func(t *testing.T, node ast.Node) {
				t.Helper()
				expr := node.(*ast.ExpressionStatement).Expression
				if _, ok := expr.(*ast.MemberExpression); !ok {
					t.Fatalf("expected MemberExpression (subscript), got %T", expr)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parseSwitchSource(t, tt.source)
			varDecl := program.Body[0].(*ast.VariableDeclaration)
			ifStmt := varDecl.Declarations[0].Init.(*ast.IfStatement)

			if len(ifStmt.Consequent) != 1 {
				t.Fatalf("expected 1 consequent node, got %d", len(ifStmt.Consequent))
			}
			tt.checkNode(t, ifStmt.Consequent[0])
		})
	}
}

func TestSwitchInlineCase_ManyCasesChainDepth(t *testing.T) {
	source := `x = switch tf
    "1m" => 1
    "3m" => 3
    "5m" => 5
    "10m" => 10
    "15m" => 15
    "30m" => 30
    "1h" => 60
    "4h" => 240
    "D" => 1440
    => 0`

	program := parseSwitchSource(t, source)
	varDecl := program.Body[0].(*ast.VariableDeclaration)
	ifStmt := varDecl.Declarations[0].Init.(*ast.IfStatement)

	depth := countIfChainDepth(ifStmt)
	if depth != 9 {
		t.Fatalf("expected 9-deep if-chain, got %d", depth)
	}

	lastIf := getLastIfInChain(ifStmt)
	if len(lastIf.Alternate) != 1 {
		t.Fatalf("expected default alternate, got %d", len(lastIf.Alternate))
	}
}

func TestSwitchInlineCase_RealStrategyPatterns(t *testing.T) {
	source := `//@version=5
indicator("test")

convertTimeframe(tf) =>
    switch tf
        '1m' => '1'
        '3m' => '3'
        '5m' => '5'
        '15m' => '15'
        '1h' => '60'
        'D' => 'D'

extensionStyle = switch lineExtension
    'Current' => extend.none
    'Left' => extend.left
    'Right' => extend.right
    'Both' => extend.both

lineStyleConverted = switch lineStyle
    'Solid' => line.style_solid
    'Dotted' => line.style_dotted
    'Dashed' => line.style_dashed
`

	program := parseSwitchSource(t, source)
	if program == nil {
		t.Fatal("parse returned nil")
	}
}
