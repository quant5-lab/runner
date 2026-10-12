package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/*
TestStrategyRuntimeAccessDetection covers the gate that decides whether the strategy runtime series
are declared. A runtime read must be seen however deeply it is nested, including inside the
named-argument object a strategy call uses, while a non-runtime member must never enable the block.
*/
func TestStrategyRuntimeAccessDetection(t *testing.T) {
	positionSize := member("strategy", "position_size")

	tests := []struct {
		name string
		expr ast.Expression
		want bool
	}{
		{name: "bare runtime member", expr: positionSize, want: true},
		{
			name: "runtime member in a comparison",
			expr: &ast.BinaryExpression{Operator: ">", Left: positionSize, Right: literal(0.0)},
			want: true,
		},
		{
			name: "runtime member in a positional argument",
			expr: &ast.CallExpression{Callee: identifier("plot"), Arguments: []ast.Expression{positionSize}},
			want: true,
		},
		{
			name: "runtime member in a named argument",
			expr: &ast.CallExpression{Callee: member("strategy", "close"), Arguments: []ast.Expression{objectArg("when", positionSize)}},
			want: true,
		},
		{
			name: "runtime member in a logical operand",
			expr: &ast.LogicalExpression{Operator: "and", Left: positionSize, Right: identifier("enabled")},
			want: true,
		},
		{
			name: "runtime member under negation",
			expr: &ast.UnaryExpression{Operator: "not", Argument: positionSize},
			want: true,
		},
		{
			name: "runtime member in a conditional branch",
			expr: &ast.ConditionalExpression{
				Test:       identifier("enabled"),
				Consequent: positionSize,
				Alternate:  literal(0.0),
			},
			want: true,
		},
		{
			name: "non-runtime strategy member",
			expr: member("strategy", "long"),
			want: false,
		},
		{
			name: "named argument without a runtime member",
			expr: &ast.CallExpression{Callee: member("strategy", "close"), Arguments: []ast.Expression{objectArg("when", identifier("enabled"))}},
			want: false,
		},
		{name: "bare identifier", expr: identifier("position_size"), want: false},
		{name: "absent expression", expr: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasStrategyRuntimeInExpression(tt.expr); got != tt.want {
				t.Errorf("hasStrategyRuntimeInExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

/*
TestDetectStrategyRuntimeAccess_Programs covers the statement walk that feeds the expression
predicate, so a runtime read reaches the gate from a call statement, a declaration initialiser and
an if statement test alike.
*/
func TestDetectStrategyRuntimeAccess_Programs(t *testing.T) {
	positionSize := member("strategy", "position_size")
	entryCall := func(args ...ast.Expression) *ast.CallExpression {
		return &ast.CallExpression{Callee: member("strategy", "entry"), Arguments: args}
	}

	tests := []struct {
		name string
		body []ast.Node
		want bool
	}{
		{
			name: "named argument in a call statement",
			body: []ast.Node{
				&ast.ExpressionStatement{Expression: entryCall(objectArg("when", positionSize))},
			},
			want: true,
		},
		{
			name: "declaration initialiser",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Kind:         "let",
					Declarations: []ast.VariableDeclarator{{ID: identifier("open"), Init: positionSize}},
				},
			},
			want: true,
		},
		{
			name: "if statement test",
			body: []ast.Node{
				&ast.IfStatement{
					Test:       &ast.BinaryExpression{Operator: ">", Left: positionSize, Right: literal(0.0)},
					Consequent: []ast.Node{&ast.ExpressionStatement{Expression: entryCall()}},
				},
			},
			want: true,
		},
		{
			name: "no runtime read",
			body: []ast.Node{
				&ast.ExpressionStatement{Expression: entryCall(objectArg("when", identifier("enabled")))},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := &ast.Program{Body: tt.body}
			if got := detectStrategyRuntimeAccess(program); got != tt.want {
				t.Errorf("detectStrategyRuntimeAccess() = %v, want %v", got, tt.want)
			}
		})
	}

	if detectStrategyRuntimeAccess(nil) {
		t.Error("a nil program must not report strategy runtime access")
	}
}
