package regression

import (
	"path/filepath"
	"testing"
	"time"
)

type weekendBarSpec struct {
	barIdx      int
	wantWeekday time.Weekday
}

type weekendBarFixtureCase struct {
	name        string
	fixtureFile string
	timezone    string
	bars        []weekendBarSpec
}

// TestBB7_DisputedBars_AreWeekendDays asserts that each disputed bar index
// falls on a weekend day in its exchange timezone. Covers both MOEX (Moscow
// timezone) and Binance (UTC) fixture files.
func TestBB7_DisputedBars_AreWeekendDays(t *testing.T) {
	root := projectRootFromCwd()
	dir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	cases := []weekendBarFixtureCase{
		{
			name:        "BTCUSDT_1h_Binance",
			fixtureFile: filepath.Join(dir, "BTCUSDT-1h.json"),
			timezone:    "UTC",
			bars: []weekendBarSpec{
				{barIdx: 2181, wantWeekday: time.Saturday},
				{barIdx: 4557, wantWeekday: time.Sunday},
				{barIdx: 4701, wantWeekday: time.Saturday},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := loadOHLCVBars(t, tc.fixtureFile)
			for _, spec := range tc.bars {
				if spec.barIdx >= len(fixture) {
					t.Fatalf("bar %d out of fixture range (len=%d)", spec.barIdx, len(fixture))
				}
				bt := barTime(fixture[spec.barIdx])
				if wd := bt.Weekday(); wd != spec.wantWeekday {
					t.Errorf("bar %d (%s UTC): weekday=%s, want %s",
						spec.barIdx, bt.Format("2006-01-02"), wd, spec.wantWeekday)
				}
			}
		})
	}
}
