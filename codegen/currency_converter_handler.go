package codegen

import "github.com/quant5-lab/runner/ast"

type CurrencyConverterHandler struct{}

func NewCurrencyConverterHandler() *CurrencyConverterHandler {
	return &CurrencyConverterHandler{}
}

func (h *CurrencyConverterHandler) CanHandle(funcName string) bool {
	return funcName == "strategy.convert_to_account" || funcName == "strategy.convert_to_symbol"
}

func (h *CurrencyConverterHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if !h.CanHandle(funcName) {
		return "", nil
	}

	if len(call.Arguments) < 1 {
		return "", nil
	}

	valueExpr, err := g.generateExpression(call.Arguments[0])
	if err != nil {
		return "", err
	}

	switch funcName {
	case "strategy.convert_to_account":
		return "strat.ConvertToAccount(" + valueExpr + ")", nil
	case "strategy.convert_to_symbol":
		return "strat.ConvertToSymbol(" + valueExpr + ")", nil
	default:
		return "", nil
	}
}
