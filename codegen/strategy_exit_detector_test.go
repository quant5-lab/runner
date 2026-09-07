package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func exitCallExpr() ast.Expression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "exit"},
		},
	}
}

func entryCallExpr() ast.Expression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "entry"},
		},
	}
}

func nonStrategyCallExpr() ast.Expression {
	return &ast.CallExpression{
		Callee: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "ta"},
			Property: &ast.Identifier{Name: "ema"},
		},
	}
}

func exprStmt(e ast.Expression) ast.Node {
	return &ast.ExpressionStatement{Expression: e}
}

func TestDetectStrategyExitCalls(t *testing.T) {
	tests := []struct {
		name    string
		program *ast.Program
		want    bool
	}{
		{
			name:    "nil program",
			program: nil,
			want:    false,
		},
		{
			name:    "empty body",
			program: &ast.Program{Body: []ast.Node{}},
			want:    false,
		},
		{
			name: "exit at top level",
			program: &ast.Program{Body: []ast.Node{
				exprStmt(exitCallExpr()),
			}},
			want: true,
		},
		{
			name: "exit inside if consequent",
			program: &ast.Program{Body: []ast.Node{
				&ast.IfStatement{
					Test:       &ast.Literal{Value: true},
					Consequent: []ast.Node{exprStmt(exitCallExpr())},
				},
			}},
			want: true,
		},
		{
			name: "exit inside if alternate (else branch)",
			program: &ast.Program{Body: []ast.Node{
				&ast.IfStatement{
					Test:       &ast.Literal{Value: true},
					Consequent: []ast.Node{exprStmt(entryCallExpr())},
					Alternate:  []ast.Node{exprStmt(exitCallExpr())},
				},
			}},
			want: true,
		},
		{
			name: "only strategy.entry, no exit",
			program: &ast.Program{Body: []ast.Node{
				exprStmt(entryCallExpr()),
			}},
			want: false,
		},
		{
			name: "only non-strategy call",
			program: &ast.Program{Body: []ast.Node{
				exprStmt(nonStrategyCallExpr()),
			}},
			want: false,
		},
		{
			name: "entry then exit: exit is found",
			program: &ast.Program{Body: []ast.Node{
				exprStmt(entryCallExpr()),
				exprStmt(exitCallExpr()),
			}},
			want: true,
		},
		{
			name: "multiple if blocks, exit only in second",
			program: &ast.Program{Body: []ast.Node{
				&ast.IfStatement{
					Test:       &ast.Literal{Value: true},
					Consequent: []ast.Node{exprStmt(entryCallExpr())},
				},
				&ast.IfStatement{
					Test:       &ast.Literal{Value: false},
					Consequent: []ast.Node{exprStmt(exitCallExpr())},
				},
			}},
			want: true,
		},
		{
			name: "exit in variable declaration init",
			program: &ast.Program{Body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{ID: &ast.Identifier{Name: "v"}, Init: exitCallExpr()},
					},
				},
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectStrategyExitCalls(tt.program)
			if got != tt.want {
				t.Errorf("detectStrategyExitCalls() = %v, want %v", got, tt.want)
			}
		})
	}
}
