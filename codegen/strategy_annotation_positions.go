package codegen

import "slices"

// noAnnotationPosition marks a strategy command parameter that has no positional slot.
const noAnnotationPosition = -1

// Leading argument positions of the strategy commands that precede every
// annotatable argument.
const (
	idArgument        = 0
	directionArgument = 1
	fromEntryArgument = 1
)

// strategyAnnotationPositions must stay the single source both the source
// collector and the call generators read, so a positional annotation resolves to
// the same slot on both sides.
var strategyAnnotationPositions = map[string]map[string]int{
	"strategy.entry": {
		"comment":       7,
		"alert_message": 8,
		"disable_alert": 9,
	},
	"strategy.order": {
		"comment":       7,
		"alert_message": 8,
		"disable_alert": 9,
	},
	"strategy.exit": {
		"comment":          12,
		"comment_profit":   13,
		"comment_loss":     14,
		"comment_trailing": 15,
		"alert_message":    16,
		"alert_profit":     17,
		"alert_loss":       18,
		"alert_trailing":   19,
		"disable_alert":    20,
	},
	"strategy.close": {
		"comment":       1,
		"alert_message": 4,
		"disable_alert": 6,
	},
	"strategy.close_all": {
		"comment":       0,
		"alert_message": 1,
		"disable_alert": 3,
	},
}

var strategyAnnotationParameters = []string{
	"comment", "comment_profit", "comment_loss", "comment_trailing",
	"alert_message", "alert_profit", "alert_loss", "alert_trailing",
	"disable_alert",
}

var strategyCommentParameters = []string{
	"comment", "comment_profit", "comment_loss", "comment_trailing",
}

func strategyAnnotationIndex(funcName, parameter string) int {
	if index, ok := strategyAnnotationPositions[funcName][parameter]; ok {
		return index
	}
	return noAnnotationPosition
}

func isStrategyAnnotationParameter(name string) bool {
	return slices.Contains(strategyAnnotationParameters, name)
}

func isPresentationOnlyStrategyPositionalArg(funcName string, index int) bool {
	for _, position := range strategyAnnotationPositions[funcName] {
		if position == index {
			return true
		}
	}
	return false
}
