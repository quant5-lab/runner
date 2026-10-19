package regression

import (
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	goldenutil "github.com/quant5-lab/runner/tests/golden/testutil"
)

// zigzagRunOutput is the top-level shape of the binary output JSON used by the
// pivot-evidence test. Only the fields this test reads are mapped.
type zigzagRunOutput struct {
	Candlestick []zigzagCandle             `json:"candlestick"`
	Indicators  map[string]zigzagIndicator `json:"indicators"`
	Strategy    struct {
		Trades []goldenutil.Trade `json:"trades"`
	} `json:"strategy"`
}

type zigzagCandle struct {
	Time int64 `json:"time"`
}

type zigzagIndicator struct {
	Data []zigzagDataPoint `json:"data"`
}

// zigzagDataPoint carries one bar of an indicator series. Value is nil when Pine
// emits na (the binary marshals NaN as JSON null).
type zigzagDataPoint struct {
	Value *float64 `json:"value"`
}

// pivotBoundaryAnchor records a traced RunnerOnly signal bar by its UTC timestamp.
// Binding by timestamp ensures a fixture regeneration that shifts bar indices fails
// loudly at the anchor-resolution step instead of silently re-pointing the check.
type pivotBoundaryAnchor struct {
	timestamp  int64
	label      string
	bcdLo      float64
	bcdHi      float64
	bcdEpsilon float64
}

// sberp1hTracedBoundaryAnchors pins two signal bars whose bcd ratio sits within
// bcdEpsilon of a harmonic pattern's upper bcd edge. Any regression of emitted-
// constant precision re-surfaces here.
//
// Timestamps resolved from the SBERP-1h.json fixture via the binary output
// candlestick array (bar 5377 → 1660114800, bar 7692 → 1680159600).
var sberp1hTracedBoundaryAnchors = []pivotBoundaryAnchor{
	{
		timestamp:  1660114800,
		label:      "2022-08-10 07:00 UTC S@120.22 ABCD(-1)",
		bcdLo:      1.130,
		bcdHi:      2.618,
		bcdEpsilon: 0.020,
	},
	{
		timestamp:  1680159600,
		label:      "2023-03-30 07:00 UTC L@217.35 AntiButterfly(1)",
		bcdLo:      1.000,
		bcdHi:      1.382,
		bcdEpsilon: 0.080,
	},
}

