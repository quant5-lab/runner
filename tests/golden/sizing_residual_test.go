package golden

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
