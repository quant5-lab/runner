package strategy

import (
	"math"
	"testing"
)

func TestExitOrder_IsEligible(t *testing.T) {
	tests := []struct {
		name               string
		firstRegisteredBar int
		currentBar         int
		want               bool
	}{
		{"same_bar_not_eligible", 5, 5, false},
		{"next_bar_eligible", 5, 6, true},
		{"many_bars_later_eligible", 5, 100, true},
		{"bar_zero_same_not_eligible", 0, 0, false},
		{"bar_zero_next_eligible", 0, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExitOrder{FirstRegisteredBar: tt.firstRegisteredBar}
			if got := e.IsEligible(tt.currentBar); got != tt.want {
				t.Errorf("IsEligible(%d) = %v, want %v (FirstRegisteredBar=%d)",
					tt.currentBar, got, tt.want, tt.firstRegisteredBar)
			}
		})
	}
}

func TestCheckPriceTrigger_Long(t *testing.T) {
	long := Trade{Direction: Long}
	nan := math.NaN()

	tests := []struct {
		name      string
		stop      float64
		limit     float64
		barOpen   float64
		barHigh   float64
		barLow    float64
		wantTrig  bool
		wantPrice float64
		wantType  string
	}{
		// Single-level cases — barOpen irrelevant to outcome
		{"stop_hit_exact", 95, nan, 100, 100, 95, true, 95, "stop"},
		{"stop_hit_below", 95, nan, 100, 100, 90, true, 95, "stop"},
		{"stop_not_hit", 95, nan, 100, 100, 96, false, 0, ""},
		{"limit_hit_exact", nan, 110, 100, 110, 100, true, 110, "limit"},
		{"limit_hit_above", nan, 110, 100, 115, 100, true, 110, "limit"},
		{"limit_not_hit", nan, 110, 100, 109, 100, false, 0, ""},
		{"nan_stop_nan_limit", nan, nan, 100, 200, 50, false, 0, ""},
		// Both breached: conservative-fill rule — stop always wins regardless of
		// intrabar path heuristics. Matches TradingView broker emulator: without
		// sub-bar tick data, the worse-for-trader outcome fills. When the bar
		// opens beyond the stop level (gap-down for long), the fill happens at
		// the open price, not the stop level (TV gap-fill rule).
		{"both_breached_open_near_high_stop_wins", 95, 110, 112, 115, 90, true, 95, "stop"},
		{"both_breached_open_near_low_stop_wins", 95, 110, 93, 115, 90, true, 93, "stop"},
		{"both_breached_equidistant_stop_wins", 95, 110, 102.5, 115, 90, true, 95, "stop"},
		{"both_above_open_gap_fills_at_open", 108, 112, 105, 115, 90, true, 105, "stop"},
		{"both_above_open_limit_higher_gap_fills_at_open", 112, 108, 105, 115, 90, true, 105, "stop"},
		{"both_below_open_stop_wins", 92, 96, 100, 115, 85, true, 92, "stop"},
		{"both_below_open_limit_higher_stop_wins", 96, 92, 100, 115, 85, true, 96, "stop"},
		{"path_low_first_both_above_gap_fills_at_open", 108, 112, 87, 115, 85, true, 87, "stop"},
		{"path_low_first_both_below_stop_wins", 92, 88, 97, 115, 85, true, 92, "stop"},
		{"equal_levels_gap_fills_at_open", 100, 100, 87, 115, 85, true, 87, "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exit := ExitOrder{StopLevel: tt.stop, LimitLevel: tt.limit}
			trig, price, typ := checkPriceTrigger(exit, long, tt.barOpen, tt.barHigh, tt.barLow)
			if trig != tt.wantTrig {
				t.Errorf("triggered = %v, want %v", trig, tt.wantTrig)
			}
			if trig {
				if price != tt.wantPrice {
					t.Errorf("fillPrice = %v, want %v", price, tt.wantPrice)
				}
				if typ != tt.wantType {
					t.Errorf("fillType = %q, want %q", typ, tt.wantType)
				}
			}
		})
	}
}

