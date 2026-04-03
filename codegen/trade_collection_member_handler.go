package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TradeCollectionMemberHandler struct {
	closedTradesProperties map[string]string
	openTradesProperties   map[string]string
}

func NewTradeCollectionMemberHandler() *TradeCollectionMemberHandler {
	return &TradeCollectionMemberHandler{
		closedTradesProperties: buildClosedTradesPropertyMap(),
		openTradesProperties:   buildOpenTradesPropertyMap(),
	}
}

func buildClosedTradesPropertyMap() map[string]string {
	return map[string]string{
		"commission":           "Commission",
		"entry_bar_index":      "EntryBarIndex",
		"entry_comment":        "EntryComment",
		"entry_id":             "EntryID",
		"entry_price":          "EntryPrice",
		"entry_time":           "EntryTime",
		"exit_bar_index":       "ExitBarIndex",
		"exit_comment":         "ExitComment",
		"exit_id":              "ExitID",
		"exit_price":           "ExitPrice",
		"exit_time":            "ExitTime",
		"max_drawdown":         "MaxDrawdown",
		"max_drawdown_percent": "MaxDrawdownPercent",
		"max_runup":            "MaxRunup",
		"max_runup_percent":    "MaxRunupPercent",
		"profit":               "Profit",
		"profit_percent":       "ProfitPercent",
		"size":                 "Size",
	}
}

func buildOpenTradesPropertyMap() map[string]string {
	return map[string]string{
		"commission":           "Commission",
		"entry_bar_index":      "EntryBarIndex",
		"entry_comment":        "EntryComment",
		"entry_id":             "EntryID",
		"entry_price":          "EntryPrice",
		"entry_time":           "EntryTime",
		"max_drawdown":         "MaxDrawdown",
		"max_drawdown_percent": "MaxDrawdownPercent",
		"max_runup":            "MaxRunup",
		"max_runup_percent":    "MaxRunupPercent",
		"profit":               "Profit",
		"profit_percent":       "ProfitPercent",
		"size":                 "Size",
	}
}

func (h *TradeCollectionMemberHandler) propertiesFor(collection string) map[string]string {
	if collection == "closedtrades" {
		return h.closedTradesProperties
	}
	return h.openTradesProperties
}

func (h *TradeCollectionMemberHandler) CanHandle(object string, member string) bool {
	switch object {
	case "closedtrades":
		_, ok := h.closedTradesProperties[member]
		return ok
	case "opentrades":
		_, ok := h.openTradesProperties[member]
		return ok
	default:
		return false
	}
}

func (h *TradeCollectionMemberHandler) GenerateAccess(
	object string,
	member string,
	args []ast.Expression,
	exprGen ExpressionGenerator,
) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("strategy.%s.%s() requires exactly 1 argument (trade_num)", object, member)
	}

	properties := h.propertiesFor(object)
	methodSuffix, exists := properties[member]
	if !exists {
		return "", fmt.Errorf("unknown trade property: %s", member)
	}

	indexExpr, err := exprGen.Generate(args[0])
	if err != nil {
		return "", fmt.Errorf("failed to generate trade index expression: %w", err)
	}

	var collectionType string
	if object == "closedtrades" {
		collectionType = "ClosedTrade"
	} else {
		collectionType = "OpenTrade"
	}

	code := fmt.Sprintf("tradeAccessor.%s%s(int(%s))", collectionType, methodSuffix, indexExpr)
	return code, nil
}
