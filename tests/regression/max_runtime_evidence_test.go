package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// minMaxClosedTrades is pinned to the exact golden count so any regression
// that re-zeroes or erodes trades is caught immediately.
const minMaxClosedTrades = 62

func TestMax_SBERP_Hourly_RuntimeEvidence(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "max.pine"))
	if err != nil {
		t.Fatalf("read max strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "max_evidence", string(source), root)
	if !ok {
		t.Fatal("max codegen/build failed")
	}

	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	outputPath := filepath.Join(tmpDir, "out.json")

	cmd := exec.Command(built.BinaryPath,
		"-symbol", "SBERP",
		"-timeframe", "1h",
		"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
		"-output", outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("max run failed (screener fixture not required): %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result struct {
		Strategy struct {
			Trades []struct{} `json:"trades"`
		} `json:"strategy"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	got := len(result.Strategy.Trades)
	if got < minMaxClosedTrades {
		t.Errorf("closed trades: got %d, want >= %d — PMax/Supertrend crossover regression suspected", got, minMaxClosedTrades)
	}
}
