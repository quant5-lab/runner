package codegen

const (
	SessionIsFirstBarSeriesName        = "session_isfirstbarSeries"
	SessionIsLastBarSeriesName         = "session_islastbarSeries"
	SessionIsFirstBarRegularSeriesName = "session_isfirstbar_regularSeries"
	SessionIsLastBarRegularSeriesName  = "session_islastbar_regularSeries"
)

var sessionSeriesNamesByKey = map[string]string{
	"session.isfirstbar":         SessionIsFirstBarSeriesName,
	"session.islastbar":          SessionIsLastBarSeriesName,
	"session.isfirstbar_regular": SessionIsFirstBarRegularSeriesName,
	"session.islastbar_regular":  SessionIsLastBarRegularSeriesName,
}

func SessionSeriesName(key string) string {
	return sessionSeriesNamesByKey[key]
}
