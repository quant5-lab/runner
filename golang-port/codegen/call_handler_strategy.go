package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// StrategyActionHandler generates code for Pine Script strategy actions.
//
// Handles: strategy.entry(), strategy.close(), strategy.close_all()
// Generates: strat.Entry(), strat.Close(), strat.CloseAll() calls
type StrategyActionHandler struct{}

func (h *StrategyActionHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "strategy.entry", "strategy.close", "strategy.close_all":
		return true
	default:
		return false
	}
}

func (h *StrategyActionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "strategy.entry":
		return h.generateEntry(g, call)
	case "strategy.close":
		return h.generateClose(g, call)
	case "strategy.close_all":
		return h.generateCloseAll(g, call)
	default:
		return "", nil
	}
}

func (h *StrategyActionHandler) generateEntry(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.entry(id, direction, qty)
	if len(call.Arguments) < 2 {
		// Invalid call - generate TODO comment instead of error for backward compatibility
		return g.ind() + "// strategy.entry() - invalid arguments\n", nil
	}

	entryID := g.extractStringLiteral(call.Arguments[0])
	direction := g.extractDirectionConstant(call.Arguments[1])
	qty := 1.0
	if len(call.Arguments) >= 3 {
		qty = g.extractFloatLiteral(call.Arguments[2])
	}

	return g.ind() + fmt.Sprintf("strat.Entry(%q, %s, %.0f)\n", entryID, direction, qty), nil
}

func (h *StrategyActionHandler) generateClose(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.close(id)
	if len(call.Arguments) < 1 {
		// Invalid call - generate TODO comment for backward compatibility
		return g.ind() + "// strategy.close() - invalid arguments\n", nil
	}

	entryID := g.extractStringLiteral(call.Arguments[0])
	return g.ind() + fmt.Sprintf("strat.Close(%q, bar.Close, bar.Time)\n", entryID), nil
}

func (h *StrategyActionHandler) generateCloseAll(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.close_all()
	return g.ind() + "strat.CloseAll(bar.Close, bar.Time)\n", nil
}
