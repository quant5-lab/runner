package ta

import (
	"math"
	"testing"
)

/* =================== Percentrank =================== */

func TestPercentrank_WarmupYieldsNaN(t *testing.T) {
	source := []float64{10, 20, 30, 40, 50}
	result := Percentrank(source, 3)

	for i := 0; i < 2; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Percentrank[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestPercentrank_MonotonicallyIncreasingSource(t *testing.T) {
	/* In an upward window the current bar is always the maximum, so
	   all bars except the current one are strictly less → rank = (period-1)/period * 100. */
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	period := 4
	result := Percentrank(source, period)

	expected := float64(period-1) / float64(period) * 100.0
	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]-expected) > 0.0001 {
			t.Errorf("Percentrank[%d] = %f, want %f (monotone ascending)", i, result[i], expected)
		}
	}
}

func TestPercentrank_ConstantSource(t *testing.T) {
	/* No bar is strictly less than the current bar → rank = 0. */
	source := []float64{5, 5, 5, 5, 5, 5}
	result := Percentrank(source, 4)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-0.0) > 0.0001 {
			t.Errorf("Percentrank[%d] of constant source = %f, want 0.0", i, result[i])
		}
	}
}

func TestPercentrank_OutputLength(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5}
	result := Percentrank(source, 3)

	if len(result) != len(source) {
		t.Errorf("Percentrank length = %d, want %d", len(result), len(source))
	}
}

func TestPercentrank_RangeIsZeroToHundred(t *testing.T) {
	source := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}
	result := Percentrank(source, 5)

	for i := 4; i < len(result); i++ {
		if result[i] < 0 || result[i] > 100.0+0.0001 {
			t.Errorf("Percentrank[%d] = %f out of [0, 100] range", i, result[i])
		}
	}
}

func TestPercentrank_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Percentrank([]float64{}, 5)
		if len(result) != 0 {
			t.Errorf("empty source should return empty, got length %d", len(result))
		}
	})

	t.Run("zero_period", func(t *testing.T) {
		result := Percentrank([]float64{1, 2, 3}, 0)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})

	t.Run("period_one", func(t *testing.T) {
		/* Period-1 window contains only the current bar itself, which is not < current → 0. */
		source := []float64{5, 3, 8, 2}
		result := Percentrank(source, 1)
		for i, v := range result {
			if math.Abs(v-0.0) > 0.0001 {
				t.Errorf("period=1 at index %d: got %f, want 0.0", i, v)
			}
		}
	})

	t.Run("period_exceeds_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Percentrank(source, 10)
		for _, v := range result {
			if !math.IsNaN(v) {
				t.Error("all bars should be NaN when period > source length")
			}
		}
	})

	t.Run("monotone_descending_has_zero_rank", func(t *testing.T) {
		/* The current bar is always the minimum in a descending window → rank = 0. */
		source := []float64{10, 9, 8, 7, 6}
		result := Percentrank(source, 3)
		for i := 2; i < len(result); i++ {
			if math.Abs(result[i]-0.0) > 0.0001 {
				t.Errorf("Percentrank[%d] descending = %f, want 0.0", i, result[i])
			}
		}
	})
}

/* =================== PercentileNearestRank =================== */

func TestPercentileNearestRank_WarmupYieldsNaN(t *testing.T) {
	result := PercentileNearestRank([]float64{10, 20, 30, 40}, 4, 50)

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("PercentileNearestRank[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestPercentileNearestRank_KnownValues(t *testing.T) {
	/* period=5, pct=50: idx=ceil(0.5*5)-1=2 → median of sorted window */
	source := []float64{10, 30, 20, 50, 40, 60, 15, 25}
	result := PercentileNearestRank(source, 5, 50)

	/* bar 4: window values [10,30,20,50,40], sorted [10,20,30,40,50], idx=2 → 30 */
	if math.Abs(result[4]-30.0) > 0.0001 {
		t.Errorf("PercentileNearestRank[4] = %f, want 30.0", result[4])
	}
}

func TestPercentileNearestRank_Pct100ReturnsMax(t *testing.T) {
	source := []float64{5, 1, 8, 3, 9, 2, 7}
	period := 5
	result := PercentileNearestRank(source, period, 100)
	expectedMaxResult := Max(source, period)

	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]-expectedMaxResult[i]) > 0.0001 {
			t.Errorf("pct=100 at [%d]: got %f, want max=%f", i, result[i], expectedMaxResult[i])
		}
	}
}

