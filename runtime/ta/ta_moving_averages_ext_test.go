package ta

import (
	"math"
	"testing"
)

/* =================== Wma =================== */

func TestWma_WarmupYieldsNaN(t *testing.T) {
	source := []float64{10, 20, 30, 40, 50}
	result := Wma(source, 4)

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Wma[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestWma_KnownValue(t *testing.T) {
	/* period=4, arithmetic source 1,2,3,4,5,6:
	   weights at i=3 (source window [1,2,3,4], j=0 is most-recent):
	     j=0 w=4 → 4*4=16, j=1 w=3 → 3*3=9, j=2 w=2 → 2*2=4, j=3 w=1 → 1*1=1
	   sum=30, weightSum=10, WMA=3.0
	   General: WMA of k..k+p-1 at any position = (2p+1)/3 + k - 1 for an arithmetic sequence. */
	source := []float64{1, 2, 3, 4, 5, 6}
	result := Wma(source, 4)

	if math.Abs(result[3]-3.0) > 0.0001 {
		t.Errorf("Wma[3] = %f, want 3.0", result[3])
	}
	if math.Abs(result[4]-4.0) > 0.0001 {
		t.Errorf("Wma[4] = %f, want 4.0", result[4])
	}
}

func TestWma_ConstantSourceReturnsSelf(t *testing.T) {
	/* Weights sum to weightSum, so WMA of constant c = c*(weightSum/weightSum) = c. */
	source := []float64{5, 5, 5, 5, 5, 5}
	result := Wma(source, 4)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-5.0) > 0.0001 {
			t.Errorf("Wma[%d] of constant source = %f, want 5.0", i, result[i])
		}
	}
}

func TestWma_OutputLength(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5}
	result := Wma(source, 3)

	if len(result) != len(source) {
		t.Errorf("Wma length = %d, want %d", len(result), len(source))
	}
}

func TestWma_ResultBoundedByWindowMinMax(t *testing.T) {
	/* WMA is a weighted average — output must be in [min(window), max(window)]. */
	source := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	period := 4
	result := Wma(source, period)
	minResult := Min(source, period)
	maxResult := Max(source, period)

	for i := period - 1; i < len(source); i++ {
		if result[i] < minResult[i]-0.0001 || result[i] > maxResult[i]+0.0001 {
			t.Errorf("Wma[%d]=%f outside window [%f, %f]", i, result[i], minResult[i], maxResult[i])
		}
	}
}

func TestWma_RecentBarsWeightedHigherThanSma(t *testing.T) {
	/* On a monotonically increasing source, WMA > SMA because it
	   assigns higher weights to the more recent (larger) values. */
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	period := 4
	wmaResult := Wma(source, period)
	smaResult := Sma(source, period)

	for i := period - 1; i < len(source); i++ {
		if wmaResult[i] <= smaResult[i]-0.0001 {
			t.Errorf("Wma[%d]=%f should exceed Sma[%d]=%f on ascending source",
				i, wmaResult[i], i, smaResult[i])
		}
	}
}

func TestWma_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Wma([]float64{}, 4)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("zero_period_returns_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Wma(source, 0)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})

	t.Run("period_one_returns_source_values", func(t *testing.T) {
		/* period=1: window=[source[i]], weight=1 → WMA = source[i]. */
		source := []float64{3, 7, 2, 9}
		result := Wma(source, 1)
		for i, v := range result {
			if math.Abs(v-source[i]) > 0.0001 {
				t.Errorf("period=1 at[%d]: got %f, want %f", i, v, source[i])
			}
		}
	})

	t.Run("period_exceeds_source_all_nan", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Wma(source, 10)
		for i, v := range result {
			if !math.IsNaN(v) {
				t.Errorf("Wma[%d] should be NaN when period > source length, got %f", i, v)
			}
		}
	})
}

/* =================== Alma =================== */

