package codegen

import "github.com/quant5-lab/runner/ast"

type ParameterUsageType int

const (
	ParameterUsageScalar ParameterUsageType = iota
	ParameterUsageSeries
)

type ParameterUsageAnalyzer struct {
	parameterTypes map[string]ParameterUsageType
}

func NewParameterUsageAnalyzer() *ParameterUsageAnalyzer {
	return &ParameterUsageAnalyzer{
		parameterTypes: make(map[string]ParameterUsageType),
	}
}

func (a *ParameterUsageAnalyzer) AnalyzeArrowFunction(arrowFunc *ast.ArrowFunctionExpression) map[string]ParameterUsageType {
	for _, param := range arrowFunc.Params {
		a.parameterTypes[param.Name] = ParameterUsageScalar
	}

	for _, stmt := range arrowFunc.Body {
		a.analyzeStatement(stmt)
	}

	return a.parameterTypes
}

func (a *ParameterUsageAnalyzer) analyzeStatement(stmt ast.Node) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		a.analyzeExpression(s.Expression)
	case *ast.VariableDeclaration:
		for _, decl := range s.Declarations {
			if decl.Init != nil {
				a.analyzeExpression(decl.Init)
			}
		}
	case *ast.ForStatement:
		for _, bodyStmt := range s.Body {
			a.analyzeStatement(bodyStmt)
		}
	case *ast.ForInStatement:
		for _, bodyStmt := range s.Body {
			a.analyzeStatement(bodyStmt)
		}
	case *ast.IfStatement:
		a.analyzeExpression(s.Test)
		for _, conseq := range s.Consequent {
			a.analyzeStatement(conseq)
		}
		for _, alt := range s.Alternate {
			a.analyzeStatement(alt)
		}
	}
}

func (a *ParameterUsageAnalyzer) analyzeExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.CallExpression:
		a.analyzeCallExpression(e)
	case *ast.BinaryExpression:
		a.analyzeExpression(e.Left)
		a.analyzeExpression(e.Right)
	case *ast.ConditionalExpression:
		a.analyzeExpression(e.Test)
		a.analyzeExpression(e.Consequent)
		a.analyzeExpression(e.Alternate)
	case *ast.UnaryExpression:
		a.analyzeExpression(e.Argument)
	case *ast.MemberExpression:
		/* Subscript access on parameter: src[i] marks src as series */
		if e.Computed {
			if obj, ok := e.Object.(*ast.Identifier); ok {
				if _, isParam := a.parameterTypes[obj.Name]; isParam {
					a.parameterTypes[obj.Name] = ParameterUsageSeries
				}
			}
		}
	case *ast.Literal:
		if elemSlice, ok := e.Value.([]ast.Expression); ok {
			for _, elem := range elemSlice {
				a.analyzeExpression(elem)
			}
		}
	}
}

func (a *ParameterUsageAnalyzer) analyzeCallExpression(call *ast.CallExpression) {
	funcName := extractCallFunctionName(call)
	argCount := len(call.Arguments)

	/* Promote first arg to series when TA function needs historical lookback on its source */
	if sharedTASignatures.Contains(funcName) && argCount >= 1 {
		promoteFirstArg := false
		if argCount >= 2 && sharedTASignatures.NeedsSourcePromotion(funcName, argCount) {
			promoteFirstArg = true
		} else if argCount == 1 && sharedTASignatures.IsSourceOnlyLookback(funcName) {
			promoteFirstArg = true
		}
		if promoteFirstArg {
			sourceArg := call.Arguments[0]
			if ident, ok := sourceArg.(*ast.Identifier); ok {
				if _, isParam := a.parameterTypes[ident.Name]; isParam {
					a.parameterTypes[ident.Name] = ParameterUsageSeries
				}
			}
		}
	}

	for _, arg := range call.Arguments {
		a.analyzeExpression(arg)
	}
}
