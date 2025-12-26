package codegen

import "github.com/quant5-lab/runner/ast"

type ArrowCallSite struct {
	FunctionName string
	CallIndex    int
	ContextVar   string
}

/*
ArrowCallSiteScanner detects arrow function calls requiring ArrowContext.

Design (SRP): Single purpose - identify call sites, delegate generation and lifecycle
*/
type ArrowCallSiteScanner struct {
	variables map[string]string
}

func NewArrowCallSiteScanner(variables map[string]string) *ArrowCallSiteScanner {
	return &ArrowCallSiteScanner{
		variables: variables,
	}
}

func (s *ArrowCallSiteScanner) ScanForArrowFunctionCalls(program *ast.Program) []ArrowCallSite {
	var callSites []ArrowCallSite
	callCounts := make(map[string]int)

	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}

		for _, declarator := range varDecl.Declarations {
			callExpr := s.extractCallExpression(declarator.Init)
			if callExpr == nil {
				continue
			}

			funcName := s.extractFunctionName(callExpr.Callee)
			if !s.isUserDefinedFunction(funcName) {
				continue
			}

			callCounts[funcName]++
			callIndex := callCounts[funcName]

			callSites = append(callSites, ArrowCallSite{
				FunctionName: funcName,
				CallIndex:    callIndex,
				ContextVar:   formatContextVariableName(funcName, callIndex),
			})
		}
	}

	return callSites
}

func (s *ArrowCallSiteScanner) extractCallExpression(expr ast.Expression) *ast.CallExpression {
	if expr == nil {
		return nil
	}

	if callExpr, ok := expr.(*ast.CallExpression); ok {
		return callExpr
	}

	return nil
}

func (s *ArrowCallSiteScanner) extractFunctionName(callee ast.Expression) string {
	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}

	if member, ok := callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}

	return ""
}

func (s *ArrowCallSiteScanner) isUserDefinedFunction(funcName string) bool {
	varType, exists := s.variables[funcName]
	return exists && varType == "function"
}

func formatContextVariableName(funcName string, callIndex int) string {
	return "arrowCtx_" + funcName + "_" + string(rune('0'+callIndex))
}