func TestAlma_WarmupYieldsNaN(t *testing.T) {
	source := []float64{10, 20, 30, 40, 50}
	result := Alma(source, 4, 0.85, 6)

	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Alma[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestAlma_ConstantSourceReturnsSelf(t *testing.T) {
	/* Gaussian weights are normalized by wSum, so ALMA of constant c = c. */
	source := []float64{7, 7, 7, 7, 7, 7}
	result := Alma(source, 4, 0.85, 6)

	for i := 3; i < len(result); i++ {
		if math.Abs(result[i]-7.0) > 0.0001 {
			t.Errorf("Alma[%d] of constant source = %f, want 7.0", i, result[i])
		}
	}
}

func TestAlma_OutputLength(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5}
	result := Alma(source, 3, 0.85, 6)

	if len(result) != len(source) {
		t.Errorf("Alma length = %d, want %d", len(result), len(source))
	}
}

func TestAlma_KnownValueSymmetricOffset(t *testing.T) {
	/* period=3, offset=0.5, sigma=6:
	     m = 0.5*(3-1) = 1.0,  s = 3/6 = 0.5
	     j=0: d=-1, w=exp(-2)≈0.1353;  j=1: d=0, w=1.0;  j=2: d=1, w=exp(-2)≈0.1353
	     wSum ≈ 1.2707
	   source=[1,2,3] at index 2: source[0]=1, source[1]=2, source[2]=3
	     val = 0.1353*1 + 1.0*2 + 0.1353*3 ≈ 2.5413
	     result = 2.5413/1.2707 = 2.0  (symmetric weights on arithmetic series pick midpoint) */
	source := []float64{1, 2, 3}
	result := Alma(source, 3, 0.5, 6)

	if math.Abs(result[2]-2.0) > 0.001 {
		t.Errorf("Alma[2] with symmetric offset = %f, want 2.0", result[2])
	}
}

func TestAlma_HighOffsetWeightsMoreRecentValues(t *testing.T) {
	/* offset near 1 shifts the Gaussian peak to the most recent bar.
	   On a strictly ascending source ALMA(offset=0.99) > ALMA(offset=0.01). */
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	period := 5
	highOffset := Alma(source, period, 0.99, 6)
	lowOffset := Alma(source, period, 0.01, 6)

	for i := period - 1; i < len(source); i++ {
		if highOffset[i] <= lowOffset[i]-0.0001 {
			t.Errorf("ALMA(offset=0.99)[%d]=%f should exceed ALMA(offset=0.01)[%d]=%f on ascending source",
				i, highOffset[i], i, lowOffset[i])
		}
	}
}

func TestAlma_ResultBoundedByWindowMinMax(t *testing.T) {
	/* ALMA is a normalized weighted average — must stay within [min(window), max(window)]. */
	source := []float64{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}
	period := 5
	result := Alma(source, period, 0.85, 6)
	minResult := Min(source, period)
	maxResult := Max(source, period)

	for i := period - 1; i < len(source); i++ {
		if result[i] < minResult[i]-0.0001 || result[i] > maxResult[i]+0.0001 {
			t.Errorf("Alma[%d]=%f outside window [%f, %f]", i, result[i], minResult[i], maxResult[i])
		}
	}
}

func TestAlma_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Alma([]float64{}, 4, 0.85, 6)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("zero_period_returns_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Alma(source, 0, 0.85, 6)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})

	t.Run("period_one_returns_source_values", func(t *testing.T) {
		/* period=1: m=0, s=1/sigma, w[0]=exp(0)=1, wSum=1 → result=source[i]. */
		source := []float64{3, 7, 2, 9}
		result := Alma(source, 1, 0.85, 6)
		for i, v := range result {
			if math.Abs(v-source[i]) > 0.0001 {
				t.Errorf("period=1 at[%d]: got %f, want %f", i, v, source[i])
			}
		}
	})
}

/* =================== Hma =================== */

func TestHma_WarmupYieldsNaN(t *testing.T) {
	/* period=4: halfPeriod=2 (wma1 warmup=1), sqrtPeriod=2 (round(sqrt(4))).
	   diff is valid from index period-1=3 onward.
	   Final WMA(sqrtPeriod=2) of diff needs 1 more bar → first valid at index 4. */
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Hma(source, 4)

	/* Indices 0..2 must be NaN regardless of period details. */
	for i := 0; i < 3; i++ {
		if !math.IsNaN(result[i]) {
			t.Errorf("Hma[%d] = %f, want NaN (warmup)", i, result[i])
		}
	}
}

func TestHma_ConstantSourceReturnsSelf(t *testing.T) {
	/* HMA = WMA(2*WMA(src,n/2) - WMA(src,n), sqrt(n)).
	   Constant c: each WMA = c → diff = 2c - c = c → outer WMA = c. */
	source := []float64{5, 5, 5, 5, 5, 5, 5, 5, 5}
	result := Hma(source, 4)

	for i := range result {
		if !math.IsNaN(result[i]) && math.Abs(result[i]-5.0) > 0.0001 {
			t.Errorf("Hma[%d] of constant source = %f, want 5.0", i, result[i])
		}
	}
}

func TestHma_OutputLength(t *testing.T) {
	source := []float64{1, 2, 3, 4, 5, 6, 7, 8}
	result := Hma(source, 4)

	if len(result) != len(source) {
		t.Errorf("Hma length = %d, want %d", len(result), len(source))
	}
}

func TestHma_ReducedLagVsWmaOnAscendingSource(t *testing.T) {
	/* HMA is designed to cancel lag vs WMA(n).
	   For a linearly increasing source with slope 1:
	     WMA(p) lag = (p-1)/3,  HMA(p) lag ≈ 0.
	   So HMA > WMA(n) on a rising source. */
	source := make([]float64, 25)
	for i := range source {
		source[i] = float64(i + 1)
	}
	period := 9
	hmaResult := Hma(source, period)
	wmaResult := Wma(source, period)

	/* Check the latter half where both are settled. */
	for i := len(source) - 5; i < len(source); i++ {
		if math.IsNaN(hmaResult[i]) || math.IsNaN(wmaResult[i]) {
			continue
		}
		if hmaResult[i] <= wmaResult[i]-0.0001 {
			t.Errorf("Hma[%d]=%f should exceed Wma[%d]=%f on ascending source (HMA has less lag)",
				i, hmaResult[i], i, wmaResult[i])
		}
	}
}

func TestHma_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Hma([]float64{}, 4)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("period_one_returns_source", func(t *testing.T) {
		/* period <= 1 returns source unchanged (guard condition in Hma). */
		source := []float64{3, 7, 2, 9}
		result := Hma(source, 1)
		if len(result) != len(source) {
			t.Errorf("period=1: length = %d, want %d", len(result), len(source))
		}
	})

	t.Run("period_two_produces_valid_output", func(t *testing.T) {
		/* period=2: halfPeriod=1, sqrtPeriod=1; warmup = period+sqrtPeriod-2 = 1. */
		source := []float64{1, 2, 3, 4, 5}
		result := Hma(source, 2)
		if math.IsNaN(result[len(result)-1]) {
			t.Error("Hma period=2: last bar should not be NaN")
		}
	})
}

