package codegen

type BuiltinIdentifierRegistry struct {
	ohlcvFields   map[string]bool
	derivedPrices map[string]bool
}

func NewBuiltinIdentifierRegistry() *BuiltinIdentifierRegistry {
	return &BuiltinIdentifierRegistry{
		ohlcvFields: map[string]bool{
			"close":     true,
			"open":      true,
			"high":      true,
			"low":       true,
			"volume":    true,
			"tr":        true,
			"bar_index": true,
		},
		derivedPrices: map[string]bool{
			"hl2":   true,
			"hlc3":  true,
			"ohlc4": true,
			"hlcc4": true,
		},
	}
}

func (r *BuiltinIdentifierRegistry) IsBuiltinSeriesIdentifier(name string) bool {
	return r.ohlcvFields[name] || r.derivedPrices[name]
}

func (r *BuiltinIdentifierRegistry) IsDerivedPrice(name string) bool {
	return r.derivedPrices[name]
}

func (r *BuiltinIdentifierRegistry) IsOHLCVField(name string) bool {
	return r.ohlcvFields[name]
}
