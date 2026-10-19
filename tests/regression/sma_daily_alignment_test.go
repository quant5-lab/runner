package regression

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type ohlcvFixtureBar struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// ohlcvFixtureWrapper covers both {"bars":[...]} and plain-array fixture formats.
type ohlcvFixtureWrapper struct {
	Bars []ohlcvFixtureBar `json:"bars"`
}

func barTime(bar ohlcvFixtureBar) time.Time {
	if bar.Time > 1e10 {
		return time.Unix(0, bar.Time*int64(time.Millisecond)).UTC()
	}
	return time.Unix(bar.Time, 0).UTC()
}

// findLastDailyBarBefore returns the index of the rightmost daily bar whose
// timestamp is ≤ cutoff, implementing Pine's barmerge.lookahead_off semantics:
// a security() value on hourly bar H reflects data as of the last completed
// daily bar strictly before H's session open.
// Returns -1 when no daily bar precedes the cutoff.
func findLastDailyBarBefore(daily []ohlcvFixtureBar, cutoff time.Time) int {
	result := -1
	for i, bar := range daily {
		if !barTime(bar).After(cutoff) {
			result = i
		} else {
			break
		}
	}
	return result
}

// Returns NaN when fewer than period bars are available, matching Pine's na warm-up behaviour.
func computeSMAFromBars(daily []ohlcvFixtureBar, endIdx, period int) float64 {
	if endIdx+1 < period {
		return math.NaN()
	}
	sum := 0.0
	for i := endIdx - period + 1; i <= endIdx; i++ {
		sum += daily[i].Close
	}
	return sum / float64(period)
}

// Returns false when either SMA is NaN, matching Pine's na-propagation rules.
func smaBullishDirection(sma20, sma50 float64) (bullish bool, entryType string) {
	bullish = !math.IsNaN(sma20) && !math.IsNaN(sma50) && sma20 > sma50
	if bullish {
		return true, "strategy.Long"
	}
	return false, "strategy.Short"
}

func loadOHLCVBars(t *testing.T, path string) []ohlcvFixtureBar {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", filepath.Base(path), err)
	}

	var w ohlcvFixtureWrapper
	if err := json.Unmarshal(data, &w); err == nil && len(w.Bars) > 0 {
		return w.Bars
	}

	var bars []ohlcvFixtureBar
	if err := json.Unmarshal(data, &bars); err != nil {
		t.Fatalf("parse fixture %s (neither object nor array): %v", filepath.Base(path), err)
	}
	return bars
}

func TestBarTime_UnixSecondsAndMilliseconds(t *testing.T) {
	cases := []struct {
		name      string
		ts        int64
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "unix seconds below threshold",
			ts:        1700000000, // 2023-11-14
			wantYear:  2023,
			wantMonth: time.November,
			wantDay:   14,
		},
		{
			name:      "unix milliseconds above threshold",
			ts:        1700000000000, // same instant in ms
			wantYear:  2023,
			wantMonth: time.November,
			wantDay:   14,
		},
		{
			name:      "zero timestamp",
			ts:        0,
			wantYear:  1970,
			wantMonth: time.January,
			wantDay:   1,
		},
		{
			name:      "boundary: last value still treated as seconds (1e10 - 1)",
			ts:        9_999_999_999,
			wantYear:  2286,
			wantMonth: time.November,
			wantDay:   20,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := barTime(ohlcvFixtureBar{Time: tc.ts})
			if got.Year() != tc.wantYear || got.Month() != tc.wantMonth || got.Day() != tc.wantDay {
				t.Errorf("barTime(%d) = %s, want %d-%02d-%02d",
					tc.ts, got.Format("2006-01-02"), tc.wantYear, int(tc.wantMonth), tc.wantDay)
			}
		})
	}
}