/* =================== Kcw =================== */

func TestKcw_WarmupYieldsNaN(t *testing.T) {
	n := 10
	high := make([]float64, n)
	low := make([]float64, n)
	closeVals := make([]float64, n)
	for i := range high {
		high[i] = float64(100 + i)
		low[i] = float64(99 + i)
		closeVals[i] = float64(100 + i)
	}
	result := Kcw(closeVals, high, low, closeVals, 5, 1.5)

	if !math.IsNaN(result[0]) {
		t.Errorf("Kcw[0] = %f, want NaN (warmup)", result[0])
	}
}

func TestKcw_ZeroAtrYieldsZero(t *testing.T) {
	/* high = low = close (flat bars) → TR = 0 every bar → ATR = 0 → KCW = 0. */
	n := 12
	price := make([]float64, n)
	for i := range price {
		price[i] = 100.0
	}
	result := Kcw(price, price, price, price, 4, 1.5)

	for i := range result {
		if !math.IsNaN(result[i]) && math.Abs(result[i]) > 0.0001 {
			t.Errorf("Kcw[%d]=%f with zero ATR, want 0.0", i, result[i])
		}
	}
}

func TestKcw_NonNegativeWithPositiveMult(t *testing.T) {
	/* ATR ≥ 0 and EMA > 0 with positive prices → KCW ≥ 0 when mult ≥ 0. */
	n := 15
	high := make([]float64, n)
	low := make([]float64, n)
	closeVals := make([]float64, n)
	for i := range high {
		high[i] = 100 + float64(i%5)
		low[i] = 100 - float64(i%3)
		closeVals[i] = 100.0
	}
	result := Kcw(closeVals, high, low, closeVals, 5, 1.5)

	for i := range result {
		if !math.IsNaN(result[i]) && result[i] < -0.0001 {
			t.Errorf("Kcw[%d]=%f is negative with positive mult", i, result[i])
		}
	}
}

