package regression

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

type preflightBar struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// preflightFixture pairs parsed bars with the raw JSON map so that
// perturbedFixtureJSON can rebuild a structurally-identical fixture with
// only the flat bars modified, preserving timezone, openDates, and all metadata.
type preflightFixture struct {
	raw  map[string]json.RawMessage
	Bars []preflightBar
}

func loadPreflightFixture(path string) (preflightFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return preflightFixture{}, fmt.Errorf("read %s: %w", path, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return preflightFixture{}, fmt.Errorf("parse %s: %w", path, err)
	}
	barsRaw, ok := raw["bars"]
	if !ok {
		return preflightFixture{}, fmt.Errorf("no bars field in %s", path)
	}
	var bars []preflightBar
	if err := json.Unmarshal(barsRaw, &bars); err != nil {
		return preflightFixture{}, fmt.Errorf("parse bars in %s: %w", path, err)
	}
	return preflightFixture{raw: raw, Bars: bars}, nil
}

// O=H=L=C makes Pine's close>=open and close<=open simultaneously true,
// causing direction-toggle strategies (e.g. zigzag) to reverse on bars
// that carry no price information.
func flatBarIndices(bars []preflightBar) []int {
	var idx []int
	for i, b := range bars {
		if b.Open == b.High && b.High == b.Low && b.Low == b.Close {
			idx = append(idx, i)
		}
	}
	return idx
}

// ohlcvContentKey fingerprints bar OHLCV+time content independently of
// fixture-level metadata (symbol, timezone, period), so two placeholder
// fixtures with identical bar data always produce the same key.
func ohlcvContentKey(bars []preflightBar) string {
	var sb strings.Builder
	for _, b := range bars {
		fmt.Fprintf(&sb, "%d:%.10f:%.10f:%.10f:%.10f:%.10f\n",
			b.Time, b.Open, b.High, b.Low, b.Close, b.Volume)
	}
	h := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(h[:8])
}

func makePreflightFixtureFromBars(bars []preflightBar, meta map[string]string) preflightFixture {
	barsJSON, _ := json.Marshal(bars)
	raw := map[string]json.RawMessage{"bars": barsJSON}
	for k, v := range meta {
		raw[k] = json.RawMessage(`"` + v + `"`)
	}
	copied := make([]preflightBar, len(bars))
	copy(copied, bars)
	return preflightFixture{raw: raw, Bars: copied}
}

