package ta

import (
	"math"
	"testing"
)

// TestNaNLevels_Length verifies the sentinel value has exactly PivotLevelsSize NaN entries.
func TestNaNLevels_Length(t *testing.T) {
	levels := NaNLevels()
	if len(levels) != PivotLevelsSize {
		t.Fatalf("NaNLevels length = %d, want %d", len(levels), PivotLevelsSize)
	}
	for i, v := range levels {
		if !math.IsNaN(v) {
			t.Errorf("NaNLevels[%d] = %v, want NaN", i, v)
		}
	}
}

// TestNaNLevelsReturnsIndependentSlices verifies each call produces a distinct
// slice so callers cannot corrupt each other's NaN arrays.
func TestNaNLevelsReturnsIndependentSlices(t *testing.T) {
	a := NaNLevels()
	b := NaNLevels()
	a[0] = 0.0
	if !math.IsNaN(b[0]) {
		t.Error("NaNLevels must return an independent slice on each call")
	}
}

// TestPivotLevelIndexConstants verifies the published index layout matches the
// documented order [P, R1, S1, R2, S2, R3, S3, R4, S4, R5, S5].
func TestPivotLevelIndexConstants(t *testing.T) {
	if PivotLevelsSize != 11 {
		t.Fatalf("PivotLevelsSize = %d, want 11", PivotLevelsSize)
	}
	want := map[string]int{
		"P": PivotP, "R1": PivotR1, "S1": PivotS1,
		"R2": PivotR2, "S2": PivotS2, "R3": PivotR3, "S3": PivotS3,
		"R4": PivotR4, "S4": PivotS4, "R5": PivotR5, "S5": PivotS5,
	}
	expected := map[string]int{
		"P": 0, "R1": 1, "S1": 2, "R2": 3, "S2": 4,
		"R3": 5, "S3": 6, "R4": 7, "S4": 8, "R5": 9, "S5": 10,
	}
	for name, idx := range want {
		if idx != expected[name] {
			t.Errorf("Pivot%s index = %d, want %d", name, idx, expected[name])
		}
	}
}

// TestComputePivotLevels_NaNInputReturnsNaNLevels verifies that an unavailable
// completed period (signalled by NaN prevH) returns all-NaN levels.
func TestComputePivotLevels_NaNInputReturnsNaNLevels(t *testing.T) {
	levels := ComputePivotLevels(PivotTraditional, math.NaN(), 90, 95, 91, 96)
	for i, v := range levels {
		if !math.IsNaN(v) {
			t.Errorf("NaN input: levels[%d] = %v, want NaN", i, v)
		}
	}
}

// TestComputePivotLevels_UnknownTypeReturnsNaNLevels verifies that an unrecognised
// pivot type returns all-NaN rather than panicking or returning garbage.
func TestComputePivotLevels_UnknownTypeReturnsNaNLevels(t *testing.T) {
	levels := ComputePivotLevels(PivotType("Unknown"), 100, 90, 95, 91, 96)
	for i, v := range levels {
		if !math.IsNaN(v) {
			t.Errorf("unknown type: levels[%d] = %v, want NaN", i, v)
		}
	}
}

// TestComputePivotLevels_LevelAvailability is a table-driven test that verifies
// which levels each pivot type defines (non-NaN) and which are unavailable (NaN).
// This is the authoritative source for "which levels exist per type".
func TestComputePivotLevels_LevelAvailability(t *testing.T) {
	tests := []struct {
		pivotType PivotType
		nanAt     []int
	}{
		{PivotTraditional, []int{}},
		{PivotFibonacci, []int{PivotR4, PivotS4, PivotR5, PivotS5}},
		{PivotWoodie, []int{PivotR5, PivotS5}},
		{PivotClassic, []int{PivotR5, PivotS5}},
		{PivotDM, []int{PivotR2, PivotS2, PivotR3, PivotS3, PivotR4, PivotS4, PivotR5, PivotS5}},
		{PivotCamarilla, []int{}},
	}

	nanSet := func(indices []int) map[int]bool {
		m := make(map[int]bool, len(indices))
		for _, i := range indices {
			m[i] = true
		}
		return m
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.pivotType), func(t *testing.T) {
			levels := ComputePivotLevels(tt.pivotType, 110, 90, 100, 100, 102)
			nans := nanSet(tt.nanAt)
			for i, v := range levels {
				if nans[i] {
					if !math.IsNaN(v) {
						t.Errorf("%s levels[%d] = %v, want NaN", tt.pivotType, i, v)
					}
				} else {
					if math.IsNaN(v) {
						t.Errorf("%s levels[%d] = NaN, want defined value", tt.pivotType, i)
					}
				}
			}
		})
	}
}

