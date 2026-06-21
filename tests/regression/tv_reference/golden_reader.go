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

	out := make([]RunnerTrade, 0, len(g.Result.Trades)+len(g.Result.OpenTrades))
	for _, t := range g.Result.Trades {
		out = append(out, RunnerTrade{
			EntryUTC:   time.Unix(t.EntryTime, 0).UTC(),
			EntryPrice: t.EntryPrice,
			Direction:  t.Direction,
			Size:       t.Size,
		})
	}
	for _, t := range g.Result.OpenTrades {
		out = append(out, RunnerTrade{
			EntryUTC:   time.Unix(t.EntryTime, 0).UTC(),
			EntryPrice: t.EntryPrice,
			Direction:  t.Direction,
			Size:       t.Size,
			IsOpen:     true,
		})
	}
	return out, nil
}
