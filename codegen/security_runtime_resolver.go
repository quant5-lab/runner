package codegen

import "strings"

func isRuntimeSymbol(symbol string) bool {
	/* Check for runtime symbols including those with modifier prefixes (e.g., HEIKINASHI:syminfo.tickerid) */
	runtimeSymbols := []string{"syminfo.tickerid", "syminfo.ticker", "tickerid", "ticker"}
	for _, rs := range runtimeSymbols {
		if symbol == rs || strings.HasSuffix(symbol, ":"+rs) {
			return true
		}
	}
	return false
}

func isRuntimeTimeframe(timeframe string) bool {
	return timeframe == "timeframe.period"
}

func runtimePlaceholder() string {
	return "%s"
}

/* extractModifierPrefix returns the modifier prefix and whether it exists */
func extractModifierPrefix(symbol string) (prefix string, baseSymbol string, hasModifier bool) {
	modifiers := []string{"HEIKINASHI:", "RENKO:", "KAGI:", "LINEBREAK:", "POINTFIG:"}
	for _, mod := range modifiers {
		if strings.HasPrefix(symbol, mod) {
			return strings.TrimSuffix(mod, ":"), strings.TrimPrefix(symbol, mod), true
		}
	}
	return "", symbol, false
}

/* generateModifierCall returns Go code for applying modifier (e.g., ticker.Heikinashi(ctx.Symbol)) */
func generateModifierCall(prefix string, baseSymbolCode string) string {
	switch prefix {
	case "HEIKINASHI":
		return "ticker.Heikinashi(" + baseSymbolCode + ")"
	case "RENKO":
		return "ticker.Renko(" + baseSymbolCode + ")"
	case "KAGI":
		return "ticker.Kagi(" + baseSymbolCode + ")"
	case "LINEBREAK":
		return "ticker.Linebreak(" + baseSymbolCode + ")"
	default:
		return baseSymbolCode
	}
}