func TestCheckPriceTrigger_Short(t *testing.T) {
	short := Trade{Direction: Short}
	nan := math.NaN()

	tests := []struct {
		name      string
		stop      float64
		limit     float64
		barOpen   float64
		barHigh   float64
		barLow    float64
		wantTrig  bool
		wantPrice float64
		wantType  string
	}{
		// Single-level cases — barOpen irrelevant to outcome
		{"stop_hit_exact", 105, nan, 100, 105, 100, true, 105, "stop"},
		{"stop_hit_above", 105, nan, 100, 110, 100, true, 105, "stop"},
		{"stop_not_hit", 105, nan, 100, 104, 100, false, 0, ""},
		{"limit_hit_exact", nan, 90, 100, 100, 90, true, 90, "limit"},
		{"limit_hit_below", nan, 90, 100, 100, 85, true, 90, "limit"},
		{"limit_not_hit", nan, 90, 100, 100, 91, false, 0, ""},
		// Both breached: conservative-fill rule — stop always wins regardless of
		// intrabar path. For shorts the stop is on the high side, limit on the
		// low side, but the precedence is identical: stop fills first. When the
		// bar opens beyond the stop level (gap-up for short), the fill happens
		// at the open price, not the stop level (TV gap-fill rule).
		{"both_breached_open_near_high_gap_fills_at_open", 105, 90, 108, 110, 85, true, 108, "stop"},
		{"both_breached_open_near_low_stop_wins", 105, 90, 87, 110, 85, true, 105, "stop"},
		{"both_above_open_stop_lower_stop_wins", 112, 108, 105, 115, 90, true, 112, "stop"},
		{"both_above_open_stop_wins", 108, 112, 105, 115, 90, true, 108, "stop"},
		{"both_below_open_stop_higher_gap_fills_at_open", 96, 92, 100, 115, 85, true, 100, "stop"},
		{"both_below_open_limit_higher_gap_fills_at_open", 92, 96, 100, 115, 85, true, 100, "stop"},
		{"path_low_first_both_above_stop_wins", 108, 112, 87, 115, 85, true, 108, "stop"},
		{"path_low_first_both_below_gap_fills_at_open", 92, 88, 97, 115, 85, true, 97, "stop"},
		{"equal_levels_stop_wins", 100, 100, 87, 115, 85, true, 100, "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exit := ExitOrder{StopLevel: tt.stop, LimitLevel: tt.limit}
			trig, price, typ := checkPriceTrigger(exit, short, tt.barOpen, tt.barHigh, tt.barLow)
			if trig != tt.wantTrig {
				t.Errorf("triggered = %v, want %v", trig, tt.wantTrig)
			}
			if trig {
				if price != tt.wantPrice {
					t.Errorf("fillPrice = %v, want %v", price, tt.wantPrice)
				}
				if typ != tt.wantType {
					t.Errorf("fillType = %q, want %q", typ, tt.wantType)
				}
			}
		})
	}
}

func TestPendingExitManager_CheckExitTriggered_EligibilityGate(t *testing.T) {
	pem := NewPendingExitManager()
	long := Trade{Direction: Long}

	const barOpen, barHigh, barLow = 100.0, 200.0, 50.0 // price always breaches stop=95 when eligible

	tests := []struct {
		name               string
		firstRegisteredBar int
		currentBar         int
		wantTriggered      bool
	}{
		{"registration_bar_blocks_trigger", 5, 5, false},
		{"next_bar_allows_trigger", 5, 6, true},
		{"many_bars_later_allows_trigger", 5, 100, true},
		{"bar_zero_same_blocks", 0, 0, false},
		{"bar_zero_next_allows", 0, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exit := ExitOrder{
				FirstRegisteredBar: tt.firstRegisteredBar,
				StopLevel:          95,
				LimitLevel:         math.NaN(),
			}
			triggered, _, _ := pem.CheckExitTriggered(exit, long, tt.currentBar, barOpen, barHigh, barLow)
			if triggered != tt.wantTriggered {
				t.Errorf("triggered = %v, want %v", triggered, tt.wantTriggered)
			}
		})
	}
}