func parseBarsFromPerturbedJSON(t *testing.T, data []byte) []preflightBar {
	t.Helper()
	var wrapper struct {
		Bars []preflightBar `json:"bars"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		t.Fatalf("parse perturbed JSON: %v", err)
	}
	return wrapper.Bars
}

// perturbedFixtureJSON breaks the close>=open / close<=open symmetry on flat bars
// by adding epsilon to Close (and adjusting High to maintain High>=Close), without
// touching any non-flat bar or any fixture metadata field.
func (f preflightFixture) perturbedFixtureJSON(epsilon float64) ([]byte, error) {
	patched := make([]preflightBar, len(f.Bars))
	copy(patched, f.Bars)
	for i, b := range patched {
		if b.Open == b.High && b.High == b.Low && b.Low == b.Close {
			patched[i].Close = b.Close + epsilon
			patched[i].High = math.Max(b.High, patched[i].Close)
		}
	}
	barsJSON, err := json.Marshal(patched)
	if err != nil {
		return nil, fmt.Errorf("marshal patched bars: %w", err)
	}
	out := make(map[string]json.RawMessage, len(f.raw))
	for k, v := range f.raw {
		out[k] = v
	}
	out["bars"] = barsJSON
	return json.Marshal(out)
}

var (
	barFlat     = preflightBar{Time: 1000, Open: 100, High: 100, Low: 100, Close: 100, Volume: 500}
	barFlatZero = preflightBar{Time: 5000, Open: 0, High: 0, Low: 0, Close: 0, Volume: 0}
	barNormal   = preflightBar{Time: 2000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 1000}
	barDoji     = preflightBar{Time: 3000, Open: 100, High: 102, Low: 98, Close: 100, Volume: 800}
	barPinBar   = preflightBar{Time: 4000, Open: 100, High: 100, Low: 98, Close: 100, Volume: 600}
)

// TestFlatBarIndices verifies the complete structural surface of flat-bar detection:
// truly flat bars (O=H=L=C) at any price level including zero, doji bars with equal
// open/close but non-zero wicks, bars with one equal pair, and all positional
// configurations (start/end/consecutive/gaps).
func TestFlatBarIndices(t *testing.T) {
	cases := []struct {
		name     string
		bars     []preflightBar
		wantIdxs []int
	}{
		{name: "nil_bars", bars: nil, wantIdxs: nil},
		{name: "empty_bars", bars: []preflightBar{}, wantIdxs: nil},
		{name: "single_flat", bars: []preflightBar{barFlat}, wantIdxs: []int{0}},
		{name: "single_non_flat", bars: []preflightBar{barNormal}, wantIdxs: nil},
		{name: "all_non_flat", bars: []preflightBar{barNormal, barNormal, barNormal}, wantIdxs: nil},
		{name: "all_flat", bars: []preflightBar{barFlat, barFlat, barFlat}, wantIdxs: []int{0, 1, 2}},
		{name: "flat_at_start", bars: []preflightBar{barFlat, barNormal, barNormal}, wantIdxs: []int{0}},
		{name: "flat_at_end", bars: []preflightBar{barNormal, barNormal, barFlat}, wantIdxs: []int{2}},
		{name: "flat_in_middle", bars: []preflightBar{barNormal, barFlat, barNormal}, wantIdxs: []int{1}},
		{name: "consecutive_flat", bars: []preflightBar{barNormal, barFlat, barFlat, barNormal}, wantIdxs: []int{1, 2}},
		{name: "flat_with_gap", bars: []preflightBar{barFlat, barNormal, barFlat}, wantIdxs: []int{0, 2}},
		{name: "doji_with_wicks_not_flat", bars: []preflightBar{barDoji}, wantIdxs: nil},
		{name: "pinbar_high_equals_open_not_flat", bars: []preflightBar{barPinBar}, wantIdxs: nil},
		{name: "mixed_flat_and_doji", bars: []preflightBar{barFlat, barDoji, barFlat}, wantIdxs: []int{0, 2}},
		// Zero-price O=H=L=C=0 is a structurally flat bar regardless of price level;
		// the detection must not treat zero as a special sentinel.
		{name: "zero_price_flat_bar", bars: []preflightBar{barFlatZero}, wantIdxs: []int{0}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := flatBarIndices(tc.bars)
			if len(got) != len(tc.wantIdxs) {
				t.Fatalf("len(got)=%d want %d: got %v", len(got), len(tc.wantIdxs), got)
			}
			for i, idx := range got {
				if idx != tc.wantIdxs[i] {
					t.Errorf("index[%d] = %d, want %d", i, idx, tc.wantIdxs[i])
				}
			}
		})
	}
}

// TestOHLCVContentKey verifies that the fingerprint changes for any field
// modification, is stable across repeated calls, and is order-sensitive
// (different bar order → different key). It also verifies that two independent
// slices with identical values produce the same key — the core property relied
// upon by familyDuplicateSymbols to detect copied fixtures.
func TestOHLCVContentKey(t *testing.T) {
	base := []preflightBar{{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}}

	mutate := func(src preflightBar, apply func(*preflightBar)) []preflightBar {
		cp := src
		apply(&cp)
		return []preflightBar{cp}
	}

	baseKey := ohlcvContentKey(base)

	t.Run("idempotent", func(t *testing.T) {
		if ohlcvContentKey(base) != baseKey {
			t.Error("two calls with identical input produced different keys")
		}
	})
	t.Run("empty_bars_stable", func(t *testing.T) {
		k1 := ohlcvContentKey(nil)
		k2 := ohlcvContentKey([]preflightBar{})
		if k1 != k2 {
			t.Errorf("nil and empty bars produced different keys: %q vs %q", k1, k2)
		}
	})
	t.Run("two_independent_equal_slices_same_key", func(t *testing.T) {
		// Two independently constructed slices with identical values must produce
		// the same key — this is the invariant that lets familyDuplicateSymbols
		// detect copied fixtures loaded from separate files.
		bars1 := []preflightBar{{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}}
		bars2 := []preflightBar{{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}}
		if ohlcvContentKey(bars1) != ohlcvContentKey(bars2) {
			t.Error("independent slices with identical values produced different keys")
		}
	})
	t.Run("close_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.Close += 0.01 })) == baseKey {
			t.Error("close change not detected")
		}
	})
	t.Run("open_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.Open += 0.01 })) == baseKey {
			t.Error("open change not detected")
		}
	})
	t.Run("high_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.High += 0.01 })) == baseKey {
			t.Error("high change not detected")
		}
	})
	t.Run("low_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.Low -= 0.01 })) == baseKey {
			t.Error("low change not detected")
		}
	})
	t.Run("volume_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.Volume += 1 })) == baseKey {
			t.Error("volume change not detected")
		}
	})
	t.Run("time_change_detected", func(t *testing.T) {
		if ohlcvContentKey(mutate(base[0], func(b *preflightBar) { b.Time += 1 })) == baseKey {
			t.Error("time change not detected")
		}
	})
	t.Run("bar_order_sensitive", func(t *testing.T) {
		b1 := preflightBar{Time: 1000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}
		b2 := preflightBar{Time: 2000, Open: 200, High: 205, Low: 195, Close: 202, Volume: 300}
		k1 := ohlcvContentKey([]preflightBar{b1, b2})
		k2 := ohlcvContentKey([]preflightBar{b2, b1})
		if k1 == k2 {
			t.Error("bar order must affect key; forward and reversed sequences must differ")
		}
	})
}

// TestPerturbedFixtureJSON verifies that flat O=H=L=C bars are modified exactly
// as required and that all other bars and all metadata fields are preserved verbatim.
// Epsilon=0 is a degenerate case where the perturbation is a no-op even for flat bars.
func TestPerturbedFixtureJSON(t *testing.T) {
	const eps = 0.001

	flat := preflightBar{Time: 1000, Open: 50, High: 50, Low: 50, Close: 50, Volume: 200}
	normal := preflightBar{Time: 2000, Open: 100, High: 102, Low: 98, Close: 101, Volume: 500}
	doji := preflightBar{Time: 3000, Open: 100, High: 103, Low: 97, Close: 100, Volume: 400}

	t.Run("non_flat_bar_unchanged", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{normal}, nil)
		out, err := f.perturbedFixtureJSON(eps)
		if err != nil {
			t.Fatalf("perturbedFixtureJSON: %v", err)
		}
		bars := parseBarsFromPerturbedJSON(t, out)
		if len(bars) != 1 {
			t.Fatalf("bar count changed: got %d want 1", len(bars))
		}
		if bars[0] != normal {
			t.Errorf("non-flat bar modified: got %+v want %+v", bars[0], normal)
		}
	})

	t.Run("doji_not_treated_as_flat", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{doji}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		if bars[0] != doji {
			t.Errorf("doji bar (O==C, wicks present) must not be perturbed: got %+v want %+v", bars[0], doji)
		}
	})

	t.Run("flat_bar_close_shifted_by_epsilon", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		want := flat.Close + eps
		if bars[0].Close != want {
			t.Errorf("flat bar Close = %.6f, want %.6f", bars[0].Close, want)
		}
	})

	t.Run("flat_bar_high_at_least_close", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		if bars[0].High < bars[0].Close {
			t.Errorf("High (%.6f) < Close (%.6f): OHLC invariant violated", bars[0].High, bars[0].Close)
		}
	})

	t.Run("flat_bar_open_low_time_volume_unchanged", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		b := bars[0]
		if b.Open != flat.Open {
			t.Errorf("Open changed: got %.6f want %.6f", b.Open, flat.Open)
		}
		if b.Low != flat.Low {
			t.Errorf("Low changed: got %.6f want %.6f", b.Low, flat.Low)
		}
		if b.Time != flat.Time {
			t.Errorf("Time changed: got %d want %d", b.Time, flat.Time)
		}
		if b.Volume != flat.Volume {
			t.Errorf("Volume changed: got %.6f want %.6f", b.Volume, flat.Volume)
		}
	})

	t.Run("zero_flat_bars_all_unchanged", func(t *testing.T) {
		bars := []preflightBar{normal, doji}
		f := makePreflightFixtureFromBars(bars, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		got := parseBarsFromPerturbedJSON(t, out)
		for i, b := range got {
			if b != bars[i] {
				t.Errorf("bar[%d] modified when no flat bars present: got %+v want %+v", i, b, bars[i])
			}
		}
	})

	t.Run("all_flat_bars_all_perturbed", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat, flat, flat}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		for i, b := range bars {
			if b.Close == flat.Close {
				t.Errorf("bar[%d] flat bar not perturbed", i)
			}
		}
	})

	t.Run("only_flat_bars_perturbed_in_mixed_fixture", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{normal, flat, normal}, nil)
		out, _ := f.perturbedFixtureJSON(eps)
		bars := parseBarsFromPerturbedJSON(t, out)
		if bars[0] != normal {
			t.Errorf("bar[0] non-flat bar was modified")
		}
		if bars[1].Close == flat.Close {
			t.Errorf("bar[1] flat bar was not perturbed")
		}
		if bars[2] != normal {
			t.Errorf("bar[2] non-flat bar was modified")
		}
	})

	t.Run("metadata_field_preserved", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat}, map[string]string{"timezone": "Europe/Moscow"})
		out, _ := f.perturbedFixtureJSON(eps)
		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal(out, &wrapper); err != nil {
			t.Fatalf("parse output: %v", err)
		}
		got := strings.Trim(string(wrapper["timezone"]), `"`)
		if got != "Europe/Moscow" {
			t.Errorf("timezone = %q, want Europe/Moscow", got)
		}
	})

	t.Run("multiple_metadata_fields_all_preserved", func(t *testing.T) {
		f := makePreflightFixtureFromBars([]preflightBar{flat}, map[string]string{
			"timezone": "Europe/Moscow",
			"period":   "1h",
		})
		out, _ := f.perturbedFixtureJSON(eps)
		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal(out, &wrapper); err != nil {
			t.Fatalf("parse output: %v", err)
		}
		if strings.Trim(string(wrapper["timezone"]), `"`) != "Europe/Moscow" {
			t.Errorf("timezone not preserved: %s", wrapper["timezone"])
		}
		if strings.Trim(string(wrapper["period"]), `"`) != "1h" {
			t.Errorf("period not preserved: %s", wrapper["period"])
		}
	})

	t.Run("epsilon_zero_leaves_flat_bar_unchanged", func(t *testing.T) {
		// epsilon=0 is the identity perturbation: Close+0==Close, max(High,Close)==High.
		// The bar must be structurally identical after the call.
		f := makePreflightFixtureFromBars([]preflightBar{flat}, nil)
		out, _ := f.perturbedFixtureJSON(0.0)
		bars := parseBarsFromPerturbedJSON(t, out)
		if bars[0] != flat {
			t.Errorf("epsilon=0 must not change any bar: got %+v want %+v", bars[0], flat)
		}
	})
}