func TestKcw_ScalesProportionallyWithMult(t *testing.T) {
	/* KCW = 2*mult*ATR/EMA → KCW(mult=2) = 2*KCW(mult=1). */
	n := 15
	high := make([]float64, n)
	low := make([]float64, n)
	closeVals := make([]float64, n)
	for i := range high {
		high[i] = 100 + float64(i%7+1)
		low[i] = 100 - float64(i%5+1)
		closeVals[i] = 100.0 + float64(i%3)
	}
	r1 := Kcw(closeVals, high, low, closeVals, 5, 1.0)
	r2 := Kcw(closeVals, high, low, closeVals, 5, 2.0)

	for i := range r1 {
		if math.IsNaN(r1[i]) || math.IsNaN(r2[i]) {
			continue
		}
		if math.Abs(r2[i]-2*r1[i]) > 0.0001 {
			t.Errorf("Kcw mult=2[%d]=%f != 2 * mult=1[%d]=%f", i, r2[i], i, r1[i])
		}
	}
}

func TestKcw_OutputLength(t *testing.T) {
	source := []float64{100, 101, 102, 103, 104}
	result := Kcw(source, source, source, source, 3, 1.5)

	if len(result) != len(source) {
		t.Errorf("Kcw length = %d, want %d", len(result), len(source))
	}
}

func TestKcw_EdgeCases(t *testing.T) {
	t.Run("empty_source", func(t *testing.T) {
		result := Kcw([]float64{}, []float64{}, []float64{}, []float64{}, 4, 1.5)
		if len(result) != 0 {
			t.Errorf("empty source: got length %d", len(result))
		}
	})

	t.Run("zero_period_returns_source", func(t *testing.T) {
		source := []float64{1, 2, 3}
		result := Kcw(source, source, source, source, 0, 1.5)
		if len(result) != 3 {
			t.Errorf("zero period: length = %d, want 3", len(result))
		}
	})
}

/* =================== Sar =================== */

func TestSar_OutputLength(t *testing.T) {
	high := []float64{10, 11, 12, 13, 12, 11, 10}
	low := []float64{9, 10, 11, 12, 11, 10, 9}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	if len(result) != len(high) {
		t.Errorf("Sar length = %d, want %d", len(result), len(high))
	}
}

func TestSar_EmptySourceReturnsNil(t *testing.T) {
	result := Sar([]float64{}, []float64{}, 0.02, 0.02, 0.2)
	if result != nil {
		t.Errorf("Sar of empty source should be nil, got length %d", len(result))
	}
}

func TestSar_SingleBarFirstIsNaN(t *testing.T) {
	/* n=1: implementation sets result[0]=NaN then returns early. */
	result := Sar([]float64{10}, []float64{9}, 0.02, 0.02, 0.2)

	if len(result) != 1 {
		t.Fatalf("single bar Sar: length = %d, want 1", len(result))
	}
	if !math.IsNaN(result[0]) {
		t.Errorf("single bar Sar[0] = %f, want NaN", result[0])
	}
}

