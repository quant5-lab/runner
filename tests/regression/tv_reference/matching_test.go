package tv_reference

import (
	"testing"
	"time"
)

// makeOpenRunnerTrade creates a long RunnerTrade with IsOpen: true.
func makeOpenRunnerTrade(datetime string, price float64) RunnerTrade {
	return RunnerTrade{
		EntryUTC:   parseUTC(datetime),
		EntryPrice: price,
		Direction:  "long",
		IsOpen:     true,
	}
}

func makeOpenRunnerTradeWithDirection(datetime string, price float64, direction string) RunnerTrade {
	return RunnerTrade{
		EntryUTC:   parseUTC(datetime),
		EntryPrice: price,
		Direction:  direction,
		IsOpen:     true,
	}
}

// TestMatchExact_OpenTradesExcluded verifies that runner trades with IsOpen == true
// participate in neither the matched count nor the runnerOnly count. Open trades must
// not consume TV slots that closed runner trades need.
func TestMatchExact_OpenTradesExcluded(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0

	cases := []struct {
		name           string
		runner         []RunnerTrade
		tv             []TVTrade
		wantMatched    int
		wantRunnerOnly int
		wantTVOnly     int
	}{
		{
			"open_trade_not_counted_as_runner_only",
			[]RunnerTrade{makeOpenRunnerTrade("2025-01-01 10:00", 100)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			0, 0, 1,
		},
		{
			"open_incompatible_with_tv_contributes_nothing",
			[]RunnerTrade{makeOpenRunnerTrade("2025-01-01 10:00", 999)},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			0, 0, 1,
		},
		{
			"open_does_not_steal_tv_slot_from_closed",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeOpenRunnerTrade("2025-01-01 10:00", 100),
			},
			[]TVTrade{makeTVTrade("2025-01-01 10:00", 100)},
			1, 0, 0,
		},
		{
			"closed_and_open_mixed_counts_only_closed",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeRunnerTrade("2025-01-02 10:00", 200),
				makeOpenRunnerTrade("2025-01-03 10:00", 300),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-02 10:00", 200),
			},
			2, 0, 0,
		},
		{
			"open_trade_in_middle_of_sequence_excluded",
			[]RunnerTrade{
				makeRunnerTrade("2025-01-01 10:00", 100),
				makeOpenRunnerTrade("2025-01-02 10:00", 200),
				makeRunnerTrade("2025-01-03 10:00", 300),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-03 10:00", 300),
			},
			2, 0, 0,
		},
		{
			"all_open_runner_is_equivalent_to_empty_runner",
			[]RunnerTrade{
				makeOpenRunnerTrade("2025-01-01 10:00", 100),
				makeOpenRunnerTrade("2025-01-02 10:00", 200),
			},
			[]TVTrade{
				makeTVTrade("2025-01-01 10:00", 100),
				makeTVTrade("2025-01-02 10:00", 200),
			},
			0, 0, 2,
		},
		{
			"open_direction_mismatch_also_excluded",
			[]RunnerTrade{makeOpenRunnerTradeWithDirection("2025-01-01 10:00", 100, "short")},
			[]TVTrade{makeTVTradeWithDirection("2025-01-01 10:00", 100, "long")},
			0, 0, 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			matched, runnerOnly, tvOnly := MatchExact(tc.runner, tc.tv, timeTol, priceTol)
			if matched != tc.wantMatched {
				t.Errorf("matched = %d, want %d", matched, tc.wantMatched)
			}
			if runnerOnly != tc.wantRunnerOnly {
				t.Errorf("runnerOnly = %d, want %d", runnerOnly, tc.wantRunnerOnly)
			}
			if tvOnly != tc.wantTVOnly {
				t.Errorf("tvOnly = %d, want %d", tvOnly, tc.wantTVOnly)
			}
		})
	}
}