func TestPercentileNearestRank_ConstantSourceReturnsSelf(t *testing.T) {
	source := []float64{7, 7, 7, 7, 7, 7}
	result := PercentileNearestRank(source, 4, 50)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-7.0) > 0.0001 {
			t.Errorf("constant source[%d] = %f, want 7.0", i, result[i])
		}
	}
}

func TestPercentileNearestRank_OutputLength(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5}
	result := PercentileNearestRank(source, 3, 50)

	if len(result) != len(source) {
		t.Errorf("length = %d, want %d", len(result), len(source))
	}
}

func TestPercentileNearestRank_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := PercentileNearestRank([]float64{}, 3, 50)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("zero_period", func(t *testing.T) {
		result := PercentileNearestRank([]float64{1, 2, 3}, 0, 50)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})

	t.Run("pct_clamps_to_last_index", func(t *testing.T) {
		/* idx = ceil(120/100 * 3) - 1 = ceil(3.6) - 1 = 4-1 = 3 >= 3, clamped to 2 */
		source := []float64{1, 3, 2, 5, 4}
		result := PercentileNearestRank(source, 3, 120)
		expectedMax := Max(source, 3)
		for i := 2; i < len(result); i++ {
			if math.Abs(result[i]-expectedMax[i]) > 0.0001 {
				t.Errorf("pct>100 at[%d]: got %f, want max %f", i, result[i], expectedMax[i])
			}
		}
	})

	t.Run("period_one", func(t *testing.T) {
		/* period=1: only one bar, idx = ceil(pct/100 * 1) - 1 = 0 → current value. */
		source := []float64{3, 7, 2, 9}
		result := PercentileNearestRank(source, 1, 50)
		for i, v := range result {
			if math.Abs(v-source[i]) > 0.0001 {
				t.Errorf("period=1 at[%d]: got %f, want %f", i, v, source[i])
			}
		}
	})
}

/* =================== PercentileLinearInterpolation =================== */

func TestPercentileLinearInterpolation_WarmupYieldsNaN(t *testing.T) {
	result := PercentileLinearInterpolation([]float64{10, 20, 30, 40}, 4, 50)

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("PLI[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestPercentileLinearInterpolation_Pct50IsMedianOfEvenPeriod(t *testing.T) {
	/* period=4, pct=50: rank = 0.5*(4-1) = 1.5
	   lower=1, upper=2, frac=0.5 → w[1] + 0.5*(w[2]-w[1]) */
	source := []float64{1, 4, 3, 2, 6, 5}
	result := PercentileLinearInterpolation(source, 4, 50)

	/* bar 3: window (newest first) = [2,3,4,1] → sorted [1,2,3,4]
	   rank=1.5, lower=1→2, upper=2→3, frac=0.5 → 2 + 0.5*(3-2) = 2.5 */
	if math.Abs(result[3]-2.5) > 0.0001 {
		t.Errorf("PLI[3] = %f, want 2.5", result[3])
	}
}

func TestPercentileLinearInterpolation_Pct100ReturnsMax(t *testing.T) {
	/* pct=100: rank = (period-1), lower=period-1, upper=period → clamped to last */
	source := []float64{5, 1, 8, 3, 9, 2, 7}
	period := 5
	result := PercentileLinearInterpolation(source, period, 100)
	maxResult := Max(source, period)

	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]-maxResult[i]) > 0.0001 {
			t.Errorf("pct=100 at[%d]: PLI=%f, Max=%f", i, result[i], maxResult[i])
		}
	}
}