func TestZigzag_SBERP_Hourly_PivotWindowBehavior(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "zigzag.pine"))
	if err != nil {
		t.Fatalf("read zigzag strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "zigzag_pivot", string(source), root)
	if !ok {
		t.Fatal("zigzag codegen/build failed — ticker-string-in-security or self-referential UDF regression suspected")
	}

	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	outputPath := filepath.Join(tmpDir, "out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "SBERP",
		"-timeframe", "1h",
		"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
		"-datadir", fixtureDir,
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zigzag pivot run failed: %v\n%s", err, out)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var result zigzagRunOutput
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("parse output JSON: %v", err)
	}

	ind, ok := result.Indicators["zigzag"]
	if !ok || len(ind.Data) == 0 {
		t.Fatal("zigzag indicator series absent — security()/plot(sz) codegen regression: sz plot not emitted")
	}

	series := indicatorToSeries(ind)
	trades := result.Strategy.Trades
	candleTimes := candleTimestamps(result.Candlestick)

	if countPivots(series, len(series)-1) < 5 {
		t.Fatalf("zigzag series has fewer than 5 non-NaN values — cannot evaluate pivot-window invariants")
	}

	t.Run("WarmupBoundary", func(t *testing.T) {
		assertWarmupBoundary(t, series, trades)
	})
	t.Run("AllSignalBarsHavePivotWindow", func(t *testing.T) {
		assertAllSignalBarsHavePivotWindow(t, series, trades)
	})
	t.Run("BothDirectionsPresent", func(t *testing.T) {
		assertBothDirectionsPresent(t, trades)
	})
	t.Run("TradeCountAboveOccurrenceMutationThreshold", func(t *testing.T) {
		if len(trades) < minZigzagClosedTrades {
			t.Errorf("closed trades: got %d, want >= %d — valuewhen occurrence mutation or harmonic-pattern gate regression",
				len(trades), minZigzagClosedTrades)
		}
	})
	t.Run("NoSimultaneousPositionsSameDirection", func(t *testing.T) {
		assertNoSimultaneousPositions(t, trades)
	})
	t.Run("BoundaryProximityAtTracedAnchors", func(t *testing.T) {
		assertBoundaryAnchorEvidence(t, series, candleTimes, sberp1hTracedBoundaryAnchors)
	})
}

// indicatorToSeries converts the binary output indicator data into a []float64 series
// where JSON null (Pine na) becomes math.NaN.
func indicatorToSeries(ind zigzagIndicator) []float64 {
	s := make([]float64, len(ind.Data))
	for i, pt := range ind.Data {
		if pt.Value == nil {
			s[i] = math.NaN()
		} else {
			s[i] = *pt.Value
		}
	}
	return s
}

// candleTimestamps extracts the Unix-second timestamps from the candlestick array
// in output bar order.
func candleTimestamps(candles []zigzagCandle) []int64 {
	ts := make([]int64, len(candles))
	for i, c := range candles {
		ts[i] = c.Time
	}
	return ts
}

// barIndexForTimestamp returns the bar index whose candle timestamp equals ts, or
// (0, false) when the timestamp is absent. A missing timestamp means the fixture
// was regenerated and the anchor must be recalibrated.
func barIndexForTimestamp(candleTimes []int64, ts int64) (int, bool) {
	for i, t := range candleTimes {
		if t == ts {
			return i, true
		}
	}
	return 0, false
}

// valuewhen(sz,sz,4) returns NaN until five occurrences have accumulated; an entry
// before that bar proves the warmup gate is bypassed.
func assertWarmupBoundary(t *testing.T, series []float64, trades []goldenutil.Trade) {
	t.Helper()
	warmupBar := fifthPivotBar(series)
	if warmupBar < 0 {
		t.Error("fewer than five pivots in zigzag series — cannot evaluate warmup boundary")
		return
	}
	for _, tr := range trades {
		if tr.EntryBar-1 < warmupBar {
			t.Errorf("entryBar=%d (signalBar=%d) fired before fifth pivot at bar %d — valuewhen warmup guard broken",
				tr.EntryBar, tr.EntryBar-1, warmupBar)
		}
	}
}

// A failure means the harmonic pattern evaluated against an incomplete x/a/b/c/d window.
func assertAllSignalBarsHavePivotWindow(t *testing.T, series []float64, trades []goldenutil.Trade) {
	t.Helper()
	failures := 0
	for _, tr := range trades {
		if _, ok := pivotWindowAt(series, tr.EntryBar-1); !ok {
			t.Errorf("entryBar=%d signalBar=%d: fewer than five pivots in sz up to signal bar — harmonic pattern evaluated against incomplete x/a/b/c/d window",
				tr.EntryBar, tr.EntryBar-1)
			failures++
			if failures >= 5 {
				t.Error("(further failures suppressed)")
				return
			}
		}
	}
}

func assertBothDirectionsPresent(t *testing.T, trades []goldenutil.Trade) {
	t.Helper()
	var hasLong, hasShort bool
	for _, tr := range trades {
		switch tr.Direction {
		case "long":
			hasLong = true
		case "short":
			hasShort = true
		}
		if hasLong && hasShort {
			return
		}
	}
	if !hasLong {
		t.Error("no long trades — bullish harmonic pattern evaluation chain broken")
	}
	if !hasShort {
		t.Error("no short trades — bearish harmonic pattern evaluation chain broken")
	}
}

// A failure means the generated runtime does not enforce the strategy's pyramiding=0 declaration.
func assertNoSimultaneousPositions(t *testing.T, trades []goldenutil.Trade) {
	t.Helper()
	check := func(dir string) {
		var group []goldenutil.Trade
		for _, tr := range trades {
			if tr.Direction == dir {
				group = append(group, tr)
			}
		}
		sort.Slice(group, func(i, j int) bool { return group[i].EntryBar < group[j].EntryBar })
		for i := 1; i < len(group); i++ {
			if group[i].EntryBar < group[i-1].ExitBar {
				t.Errorf("%s trade %d (entryBar=%d) overlaps prior %s trade %d (exitBar=%d) — pyramiding=0 violated",
					dir, i, group[i].EntryBar, dir, i-1, group[i-1].ExitBar)
			}
		}
	}
	check("long")
	check("short")
}

// A failure means the sub-tick boundary proximity no longer holds — the pivot series
// shifted by more than the epsilon can absorb, requiring a deliberate cause.
func assertBoundaryAnchorEvidence(t *testing.T, series []float64, candleTimes []int64, anchors []pivotBoundaryAnchor) {
	t.Helper()
	for _, anchor := range anchors {
		bar, ok := barIndexForTimestamp(candleTimes, anchor.timestamp)
		if !ok {
			t.Errorf("anchor %s: timestamp %d not found in candlestick — fixture regenerated; recalibrate anchor",
				anchor.label, anchor.timestamp)
			continue
		}
		w, ok := pivotWindowAt(series, bar)
		if !ok {
			t.Errorf("anchor %s bar %d: fewer than five pivots — boundary evidence unavailable",
				anchor.label, bar)
			continue
		}
		r := harmonicRatiosAt(w)
		if r.bcd < anchor.bcdLo || r.bcd > anchor.bcdHi {
			t.Errorf("anchor %s bar %d: bcd=%.4f outside band [%.3f, %.3f] — pattern gate broken; pivot-window or series regression",
				anchor.label, bar, r.bcd, anchor.bcdLo, anchor.bcdHi)
			continue
		}
		if distFromEdge := anchor.bcdHi - r.bcd; distFromEdge > anchor.bcdEpsilon {
			t.Errorf("anchor %s bar %d: bcd=%.4f is %.4f from upper edge %.3f (want <= %.3f) — boundary proximity claim not held",
				anchor.label, bar, r.bcd, distFromEdge, anchor.bcdHi, anchor.bcdEpsilon)
		}
	}
}

// TestBarIndexForTimestamp verifies timestamp-to-bar-index resolution for all
// boundary conditions: match at start, middle, end; absent timestamp; empty slice.
func TestBarIndexForTimestamp(t *testing.T) {
	times := []int64{1000, 2000, 3000, 4000, 5000}

	tests := []struct {
		name    string
		times   []int64
		ts      int64
		wantIdx int
		wantOK  bool
	}{
		{name: "match_at_start", times: times, ts: 1000, wantIdx: 0, wantOK: true},
		{name: "match_at_middle", times: times, ts: 3000, wantIdx: 2, wantOK: true},
		{name: "match_at_end", times: times, ts: 5000, wantIdx: 4, wantOK: true},
		{name: "absent_timestamp", times: times, ts: 9999, wantIdx: 0, wantOK: false},
		{name: "empty_slice", times: []int64{}, ts: 1000, wantIdx: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, ok := barIndexForTimestamp(tt.times, tt.ts)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && idx != tt.wantIdx {
				t.Errorf("idx = %d, want %d", idx, tt.wantIdx)
			}
		})
	}
}