// TestLoadPreflightFixture covers the fixture loader's contract for valid,
// malformed, and missing inputs.
func TestLoadPreflightFixture(t *testing.T) {
	writeTemp := func(t *testing.T, content string) string {
		t.Helper()
		path := fmt.Sprintf("%s/fixture.json", t.TempDir())
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write temp fixture: %v", err)
		}
		return path
	}

	t.Run("valid_fixture", func(t *testing.T) {
		path := writeTemp(t, `{"timezone":"UTC","bars":[{"time":1000,"open":100,"high":102,"low":98,"close":101,"volume":500}]}`)
		f, err := loadPreflightFixture(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(f.Bars) != 1 {
			t.Errorf("bar count = %d, want 1", len(f.Bars))
		}
		if f.Bars[0].Time != 1000 {
			t.Errorf("bar.Time = %d, want 1000", f.Bars[0].Time)
		}
	})

	t.Run("empty_bars_array_succeeds_with_zero_bars", func(t *testing.T) {
		path := writeTemp(t, `{"timezone":"UTC","bars":[]}`)
		f, err := loadPreflightFixture(path)
		if err != nil {
			t.Fatalf("unexpected error for empty bars array: %v", err)
		}
		if len(f.Bars) != 0 {
			t.Errorf("bar count = %d, want 0", len(f.Bars))
		}
	})

	t.Run("missing_file", func(t *testing.T) {
		_, err := loadPreflightFixture("/nonexistent/path/fixture.json")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})

	t.Run("malformed_json", func(t *testing.T) {
		path := writeTemp(t, `{not valid json}`)
		_, err := loadPreflightFixture(path)
		if err == nil {
			t.Error("expected error for malformed JSON, got nil")
		}
	})

	t.Run("missing_bars_field", func(t *testing.T) {
		path := writeTemp(t, `{"timezone":"UTC"}`)
		_, err := loadPreflightFixture(path)
		if err == nil {
			t.Error("expected error for missing bars field, got nil")
		}
	})

	t.Run("bars_field_wrong_type_returns_error", func(t *testing.T) {
		path := writeTemp(t, `{"timezone":"UTC","bars":"not_an_array"}`)
		_, err := loadPreflightFixture(path)
		if err == nil {
			t.Error("expected error when bars field is not an array, got nil")
		}
	})
}
