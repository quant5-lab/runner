package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TickerFunctionHandler struct{}

func NewTickerFunctionHandler() *TickerFunctionHandler {
	return &TickerFunctionHandler{}
}

func (h *TickerFunctionHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "heikinashi", "ticker.heikinashi":
		return true
	case "renko", "ticker.renko":
		return true
	case "kagi", "ticker.kagi":
		return true
	case "linebreak", "ticker.linebreak":
		return true
	case "pointfigure", "ticker.pointfigure":
		return true
	case "range", "ticker.range":
		return true
	case "ticker.new", "ticker.modify", "ticker.standard", "ticker.inherit":
		return true
	default:
		return false
	}
}

func (h *TickerFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if !h.CanHandle(funcName) {
		return "", nil
	}

	g.hasTickerCalls = true

	switch funcName {
	case "heikinashi", "ticker.heikinashi":
		return h.generateHeikinashi(g, call)
	case "renko", "ticker.renko":
		return h.generateRenko(g, call)
	case "kagi", "ticker.kagi":
		return h.generateKagi(g, call)
	case "linebreak", "ticker.linebreak":
		return h.generateLineBreak(g, call)
	case "pointfigure", "ticker.pointfigure":
		return h.generatePointFigure(g, call)
	case "range", "ticker.range":
		return h.generateRange(g, call)
	case "ticker.new":
		return h.generateTickerNew(g, call)
	case "ticker.modify":
		return h.generateTickerModify(g, call)
	case "ticker.standard":
		return h.generateTickerStandard(g, call)
	case "ticker.inherit":
		return h.generateTickerInherit(g, call)
	default:
		return "", fmt.Errorf("unsupported ticker function: %s", funcName)
	}
}

func (h *TickerFunctionHandler) extractTickerExpression(g *generator, expr ast.Expression) (string, error) {
	switch exp := expr.(type) {
	case *ast.MemberExpression:
		if obj, ok := exp.Object.(*ast.Identifier); ok {
			if prop, ok := exp.Property.(*ast.Identifier); ok {
				if obj.Name == "syminfo" && (prop.Name == "tickerid" || prop.Name == "ticker") {
					return "ctx.Symbol", nil
				}
			}
		}
		return "", fmt.Errorf("unsupported member expression for ticker")
	case *ast.Literal:
		if s, ok := exp.Value.(string); ok {
			return fmt.Sprintf("%q", s), nil
		}
		return "", fmt.Errorf("invalid ticker literal type")
	case *ast.Identifier:
		if constVal, exists := g.constants[exp.Name]; exists {
			if strVal, ok := constVal.(string); ok {
				return fmt.Sprintf("%q", strVal), nil
			}
		}
		if varType, exists := g.variables[exp.Name]; exists && varType == "string" {
			return exp.Name, nil
		}
		return fmt.Sprintf("%q", exp.Name), nil
	default:
		return "", fmt.Errorf("unsupported ticker expression type: %T", expr)
	}
}

func (h *TickerFunctionHandler) generateHeikinashi(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("heikinashi() requires 1 argument")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ticker.Heikinashi(%s)", symbolArg), nil
}

func (h *TickerFunctionHandler) generateRenko(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("renko() requires at least 3 arguments")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	styleArg, err := h.extractTickerExpression(g, call.Arguments[1])
	if err != nil {
		return "", err
	}
	paramArg := g.extractSeriesExpression(call.Arguments[2])

	return fmt.Sprintf("ticker.Renko(%s, %s, float64(%s))", symbolArg, styleArg, paramArg), nil
}

func (h *TickerFunctionHandler) generateKagi(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("kagi() requires 2 arguments")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	reversalArg := g.extractSeriesExpression(call.Arguments[1])

	return fmt.Sprintf("ticker.Kagi(%s, float64(%s))", symbolArg, reversalArg), nil
}

func (h *TickerFunctionHandler) generateLineBreak(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("linebreak() requires 2 arguments")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	linesArg := g.extractSeriesExpression(call.Arguments[1])

	return fmt.Sprintf("ticker.LineBreak(%s, int(%s))", symbolArg, linesArg), nil
}

func (h *TickerFunctionHandler) generatePointFigure(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("pointfigure() requires at least 3 arguments (symbol, source, style)")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	sourceArg := g.extractSeriesExpression(call.Arguments[1])
	styleArg, err := h.extractTickerExpression(g, call.Arguments[2])
	if err != nil {
		return "", err
	}

	paramArg := "14"
	if len(call.Arguments) > 3 {
		paramArg = g.extractSeriesExpression(call.Arguments[3])
	}

	reversalArg := "3"
	if len(call.Arguments) > 4 {
		reversalArg = g.extractSeriesExpression(call.Arguments[4])
	}

	return fmt.Sprintf("ticker.PointFigure(%s, %s, %s, float64(%s), float64(%s))", symbolArg, sourceArg, styleArg, paramArg, reversalArg), nil
}

func (h *TickerFunctionHandler) generateRange(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ticker.range() requires at least 1 argument (symbol)")
	}

	symbolArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ticker.Range(%s)", symbolArg), nil
}

func (h *TickerFunctionHandler) generateTickerNew(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("ticker.new() requires at least 2 arguments (prefix, ticker)")
	}

	prefixArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	tickerArg, err := h.extractTickerExpression(g, call.Arguments[1])
	if err != nil {
		return "", err
	}

	sessionArg := resolveTickerModifierArg(g, call.Arguments, 2)
	adjustmentArg := resolveTickerModifierArg(g, call.Arguments, 3)
	backAdjustmentArg := resolveTickerModifierArg(g, call.Arguments, 4)
	settlementArg := resolveTickerModifierArg(g, call.Arguments, 5)

	return fmt.Sprintf("ticker.New(%s, %s, %s, %s, %s, %s)", prefixArg, tickerArg, sessionArg, adjustmentArg, backAdjustmentArg, settlementArg), nil
}

func (h *TickerFunctionHandler) generateTickerModify(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("ticker.modify() requires at least 1 argument (tickerid)")
	}

	tickeridArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}

	sessionArg := resolveTickerModifierArg(g, call.Arguments, 1)
	adjustmentArg := resolveTickerModifierArg(g, call.Arguments, 2)
	backAdjustmentArg := resolveTickerModifierArg(g, call.Arguments, 3)
	settlementArg := resolveTickerModifierArg(g, call.Arguments, 4)

	return fmt.Sprintf("ticker.Modify(%s, %s, %s, %s, %s)", tickeridArg, sessionArg, adjustmentArg, backAdjustmentArg, settlementArg), nil
}

func (h *TickerFunctionHandler) generateTickerStandard(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "ctx.Symbol", nil
	}

	tickerIDArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ticker.Standard(%s)", tickerIDArg), nil
}

func (h *TickerFunctionHandler) generateTickerInherit(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("ticker.inherit() requires 2 arguments (from_tickerid, symbol)")
	}

	fromArg, err := h.extractTickerExpression(g, call.Arguments[0])
	if err != nil {
		return "", err
	}
	symbolArg, err := h.extractTickerExpression(g, call.Arguments[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ticker.Inherit(%s, %s)", fromArg, symbolArg), nil
}