func TestPercentileLinearInterpolation_Pct0ReturnsMin(t *testing.T) {
	/* pct=0: rank=0, lower=0, frac=0 → w[0] = min. */
	source := []float64{5, 1, 8, 3, 9, 2, 7}
	period := 5
	result := PercentileLinearInterpolation(source, period, 0)
	minResult := Min(source, period)

	for i := period - 1; i < len(result); i++ {
		if math.Abs(result[i]-minResult[i]) > 0.0001 {
			t.Errorf("pct=0 at[%d]: PLI=%f, Min=%f", i, result[i], minResult[i])
		}
	}
}

func TestPercentileLinearInterpolation_ConstantSourceReturnsSelf(t *testing.T) {
	source := []float64{7, 7, 7, 7, 7, 7}
	result := PercentileLinearInterpolation(source, 4, 75)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-7.0) > 0.0001 {
			t.Errorf("constant source[%d] = %f, want 7.0", i, result[i])
		}
	}
}

func TestPercentileLinearInterpolation_ResultBoundedByMinMax(t *testing.T) {
	source := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	period := 4

	pliLow := PercentileLinearInterpolation(source, period, 25)
	pliHigh := PercentileLinearInterpolation(source, period, 75)
	minResult := Min(source, period)
	maxResult := Max(source, period)

	for i := period - 1; i < len(source); i++ {
		if pliLow[i] < minResult[i]-0.0001 || pliLow[i] > maxResult[i]+0.0001 {
			t.Errorf("PLI25[%d]=%f outside [%f, %f]", i, pliLow[i], minResult[i], maxResult[i])
		}
		if pliHigh[i] < minResult[i]-0.0001 || pliHigh[i] > maxResult[i]+0.0001 {
			t.Errorf("PLI75[%d]=%f outside [%f, %f]", i, pliHigh[i], minResult[i], maxResult[i])
		}
	}
}

func TestPercentileLinearInterpolation_MonotonicallyIncreasingPct(t *testing.T) {
	/* PLI(pct=25) <= PLI(pct=50) <= PLI(pct=75) for every bar. */
	source := []float64{4, 2, 7, 1, 9, 3, 8, 5, 6, 10}
	period := 5
	p25 := PercentileLinearInterpolation(source, period, 25)
	p50 := PercentileLinearInterpolation(source, period, 50)
	p75 := PercentileLinearInterpolation(source, period, 75)

	for i := period - 1; i < len(source); i++ {
		if p25[i] > p50[i]+0.0001 {
			t.Errorf("PLI25[%d]=%f > PLI50[%d]=%f (ordering violated)", i, p25[i], i, p50[i])
		}
		if p50[i] > p75[i]+0.0001 {
			t.Errorf("PLI50[%d]=%f > PLI75[%d]=%f (ordering violated)", i, p50[i], i, p75[i])
		}
	}
}

func TestPercentileLinearInterpolation_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := PercentileLinearInterpolation([]float64{}, 3, 50)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("zero_period", func(t *testing.T) {
		result := PercentileLinearInterpolation([]float64{1, 2, 3}, 0, 50)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})

	t.Run("period_one_returns_source", func(t *testing.T) {
		source := []float64{3, 7, 2, 9}
		result := PercentileLinearInterpolation(source, 1, 50)
		for i, v := range result {
			if math.Abs(v-source[i]) > 0.0001 {
				t.Errorf("period=1 at[%d]: got %f, want %f", i, v, source[i])
			}
		}
	})
}

/* =================== Correlation =================== */

