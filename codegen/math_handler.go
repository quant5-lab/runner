package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type MathHandler struct {
	registry *MathFunctionRegistry
}

func NewMathHandler() *MathHandler {
	return &MathHandler{
		registry: NewMathFunctionRegistry(),
	}
}

func (mh *MathHandler) normalizeToGoMathFunc(pineFuncName string) string {
	normalized := strings.ToLower(pineFuncName)
	if strings.HasPrefix(normalized, "math.") {
		shortName := normalized[5:]
		return "math." + strings.ToUpper(shortName[:1]) + shortName[1:]
	}
	return "math." + strings.ToUpper(normalized[:1]) + normalized[1:]
}

func (mh *MathHandler) CanHandle(funcName string) bool {
	_, ok := mh.registry.Lookup(funcName)
	return ok
}

func (mh *MathHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	funcName := g.extractFunctionName(expr.Callee)
	return mh.GenerateMathCall(funcName, expr.Arguments, g)
}

func (mh *MathHandler) GenerateMathCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	spec, ok := mh.registry.Lookup(funcName)
	if !ok {
		return "", fmt.Errorf("unsupported math function: %s", funcName)
	}

	switch spec.GeneratorMethod {
	case "unary":
		return mh.generateUnaryMath(funcName, args, g)
	case "binary":
		return mh.generateBinaryMath(funcName, args, g)
	case "sign":
		return mh.generateSign(args, g)
	case "todegrees":
		return mh.generateToDegrees(args, g)
	case "toradians":
		return mh.generateToRadians(args, g)
	case "avg":
		return mh.generateAvg(args, g)
	case "random":
		return mh.generateRandom(args, g)
	case "round_to_mintick":
		return mh.generateRoundToMintick(args, g)
	default:
		return "", fmt.Errorf("no generation strategy for %s", funcName)
	}
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
		return "rand.Float64()", nil
	case 2:
		minArg := g.extractSeriesExpression(args[0])
		maxArg := g.extractSeriesExpression(args[1])
		return fmt.Sprintf("(%s + rand.Float64()*(%s - %s))", minArg, maxArg, minArg), nil
	case 3:
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
	return fmt.Sprintf("math.Round(%s/ctx.Mintick)*ctx.Mintick", arg), nil
}
