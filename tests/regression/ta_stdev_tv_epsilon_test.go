package regression

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/ta"
)

/*
TestStdev_TVEpsilonCompensation locks the TradingView ta.stdev per-deviation
compensation rule documented in Pine's reference implementation. For each
(source-mean) deviation the rule is:

	|diff| <= 1e-10 -> 0
	|diff| <= 1e-4  -> 1e-5
	otherwise       -> diff (unchanged)

The rule mitigates floating-point cancellation drift that would otherwise
flip marginal Bollinger Band crossing comparisons. The cases below are
hand-constructed so the rule's effect is observable in the returned stdev
(removing the rule changes the value).
*/
func TestStdev_TVEpsilonCompensation(t *testing.T) {
	t.Run("sub-1e-10 deviation collapses to zero", func(t *testing.T) {
		// All values identical → mean equals every value → every diff is 0.
		// Float arithmetic over many bars can produce drift on the order of
		// 1e-12. The epsilon rule clamps that to exactly 0.
		src := []float64{1234.5678, 1234.5678, 1234.5678, 1234.5678, 1234.5678}
		got := ta.Stdev(src, 5)
		if math.IsNaN(got[4]) {
			t.Fatalf("expected finite stdev, got NaN")
		}
		if got[4] != 0 {
			t.Errorf("expected exactly 0 stdev for constant input, got %g", got[4])
		}
	})

	t.Run("deviation between 1e-10 and 1e-4 floors at 1e-5", func(t *testing.T) {
		// Construct a window where every (src-mean) diff has |diff| in
		// (1e-10, 1e-4]. With the rule each diff becomes exactly 1e-5 and
		// variance = period * (1e-5)^2 / period = 1e-10, stdev = 1e-5.
		// Without the rule the diffs are tiny non-uniform values producing
		// a much smaller stdev.
		mean := 100.0
		src := []float64{
			mean - 5e-5,
			mean - 1e-5,
			mean,
			mean + 1e-5,
			mean + 5e-5,
		}
		got := ta.Stdev(src, 5)
		if math.IsNaN(got[4]) {
			t.Fatalf("expected finite stdev, got NaN")
		}
		// With the rule, each |diff|<=1e-4 becomes 1e-5, so stdev = sqrt(5*(1e-5)^2/5) = 1e-5.
		// (One of the diffs is exactly 0 and gets clamped to 0 — but |0|<=1e-10 → 0
		// anyway, contributing 0 to variance. The non-zero diffs all become 1e-5
		// → variance = 4*(1e-5)^2/5 = 8e-11 → stdev ≈ 8.944e-6.)
		want := math.Sqrt(4 * 1e-10 / 5)
		if math.Abs(got[4]-want) > 1e-12 {
			t.Errorf("stdev = %g, want %g (TV epsilon-floor rule)", got[4], want)
		}
	})

	t.Run("large deviations pass through unchanged", func(t *testing.T) {
		// Standard population stdev with diffs all > 1e-4 should match the
		// naive (no-epsilon) formula exactly.
		src := []float64{10, 12, 14, 16, 18, 20}
		got := ta.Stdev(src, 3)
		// Naive: stdev of {10,12,14} = stdev of {12,14,16} = ... = sqrt(((±2)^2+0^2+(±2)^2)/3) ≈ 1.632993
		want := math.Sqrt((4.0 + 0.0 + 4.0) / 3.0)
		for i := 2; i < 6; i++ {
			if math.Abs(got[i]-want) > 1e-9 {
				t.Errorf("stdev[%d] = %g, want %g (unchanged by epsilon rule)", i, got[i], want)
			}
		}
	})
}

/*
TestStdev_BandaidGuard fails if the epsilon-compensation is removed.
This guards against the rule being silently reverted to the naive variance
formula. Specifically, with constant input the runner must return EXACTLY 0
(not a tiny epsilon-scale residue from floating cancellation).
*/
func TestStdev_BandaidGuard(t *testing.T) {
	// Repeating the same large value 20 times: any naive variance loop
	// summing (v-mean)^2 yields exactly 0 too, so we need a stronger probe.
	// We use values whose summation order is known to produce float noise.
	src := make([]float64, 20)
	// 20 copies of a value with non-trivial binary representation
	for i := range src {
		src[i] = 0.1 + 0.2 // not exactly 0.3 in IEEE-754
	}
	got := ta.Stdev(src, 20)
	if got[19] != 0 {
		t.Errorf("constant 0.1+0.2 stdev = %g, want exactly 0 (TV epsilon rule should clamp drift)", got[19])
	}
}
