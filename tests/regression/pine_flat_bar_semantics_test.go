package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// pineSourceFlatBarSemantics: >= and <= are simultaneously true when close == open,
// so a direction-tracking UDF reverses on every flat bar — correct Pine semantics,
// not a runner defect, but the source of spurious pivots in strategies run against
// fixtures that contain synthetic flat session-open bars (e.g., SBERP-1h.json).
const pineSourceFlatBarSemantics = `//@version=4
strategy("Flat Bar Equality Semantics", overlay=true)
is_up   = close >= open
is_down = close <= open
if is_up and is_down
    strategy.entry("FlatLong", strategy.long, qty=1)
if strategy.position_size > 0 and not (is_up and is_down)
    strategy.close("FlatLong")
`

func TestPineFlatBar_BothGEAndLEFireOnFlatBar(t *testing.T) {
	root := projectRootFromCwd()
	tmpDir := t.TempDir()

	built, ok := codegenAndBuild(t, tmpDir, "flat_bar_semantics", pineSourceFlatBarSemantics, root)
	if !ok {
		t.Fatal("flat-bar semantics strategy codegen/build failed — >= / <= comparison codegen regression suspected")
	}

	fixturePath := filepath.Join(tmpDir, "flat_bar_fixture.json")
	if err := os.WriteFile(fixturePath, []byte(generateFlatBarFixture(5, 20)), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "out.json")
	cmd := exec.Command(built.BinaryPath,
		"-symbol", "TEST",
		"-timeframe", "1h",
		"-data", fixturePath,
		"-output", outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	var result struct {
		Strategy struct {
			Trades []struct {
				Direction string `json:"direction"`
			} `json:"trades"`
			OpenTrades []struct{} `json:"openTrades"`
		} `json:"strategy"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parse output: %v", err)
	}

	if n := len(result.Strategy.OpenTrades); n != 0 {
		t.Errorf("open trades: got %d, want 0 — flat-bar entry must be closed by the following non-flat bar", n)
	}
	if n := len(result.Strategy.Trades); n != 1 {
		t.Fatalf("closed trades: got %d, want 1 — exactly one flat bar must produce exactly one entry+exit", n)
	}
	if dir := result.Strategy.Trades[0].Direction; dir != "long" {
		t.Errorf("direction = %q, want long", dir)
	}
}

func generateFlatBarFixture(flatIdx, totalBars int) string {
	type Bar struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	}
	type Fixture struct {
		Timezone string `json:"timezone"`
		Bars     []Bar  `json:"bars"`
	}

	const startSec = int64(1640000000)
	const intervalSec = int64(3600)

	bars := make([]Bar, totalBars)
	for i := range bars {
		open := 100.0 + float64(i)
		ts := startSec + int64(i)*intervalSec
		if i == flatIdx {
			bars[i] = Bar{Time: ts, Open: open, High: open, Low: open, Close: open, Volume: 100}
		} else {
			bars[i] = Bar{Time: ts, Open: open, High: open + 1.0, Low: open - 0.5, Close: open + 0.7, Volume: 100}
		}
	}

	out, _ := json.Marshal(Fixture{Timezone: "UTC", Bars: bars})
	return string(out)
}
