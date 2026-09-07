package datafetcher

import (
	"testing"
)

func TestEquivalentTimeframeTokens(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		// Intraday suffixed — canonical first, then pure-minute-count alternate.
		{"1m", []string{"1m", "1"}},
		{"5m", []string{"5m", "5"}},
		{"15m", []string{"15m", "15"}},
		{"30m", []string{"30m", "30"}},
		{"1h", []string{"1h", "60"}},
		{"2h", []string{"2h", "120"}},
		{"3h", []string{"3h", "180"}},
		{"4h", []string{"4h", "240"}},
		// Intraday numeric — normalises to canonical first, same alternate set.
		{"1", []string{"1m", "1"}},
		{"5", []string{"5m", "5"}},
		{"15", []string{"15m", "15"}},
		{"30", []string{"30m", "30"}},
		{"60", []string{"1h", "60"}},
		{"120", []string{"2h", "120"}},
		{"180", []string{"3h", "180"}},
		{"240", []string{"4h", "240"}},
		// Sub-minute — not a whole number of minutes, so no numeric alternate.
		{"1s", []string{"1s"}},
		{"5s", []string{"5s"}},
		// Calendar-scale — daily and above have no minute-count alternate.
		{"1D", []string{"1D"}},
		{"2D", []string{"2D"}},
		{"1W", []string{"1W"}},
		{"1M", []string{"1M"}},
		{"12M", []string{"12M"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			got := equivalentTimeframeTokens(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("[%d]: got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

// TestEquivalentTimeframeTokens_SymmetricEquivalence verifies that two tokens
// denoting the same period produce identical slices regardless of which form
// is supplied — the canonical form is always first.
func TestEquivalentTimeframeTokens_SymmetricEquivalence(t *testing.T) {
	pairs := [][2]string{
		{"1m", "1"},
		{"5m", "5"},
		{"15m", "15"},
		{"30m", "30"},
		{"1h", "60"},
		{"2h", "120"},
		{"3h", "180"},
		{"4h", "240"},
	}
	for _, pair := range pairs {
		a := equivalentTimeframeTokens(pair[0])
		b := equivalentTimeframeTokens(pair[1])
		if len(a) != len(b) {
			t.Errorf("(%q,%q) length mismatch: %v vs %v", pair[0], pair[1], a, b)
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("(%q,%q)[%d]: %q vs %q", pair[0], pair[1], i, a[i], b[i])
			}
		}
	}
}

// TestEquivalentTimeframeTokens_UnknownTokenPassthrough verifies that tokens
// unrecognised by TimeframeToSeconds (secs == 0) are returned as a single-element
// slice containing the original token unchanged. This ensures callers that rely
// on "at least try the literal token" behaviour never receive an empty path list.
func TestEquivalentTimeframeTokens_UnknownTokenPassthrough(t *testing.T) {
	for _, token := range []string{"", "unknown", "wtf", "9999x"} {
		token := token
		t.Run(token, func(t *testing.T) {
			got := equivalentTimeframeTokens(token)
			if len(got) != 1 {
				t.Fatalf("got %v (len %d), want single-element passthrough", got, len(got))
			}
			if got[0] != token {
				t.Errorf("got %q, want passthrough of original token %q", got[0], token)
			}
		})
	}
}

// TestEquivalentTimeframeTokens_SubMinutePeriodHasNoNumericAlternate verifies
// that seconds-scale periods are not assigned a minute-count alternate because
// they are not expressible as a whole number of minutes.
func TestEquivalentTimeframeTokens_SubMinutePeriodHasNoNumericAlternate(t *testing.T) {
	for _, token := range []string{"1s", "5s", "10s", "30s"} {
		token := token
		t.Run(token, func(t *testing.T) {
			got := equivalentTimeframeTokens(token)
			if len(got) != 1 {
				t.Errorf("got %v, want exactly one token (no minute-count alternate for sub-minute period)", got)
			}
		})
	}
}

// TestEquivalentTimeframeTokens_CalendarScaleHasNoNumericAlternate verifies
// that daily-and-above periods carry no numeric minute-count alternate because
// Pine Script does not accept minute counts for calendar-scale timeframes.
func TestEquivalentTimeframeTokens_CalendarScaleHasNoNumericAlternate(t *testing.T) {
	for _, token := range []string{"1D", "2D", "3D", "1W", "1M", "12M"} {
		token := token
		t.Run(token, func(t *testing.T) {
			got := equivalentTimeframeTokens(token)
			if len(got) != 1 {
				t.Errorf("got %v, want exactly one token (no numeric alternate for calendar-scale)", got)
			}
		})
	}
}