// TestMatchSize_InputOrderInvariant verifies that MatchSize returns the same
// result regardless of the order in which runner and TV trades are supplied.
func TestMatchSize_InputOrderInvariant(t *testing.T) {
	timeTol := 2 * time.Hour
	priceTol := 1.0
	relTol := 0.01

	runner := []RunnerTrade{
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", Size: 3.0},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 1.0},
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 2.0},
	}
	tv := []TVTrade{
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 2.0},
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", Size: 3.0},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 1.0},
	}

	reverseRunner := func(s []RunnerTrade) []RunnerTrade {
		cp := append([]RunnerTrade(nil), s...)
		for i, j := 0, len(cp)-1; i < j; i, j = i+1, j-1 {
			cp[i], cp[j] = cp[j], cp[i]
		}
		return cp
	}
	reverseTV := func(s []TVTrade) []TVTrade {
		cp := append([]TVTrade(nil), s...)
		for i, j := 0, len(cp)-1; i < j; i, j = i+1, j-1 {
			cp[i], cp[j] = cp[j], cp[i]
		}
		return cp
	}

	wantMatched, wantMismatch, _ := MatchSize(runner, tv, timeTol, priceTol, relTol)

	cases := []struct {
		name   string
		runner []RunnerTrade
		tv     []TVTrade
	}{
		{"runner_reversed", reverseRunner(runner), tv},
		{"tv_reversed", runner, reverseTV(tv)},
		{"both_reversed", reverseRunner(runner), reverseTV(tv)},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			matched, mismatch, _ := MatchSize(tc.runner, tc.tv, timeTol, priceTol, relTol)
			if matched != wantMatched {
				t.Errorf("matched = %d, want %d", matched, wantMatched)
			}
			if mismatch != wantMismatch {
				t.Errorf("sizeMismatch = %d, want %d", mismatch, wantMismatch)
			}
		})
	}
}

// TestMatchSize_OpenTradesExcluded verifies that open runner trades do not
// participate in size comparison — neither as matched pairs nor via sizeMismatch.
func TestMatchSize_OpenTradesExcluded(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0
	relTol := 0.01

	t.Run("open_trade_produces_no_match", func(t *testing.T) {
		runner := []RunnerTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 5.0, IsOpen: true},
		}
		tv := []TVTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 5.0},
		}
		matched, sizeMismatch, _ := MatchSize(runner, tv, timeTol, priceTol, relTol)
		if matched != 0 {
			t.Errorf("matched = %d, want 0 (open trade excluded)", matched)
		}
		if sizeMismatch != 0 {
			t.Errorf("sizeMismatch = %d, want 0", sizeMismatch)
		}
	})

	t.Run("open_size_discrepancy_invisible_to_metric", func(t *testing.T) {
		runner := []RunnerTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 1.0},
			{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 999.0, IsOpen: true},
		}
		tv := []TVTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", Size: 1.0},
			{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", Size: 1.0},
		}
		matched, sizeMismatch, _ := MatchSize(runner, tv, timeTol, priceTol, relTol)
		if matched != 1 {
			t.Errorf("matched = %d, want 1 (only the closed trade)", matched)
		}
		if sizeMismatch != 0 {
			t.Errorf("sizeMismatch = %d, want 0 (open size discrepancy must be invisible)", sizeMismatch)
		}
	})
}

// TestMatchNetPnL_OpenTradesExcluded verifies that open runner trades do not
// participate in PnL comparison.
func TestMatchNetPnL_OpenTradesExcluded(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0
	relTol := 0.01

	t.Run("open_trade_with_matching_tv_not_counted_as_matched", func(t *testing.T) {
		runner := []RunnerTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 0, IsOpen: true},
		}
		tv := []TVTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 50},
		}
		matched, pnlMismatch := MatchNetPnL(runner, tv, 1.0, timeTol, priceTol, relTol)
		if matched != 0 {
			t.Errorf("matched = %d, want 0 (open trade excluded)", matched)
		}
		if pnlMismatch != 0 {
			t.Errorf("pnlMismatch = %d, want 0", pnlMismatch)
		}
	})

	t.Run("open_pnl_mismatch_invisible_to_metric", func(t *testing.T) {
		runner := []RunnerTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 50},
			{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", NetPnL: 0, IsOpen: true},
		}
		tv := []TVTrade{
			{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 50},
			{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "short", NetPnL: 50},
		}
		matched, pnlMismatch := MatchNetPnL(runner, tv, 1.0, timeTol, priceTol, relTol)
		if matched != 1 {
			t.Errorf("matched = %d, want 1 (only the closed trade)", matched)
		}
		if pnlMismatch != 0 {
			t.Errorf("pnlMismatch = %d, want 0 (open PnL gap invisible)", pnlMismatch)
		}
	})
}

