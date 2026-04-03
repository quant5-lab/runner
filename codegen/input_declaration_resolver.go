package codegen

import "github.com/quant5-lab/runner/ast"

func (g *generator) tryResolveInputConstant(d ast.VariableDeclarator) bool {
	id, ok := d.ID.(*ast.Identifier)
	if !ok {
		return false
	}
	callExpr, ok := d.Init.(*ast.CallExpression)
	if !ok {
		return false
	}

	varName := SanitizeGoIdentifier(id.Name)
	funcName := g.extractFunctionName(callExpr.Callee)

	if g.inputHandler != nil {
		if funcName == "input" && len(callExpr.Arguments) > 0 {
			if resolved := resolveInputFuncName(callExpr); resolved != "" {
				funcName = resolved
			}
		}
		if g.tryRegisterInputHandlerConstant(varName, funcName, callExpr) {
			return true
		}
	}

	if funcName == "input.source" {
		g.constants[varName] = funcName
		g.variables[varName] = "float"
		g.typeSystem.RegisterVariable(varName, "float")
		return true
	}

	return false
}

func (g *generator) tryRegisterInputHandlerConstant(varName, funcName string, callExpr *ast.CallExpression) bool {
	handler := g.inputHandlerForFunc(funcName)
	if handler == nil {
		return false
	}

	code, _ := handler(callExpr, varName)
	if code != "" {
		if val := g.constantRegistry.ExtractFromGeneratedCode(code); val != nil {
			g.constants[varName] = val
			g.constantRegistry.Register(varName, val)
		}
	}
	return true
}

func (g *generator) inputHandlerForFunc(funcName string) func(*ast.CallExpression, string) (string, error) {
	switch InputValueKindOf(funcName) {
	case InputValueFloat:
		return g.inputHandler.GenerateInputFloat
	case InputValueInt:
		return g.inputHandler.GenerateInputInt
	case InputValueBool:
		return g.inputHandler.GenerateInputBool
	case InputValueString:
		return g.inputHandler.GenerateInputString
	case InputValueSession:
		return g.inputHandler.GenerateInputSession
	case InputValueColor:
		return g.inputHandler.GenerateInputColor
	default:
		return nil
	}
}
