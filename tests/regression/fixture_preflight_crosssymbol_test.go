package regression

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// crossSymbolFamily holds alignment cases for one strategy wired against two or
// more distinct symbols. Placeholder or copied fixtures in such a family silently
// corrupt TV-match attempts because price scales differ by orders of magnitude
// across symbols.
type crossSymbolFamily struct {
	Strategy string
	Cases    []tvAlignmentCase
}

func groupCrossSymbolFamilies(cases []tvAlignmentCase) []crossSymbolFamily {
	byStrategy := make(map[string][]tvAlignmentCase)
	for _, c := range cases {
		byStrategy[c.Strategy] = append(byStrategy[c.Strategy], c)
	}
	var families []crossSymbolFamily
	for _, stratCases := range byStrategy {
		symbols := make(map[string]struct{})
		for _, c := range stratCases {
			symbols[c.Symbol] = struct{}{}
		}
		if len(symbols) < 2 {
			continue
		}
		sorted := append([]tvAlignmentCase(nil), stratCases...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Symbol < sorted[j].Symbol })
		families = append(families, crossSymbolFamily{Strategy: sorted[0].Strategy, Cases: sorted})
	}
	sort.Slice(families, func(i, j int) bool { return families[i].Strategy < families[j].Strategy })
	return families
}

// familyOHLCVKeysBySymbol returns the ohlcvContentKey for each fixture keyed by
// symbol. Identical keys mean bar-level OHLCV+time content is the same across
// symbols — the definitive signature of synthetic placeholder fixtures regardless
// of file-level metadata differences (symbol name, timezone string, etc.).
func familyOHLCVKeysBySymbol(family crossSymbolFamily, fixtureDir string) (map[string]string, error) {
	out := make(map[string]string, len(family.Cases))
	for _, c := range family.Cases {
		f, err := loadPreflightFixture(filepath.Join(fixtureDir, c.Data))
		if err != nil {
			return nil, err
		}
		out[c.Symbol] = ohlcvContentKey(f.Bars)
	}
	return out, nil
}

// familyDuplicateSymbols scans the family's cases in their declared order
// (symbol-sorted, guaranteed by groupCrossSymbolFamilies) and returns the first
// pair of symbols whose OHLCV bar content is identical. Identical content is the
// definitive signature of synthetic placeholder fixtures regardless of file-level
// metadata differences (symbol field, timezone string). Returns [2]string{} and
// false when all symbols have distinct content.
func familyDuplicateSymbols(family crossSymbolFamily, fixtureDir string) ([2]string, bool, error) {
	seen := make(map[string]string, len(family.Cases))
	for _, c := range family.Cases {
		f, err := loadPreflightFixture(filepath.Join(fixtureDir, c.Data))
		if err != nil {
			return [2]string{}, false, fmt.Errorf("load fixture %s: %w", c.Data, err)
		}
		key := ohlcvContentKey(f.Bars)
		if prior, dup := seen[key]; dup {
			return [2]string{prior, c.Symbol}, true, nil
		}
		seen[key] = c.Symbol
	}
	return [2]string{}, false, nil
}

func makeCase(strategy, symbol, data string) tvAlignmentCase {
	return tvAlignmentCase{Strategy: strategy, Symbol: symbol, Data: data, Name: strategy + "/" + symbol}
}

