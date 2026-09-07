package testutil

import (
	"path/filepath"
	"testing"
)

// measuredSizeResidualBound caps drift between a fresh runner execution and the
// persisted golden file for this focused deterministic-sizing sentinel.
const measuredSizeResidualBound = 3.8e-4

// TestPositionSizing_GoldenDriftWithinBound grounds the financialRelEps =
// 1e-3 tolerance choice in tolerance.go by asserting that runner-vs-golden
// position-size drift stays within the documented empirical bound.
// SBERP/bb9 is used as the canonical fixture because it has a regenerated
// zero-drift baseline; the bound also has to hold for any future drift up to
// the documented headroom.
func TestPositionSizing_GoldenDriftWithinBound(t *testing.T) {
	root, err := findWorkspaceRoot()
	if err != nil {
		t.Fatalf("findWorkspaceRoot: %v", err)
	}

	mgr := NewGoldenManager(t, false)
	expected := mgr.LoadExpected(t, "bb9-sberp-1h.json")
	if expected == nil {
		t.Fatal("golden file bb9-sberp-1h.json not found; cannot ground size-drift bound")
	}

	runner := NewStrategyRunner(t)
	stratPath := filepath.Join(root, "strategies", "bb-strategy-9-rus.pine")
	dataPath := filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json")
	actual := runner.Execute(t, stratPath, dataPath, "SBERP", "1h")

	residual := SizeResidual(expected, actual)
	if residual > measuredSizeResidualBound {
		t.Errorf(
			"position-size residual %.6e exceeds measuredSizeResidualBound %.6e;\n"+
				"  if this is an intentional change, regenerate the golden and reassess the bound\n"+
				"  in docs/audits/VM7_GOLDEN_DELTA_AUDIT.md",
			residual, measuredSizeResidualBound,
		)
	}
}

func TestPositionSizing_Deterministic(t *testing.T) {
	root, err := findWorkspaceRoot()
	if err != nil {
		t.Fatalf("findWorkspaceRoot: %v", err)
	}

	runner := NewStrategyRunner(t)
	stratPath := filepath.Join(root, "strategies", "bb-strategy-9-rus.pine")
	dataPath := filepath.Join(root, "tests", "golden", "fixtures", "data", "SBERP-1h.json")

	first := runner.Execute(t, stratPath, dataPath, "SBERP", "1h")
	second := runner.Execute(t, stratPath, dataPath, "SBERP", "1h")

	for i := range first.Trades {
		if i >= len(second.Trades) {
			break
		}
		if first.Trades[i].Size != second.Trades[i].Size {
			t.Errorf("trades[%d].size: run1=%.16f run2=%.16f — non-deterministic", i,
				first.Trades[i].Size, second.Trades[i].Size)
		}
	}
	for i := range first.OpenTrades {
		if i >= len(second.OpenTrades) {
			break
		}
		if first.OpenTrades[i].Size != second.OpenTrades[i].Size {
			t.Errorf("openTrades[%d].size: run1=%.16f run2=%.16f — non-deterministic", i,
				first.OpenTrades[i].Size, second.OpenTrades[i].Size)
		}
	}
}
