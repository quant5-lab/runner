package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type strategyDirectionPolicyCase struct {
	name               string
	strategyPath       string
	symbol             string
	timeframe          string
	dataPath           string
	dataDir            string
	forbiddenDirection string
	after              time.Time
}

type strategyChartOutput struct {
	Strategy strategyTradeOutput `json:"strategy"`
}

type strategyTradeOutput struct {
	Trades []strategyTrade `json:"trades"`
}

type strategyTrade struct {
	EntryTime int64  `json:"entryTime"`
	Direction string `json:"direction"`
}

func TestStrategyDirectionPolicy_NoForbiddenDirectionsAfterCutoff(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	regularFixture := regularReferenceSessionFixture(t, fixtureDir, "SBERP-1h.json")

	cases := []strategyDirectionPolicyCase{
		{
			name:               "regular exchange reference fixture rejects forbidden short direction",
			strategyPath:       filepath.Join(root, "strategies", "bb-strategy-7-rus.pine"),
			symbol:             "SBERP",
			timeframe:          "1h",
			dataPath:           regularFixture,
			dataDir:            fixtureDir,
			forbiddenDirection: "short",
			after:              time.Date(2023, time.April, 16, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := runStrategyDirectionPolicyCase(t, root, tc)
			assertNoForbiddenDirectionAfter(t, result.Strategy.Trades, tc.forbiddenDirection, tc.after)
		})
	}
}

func regularReferenceSessionFixture(t *testing.T, dataDir, filename string) string {
	t.Helper()
	source := filepath.Join(dataDir, filename)
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read source fixture: %v", err)
	}

	var fixture struct {
		Timezone string          `json:"timezone"`
		Bars     json.RawMessage `json:"bars"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("parse source fixture: %v", err)
	}
	if len(fixture.Bars) == 0 {
		t.Fatalf("source fixture %s has no bars", source)
	}

	output := struct {
		Timezone         string          `json:"timezone"`
		ReferenceSession string          `json:"referenceSession"`
		Bars             json.RawMessage `json:"bars"`
	}{
		Timezone:         fixture.Timezone,
		ReferenceSession: "regular",
		Bars:             fixture.Bars,
	}

	path := filepath.Join(t.TempDir(), filename)
	encoded, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("marshal regular fixture: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0644); err != nil {
		t.Fatalf("write regular fixture: %v", err)
	}
	return path
}

func runStrategyDirectionPolicyCase(t *testing.T, root string, tc strategyDirectionPolicyCase) strategyChartOutput {
	t.Helper()
	source, err := os.ReadFile(tc.strategyPath)
	if err != nil {
		t.Fatalf("read strategy source: %v", err)
	}

	tmpDir := t.TempDir()
	built, ok := codegenAndBuild(t, tmpDir, "direction_policy", string(source), root)
	if !ok {
		t.Fatal("strategy codegen/build failed")
	}
	outputPath := filepath.Join(tmpDir, "chart-data.json")
	cmd := exec.Command(
		built.BinaryPath,
		"-symbol", tc.symbol,
		"-timeframe", tc.timeframe,
		"-data", tc.dataPath,
		"-datadir", tc.dataDir,
		"-output", outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run strategy fixture: %v\n%s", err, output)
	}

	return readStrategyChartOutput(t, outputPath)
}

func assertNoForbiddenDirectionAfter(t *testing.T, trades []strategyTrade, forbiddenDirection string, after time.Time) {
	t.Helper()
	cutoff := after.Unix()
	for _, trade := range trades {
		if trade.Direction == forbiddenDirection && trade.EntryTime > cutoff {
			t.Fatalf("%s trade after cutoff: entryTime=%d cutoff=%d", forbiddenDirection, trade.EntryTime, cutoff)
		}
	}
}

func readStrategyChartOutput(t *testing.T, path string) strategyChartOutput {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read chart output: %v", err)
	}
	var output strategyChartOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("parse chart output: %v", err)
	}
	return output
}
