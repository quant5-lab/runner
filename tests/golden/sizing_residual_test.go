package golden

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/quant5-lab/runner/tests/golden/testutil"
)

func TestSizeResidual_AllGoldens(t *testing.T) {
	root, err := testutil.FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot: %v", err)
	}
	gm := testutil.NewGoldenManager(t, false)

	for _, tc := range allStrategyGoldens {
		tc := tc
		t.Run(tc.goldenFile, func(t *testing.T) {
			t.Parallel()
			runner := testutil.NewStrategyRunner(t)
			golden := gm.LoadExpected(t, tc.goldenFile)
			live := runner.Execute(t,
				filepath.Join(root, tc.strategyRelPath),
				gm.DataPath(tc.dataFile),
				tc.symbol, tc.timeframe,
			)
			dev := testutil.SizeResidual(golden, live)
			if dev > testutil.MeasuredSizeResidualBound {
				t.Errorf("size deviation %.4e exceeds MeasuredSizeResidualBound %.4e — recalibrate bound or regenerate golden",
					dev, testutil.MeasuredSizeResidualBound)
			}
		})
	}
}

func TestRegistry_CoversAllStrategyGoldens(t *testing.T) {
	gm := testutil.NewGoldenManager(t, false)

	registered := make(map[string]bool, len(allStrategyGoldens))
	for _, tc := range allStrategyGoldens {
		if registered[tc.goldenFile] {
			t.Errorf("duplicate in allStrategyGoldens registry: %s", tc.goldenFile)
		}
		registered[tc.goldenFile] = true
	}

	for _, tc := range allStrategyGoldens {
		result := gm.LoadExpected(t, tc.goldenFile)
		if result == nil || len(result.Trades) == 0 {
			t.Errorf("registry entry %s has no trades — remove from allStrategyGoldens or regenerate golden",
				tc.goldenFile)
		}
	}

	expectedDir := filepath.Dir(gm.ExpectedPath("sentinel"))
	entries, err := os.ReadDir(expectedDir)
	if err != nil {
		t.Fatalf("read expected dir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		if !onDiskGoldenHasTrades(t, gm.ExpectedPath(e.Name())) {
			continue
		}
		if !registered[e.Name()] {
			t.Errorf("strategy golden %s has trades but is absent from allStrategyGoldens — add it to registry.go",
				e.Name())
		}
	}
}

