package datafetcher

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/market"
)

type FileFetcher struct {
	dataDir string
	latency time.Duration
}

type MarketData struct {
	market.SourceMetadata
	Bars []context.OHLCV `json:"bars"`
}

func NewFileFetcher(dataDir string, latency time.Duration) *FileFetcher {
	return &FileFetcher{
		dataDir: dataDir,
		latency: latency,
	}
}

func (f *FileFetcher) Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error) {
	marketData, err := f.FetchWithMetadata(symbol, timeframe, limit)
	if err != nil {
		return nil, err
	}
	return marketData.Bars, nil
}

func (f *FileFetcher) FetchWithMetadata(symbol, timeframe string, limit int) (MarketData, error) {
	if f.latency > 0 {
		time.Sleep(f.latency)
	}

	rawBytes, resolvedPath, err := readFixture(f.dataDir, symbol, timeframe)
	if err != nil {
		return MarketData{}, err
	}

	marketData, err := ParseMarketDataJSON(rawBytes)
	if err != nil {
		return MarketData{}, fmt.Errorf("failed to parse %s: %w", resolvedPath, err)
	}

	if limit > 0 && limit < len(marketData.Bars) {
		marketData.Bars = marketData.Bars[len(marketData.Bars)-limit:]
	}

	return marketData, nil
}

// Hyphen convention (operator default) is probed before underscore (legacy); canonical token before numeric alternate (e.g. "1h" before "60").
func fixtureCandidatePaths(dir, symbol, timeframe string) []string {
	base := filepath.Join(dir, symbol)
	var paths []string
	for _, token := range equivalentTimeframeTokens(timeframe) {
		paths = append(paths, base+"-"+token+".json", base+"_"+token+".json")
	}
	return paths
}

func readFixture(dir, symbol, timeframe string) (data []byte, path string, err error) {
	for _, p := range fixtureCandidatePaths(dir, symbol, timeframe) {
		if d, e := os.ReadFile(p); e == nil {
			return d, p, nil
		}
	}
	candidates := fixtureCandidatePaths(dir, symbol, timeframe)
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
