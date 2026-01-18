package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// StrategyActionHandler generates code for Pine Script strategy actions.
//
// Handles: strategy.entry(), strategy.close(), strategy.close_all()
// Generates: strat.Entry(), strat.Close(), strat.CloseAll() calls
type StrategyActionHandler struct {
	qtyResolver        *EntryQuantityResolver
	conditionalWrapper *ConditionalEntryGenerator
}

// NewStrategyActionHandler creates a handler.
func NewStrategyActionHandler() *StrategyActionHandler {
	return &StrategyActionHandler{
		qtyResolver:        NewEntryQuantityResolver(),
		conditionalWrapper: NewConditionalEntryGenerator(""),
	}
}

func (h *StrategyActionHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "strategy.entry", "strategy.close", "strategy.close_all", "strategy.exit":
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
	case "strategy.exit":
		return h.generateExit(g, call)
	default:
		return "", nil
	}
}

func (h *StrategyActionHandler) generateEntry(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return g.ind() + "// strategy.entry() - invalid arguments\n", nil
	}

	entryID := g.extractStringLiteral(call.Arguments[0])
	direction := g.extractDirectionConstant(call.Arguments[1])
	qty := h.qtyResolver.ResolveQuantity(
		call.Arguments,
		g.strategyConfig.DefaultQtyValue,
		g.extractFloatLiteral,
	)

	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments[2:], "comment", 1, `""`)
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	/* Runtime qty calculation per PineScript spec: https://www.tradingview.com/pine-script-reference/v5/#fun_strategy */
	var entryCode string
	switch g.strategyConfig.DefaultQtyType {
	case "strategy.cash", "cash":
		entryCode = g.ind() + fmt.Sprintf("entryQty := %.0f / closeSeries.GetCurrent()\n", qty)
		entryCode += g.ind() + fmt.Sprintf("strat.Entry(%q, %s, entryQty, %s)\n", entryID, direction, comment)
	case "strategy.percent_of_equity", "percent_of_equity":
		entryCode = g.ind() + fmt.Sprintf("entryQty := (strat.Equity() * %.2f / 100) / closeSeries.GetCurrent()\n", qty)
		entryCode += g.ind() + fmt.Sprintf("strat.Entry(%q, %s, entryQty, %s)\n", entryID, direction, comment)
	case "strategy.fixed", "fixed", "":
		entryCode = g.ind() + fmt.Sprintf("strat.Entry(%q, %s, %.0f, %s)\n", entryID, direction, qty, comment)
	default:
		entryCode = g.ind() + fmt.Sprintf("// WARNING: Unknown default_qty_type '%s', using qty as fixed\n", g.strategyConfig.DefaultQtyType)
		entryCode += g.ind() + fmt.Sprintf("strat.Entry(%q, %s, %.0f, %s)\n", entryID, direction, qty, comment)
	}

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, entryCode, g.ind()), nil
	}

	return entryCode, nil
}

func (h *StrategyActionHandler) generateClose(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.close(id)
	if len(call.Arguments) < 1 {
		// Invalid call - generate TODO comment for backward compatibility
		return g.ind() + "// strategy.close() - invalid arguments\n", nil
	}

	entryID := g.extractStringLiteral(call.Arguments[0])

	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments[1:], "comment", 0, `""`)

	return g.ind() + fmt.Sprintf("strat.Close(%q, bar.Close, bar.Time, %s)\n", entryID, comment), nil
}

func (h *StrategyActionHandler) generateCloseAll(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.close_all()
	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments, "comment", 0, `""`)

	return g.ind() + fmt.Sprintf("strat.CloseAll(bar.Close, bar.Time, %s)\n", comment), nil
}

func (h *StrategyActionHandler) generateExit(g *generator, call *ast.CallExpression) (string, error) {
	// strategy.exit(id, from_entry, qty, qty_percent, profit, limit, loss, stop, ...)
	//               0   1           2    3           4       5      6     7
	if len(call.Arguments) < 2 {
		return g.ind() + "// strategy.exit() - invalid arguments\n", nil
	}

	exitID := g.extractStringLiteral(call.Arguments[0])
	fromEntry := g.extractStringLiteral(call.Arguments[1])

	extractor := &ArgumentExtractor{generator: g}
	limitExpr := extractor.ExtractNamedOrPositional(call.Arguments[2:], "limit", 3, "math.NaN()")
	stopExpr := extractor.ExtractNamedOrPositional(call.Arguments[2:], "stop", 5, "math.NaN()")
	comment := extractor.ExtractCommentArgument(call.Arguments[2:], "comment", 6, `""`)

	return g.ind() + fmt.Sprintf("strat.ExitWithLevels(%q, %q, %s, %s, bar.High, bar.Low, bar.Close, bar.Time, %s)\n",
		exitID, fromEntry, stopExpr, limitExpr, comment), nil
}
