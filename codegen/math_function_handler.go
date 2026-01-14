package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// MathFunctionHandler generates Series-based code for math functions with TA dependencies.
//
// Purpose: When math functions like max(change(x), 0) contain TA calls,
// they need Series storage (ForwardSeriesBuffer paradigm) rather than inline evaluation.
//
// Example:
//
//	max(change(close), 0) →
//	Step 1: ta_changeSeries.Set(change(close))
//	Step 2: maxSeries.Set(math.Max(ta_changeSeries.GetCurrent(), 0))
type MathFunctionHandler struct{}

func NewMathFunctionHandler() *MathFunctionHandler {
	return &MathFunctionHandler{}
}

// CanHandle checks if this is a math function that might need Series storage
func (h *MathFunctionHandler) CanHandle(funcName string) bool {
	return funcName == "max" || funcName == "min" ||
		funcName == "abs" || funcName == "sqrt" ||
		funcName == "floor" || funcName == "ceil" ||
		funcName == "round" || funcName == "log" || funcName == "exp"
}

// GenerateCode generates Series.Set() code for math function
//
// This is called when the math function has TA dependencies and needs
// to store its result in a Series variable (ForwardSeriesBuffer paradigm)
func (h *MathFunctionHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	// Generate inline math expression using MathHandler
	mathExpr, err := g.mathHandler.GenerateMathCall(funcName, call.Arguments, g)
	if err != nil {
		return "", fmt.Errorf("failed to generate math expression for %s: %w", funcName, err)
	}

	// Wrap in Series.Set() for bar-to-bar storage
	code := g.ind() + fmt.Sprintf("/* Inline %s() with TA dependencies */\n", funcName)
	code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, mathExpr)

	return code, nil
}
