package codegen

// VariableType represents the type classification of a PineScript variable
type VariableType int

const (
	// VariableTypeUnknown indicates type has not been determined
	VariableTypeUnknown VariableType = iota

	// VariableTypeScalar represents simple scalar values (int, float, bool, string)
	VariableTypeScalar

	// VariableTypeSeries represents time-series values that support historical indexing
	VariableTypeSeries

	// VariableTypeFunction represents function references
	VariableTypeFunction
)

func (v VariableType) String() string {
	switch v {
	case VariableTypeScalar:
		return "scalar"
	case VariableTypeSeries:
		return "series"
	case VariableTypeFunction:
		return "function"
	default:
		return "unknown"
	}
}

func (v VariableType) IsSeries() bool {
	return v == VariableTypeSeries
}

func (v VariableType) IsScalar() bool {
	return v == VariableTypeScalar
}
