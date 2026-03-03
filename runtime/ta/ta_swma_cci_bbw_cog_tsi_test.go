package ta

import (
	"math"
	"testing"
)

func TestSwma(t *testing.T) {
	source := []float64{10, 20, 30, 40, 50, 60}
	result := Swma(source)

	if len(result) != len(source) {
		t.Fatalf("Swma length = %d, want %d", len(result), len(source))
	}

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Swma[%d] should be NaN (warmup), got %f", i, result[i])
		}
	}

	/* index 3: 10*(1/6) + 20*(2/6) + 30*(2/6) + 40*(1/6) = 25.0 */
	expected3 := 10.0*(1.0/6.0) + 20.0*(2.0/6.0) + 30.0*(2.0/6.0) + 40.0*(1.0/6.0)
	if math.Abs(result[3]-expected3) > 0.0001 {
		t.Errorf("Swma[3] = %f, want %f", result[3], expected3)
	}

	/* index 5: 30*(1/6) + 40*(2/6) + 50*(2/6) + 60*(1/6) = 45.0 */
	expected5 := 30.0*(1.0/6.0) + 40.0*(2.0/6.0) + 50.0*(2.0/6.0) + 60.0*(1.0/6.0)
	if math.Abs(result[5]-expected5) > 0.0001 {
		t.Errorf("Swma[5] = %f, want %f", result[5], expected5)
	}
}

func TestSwma_EdgeCases(t *testing.T) {
	t.Run("short_slice_all_warmup", func(t *testing.T) {
		for n := 0; n <= 3; n++ {
			src := make([]float64, n)
			for i := range src {
				src[i] = float64(i + 1)
			}
			result := Swma(src)
			for i, v := range result {
				if !math.IsNaN(v) {
					t.Errorf("len=%d: Swma[%d] = %f, want NaN", n, i, v)
				}
			}
		}
	})

	t.Run("constant_source_equals_constant", func(t *testing.T) {
		src := []float64{7, 7, 7, 7, 7}
		result := Swma(src)
		for i := 3; i < len(result); i++ {
			if math.Abs(result[i]-7.0) > 0.0001 {
				t.Errorf("Swma[%d] of constant series = %f, want 7.0", i, result[i])
			}
		}
	})

	t.Run("all_nan_source", func(t *testing.T) {
		src := []float64{math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()}
		result := Swma(src)
		for i := 3; i < len(result); i++ {
			if !math.IsNaN(result[i]) {
				t.Errorf("Swma[%d] with NaN source should be NaN, got %f", i, result[i])
			}
		}
	})

	t.Run("weights_sum_to_one", func(t *testing.T) {
		/* SWMA of constant c must equal c */
		src := []float64{100, 100, 100, 100}
		result := Swma(src)
		if math.Abs(result[3]-100.0) > 0.0001 {
			t.Errorf("Swma weights don't sum to 1: got %f, want 100.0", result[3])
		}
	})
}

func TestCci(t *testing.T) {
	source := []float64{10, 12, 11, 13, 14, 15, 13, 12}
	result := Cci(source, 5)

	if len(result) != len(source) {
		t.Fatalf("Cci length = %d, want %d", len(result), len(source))
	}

	for i := 0; i < 4; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Cci[%d] should be NaN (warmup), got %f", i, result[i])
		}
	}

	if math.IsNaN(result[4]) {
		t.Error("Cci[4] should have a valid value at first valid bar")
	}
}

func TestCci_FlatSource(t *testing.T) {
	/* Zero deviation → CCI = 0 */
	src := []float64{10, 10, 10, 10, 10, 10}
	result := Cci(src, 5)

	for i := 4; i < len(result); i++ {
		if math.Abs(result[i]-0.0) > 0.0001 {
			t.Errorf("Cci[%d] with flat source = %f, want 0.0", i, result[i])
		}
	}
}

func TestCci_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Cci([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("Cci of empty slice: length = %d, want 0", len(result))
		}
	})

	t.Run("zero_length_param", func(t *testing.T) {
		src := []float64{1, 2, 3}
		result := Cci(src, 0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Cci with length=0: [%d] = %f, want NaN", i, v)
			}
		}
	})

	t.Run("single_bar_always_warmup", func(t *testing.T) {
		src := []float64{50}
		result := Cci(src, 5)
		if !math.IsNaN(result[0]) {
			t.Errorf("Cci of single bar with period=5 should be NaN, got %f", result[0])
		}
	})

	t.Run("above_sma_positive_cci", func(t *testing.T) {
		/* Linearly increasing series: last bar above SMA → positive CCI */
		src := []float64{10, 11, 12, 13, 100}
		result := Cci(src, 5)
		if result[4] <= 0 {
			t.Errorf("Expected positive CCI when current bar >> SMA, got %f", result[4])
		}
	})

	t.Run("below_sma_negative_cci", func(t *testing.T) {
		/* Last bar well below SMA → negative CCI */
		src := []float64{100, 100, 100, 100, 1}
		result := Cci(src, 5)
		if result[4] >= 0 {
			t.Errorf("Expected negative CCI when current bar << SMA, got %f", result[4])
		}
	})
}

