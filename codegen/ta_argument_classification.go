package codegen

type TAArgumentClassification int

const (
	TAArgSeriesRequired TAArgumentClassification = iota
	TAArgSeriesOptional
	TAArgScalarInt
	TAArgScalarFloat
	TAArgScalarBool
	TAArgImplicitOHLC
)

func (c TAArgumentClassification) IsSeries() bool {
	return c == TAArgSeriesRequired || c == TAArgSeriesOptional
}

func (c TAArgumentClassification) IsScalar() bool {
	return c == TAArgScalarInt || c == TAArgScalarFloat || c == TAArgScalarBool
}

func (c TAArgumentClassification) RequiresHistoricalAccess() bool {
	return c == TAArgSeriesRequired || c == TAArgSeriesOptional || c == TAArgImplicitOHLC
}
