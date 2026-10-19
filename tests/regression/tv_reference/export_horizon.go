package tv_reference

import "time"

// TVExportHorizon returns the latest entry time among all TV trades in the slice.
// It represents the point beyond which the CSV export carries no data.
// Returns the zero time when the slice is empty.
func TVExportHorizon(trades []TVTrade) time.Time {
	var horizon time.Time
	for _, t := range trades {
		if t.EntryUTC.After(horizon) {
			horizon = t.EntryUTC
		}
	}
	return horizon
}

// SplitRunnerAtHorizon partitions runner trades into those whose entry is at or
// before horizon (within) and those whose entry is strictly after (beyond).
// A zero horizon places all trades in within.
func SplitRunnerAtHorizon(runner []RunnerTrade, horizon time.Time) (within, beyond []RunnerTrade) {
	if horizon.IsZero() {
		return runner, nil
	}
	for _, r := range runner {
		if !r.EntryUTC.After(horizon) {
			within = append(within, r)
		} else {
			beyond = append(beyond, r)
		}
	}
	return
}

// ClosedCount returns the number of runner trades that are not open positions.
func ClosedCount(runner []RunnerTrade) int {
	n := 0
	for _, r := range runner {
		if !r.IsOpen {
			n++
		}
	}
	return n
}
