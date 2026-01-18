package codegen

import "fmt"

/*
BuiltinIdentifierAccessor provides access to builtin identifiers (high, low, close, etc.) in inline TA loops.

Responsibility (SRP):
  - Single purpose: generate loop-based access for builtin OHLCV fields
  - No knowledge of identifier resolution or expression evaluation
  - Uses pre-resolved builtin code as template

Design:
  - Implements AccessGenerator interface for compatibility with inline TA generators
  - Adapts current-bar access code (ctx.Data[ctx.BarIndex].High) to offset-based access
  - KISS: simple string manipulation, no complex logic
*/
type BuiltinIdentifierAccessor struct {
	baseCode  string // Pre-resolved builtin code (e.g., "ctx.Data[ctx.BarIndex].High")
	fieldName string // Extracted field name (e.g., "High")
}

func NewBuiltinIdentifierAccessor(resolvedCode string) *BuiltinIdentifierAccessor {
	fieldName := extractFieldName(resolvedCode)
	return &BuiltinIdentifierAccessor{
		baseCode:  resolvedCode,
		fieldName: fieldName,
	}
}

/*
GenerateLoopValueAccess generates offset-based access for loop iterations.
*/
func (a *BuiltinIdentifierAccessor) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%s].%s", loopVar, a.fieldName)
}

/*
GenerateInitialValueAccess generates access for initial value in windowed calculations.
*/
func (a *BuiltinIdentifierAccessor) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", period-1, a.fieldName)
}

/*
GenerateCurrentValueAccess generates access for the current bar's value.
*/
func (a *BuiltinIdentifierAccessor) GenerateCurrentValueAccess() string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex].%s", a.fieldName)
}

/*
GetPreamble returns any setup code needed before the accessor is used.
*/
func (a *BuiltinIdentifierAccessor) GetPreamble() string {
	return ""
}

func extractFieldName(resolvedCode string) string {
	// Extract field name from "ctx.Data[ctx.BarIndex].High" → "High"
	// This is a simple heuristic - assumes last dotted component is the field name
	for i := len(resolvedCode) - 1; i >= 0; i-- {
		if resolvedCode[i] == '.' {
			return resolvedCode[i+1:]
		}
	}
	return resolvedCode
}
