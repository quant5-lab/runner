package regression

import (
	"testing"

	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

// TestGoldenRunnerSizeEmission_NonZeroWhenTVHasSizeData asserts that the
// golden snapshot for every alignment case never contains a runner trade with
// Size=0 that can be paired to a TV reference trade.
//
// This invariant is orthogonal to magnitude comparison:
//   - SkipSizeRatchetReason suppresses MatchSize (magnitude comparison) when TV
//     and runner equity diverge, making the size ratio incomparable.
//   - It does NOT suppress the quantity-emission check: a Size=0 on a matched
//     runner trade means the runner failed to emit a position quantity at all,
//     which is a codegen or sizing defect regardless of the equity basis.
//
// Cases where the TV CSV carries no size column are skipped; there is no size
// reference to pair against and no quantity check is applicable.
func TestGoldenRunnerSizeEmission_NonZeroWhenTVHasSizeData(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tvTrades := loadTVTradesInFixtureWindow(t, root, tc)
			if !tvref.HasSizeData(tvTrades) {
				t.Skip("TV reference CSV carries no size column; quantity emission check does not apply")
			}
			runnerTrades := loadGoldenRunnerTrades(t, root, tc)
			n := tvref.ZeroSizeRunnerMatches(runnerTrades, tvTrades, tc.Tolerance.Time, tc.Tolerance.Price)
			if n > 0 {
				t.Errorf(
					"%d golden runner trade(s) carry Size=0 on a TV-matched pair — "+
						"runner must emit a positive quantity for every matched trade; "+
						"SkipSizeRatchetReason=%q suppresses magnitude comparison only, not quantity emission",
					n, tc.SkipSizeRatchetReason,
				)
			}
		})
	}
}
