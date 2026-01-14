package codegen

/* BarFieldSeriesRegistry manages OHLCV bar field Series names */
type BarFieldSeriesRegistry struct {
	fields map[string]string
}

func NewBarFieldSeriesRegistry() *BarFieldSeriesRegistry {
	return &BarFieldSeriesRegistry{
		fields: map[string]string{
			"bar.Close":  "closeSeries",
			"bar.High":   "highSeries",
			"bar.Low":    "lowSeries",
			"bar.Open":   "openSeries",
			"bar.Volume": "volumeSeries",
		},
	}
}

func (r *BarFieldSeriesRegistry) GetSeriesName(barField string) (string, bool) {
	name, exists := r.fields[barField]
	return name, exists
}

func (r *BarFieldSeriesRegistry) AllFields() []string {
	return []string{"Close", "High", "Low", "Open", "Volume"}
}

func (r *BarFieldSeriesRegistry) AllSeriesNames() []string {
	return []string{"closeSeries", "highSeries", "lowSeries", "openSeries", "volumeSeries"}
}
