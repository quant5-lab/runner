package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type coverageGap struct {
	RelStrategyPath string
	Symbol          string
	Timeframe       string
	TriedFilenames  []string
}

// Enables a single fixture file to satisfy multiple Pine TF spellings.
func timeframeVariants(tf string) []string {
	switch strings.ToUpper(tf) {
	case "D", "1D":
		return []string{"1D", "D"}
	case "W", "1W":
		return []string{"1W", "W"}
	case "M", "1M":
		return []string{"1M", "M"}
	case "1H", "60":
		return []string{"1h", "60", "1H"}
	case "4H", "240":
		return []string{"4h", "240", "4H"}
	default:
		return []string{tf}
	}
}

// Exchange-prefixed symbols ("BINANCE:BTCUSDT") are normalized to the bare ticker part.
func candidateFilenames(symbol, tf string) []string {
	bare := symbol
	if idx := strings.LastIndex(symbol, ":"); idx >= 0 {
		bare = symbol[idx+1:]
	}

	seen := map[string]bool{}
	var names []string
	for _, variant := range timeframeVariants(tf) {
		for _, sep := range []string{"_", "-"} {
			name := bare + sep + variant + ".json"
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
		}
	}
	return names
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func anyExists(paths []string) bool {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func collectCoverageGaps(literalRefs []securityRef, fixtureDirs []string) []coverageGap {
	var gaps []coverageGap
	for _, ref := range literalRefs {
		names := candidateFilenames(ref.Symbol, ref.Timeframe)
		var candidates []string
		for _, dir := range fixtureDirs {
			for _, name := range names {
				candidates = append(candidates, filepath.Join(dir, name))
			}
		}
		if !anyExists(candidates) {
			gaps = append(gaps, coverageGap{
				RelStrategyPath: ref.RelStrategyPath,
				Symbol:          ref.Symbol,
				Timeframe:       ref.Timeframe,
				TriedFilenames:  uniqueStrings(names),
			})
		}
	}
	return gaps
}

func TestTimeframeVariants(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"D", []string{"1D", "D"}},
		{"1D", []string{"1D", "D"}},
		{"W", []string{"1W", "W"}},
		{"1W", []string{"1W", "W"}},
		{"M", []string{"1M", "M"}},
		{"1M", []string{"1M", "M"}},
		{"1h", []string{"1h", "60", "1H"}},
		{"60", []string{"1h", "60", "1H"}},
		{"1H", []string{"1h", "60", "1H"}},
		{"4h", []string{"4h", "240", "4H"}},
		{"240", []string{"4h", "240", "4H"}},
		{"15", []string{"15"}},
		{"5m", []string{"5m"}},
		{"custom", []string{"custom"}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := timeframeVariants(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("[%d] got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

func TestCandidateFilenames(t *testing.T) {
	cases := []struct {
		name        string
		symbol      string
		tf          string
		mustContain []string
	}{
		{
			name:        "plain symbol underscore and hyphen",
			symbol:      "SBERP",
			tf:          "1D",
			mustContain: []string{"SBERP_1D.json", "SBERP-1D.json", "SBERP_D.json", "SBERP-D.json"},
		},
		{
			name:        "exchange-prefixed symbol strips prefix",
			symbol:      "BINANCE:BTCUSDT",
			tf:          "D",
			mustContain: []string{"BTCUSDT_1D.json", "BTCUSDT-1D.json"},
		},
		{
			name:        "hourly TF produces multiple aliases",
			symbol:      "AAPL",
			tf:          "1h",
			mustContain: []string{"AAPL_1h.json", "AAPL-1h.json", "AAPL_60.json", "AAPL-60.json"},
		},
		{
			name:        "multi-colon exchange prefix uses rightmost segment",
			symbol:      "NSE:BSE:INFY",
			tf:          "1D",
			mustContain: []string{"INFY_1D.json", "INFY-1D.json"},
		},
		{
			name:        "no duplicates in output",
			symbol:      "X",
			tf:          "1D",
			mustContain: []string{"X_1D.json", "X-1D.json", "X_D.json", "X-D.json"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := candidateFilenames(tc.symbol, tc.tf)
			gotSet := make(map[string]bool, len(got))
			for _, name := range got {
				if gotSet[name] {
					t.Errorf("duplicate filename in output: %q", name)
				}
				gotSet[name] = true
			}
			for _, want := range tc.mustContain {
				if !gotSet[want] {
					t.Errorf("expected %q in result %v", want, got)
				}
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	cases := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no duplicates preserved as-is",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "adjacent duplicates deduplicated",
			input: []string{"x", "x", "y"},
			want:  []string{"x", "y"},
		},
		{
			name:  "non-adjacent duplicates deduplicated",
			input: []string{"a", "b", "a"},
			want:  []string{"a", "b"},
		},
		{
			name:  "all same",
			input: []string{"z", "z", "z"},
			want:  []string{"z"},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  []string{},
		},
		{
			name:  "order preserved",
			input: []string{"c", "a", "b"},
			want:  []string{"c", "a", "b"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := uniqueStrings(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i, w := range tc.want {
				if got[i] != w {
					t.Errorf("[%d] got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

// TestSecurityFixtureCoverage scans all non-skip Pine strategy files for
// security() / request.security() calls and reports which literal-symbol
// (symbol, timeframe) pairs lack fixture files in the known fixture directories.
func TestAnyExists(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "present.json")
	if err := os.WriteFile(present, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	absent := filepath.Join(dir, "absent.json")

	cases := []struct {
		name  string
		paths []string
		want  bool
	}{
		{"single existing path", []string{present}, true},
		{"single absent path", []string{absent}, false},
		{"first absent, second present", []string{absent, present}, true},
		{"first present, second absent", []string{present, absent}, true},
		{"all absent", []string{absent, filepath.Join(dir, "also-absent.json")}, false},
		{"empty paths list", []string{}, false},
		{"nil paths", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := anyExists(tc.paths); got != tc.want {
				t.Errorf("anyExists(%v) = %v, want %v", tc.paths, got, tc.want)
			}
		})
	}
}

func TestCollectCoverageGaps(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"BTC_1D.json", "ETH-1h.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	refs := []securityRef{
		{RelStrategyPath: "has_fixture.pine", Symbol: "BTC", Timeframe: "1D"},
		{RelStrategyPath: "also_has.pine", Symbol: "ETH", Timeframe: "1h"},
		{RelStrategyPath: "no_fixture.pine", Symbol: "MISSING", Timeframe: "1D"},
		{RelStrategyPath: "exchange_prefix.pine", Symbol: "BINANCE:MISSING2", Timeframe: "D"},
	}

	gaps := collectCoverageGaps(refs, []string{dir})

	if len(gaps) != 2 {
		t.Fatalf("got %d gaps, want 2: %v", len(gaps), gaps)
	}
	gapSymbols := make(map[string]bool)
	for _, g := range gaps {
		gapSymbols[g.Symbol] = true
	}
	if !gapSymbols["MISSING"] {
		t.Error("expected gap for MISSING symbol")
	}
	if !gapSymbols["BINANCE:MISSING2"] {
		t.Error("expected gap for BINANCE:MISSING2 symbol")
	}
}

func TestCollectCoverageGaps_EmptyRefs(t *testing.T) {
	gaps := collectCoverageGaps(nil, []string{t.TempDir()})
	if len(gaps) != 0 {
		t.Errorf("got %d gaps for nil refs, want 0", len(gaps))
	}
}

func TestCollectCoverageGaps_MultipleFixtureDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir2, "SYM_1D.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	refs := []securityRef{{Symbol: "SYM", Timeframe: "1D"}}
	gaps := collectCoverageGaps(refs, []string{dir1, dir2})
	if len(gaps) != 0 {
		t.Errorf("got %d gaps, want 0 (fixture exists in dir2)", len(gaps))
	}
}

func TestCollectCoverageGaps_TriedFilenamesNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	refs := []securityRef{{Symbol: "X", Timeframe: "1D"}}
	gaps := collectCoverageGaps(refs, []string{dir})
	if len(gaps) != 1 {
		t.Fatalf("expected 1 gap")
	}
	seen := make(map[string]bool)
	for _, name := range gaps[0].TriedFilenames {
		if seen[name] {
			t.Errorf("duplicate filename in TriedFilenames: %q", name)
		}
		seen[name] = true
	}
}

// syminfo.tickerid calls are informational-only; fixture availability depends on the runtime symbol.
func TestSecurityFixtureCoverage(t *testing.T) {
	root := projectRootFromCwd()

	refs := scanStrategiesForSecurityCalls(t, filepath.Join(root, "strategies"))
	if len(refs) == 0 {
		t.Log("no security() calls found in strategy files")
		return
	}

	fixtureDirs := []string{
		filepath.Join(root, "tests", "fixtures", "ohlcv"),
		filepath.Join(root, "tests", "golden", "fixtures", "data"),
	}

	literal, runtime := partitionBySymbolType(refs)
	gaps := collectCoverageGaps(literal, fixtureDirs)

	t.Logf("security() call scan: %d literal-symbol refs, %d runtime-symbol refs (syminfo.tickerid)",
		len(literal), len(runtime))
	t.Logf("literal-symbol gaps: %d missing fixtures", len(gaps))

	for _, g := range gaps {
		t.Logf("MISSING  strategy=%-45s  symbol=%-25s  tf=%-5s  tried=%s",
			g.RelStrategyPath, g.Symbol, g.Timeframe, strings.Join(g.TriedFilenames, ", "))
	}

	if len(runtime) > 0 {
		t.Log("runtime-symbol refs (verify fixture exists for each test symbol):")
		for _, r := range runtime {
			t.Logf("  INFO  strategy=%-45s  tf=%s", r.RelStrategyPath, r.Timeframe)
		}
	}

	if len(gaps) > 0 {
		t.Log("Resolution: add fixtures via 'make fetch-strategy' or add skip-with-reason entries to smoke test skipList")
	}
}

// TestCollectCoverageGaps_SameSymbolMultipleStrategies verifies that when two
// strategies both reference the same missing (symbol, timeframe) fixture, both
// are reported as separate gaps.  Gaps are per-strategy, not per-fixture.
func TestCollectCoverageGaps_SameSymbolMultipleStrategies(t *testing.T) {
	dir := t.TempDir()
	refs := []securityRef{
		{RelStrategyPath: "strat_a.pine", Symbol: "MISSING", Timeframe: "1D"},
		{RelStrategyPath: "strat_b.pine", Symbol: "MISSING", Timeframe: "1D"},
	}

	gaps := collectCoverageGaps(refs, []string{dir})

	if len(gaps) != 2 {
		t.Fatalf("got %d gaps, want 2 (one per strategy even for the same missing fixture)", len(gaps))
	}
	strategies := map[string]bool{gaps[0].RelStrategyPath: true, gaps[1].RelStrategyPath: true}
	if !strategies["strat_a.pine"] || !strategies["strat_b.pine"] {
		t.Errorf("expected gaps for strat_a.pine and strat_b.pine, got %v", gaps)
	}
}

// TestCollectCoverageGaps_FixtureAddedResolvesBothStrategies verifies that once
// a fixture is present, both strategies that referenced the missing fixture are
// no longer reported as gaps.
func TestCollectCoverageGaps_FixtureAddedResolvesBothStrategies(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SYM_1D.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	refs := []securityRef{
		{RelStrategyPath: "strat_a.pine", Symbol: "SYM", Timeframe: "1D"},
		{RelStrategyPath: "strat_b.pine", Symbol: "SYM", Timeframe: "1D"},
	}

	gaps := collectCoverageGaps(refs, []string{dir})

	if len(gaps) != 0 {
		t.Errorf("got %d gaps, want 0 after fixture added: %v", len(gaps), gaps)
	}
}

// TestCandidateFilenames_ExchangePrefixVariants verifies stripping for exchange
// prefixes with different separators and casing.
func TestCandidateFilenames_ExchangePrefixVariants(t *testing.T) {
	cases := []struct {
		symbol      string
		tf          string
		mustContain string
		mustAbsent  string
	}{
		// Exchange prefix stripped; only the ticker part appears in filenames.
		{"BINANCE:BTCUSDT", "1D", "BTCUSDT_1D.json", "BINANCE"},
		{"NYSE:AAPL", "1h", "AAPL_1h.json", "NYSE"},
		// Symbol with no prefix is used as-is.
		{"SBERP", "1D", "SBERP_1D.json", ""},
	}
	for _, tc := range cases {
		t.Run(tc.symbol, func(t *testing.T) {
			names := candidateFilenames(tc.symbol, tc.tf)
			found := false
			for _, n := range names {
				if n == tc.mustContain {
					found = true
				}
				if tc.mustAbsent != "" && strings.Contains(n, tc.mustAbsent) {
					t.Errorf("filename %q must not contain %q", n, tc.mustAbsent)
				}
			}
			if !found {
				t.Errorf("expected %q in candidateFilenames(%q,%q)=%v", tc.mustContain, tc.symbol, tc.tf, names)
			}
		})
	}
}

// TestTimeframeVariants_AllKnownAliases verifies that every supported canonical
// timeframe spelling and its numeric alias resolve to the same candidate set.
func TestTimeframeVariants_AllKnownAliases(t *testing.T) {
	equivalentPairs := []struct {
		a, b string
	}{
		{"1h", "1H"},
		{"1h", "60"},
		{"4h", "4H"},
		{"4h", "240"},
		{"1D", "D"},
		{"1W", "W"},
		{"1M", "M"},
	}
	for _, p := range equivalentPairs {
		t.Run(p.a+"=="+p.b, func(t *testing.T) {
			va := timeframeVariants(p.a)
			vb := timeframeVariants(p.b)
			if len(va) != len(vb) {
				t.Fatalf("variant counts differ: %v vs %v", va, vb)
			}
			setA := make(map[string]bool, len(va))
			for _, v := range va {
				setA[v] = true
			}
			for _, v := range vb {
				if !setA[v] {
					t.Errorf("variant %q from %q not found in variants of %q: %v", v, p.b, p.a, va)
				}
			}
		})
	}
}
