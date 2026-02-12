package codegen

type CalendarBuiltinInfo struct {
	PineName    string
	SeriesName  string
	StructField string
}

type BuiltinIdentifierRegistry struct {
	ohlcvFields        map[string]bool
	derivedPrices      map[string]bool
	timeSeriesBuiltins map[string]bool
	calendarBuiltins   map[string]CalendarBuiltinInfo
	constantBuiltins   map[string]bool
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
		timeSeriesBuiltins: map[string]bool{
			"time": true,
		},
		calendarBuiltins: map[string]CalendarBuiltinInfo{
			"dayofweek":  {PineName: "dayofweek", SeriesName: "dayofweekSeries", StructField: "DayOfWeek"},
			"dayofmonth": {PineName: "dayofmonth", SeriesName: "dayofmonthSeries", StructField: "DayOfMonth"},
			"hour":       {PineName: "hour", SeriesName: "hourSeries", StructField: "Hour"},
			"minute":     {PineName: "minute", SeriesName: "minuteSeries", StructField: "Minute"},
			"month":      {PineName: "month", SeriesName: "monthSeries", StructField: "Month"},
			"second":     {PineName: "second", SeriesName: "secondSeries", StructField: "Second"},
			"year":       {PineName: "year", SeriesName: "yearSeries", StructField: "Year"},
			"weekofyear": {PineName: "weekofyear", SeriesName: "weekofyearSeries", StructField: "WeekOfYear"},
		},
		constantBuiltins: map[string]bool{
			"last_bar_index": true,
		},
	}
}

func (r *BuiltinIdentifierRegistry) IsBuiltinSeriesIdentifier(name string) bool {
	if r.ohlcvFields[name] || r.derivedPrices[name] || r.timeSeriesBuiltins[name] {
		return true
	}
	_, isCalendar := r.calendarBuiltins[name]
	return isCalendar
}

func (r *BuiltinIdentifierRegistry) IsDerivedPrice(name string) bool {
	return r.derivedPrices[name]
}

func (r *BuiltinIdentifierRegistry) IsOHLCVField(name string) bool {
	return r.ohlcvFields[name]
}

func (r *BuiltinIdentifierRegistry) IsTimeSeriesBuiltin(name string) bool {
	return r.timeSeriesBuiltins[name]
}

func (r *BuiltinIdentifierRegistry) IsCalendarBuiltin(name string) bool {
	_, exists := r.calendarBuiltins[name]
	return exists
}

func (r *BuiltinIdentifierRegistry) CalendarInfo(name string) (CalendarBuiltinInfo, bool) {
	info, exists := r.calendarBuiltins[name]
	return info, exists
}

func (r *BuiltinIdentifierRegistry) CalendarBuiltinNames() []string {
	names := make([]string, 0, len(r.calendarBuiltins))
	for name := range r.calendarBuiltins {
		names = append(names, name)
	}
	return names
}

func (r *BuiltinIdentifierRegistry) IsConstantBuiltin(name string) bool {
	return r.constantBuiltins[name]
}
