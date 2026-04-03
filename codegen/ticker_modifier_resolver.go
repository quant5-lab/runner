package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

var pineModifierConstants = map[string]map[string]string{
	"session": {
		"regular":  "ticker.SessionRegular",
		"extended": "ticker.SessionExtended",
	},
	"adjustment": {
		"none":      "ticker.AdjustmentNone",
		"splits":    "ticker.AdjustmentSplits",
		"dividends": "ticker.AdjustmentDividends",
	},
	"backadjustment": {
		"inherit": "ticker.BackAdjustmentInherit",
		"on":      "ticker.BackAdjustmentOn",
		"off":     "ticker.BackAdjustmentOff",
	},
	"settlement_as_close": {
		"inherit": "ticker.SettlementInherit",
		"on":      "ticker.SettlementOn",
		"off":     "ticker.SettlementOff",
	},
}

func resolveTickerModifierArg(g *generator, args []ast.Expression, index int) string {
	if index >= len(args) {
		return `""`
	}
	return resolveTickerModifierExpr(g, args[index])
}

func resolveTickerModifierExpr(g *generator, expr ast.Expression) string {
	if expr == nil {
		return `""`
	}

	if member, ok := expr.(*ast.MemberExpression); ok {
		obj, objOk := member.Object.(*ast.Identifier)
		prop, propOk := member.Property.(*ast.Identifier)
		if objOk && propOk {
			if props, nsFound := pineModifierConstants[obj.Name]; nsFound {
				if goConst, found := props[prop.Name]; found {
					return goConst
				}
			}
		}
	}

	return g.extractSeriesExpression(expr)
}
