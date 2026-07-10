package datafetcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileFetcher_FetchSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BTC_1h.json")

	testData := `[
		{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
		{"time": 1700003600, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100}
	]`

	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)

	bars, err := fetcher.Fetch("BTC", "1h", 0)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(bars))
	}

	if bars[0].Close != 102 {
		t.Errorf("Expected first close 102, got %.2f", bars[0].Close)
	}
}

func TestFileFetcher_FetchWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "ETH_1D.json")

	testData := `[
		{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
		{"time": 1700086400, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100},
		{"time": 1700172800, "open": 104, "high": 109, "low": 99, "close": 106, "volume": 1200}
	]`

	os.WriteFile(testFile, []byte(testData), 0644)

	fetcher := NewFileFetcher(tmpDir, 0)

	bars, err := fetcher.Fetch("ETH", "1D", 2)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(bars) != 2 {
		t.Errorf("Expected 2 bars, got %d", len(bars))
	}

	if bars[0].Close != 104 {
		t.Errorf("Expected first close 104, got %.2f", bars[0].Close)
	}
}

// TestFileFetcher_Fetch_LimitLargerThanBarsReturnsAll verifies that when limit
// exceeds the number of bars in the file, all bars are returned without error.
func TestFileFetcher_Fetch_LimitLargerThanBarsReturnsAll(t *testing.T) {
	tmpDir := t.TempDir()
	barsJSON := `[
		{"time":1700000000,"open":1,"high":2,"low":0.5,"close":1.5,"volume":10},
		{"time":1700003600,"open":1.5,"high":2.5,"low":1,"close":2,"volume":20}
	]`
	if err := os.WriteFile(filepath.Join(tmpDir, "X-1h.json"), []byte(barsJSON), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	bars, err := NewFileFetcher(tmpDir, 0).Fetch("X", "1h", 9999)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(bars) != 2 {
		t.Errorf("limit > len: got %d bars, want 2 (all bars)", len(bars))
	}
}

// TestFileFetcher_Fetch_LimitEqualsBarsReturnsAll verifies that when limit
// equals the number of bars exactly, all bars are returned (boundary condition).
func TestFileFetcher_Fetch_LimitEqualsBarsReturnsAll(t *testing.T) {
	tmpDir := t.TempDir()
	barsJSON := `[
		{"time":1700000000,"open":1,"high":2,"low":0.5,"close":1.5,"volume":10},
		{"time":1700003600,"open":1.5,"high":2.5,"low":1,"close":2,"volume":20},
		{"time":1700007200,"open":2,"high":3,"low":1.5,"close":2.5,"volume":30}
	]`
	if err := os.WriteFile(filepath.Join(tmpDir, "X-1h.json"), []byte(barsJSON), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	bars, err := NewFileFetcher(tmpDir, 0).Fetch("X", "1h", 3)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(bars) != 3 {
		t.Errorf("limit == len: got %d bars, want 3 (all bars)", len(bars))
	}
}

func TestFileFetcher_FetchWithMetadata_PreservesSourceMetadataAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "SBERP_1h.json")

	testData := `{
		"timezone": "Europe/Moscow",
		"exchange": "MOEX",
		"referenceSession": "regular",
		"openDates": ["2025-08-16"],
		"bars": [
			{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000},
			{"time": 1700003600, "open": 102, "high": 107, "low": 97, "close": 104, "volume": 1100},
			{"time": 1700007200, "open": 104, "high": 109, "low": 99, "close": 106, "volume": 1200}
		]
	}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("SBERP", "1h", 2)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q, want Europe/Moscow", marketData.Timezone)
	}
	if marketData.ReferenceSession != "regular" {
		t.Fatalf("reference session = %q, want regular", marketData.ReferenceSession)
	}
	if marketData.Exchange != "MOEX" {
		t.Fatalf("exchange = %q, want MOEX", marketData.Exchange)
	}
	if len(marketData.OpenDates) != 1 || marketData.OpenDates[0] != "2025-08-16" {
		t.Fatalf("open dates = %#v, want [2025-08-16]", marketData.OpenDates)
	}
	if len(marketData.Bars) != 2 {
		t.Fatalf("bars = %d, want 2", len(marketData.Bars))
	}
	if marketData.Bars[0].Close != 104 {
		t.Fatalf("first limited close = %.2f, want 104.00", marketData.Bars[0].Close)
	}
}

func TestFileFetcher_FetchWithMetadata_ArrayInputHasEmptyMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BTCUSDT_1h.json")

	testData := `[{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}]`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("BTCUSDT", "1h", 0)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "" {
		t.Fatalf("timezone = %q, want empty", marketData.Timezone)
	}
	if marketData.ReferenceSession != "" {
		t.Fatalf("reference session = %q, want empty", marketData.ReferenceSession)
	}
	if len(marketData.Bars) != 1 {
		t.Fatalf("bars = %d, want 1", len(marketData.Bars))
	}
}

func TestFileFetcher_FetchWithMetadata_ObjectInputAllowsEmptyBars(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "SBERP_1h.json")

	testData := `{"timezone":"Europe/Moscow","referenceSession":"regular","bars":[]}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	marketData, err := fetcher.FetchWithMetadata("SBERP", "1h", 0)
	if err != nil {
		t.Fatalf("FetchWithMetadata failed: %v", err)
	}

	if marketData.Timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q, want Europe/Moscow", marketData.Timezone)
	}
	if marketData.ReferenceSession != "regular" {
		t.Fatalf("reference session = %q, want regular", marketData.ReferenceSession)
	}
	if len(marketData.Bars) != 0 {
		t.Fatalf("bars = %d, want 0", len(marketData.Bars))
	}
}