func TestBbw(t *testing.T) {
	source := []float64{10, 11, 12, 13, 14, 15, 14, 13}
	result := Bbw(source, 5, 2.0)

	if len(result) != len(source) {
		t.Fatalf("Bbw length = %d, want %d", len(result), len(source))
	}

	for i := 0; i < 4; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Bbw[%d] should be NaN (warmup), got %f", i, result[i])
		}
	}

	for i := 4; i < len(result); i++ {
		if math.IsNaN(result[i]) {
			t.Errorf("Bbw[%d] should not be NaN at valid bar", i)
		}
		if result[i] < 0 {
			t.Errorf("Bbw[%d] = %f should be non-negative", i, result[i])
		}
	}
}

func TestBbw_MultiplierScaling(t *testing.T) {
	src := []float64{10, 12, 8, 15, 11, 9, 13}
	r1 := Bbw(src, 5, 1.0)
	r2 := Bbw(src, 5, 2.0)

	for i := 4; i < len(src); i++ {
		if math.IsNaN(r1[i]) || math.IsNaN(r2[i]) {
			continue
		}
		if math.Abs(r2[i]-2.0*r1[i]) > 0.0001 {
			t.Errorf("Bbw[%d]: mult=2 result %.6f != 2*mult=1 result %.6f", i, r2[i], r1[i])
		}
	}
}

func TestBbw_EdgeCases(t *testing.T) {
	t.Run("flat_source_zero_stdev", func(t *testing.T) {
		/* Zero stdev → BBW = 0 */
		src := []float64{10, 10, 10, 10, 10}
		result := Bbw(src, 5, 2.0)
		if math.Abs(result[4]-0.0) > 0.0001 {
			t.Errorf("Bbw with flat source = %f, want 0.0", result[4])
		}
	})

	t.Run("empty_source", func(t *testing.T) {
		result := Bbw([]float64{}, 5, 2.0)
		if len(result) != 0 {
			t.Errorf("Bbw of empty slice: length = %d, want 0", len(result))
		}
	})

	t.Run("zero_length_param", func(t *testing.T) {
		src := []float64{1, 2, 3}
		result := Bbw(src, 0, 2.0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Bbw with length=0: [%d] = %f, want NaN", i, v)
			}
		}
	})
}

func TestCog(t *testing.T) {
	source := []float64{10, 20, 30, 40, 50}
	result := Cog(source, 3)

	if len(result) != len(source) {
		t.Fatalf("Cog length = %d, want %d", len(result), len(source))
	}

	for i := 0; i < 2; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Cog[%d] should be NaN (warmup), got %f", i, result[i])
		}
	}

	/* index 2: window=[10,20,30]  most-recent=30 (j=0), middle=20 (j=1), oldest=10 (j=2)
	 * num = 30*1 + 20*2 + 10*3 = 100; den = 60; COG = -100/60 = -1.6667 */
	expected2 := -(30.0*1 + 20.0*2 + 10.0*3) / (30.0 + 20.0 + 10.0)
	if math.Abs(result[2]-expected2) > 0.0001 {
		t.Errorf("Cog[2] = %f, want %f", result[2], expected2)
	}
}

func TestCog_UniformSource(t *testing.T) {
	/* Uniform weights: COG = -(1+2+...+n)/n = -(n+1)/2 */
	src := []float64{5, 5, 5, 5, 5}
	result := Cog(src, 3)

	/* -(5*1+5*2+5*3)/(5*3) = -30/15 = -2.0 */
	expected := -(5.0*1 + 5.0*2 + 5.0*3) / (5.0 * 3)
	for i := 2; i < len(result); i++ {
		if math.Abs(result[i]-expected) > 0.0001 {
			t.Errorf("Cog[%d] uniform source = %f, want %f", i, result[i], expected)
		}
	}
}

