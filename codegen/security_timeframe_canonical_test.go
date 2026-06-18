package codegen

import "testing"

func TestCanonicalizeTimeframe_NumericMinuteTokens(t *testing.T) {
	cases := []struct {
		token string
		want  string
	}{
		{"1", "1m"},
		{"5", "5m"},
		{"10", "10m"},
		{"15", "15m"},
		{"30", "30m"},
		{"45", "45m"},
		{"60", "1h"},
		{"90", "90m"},
		{"120", "2h"},
		{"180", "3h"},
		{"240", "4h"},
		{"360", "6h"},
		{"480", "8h"},
		{"720", "12h"},
		{"1440", "24h"},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			got := canonicalizeTimeframe(tc.token)
			if got != tc.want {
				t.Errorf("canonicalizeTimeframe(%q) = %q, want %q", tc.token, got, tc.want)
			}
		})
	}
}

func TestCanonicalizeTimeframe_HourBoundaryOnlyForExactMultiples(t *testing.T) {
	// Tokens that are NOT multiples of 60 must receive the "m" suffix, never "h".
	cases := []struct {
		token string
		want  string
	}{
		{"59", "59m"},
		{"61", "61m"},
		{"90", "90m"},
		{"91", "91m"},
		{"119", "119m"},
		{"121", "121m"},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			got := canonicalizeTimeframe(tc.token)
			if got != tc.want {
				t.Errorf("canonicalizeTimeframe(%q) = %q, want %q", tc.token, got, tc.want)
			}
		})
	}
}

func TestCanonicalizeTimeframe_CalendarUnitShortcuts(t *testing.T) {
	cases := []struct {
		token string
		want  string
	}{
		{"D", "1D"},
		{"W", "1W"},
		{"M", "1M"},
	}
	for _, tc := range cases {
		t.Run(tc.token, func(t *testing.T) {
			got := canonicalizeTimeframe(tc.token)
			if got != tc.want {
				t.Errorf("canonicalizeTimeframe(%q) = %q, want %q", tc.token, got, tc.want)
			}
		})
	}
}

func TestCanonicalizeTimeframe_AlreadyCanonicalTokensPassThrough(t *testing.T) {
	tokens := []string{
		"1m", "5m", "10m", "15m", "30m", "45m",
		"1h", "2h", "3h", "4h", "6h", "8h", "12h", "24h",
		"1D", "2D", "3D", "1W", "2W", "1M",
	}
	for _, token := range tokens {
		t.Run(token, func(t *testing.T) {
			got := canonicalizeTimeframe(token)
			if got != token {
				t.Errorf("canonicalizeTimeframe(%q) = %q, want unchanged %q", token, got, token)
			}
		})
	}
}

func TestCanonicalizeTimeframe_NonParseableTokensPassThrough(t *testing.T) {
	// Tokens that contain non-digit characters (other than the bare D/W/M switches)
	// must pass through unchanged — they may be valid named forms or future tokens.
	tokens := []string{
		"",    // empty
		"abc", // purely alphabetic
		"1h2", // mixed: digits then non-suffix letters
		"1.5", // decimal: not a pure integer
		"0",   // zero is not a valid period
		"00",  // zero with leading zero
		"-60", // negative sign is not a digit
		" 60", // leading space is not a digit
	}
	for _, token := range tokens {
		t.Run("token="+token, func(t *testing.T) {
			got := canonicalizeTimeframe(token)
			if got != token {
				t.Errorf("canonicalizeTimeframe(%q) = %q, want pass-through %q", token, got, token)
			}
		})
	}
}

func TestDigitsOnlyInt_ParsesAllDigitPositiveIntegers(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"9", 9},
		{"60", 60},
		{"240", 240},
		{"1440", 1440},
		{"99999", 99999},
	}
	for _, tc := range cases {
		got := digitsOnlyInt(tc.input)
		if got != tc.want {
			t.Errorf("digitsOnlyInt(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestDigitsOnlyInt_ReturnsZeroForNonPositiveOrNonDigitInput(t *testing.T) {
	// All of these should return 0: they are either empty, non-numeric,
	// contain non-digit characters, or evaluate to zero.
	cases := []string{
		"",    // empty string
		"0",   // integer zero is not a valid period count
		"00",  // zero with leading zeros
		"abc", // purely alphabetic
		"1a",  // digit followed by letter
		"a1",  // letter followed by digit
		"1.5", // decimal point
		"-1",  // minus sign
		" 1",  // leading space
		"1 ",  // trailing space
	}
	for _, s := range cases {
		got := digitsOnlyInt(s)
		if got != 0 {
			t.Errorf("digitsOnlyInt(%q) = %d, want 0", s, got)
		}
	}
}
