package codegen

import "fmt"

type SeriesAccessCodeGenerator = AccessGenerator

// SeriesVariableAccessGenerator generates access code for user-defined Series variables.
type SeriesVariableAccessGenerator struct {
	variableName string
	baseOffset   int
}

// NewSeriesVariableAccessGenerator creates a generator for Series variable access.
func NewSeriesVariableAccessGenerator(variableName string) *SeriesVariableAccessGenerator {
	return &SeriesVariableAccessGenerator{
		variableName: variableName,
		baseOffset:   0,
	}
}

// NewSeriesVariableAccessGeneratorWithOffset creates a generator with a base offset for Series variable access.
func NewSeriesVariableAccessGeneratorWithOffset(variableName string, baseOffset int) *SeriesVariableAccessGenerator {
	return &SeriesVariableAccessGenerator{
		variableName: variableName,
		baseOffset:   baseOffset,
	}
}

func (g *SeriesVariableAccessGenerator) GenerateInitialValueAccess(period int) string {
	totalOffset := period - 1 + g.baseOffset
	return fmt.Sprintf("%sSeries.Get(%d)", g.variableName, totalOffset)
}

func (g *SeriesVariableAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("%sSeries.Get(%s)", g.variableName, loopVar)
	}
	return fmt.Sprintf("%sSeries.Get(%s+%d)", g.variableName, loopVar, g.baseOffset)
}

func (g *SeriesVariableAccessGenerator) GenerateCurrentValueAccess() string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("%sSeries.GetCurrent()", g.variableName)
	}
	return fmt.Sprintf("%sSeries.Get(%d)", g.variableName, g.baseOffset)
}

// OHLCVFieldAccessGenerator generates access code for built-in OHLCV fields.
type OHLCVFieldAccessGenerator struct {
	fieldName  string
	baseOffset int
}

// NewOHLCVFieldAccessGenerator creates a generator for OHLCV field access.
func NewOHLCVFieldAccessGenerator(fieldName string) *OHLCVFieldAccessGenerator {
	return &OHLCVFieldAccessGenerator{
		fieldName:  fieldName,
		baseOffset: 0,
	}
}

// NewOHLCVFieldAccessGeneratorWithOffset creates a generator with a base offset for OHLCV field access.
func NewOHLCVFieldAccessGeneratorWithOffset(fieldName string, baseOffset int) *OHLCVFieldAccessGenerator {
	return &OHLCVFieldAccessGenerator{
		fieldName:  fieldName,
		baseOffset: baseOffset,
	}
}

func (g *OHLCVFieldAccessGenerator) GenerateInitialValueAccess(period int) string {
	totalOffset := period - 1 + g.baseOffset
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", totalOffset, g.fieldName)
}

func (g *OHLCVFieldAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("ctx.Data[ctx.BarIndex-%s].%s", loopVar, g.fieldName)
	}
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-(%s+%d)].%s", loopVar, g.baseOffset, g.fieldName)
}

func (g *OHLCVFieldAccessGenerator) GenerateCurrentValueAccess() string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("ctx.Data[ctx.BarIndex].%s", g.fieldName)
	}
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", g.baseOffset, g.fieldName)
}

// CreateAccessGenerator creates the appropriate access generator based on source info.
func CreateAccessGenerator(source SourceInfo) SeriesAccessCodeGenerator {
	if source.IsSeriesVariable() {
		if source.BaseOffset != 0 {
			return NewSeriesVariableAccessGeneratorWithOffset(source.VariableName, source.BaseOffset)
		}
		return NewSeriesVariableAccessGenerator(source.VariableName)
	}
	return NewOHLCVFieldAccessGeneratorWithOffset(source.FieldName, source.BaseOffset)
}
