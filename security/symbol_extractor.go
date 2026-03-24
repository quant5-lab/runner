package security

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/ticker"
)

/* SymbolExtractor handles symbol extraction from all expression types */
type SymbolExtractor struct{}

func NewSymbolExtractor() *SymbolExtractor {
	return &SymbolExtractor{}
}

/* Extract returns base symbol and modifier type from expression */
func (e *SymbolExtractor) Extract(expr ast.Expression) (symbol string, modifierType ticker.ModifierType) {
	rawSymbol := e.extractRaw(expr)
	if rawSymbol == "" {
		return "", ""
	}

	baseSymbol, modType, hasModifier := ticker.ParseModifiedSymbol(rawSymbol)
	if hasModifier {
		return baseSymbol, modType
	}

	return rawSymbol, ""
}

/* extractRaw extracts raw symbol string from expression */
func (e *SymbolExtractor) extractRaw(expr ast.Expression) string {
	switch exp := expr.(type) {
	case *ast.Literal:
		if s, ok := exp.Value.(string); ok {
			return strings.Trim(s, "\"'")
		}

	case *ast.Identifier:
		return exp.Name

	case *ast.MemberExpression:
		obj := extractIdentifier(exp.Object)
		prop := extractIdentifier(exp.Property)
		if obj != "" && prop != "" {
			return obj + "." + prop
		}

	case *ast.CallExpression:
		return e.extractFromTickerCall(exp)
	}

	return ""
}

/* extractFromTickerCall extracts resulting symbol from ticker modifier functions */
func (e *SymbolExtractor) extractFromTickerCall(call *ast.CallExpression) string {
	funcName := extractFunctionName(call.Callee)

	switch funcName {
	case "heikinashi", "ticker.heikinashi":
		if len(call.Arguments) >= 1 {
			baseSymbol := e.extractRaw(call.Arguments[0])
			if baseSymbol != "" {
				return string(ticker.ModifierHeikinAshi) + ":" + baseSymbol
			}
		}

	case "renko", "ticker.renko":
		if len(call.Arguments) >= 3 {
			baseSymbol := e.extractRaw(call.Arguments[0])
			if baseSymbol != "" {
				return string(ticker.ModifierRenko) + ":" + baseSymbol
			}
		}

	case "kagi", "ticker.kagi":
		if len(call.Arguments) >= 2 {
			baseSymbol := e.extractRaw(call.Arguments[0])
			if baseSymbol != "" {
				return string(ticker.ModifierKagi) + ":" + baseSymbol
			}
		}

	case "linebreak", "ticker.linebreak":
		if len(call.Arguments) >= 2 {
			baseSymbol := e.extractRaw(call.Arguments[0])
			if baseSymbol != "" {
				return string(ticker.ModifierLineBreak) + ":" + baseSymbol
			}
		}

	case "pointfigure", "ticker.pointfigure":
		if len(call.Arguments) >= 5 {
			baseSymbol := e.extractRaw(call.Arguments[0])
			if baseSymbol != "" {
				return string(ticker.ModifierPointFig) + ":" + baseSymbol
			}
		}

	case "ticker.standard":
		if len(call.Arguments) >= 1 {
			modifiedSymbol := e.extractRaw(call.Arguments[0])
			return ticker.ExtractBaseSymbol(modifiedSymbol)
		}
	}

	return ""
}
