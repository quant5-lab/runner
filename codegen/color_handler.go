package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ColorHandler struct{}

func NewColorHandler() *ColorHandler {
	return &ColorHandler{}
}

func (ch *ColorHandler) CanHandle(funcName string) bool {
	return IsColorFunction(funcName)
}

/* Returns inline-ready Go expression (no trailing newline, no indentation) */
func (ch *ColorHandler) GenerateColorCall(funcName string, args []ast.Expression, g *generator) (string, error) {
	goFunc := ColorFunctionGoName(funcName)
	if goFunc == "" {
		return "", fmt.Errorf("unsupported color function: %s", funcName)
	}

	switch funcName {
	case "color.new":
		return ch.generateColorNew(goFunc, args, g)
	case "color.rgb":
		return ch.generateColorRGB(goFunc, args, g)
	case "color.from_gradient":
		return ch.generateColorFromGradient(goFunc, args, g)
	case "color.r", "color.g", "color.b", "color.t":
		return ch.generateColorComponent(goFunc, args, g)
	default:
		return "", fmt.Errorf("no generation strategy for %s", funcName)
	}
}

func (ch *ColorHandler) generateColorNew(goFunc string, args []ast.Expression, g *generator) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("color.new requires at least 1 argument")
	}

	colorArg := ch.ResolveColorExpression(args[0], g)
	transpArg := "0.0"
	if len(args) >= 2 {
		transpArg = g.extractSeriesExpression(args[1])
	}

	return fmt.Sprintf("%s(%s, %s)", goFunc, colorArg, transpArg), nil
}

func (ch *ColorHandler) generateColorRGB(goFunc string, args []ast.Expression, g *generator) (string, error) {
	if len(args) < 3 {
		return "", fmt.Errorf("color.rgb requires at least 3 arguments")
	}

	r := g.extractSeriesExpression(args[0])
	gArg := g.extractSeriesExpression(args[1])
	b := g.extractSeriesExpression(args[2])
	t := "0.0"
	if len(args) >= 4 {
		t = g.extractSeriesExpression(args[3])
	}

	return fmt.Sprintf("%s(%s, %s, %s, %s)", goFunc, r, gArg, b, t), nil
}

func (ch *ColorHandler) generateColorFromGradient(goFunc string, args []ast.Expression, g *generator) (string, error) {
	if len(args) != 5 {
		return "", fmt.Errorf("color.from_gradient requires exactly 5 arguments")
	}

	value := g.extractSeriesExpression(args[0])
	bottomVal := g.extractSeriesExpression(args[1])
	topVal := g.extractSeriesExpression(args[2])
	bottomColor := ch.ResolveColorExpression(args[3], g)
	topColor := ch.ResolveColorExpression(args[4], g)

	return fmt.Sprintf("%s(%s, %s, %s, %s, %s)", goFunc, value, bottomVal, topVal, bottomColor, topColor), nil
}

func (ch *ColorHandler) generateColorComponent(goFunc string, args []ast.Expression, g *generator) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("%s requires exactly 1 argument", goFunc)
	}

	colorArg := ch.ResolveColorExpression(args[0], g)
	return fmt.Sprintf("%s(%s)", goFunc, colorArg), nil
}

/* Returned string is ready for %s embedding: quoted for literals, unquoted for runtime calls */
func (ch *ColorHandler) ResolveColorExpression(expr ast.Expression, g *generator) string {
	switch e := expr.(type) {
	case *ast.Literal:
		if strVal, ok := e.Value.(string); ok {
			return fmt.Sprintf("%q", strVal)
		}
	case *ast.MemberExpression:
		if hex, ok := g.builtinHandler.ResolveMemberExpressionColorHex(e); ok && hex != "" {
			return fmt.Sprintf("%q", hex)
		}
	case *ast.Identifier:
		if hex, ok := g.builtinHandler.ResolveColorHex(e.Name); ok && hex != "" {
			return fmt.Sprintf("%q", hex)
		}
		return e.Name
	case *ast.CallExpression:
		funcName := g.extractFunctionName(e.Callee)
		if ch.CanHandle(funcName) {
			code, err := ch.GenerateColorCall(funcName, e.Arguments, g)
			if err == nil {
				return code
			}
		}
	}
	return g.extractSeriesExpression(expr)
}
