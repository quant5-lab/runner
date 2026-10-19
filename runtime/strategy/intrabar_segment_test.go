package strategy

import "testing"

func TestIntrabarSegments(t *testing.T) {
	tests := []struct {
		name     string
		barOpen  float64
		barHigh  float64
		barLow   float64
		wantSeg0 [2]float64
		wantSeg1 [2]float64
	}{
		{
			"open_near_high_path_high_before_low",
			108, 115, 90,
			[2]float64{108, 115}, [2]float64{115, 90},
		},
		{
			"open_near_low_path_low_before_high",
			93, 115, 90,
			[2]float64{93, 90}, [2]float64{90, 115},
		},
		{
			"equidistant_resolves_to_path_high_before_low",
			102.5, 115, 90,
			[2]float64{102.5, 115}, [2]float64{115, 90},
		},
		{
			"open_at_high_produces_zero_length_first_segment",
			115, 115, 90,
			[2]float64{115, 115}, [2]float64{115, 90},
		},
		{
			"open_at_low_produces_zero_length_first_segment",
			90, 115, 90,
			[2]float64{90, 90}, [2]float64{90, 115},
		},
		{
			"doji_produces_two_zero_length_segments",
			100, 100, 100,
			[2]float64{100, 100}, [2]float64{100, 100},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segs := intrabarSegments(tt.barOpen, tt.barHigh, tt.barLow)
			if segs[0].start != tt.wantSeg0[0] || segs[0].end != tt.wantSeg0[1] {
				t.Errorf("seg[0] = {%v→%v}, want {%v→%v}",
					segs[0].start, segs[0].end, tt.wantSeg0[0], tt.wantSeg0[1])
			}
			if segs[1].start != tt.wantSeg1[0] || segs[1].end != tt.wantSeg1[1] {
				t.Errorf("seg[1] = {%v→%v}, want {%v→%v}",
					segs[1].start, segs[1].end, tt.wantSeg1[0], tt.wantSeg1[1])
			}
		})
	}
}

func TestIntrabarSegment_Contains(t *testing.T) {
	tests := []struct {
		name  string
		start float64
		end   float64
		level float64
		want  bool
	}{
		{"asc_within", 100, 110, 105, true},
		{"asc_at_start", 100, 110, 100, true},
		{"asc_at_end", 100, 110, 110, true},
		{"asc_below_range", 100, 110, 99, false},
		{"asc_above_range", 100, 110, 111, false},
		{"desc_within", 110, 90, 100, true},
		{"desc_at_start", 110, 90, 110, true},
		{"desc_at_end", 110, 90, 90, true},
		{"desc_below_range", 110, 90, 89, false},
		{"desc_above_range", 110, 90, 111, false},
		{"zero_length_exact_match", 100, 100, 100, true},
		{"zero_length_no_match", 100, 100, 101, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg := intrabarSegment{tt.start, tt.end}
			if got := seg.contains(tt.level); got != tt.want {
				t.Errorf("contains(%v) = %v, want %v (seg {%v→%v})",
					tt.level, got, tt.want, tt.start, tt.end)
			}
		})
	}
}

func TestIntrabarSegment_ProximityToStart(t *testing.T) {
	tests := []struct {
		name  string
		start float64
		end   float64
		level float64
		want  float64
	}{
		{"asc_near_start", 100, 110, 102, 2},
		{"asc_far_from_start", 100, 110, 108, 8},
		{"asc_at_end", 100, 110, 110, 10},
		{"desc_near_start", 110, 90, 108, 2},
		{"desc_far_from_start", 110, 90, 92, 18},
		{"desc_at_end", 110, 90, 90, 20},
		{"at_start", 100, 110, 100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg := intrabarSegment{tt.start, tt.end}
			if got := seg.proximityToStart(tt.level); got != tt.want {
				t.Errorf("proximityToStart(%v) = %v, want %v (seg {%v→%v})",
					tt.level, got, tt.want, tt.start, tt.end)
			}
		})
	}
}
