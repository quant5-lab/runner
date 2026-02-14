package codegen

var ohlcvFieldNames = map[string]string{
	"close":  "Close",
	"open":   "Open",
	"high":   "High",
	"low":    "Low",
	"volume": "Volume",
}

func OHLCVFieldName(builtin string) (string, bool) {
	field, ok := ohlcvFieldNames[builtin]
	return field, ok
}
