package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type StatementConditionalAnalyzer struct {
	gen *generator
}

func NewStatementConditionalAnalyzer(g *generator) *StatementConditionalAnalyzer {
	return &StatementConditionalAnalyzer{gen: g}
}

func (a *StatementConditionalAnalyzer) Analyze(program *ast.Program) {
	for _, stmt := range program.Body {
		a.analyzeStatement(stmt)
	}
}

func (a *StatementConditionalAnalyzer) analyzeStatement(stmt ast.Node) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		a.registerConditionalsInExpression(s.Expression)

	case *ast.IfStatement:
		if s.Test != nil {
			a.registerConditionalsInExpression(s.Test)
		}
		a.analyzeStatementBlock(s.Consequent)
		a.analyzeStatementBlock(s.Alternate)
	}
}

func (a *StatementConditionalAnalyzer) analyzeStatementBlock(nodes []ast.Node) {
	for _, node := range nodes {
		a.analyzeStatement(node)
	}
}

func (a *StatementConditionalAnalyzer) registerConditionalsInExpression(expr ast.Expression) {
	conditionals := a.gen.conditionalArgAnalyzer.FindInExpression(expr)
	for _, condInfo := range conditionals {
		a.gen.tempVarMgr.RegisterConditional(condInfo.ContentHash, condInfo.Conditional)
	}
}
