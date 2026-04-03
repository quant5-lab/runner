package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// StrategyActionHandler generates code for strategy.entry/close/cancel/order calls.
type StrategyActionHandler struct {
	qtyResolver        *EntryQuantityResolver
	conditionalWrapper *ConditionalEntryGenerator
}

func NewStrategyActionHandler() *StrategyActionHandler {
	return &StrategyActionHandler{
		qtyResolver:        NewEntryQuantityResolver(),
		conditionalWrapper: NewConditionalEntryGenerator(""),
	}
}

func (h *StrategyActionHandler) CanHandle(funcName string) bool {
	switch funcName {
	case "strategy.entry", "strategy.close", "strategy.close_all", "strategy.exit",
		"strategy.order", "strategy.cancel", "strategy.cancel_all",
		"strategy.default_entry_qty",
		"strategy.risk.allow_entry_in",
		"strategy.risk.max_cons_loss_days", "strategy.risk.max_drawdown",
		"strategy.risk.max_intraday_filled_orders", "strategy.risk.max_intraday_loss",
		"strategy.risk.max_position_size":
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
	case "strategy.order":
		return h.generateOrder(g, call)
	case "strategy.cancel":
		return h.generateCancel(g, call)
	case "strategy.cancel_all":
		return h.generateCancelAll(g, call)
	case "strategy.default_entry_qty":
		return h.generateDefaultEntryQty(g, call)
	case "strategy.risk.allow_entry_in":
		return h.generateAllowEntryIn(g, call)
	case "strategy.risk.max_cons_loss_days",
		"strategy.risk.max_drawdown",
		"strategy.risk.max_intraday_filled_orders",
		"strategy.risk.max_intraday_loss",
		"strategy.risk.max_position_size":
		return "", nil
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

	entryCode := h.generateQtyBlock(g, "Entry", "entryQty", entryID, direction, comment, qty)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, entryCode, g.ind()), nil
	}

	return entryCode, nil
}

func (h *StrategyActionHandler) generateClose(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return g.ind() + "// strategy.close() - invalid arguments\n", nil
	}

	entryID := g.extractStringLiteral(call.Arguments[0])

	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments[1:], "comment", 0, `""`)
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	closeCode := g.ind() + fmt.Sprintf("strat.Close(%q, bar.Close, bar.Time, %s)\n", entryID, comment)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, closeCode, g.ind()), nil
	}

	return closeCode, nil
}

func (h *StrategyActionHandler) generateCloseAll(g *generator, call *ast.CallExpression) (string, error) {
	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments, "comment", 0, `""`)
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	closeAllCode := g.ind() + fmt.Sprintf("strat.CloseAll(bar.Close, bar.Time, %s)\n", comment)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, closeAllCode, g.ind()), nil
	}

	return closeAllCode, nil
}

func (h *StrategyActionHandler) generateExit(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return g.ind() + "// strategy.exit() - invalid arguments\n", nil
	}

	exitID := g.extractStringLiteral(call.Arguments[0])
	fromEntry := g.extractStringLiteral(call.Arguments[1])

	extractor := &ArgumentExtractor{generator: g}
	limitExpr := extractor.ExtractNamedOrPositional(call.Arguments[2:], "limit", 3, "math.NaN()")
	stopExpr := extractor.ExtractNamedOrPositional(call.Arguments[2:], "stop", 5, "math.NaN()")
	comment := extractor.ExtractCommentArgument(call.Arguments[2:], "comment", 6, `""`)
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	exitCode := g.ind() + fmt.Sprintf("strat.ExitWithLevels(%q, %q, %s, %s, bar.High, bar.Low, bar.Close, bar.Time, %s)\n",
		exitID, fromEntry, stopExpr, limitExpr, comment)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, exitCode, g.ind()), nil
	}

	return exitCode, nil
}

func (h *StrategyActionHandler) generateOrder(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 2 {
		return g.ind() + "// strategy.order() - invalid arguments\n", nil
	}

	orderID := g.extractStringLiteral(call.Arguments[0])
	direction := g.extractDirectionConstant(call.Arguments[1])
	qty := h.qtyResolver.ResolveQuantity(call.Arguments, g.strategyConfig.DefaultQtyValue, g.extractFloatLiteral)

	extractor := &ArgumentExtractor{generator: g}
	comment := extractor.ExtractCommentArgument(call.Arguments[2:], "comment", 1, `""`)
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	orderCode := h.generateQtyBlock(g, "Order", "orderQty", orderID, direction, comment, qty)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, orderCode, g.ind()), nil
	}

	return orderCode, nil
}

func (h *StrategyActionHandler) generateQtyBlock(g *generator, method, qtyVar, id, direction, comment string, qty float64) string {
	dynamicCall := g.ind() + "\t" + fmt.Sprintf("strat.%s(%q, %s, %s, %s)\n", method, id, direction, qtyVar, comment)
	fixedCall := g.ind() + fmt.Sprintf("strat.%s(%q, %s, %.0f, %s)\n", method, id, direction, qty, comment)

	switch g.strategyConfig.DefaultQtyType {
	case "strategy.cash", "cash":
		return g.ind() + "{\n" +
			g.ind() + "\t" + fmt.Sprintf("%s := %.0f / closeSeries.GetCurrent()\n", qtyVar, qty) +
			dynamicCall +
			g.ind() + "}\n"
	case "strategy.percent_of_equity", "percent_of_equity":
		return g.ind() + "{\n" +
			g.ind() + "\t" + fmt.Sprintf("%s := (strat.Equity() * %.2f / 100) / closeSeries.GetCurrent()\n", qtyVar, qty) +
			dynamicCall +
			g.ind() + "}\n"
	case "strategy.fixed", "fixed", "":
		return fixedCall
	default:
		return g.ind() + fmt.Sprintf("// WARNING: Unknown default_qty_type '%s', using qty as fixed\n", g.strategyConfig.DefaultQtyType) +
			fixedCall
	}
}

func (h *StrategyActionHandler) generateCancel(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return g.ind() + "// strategy.cancel() - invalid arguments\n", nil
	}

	orderID := g.extractStringLiteral(call.Arguments[0])

	extractor := &ArgumentExtractor{generator: g}
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	cancelCode := g.ind() + fmt.Sprintf("strat.Cancel(%q)\n", orderID)

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, cancelCode, g.ind()), nil
	}

	return cancelCode, nil
}

func (h *StrategyActionHandler) generateCancelAll(g *generator, call *ast.CallExpression) (string, error) {
	extractor := &ArgumentExtractor{generator: g}
	whenCondition, hasWhen := extractor.ExtractWhenCondition(call.Arguments)

	cancelAllCode := g.ind() + "strat.CancelAll()\n"

	if hasWhen {
		wrapper := &ConditionalWrapperGenerator{}
		return wrapper.WrapIfNeeded(whenCondition, cancelAllCode, g.ind()), nil
	}

	return cancelAllCode, nil
}

func (h *StrategyActionHandler) generateDefaultEntryQty(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", nil
	}

	fillPriceExpr, err := g.generateExpression(call.Arguments[0])
	if err != nil {
		return "", err
	}

	return "strat.DefaultEntryQty(" + fillPriceExpr + ")", nil
}

func (h *StrategyActionHandler) generateAllowEntryIn(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return g.ind() + "// strategy.risk.allow_entry_in() - invalid arguments\n", nil
	}

	direction := g.extractDirectionConstant(call.Arguments[0])
	return g.ind() + fmt.Sprintf("strat.SetAllowedDirection(%s)\n", direction), nil
}