func writePreflightFixtureToDisk(t *testing.T, dir, name string, bars []preflightBar) {
	t.Helper()
	data, err := json.Marshal(map[string]interface{}{"timezone": "UTC", "bars": bars})
	if err != nil {
		t.Fatalf("marshal fixture %s: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
}

// TestGroupCrossSymbolFamilies covers family detection, single-symbol exclusion,
// deterministic ordering, and mixed registries.
func TestGroupCrossSymbolFamilies(t *testing.T) {
	cases := []struct {
		name            string
		input           []tvAlignmentCase
		wantStrategies  []string
		wantFamilySizes map[string]int
	}{
		{
			name:           "nil_input",
			input:          nil,
			wantStrategies: nil,
		},
		{
			name:           "empty_input",
			input:          []tvAlignmentCase{},
			wantStrategies: nil,
		},
		{
			name:           "all_single_symbol_no_families",
			input:          []tvAlignmentCase{makeCase("a.pine", "AAPL", "aapl.json"), makeCase("b.pine", "SBERP", "sberp.json")},
			wantStrategies: nil,
		},
		{
			name:           "same_strategy_same_symbol_twice_not_a_family",
			input:          []tvAlignmentCase{makeCase("a.pine", "AAPL", "aapl1.json"), makeCase("a.pine", "AAPL", "aapl2.json")},
			wantStrategies: nil,
		},
		{
			name:            "two_symbols_same_strategy_forms_family",
			input:           []tvAlignmentCase{makeCase("a.pine", "AAPL", "aapl.json"), makeCase("a.pine", "BTC", "btc.json")},
			wantStrategies:  []string{"a.pine"},
			wantFamilySizes: map[string]int{"a.pine": 2},
		},
		{
			name: "three_symbols_same_strategy_forms_one_family",
			input: []tvAlignmentCase{
				makeCase("a.pine", "SBERP", "sberp.json"),
				makeCase("a.pine", "AAPL", "aapl.json"),
				makeCase("a.pine", "BTC", "btc.json"),
			},
			wantStrategies:  []string{"a.pine"},
			wantFamilySizes: map[string]int{"a.pine": 3},
		},
		{
			name: "two_strategies_each_multi_symbol",
			input: []tvAlignmentCase{
				makeCase("b.pine", "BTC", "btc.json"),
				makeCase("a.pine", "AAPL", "aapl.json"),
				makeCase("a.pine", "SBERP", "sberp.json"),
				makeCase("b.pine", "AAPL", "aapl.json"),
			},
			wantStrategies:  []string{"a.pine", "b.pine"},
			wantFamilySizes: map[string]int{"a.pine": 2, "b.pine": 2},
		},
		{
			name: "mixed_single_and_multi",
			input: []tvAlignmentCase{
				makeCase("single.pine", "AAPL", "aapl.json"),
				makeCase("multi.pine", "AAPL", "aapl.json"),
				makeCase("multi.pine", "BTC", "btc.json"),
			},
			wantStrategies:  []string{"multi.pine"},
			wantFamilySizes: map[string]int{"multi.pine": 2},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := groupCrossSymbolFamilies(tc.input)
			if len(got) != len(tc.wantStrategies) {
				t.Fatalf("family count = %d, want %d (strategies: %v)", len(got), len(tc.wantStrategies), tc.wantStrategies)
			}
			for i, f := range got {
				if f.Strategy != tc.wantStrategies[i] {
					t.Errorf("family[%d].Strategy = %q, want %q (result must be sorted)", i, f.Strategy, tc.wantStrategies[i])
				}
				if want, ok := tc.wantFamilySizes[f.Strategy]; ok && len(f.Cases) != want {
					t.Errorf("family %q: case count = %d, want %d", f.Strategy, len(f.Cases), want)
				}
			}
		})
	}
}

// TestGroupCrossSymbolFamilies_CasesWithinFamilySortedBySymbol verifies that the
// cases within each family are sorted by symbol name for deterministic iteration
// by familyDuplicateSymbols and other callers.
func TestGroupCrossSymbolFamilies_CasesWithinFamilySortedBySymbol(t *testing.T) {
	input := []tvAlignmentCase{
		makeCase("s.pine", "SBERP", "sberp.json"),
		makeCase("s.pine", "AAPL", "aapl.json"),
		makeCase("s.pine", "BTC", "btc.json"),
	}
	families := groupCrossSymbolFamilies(input)
	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}
	cases := families[0].Cases
	for i := 1; i < len(cases); i++ {
		if cases[i].Symbol < cases[i-1].Symbol {
			t.Errorf("cases not sorted: cases[%d].Symbol=%q < cases[%d].Symbol=%q",
				i, cases[i].Symbol, i-1, cases[i-1].Symbol)
		}
	}
}

