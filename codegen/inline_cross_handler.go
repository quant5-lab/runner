package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CrossInlineHandler generates inline expressions for ta.crossover and ta.crossunder */
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
	if len(expr.Arguments) < 2 {
		funcName := "ta.crossover"
		if h.isUnder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	arg1Call, isCall1 := expr.Arguments[0].(*ast.CallExpression)
	arg2Call, isCall2 := expr.Arguments[1].(*ast.CallExpression)

	if !isCall1 || !isCall2 {
		funcName := "ta.crossover"
		if h.isUnder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s requires CallExpression arguments for inline generation", funcName)
	}

	inline1, err := g.plotExprHandler.Generate(arg1Call)
	if err != nil {
		funcName := "ta.crossover"
		if h.isUnder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s arg1 inline generation failed: %w", funcName, err)
	}

	inline2, err := g.plotExprHandler.Generate(arg2Call)
	if err != nil {
		funcName := "ta.crossover"
		if h.isUnder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s arg2 inline generation failed: %w", funcName, err)
	}

	/* Generate IIFE that:
	 * 1. Evaluates both expressions at current bar
	 * 2. Temporarily decrements ctx.BarIndex to evaluate at previous bar
	 * 3. Compares current vs previous to detect crossover/crossunder
	 * 4. Restores ctx.BarIndex
	 */
	if h.isUnder {
		/* crossunder: curr1 < curr2 && prev1 >= prev2 (series1 crosses BELOW series2) */
		return fmt.Sprintf("(func() bool { if ctx.BarIndex == 0 { return false }; curr1 := (%s); curr2 := (%s); prevBarIdx := ctx.BarIndex; ctx.BarIndex--; prev1 := (%s); prev2 := (%s); ctx.BarIndex = prevBarIdx; return curr1 < curr2 && prev1 >= prev2 }())",
			inline1, inline2, inline1, inline2), nil
	}

	/* crossover: curr1 > curr2 && prev1 <= prev2 (series1 crosses ABOVE series2) */
	return fmt.Sprintf("(func() bool { if ctx.BarIndex == 0 { return false }; curr1 := (%s); curr2 := (%s); prevBarIdx := ctx.BarIndex; ctx.BarIndex--; prev1 := (%s); prev2 := (%s); ctx.BarIndex = prevBarIdx; return curr1 > curr2 && prev1 <= prev2 }())",
		inline1, inline2, inline1, inline2), nil
}
