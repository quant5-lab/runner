package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// minZigzagClosedTrades is pinned to the exact golden count so any regression
// that re-zeroes or erodes the harmonic-pattern trade series is caught immediately.
const minZigzagClosedTrades = 97

// TestZigzag_SBERP_Hourly_RuntimeEvidence verifies that the ZigZag PA Strategy
// (Pine v4, harmonic pattern recognition) compiles and produces at least the
// golden trade count end-to-end.  The strategy uses a ticker string variable in
// security() and a self-referential UDF with series-history lookback
// (_direction[1]) — both generic codegen capabilities whose absence would
// silently zero the trade series or fail compilation.
func TestZigzag_SBERP_Hourly_RuntimeEvidence(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "zigzag.pine"))
	if err != nil {
		t.Fatalf("read zigzag strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "zigzag_evidence", string(source), root)
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
