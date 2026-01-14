package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/* FetchTestData fetches market data for testing using Node.js data fetchers.
 * Automatically downloads data if not cached in testdata/ohlcv/.
 *
 * This ensures tests are self-contained and can fetch required data on-demand.
 * Data is cached to avoid repeated network calls in local development.
 *
 * Example:
 *   dataFile := testutil.FetchTestData(t, "SPY", "M", 120) // 10 years monthly
 *   dataFile := testutil.FetchTestData(t, "BTCUSDT", "1h", 500) // 500 hours
 */
func FetchTestData(t *testing.T, symbol, timeframe string, bars int) string {
	t.Helper()

	projectRoot := findProjectRoot(t)
	testdataDir := filepath.Join(projectRoot, "testdata", "ohlcv")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}

	normTimeframe := timeframe
	if timeframe == "D" {
		normTimeframe = "1D"
	} else if timeframe == "W" {
		normTimeframe = "1W"
	} else if timeframe == "M" {
		normTimeframe = "1M"
	}

	dataFile := filepath.Join(testdataDir, fmt.Sprintf("%s_%s.json", symbol, normTimeframe))

	t.Logf("📡 Fetching %d bars of %s %s data...", bars, symbol, timeframe)

	nodeCmd := fmt.Sprintf(`
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('%s', '%s', %d, '%s')
      .then(result => {
        console.log('Done: ' + result.message);
      })
      .catch(err => {
        console.error('Error:', err.message);
        process.exit(1);
      });
  });
});`, symbol, timeframe, bars, dataFile)

	fetchCmd := exec.Command("node", "-e", nodeCmd)
	fetchCmd.Dir = projectRoot
	output, err := fetchCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to fetch data: %v\nOutput: %s", err, output)
	}

	t.Logf("%s", output)
	return dataFile
}

/* findProjectRoot locates the project root directory by searching for go.mod */
func findProjectRoot(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	dir := cwd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root (go.mod) from: %s", cwd)
		}
		dir = parent
	}
}
