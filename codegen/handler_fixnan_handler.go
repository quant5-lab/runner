package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* FixnanHandler generates inline code for forward-filling NaN values */
type FixnanHandler struct{}

func (h *FixnanHandler) CanHandle(funcName string) bool {
	return funcName == "fixnan"
}

func (h *FixnanHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("fixnan requires 1 argument")
	}

	var code string
	argExpr := call.Arguments[0]

	/* Handle nested function: fixnan(pivothigh()[1]) */
	if memberExpr, ok := argExpr.(*ast.MemberExpression); ok {
		if nestedCall, isCall := memberExpr.Object.(*ast.CallExpression); isCall {
			/* Generate intermediate variable for nested function */
			nestedFuncName := g.extractFunctionName(nestedCall.Callee)
			tempVarName := strings.ReplaceAll(nestedFuncName, ".", "_")

			/* Generate nested function code */
			nestedCode, err := g.generateVariableFromCall(tempVarName, nestedCall)
			if err != nil {
				return "", fmt.Errorf("failed to generate nested function in fixnan: %w", err)
			}
			code += nestedCode
		}
	}

	sourceExpr := g.extractSeriesExpression(argExpr)
	stateVar := "fixnanState_" + varName

	code += g.ind() + fmt.Sprintf("if !math.IsNaN(%s) {\n", sourceExpr)
	g.indent++
	code += g.ind() + fmt.Sprintf("%s = %s\n", stateVar, sourceExpr)
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, stateVar)

	return code, nil
}
