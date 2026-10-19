package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

/*
TestWeekendEntryDistribution_BB7 ratchets the BB7 phantom-shorts unified
fix: a v4 strategy whose session is "0950-1645" (no ":DAYS" suffix) must
never emit entries on Saturday or Sunday in the exchange timezone. Two
representative fixtures (SBERP/Europe/Moscow, BTCUSDT/UTC) cover both the
non-UTC-timezone and the 24/7-market edge cases.

Symbol-agnostic shape: the test asserts the day-of-week property only; it
does not pin trade counts, prices, or exit semantics so it survives
unrelated tactical changes in BB7.
*/
func TestWeekendEntryDistribution_BB7(t *testing.T) {
	root := projectRootFromCwd()
	source, err := os.ReadFile(filepath.Join(root, "strategies", "bb-strategy-7-rus.pine"))
	if err != nil {
		t.Fatalf("read BB7 strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "bb7_weekend", string(source), root)
	if !ok {
		t.Fatal("BB7 codegen/build failed")
	}

	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	cases := []struct {
		name     string
		symbol   string
		dataFile string
		timezone string
	}{
		{"SBERP_1h_MOEX", "SBERP", "SBERP-1h.json", "Europe/Moscow"},
		{"BTCUSDT_1h_Binance", "BTCUSDT", "BTCUSDT-1h.json", "UTC"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.timezone)
			if err != nil {
				t.Fatalf("load tz %q: %v", tc.timezone, err)
			}

			outputPath := filepath.Join(tmpDir, tc.symbol+".json")
			cmd := exec.Command(built.BinaryPath,
				"-symbol", tc.symbol,
				"-timeframe", "1h",
				"-data", filepath.Join(fixtureDir, tc.dataFile),
				"-datadir", fixtureDir,
				"-output", outputPath,
			)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("run BB7 %s: %v\n%s", tc.symbol, err, out)
			}

			data, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read chart-data: %v", err)
			}
			var chart strategyChartOutput
			if err := json.Unmarshal(data, &chart); err != nil {
				t.Fatalf("parse chart-data: %v", err)
			}

			if len(chart.Strategy.Trades) == 0 {
				t.Fatalf("%s: zero closed trades — phantom-fix ratchet cannot assert weekend distribution", tc.symbol)
			}

			for i, tr := range chart.Strategy.Trades {
				entry := time.Unix(tr.EntryTime, 0).In(loc)
				wd := entry.Weekday()
				if wd == time.Saturday || wd == time.Sunday {
					t.Fatalf("%s trade #%d entry on %s (%s) — v4 default DAYS must reject weekends",
						tc.symbol, i, entry.Format(time.RFC3339), wd)
				}
			}
		})
	}
}
