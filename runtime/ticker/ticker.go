package ticker

import (
	"fmt"
	"strings"
)

type ModifierType string

const (
	ModifierHeikinAshi ModifierType = "HEIKINASHI"
	ModifierRenko      ModifierType = "RENKO"
	ModifierKagi       ModifierType = "KAGI"
	ModifierLineBreak  ModifierType = "LINEBREAK"
	ModifierPointFig   ModifierType = "POINTFIG"
	ModifierRange      ModifierType = "RANGE"
)

var knownModifiers = []ModifierType{
	ModifierHeikinAshi,
	ModifierRenko,
	ModifierKagi,
	ModifierLineBreak,
	ModifierPointFig,
	ModifierRange,
}

func Heikinashi(symbol string) string {
	return fmt.Sprintf("%s:%s", ModifierHeikinAshi, symbol)
}

func Renko(symbol, style string, param float64) string {
	return fmt.Sprintf("%s:%s:%s:%.2f", ModifierRenko, symbol, style, param)
}

func Kagi(symbol string, reversal float64) string {
	return fmt.Sprintf("%s:%s:%.2f", ModifierKagi, symbol, reversal)
}

func LineBreak(symbol string, numberOfLines int) string {
	return fmt.Sprintf("%s:%s:%d", ModifierLineBreak, symbol, numberOfLines)
}

func Range(symbol string) string {
	return fmt.Sprintf("%s:%s", ModifierRange, symbol)
}

func ParseModifiedSymbol(tickerID string) (baseSymbol string, modifierType ModifierType, hasModifier bool) {
	stripped := StripModifiers(tickerID)
	parts := strings.Split(stripped, ":")
	if len(parts) < 2 {
		return stripped, "", false
	}

	for _, mod := range knownModifiers {
		if parts[0] == string(mod) {
			return parts[1], mod, true
		}
	}

	return stripped, "", false
}

func IsModified(tickerID string) bool {
	_, _, hasModifier := ParseModifiedSymbol(tickerID)
	return hasModifier
}

func ExtractBaseSymbol(tickerID string) string {
	baseSymbol, _, _ := ParseModifiedSymbol(tickerID)
	return baseSymbol
}