func TestFileFetcher_FetchWithMetadata_InvalidObjectBarsFails(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BAD_1h.json")

	testData := `{"timezone":"UTC","bars":{}}`
	if err := os.WriteFile(testFile, []byte(testData), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	fetcher := NewFileFetcher(tmpDir, 0)
	if _, err := fetcher.FetchWithMetadata("BAD", "1h", 0); err == nil {
		t.Fatal("expected invalid object bars error")
	}
}

func TestFileFetcher_SimulatedLatency(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "TEST_1m.json")

	testData := `[{"time": 1700000000, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}]`
	os.WriteFile(testFile, []byte(testData), 0644)

	fetcher := NewFileFetcher(tmpDir, 50*time.Millisecond)

	start := time.Now()
	_, err := fetcher.Fetch("TEST", "1m", 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected latency >=50ms, got %v", elapsed)
	}
}

func TestFileFetcher_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	fetcher := NewFileFetcher(tmpDir, 0)

	_, err := fetcher.Fetch("NONEXISTENT", "1h", 0)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestFileFetcher_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "BAD_1h.json")

	os.WriteFile(testFile, []byte("not valid json"), 0644)

	fetcher := NewFileFetcher(tmpDir, 0)

	_, err := fetcher.Fetch("BAD", "1h", 0)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestFixtureCandidatePaths_OrderAndEncoding verifies that fixtureCandidatePaths