// TestMatchExact_ToleranceOverlapCollision pins the behaviour when two runner trades
// are both compatible with the same TV trade but only one of them has an alternative.
//
// Setup (timeTol=2h, priceTol=1.0):
//
//	R_multi  t=09:59  p=101.0 long  →  compat T1 (Δt=1min, Δp=0.5) AND T2 (Δt=91min, Δp=0.5)
//	R_single t=10:00  p=100.0 long  →  compat T1 (Δt=0,    Δp=0.5)  NOT T2 (Δp=1.5 > 1.0)
//	T1       t=10:00  p=100.5 long
//	T2       t=11:30  p=101.5 long
//
// Greedy (time-sorted, R_multi processed first) steals T1, leaving R_single with no
// compatible slot → matched=1.  Maximum bipartite matching re-routes R_multi to T2,
// freeing T1 for R_single → matched=2.
func TestMatchExact_ToleranceOverlapCollision(t *testing.T) {
	timeTol := 2 * time.Hour
	priceTol := 1.0

	runner := []RunnerTrade{
		makeRunnerTradeWithDirection("2025-01-01 09:59", 101.0, "long"),
		makeRunnerTradeWithDirection("2025-01-01 10:00", 100.0, "long"),
	}
	tv := []TVTrade{
		makeTVTradeWithDirection("2025-01-01 10:00", 100.5, "long"),
		makeTVTradeWithDirection("2025-01-01 11:30", 101.5, "long"),
	}

	matched, runnerOnly, tvOnly := MatchExact(runner, tv, timeTol, priceTol)
	if matched != 2 {
		t.Errorf("matched = %d, want 2 (max matching must re-route R_multi to T2)", matched)
	}
	if runnerOnly != 0 {
		t.Errorf("runnerOnly = %d, want 0", runnerOnly)
	}
	if tvOnly != 0 {
		t.Errorf("tvOnly = %d, want 0", tvOnly)
	}
}

// TestMatchExact_ToleranceOverlapCollision_InputOrderInvariant verifies that the
// collision case resolves to matched=2 regardless of which runner appears first in
// the input slice, guarding against order-sensitive regression.
func TestMatchExact_ToleranceOverlapCollision_InputOrderInvariant(t *testing.T) {
	timeTol := 2 * time.Hour
	priceTol := 1.0

	rMulti := makeRunnerTradeWithDirection("2025-01-01 09:59", 101.0, "long")
	rSingle := makeRunnerTradeWithDirection("2025-01-01 10:00", 100.0, "long")
	tv := []TVTrade{
		makeTVTradeWithDirection("2025-01-01 10:00", 100.5, "long"),
		makeTVTradeWithDirection("2025-01-01 11:30", 101.5, "long"),
	}

	for _, tc := range []struct {
		name   string
		runner []RunnerTrade
	}{
		{"multi_first", []RunnerTrade{rMulti, rSingle}},
		{"single_first", []RunnerTrade{rSingle, rMulti}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			matched, _, _ := MatchExact(tc.runner, tv, timeTol, priceTol)
			if matched != 2 {
				t.Errorf("matched = %d, want 2", matched)
			}
		})
	}
}