func TestCog_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Cog([]float64{}, 3)
		if len(result) != 0 {
			t.Errorf("Cog of empty slice: length = %d, want 0", len(result))
		}
	})

	t.Run("zero_length_param", func(t *testing.T) {
		src := []float64{1, 2, 3}
		result := Cog(src, 0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Cog with length=0: [%d] = %f, want NaN", i, v)
			}
		}
	})

	t.Run("single_bar_is_negative_one", func(t *testing.T) {
		/* period=1: num = v*1, den = v → COG = -1 */
		src := []float64{42, 42, 42}
		result := Cog(src, 1)
		for i, v := range result {
			if math.Abs(v-(-1.0)) > 0.0001 {
				t.Errorf("Cog[%d] with period=1 = %f, want -1.0", i, v)
			}
		}
	})

	t.Run("zero_denominator_returns_zero", func(t *testing.T) {
		/* All-zero source: den = 0 → COG = 0 */
		src := []float64{0, 0, 0}
		result := Cog(src, 3)
		if math.Abs(result[2]-0.0) > 0.0001 {
			t.Errorf("Cog with zero-sum source = %f, want 0.0", result[2])
		}
	})

	t.Run("cog_is_negative", func(t *testing.T) {
		/* COG is always negative for positive sources */
		src := []float64{1, 2, 3, 4, 5}
		result := Cog(src, 3)
		for i := 2; i < len(result); i++ {
			if result[i] > 0 {
				t.Errorf("Cog[%d] = %f should be non-positive for positive source", i, result[i])
			}
		}
	})
}

func TestTsi(t *testing.T) {
	source := make([]float64, 40)
	for i := range source {
		source[i] = float64(100 + i)
	}

	result := Tsi(source, 5, 13)

	if len(result) != len(source) {
		t.Fatalf("Tsi length = %d, want %d", len(result), len(source))
	}

	/* warmup = longLength + shortLength - 1: EMA(momentum,long) seeds at bar long, EMA(ema1,short) seeds at bar long+short-1 */
	warmupEdge := 5 + 13 - 1
	for i := 0; i < warmupEdge; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Tsi[%d] should be NaN during warmup, got %f", i, result[i])
		}
	}

	for i := warmupEdge; i < len(result); i++ {
		if math.IsNaN(result[i]) {
			t.Errorf("Tsi[%d] should have a valid value after warmup", i)
		}
	}
}

func TestTsi_Bounds(t *testing.T) {
	/* TSI is bounded [-100, 100] */
	source := make([]float64, 60)
	for i := range source {
		if i%2 == 0 {
			source[i] = 100 + float64(i)
		} else {
			source[i] = 100 - float64(i)
		}
	}

	result := Tsi(source, 5, 13)

	for i, v := range result {
		if math.IsNaN(v) {
			continue
		}
		if v < -100.0 || v > 100.0 {
			t.Errorf("Tsi[%d] = %f, out of [-100, 100] bounds", i, v)
		}
	}
}

func TestTsi_FlatSource(t *testing.T) {
	/* Zero momentum everywhere → TSI = 0 */
	src := make([]float64, 40)
	for i := range src {
		src[i] = 50.0
	}
	result := Tsi(src, 5, 13)

	for i, v := range result {
		if math.IsNaN(v) {
			continue
		}
		if math.Abs(v-0.0) > 0.0001 {
			t.Errorf("Tsi[%d] with flat source = %f, want 0.0", i, v)
		}
	}
}

func TestTsi_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Tsi([]float64{}, 5, 13)
		if len(result) != 0 {
			t.Errorf("Tsi of empty slice: length = %d, want 0", len(result))
		}
	})

	t.Run("single_bar_source", func(t *testing.T) {
		result := Tsi([]float64{100}, 5, 13)
		if len(result) != 1 {
			t.Fatalf("Tsi of single bar: length = %d, want 1", len(result))
		}
		if !math.IsNaN(result[0]) {
			t.Errorf("Tsi of single bar should be NaN, got %f", result[0])
		}
	})

	t.Run("zero_short_length_all_nan", func(t *testing.T) {
		src := []float64{1, 2, 3, 4, 5}
		result := Tsi(src, 0, 13)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Tsi with shortLength=0: [%d] = %f, want NaN", i, v)
			}
		}
	})

	t.Run("zero_long_length_all_nan", func(t *testing.T) {
		src := []float64{1, 2, 3, 4, 5}
		result := Tsi(src, 5, 0)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Tsi with longLength=0: [%d] = %f, want NaN", i, v)
			}
		}
	})
}