// TestFamilyOHLCVKeysBySymbol verifies that the keys map is keyed by symbol,
// that distinct bar content produces distinct keys, that identical bar content
// across symbols produces the same key (the property exploited by
// familyDuplicateSymbols), and that missing fixture files surface as errors.
func TestFamilyOHLCVKeysBySymbol(t *testing.T) {
	dir := t.TempDir()

	barA := preflightBar{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}
	barB := preflightBar{Time: 2000, Open: 200, High: 205, Low: 195, Close: 202, Volume: 300}

	writePreflightFixtureToDisk(t, dir, "AAPL.json", []preflightBar{barA})
	writePreflightFixtureToDisk(t, dir, "BTC.json", []preflightBar{barB})
	writePreflightFixtureToDisk(t, dir, "SBERP.json", []preflightBar{barA}) // same content as AAPL

	twoDistinct := crossSymbolFamily{
		Strategy: "s.pine",
		Cases: []tvAlignmentCase{
			makeCase("s.pine", "AAPL", "AAPL.json"),
			makeCase("s.pine", "BTC", "BTC.json"),
		},
	}

	t.Run("keys_keyed_by_symbol", func(t *testing.T) {
		keys, err := familyOHLCVKeysBySymbol(twoDistinct, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := keys["AAPL"]; !ok {
			t.Error("key missing for symbol AAPL")
		}
		if _, ok := keys["BTC"]; !ok {
			t.Error("key missing for symbol BTC")
		}
	})

	t.Run("distinct_bar_content_produces_distinct_keys", func(t *testing.T) {
		keys, err := familyOHLCVKeysBySymbol(twoDistinct, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if keys["AAPL"] == keys["BTC"] {
			t.Error("distinct bar content produced the same OHLCV key")
		}
	})

	t.Run("identical_bar_content_produces_identical_key", func(t *testing.T) {
		sameContent := crossSymbolFamily{
			Strategy: "s.pine",
			Cases: []tvAlignmentCase{
				makeCase("s.pine", "AAPL", "AAPL.json"),
				makeCase("s.pine", "SBERP", "SBERP.json"),
			},
		}
		keys, err := familyOHLCVKeysBySymbol(sameContent, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if keys["AAPL"] != keys["SBERP"] {
			t.Error("identical bar content produced different OHLCV keys")
		}
	})

	t.Run("empty_family_returns_empty_map", func(t *testing.T) {
		empty := crossSymbolFamily{Strategy: "s.pine", Cases: nil}
		keys, err := familyOHLCVKeysBySymbol(empty, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(keys) != 0 {
			t.Errorf("empty family produced non-empty map: %v", keys)
		}
	})

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		bad := crossSymbolFamily{
			Strategy: "s.pine",
			Cases:    []tvAlignmentCase{makeCase("s.pine", "MISSING", "no_such_file.json")},
		}
		_, err := familyOHLCVKeysBySymbol(bad, dir)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})
}

// TestFamilyDuplicateSymbols verifies the full structural surface of duplicate
// detection: empty and single-case families (can't have duplicates), two-symbol
// families with distinct or identical content, three-symbol families with one or
// all duplicated, and missing fixture error propagation.
func TestFamilyDuplicateSymbols(t *testing.T) {
	dir := t.TempDir()

	barA := preflightBar{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}
	barB := preflightBar{Time: 2000, Open: 200, High: 205, Low: 195, Close: 202, Volume: 300}
	barC := preflightBar{Time: 3000, Open: 150, High: 155, Low: 145, Close: 152, Volume: 400}

	writePreflightFixtureToDisk(t, dir, "aapl.json", []preflightBar{barA})
	writePreflightFixtureToDisk(t, dir, "btc.json", []preflightBar{barB})
	writePreflightFixtureToDisk(t, dir, "nvda.json", []preflightBar{barC})
	writePreflightFixtureToDisk(t, dir, "sberp.json", []preflightBar{barA})
	writePreflightFixtureToDisk(t, dir, "eth.json", []preflightBar{barA})

	makeFamily := func(cases ...tvAlignmentCase) crossSymbolFamily {
		return crossSymbolFamily{Strategy: "test.pine", Cases: cases}
	}

	t.Run("empty_family_returns_false", func(t *testing.T) {
		_, isDup, err := familyDuplicateSymbols(makeFamily(), dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isDup {
			t.Error("empty family reported as having duplicates")
		}
	})

	t.Run("single_case_returns_false", func(t *testing.T) {
		_, isDup, err := familyDuplicateSymbols(makeFamily(makeCase("test.pine", "AAPL", "aapl.json")), dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isDup {
			t.Error("single-case family reported as having duplicates")
		}
	})

	t.Run("no_duplicates_returns_false", func(t *testing.T) {
		f := makeFamily(
			makeCase("test.pine", "AAPL", "aapl.json"),
			makeCase("test.pine", "BTC", "btc.json"),
		)
		_, isDup, err := familyDuplicateSymbols(f, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isDup {
			t.Error("distinct fixtures reported as duplicate")
		}
	})

	t.Run("identical_ohlcv_with_different_metadata_is_duplicate", func(t *testing.T) {
		f := makeFamily(
			makeCase("test.pine", "AAPL", "aapl.json"),
			makeCase("test.pine", "SBERP", "sberp.json"),
		)
		dup, isDup, err := familyDuplicateSymbols(f, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDup {
			t.Fatal("identical bar content not detected as duplicate")
		}
		if dup[0] != "AAPL" || dup[1] != "SBERP" {
			t.Errorf("duplicate pair = %v, want [AAPL SBERP]", dup)
		}
	})

	t.Run("three_symbols_first_duplicate_pair_reported", func(t *testing.T) {
		f := makeFamily(
			makeCase("test.pine", "AAPL", "aapl.json"),
			makeCase("test.pine", "BTC", "btc.json"),
			makeCase("test.pine", "SBERP", "sberp.json"),
		)
		dup, isDup, err := familyDuplicateSymbols(f, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDup {
			t.Fatal("duplicate not detected in 3-symbol family")
		}
		if dup[0] != "AAPL" || dup[1] != "SBERP" {
			t.Errorf("duplicate pair = %v, want [AAPL SBERP]", dup)
		}
	})

	t.Run("three_all_same_content_first_collision_reported", func(t *testing.T) {
		// AAPL, ETH, SBERP all share identical bar content (barA). Symbol-sort
		// order is AAPL < ETH < SBERP, so the scanner encounters AAPL first,
		// then ETH triggers the collision. The returned pair must be [AAPL, ETH],
		// not [AAPL, SBERP] or [ETH, SBERP] — proving scan stops at the FIRST
		// duplicate pair in traversal order.
		f := makeFamily(
			makeCase("test.pine", "AAPL", "aapl.json"),
			makeCase("test.pine", "ETH", "eth.json"),
			makeCase("test.pine", "SBERP", "sberp.json"),
		)
		dup, isDup, err := familyDuplicateSymbols(f, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !isDup {
			t.Fatal("duplicate not detected in all-same-content 3-symbol family")
		}
		if dup[0] != "AAPL" || dup[1] != "ETH" {
			t.Errorf("first collision = %v, want [AAPL ETH] (scan stops at first pair)", dup)
		}
	})

	t.Run("three_all_distinct_returns_false", func(t *testing.T) {
		f := makeFamily(
			makeCase("test.pine", "AAPL", "aapl.json"),
			makeCase("test.pine", "BTC", "btc.json"),
			makeCase("test.pine", "NVDA", "nvda.json"),
		)
		_, isDup, err := familyDuplicateSymbols(f, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isDup {
			t.Error("3 distinct fixtures reported as duplicate")
		}
	})

	t.Run("missing_fixture_returns_error", func(t *testing.T) {
		f := makeFamily(makeCase("test.pine", "X", "nonexistent.json"))
		_, _, err := familyDuplicateSymbols(f, dir)
		if err == nil {
			t.Error("expected error for missing fixture, got nil")
		}
	})
}