// TestIndicatorToSeries verifies that nil entries (Pine na) become math.NaN and
// non-nil entries are preserved exactly.
func TestIndicatorToSeries(t *testing.T) {
	val := func(v float64) *float64 { return &v }

	tests := []struct {
		name    string
		input   zigzagIndicator
		wantLen int
		nilAt   []int
		nonNil  map[int]float64
	}{
		{
			name:    "empty_indicator",
			input:   zigzagIndicator{},
			wantLen: 0,
		},
		{
			name:    "all_nil_yields_all_nan",
			input:   zigzagIndicator{Data: []zigzagDataPoint{{nil}, {nil}, {nil}}},
			wantLen: 3,
			nilAt:   []int{0, 1, 2},
		},
		{
			name:    "all_non_nil_preserved",
			input:   zigzagIndicator{Data: []zigzagDataPoint{{val(1.5)}, {val(99.5)}, {val(0.0)}}},
			wantLen: 3,
			nonNil:  map[int]float64{0: 1.5, 1: 99.5, 2: 0.0},
		},
		{
			name: "mixed_nil_and_non_nil",
			input: zigzagIndicator{Data: []zigzagDataPoint{
				{val(120.22)}, {nil}, {val(217.35)}, {nil},
			}},
			wantLen: 4,
			nilAt:   []int{1, 3},
			nonNil:  map[int]float64{0: 120.22, 2: 217.35},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indicatorToSeries(tt.input)
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
			for _, i := range tt.nilAt {
				if !math.IsNaN(got[i]) {
					t.Errorf("got[%d] = %v, want NaN", i, got[i])
				}
			}
			for i, want := range tt.nonNil {
				if got[i] != want {
					t.Errorf("got[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}