// TestComputePivotLevels_Traditional verifies all 11 levels using the standard
// H=110, L=90, C=100 fixture (P=100, hl=20).
func TestComputePivotLevels_Traditional(t *testing.T) {
	levels := ComputePivotLevels(PivotTraditional, 110, 90, 100, 0, 0)
	checks := []struct {
		name string
		idx  int
		want float64
	}{
		{"P", PivotP, 100.0},
		{"R1", PivotR1, 110.0},
		{"S1", PivotS1, 90.0},
		{"R2", PivotR2, 120.0},
		{"S2", PivotS2, 80.0},
		{"R3", PivotR3, 130.0},
		{"S3", PivotS3, 70.0},
		{"R4", PivotR4, 140.0},
		{"S4", PivotS4, 60.0},
		{"R5", PivotR5, 150.0},
		{"S5", PivotS5, 50.0},
	}
	for _, c := range checks {
		if !approx(levels[c.idx], c.want) {
			t.Errorf("Traditional %s = %.6f, want %.6f", c.name, levels[c.idx], c.want)
		}
	}
}

// TestComputePivotLevels_Fibonacci verifies all defined levels using the
// Fibonacci retracement ratios (0.382, 0.618, 1.0).
func TestComputePivotLevels_Fibonacci(t *testing.T) {
	levels := ComputePivotLevels(PivotFibonacci, 110, 90, 100, 0, 0)
	checks := []struct {
		name string
		idx  int
		want float64
	}{
		{"P", PivotP, 100.0},
		{"R1", PivotR1, 100 + 0.382*20},
		{"S1", PivotS1, 100 - 0.382*20},
		{"R2", PivotR2, 100 + 0.618*20},
		{"S2", PivotS2, 100 - 0.618*20},
		{"R3", PivotR3, 120.0},
		{"S3", PivotS3, 80.0},
	}
	for _, c := range checks {
		if !approx(levels[c.idx], c.want) {
			t.Errorf("Fibonacci %s = %.6f, want %.6f", c.name, levels[c.idx], c.want)
		}
	}
}

// TestComputePivotLevels_Woodie verifies all defined levels. Woodie uses the
// new period's open (woodieCurrentOpen) instead of prevClose for the pivot.
func TestComputePivotLevels_Woodie(t *testing.T) {
	levels := ComputePivotLevels(PivotWoodie, 110, 90, 0, 0, 102)
	checks := []struct {
		name string
		idx  int
		want float64
	}{
		{"P", PivotP, 101.0},
		{"R1", PivotR1, 112.0},
		{"S1", PivotS1, 92.0},
		{"R2", PivotR2, 121.0},
		{"S2", PivotS2, 81.0},
		{"R3", PivotR3, 132.0},
		{"S3", PivotS3, 72.0},
		{"R4", PivotR4, 152.0},
		{"S4", PivotS4, 52.0},
	}
	for _, c := range checks {
		if !approx(levels[c.idx], c.want) {
			t.Errorf("Woodie %s = %.6f, want %.6f", c.name, levels[c.idx], c.want)
		}
	}
}

// TestComputePivotLevels_Classic verifies all defined levels.
// Classic differs from Traditional only in R3–R4 (multiples of hl from P,
// not from H/L).
func TestComputePivotLevels_Classic(t *testing.T) {
	levels := ComputePivotLevels(PivotClassic, 110, 90, 100, 0, 0)
	checks := []struct {
		name string
		idx  int
		want float64
	}{
		{"P", PivotP, 100.0},
		{"R1", PivotR1, 110.0},
		{"S1", PivotS1, 90.0},
		{"R2", PivotR2, 120.0},
		{"S2", PivotS2, 80.0},
		{"R3", PivotR3, 140.0},
		{"S3", PivotS3, 60.0},
		{"R4", PivotR4, 160.0},
		{"S4", PivotS4, 40.0},
	}
	for _, c := range checks {
		if !approx(levels[c.idx], c.want) {
			t.Errorf("Classic %s = %.6f, want %.6f", c.name, levels[c.idx], c.want)
		}
	}
}

