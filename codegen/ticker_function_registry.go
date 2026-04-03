package codegen

// Type inference runs after preprocessing, so both the canonical v5 form (ticker.X)
// and the v4 bare aliases must be covered: a v5 script may use bare forms too.
// All ticker constructors return a symbol string, never a numeric series.
var tickerConstructorFunctions = map[string]bool{
	"ticker.heikinashi":  true,
	"ticker.renko":       true,
	"ticker.kagi":        true,
	"ticker.linebreak":   true,
	"ticker.pointfigure": true,
	"ticker.range":       true,
	"ticker.new":         true,
	"ticker.modify":      true,
	"ticker.standard":    true,
	"ticker.inherit":     true,
	"heikinashi":         true,
	"heikenashi":         true,
	"renko":              true,
	"kagi":               true,
	"linebreak":          true,
	"pointfigure":        true,
	"range":              true,
}

// IsTickerConstructorFunction reports whether name is a ticker constructor that
// returns a symbol string rather than a float64 series.
func IsTickerConstructorFunction(name string) bool {
	return tickerConstructorFunctions[name]
}
