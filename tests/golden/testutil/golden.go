package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type GoldenManager struct {
	fixturesRoot string
	updateMode   bool
}

func NewGoldenManager(t *testing.T, updateMode bool) *GoldenManager {
	t.Helper()

	workspaceRoot, err := findWorkspaceRoot()
	if err != nil {
		t.Fatalf("Find workspace root: %v", err)
	}

	return &GoldenManager{
		fixturesRoot: filepath.Join(workspaceRoot, "tests", "golden", "fixtures"),
		updateMode:   updateMode,
	}
}

func (m *GoldenManager) DataPath(filename string) string {
	return filepath.Join(m.fixturesRoot, "data", filename)
}

func (m *GoldenManager) ExpectedPath(filename string) string {
	return filepath.Join(m.fixturesRoot, "expected", filename)
}

func (m *GoldenManager) LoadExpected(t *testing.T, filename string) *StrategyResult {
	t.Helper()

	path := m.ExpectedPath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && m.updateMode {
			return nil
		}
		t.Fatalf("Read golden file %s: %v", filename, err)
	}

	var golden GoldenFile
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("Parse golden file %s: %v", filename, err)
	}

	return golden.StrategyResult
}

func (m *GoldenManager) SaveGolden(t *testing.T, filename, strategyName, dataSource string, result *StrategyResult) {
	t.Helper()

	golden := GoldenFile{
		Version:        "1.0",
		Strategy:       strategyName,
		DataSource:     dataSource,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
		StrategyResult: result,
	}

	data, err := json.MarshalIndent(golden, "", "  ")
	if err != nil {
		t.Fatalf("Marshal golden file: %v", err)
	}

	path := m.ExpectedPath(filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Create golden directory: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Write golden file %s: %v", filename, err)
	}

	t.Logf("Saved golden file: %s", path)
}

func (m *GoldenManager) LoadMarketData(t *testing.T, filename string) *MarketData {
	t.Helper()

	path := m.DataPath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Read market data %s: %v", filename, err)
	}

	var marketData MarketData
	if err := json.Unmarshal(data, &marketData); err != nil {
		var bars []Bar
		if err := json.Unmarshal(data, &bars); err != nil {
			t.Fatalf("Parse market data %s: %v", filename, err)
		}
		marketData.Bars = bars
	}

	if len(marketData.Bars) == 0 {
		t.Fatalf("Market data %s contains no bars", filename)
	}

	return &marketData
}

func (m *GoldenManager) EnsureDataFile(t *testing.T, filename string) string {
	t.Helper()

	path := m.DataPath(filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Market data file %s not found - run data generation first", filename)
	}

	return path
}

func (m *GoldenManager) ValidateOrUpdate(t *testing.T, goldenFile, strategyName, dataSource string, actual *StrategyResult) {
	t.Helper()

	if m.updateMode {
		m.SaveGolden(t, goldenFile, strategyName, dataSource, actual)
		t.Logf("Updated golden file: %s", goldenFile)
		return
	}

	expected := m.LoadExpected(t, goldenFile)
	if expected == nil {
		t.Fatalf("Golden file %s not found - run with -update-golden flag to generate", goldenFile)
	}

	if err := CompareResults(expected, actual); err != nil {
		t.Fatalf("Golden file mismatch:\n%v", err)
	}
}

func CompareResults(expected, actual *StrategyResult) error {
	if len(expected.Trades) != len(actual.Trades) {
		return fmt.Errorf("trade count: expected %d, got %d", len(expected.Trades), len(actual.Trades))
	}

	for i := range expected.Trades {
		if err := compareTrade(i, &expected.Trades[i], &actual.Trades[i]); err != nil {
			return err
		}
	}

	// Compare equity
	if !floatEquals(expected.Equity, actual.Equity, 0.01) {
		return fmt.Errorf("equity: expected %.2f, got %.2f", expected.Equity, actual.Equity)
	}

	return nil
}

func compareTrade(index int, expected, actual *Trade) error {
	if expected.EntryBar != actual.EntryBar {
		return fmt.Errorf("trade[%d].entryBar: expected %d, got %d", index, expected.EntryBar, actual.EntryBar)
	}

	if expected.ExitBar != actual.ExitBar {
		return fmt.Errorf("trade[%d].exitBar: expected %d, got %d", index, expected.ExitBar, actual.ExitBar)
	}

	if !floatEquals(expected.EntryPrice, actual.EntryPrice, 0.01) {
		return fmt.Errorf("trade[%d].entryPrice: expected %.2f, got %.2f", index, expected.EntryPrice, actual.EntryPrice)
	}

	if !floatEquals(expected.ExitPrice, actual.ExitPrice, 0.01) {
		return fmt.Errorf("trade[%d].exitPrice: expected %.2f, got %.2f", index, expected.ExitPrice, actual.ExitPrice)
	}

	if !floatEquals(expected.Profit, actual.Profit, 0.01) {
		return fmt.Errorf("trade[%d].profit: expected %.2f, got %.2f", index, expected.Profit, actual.Profit)
	}

	return nil
}

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
