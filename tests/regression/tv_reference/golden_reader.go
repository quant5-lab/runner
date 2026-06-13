package tv_reference

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type GoldenTrade struct {
	EntryTime  int64   `json:"entryTime"`
	EntryPrice float64 `json:"entryPrice"`
	Direction  string  `json:"direction"`
	Size       float64 `json:"size"`
}

type goldenResult struct {
	Trades     []GoldenTrade `json:"trades"`
	OpenTrades []GoldenTrade `json:"openTrades"`
}

type goldenFile struct {
	Result goldenResult `json:"result"`
}

func LoadRunnerTrades(path string) ([]RunnerTrade, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read golden %s: %w", path, err)
	}
	var g goldenFile
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("parse golden %s: %w", path, err)
	}

	all := append(g.Result.Trades, g.Result.OpenTrades...)
	out := make([]RunnerTrade, 0, len(all))
	for _, t := range all {
		out = append(out, RunnerTrade{
			EntryUTC:   time.Unix(t.EntryTime, 0).UTC(),
			EntryPrice: t.EntryPrice,
			Direction:  t.Direction,
			Size:       t.Size,
		})
	}
	return out, nil
}