func TestCorrelation_WarmupYieldsNaN(t *testing.T) {
	src1 := []float64{1, 2, 3, 4, 5}
	src2 := []float64{2, 4, 6, 8, 10}
	result := Correlation(src1, src2, 4)

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Correlation[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestCorrelation_IdenticalSeriesIsOne(t *testing.T) {
	src := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	result := Correlation(src, src, 5)

	for i := 4; i < len(result); i++ {
		if math.Abs(result[i]-1.0) > 0.0001 {
			t.Errorf("Correlation[%d] identical series = %f, want 1.0", i, result[i])
		}
	}
}

func TestCorrelation_NegatedSeriesIsMinusOne(t *testing.T) {
	src := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	negSrc := make([]float64, len(src))
	for i, v := range src {
		negSrc[i] = -v
	}

	result := Correlation(src, negSrc, 5)

	for i := 4; i < len(result); i++ {
		if math.Abs(result[i]+1.0) > 0.0001 {
			t.Errorf("Correlation[%d] negated series = %f, want -1.0", i, result[i])
		}
	}
}

func TestCorrelation_ConstantSourceIsZero(t *testing.T) {
	/* Zero variance → correlation is undefined; implementation returns 0 */
	src1 := []float64{5, 5, 5, 5, 5, 5}
	src2 := []float64{1, 2, 3, 4, 5, 6}
	result := Correlation(src1, src2, 4)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-0.0) > 0.0001 {
			t.Errorf("Correlation[%d] with constant source = %f, want 0.0", i, result[i])
		}
	}
}

func TestCorrelation_ResultBoundedBetweenMinusOneAndOne(t *testing.T) {
	src1 := []float64{1, 3, 2, 5, 4, 7, 6, 8, 9, 2}
	src2 := []float64{2, 1, 4, 3, 6, 5, 8, 7, 1, 9}
	result := Correlation(src1, src2, 4)

	for i := 3; i < len(result); i++ {
		if result[i] < -1.0-0.0001 || result[i] > 1.0+0.0001 {
			t.Errorf("Correlation[%d] = %f outside [-1, 1]", i, result[i])
		}
	}
}

func TestCorrelation_SymmetricArguments(t *testing.T) {
	/* correlation(a, b) == correlation(b, a) */
	src1 := []float64{1, 4, 2, 5, 3, 7, 2, 8}
	src2 := []float64{3, 2, 5, 1, 6, 4, 7, 3}
	result12 := Correlation(src1, src2, 4)
	result21 := Correlation(src2, src1, 4)

	for i := 3; i < len(result12); i++ {
		if math.Abs(result12[i]-result21[i]) > 0.0001 {
			t.Errorf("Correlation[%d] not symmetric: corr(a,b)=%f vs corr(b,a)=%f", i, result12[i], result21[i])
		}
	}
}

func TestCorrelation_OutputLength(t *testing.T) {
	src1 := []float64{1, 2, 3, 4, 5}
	src2 := []float64{5, 4, 3, 2, 1}
	result := Correlation(src1, src2, 3)

	if len(result) != len(src1) {
		t.Errorf("length = %d, want %d", len(result), len(src1))
	}
}

func TestCorrelation_KnownValue(t *testing.T) {
	/* Manually computed: src1=[1,2,3], src2=[2,4,6], period=3
	   Both are perfectly correlated → r=1 (identical shapes, scale doesn't matter). */
	src1 := []float64{1, 2, 3}
	src2 := []float64{2, 4, 6}
	result := Correlation(src1, src2, 3)

	if math.Abs(result[2]-1.0) > 0.0001 {
		t.Errorf("Correlation of proportional series = %f, want 1.0", result[2])
	}
}

func TestCorrelation_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Correlation([]float64{}, []float64{}, 3)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("mismatched_lengths_uses_shorter", func(t *testing.T) {
		src1 := []float64{1, 2, 3, 4, 5}
		src2 := []float64{1, 2, 3}
		result := Correlation(src1, src2, 2)
		if len(result) != 3 {
			t.Errorf("mismatched lengths: got %d, want 3 (shorter)", len(result))
		}
	})

	t.Run("zero_period_returns_empty_array", func(t *testing.T) {
		result := Correlation([]float64{1, 2, 3}, []float64{1, 2, 3}, 0)
		if len(result) != 3 {
			t.Errorf("zero period: got length %d, want 3", len(result))
		}
	})
}
