package security

import (
	"github.com/quant5-lab/runner/ast"
)

/* SecurityCall represents a detected request.security() invocation */
type SecurityCall struct {
	Symbol     string         /* Symbol parameter (e.g., "BTCUSDT", "syminfo.tickerid") */
	Timeframe  string         /* Timeframe parameter (e.g., "1D", "1h") */
	Expression ast.Expression /* AST node of expression argument for evaluation */
	ExprName   string         /* Optional name from array notation: [expr, "name"] */
}

/* AnalyzeAST scans Pine Script AST for request.security() calls */
func AnalyzeAST(program *ast.Program) []SecurityCall {
	if program == nil {
		return []SecurityCall{}
	}

	exprWalker := NewExpressionSecurityWalker()
	stmtWalker := NewStatementSecurityWalker(exprWalker)

	for _, stmt := range program.Body {
		stmtWalker.Walk(stmt)
	}

	return exprWalker.GetCalls()
}
