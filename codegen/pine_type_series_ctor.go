package codegen

const (
	seriesCtorBool    = "series.NewBoolSeries"
	seriesCtorDefault = "series.NewSeries"
)

// SeriesCtorForType returns the Series constructor name that matches the
// pre-history default mandated by Pine's type system for the given varType:
//
//   - "bool" → NewBoolSeries  (Pine: bool is never na, pre-history = false)
//   - everything else → NewSeries  (Pine: pre-history = na)
func SeriesCtorForType(varType string) string {
	if varType == "bool" {
		return seriesCtorBool
	}
	return seriesCtorDefault
}