// TestComputePivotLevels_DM_OpenEqualsClose verifies the DM formula when
// prevOpen == prevClose, where X = H + L + 2*C.
func TestComputePivotLevels_DM_OpenEqualsClose(t *testing.T) {
	levels := ComputePivotLevels(PivotDM, 110, 90, 100, 100, 0)
	x := 110.0 + 90.0 + 2*100.0
	if !approx(levels[PivotP], x/4) {
		t.Errorf("DM P (O==C) = %.4f, want %.4f", levels[PivotP], x/4)
	}
	if !approx(levels[PivotR1], x/2-90) {
		t.Errorf("DM R1 (O==C) = %.4f, want %.4f", levels[PivotR1], x/2-90)
	}
	if !approx(levels[PivotS1], x/2-110) {
		t.Errorf("DM S1 (O==C) = %.4f, want %.4f", levels[PivotS1], x/2-110)
	}
}

// TestComputePivotLevels_DM_CloseAboveOpen verifies the DM formula when
// prevClose > prevOpen, where X = 2*H + L + C.
func TestComputePivotLevels_DM_CloseAboveOpen(t *testing.T) {
	levels := ComputePivotLevels(PivotDM, 110, 90, 102, 98, 0)
	x := 2*110.0 + 90.0 + 102.0
	if !approx(levels[PivotP], x/4) {
		t.Errorf("DM P (C>O) = %.4f, want %.4f", levels[PivotP], x/4)
	}
	if !approx(levels[PivotR1], x/2-90) {
		t.Errorf("DM R1 (C>O) = %.4f, want %.4f", levels[PivotR1], x/2-90)
	}
	if !approx(levels[PivotS1], x/2-110) {
		t.Errorf("DM S1 (C>O) = %.4f, want %.4f", levels[PivotS1], x/2-110)
	}
}

// TestComputePivotLevels_DM_CloseBelowOpen verifies the DM formula when
// prevClose < prevOpen, where X = 2*L + H + C.
func TestComputePivotLevels_DM_CloseBelowOpen(t *testing.T) {
	levels := ComputePivotLevels(PivotDM, 110, 90, 96, 100, 0)
	x := 2*90.0 + 110.0 + 96.0
	if !approx(levels[PivotP], x/4) {
		t.Errorf("DM P (C<O) = %.4f, want %.4f", levels[PivotP], x/4)
	}
	if !approx(levels[PivotR1], x/2-90) {
		t.Errorf("DM R1 (C<O) = %.4f, want %.4f", levels[PivotR1], x/2-90)
	}
	if !approx(levels[PivotS1], x/2-110) {
		t.Errorf("DM S1 (C<O) = %.4f, want %.4f", levels[PivotS1], x/2-110)
	}
}

// TestComputePivotLevels_Camarilla verifies all 11 Camarilla levels including
// the ratio-based R5/S5 formula.
func TestComputePivotLevels_Camarilla(t *testing.T) {
	levels := ComputePivotLevels(PivotCamarilla, 110, 90, 100, 0, 0)
	hl := 110.0 - 90.0
	r5 := (110.0 / 90.0) * 100.0

	checks := []struct {
		name string
		idx  int
		want float64
	}{
		{"P", PivotP, (110 + 90 + 100) / 3.0},
		{"R1", PivotR1, 100 + 1.1*hl/12},
		{"S1", PivotS1, 100 - 1.1*hl/12},
		{"R2", PivotR2, 100 + 1.1*hl/6},
		{"S2", PivotS2, 100 - 1.1*hl/6},
		{"R3", PivotR3, 100 + 1.1*hl/4},
		{"S3", PivotS3, 100 - 1.1*hl/4},
		{"R4", PivotR4, 100 + 1.1*hl/2},
		{"S4", PivotS4, 100 - 1.1*hl/2},
		{"R5", PivotR5, r5},
		{"S5", PivotS5, 100 - (r5 - 100)},
	}
	for _, c := range checks {
		if !approx(levels[c.idx], c.want) {
			t.Errorf("Camarilla %s = %.6f, want %.6f", c.name, levels[c.idx], c.want)
		}
	}
}

func approx(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
