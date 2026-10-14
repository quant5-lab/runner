package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"
)

// The classification column for the top-10 set, which acceptance row A of the
// parity card asks for.
//
// Each member carries exactly one class from the runner's own outcome space:
// complete, degraded or blocked. The class and the closed-trade count are measured
// on the SBERP 1h golden fixture, the fixture the alignment battery reads. The
// table is checked in both directions: a member on disk that is absent from the
// table fails, and a row that names no member on disk also fails, so a new member
// cannot enter the set unclassified and a removed member cannot leave a stale row.
type top10ClassRow struct {
	Strategy string
	Class    string
	Trades   int
}

var top10Classification = []top10ClassRow{
	{"alpha", "degraded", 545},
	{"ann", "complete", 1452},
	{"aostoch", "degraded", 230},
	{"hull", "degraded", 490},
	{"max", "degraded", 244},
	{"moon", "degraded", 131},
	{"supertrend", "degraded", 536},
	{"ut", "complete", 2535},
	{"ut+", "complete", 2535},
	{"zigzag", "complete", 484},
}

func TestTop10_ClassificationColumnCoversEveryMember(t *testing.T) {
	root := projectRootFromCwd()
	onDisk := []string{}
	paths, err := filepath.Glob(filepath.Join(root, "strategies", "top10", "*.pine"))
	if err != nil {
		t.Fatalf("glob top-10: %v", err)
	}
	for _, p := range paths {
		onDisk = append(onDisk, filepath.Base(p[:len(p)-len(".pine")]))
	}
	sort.Strings(onDisk)

	inTable := []string{}
	for _, r := range top10Classification {
		inTable = append(inTable, r.Strategy)
	}
	sort.Strings(inTable)

	if len(onDisk) != len(inTable) {
		t.Fatalf("the set holds %d members and the column holds %d rows: %v against %v", len(onDisk), len(inTable), onDisk, inTable)
	}
	for i := range onDisk {
		if onDisk[i] != inTable[i] {
			t.Errorf("column mismatch at position %d: on disk %q, in the column %q", i, onDisk[i], inTable[i])
		}
	}
}

func TestTop10_ClassificationColumnMatchesTheRunner(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	for _, row := range top10Classification {
		row := row
		t.Run(row.Strategy, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(root, "strategies", "top10", row.Strategy+".pine"))
			if err != nil {
				t.Fatalf("read strategy: %v", err)
			}
			tmpDir := t.TempDir()
			built, ok := codegenAndBuild(t, tmpDir, "class_"+row.Strategy, string(source), root)
			if !ok {
				t.Fatalf("codegen or build failed for %s", row.Strategy)
			}
			outputPath := filepath.Join(tmpDir, "out.json")
			cmd := exec.Command(built.BinaryPath,
				"-symbol", "SBERP", "-timeframe", "1h",
				"-data", filepath.Join(fixtureDir, "SBERP-1h.json"),
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
				ScriptCompatibility struct {
					Status string `json:"status"`
				} `json:"scriptCompatibility"`
				Strategy struct {
					Trades []struct{} `json:"trades"`
				} `json:"strategy"`
			}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("parse output: %v", err)
			}
			if result.ScriptCompatibility.Status != row.Class {
				t.Errorf("class changed: got %q, want %q", result.ScriptCompatibility.Status, row.Class)
			}
			if got := len(result.Strategy.Trades); got != row.Trades {
				t.Errorf("closed trades changed: got %d, want %d", got, row.Trades)
			}
		})
	}
}
