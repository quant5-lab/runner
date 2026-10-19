package codegen

import "fmt"

// canonicalizeTimeframe normalises a Pine timeframe token to the canonical form
// used throughout the runner.  Three classes of input are handled:
//
//	Bare calendar units:   "D" → "1D",  "W" → "1W",  "M" → "1M"
//	Pure-numeric (minutes): "240" → "4h", "60" → "1h", "30" → "30m"
//	Already-canonical:     "4h", "1D", "30m" → unchanged
//
// TradingView accepts bare numeric strings as minute counts in security()
// timeframe arguments.  Normalising them to the suffixed form ensures that
// context.TimeframeToSeconds produces the correct period and that dedup-map
// keys are stable regardless of which notation the Pine source used.
func canonicalizeTimeframe(token string) string {
	switch token {
	case "D":
		return "1D"
	case "W":
		return "1W"
	case "M":
		return "1M"
	}
	return canonicalizeMinuteToken(token)
}

// canonicalizeMinuteToken converts a pure-numeric TradingView minute-count token
// to the canonical suffixed form.  Strings that contain any non-digit character
// are returned unchanged — they are already in suffixed or named form.
func canonicalizeMinuteToken(token string) string {
	minutes := digitsOnlyInt(token)
	if minutes <= 0 {
		return token
	}
	if minutes%60 == 0 {
		return fmt.Sprintf("%dh", minutes/60)
	}
	return fmt.Sprintf("%dm", minutes)
}

// digitsOnlyInt parses a string as a positive integer if and only if every
// character is an ASCII digit.  Returns 0 for empty strings, strings with
// non-digit characters, or strings whose numeric value is zero.
func digitsOnlyInt(s string) int {
	if len(s) == 0 {
		return 0
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
