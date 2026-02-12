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
	switch funcName {
	case "input.float":
		return g.inputHandler.GenerateInputFloat
	case "input.int":
		return g.inputHandler.GenerateInputInt
	case "input.bool":
		return g.inputHandler.GenerateInputBool
	case "input.string":
		return g.inputHandler.GenerateInputString
	case "input.session":
		return g.inputHandler.GenerateInputSession
	default:
		return nil
	}
}
