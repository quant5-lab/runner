package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/*
LocalSeriesAnalyzer determines which arrow function local variables need Series storage.

	Variables need Series if: used in TA functions (rma, sma, ema), accessed historically (var[1]), or in nested TA calls.
*/
type LocalSeriesAnalyzer struct {
	taFunctionsRequiringHistory map[string]bool
}

/* NewLocalSeriesAnalyzer creates analyzer with TA function registry */
func NewLocalSeriesAnalyzer() *LocalSeriesAnalyzer {
	return &LocalSeriesAnalyzer{
		taFunctionsRequiringHistory: map[string]bool{
			"rma":          true,
			"sma":          true,
			"ema":          true,
			"wma":          true,
			"vwma":         true,
			"alma":         true,
			"hma":          true,
			"linreg":       true,
			"ta.rma":       true,
			"ta.sma":       true,
			"ta.ema":       true,
			"ta.wma":       true,
			"ta.vwma":      true,
			"ta.alma":      true,
			"ta.hma":       true,
			"ta.linreg":    true,
			"valuewhen":    true,
			"ta.valuewhen": true,
			"highest":      true,
			"lowest":       true,
			"ta.highest":   true,
			"ta.lowest":    true,
			"tsi":          true,
			"ta.tsi":       true,
		},
	}
}

/* Analyze returns map of variable names requiring Series storage */
func (a *LocalSeriesAnalyzer) Analyze(arrowFunc *ast.ArrowFunctionExpression) map[string]bool {
	needsSeries := make(map[string]bool)
	localVars := a.extractLocalVariables(arrowFunc)

	for _, stmt := range arrowFunc.Body {
		a.analyzeStatement(stmt, localVars, needsSeries)
	}

	return needsSeries
}

func (a *LocalSeriesAnalyzer) extractLocalVariables(arrowFunc *ast.ArrowFunctionExpression) map[string]bool {
	localVars := make(map[string]bool)

	for _, stmt := range arrowFunc.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					localVars[id.Name] = true
				}
			}
		}
	}

	return localVars
}

func (a *LocalSeriesAnalyzer) analyzeStatement(stmt ast.Node, localVars, needsSeries map[string]bool) {
	switch s := stmt.(type) {
	case *ast.VariableDeclaration:
		for _, declarator := range s.Declarations {
			if declarator.Init != nil {
				a.analyzeExpression(declarator.Init, localVars, needsSeries)
			}
		}

	case *ast.ExpressionStatement:
		a.analyzeExpression(s.Expression, localVars, needsSeries)
	}
}

func (a *LocalSeriesAnalyzer) analyzeExpression(expr ast.Expression, localVars, needsSeries map[string]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		a.analyzeCall(e, localVars, needsSeries)

	case *ast.BinaryExpression:
		a.analyzeExpression(e.Left, localVars, needsSeries)
		a.analyzeExpression(e.Right, localVars, needsSeries)

	case *ast.UnaryExpression:
		a.analyzeExpression(e.Argument, localVars, needsSeries)

	case *ast.ConditionalExpression:
		a.analyzeExpression(e.Test, localVars, needsSeries)
		a.analyzeExpression(e.Consequent, localVars, needsSeries)
		a.analyzeExpression(e.Alternate, localVars, needsSeries)

	case *ast.MemberExpression:
		a.analyzeHistoricalAccess(e, localVars, needsSeries)
	}
}

func (a *LocalSeriesAnalyzer) analyzeCall(call *ast.CallExpression, localVars, needsSeries map[string]bool) {
	funcName := extractCallFunctionName(call)

	if a.taFunctionsRequiringHistory[funcName] {
		for _, arg := range call.Arguments {
			if id, ok := arg.(*ast.Identifier); ok {
				if localVars[id.Name] {
					needsSeries[id.Name] = true
				}
			}
			a.analyzeExpression(arg, localVars, needsSeries)
		}
	}

	for _, arg := range call.Arguments {
		a.analyzeExpression(arg, localVars, needsSeries)
	}
}

func (a *LocalSeriesAnalyzer) analyzeHistoricalAccess(member *ast.MemberExpression, localVars, needsSeries map[string]bool) {
	if id, ok := member.Object.(*ast.Identifier); ok {
		if localVars[id.Name] {
			needsSeries[id.Name] = true
		}
	}

	a.analyzeExpression(member.Object, localVars, needsSeries)
	if member.Property != nil {
		if propExpr, ok := member.Property.(ast.Expression); ok {
			a.analyzeExpression(propExpr, localVars, needsSeries)
		}
	}
}
