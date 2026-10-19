package codegen

// SeriesSuffixResolver maps a variable name to the Go suffix of its series variable.
//
//	Regular float variables  → "Series"       (varNameSeries *series.Series)
//	Float array variables    → "ArraySeries"  (varNameArraySeries *series.ArraySeries)
//	String array variables   → "StringArraySeries"
type SeriesSuffixResolver interface {
	ResolveSuffix(varName string) string
}

// FixedSeriesSuffixResolver always returns "Series".
// Used when no array registry is available (e.g. test helpers, string-variable paths).
type FixedSeriesSuffixResolver struct{}

func (r *FixedSeriesSuffixResolver) ResolveSuffix(_ string) string { return "Series" }

// ArrayRegistrySuffixResolver delegates to ArrayVariableRegistry for array types
// and falls back to "Series" for all other variables.
type ArrayRegistrySuffixResolver struct {
	registry *ArrayVariableRegistry
}

func NewArrayRegistrySuffixResolver(registry *ArrayVariableRegistry) *ArrayRegistrySuffixResolver {
	return &ArrayRegistrySuffixResolver{registry: registry}
}

func (r *ArrayRegistrySuffixResolver) ResolveSuffix(varName string) string {
	if elemType, ok := r.registry.Lookup(varName); ok {
		return elemType.VariableSuffix()
	}
	return "Series"
}
