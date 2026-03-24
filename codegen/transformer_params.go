package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// transformerConstructorCode returns the Go expression that constructs the
// appropriate BarTransformer for the given symbol expression and modifier prefix.
//
// When the ticker constructor arguments are all literals they are forwarded to
// the typed constructor (e.g. ticker.NewRenkoTransformer("ATR", 14.0)).
// Non-literal arguments fall back to NewTransformer(modifierType) with defaults.
func transformerConstructorCode(symbolExpr ast.Expression, modifierPrefix string) string {
	call, ok := symbolExpr.(*ast.CallExpression)
	if !ok {
		return fmt.Sprintf("ticker.NewTransformer(%q)", modifierPrefix)
	}

	switch strings.ToUpper(modifierPrefix) {
	case "HEIKINASHI":
		return "&ticker.HeikinAshiTransformer{}"

	case "RENKO":
		style, styleOK := extractStringLiteral(call, 1)
		param, paramOK := extractFloatLiteral(call, 2)
		if styleOK && paramOK {
			return fmt.Sprintf("ticker.NewRenkoTransformer(%q, %g)", style, param)
		}

	case "KAGI":
		reversal, ok := extractFloatLiteral(call, 1)
		if ok {
			return fmt.Sprintf("ticker.NewKagiTransformer(%g)", reversal)
		}

	case "LINEBREAK":
		lines, ok := extractFloatLiteral(call, 1)
		if ok {
			return fmt.Sprintf("ticker.NewLineBreakTransformer(%d)", int(lines))
		}

	case "POINTFIG":
		source, srcOK := extractStringLiteral(call, 1)
		style, styleOK := extractStringLiteral(call, 2)
		param, paramOK := extractFloatLiteral(call, 3)
		reversal, revOK := extractFloatLiteral(call, 4)
		if srcOK && styleOK && paramOK && revOK {
			return fmt.Sprintf("ticker.NewPointFigureTransformer(%q, %q, %g, %g)", source, style, param, reversal)
		}
	}

	return fmt.Sprintf("ticker.NewTransformer(%q)", modifierPrefix)
}

// isVariableBarCountModifier returns true for modifier types that produce
// a different number of bars than the source (Renko, Kagi, LineBreak, PointFig).
// HeikinAshi preserves bar count 1:1 and therefore returns false.
func isVariableBarCountModifier(prefix string) bool {
	switch strings.ToUpper(prefix) {
	case "RENKO", "KAGI", "LINEBREAK", "POINTFIG":
		return true
	}
	return false
}

func extractStringLiteral(call *ast.CallExpression, argIdx int) (string, bool) {
	if call == nil || argIdx >= len(call.Arguments) {
		return "", false
	}
	lit, ok := call.Arguments[argIdx].(*ast.Literal)
	if !ok {
		return "", false
	}
	s, ok := lit.Value.(string)
	return strings.Trim(s, "\"'"), ok
}

func extractFloatLiteral(call *ast.CallExpression, argIdx int) (float64, bool) {
	if call == nil || argIdx >= len(call.Arguments) {
		return 0, false
	}
	lit, ok := call.Arguments[argIdx].(*ast.Literal)
	if !ok {
		return 0, false
	}
	f, ok := lit.Value.(float64)
	return f, ok
}
