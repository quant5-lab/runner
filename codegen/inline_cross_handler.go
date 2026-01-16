package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type CrossInlineHandler struct {
	isUnder bool
}

func NewCrossoverInlineHandler() *CrossInlineHandler {
	return &CrossInlineHandler{isUnder: false}
}

func NewCrossunderInlineHandler() *CrossInlineHandler {
	return &CrossInlineHandler{isUnder: true}
}

func (h *CrossInlineHandler) CanHandle(funcName string) bool {
	if h.isUnder {
		return funcName == "ta.crossunder" || funcName == "crossunder"
	}
	return funcName == "ta.crossover" || funcName == "crossover"
}

func (h *CrossInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	funcName := h.functionName()

	if len(expr.Arguments) < 2 {
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	inline1, err := g.plotExprHandler.Generate(expr.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("%s arg1 inline generation failed: %w", funcName, err)
	}

	inline2, err := g.plotExprHandler.Generate(expr.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("%s arg2 inline generation failed: %w", funcName, err)
	}

	return h.buildCrossDetectionIIFE(inline1, inline2), nil
}

func (h *CrossInlineHandler) functionName() string {
	if h.isUnder {
		return "ta.crossunder"
	}
	return "ta.crossover"
}

func (h *CrossInlineHandler) buildCrossDetectionIIFE(expr1, expr2 string) string {
	operator := h.crossOperator()
	reverseOp := h.reverseCrossOperator()

	return fmt.Sprintf(
		"(func() bool { if ctx.BarIndex == 0 { return false }; "+
			"curr1 := (%s); curr2 := (%s); "+
			"prevBarIdx := ctx.BarIndex; ctx.BarIndex--; "+
			"prev1 := (%s); prev2 := (%s); "+
			"ctx.BarIndex = prevBarIdx; "+
			"return curr1 %s curr2 && prev1 %s prev2 }())",
		expr1, expr2, expr1, expr2, operator, reverseOp,
	)
}

func (h *CrossInlineHandler) crossOperator() string {
	if h.isUnder {
		return "<"
	}
	return ">"
}

func (h *CrossInlineHandler) reverseCrossOperator() string {
	if h.isUnder {
		return ">="
	}
	return "<="
}