func TestSar_UptrendSarBelowHighs(t *testing.T) {
	/* Strictly ascending highs → initial trend is upward.
	   In uptrend, SAR represents a stop below the current price. */
	high := []float64{10, 11, 12, 13, 14, 15, 16, 17}
	low := []float64{9, 10, 11, 12, 13, 14, 15, 16}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	for i := 0; i < len(result); i++ {
		if result[i] > high[i]+0.0001 {
			t.Errorf("Sar[%d]=%f exceeds high=%f during uptrend", i, result[i], high[i])
		}
	}
}

func TestSar_DowntrendSarAboveLows(t *testing.T) {
	/* Strictly descending highs → initial trend is downward.
	   In downtrend, SAR sits above the current price as a resistance stop. */
	high := []float64{17, 16, 15, 14, 13, 12, 11, 10}
	low := []float64{16, 15, 14, 13, 12, 11, 10, 9}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	/* Skip first bar (initial sar = high[0], same as bar high — boundary). */
	for i := 1; i < len(result)-1; i++ {
		if result[i] < low[i]-0.0001 {
			t.Errorf("Sar[%d]=%f below low=%f during downtrend", i, result[i], low[i])
		}
	}
}

func TestSar_AllValuesFinite(t *testing.T) {
	/* SAR must produce finite values for every bar (n >= 2). */
	high := []float64{10, 11, 12, 11, 10, 9, 10, 11, 12}
	low := []float64{9, 10, 11, 10, 9, 8, 9, 10, 11}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	for i, v := range result {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("Sar[%d] = %v, want finite value", i, v)
		}
	}
}

func TestSar_AccelerationFactorCappedAtMax(t *testing.T) {
	/* Even in a long, unbroken trend the AF cannot exceed max.
	   Verify SAR remains finite and non-NaN over many bars. */
	n := 100
	high := make([]float64, n)
	low := make([]float64, n)
	for i := range high {
		high[i] = float64(100 + i)
		low[i] = float64(99 + i)
	}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	for i, v := range result {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("Sar[%d] = %v after long trend, want finite", i, v)
		}
	}
}

func TestSar_ReversalFlipsTrendSide(t *testing.T) {
	/* A sharp price reversal must cause the trend direction to flip.
	   Start: uptrend for 6 bars, then price collapses sharply.
	   After reversal, SAR should move to the other side of price. */
	high := []float64{10, 11, 12, 13, 14, 15, 5, 4, 3, 2}
	low := []float64{9, 10, 11, 12, 13, 14, 4, 3, 2, 1}
	result := Sar(high, low, 0.02, 0.02, 0.2)

	/* All values must be finite after the reversal. */
	for i, v := range result {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("Sar[%d]=%v after sharp reversal, want finite", i, v)
		}
	}

	/* Bar 6: high=5, low=4 — price collapsed below prior SAR → downtrend begins.
	   SAR at bar 6 should be set to the prior EP (high of the uptrend), which is ≥ high[6]=5. */
	if result[6] < high[6] {
		t.Errorf("Sar[6]=%f should be >= high[6]=%f after bearish reversal (SAR flips to EP)", result[6], high[6])
	}
}

func TestSar_StartParamSetsInitialAF(t *testing.T) {
	/* The start AF controls initial sensitivity.
	   Larger start → SAR moves faster toward price in early bars.
	   With start=0.1 the SAR should converge faster than start=0.01. */
	high := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	low := []float64{9, 10, 11, 12, 13, 14, 15, 16, 17, 18}

	fastSAR := Sar(high, low, 0.1, 0.1, 0.2)
	slowSAR := Sar(high, low, 0.01, 0.01, 0.2)

	/* In an uptrend, a larger AF causes SAR to move up faster (larger SAR values). */
	found := false
	for i := 1; i < len(fastSAR); i++ {
		if fastSAR[i] > slowSAR[i]+0.0001 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Sar with larger start AF should produce higher SAR values in uptrend than smaller AF")
	}
}
