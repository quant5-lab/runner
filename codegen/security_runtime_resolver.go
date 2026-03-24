package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

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

// generateModifierCall returns Go code that produces the runtime security key prefix for the
// given modifier. symbolExpr is the full ticker constructor call (e.g. ticker.renko(...)) from
// the AST; when all modifier params are literals they are forwarded so the key matches the
// bar-loop lookup exactly.  Non-literal params fall back to a best-effort key.
func generateModifierCall(prefix string, baseSymbolCode string, symbolExpr ast.Expression) string {
	call, _ := symbolExpr.(*ast.CallExpression)

	switch prefix {
	case "HEIKINASHI":
		return "ticker.Heikinashi(" + baseSymbolCode + ")"

	case "RENKO":
		style, styleOK := extractStringLiteral(call, 1)
		param, paramOK := extractFloatLiteral(call, 2)
		if styleOK && paramOK {
			return fmt.Sprintf("ticker.Renko(%s, %q, %g)", baseSymbolCode, style, param)
		}
		return fmt.Sprintf("fmt.Sprintf(\"RENKO:%%s\", %s)", baseSymbolCode)

	case "KAGI":
		reversal, ok := extractFloatLiteral(call, 1)
		if ok {
			return fmt.Sprintf("ticker.Kagi(%s, %g)", baseSymbolCode, reversal)
		}
		return fmt.Sprintf("fmt.Sprintf(\"KAGI:%%s\", %s)", baseSymbolCode)

	case "LINEBREAK":
		lines, ok := extractFloatLiteral(call, 1)
		if ok {
			return fmt.Sprintf("ticker.LineBreak(%s, %d)", baseSymbolCode, int(lines))
		}
		return fmt.Sprintf("fmt.Sprintf(\"LINEBREAK:%%s\", %s)", baseSymbolCode)

	case "POINTFIG":
		source, srcOK := extractStringLiteral(call, 1)
		style, styleOK := extractStringLiteral(call, 2)
		param, paramOK := extractFloatLiteral(call, 3)
		reversal, revOK := extractFloatLiteral(call, 4)
		if srcOK && styleOK && paramOK && revOK {
			return fmt.Sprintf("ticker.PointFigure(%s, %q, %q, %g, %g)", baseSymbolCode, source, style, param, reversal)
		}
		return fmt.Sprintf("fmt.Sprintf(\"POINTFIG:%%s\", %s)", baseSymbolCode)
	}

	return baseSymbolCode
}
