package datafetcher

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/context"
)

// equivalentTimeframeTokens returns every Pine-native string token that names
// the same period, in probe order. The canonical suffixed form (e.g. "1h")
// is always first; for intraday periods (< 1 day) Pine also accepts a bare
// minute-count integer (e.g. "60"), which is appended as an alternate.
// Calendar-scale timeframes (daily and above) have no numeric alternate.
// Unrecognised tokens are returned as-is.
func equivalentTimeframeTokens(timeframe string) []string {
	secs := context.TimeframeToSeconds(timeframe)
	if secs <= 0 {
		return []string{timeframe}
	}

	canonical := context.TimeframeFromSeconds(secs)
	tokens := []string{canonical}

	const secondsPerDay = 86400
	if secs < secondsPerDay && secs%60 == 0 {
		numeric := fmt.Sprintf("%d", secs/60)
		if numeric != canonical {
			tokens = append(tokens, numeric)
		}
	}

	return tokens
}
