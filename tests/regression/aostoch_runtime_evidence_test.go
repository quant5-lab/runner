package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// minAostochClosedTrades is the verified closed-trade count for aostoch on SBERP-1h.
// Regressions in exit codegen or OnBarMetrics emission will drop it below this floor.
const minAostochClosedTrades = 42

func TestAostoch_SBERP_Hourly_RuntimeEvidence(t *testing.T) {
	root := projectRootFromCwd()

	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "aostoch.pine"))
	if err != nil {
		t.Fatalf("read aostoch strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "aostoch_evidence", string(source), root)
	if !ok {
		t.Fatal("aostoch codegen/build failed")
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
		t.Fatalf("aostoch run failed: %v\n%s", err, out)
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
	if got < minAostochClosedTrades {
		t.Errorf("closed trades: got %d, want >= %d — strategy.exit stop/limit regression suspected (exits not firing)", got, minAostochClosedTrades)
	}
}
