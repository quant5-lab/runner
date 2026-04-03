package golden

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/golden/testutil"
)

var updateGolden = flag.Bool("update-golden", false, "Update golden files with actual results")

type TestSuite struct {
	golden  *testutil.GoldenManager
	runner  *testutil.StrategyRunner
	assets  map[string]testutil.TestAsset
	workDir string
}

func NewTestSuite(t *testing.T) *TestSuite {
	t.Helper()

	workDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Get working directory: %v", err)
	}

	return &TestSuite{
		golden:  testutil.NewGoldenManager(t, *updateGolden),
		runner:  testutil.NewStrategyRunner(t),
		assets:  makeTestAssets(),
		workDir: workDir,
	}
}

func (s *TestSuite) StrategyPath(filename string) string {
	root := s.findWorkspaceRoot()
	return filepath.Join(root, "strategies", filename)
}

func (s *TestSuite) TestFixturePath(filename string) string {
	root := s.findWorkspaceRoot()
	return filepath.Join(root, "e2e", "fixtures", "strategies", filename)
}

func (s *TestSuite) DataPath(filename string) string {
	return s.golden.DataPath(filename)
}

func (s *TestSuite) RunAndValidate(t *testing.T, cfg TestConfig) {
	t.Helper()

	strategyPath := s.StrategyPath(cfg.StrategyFile)
	dataPath := s.golden.EnsureDataFile(t, cfg.DataFile)

	actual := s.runner.Execute(t, strategyPath, dataPath, cfg.Symbol, cfg.Timeframe)

	testutil.PrintTradeSummary(t, actual)

	s.golden.ValidateOrUpdate(t, cfg.GoldenFile, cfg.StrategyName, cfg.DataFile, actual)
}

func (s *TestSuite) RunTestFixtureAndValidate(t *testing.T, cfg TestConfig) {
	t.Helper()

	strategyPath := s.TestFixturePath(cfg.StrategyFile)
	dataPath := s.golden.EnsureDataFile(t, cfg.DataFile)

	actual := s.runner.Execute(t, strategyPath, dataPath, cfg.Symbol, cfg.Timeframe)

	testutil.PrintTradeSummary(t, actual)

	s.golden.ValidateOrUpdate(t, cfg.GoldenFile, cfg.StrategyName, cfg.DataFile, actual)
}

func (s *TestSuite) findWorkspaceRoot() string {
	dir := s.workDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return s.workDir
		}
		dir = parent
	}
}

func makeTestAssets() map[string]testutil.TestAsset {
	return map[string]testutil.TestAsset{
		"AAPL": {
			Symbol:    "AAPL",
			Exchange:  "NYSE",
			Timeframe: "1h",
			DataFile:  "AAPL-1h.json",
		},
		"BTCUSDT": {
			Symbol:    "BTCUSDT",
			Exchange:  "Binance",
			Timeframe: "1h",
			DataFile:  "BTCUSDT-1h.json",
		},
		"SBERP": {
			Symbol:    "SBERP",
			Exchange:  "MOEX",
			Timeframe: "1h",
			DataFile:  "SBERP-1h.json",
		},
	}
}

type TestConfig struct {
	StrategyName string
	StrategyFile string
	Symbol       string
	Timeframe    string
	DataFile     string
	GoldenFile   string
}
