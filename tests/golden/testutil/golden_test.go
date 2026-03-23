package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func goldenManagerOverTempDir(t *testing.T) (*GoldenManager, string) {
	t.Helper()
	root := t.TempDir()
	return &GoldenManager{fixturesRoot: root, updateMode: true}, root
}

func writeGoldenFile(t *testing.T, root, filename string, g GoldenFile) {
	t.Helper()
	dir := filepath.Join(root, "expected")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create expected dir: %v", err)
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0644); err != nil {
		t.Fatalf("write golden file: %v", err)
	}
}

func readGoldenFile(t *testing.T, root, filename string) GoldenFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "expected", filename))
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	var g GoldenFile
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("parse golden file: %v", err)
	}
	return g
}

func TestGoldenManager_SaveGolden_CreatesNewFile(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	result := baselineResult()

	mgr.SaveGolden(t, "new.json", "TestStrategy", "data.json", result)

	got := readGoldenFile(t, root, "new.json")
	if got.Version != "1.0" {
		t.Errorf("version: expected %q, got %q", "1.0", got.Version)
	}
	if got.Strategy != "TestStrategy" {
		t.Errorf("strategy: expected %q, got %q", "TestStrategy", got.Strategy)
	}
	if got.DataSource != "data.json" {
		t.Errorf("dataSource: expected %q, got %q", "data.json", got.DataSource)
	}
	if got.GeneratedAt == "" {
		t.Error("generatedAt must not be empty in new file")
	}
	if !ResultsEqual(got.StrategyResult, result) {
		t.Error("saved result does not match input")
	}
}

func TestGoldenManager_SaveGolden_SkipsWriteWhenResultsUnchanged(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	result := baselineResult()
	const pastTimestamp = "2020-01-01T00:00:00Z"

	writeGoldenFile(t, root, "stable.json", GoldenFile{
		Version:        "1.0",
		Strategy:       "TestStrategy",
		DataSource:     "data.json",
		GeneratedAt:    pastTimestamp,
		StrategyResult: result,
	})

	mgr.SaveGolden(t, "stable.json", "TestStrategy", "data.json", result)

	got := readGoldenFile(t, root, "stable.json")
	if got.GeneratedAt != pastTimestamp {
		t.Errorf("timestamp changed on identical results: expected %q preserved, got %q",
			pastTimestamp, got.GeneratedAt)
	}
}

func TestGoldenManager_SaveGolden_OverwritesWhenResultsChange(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	initial := baselineResult()
	const pastTimestamp = "2020-01-01T00:00:00Z"

	writeGoldenFile(t, root, "changing.json", GoldenFile{
		Version:        "1.0",
		Strategy:       "TestStrategy",
		DataSource:     "data.json",
		GeneratedAt:    pastTimestamp,
		StrategyResult: initial,
	})

	updated := baselineResult()
	updated.Trades[0].Profit = 999.00
	updated.Equity = 99999.00

	mgr.SaveGolden(t, "changing.json", "TestStrategy", "data.json", updated)

	got := readGoldenFile(t, root, "changing.json")
	if got.GeneratedAt == pastTimestamp {
		t.Error("expected timestamp to be updated when results changed")
	}
	if !ResultsEqual(got.StrategyResult, updated) {
		t.Error("saved result does not match updated input")
	}
}

func TestGoldenManager_SaveGolden_ToleranceBoundaryOnSkip(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	result := baselineResult()
	const pastTimestamp = "2020-01-01T00:00:00Z"

	writeGoldenFile(t, root, "tolerance.json", GoldenFile{
		Version:        "1.0",
		Strategy:       "TestStrategy",
		DataSource:     "data.json",
		GeneratedAt:    pastTimestamp,
		StrategyResult: result,
	})

	// Profit differs by less than priceTolerance — still considered equal, must not rewrite.
	nearlyIdentical := baselineResult()
	nearlyIdentical.Trades[0].Profit += priceTolerance * 0.5

	mgr.SaveGolden(t, "tolerance.json", "TestStrategy", "data.json", nearlyIdentical)

	got := readGoldenFile(t, root, "tolerance.json")
	if got.GeneratedAt != pastTimestamp {
		t.Errorf("file rewritten for difference within tolerance: expected %q preserved, got %q",
			pastTimestamp, got.GeneratedAt)
	}
}

func TestGoldenManager_SaveGolden_CorruptedFileIsOverwritten(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	dir := filepath.Join(root, "expected")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "corrupt.json"), []byte("not json {{{"), 0644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	result := baselineResult()
	mgr.SaveGolden(t, "corrupt.json", "TestStrategy", "data.json", result)

	got := readGoldenFile(t, root, "corrupt.json")
	if !ResultsEqual(got.StrategyResult, result) {
		t.Error("expected corrupt file to be overwritten with valid result")
	}
}

func TestGoldenManager_LoadExpected_ReturnsNilWhenMissingInUpdateMode(t *testing.T) {
	mgr, _ := goldenManagerOverTempDir(t)

	result := mgr.LoadExpected(t, "nonexistent.json")
	if result != nil {
		t.Errorf("expected nil for missing file in update mode, got non-nil")
	}
}

func TestGoldenManager_LoadExpected_ParsesPersistedResult(t *testing.T) {
	mgr, root := goldenManagerOverTempDir(t)
	result := baselineResult()

	writeGoldenFile(t, root, "existing.json", GoldenFile{
		Version:        "1.0",
		Strategy:       "TestStrategy",
		DataSource:     "data.json",
		GeneratedAt:    "2020-01-01T00:00:00Z",
		StrategyResult: result,
	})

	loaded := mgr.LoadExpected(t, "existing.json")
	if loaded == nil {
		t.Fatal("expected non-nil result for existing file")
	}
	if !ResultsEqual(loaded, result) {
		t.Error("loaded result does not match original")
	}
}