// TestMatchExact_ThreeWayToleranceChain verifies that a multi-option runner
// acting as a bridge between two single-option runners is correctly re-routed to
// its escape slot so both single-option runners can claim their only compatible TV trade.
//
// Setup (timeTol=2h, priceTol=1.0):
//
//	R_multi  t=09:59  p=101.0 long → compat T1 (Δt=1min, Δp=1.0) and T2 (Δt=31min, Δp=0)
//	                                   NOT T3 (Δt=121min > 120min tolerance)
//	R_single1 t=10:00 p= 99.5 long → compat T1 only (Δp=0.5; T2 Δp=1.5, T3 Δp=2.5)
//	R_single3 t=12:00 p=102.5 long → compat T3 only (Δp=0.5; T1 Δp=2.5, T2 Δp=1.5)
//	T1 t=10:00 p=100.0   T2 t=10:30 p=101.0   T3 t=12:00 p=102.0
//
// Greedy (time-sorted, R_multi processed first) steals T1, stranding R_single1.
// Maximum matching routes R_multi to its escape T2, freeing T1 for R_single1.
func TestMatchExact_ThreeWayToleranceChain(t *testing.T) {
	timeTol := 2 * time.Hour
	priceTol := 1.0

	runner := []RunnerTrade{
		makeRunnerTradeWithDirection("2025-01-01 09:59", 101.0, "long"),
		makeRunnerTradeWithDirection("2025-01-01 10:00", 99.5, "long"),
		makeRunnerTradeWithDirection("2025-01-01 12:00", 102.5, "long"),
	}
	tv := []TVTrade{
		makeTVTradeWithDirection("2025-01-01 10:00", 100.0, "long"),
		makeTVTradeWithDirection("2025-01-01 10:30", 101.0, "long"),
		makeTVTradeWithDirection("2025-01-01 12:00", 102.0, "long"),
	}

	matched, runnerOnly, tvOnly := MatchExact(runner, tv, timeTol, priceTol)
	if matched != 3 {
		t.Errorf("matched = %d, want 3 (R_multi must escape to T2, freeing T1 for R_single1)", matched)
	}
	if runnerOnly != 0 {
		t.Errorf("runnerOnly = %d, want 0", runnerOnly)
	}
	if tvOnly != 0 {
		t.Errorf("tvOnly = %d, want 0", tvOnly)
	}
}

// TestMatchNetPnL_InputOrderInvariant verifies that MatchNetPnL returns the same
// result regardless of the order of runner and TV trades.
func TestMatchNetPnL_InputOrderInvariant(t *testing.T) {
	timeTol := time.Hour
	priceTol := 1.0
	relTol := 0.01

	runner := []RunnerTrade{
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", NetPnL: 30},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 10},
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "long", NetPnL: 20},
	}
	tv := []TVTrade{
		{EntryUTC: parseUTC("2025-01-02 10:00"), EntryPrice: 200, Direction: "long", NetPnL: 20},
		{EntryUTC: parseUTC("2025-01-03 10:00"), EntryPrice: 300, Direction: "long", NetPnL: 30},
		{EntryUTC: parseUTC("2025-01-01 10:00"), EntryPrice: 100, Direction: "long", NetPnL: 10},
	}

	wantMatched, wantMismatch := MatchNetPnL(runner, tv, 1.0, timeTol, priceTol, relTol)

	for i, j := 0, len(runner)-1; i < j; i, j = i+1, j-1 {
		runner[i], runner[j] = runner[j], runner[i]
	}
	for i, j := 0, len(tv)-1; i < j; i, j = i+1, j-1 {
		tv[i], tv[j] = tv[j], tv[i]
	}

	gotMatched, gotMismatch := MatchNetPnL(runner, tv, 1.0, timeTol, priceTol, relTol)
	if gotMatched != wantMatched || gotMismatch != wantMismatch {
		t.Errorf("reversed inputs: matched=%d pnlMismatch=%d, want %d/%d",
			gotMatched, gotMismatch, wantMatched, wantMismatch)
	}
}
