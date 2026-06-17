package datafetcher

import (
	"path/filepath"
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

func TestCandidateFixturePaths(t *testing.T) {
	dir := "/data"
	cases := []struct {
		symbol    string
		timeframe string
		want      []string
	}{
		// Intraday suffixed — canonical then numeric alternate.
		{"SBERP", "1h", []string{filepath.Join(dir, "SBERP_1h.json"), filepath.Join(dir, "SBERP_60.json")}},
		{"SBERP", "4h", []string{filepath.Join(dir, "SBERP_4h.json"), filepath.Join(dir, "SBERP_240.json")}},
		{"BTCUSDT", "1h", []string{filepath.Join(dir, "BTCUSDT_1h.json"), filepath.Join(dir, "BTCUSDT_60.json")}},
		// Intraday numeric — identical path list as the suffixed counterpart.
		{"SBERP", "60", []string{filepath.Join(dir, "SBERP_1h.json"), filepath.Join(dir, "SBERP_60.json")}},
		{"SBERP", "240", []string{filepath.Join(dir, "SBERP_4h.json"), filepath.Join(dir, "SBERP_240.json")}},
		// Calendar-scale — single candidate, no numeric fallback.
		{"AAPL", "1D", []string{filepath.Join(dir, "AAPL_1D.json")}},
		{"AAPL", "2D", []string{filepath.Join(dir, "AAPL_2D.json")}},
		{"AAPL", "1W", []string{filepath.Join(dir, "AAPL_1W.json")}},
		{"AAPL", "1M", []string{filepath.Join(dir, "AAPL_1M.json")}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.symbol+":"+tc.timeframe, func(t *testing.T) {
			got := candidateFixturePaths(dir, tc.symbol, tc.timeframe)
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

// TestCandidateFixturePaths_EquivalentInputsProduceSamePaths verifies that two
// tokens naming the same period generate the same ordered candidate list —
// resolution is a property of the period, not the token encoding.
func TestCandidateFixturePaths_EquivalentInputsProduceSamePaths(t *testing.T) {
	pairs := [][2]string{
		{"1h", "60"},
		{"4h", "240"},
		{"2h", "120"},
		{"15m", "15"},
		{"30m", "30"},
	}
	for _, pair := range pairs {
		a := candidateFixturePaths("/dir", "SYM", pair[0])
		b := candidateFixturePaths("/dir", "SYM", pair[1])
		if len(a) != len(b) {
			t.Errorf("(%q,%q) path count mismatch: %d vs %d", pair[0], pair[1], len(a), len(b))
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("(%q,%q)[%d]: %q vs %q", pair[0], pair[1], i, a[i], b[i])
			}
		}
	}
}
