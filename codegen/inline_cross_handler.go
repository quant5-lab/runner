package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* CrossInlineHandler generates inline access for crossover/crossunder in conditions.
 * Delegates to temp variable system - crossover is just another TA function.
 */
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

	// Register as temp var (standard TA function flow)
	argHash := g.exprAnalyzer.ComputeArgHash(expr)
	callInfo := CallInfo{
		Call:     expr,
		FuncName: funcName,
		ArgHash:  argHash,
	}
	varName := g.tempVarMgr.GetOrCreate(callInfo)

	// Return Series access wrapped in value.IsTrue() for boolean context
	return fmt.Sprintf("value.IsTrue(%sSeries.GetCurrent())", varName), nil
}

func (h *CrossInlineHandler) functionName() string {
	if h.isUnder {
		return "ta.crossunder"
	}
	return "ta.crossover"
}
