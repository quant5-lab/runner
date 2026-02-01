package codegen

/* OHLCVFieldToSeriesName maps OHLCV field names to their ForwardSeriesBuffer variable names */
func OHLCVFieldToSeriesName(fieldName string) string {
	seriesNameMap := map[string]string{
		"Close":  "closeSeries",
		"High":   "highSeries",
		"Low":    "lowSeries",
		"Open":   "openSeries",
		"Volume": "volumeSeries",
	}
	if seriesName, exists := seriesNameMap[fieldName]; exists {
		return seriesName
	}
	return fieldName + "Series"
}
