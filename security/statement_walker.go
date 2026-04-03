package security

import "github.com/quant5-lab/runner/ast"

/* StatementSecurityWalker recursively walks statements to find security() calls */
type StatementSecurityWalker struct {
	exprWalker *ExpressionSecurityWalker
}

func NewStatementSecurityWalker(exprWalker *ExpressionSecurityWalker) *StatementSecurityWalker {
	return &StatementSecurityWalker{
		exprWalker: exprWalker,
	}
}

func (w *StatementSecurityWalker) Walk(stmt ast.Node) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		w.walkVariableDeclaration(s)
	case *ast.ExpressionStatement:
		w.walkExpressionStatement(s)
	case *ast.IfStatement:
		w.walkIfStatement(s)
	case *ast.ForStatement:
		w.walkForStatement(s)
	case *ast.WhileStatement:
		w.walkWhileStatement(s)
	}
}

func (w *StatementSecurityWalker) walkVariableDeclaration(varDecl *ast.VariableDeclaration) {
	for _, declarator := range varDecl.Declarations {
		w.exprWalker.Walk(declarator.Init)
	}
}

func (w *StatementSecurityWalker) walkExpressionStatement(exprStmt *ast.ExpressionStatement) {
	w.exprWalker.Walk(exprStmt.Expression)
}

func (w *StatementSecurityWalker) walkIfStatement(ifStmt *ast.IfStatement) {
	w.exprWalker.Walk(ifStmt.Test)

	for _, consequent := range ifStmt.Consequent {
		w.Walk(consequent)
	}

	for _, alternate := range ifStmt.Alternate {
		w.Walk(alternate)
	}
}

func (w *StatementSecurityWalker) walkForStatement(forStmt *ast.ForStatement) {
	w.exprWalker.Walk(forStmt.From)
	w.exprWalker.Walk(forStmt.To)
	w.exprWalker.Walk(forStmt.Step)

	for _, bodyStmt := range forStmt.Body {
		w.Walk(bodyStmt)
	}
}

func (w *StatementSecurityWalker) walkWhileStatement(whileStmt *ast.WhileStatement) {
	w.exprWalker.Walk(whileStmt.Condition)

	for _, bodyStmt := range whileStmt.Body {
		w.Walk(bodyStmt)
	}
}
