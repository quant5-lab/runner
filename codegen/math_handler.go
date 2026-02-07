package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type MathHandler struct{}

func NewMathHandler() *MathHandler {
	return &MathHandler{}
}

func (mh *MathHandler) normalizeToGoMathFunc(pineFuncName string) string {
	if strings.HasPrefix(pineFuncName, "math.") {
		shortName := pineFuncName[5:]
		return "math." + strings.ToUpper(shortName[:1]) + shortName[1:]
	}
	return "math." + strings.ToUpper(pineFuncName[:1]) + pineFuncName[1:]
}

/* CanHandle checks if this is an inline math function */
func (mh *MathHandler) CanHandle(funcName string) bool {
	funcName = strings.ToLower(funcName)
	switch funcName {
	case "math.pow",
		"math.abs", "abs",
		"math.sqrt", "sqrt",
		"math.floor", "floor",
		"math.ceil", "ceil",
		"math.round", "round",
		"math.log", "log",
		"math.log10", "log10",
		"math.exp", "exp",
		"math.max", "max",
		"math.min", "min",
		"math.sin", "sin",
		"math.cos", "cos",
		"math.tan", "tan",
		"math.asin", "asin",
		"math.acos", "acos",
		"math.atan", "atan",
		"math.sign", "sign",
		"math.todegrees", "todegrees",
		"math.toradians", "toradians",
		"math.avg", "avg",
		"math.random", "random",
		"math.round_to_mintick", "round_to_mintick":
		return true
	default:
		return false
	}
}

/* GenerateInline implements InlineConditionHandler interface */
func (mh *MathHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	funcName := g.extractFunctionName(expr.Callee)
	return mh.GenerateMathCall(funcName, expr.Arguments, g)
}

func (mh *MathHandler) GenerateMathCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	funcName = strings.ToLower(funcName)

	switch funcName {
	case "math.pow":
		return mh.generatePow(args, g)
	case "math.abs", "abs", "math.sqrt", "sqrt", "math.floor", "floor", "math.ceil", "ceil",
		"math.round", "round", "math.log", "log", "math.log10", "log10", "math.exp", "exp",
		"math.sin", "sin", "math.cos", "cos", "math.tan", "tan",
		"math.asin", "asin", "math.acos", "acos", "math.atan", "atan":
		return mh.generateUnaryMath(funcName, args, g)
	case "math.max", "max", "math.min", "min":
		return mh.generateBinaryMath(funcName, args, g)
	case "math.sign", "sign":
		return mh.generateSign(args, g)
	case "math.todegrees", "todegrees":
		return mh.generateToDegrees(args, g)
	case "math.toradians", "toradians":
		return mh.generateToRadians(args, g)
	case "math.avg", "avg":
		return mh.generateAvg(args, g)
	case "math.random", "random":
		return mh.generateRandom(args, g)
	case "math.round_to_mintick", "round_to_mintick":
		return mh.generateRoundToMintick(args, g)
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
	goFuncName := mh.normalizeToGoMathFunc(funcName)

	return fmt.Sprintf("%s(%s)", goFuncName, arg), nil
}

func (mh *MathHandler) generateBinaryMath(funcName string, args []ast.Expression, g *generator) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("%s requires exactly 2 arguments", funcName)
	}

	arg1 := g.extractSeriesExpression(args[0])
	arg2 := g.extractSeriesExpression(args[1])
	goFuncName := mh.normalizeToGoMathFunc(funcName)

	return fmt.Sprintf("%s(%s, %s)", goFuncName, arg1, arg2), nil
}

func (mh *MathHandler) generateSign(args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("sign requires exactly 1 argument")
	}
	arg := g.extractSeriesExpression(args[0])
	/* Sign returns -1, 0, or 1 */
	return fmt.Sprintf("func() float64 { v := %s; if v > 0 { return 1 } else if v < 0 { return -1 } else { return 0 } }()", arg), nil
}

func (mh *MathHandler) generateToDegrees(args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("todegrees requires exactly 1 argument")
	}
	arg := g.extractSeriesExpression(args[0])
	return fmt.Sprintf("(%s * 180.0 / math.Pi)", arg), nil
}

func (mh *MathHandler) generateToRadians(args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("toradians requires exactly 1 argument")
	}
	arg := g.extractSeriesExpression(args[0])
	return fmt.Sprintf("(%s * math.Pi / 180.0)", arg), nil
}

func (mh *MathHandler) generateAvg(args []ast.Expression, g *generator) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("avg requires at least 1 argument")
	}
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, g.extractSeriesExpression(arg))
	}
	sum := strings.Join(argStrs, " + ")
	return fmt.Sprintf("((%s) / float64(%d))", sum, len(args)), nil
}

func (mh *MathHandler) generateRandom(args []ast.Expression, g *generator) (string, error) {
	switch len(args) {
	case 0:
		return "rand.Float64()", nil
	case 1:
		/* random(seed) - use seed but return [0,1) */
		return "rand.Float64()", nil
	case 2:
		/* random(min, max) */
		minArg := g.extractSeriesExpression(args[0])
		maxArg := g.extractSeriesExpression(args[1])
		return fmt.Sprintf("(%s + rand.Float64()*(%s - %s))", minArg, maxArg, minArg), nil
	case 3:
		/* random(min, max, seed) */
		minArg := g.extractSeriesExpression(args[0])
		maxArg := g.extractSeriesExpression(args[1])
		return fmt.Sprintf("(%s + rand.Float64()*(%s - %s))", minArg, maxArg, minArg), nil
	default:
		return "", fmt.Errorf("random requires 0-3 arguments")
	}
}

func (mh *MathHandler) generateRoundToMintick(args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("round_to_mintick requires exactly 1 argument")
	}
	arg := g.extractSeriesExpression(args[0])
	/* Round to mintick precision using ctx.Mintick */
	return fmt.Sprintf("math.Round(%s/ctx.Mintick)*ctx.Mintick", arg), nil
}
