package codegen

import "fmt"

/*
BuiltinIdentifierAccessor generates ForwardSeriesBuffer access for OHLCV fields in inline TA loops.

Responsibility: Single-purpose accessor implementing AccessGenerator interface
Design: KISS - maps field names to Series.Get() patterns
*/
type BuiltinIdentifierAccessor struct {
	baseCode  string
	fieldName string
}

func NewBuiltinIdentifierAccessor(resolvedCode string) *BuiltinIdentifierAccessor {
	fieldName := extractFieldName(resolvedCode)
	return &BuiltinIdentifierAccessor{
		baseCode:  resolvedCode,
		fieldName: fieldName,
	}
}

/* GenerateLoopValueAccess generates offset-based access for loop iterations */
func (a *BuiltinIdentifierAccessor) GenerateLoopValueAccess(loopVar string) string {
	seriesName := a.fieldNameToSeriesName()
	return fmt.Sprintf("%s.Get(%s)", seriesName, loopVar)
}

/* GenerateInitialValueAccess generates access for initial value in windowed calculations */
func (a *BuiltinIdentifierAccessor) GenerateInitialValueAccess(period int) string {
	seriesName := a.fieldNameToSeriesName()
	return fmt.Sprintf("%s.Get(%d)", seriesName, period-1)
}

/* GenerateCurrentValueAccess generates access for current bar value */
func (a *BuiltinIdentifierAccessor) GenerateCurrentValueAccess() string {
	seriesName := a.fieldNameToSeriesName()
	return fmt.Sprintf("%s.GetCurrent()", seriesName)
}

/* GetPreamble returns setup code needed before accessor usage */
func (a *BuiltinIdentifierAccessor) GetPreamble() string {
	return ""
}

/* GetBaseOffset returns 0 - builtin identifier access is current bar relative */
func (a *BuiltinIdentifierAccessor) GetBaseOffset() int {
	return 0
}

func (a *BuiltinIdentifierAccessor) fieldNameToSeriesName() string {
	return OHLCVFieldToSeriesName(a.fieldName)
}

func extractFieldName(resolvedCode string) string {
	/* Extract last dotted component as field name */
	for i := len(resolvedCode) - 1; i >= 0; i-- {
		if resolvedCode[i] == '.' {
			return resolvedCode[i+1:]
		}
	}
	return resolvedCode
}
