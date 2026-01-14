package codegen

import (
	"fmt"
	"regexp"
	"strings"
)

// SeriesAccessor determines how to access data from different source types
type SeriesAccessor interface {
	// IsApplicable checks if this accessor handles the given source expression
	IsApplicable(sourceExpr string) bool

	// GetAccessExpression returns the Go code to access data at given offset
	GetAccessExpression(offset string) string

	// GetSourceIdentifier returns the underlying source identifier (for Series: variable name, for OHLCV: field name)
	GetSourceIdentifier() string

	// RequiresNaNCheck indicates whether NaN checks are needed
	RequiresNaNCheck() bool
}

// SeriesVariableAccessor handles user-defined Series variables
type SeriesVariableAccessor struct {
	variableName string
}

func NewSeriesVariableAccessor(sourceExpr string) *SeriesVariableAccessor {
	re := regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)Series\.Get\(`)
	if matches := re.FindStringSubmatch(sourceExpr); len(matches) == 2 {
		return &SeriesVariableAccessor{variableName: matches[1]}
	}
	return nil
}

func (a *SeriesVariableAccessor) IsApplicable(sourceExpr string) bool {
	return NewSeriesVariableAccessor(sourceExpr) != nil
}

func (a *SeriesVariableAccessor) GetAccessExpression(offset string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", a.variableName, offset)
}

func (a *SeriesVariableAccessor) GetSourceIdentifier() string {
	return a.variableName
}

func (a *SeriesVariableAccessor) RequiresNaNCheck() bool {
	return true
}

// OHLCVFieldAccessor handles built-in OHLCV fields
type OHLCVFieldAccessor struct {
	fieldName string
}

func NewOHLCVFieldAccessor(sourceExpr string) *OHLCVFieldAccessor {
	var fieldName string
	if strings.Contains(sourceExpr, ".") {
		parts := strings.Split(sourceExpr, ".")
		fieldName = parts[len(parts)-1]
	} else {
		fieldName = sourceExpr
	}
	return &OHLCVFieldAccessor{fieldName: fieldName}
}

func (a *OHLCVFieldAccessor) IsApplicable(sourceExpr string) bool {
	// OHLCV accessor is the fallback - always applicable
	return true
}

func (a *OHLCVFieldAccessor) GetAccessExpression(offset string) string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%s].%s", offset, a.fieldName)
}

func (a *OHLCVFieldAccessor) GetSourceIdentifier() string {
	return a.fieldName
}

func (a *OHLCVFieldAccessor) RequiresNaNCheck() bool {
	return false
}

// CreateSeriesAccessor factory function that returns appropriate accessor
func CreateSeriesAccessor(sourceExpr string) SeriesAccessor {
	// Try Series variable first (more specific)
	if accessor := NewSeriesVariableAccessor(sourceExpr); accessor != nil {
		return accessor
	}

	// Fallback to OHLCV field
	return NewOHLCVFieldAccessor(sourceExpr)
}
