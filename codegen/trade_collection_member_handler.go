package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TradeCollectionMemberHandler struct {
	propertyMap map[string]string
}

func NewTradeCollectionMemberHandler() *TradeCollectionMemberHandler {
	return &TradeCollectionMemberHandler{
		propertyMap: buildTradePropertyMap(),
	}
}

func buildTradePropertyMap() map[string]string {
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
		"max_drawdown_percent": "MaxDrawdown",
		"max_runup":            "MaxRunup",
		"max_runup_percent":    "MaxRunup",
		"profit":               "Profit",
		"profit_percent":       "ProfitPercent",
		"size":                 "Size",
	}
}

func (h *TradeCollectionMemberHandler) CanHandle(object string, member string) bool {
	if object != "closedtrades" && object != "opentrades" {
		return false
	}
	_, exists := h.propertyMap[member]
	return exists
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

	methodSuffix, exists := h.propertyMap[member]
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
