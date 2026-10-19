package tv_reference

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type fixtureWrapper struct {
	Bars []struct {
		Time int64 `json:"time"`
	} `json:"bars"`
}

func FixtureWindow(path string) (start, end time.Time, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("read fixture %s: %w", path, err)
	}
	var w fixtureWrapper
	if err := json.Unmarshal(data, &w); err != nil || len(w.Bars) == 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("parse fixture bars %s: %w", path, err)
	}
	first := w.Bars[0].Time
	last := w.Bars[len(w.Bars)-1].Time
	if first > 1e10 {
		first /= 1000
	}
	if last > 1e10 {
		last /= 1000
	}
	return time.Unix(first, 0).UTC(), time.Unix(last, 0).UTC(), nil
}