func TestFindLastDailyBarBefore(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	makeBar := func(dayOffset int) ohlcvFixtureBar {
		ts := base.AddDate(0, 0, dayOffset).Unix()
		return ohlcvFixtureBar{Time: ts, Close: float64(dayOffset + 1)}
	}

	daily := []ohlcvFixtureBar{makeBar(0), makeBar(1), makeBar(2)}

	cases := []struct {
		name    string
		cutoff  time.Time
		wantIdx int
	}{
		{
			name:    "cutoff before all daily bars",
			cutoff:  base.Add(-time.Hour),
			wantIdx: -1,
		},
		{
			name:    "cutoff exactly at first bar timestamp",
			cutoff:  base,
			wantIdx: 0,
		},
		{
			name:    "cutoff between bar 0 and bar 1",
			cutoff:  base.Add(12 * time.Hour),
			wantIdx: 0,
		},
		{
			name:    "cutoff exactly at bar 1 timestamp",
			cutoff:  base.AddDate(0, 0, 1),
			wantIdx: 1,
		},
		{
			name:    "cutoff after last bar",
			cutoff:  base.AddDate(0, 0, 10),
			wantIdx: 2,
		},
		{
			name:    "empty daily slice",
			cutoff:  base,
			wantIdx: -1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			slice := daily
			if tc.name == "empty daily slice" {
				slice = nil
			}
			got := findLastDailyBarBefore(slice, tc.cutoff)
			if got != tc.wantIdx {
				t.Errorf("findLastDailyBarBefore cutoff=%s: got %d, want %d",
					tc.cutoff.Format("2006-01-02 15:04"), got, tc.wantIdx)
			}
		})
	}
}

func TestComputeSMAFromBars(t *testing.T) {
	makeBars := func(closes ...float64) []ohlcvFixtureBar {
		bars := make([]ohlcvFixtureBar, len(closes))
		for i, c := range closes {
			bars[i] = ohlcvFixtureBar{Close: c}
		}
		return bars
	}

	cases := []struct {
		name    string
		bars    []ohlcvFixtureBar
		endIdx  int
		period  int
		wantNaN bool
		want    float64
	}{
		{
			name:   "exact period window",
			bars:   makeBars(1, 2, 3, 4, 5),
			endIdx: 4,
			period: 5,
			want:   3.0,
		},
		{
			name:    "window larger than available bars",
			bars:    makeBars(10, 20),
			endIdx:  1,
			period:  5,
			wantNaN: true,
		},
		{
			name:   "period of one",
			bars:   makeBars(42),
			endIdx: 0,
			period: 1,
			want:   42.0,
		},
		{
			name:   "mid-slice window (period=3, endIdx=4 of 7)",
			bars:   makeBars(1, 2, 3, 4, 5, 6, 7),
			endIdx: 4,
			period: 3,
			want:   4.0, // avg(3,4,5)
		},
		{
			name:    "warm-up: endIdx 0 with period 2",
			bars:    makeBars(100),
			endIdx:  0,
			period:  2,
			wantNaN: true,
		},
		{
			name:    "period exactly one more than available",
			bars:    makeBars(10, 20, 30),
			endIdx:  2,
			period:  4,
			wantNaN: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeSMAFromBars(tc.bars, tc.endIdx, tc.period)
			if tc.wantNaN {
				if !math.IsNaN(got) {
					t.Errorf("want NaN, got %.4f", got)
				}
				return
			}
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("got %.6f, want %.6f", got, tc.want)
			}
		})
	}
}

