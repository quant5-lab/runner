package codegen

import "fmt"

// SeriesAccessCodeGenerator generates Go code for accessing series data sources.
type SeriesAccessCodeGenerator interface {
	GenerateInitialValueAccess(period int) string
	GenerateLoopValueAccess(loopVar string) string
}

// SeriesVariableAccessGenerator generates access code for user-defined Series variables.
type SeriesVariableAccessGenerator struct {
	variableName string
}

// NewSeriesVariableAccessGenerator creates a generator for Series variable access.
func NewSeriesVariableAccessGenerator(variableName string) *SeriesVariableAccessGenerator {
	return &SeriesVariableAccessGenerator{
		variableName: variableName,
	}
}

// GenerateInitialValueAccess returns code to access the initial value for windowed calculations.
func (g *SeriesVariableAccessGenerator) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("%sSeries.Get(%d-1)", g.variableName, period)
}

// GenerateLoopValueAccess returns code to access values within a loop.
func (g *SeriesVariableAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", g.variableName, loopVar)
}

// OHLCVFieldAccessGenerator generates access code for built-in OHLCV fields.
type OHLCVFieldAccessGenerator struct {
	fieldName string
}

// NewOHLCVFieldAccessGenerator creates a generator for OHLCV field access.
func NewOHLCVFieldAccessGenerator(fieldName string) *OHLCVFieldAccessGenerator {
	return &OHLCVFieldAccessGenerator{
		fieldName: fieldName,
	}
}

// GenerateInitialValueAccess returns code to access the initial OHLCV field value.
func (g *OHLCVFieldAccessGenerator) GenerateInitialValueAccess(period int) string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-(%d-1)].%s", period, g.fieldName)
}

// GenerateLoopValueAccess returns code to access OHLCV field values within a loop.
func (g *OHLCVFieldAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%s].%s", loopVar, g.fieldName)
}

// CreateAccessGenerator creates the appropriate access generator based on source info.
func CreateAccessGenerator(source SourceInfo) SeriesAccessCodeGenerator {
	if source.IsSeriesVariable() {
		return NewSeriesVariableAccessGenerator(source.VariableName)
	}
	return NewOHLCVFieldAccessGenerator(source.FieldName)
}
