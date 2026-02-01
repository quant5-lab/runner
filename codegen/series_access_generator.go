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

func (g *SeriesVariableAccessGenerator) GetBaseOffset() int {
	return g.baseOffset
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
	seriesName := g.fieldNameToSeriesName()
	return fmt.Sprintf("%s.Get(%d)", seriesName, totalOffset)
}

func (g *OHLCVFieldAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	seriesName := g.fieldNameToSeriesName()
	if g.baseOffset == 0 {
		return fmt.Sprintf("%s.Get(%s)", seriesName, loopVar)
	}
	return fmt.Sprintf("%s.Get(%s+%d)", seriesName, loopVar, g.baseOffset)
}

func (g *OHLCVFieldAccessGenerator) GenerateCurrentValueAccess() string {
	seriesName := g.fieldNameToSeriesName()
	if g.baseOffset == 0 {
		return fmt.Sprintf("%s.GetCurrent()", seriesName)
	}
	return fmt.Sprintf("%s.Get(%d)", seriesName, g.baseOffset)
}

func (g *OHLCVFieldAccessGenerator) GetBaseOffset() int {
	return g.baseOffset
}

func (g *OHLCVFieldAccessGenerator) fieldNameToSeriesName() string {
	return OHLCVFieldToSeriesName(g.fieldName)
}

// CreateAccessGenerator creates the appropriate access generator based on source info.
func CreateAccessGenerator(source SourceInfo) SeriesAccessCodeGenerator {
	if source.IsSeriesVariable() {
		if source.BaseOffset != 0 {
			return NewSeriesVariableAccessGeneratorWithOffset(source.VariableName, source.BaseOffset)
		}
		return NewSeriesVariableAccessGenerator(source.VariableName)
	}
	if source.IsDerivedPrice() {
		return NewDerivedPriceAccessor(source.PriceName, source.BaseOffset)
	}
	return NewOHLCVFieldAccessGeneratorWithOffset(source.FieldName, source.BaseOffset)
}

/* CreatePreviousBarAccessGenerator creates accessor shifted 1 bar back for crossover previous bar calculation */
func CreatePreviousBarAccessGenerator(source SourceInfo) SeriesAccessCodeGenerator {
	if source.IsSeriesVariable() {
		return NewSeriesVariableAccessGeneratorWithOffset(source.VariableName, source.BaseOffset+1)
	}
	if source.IsDerivedPrice() {
		return NewDerivedPriceAccessor(source.PriceName, source.BaseOffset+1)
	}
	return NewOHLCVFieldAccessGeneratorWithOffset(source.FieldName, source.BaseOffset+1)
}
