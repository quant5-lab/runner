package codegen

type TAArgumentSpec struct {
	Position       int
	Classification TAArgumentClassification
	DefaultValue   string
}

func NewSeriesArgument(position int, defaultValue string) TAArgumentSpec {
	classification := TAArgSeriesRequired
	if defaultValue != "" {
		classification = TAArgSeriesOptional
	}
	return TAArgumentSpec{
		Position:       position,
		Classification: classification,
		DefaultValue:   defaultValue,
	}
}

func NewScalarIntArgument(position int) TAArgumentSpec {
	return TAArgumentSpec{
		Position:       position,
		Classification: TAArgScalarInt,
	}
}

func NewScalarFloatArgument(position int) TAArgumentSpec {
	return TAArgumentSpec{
		Position:       position,
		Classification: TAArgScalarFloat,
	}
}

func NewScalarBoolArgument(position int) TAArgumentSpec {
	return TAArgumentSpec{
		Position:       position,
		Classification: TAArgScalarBool,
	}
}

func NewImplicitOHLCArgument() TAArgumentSpec {
	return TAArgumentSpec{
		Position:       -1,
		Classification: TAArgImplicitOHLC,
	}
}