func TestSmaBullishDirection(t *testing.T) {
	nan := math.NaN()
	cases := []struct {
		name        string
		sma20       float64
		sma50       float64
		wantBullish bool
		wantEntry   string
	}{
		{
			name:        "SMA20 above SMA50 → Long",
			sma20:       310.0,
			sma50:       305.0,
			wantBullish: true,
			wantEntry:   "strategy.Long",
		},
		{
			name:        "SMA20 below SMA50 → Short",
			sma20:       300.0,
			sma50:       305.0,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
		{
			name:        "SMA20 equal to SMA50 → Short (not strictly bullish)",
			sma20:       305.0,
			sma50:       305.0,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
		{
			name:        "SMA20 is NaN → Short (na propagation)",
			sma20:       nan,
			sma50:       305.0,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
		{
			name:        "SMA50 is NaN → Short (na propagation)",
			sma20:       305.0,
			sma50:       nan,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
		{
			name:        "both NaN → Short",
			sma20:       nan,
			sma50:       nan,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
		{
			name:        "large spread bullish",
			sma20:       1000.0,
			sma50:       1.0,
			wantBullish: true,
			wantEntry:   "strategy.Long",
		},
		{
			name:        "tiny difference bearish",
			sma20:       100.0001,
			sma50:       100.0002,
			wantBullish: false,
			wantEntry:   "strategy.Short",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotBullish, gotEntry := smaBullishDirection(tc.sma20, tc.sma50)
			if gotBullish != tc.wantBullish {
				t.Errorf("bullish: got %v, want %v", gotBullish, tc.wantBullish)
			}
			if gotEntry != tc.wantEntry {
				t.Errorf("entryType: got %q, want %q", gotEntry, tc.wantEntry)
			}
		})
	}
}

func TestLoadOHLCVBars_BothFixtureFormats(t *testing.T) {
	bar1 := `{"time":1700000000,"open":10,"high":11,"low":9,"close":10.5,"volume":1000}`
	bar2 := `{"time":1700086400,"open":11,"high":12,"low":10,"close":11.5,"volume":2000}`

	cases := []struct {
		name       string
		content    string
		wantLen    int
		wantClose0 float64
	}{
		{
			name:       "plain JSON array",
			content:    "[" + bar1 + "," + bar2 + "]",
			wantLen:    2,
			wantClose0: 10.5,
		},
		{
			name:       "object wrapper with bars key",
			content:    `{"timezone":"Europe/Moscow","bars":[` + bar1 + "," + bar2 + "]}",
			wantLen:    2,
			wantClose0: 10.5,
		},
		{
			name:       "single-bar plain array",
			content:    "[" + bar1 + "]",
			wantLen:    1,
			wantClose0: 10.5,
		},
		{
			name:       "single-bar wrapped object",
			content:    `{"bars":[` + bar1 + "]}",
			wantLen:    1,
			wantClose0: 10.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "data.json")
			if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}
			bars := loadOHLCVBars(t, path)
			if len(bars) != tc.wantLen {
				t.Errorf("len: got %d, want %d", len(bars), tc.wantLen)
			}
			if len(bars) > 0 && bars[0].Close != tc.wantClose0 {
				t.Errorf("bars[0].Close: got %v, want %v", bars[0].Close, tc.wantClose0)
			}
		})
	}
}

func TestFindLastDailyBarBefore_SingleElementSlice(t *testing.T) {
	ts := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	single := []ohlcvFixtureBar{{Time: ts.Unix(), Close: 42}}

	cases := []struct {
		name    string
		cutoff  time.Time
		wantIdx int
	}{
		{"cutoff before the one bar", ts.Add(-time.Second), -1},
		{"cutoff exactly at the one bar", ts, 0},
		{"cutoff well after the one bar", ts.Add(24 * time.Hour), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findLastDailyBarBefore(single, tc.cutoff)
			if got != tc.wantIdx {
				t.Errorf("got %d, want %d", got, tc.wantIdx)
			}
		})
	}
}

func TestComputeSMAFromBars_ArithmeticAccuracy(t *testing.T) {
	makeBars := func(closes ...float64) []ohlcvFixtureBar {
		bars := make([]ohlcvFixtureBar, len(closes))
		for i, c := range closes {
			bars[i] = ohlcvFixtureBar{Close: c}
		}
		return bars
	}

	cases := []struct {
		name   string
		bars   []ohlcvFixtureBar
		endIdx int
		period int
		want   float64
	}{
		{
			name:   "uniform values: SMA equals that value",
			bars:   makeBars(7, 7, 7, 7, 7),
			endIdx: 4,
			period: 5,
			want:   7.0,
		},
		{
			name:   "arithmetic series: SMA equals midpoint",
			bars:   makeBars(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			endIdx: 9,
			period: 10,
			want:   5.5,
		},
		{
			name:   "period-1 window at first bar",
			bars:   makeBars(99.99),
			endIdx: 0,
			period: 1,
			want:   99.99,
		},
		{
			name:   "sliding window does not include bars outside period",
			bars:   makeBars(1000, 2, 3, 4),
			endIdx: 3,
			period: 3,
			want:   3.0, // avg(2,3,4); bar[0]=1000 excluded
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeSMAFromBars(tc.bars, tc.endIdx, tc.period)
			if math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("got %.10f, want %.10f", got, tc.want)
			}
		})
	}
}
