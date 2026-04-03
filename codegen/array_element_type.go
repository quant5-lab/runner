package codegen

type ArrayElementType int

const (
	ArrayElementFloat64 ArrayElementType = iota
	ArrayElementString
)

func (t ArrayElementType) IsNumeric() bool {
	return t == ArrayElementFloat64
}

func (t ArrayElementType) IsString() bool {
	return t == ArrayElementString
}

func (t ArrayElementType) GoType() string {
	switch t {
	case ArrayElementString:
		return "string"
	default:
		return "float64"
	}
}

func (t ArrayElementType) GoSliceType() string {
	return "[]" + t.GoType()
}

func (t ArrayElementType) SeriesType() string {
	switch t {
	case ArrayElementString:
		return "*series.StringArraySeries"
	default:
		return "*series.ArraySeries"
	}
}

func (t ArrayElementType) NewSeriesCall(capacity string) string {
	switch t {
	case ArrayElementString:
		return "series.NewStringArraySeries(" + capacity + ")"
	default:
		return "series.NewArraySeries(" + capacity + ")"
	}
}

func (t ArrayElementType) VariableSuffix() string {
	switch t {
	case ArrayElementString:
		return "StringArraySeries"
	default:
		return "ArraySeries"
	}
}

func (t ArrayElementType) TypeTag() string {
	switch t {
	case ArrayElementString:
		return "array_series_string"
	default:
		return "array_series_float"
	}
}

func (t ArrayElementType) MutatorConstructor() string {
	switch t {
	case ArrayElementString:
		return "arrayops.NewStringMutator()"
	default:
		return "arrayops.NewMutator()"
	}
}

func (t ArrayElementType) AccessorConstructor() string {
	switch t {
	case ArrayElementString:
		return "arrayops.NewStringAccessor()"
	default:
		return "arrayops.NewAccessor()"
	}
}

func (t ArrayElementType) TransformerConstructor() string {
	switch t {
	case ArrayElementString:
		return "arrayops.NewStringTransformers()"
	default:
		return "arrayops.NewTransformer()"
	}
}

func (t ArrayElementType) StatisticsConstructor() string {
	return "arrayops.NewStatistics()"
}

func (t ArrayElementType) SearchConstructor() string {
	return "arrayops.NewSearch()"
}

func (t ArrayElementType) PredicatesConstructor() string {
	return "arrayops.NewPredicates()"
}

func (t ArrayElementType) FormattersConstructor() string {
	switch t {
	case ArrayElementString:
		return "arrayops.NewStringFormatters()"
	default:
		return "arrayops.NewFormatters()"
	}
}

func (t ArrayElementType) SupportsStatistics() bool {
	return t.IsNumeric()
}

func ParseArrayElementType(typeTag string) (ArrayElementType, bool) {
	switch typeTag {
	case "array_series_float":
		return ArrayElementFloat64, true
	case "array_series_string":
		return ArrayElementString, true
	default:
		return 0, false
	}
}
