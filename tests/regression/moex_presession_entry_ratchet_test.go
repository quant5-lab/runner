package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Guards against SBERP-1h fixture bars at 06:00 MSK — single-price pre-session
// auction prints (O=H=L=C) that market normalisation must drop before strategies run.
// Pins entry-time property only; does not couple to trade counts or profit figures.
func TestPresessionEntryRatchet_UtPlusSBERP(t *testing.T) {
	root := projectRootFromCwd()
	source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", "ut+.pine"))
	if err != nil {
		t.Fatalf("read UtPlus strategy: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "utplus_presession", string(source), root)
	if !ok {
		t.Fatal("UtPlus codegen/build failed")
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
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run UtPlus: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	var chart strategyChartOutput
	if err := json.Unmarshal(data, &chart); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if len(chart.Strategy.Trades) == 0 {
		t.Fatal("zero closed trades — pre-session entry ratchet cannot assert entry-time distribution")
	}

	msk, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load Europe/Moscow: %v", err)
	}

	for i, tr := range chart.Strategy.Trades {
		entry := time.Unix(tr.EntryTime, 0).In(msk)
		if entry.Hour() < 7 {
			t.Fatalf("trade #%d entered at %s — pre-session bar (before 07:00 MSK) not filtered",
				i, entry.Format(time.RFC3339))
		}
	}
}
