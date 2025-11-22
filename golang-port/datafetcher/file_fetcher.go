package datafetcher

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
)

/* FileFetcher reads OHLCV data from local JSON files */
type FileFetcher struct {
	dataDir string        /* Directory containing JSON files */
	latency time.Duration /* Simulated network latency */
}

/* NewFileFetcher creates fetcher with data directory and simulated latency */
func NewFileFetcher(dataDir string, latency time.Duration) *FileFetcher {
	return &FileFetcher{
		dataDir: dataDir,
		latency: latency,
	}
}

/* Fetch reads OHLCV data from {dataDir}/{symbol}_{timeframe}.json */
func (f *FileFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
	/* Simulate async network delay */
	if f.latency > 0 {
		time.Sleep(f.latency)
	}

	/* Construct file path: BTCUSDT_1D.json */
	filename := fmt.Sprintf("%s/%s_%s.json", f.dataDir, symbol, timeframe)

	/* Read JSON file */
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	/* Parse OHLCV array */
	var bars []context.OHLCV
	if err := json.Unmarshal(data, &bars); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	/* Limit bars if requested */
	if limit > 0 && limit < len(bars) {
		bars = bars[len(bars)-limit:]
	}

	return bars, nil
}
