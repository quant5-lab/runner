package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SecurityExpressionHandler struct {
	indentFunc           func() string
	incrementIndent      func()
	decrementIndent      func()
	serializeExpr        func(ast.Expression) (string, error)
	markSecurityExprEval func()
	symbolTable          SymbolTable
	gen                  *generator // Access to generator for input constants
}

type SecurityExpressionConfig struct {
	IndentFunc           func() string
	IncrementIndent      func()
	DecrementIndent      func()
	SerializeExpr        func(ast.Expression) (string, error)
	MarkSecurityExprEval func()
	SymbolTable          SymbolTable
	Generator            *generator
}

func NewSecurityExpressionHandler(config SecurityExpressionConfig) *SecurityExpressionHandler {
	return &SecurityExpressionHandler{
		indentFunc:           config.IndentFunc,
		incrementIndent:      config.IncrementIndent,
		decrementIndent:      config.DecrementIndent,
		serializeExpr:        config.SerializeExpr,
		markSecurityExprEval: config.MarkSecurityExprEval,
		symbolTable:          config.SymbolTable,
		gen:                  config.Generator,
	}
}

func (h *SecurityExpressionHandler) GenerateEvaluationCode(
	varName string,
	exprArg ast.Expression,
	secBarIdxVar string,
) (string, error) {
	if ident, ok := exprArg.(*ast.Identifier); ok {
		return h.generateOHLCVAccess(varName, ident, secBarIdxVar), nil
	}

	code := ""
	h.markSecurityExprEval()

	initializer := NewSecurityEvaluatorInitializer(h.symbolTable, h.gen)
	code += initializer.EmitInitialization(h.indentFunc, h.incrementIndent, h.decrementIndent)

	exprJSON, err := h.serializeExpr(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	code += h.indentFunc() + fmt.Sprintf("secValue, err := secBarEvaluator.EvaluateAtBar(%s, secCtx, %s)\n", exprJSON, secBarIdxVar)
	code += h.indentFunc() + "if err != nil {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	h.decrementIndent()
	code += h.indentFunc() + "} else {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	return code, nil
}

func (h *SecurityExpressionHandler) generateOHLCVAccess(varName string, ident *ast.Identifier, barIdxVar string) string {
	barAccess := fmt.Sprintf("secCtx.Data[%s]", barIdxVar)

	if ident.Name == "bar_index" {
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(float64(%s))\n", varName, barIdxVar)
	}

	if fieldExpr, ok := SecurityBarFieldExpression(ident.Name, barAccess); ok {
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, fieldExpr)
	}

	return h.indentFunc() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
}

func (h *SecurityExpressionHandler) extractHistoricalOffset(expr ast.Expression) (ast.Expression, int) {
	// Direct subscript: close[1]
	if memberExpr, ok := expr.(*ast.MemberExpression); ok {
		if offsetLit, ok := memberExpr.Property.(*ast.Literal); ok {
			if offsetVal, ok := offsetLit.Value.(float64); ok {
				return memberExpr.Object, int(offsetVal)
			}
		}
	}

	// Nested subscript: fixnan(pivothigh()[1])
	if callExpr, ok := expr.(*ast.CallExpression); ok {
		for i, arg := range callExpr.Arguments {
			if memberExpr, ok := arg.(*ast.MemberExpression); ok {
				if offsetLit, ok := memberExpr.Property.(*ast.Literal); ok {
					if offsetVal, ok := offsetLit.Value.(float64); ok {
						newArgs := make([]ast.Expression, len(callExpr.Arguments))
						copy(newArgs, callExpr.Arguments)
						newArgs[i] = memberExpr.Object

						newCall := &ast.CallExpression{
							Callee:    callExpr.Callee,
							Arguments: newArgs,
						}
						return newCall, int(offsetVal)
					}
				}
			}
		}
	}

	return expr, 0
}