func TestPendingExitManager_RegisterExit_FirstRegisteredBarPreservation(t *testing.T) {
	type reg struct {
		bar    int
		exitID string
		stop   float64
	}

	tests := []struct {
		name          string
		registrations []reg
		checkExitID   string
		wantBar       int
	}{
		{
			name:          "first_registration_uses_current_bar",
			registrations: []reg{{7, "e", 95}},
			checkExitID:   "e",
			wantBar:       7,
		},
		{
			name:          "level_update_preserves_first_bar",
			registrations: []reg{{3, "e", 95}, {5, "e", 90}},
			checkExitID:   "e",
			wantBar:       3,
		},
		{
			name:          "many_updates_preserve_first_bar",
			registrations: []reg{{3, "e", 95}, {4, "e", 90}, {5, "e", 85}, {10, "e", 80}},
			checkExitID:   "e",
			wantBar:       3,
		},
		{
			name:          "distinct_exit_ids_track_independently",
			registrations: []reg{{3, "e1", 95}, {5, "e2", 90}},
			checkExitID:   "e2",
			wantBar:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pem := NewPendingExitManager()
			for _, r := range tt.registrations {
				pem.RegisterExit(r.exitID, "entry", r.stop, math.NaN(), r.bar, "")
			}
			var found *ExitOrder
			for i := range pem.exitOrders {
				if pem.exitOrders[i].ExitID == tt.checkExitID {
					found = &pem.exitOrders[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("exit %q not found", tt.checkExitID)
			}
			if found.FirstRegisteredBar != tt.wantBar {
				t.Errorf("FirstRegisteredBar = %d, want %d", found.FirstRegisteredBar, tt.wantBar)
			}
		})
	}
}

// TestPendingExitManager_RegisterExit_AfterRemoval verifies that a re-registration
// after removal (e.g. a new trade opening with the same exit ID) starts fresh.
func TestPendingExitManager_RegisterExit_AfterRemoval(t *testing.T) {
	pem := NewPendingExitManager()
	pem.RegisterExit("e", "entry", 95, math.NaN(), 3, "")
	pem.RemoveExit("e", "entry")
	pem.RegisterExit("e", "entry", 90, math.NaN(), 8, "")

	var found *ExitOrder
	for i := range pem.exitOrders {
		if pem.exitOrders[i].ExitID == "e" {
			found = &pem.exitOrders[i]
			break
		}
	}
	if found == nil {
		t.Fatal("exit not found after re-registration")
	}
	if found.FirstRegisteredBar != 8 {
		t.Errorf("FirstRegisteredBar = %d, want 8", found.FirstRegisteredBar)
	}
}

func TestPendingExitManager_GetExitsForEntry(t *testing.T) {
	tests := []struct {
		name    string
		seeded  []ExitOrder
		query   string
		wantIDs []string
	}{
		{
			name:    "exact_from_entry_match",
			seeded:  []ExitOrder{{ExitID: "e1", FromEntry: "buy"}, {ExitID: "e2", FromEntry: "sell"}},
			query:   "buy",
			wantIDs: []string{"e1"},
		},
		{
			name:    "empty_from_entry_matches_any_query",
			seeded:  []ExitOrder{{ExitID: "e1", FromEntry: ""}, {ExitID: "e2", FromEntry: "sell"}},
			query:   "buy",
			wantIDs: []string{"e1"},
		},
		{
			name:    "mixed_exact_and_wildcard",
			seeded:  []ExitOrder{{ExitID: "e1", FromEntry: ""}, {ExitID: "e2", FromEntry: "buy"}},
			query:   "buy",
			wantIDs: []string{"e1", "e2"},
		},
		{
			name:    "no_match_returns_empty",
			seeded:  []ExitOrder{{ExitID: "e1", FromEntry: "sell"}},
			query:   "buy",
			wantIDs: []string{},
		},
		{
			name:    "empty_manager_returns_empty",
			seeded:  []ExitOrder{},
			query:   "buy",
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pem := &PendingExitManager{exitOrders: append([]ExitOrder{}, tt.seeded...)}
			got := pem.GetExitsForEntry(tt.query)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("len = %d, want %d; got IDs %v", len(got), len(tt.wantIDs), exitIDs(got))
			}
			for i, id := range tt.wantIDs {
				if got[i].ExitID != id {
					t.Errorf("[%d].ExitID = %q, want %q", i, got[i].ExitID, id)
				}
			}
		})
	}
}

func TestPendingExitManager_RemoveExit(t *testing.T) {
	tests := []struct {
		name      string
		seeded    []ExitOrder
		exitID    string
		fromEntry string
		wantIDs   []string
	}{
		{
			name:   "removes_matching_exit",
			seeded: []ExitOrder{{ExitID: "e1", FromEntry: "buy"}, {ExitID: "e2", FromEntry: "buy"}},
			exitID: "e1", fromEntry: "buy",
			wantIDs: []string{"e2"},
		},
		{
			name:   "noop_when_not_found",
			seeded: []ExitOrder{{ExitID: "e1", FromEntry: "buy"}},
			exitID: "e2", fromEntry: "buy",
			wantIDs: []string{"e1"},
		},
		{
			name:   "same_exit_id_different_entry_not_removed",
			seeded: []ExitOrder{{ExitID: "e1", FromEntry: "buy"}, {ExitID: "e1", FromEntry: "sell"}},
			exitID: "e1", fromEntry: "buy",
			wantIDs: []string{"e1"},
		},
		{
			name:   "empty_manager_is_noop",
			seeded: []ExitOrder{},
			exitID: "e1", fromEntry: "buy",
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pem := &PendingExitManager{exitOrders: append([]ExitOrder{}, tt.seeded...)}
			pem.RemoveExit(tt.exitID, tt.fromEntry)
			if len(pem.exitOrders) != len(tt.wantIDs) {
				t.Fatalf("len = %d, want %d; got IDs %v", len(pem.exitOrders), len(tt.wantIDs), exitIDs(pem.exitOrders))
			}
			for i, id := range tt.wantIDs {
				if pem.exitOrders[i].ExitID != id {
					t.Errorf("[%d].ExitID = %q, want %q", i, pem.exitOrders[i].ExitID, id)
				}
			}
		})
	}
}

