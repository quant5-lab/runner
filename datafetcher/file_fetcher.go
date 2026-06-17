package datafetcher

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/market"
)

// FileFetcher reads OHLCV fixture files from a local directory.
// It probes equivalent timeframe encodings in order (see candidateFixturePaths)
// so "1h" and "60" both resolve to the same fixture regardless of how the file
// was named on disk.
type FileFetcher struct {
	dataDir string
	latency time.Duration
}

// MarketData pairs raw OHLCV bars with their source-level metadata.
type MarketData struct {
	market.SourceMetadata
	Bars []context.OHLCV `json:"bars"`
}

// NewFileFetcher creates a FileFetcher rooted at dataDir. Latency is injected
// only in tests that measure I/O timing; production callers pass 0.
func NewFileFetcher(dataDir string, latency time.Duration) *FileFetcher {
	return &FileFetcher{
		dataDir: dataDir,
		latency: latency,
	}
}

// Fetch returns bars for symbol+timeframe, trimmed to the most-recent limit
// bars when limit > 0.
func (f *FileFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
	marketData, err := f.FetchWithMetadata(symbol, timeframe, limit)
	if err != nil {
		return nil, err
	}
	return marketData.Bars, nil
}

// FetchWithMetadata returns bars and source metadata for symbol+timeframe.
// All equivalent encodings of timeframe are probed in order; the first
// readable file wins (see candidateFixturePaths).
func (f *FileFetcher) FetchWithMetadata(symbol, timeframe string, limit int) (MarketData, error) {
	if f.latency > 0 {
		time.Sleep(f.latency)
	}

	data, resolvedPath, err := f.readFirstAvailable(symbol, timeframe)
	if err != nil {
		return MarketData{}, err
	}

	marketData, err := ParseMarketDataJSON(data)
	if err != nil {
		return MarketData{}, fmt.Errorf("failed to parse %s: %w", resolvedPath, err)
	}

	if limit > 0 && limit < len(marketData.Bars) {
		marketData.Bars = marketData.Bars[len(marketData.Bars)-limit:]
	}

	return marketData, nil
}

func (f *FileFetcher) readFirstAvailable(symbol, timeframe string) (data []byte, path string, err error) {
	candidates := candidateFixturePaths(f.dataDir, symbol, timeframe)
	for _, p := range candidates {
		if d, readErr := os.ReadFile(p); readErr == nil {
			return d, p, nil
		}
	}
	return nil, "", fmt.Errorf("no fixture for %s:%s (tried: %s)",
		symbol, timeframe, strings.Join(candidates, ", "))
}

func ParseMarketDataJSON(data []byte) (MarketData, error) {
	var envelope struct {
		market.SourceMetadata
		Bars    []context.OHLCV  `json:"bars"`
		RawBars *json.RawMessage `json:"-"`
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if barsRaw, ok := raw["bars"]; ok {
			envelope.RawBars = &barsRaw
			if err := json.Unmarshal(data, &envelope); err != nil {
				return MarketData{}, err
			}
			return MarketData{
				SourceMetadata: envelope.SourceMetadata,
				Bars:           envelope.Bars,
			}, nil
		}
	}

	var bars []context.OHLCV
	if err := json.Unmarshal(data, &bars); err != nil {
		return MarketData{}, err
	}
	return MarketData{Bars: bars}, nil
}
