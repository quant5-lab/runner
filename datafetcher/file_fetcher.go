package datafetcher

import (
	"encoding/json"
	"fmt"
	"os"
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

	filename := fmt.Sprintf("%s/%s_%s.json", f.dataDir, symbol, timeframe)
	data, err := os.ReadFile(filename)
	if err != nil {
		return MarketData{}, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	marketData, err := ParseMarketDataJSON(data)
	if err != nil {
		return MarketData{}, fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	if limit > 0 && limit < len(marketData.Bars) {
		marketData.Bars = marketData.Bars[len(marketData.Bars)-limit:]
	}

	return marketData, nil
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