// emits paths in priority order: hyphen-canonical, underscore-canonical, then
// hyphen-numeric and underscore-numeric (intraday only). Calendar-scale tokens
// have no numeric alternate so only two paths are produced.
func TestFixtureCandidatePaths_OrderAndEncoding(t *testing.T) {
	cases := []struct {
		symbol    string
		timeframe string
		want      []string
	}{
		// Intraday: 4 paths (canonical + numeric alternate, each in hyphen then underscore form).
		{"SBERP", "1h", []string{"SBERP-1h.json", "SBERP_1h.json", "SBERP-60.json", "SBERP_60.json"}},
		{"SBERP", "60", []string{"SBERP-1h.json", "SBERP_1h.json", "SBERP-60.json", "SBERP_60.json"}},
		{"AAPL", "4h", []string{"AAPL-4h.json", "AAPL_4h.json", "AAPL-240.json", "AAPL_240.json"}},
		{"AAPL", "240", []string{"AAPL-4h.json", "AAPL_4h.json", "AAPL-240.json", "AAPL_240.json"}},
		{"NVDA", "30m", []string{"NVDA-30m.json", "NVDA_30m.json", "NVDA-30.json", "NVDA_30.json"}},
		{"NVDA", "30", []string{"NVDA-30m.json", "NVDA_30m.json", "NVDA-30.json", "NVDA_30.json"}},
		// Calendar-scale: 2 paths (no numeric alternate).
		{"AAPL", "1D", []string{"AAPL-1D.json", "AAPL_1D.json"}},
		{"BTCUSDT", "1M", []string{"BTCUSDT-1M.json", "BTCUSDT_1M.json"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.symbol+":"+tc.timeframe, func(t *testing.T) {
			paths := fixtureCandidatePaths("/data", tc.symbol, tc.timeframe)
			if len(paths) != len(tc.want) {
				t.Fatalf("got %d paths %v, want %d %v", len(paths), paths, len(tc.want), tc.want)
			}
			for i, w := range tc.want {
				if !strings.HasSuffix(paths[i], w) {
					t.Errorf("[%d]: got %q, want suffix %q", i, paths[i], w)
				}
			}
		})
	}
}

// TestFixtureCandidatePaths_SymmetricEquivalence verifies the invariant that two
// tokens denoting the same period produce identical candidate path lists. This
// ensures fixture resolution is stable regardless of whether the caller uses the
// suffixed form ("1h") or the numeric form ("60").
func TestFixtureCandidatePaths_SymmetricEquivalence(t *testing.T) {
	pairs := [][2]string{
		{"1m", "1"}, {"5m", "5"}, {"15m", "15"}, {"30m", "30"},
		{"1h", "60"}, {"2h", "120"}, {"3h", "180"}, {"4h", "240"},
	}
	const dir = "/fixtures"
	const sym = "SYM"
	for _, pair := range pairs {
		a := fixtureCandidatePaths(dir, sym, pair[0])
		b := fixtureCandidatePaths(dir, sym, pair[1])
		if len(a) != len(b) {
			t.Errorf("(%q,%q) path count mismatch: %v vs %v", pair[0], pair[1], a, b)
			continue
		}
		for i := range a {
			if a[i] != b[i] {
				t.Errorf("(%q,%q)[%d]: %q vs %q", pair[0], pair[1], i, a[i], b[i])
			}
		}
	}
}

// TestFixtureCandidatePaths_UnknownTokenPassthrough verifies that an unrecognised
// timeframe token is passed through unchanged and results in exactly two candidate
// paths: hyphen-separated and underscore-separated. No panic or empty list.
func TestFixtureCandidatePaths_UnknownTokenPassthrough(t *testing.T) {
	for _, token := range []string{"", "unknown", "9999x"} {
		token := token
		t.Run(token, func(t *testing.T) {
			paths := fixtureCandidatePaths("/data", "SYM", token)
			if len(paths) != 2 {
				t.Fatalf("got %d paths %v, want 2 (hyphen + underscore)", len(paths), paths)
			}
			if !strings.HasSuffix(paths[0], "SYM-"+token+".json") {
				t.Errorf("paths[0] = %q, want hyphen-separated suffix", paths[0])
			}
			if !strings.HasSuffix(paths[1], "SYM_"+token+".json") {
				t.Errorf("paths[1] = %q, want underscore-separated suffix", paths[1])
			}
		})
	}
}

// TestFileFetcher_HyphenWinsOverUnderscore verifies that when both
// SYMBOL-TF.json (hyphen) and SYMBOL_TF.json (underscore) exist, the
// hyphen-named file is returned (first candidate wins).
func TestFileFetcher_HyphenWinsOverUnderscore(t *testing.T) {
	tmpDir := t.TempDir()
	hyphenData := `[{"time":1,"open":10,"high":11,"low":9,"close":10,"volume":1}]`
	underscoreData := `[{"time":2,"open":20,"high":21,"low":19,"close":20,"volume":2}]`

	if err := os.WriteFile(filepath.Join(tmpDir, "SYM-1h.json"), []byte(hyphenData), 0644); err != nil {
		t.Fatalf("write hyphen: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "SYM_1h.json"), []byte(underscoreData), 0644); err != nil {
		t.Fatalf("write underscore: %v", err)
	}

	bars, err := NewFileFetcher(tmpDir, 0).Fetch("SYM", "1h", 0)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(bars) == 0 || bars[0].Time != 1 {
		t.Errorf("expected hyphen fixture (time=1), got time=%d", bars[0].Time)
	}
}

// TestFileFetcher_UnderscoreFallbackWhenHyphenAbsent verifies that when only
// SYMBOL_TF.json (underscore, legacy convention) exists and no hyphen file is
// present, the fetcher finds it as the second candidate.
func TestFileFetcher_UnderscoreFallbackWhenHyphenAbsent(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "SYM_1h.json"), []byte(
		`[{"time":42,"open":1,"high":2,"low":0.5,"close":1.5,"volume":10}]`,
	), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	bars, err := NewFileFetcher(tmpDir, 0).Fetch("SYM", "1h", 0)
	if err != nil {
		t.Fatalf("Fetch with underscore-only fixture: %v", err)
	}
	if len(bars) != 1 || bars[0].Time != 42 {
		t.Errorf("expected 1 bar with time=42, got %v", bars)
	}
}

// TestFileFetcher_CrossEncodingResolution verifies that a fixture named with one
// encoding of a timeframe period is resolved when the fetcher is called with the
// equivalent encoding, and vice versa. Covers both hyphen-canonical and
// underscore-numeric naming so that all four candidate path positions are exercised.
func TestFileFetcher_CrossEncodingResolution(t *testing.T) {
	onebar := `[{"time":1700000000,"open":1,"high":2,"low":0.5,"close":1.5,"volume":100}]`

	cases := []struct {
		fixtureFile  string // filename created in tmpDir
		requestToken string // timeframe arg passed to Fetch
	}{
		// Hyphen-canonical fixture resolved via numeric request (candidate 3).
		{"SYM-1h.json", "60"},
		{"SYM-4h.json", "240"},
		{"SYM-5m.json", "5"},
		{"SYM-30m.json", "30"},
		// Hyphen-numeric fixture resolved via canonical request (candidate 1 misses, candidate 3 matches).
		{"SYM-60.json", "1h"},
		{"SYM-240.json", "4h"},
		// Hyphen-canonical fixture resolved via same token (candidate 1 matches).
		{"SYM-1h.json", "1h"},
		{"SYM-1D.json", "1D"},
		// Underscore-canonical fixture resolved via same token (candidate 2 matches).
		{"SYM_1h.json", "1h"},
		{"SYM_1D.json", "1D"},
		// Underscore-numeric fixture resolved via canonical request (candidate 4 matches).
		{"SYM_60.json", "1h"},
		{"SYM_240.json", "4h"},
	}

	for _, tc := range cases {
		tc := tc
		name := strings.TrimSuffix(tc.fixtureFile, ".json") + "_via_" + tc.requestToken
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, tc.fixtureFile), []byte(onebar), 0644); err != nil {
				t.Fatal(err)
			}
			bars, err := NewFileFetcher(tmpDir, 0).Fetch("SYM", tc.requestToken, 0)
			if err != nil {
				t.Fatalf("Fetch(SYM,%q) with fixture %q: %v", tc.requestToken, tc.fixtureFile, err)
			}
			if len(bars) != 1 {
				t.Fatalf("want 1 bar, got %d", len(bars))
			}
		})
	}
}
