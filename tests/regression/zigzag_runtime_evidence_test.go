package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const minZigzagClosedTrades = 50

// TestZigzag_SBERP_Hourly_RuntimeEvidence guards against regressions that
// re-introduce the syminfo_tickerid scope compile failure or silently zero trades
// for the ZigZag PA Strategy (Pine v4, harmonic pattern recognition, ticker via
// string variable in security()).
func TestZigzag_SBERP_Hourly_RuntimeEvidence(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "zigzag.pine"))
	if err != nil {
		t.Fatalf("read zigzag strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "zigzag_evidence", string(source), root)
	if !ok {
		t.Fatal("zigzag codegen/build failed — syminfo_tickerid scope regression suspected")
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
		t.Fatalf("zigzag run failed: %v\n%s", err, out)
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
	if got < minZigzagClosedTrades {
		t.Errorf("closed trades: got %d, want >= %d", got, minZigzagClosedTrades)
	}
}
