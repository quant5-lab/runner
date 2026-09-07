package request

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/context"
)

func TestSecurityBarMapper_BuildMappingByTimestamp_EmptyInputs(t *testing.T) {
	primary := []context.OHLCV{{Time: parseTime("2025-01-01 10:00:00"), Close: 100}}
	coarser := []context.OHLCV{{Time: parseTime("2025-01-01 10:00:00"), Close: 100}}

	cases := []struct {
		name    string
		coarser []context.OHLCV
		primary []context.OHLCV
	}{
		{"nil coarser", nil, primary},
		{"nil primary", coarser, nil},
		{"both nil", nil, nil},
		{"empty coarser", []context.OHLCV{}, primary},
		{"empty primary", coarser, []context.OHLCV{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewSecurityBarMapper()
			priorMode := m.Mode()
			m.BuildMappingByTimestamp(tc.coarser, tc.primary)
			if m.Mode() != priorMode {
				t.Errorf("mode changed from %v to %v; empty inputs must be no-op", priorMode, m.Mode())
			}
			if len(m.GetRanges()) != 0 {
				t.Errorf("got %d ranges, want 0 for empty input", len(m.GetRanges()))
			}
		})
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_SingleSlot(t *testing.T) {
	tests := []struct {
		name         string
		primaryBars  []context.OHLCV
		wantStartIdx int
		wantEndIdx   int
	}{
		{
			name: "single primary in slot",
			primaryBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 09:30:00"), Close: 100},
			},
			wantStartIdx: 0, wantEndIdx: 0,
		},
		{
			name: "four primaries in single slot",
			primaryBars: []context.OHLCV{
				{Time: parseTime("2025-01-01 09:30:00"), Close: 100},
				{Time: parseTime("2025-01-01 10:30:00"), Close: 101},
				{Time: parseTime("2025-01-01 11:30:00"), Close: 102},
				{Time: parseTime("2025-01-01 12:30:00"), Close: 103},
			},
			wantStartIdx: 0, wantEndIdx: 3,
		},
	}

	coarser := []context.OHLCV{{Time: parseTime("2025-01-01 09:30:00"), Close: 100}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSecurityBarMapper()
			m.BuildMappingByTimestamp(coarser, tt.primaryBars)

			ranges := m.GetRanges()
			if len(ranges) != 1 {
				t.Fatalf("got %d ranges, want 1", len(ranges))
			}
			r := ranges[0]
			if r.DailyBarIndex != 0 {
				t.Errorf("DailyBarIndex = %d, want 0", r.DailyBarIndex)
			}
			if r.StartHourlyIndex != tt.wantStartIdx || r.EndHourlyIndex != tt.wantEndIdx {
				t.Errorf("primary span [%d,%d], want [%d,%d]",
					r.StartHourlyIndex, r.EndHourlyIndex, tt.wantStartIdx, tt.wantEndIdx)
			}
		})
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_MultipleSlots(t *testing.T) {
	// Two intraday coarser bars on the same calendar date.
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // slot 0: [09:30–11:30)
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // slot 1: [11:30–…)
	}
	primary := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // idx 0 → slot 0
		{Time: parseTime("2025-01-01 10:30:00"), Close: 101}, // idx 1 → slot 0
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // idx 2 → slot 1
		{Time: parseTime("2025-01-01 12:30:00"), Close: 103}, // idx 3 → slot 1
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	ranges := m.GetRanges()
	if len(ranges) != 2 {
		t.Fatalf("got %d ranges, want 2", len(ranges))
	}

	if ranges[0].DailyBarIndex != 0 || ranges[0].StartHourlyIndex != 0 || ranges[0].EndHourlyIndex != 1 {
		t.Errorf("slot 0: got daily=%d start=%d end=%d, want 0 0 1",
			ranges[0].DailyBarIndex, ranges[0].StartHourlyIndex, ranges[0].EndHourlyIndex)
	}
	if ranges[1].DailyBarIndex != 1 || ranges[1].StartHourlyIndex != 2 || ranges[1].EndHourlyIndex != 3 {
		t.Errorf("slot 1: got daily=%d start=%d end=%d, want 1 2 3",
			ranges[1].DailyBarIndex, ranges[1].StartHourlyIndex, ranges[1].EndHourlyIndex)
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_LeadingPrimariesSkipped(t *testing.T) {
	// Primary bars 0..1 predate the coarser bar; bars 2..3 belong to it.
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102},
	}
	primary := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // pre-coarser, skip
		{Time: parseTime("2025-01-01 10:30:00"), Close: 101}, // pre-coarser, skip
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // idx 2 → slot 0
		{Time: parseTime("2025-01-01 12:30:00"), Close: 103}, // idx 3 → slot 0
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	ranges := m.GetRanges()
	if len(ranges) != 1 {
		t.Fatalf("got %d ranges, want 1", len(ranges))
	}
	if ranges[0].StartHourlyIndex != 2 || ranges[0].EndHourlyIndex != 3 {
		t.Errorf("primary span [%d,%d], want [2,3]", ranges[0].StartHourlyIndex, ranges[0].EndHourlyIndex)
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_LastSlotAbsorbsRemaining(t *testing.T) {
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // slot 0
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // slot 1 (last)
	}
	primary := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // idx 0 → slot 0
		{Time: parseTime("2025-01-01 10:30:00"), Close: 101}, // idx 1 → slot 0
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // idx 2 → slot 0
		{Time: parseTime("2025-01-01 12:30:00"), Close: 103}, // idx 3 → slot 0
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // idx 4 → slot 1
		{Time: parseTime("2025-01-01 14:30:00"), Close: 105}, // idx 5 → slot 1
		{Time: parseTime("2025-01-01 15:30:00"), Close: 106}, // idx 6 → slot 1
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	ranges := m.GetRanges()
	if len(ranges) != 2 {
		t.Fatalf("got %d ranges, want 2", len(ranges))
	}
	if ranges[0].EndHourlyIndex != 3 {
		t.Errorf("slot 0 end = %d, want 3", ranges[0].EndHourlyIndex)
	}
	if ranges[1].StartHourlyIndex != 4 || ranges[1].EndHourlyIndex != 6 {
		t.Errorf("slot 1: start=%d end=%d, want 4 6", ranges[1].StartHourlyIndex, ranges[1].EndHourlyIndex)
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_CoarserSlotWithNoPrimaries(t *testing.T) {
	// Primary bars skip the time window of coarser slot 1.
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // slot 0: has primaries
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // slot 1: no primaries (gap)
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // slot 2: has primaries
	}
	primary := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // → slot 0
		{Time: parseTime("2025-01-01 10:30:00"), Close: 101}, // → slot 0
		// slot 1 window [11:30–13:30): no primary bars
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // → slot 2
		{Time: parseTime("2025-01-01 14:30:00"), Close: 105}, // → slot 2
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	ranges := m.GetRanges()
	if len(ranges) != 2 {
		t.Fatalf("got %d ranges, want 2 (slot 1 has no primaries so no range)", len(ranges))
	}
	if ranges[0].DailyBarIndex != 0 {
		t.Errorf("first range DailyBarIndex = %d, want 0", ranges[0].DailyBarIndex)
	}
	if ranges[1].DailyBarIndex != 2 {
		t.Errorf("second range DailyBarIndex = %d, want 2", ranges[1].DailyBarIndex)
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_FindDailyBarIndex(t *testing.T) {
	// Three 2h coarser slots, each with two 1h primary bars.
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // coarser 0 → primaries 0,1
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // coarser 1 → primaries 2,3
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // coarser 2 → primaries 4,5
	}
	primary := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00"), Close: 100}, // 0 → coarser 0
		{Time: parseTime("2025-01-01 10:30:00"), Close: 101}, // 1 → coarser 0
		{Time: parseTime("2025-01-01 11:30:00"), Close: 102}, // 2 → coarser 1
		{Time: parseTime("2025-01-01 12:30:00"), Close: 103}, // 3 → coarser 1
		{Time: parseTime("2025-01-01 13:30:00"), Close: 104}, // 4 → coarser 2
		{Time: parseTime("2025-01-01 14:30:00"), Close: 105}, // 5 → coarser 2
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	cases := []struct {
		primaryIdx int
		lookahead  bool
		want       int
		desc       string
	}{
		// First range — no previous coarser bar, so no-lookahead still returns coarser 0.
		{0, true, 0, "first primary of slot 0, lookahead=true → coarser 0"},
		{0, false, 0, "first primary of slot 0, lookahead=false → coarser 0 (no prior)"},
		{1, true, 0, "last primary of slot 0, lookahead=true → coarser 0"},
		{1, false, 0, "last primary of slot 0, lookahead=false → coarser 0"},
		// Second range: slot boundary between primary 1 and 2.
		{2, true, 1, "first primary of slot 1, lookahead=true → coarser 1"},
		{2, false, 0, "first primary of slot 1, lookahead=false → previous coarser 0"},
		{3, true, 1, "last primary of slot 1, lookahead=true → coarser 1"},
		{3, false, 0, "last primary of slot 1, lookahead=false → previous coarser 0"},
		// Third range.
		{4, true, 2, "first primary of slot 2, lookahead=true → coarser 2"},
		{4, false, 1, "first primary of slot 2, lookahead=false → previous coarser 1"},
		{5, true, 2, "last primary, lookahead=true → coarser 2"},
		{5, false, 1, "last primary, lookahead=false → previous coarser 1"},
		// Out of bounds.
		{-1, true, -1, "negative index → -1"},
		{-1, false, -1, "negative index → -1"},
		{6, true, 2, "beyond last primary, lookahead=true → last coarser"},
		{6, false, 2, "beyond last primary, lookahead=false → last coarser"},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := m.FindDailyBarIndex(tc.primaryIdx, tc.lookahead)
			if got != tc.want {
				t.Errorf("FindDailyBarIndex(%d, %v) = %d, want %d", tc.primaryIdx, tc.lookahead, got, tc.want)
			}
		})
	}
}

func TestSecurityBarMapper_BuildMappingByTimestamp_RangeIntegrity(t *testing.T) {
	coarser := []context.OHLCV{
		{Time: parseTime("2025-01-01 09:30:00")},
		{Time: parseTime("2025-01-01 11:30:00")},
		{Time: parseTime("2025-01-01 13:30:00")},
		{Time: parseTime("2025-01-01 15:30:00")},
	}
	primary := make([]context.OHLCV, 8)
	base := int64(1735731000) // 2025-01-01 09:30:00 UTC
	for i := range primary {
		primary[i] = context.OHLCV{Time: base + int64(i)*3600}
	}

	m := NewSecurityBarMapper()
	m.BuildMappingByTimestamp(coarser, primary)

	ranges := m.GetRanges()
	if len(ranges) == 0 {
		t.Fatal("no ranges produced")
	}

	for i := 1; i < len(ranges); i++ {
		prev := ranges[i-1]
		curr := ranges[i]
		if curr.DailyBarIndex <= prev.DailyBarIndex {
			t.Errorf("ranges[%d].DailyBarIndex=%d not > ranges[%d].DailyBarIndex=%d",
				i, curr.DailyBarIndex, i-1, prev.DailyBarIndex)
		}
		if curr.StartHourlyIndex <= prev.EndHourlyIndex {
			t.Errorf("ranges[%d].StartHourlyIndex=%d overlaps ranges[%d].EndHourlyIndex=%d",
				i, curr.StartHourlyIndex, i-1, prev.EndHourlyIndex)
		}
		if curr.StartHourlyIndex != prev.EndHourlyIndex+1 {
			t.Errorf("gap between ranges[%d] end=%d and ranges[%d] start=%d — should be contiguous",
				i-1, prev.EndHourlyIndex, i, curr.StartHourlyIndex)
		}
	}
}
