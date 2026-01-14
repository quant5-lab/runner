package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/* ParseGeneratedFilePath extracts the generated Go file path from pine-gen output.
 * Pine-gen now creates unique temp files to support parallel test execution.
 */
func ParseGeneratedFilePath(t *testing.T, pineGenOutput []byte) string {
	t.Helper()

	outputStr := string(pineGenOutput)
	genPrefix := "Generated: "
	startIdx := strings.Index(outputStr, genPrefix)
	if startIdx == -1 {
		t.Fatalf("Could not find 'Generated: ' in pine-gen output: %s", outputStr)
	}
	startIdx += len(genPrefix)
	endIdx := strings.Index(outputStr[startIdx:], "\n")
	if endIdx == -1 {
		endIdx = len(outputStr)
	} else {
		endIdx += startIdx
	}
	return outputStr[startIdx:endIdx]
}

/* FetchTestData fetches market data for testing using Node.js data fetchers.
 * Automatically downloads data if not cached in testdata/ohlcv/.
 *
 * This ensures tests are self-contained and can fetch required data on-demand.
 * Data is cached to avoid repeated network calls in local development.
 *
 * Example:
 *   dataFile := FetchTestData(t, "SPY", "M", 120) // 10 years monthly
 *   dataFile := FetchTestData(t, "BTCUSDT", "1h", 500) // 500 hours
 */
func FetchTestData(t *testing.T, symbol, timeframe string, bars int) string {
	t.Helper()

	// Path to testdata directory
	testdataDir := "../../testdata/ohlcv"
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}

	// Normalize timeframe for filename (D → 1D, W → 1W, M → 1M)
	normTimeframe := timeframe
	if timeframe == "D" {
		normTimeframe = "1D"
	} else if timeframe == "W" {
		normTimeframe = "1W"
	} else if timeframe == "M" {
		normTimeframe = "1M"
	}

	dataFile := filepath.Join(testdataDir, fmt.Sprintf("%s_%s.json", symbol, normTimeframe))

	// Check if data already exists (cached)
	if _, err := os.Stat(dataFile); err == nil {
		t.Logf("✓ Using cached data: %s", dataFile)
		return dataFile
	}

	// Fetch data using Node.js fetchers (Binance/Yahoo/MOEX)
	t.Logf("📡 Fetching %d bars of %s %s data...", bars, symbol, timeframe)

	tmpDir := t.TempDir()
	binanceFile := filepath.Join(tmpDir, "binance.json")
	metadataFile := filepath.Join(tmpDir, "metadata.json")
	standardFile := filepath.Join(tmpDir, "standard.json")

	// Node.js fetch command
	nodeCmd := fmt.Sprintf(`
import('./fetchers/src/container.js').then(({ createContainer }) => {
  import('./fetchers/src/config.js').then(({ createProviderChain, DEFAULTS }) => {
    const container = createContainer(createProviderChain, DEFAULTS);
    const providerManager = container.resolve('providerManager');
    
    providerManager.fetchMarketData('%s', '%s', %d)
      .then(result => {
        const fs = require('fs');
        fs.writeFileSync('%s', JSON.stringify(result.data, null, 2));
        fs.writeFileSync('%s', JSON.stringify({ timezone: result.timezone, provider: result.provider }, null, 2));
        console.log('✓ Fetched ' + result.data.length + ' bars from ' + result.provider);
      })
      .catch(err => {
        console.error('Error fetching data:', err.message);
        process.exit(1);
      });
  });
});`, symbol, timeframe, bars, binanceFile, metadataFile)

	fetchCmd := exec.Command("node", "-e", nodeCmd)
	fetchCmd.Dir = "../../"
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to fetch data: %v\nOutput: %s", err, output)
	}

	// Convert Binance format to standard OHLCV format
	convertCmd := exec.Command("node", "scripts/convert-binance-to-standard.cjs", binanceFile, standardFile, metadataFile)
	convertCmd.Dir = "../../"
	if output, err := convertCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to convert data format: %v\nOutput: %s", err, output)
	}

	// Copy to testdata for caching
	if err := exec.Command("cp", standardFile, dataFile).Run(); err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	t.Logf("✓ Saved data: %s", dataFile)
	return dataFile
}