func TestPendingExitManager_RemoveAllExitsForEntry(t *testing.T) {
	tests := []struct {
		name    string
		seeded  []ExitOrder
		entryID string
		wantIDs []string
	}{
		{
			name: "removes_all_for_entry",
			seeded: []ExitOrder{
				{ExitID: "e1", FromEntry: "buy"},
				{ExitID: "e2", FromEntry: "buy"},
				{ExitID: "e3", FromEntry: "sell"},
			},
			entryID: "buy",
			wantIDs: []string{"e3"},
		},
		{
			// Empty FromEntry matches any trade, so the exit is consumed on any close.
			name: "wildcard_empty_from_entry_removed_on_any_close",
			seeded: []ExitOrder{
				{ExitID: "e1", FromEntry: ""},
				{ExitID: "e2", FromEntry: "buy"},
			},
			entryID: "buy",
			wantIDs: []string{},
		},
		{
			name: "different_entry_id_is_preserved",
			seeded: []ExitOrder{
				{ExitID: "e1", FromEntry: "sell"},
				{ExitID: "e2", FromEntry: "buy"},
			},
			entryID: "buy",
			wantIDs: []string{"e1"},
		},
		{
			name: "no_match_is_noop",
			seeded: []ExitOrder{
				{ExitID: "e1", FromEntry: "sell"},
			},
			entryID: "buy",
			wantIDs: []string{"e1"},
		},
		{
			name:    "empty_manager_is_noop",
			seeded:  []ExitOrder{},
			entryID: "buy",
			wantIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pem := &PendingExitManager{exitOrders: append([]ExitOrder{}, tt.seeded...)}
			pem.RemoveAllExitsForEntry(tt.entryID)
			if len(pem.exitOrders) != len(tt.wantIDs) {
				t.Fatalf("len = %d, want %d; got IDs %v", len(pem.exitOrders), len(tt.wantIDs), exitIDs(pem.exitOrders))
			}
			for i, id := range tt.wantIDs {
				if pem.exitOrders[i].ExitID != id {
					t.Errorf("[%d].ExitID = %q, want %q", i, pem.exitOrders[i].ExitID, id)
				}
			}
		})
	}
}

func exitIDs(exits []ExitOrder) []string {
	ids := make([]string, len(exits))
	for i, e := range exits {
		ids[i] = e.ExitID
	}
	return ids
}

// TestStopFillPrice verifies the gap-fill rule for stop orders: when price opens
// beyond the stop level (long: open < stop; short: open > stop), the fill price
// is the bar open, not the stop level. When no gap exists, the fill price is the
// stop level. The boundary case (open exactly at stop) does not trigger gap-fill.
func TestStopFillPrice(t *testing.T) {
	tests := []struct {
		name      string
		isLong    bool
		stopLevel float64
		barOpen   float64
		want      float64
	}{
		// Long: adverse gap when open < stop (gap below stop level)
		{"long_gap_below_stop_fills_at_open", true, 95, 80, 80},
		{"long_open_exactly_at_stop_fills_at_stop", true, 95, 95, 95},
		{"long_open_above_stop_no_gap_fills_at_stop", true, 95, 100, 95},
		// Short: adverse gap when open > stop (gap above stop level)
		{"short_gap_above_stop_fills_at_open", false, 105, 120, 120},
		{"short_open_exactly_at_stop_fills_at_stop", false, 105, 105, 105},
		{"short_open_below_stop_no_gap_fills_at_stop", false, 105, 100, 105},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stopFillPrice(tt.stopLevel, tt.isLong, tt.barOpen)
			if got != tt.want {
				t.Errorf("stopFillPrice(level=%.2f, isLong=%v, open=%.2f) = %.2f, want %.2f",
					tt.stopLevel, tt.isLong, tt.barOpen, got, tt.want)
			}
		})
	}
}
