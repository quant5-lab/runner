package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/*
MathHandler generates code for math.* function calls with expression arguments.

Design: Evaluate each argument expression, then call runtime math function.
Rationale: Reuses existing expression evaluation, no special cases needed.
*/
type MathHandler struct{}

func NewMathHandler() *MathHandler {
	return &MathHandler{}
}

/*
GenerateMathCall generates code for math.* functions.

Returns: Go expression string that evaluates the math function
*/
func (mh *MathHandler) GenerateMathCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	// Normalize function name
	funcName = strings.ToLower(funcName)

	switch funcName {
	case "math.pow":
		return mh.generatePow(args, g)
	case "math.abs", "abs", "math.sqrt", "sqrt", "math.floor", "floor", "math.ceil", "ceil", "math.round", "round", "math.log", "log", "math.exp", "exp":
		return mh.generateUnaryMath(funcName, args, g)
	case "math.max", "max", "math.min", "min":
		return mh.generateBinaryMath(funcName, args, g)
	default:
		return "", fmt.Errorf("unsupported math function: %s", funcName)
	}
}

func (mh *MathHandler) generatePow(args []ast.Expression, g *generator) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("math.pow requires exactly 2 arguments")
	}

	base := g.extractSeriesExpression(args[0])
	exponent := g.extractSeriesExpression(args[1])

	return fmt.Sprintf("math.Pow(%s, %s)", base, exponent), nil
}

func (mh *MathHandler) generateUnaryMath(funcName string, args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s requires exactly 1 argument", funcName)
	}

	arg := g.extractSeriesExpression(args[0])

	// Extract function name after "math." and capitalize
	// "math.abs" -> "Abs"
	shortName := funcName[5:] // Remove "math."
	goFuncName := "math." + strings.ToUpper(shortName[:1]) + shortName[1:]

	return fmt.Sprintf("%s(%s)", goFuncName, arg), nil
}

func (mh *MathHandler) generateBinaryMath(funcName string, args []ast.Expression, g *generator) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("%s requires exactly 2 arguments", funcName)
	}

	arg1 := g.extractSeriesExpression(args[0])
	arg2 := g.extractSeriesExpression(args[1])

	// Normalize function name (Pine "max" → Go "math.Max")
	var goFuncName string
	if strings.HasPrefix(funcName, "math.") {
		// Already has math. prefix (math.max → math.Max)
		shortName := funcName[5:]
		goFuncName = "math." + strings.ToUpper(shortName[:1]) + shortName[1:]
	} else {
		// Pine function without prefix (max → math.Max)
		goFuncName = "math." + strings.ToUpper(funcName[:1]) + funcName[1:]
	}

	return fmt.Sprintf("%s(%s, %s)", goFuncName, arg1, arg2), nil
}