func onDiskGoldenHasTrades(t *testing.T, path string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	var f struct {
		Result struct {
			Trades []struct{} `json:"trades"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("parse golden %s: %v", path, err)
	}
	return len(f.Result.Trades) > 0
}

// TestGoldenTradeCountFidelity_AllGoldens asserts that the live runner's closed-trade
// count is within MaxGoldenTradeCountDrift of the stored golden for every
// registered (strategy, fixture) pair.
//
// SizeResidual cannot catch fixture-window expansions because it compares only
// the overlapping prefix of trades and ignores surplus. This sweep closes that
// gap: a 4x window expansion (e.g. 5499 to 21927 bars, 103 to 422 trades, drift
// approximately 0.76) fails immediately here, forcing explicit golden regeneration
// before the change can be merged.
func TestGoldenTradeCountFidelity_AllGoldens(t *testing.T) {
	root, err := testutil.FindWorkspaceRoot()
	if err != nil {
		t.Fatalf("FindWorkspaceRoot: %v", err)
	}
	gm := testutil.NewGoldenManager(t, false)

	for _, tc := range allStrategyGoldens {
		tc := tc
		t.Run(tc.goldenFile, func(t *testing.T) {
			t.Parallel()
			runner := testutil.NewStrategyRunner(t)
			golden := gm.LoadExpected(t, tc.goldenFile)
			live := runner.Execute(t,
				filepath.Join(root, tc.strategyRelPath),
				gm.DataPath(tc.dataFile),
				tc.symbol, tc.timeframe,
			)
			drift := testutil.TradeCountFidelity(golden, live)
			if drift > testutil.MaxGoldenTradeCountDrift {
				t.Errorf(
					"trade count drift %.4f exceeds MaxGoldenTradeCountDrift %.4f: "+
						"golden=%d live=%d -- regenerate golden or investigate fixture change",
					drift, testutil.MaxGoldenTradeCountDrift,
					len(golden.Trades), len(live.Trades),
				)
			}
		})
	}
}

// TestGoldenBarBoundsConsistency_AllGoldens asserts that every closed-trade entry
// and exit bar index stored in the golden falls within the current fixture's
// bar range [0, barCount). A golden generated against a larger fixture than
// currently on disk would have out-of-range bar indices, signalling that the
// fixture was narrowed without a corresponding golden regen.
func TestGoldenBarBoundsConsistency_AllGoldens(t *testing.T) {
	gm := testutil.NewGoldenManager(t, false)

	for _, tc := range allStrategyGoldens {
		tc := tc
		t.Run(tc.goldenFile, func(t *testing.T) {
			t.Parallel()
			golden := gm.LoadExpected(t, tc.goldenFile)
			fixture := gm.LoadMarketData(t, tc.dataFile)
			barCount := len(fixture.Bars)
			for i, tr := range golden.Trades {
				if tr.EntryBar < 0 || tr.EntryBar >= barCount {
					t.Errorf(
						"trades[%d].entryBar=%d outside fixture %s bar range [0, %d)",
						i, tr.EntryBar, tc.dataFile, barCount,
					)
				}
				if tr.ExitBar < 0 || tr.ExitBar >= barCount {
					t.Errorf(
						"trades[%d].exitBar=%d outside fixture %s bar range [0, %d)",
						i, tr.ExitBar, tc.dataFile, barCount,
					)
				}
			}
		})
	}
}

func goldenRawMeta(t *testing.T, path string) (dataSource, generatedAt string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	var f struct {
		DataSource  string `json:"dataSource"`
		GeneratedAt string `json:"generatedAt"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("parse golden %s: %v", path, err)
	}
	return f.DataSource, f.GeneratedAt
}

// TestGoldenDataSourceMatchesRegistry asserts that every golden file on disk declares
// a dataSource field that matches the data fixture its registry entry names. A mismatch
// means the golden was either generated from a different fixture than the one the test
// will load, or the registry entry was updated without regenerating the golden.
func TestGoldenDataSourceMatchesRegistry(t *testing.T) {
	gm := testutil.NewGoldenManager(t, false)
	for _, tc := range allStrategyGoldens {
		tc := tc
		t.Run(tc.goldenFile, func(t *testing.T) {
			t.Parallel()
			dataSource, _ := goldenRawMeta(t, gm.ExpectedPath(tc.goldenFile))
			if dataSource != tc.dataFile {
				t.Errorf("golden dataSource=%q but registry dataFile=%q — regenerate golden against the correct fixture",
					dataSource, tc.dataFile)
			}
		})
	}
}

// TestGoldenGeneratedAtIsValidRFC3339 asserts that every golden file carries a
// generatedAt timestamp that parses as valid RFC3339. A missing or malformed
// timestamp indicates a corrupt or hand-edited golden file.
func TestGoldenGeneratedAtIsValidRFC3339(t *testing.T) {
	gm := testutil.NewGoldenManager(t, false)
	for _, tc := range allStrategyGoldens {
		tc := tc
		t.Run(tc.goldenFile, func(t *testing.T) {
			t.Parallel()
			_, generatedAt := goldenRawMeta(t, gm.ExpectedPath(tc.goldenFile))
			if generatedAt == "" {
				t.Errorf("generatedAt is empty in %s", tc.goldenFile)
				return
			}
			if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
				t.Errorf("generatedAt %q in %s is not valid RFC3339: %v",
					generatedAt, tc.goldenFile, err)
			}
		})
	}
}
