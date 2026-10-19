package regression

import "testing"

func TestLiveRunnerTVAlignment(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		t.Run(tc.Name, func(t *testing.T) {
			assertTVAlignment(t, loadLiveRunnerTrades(t, root, tc), loadTVTradesInFixtureWindow(t, root, tc), tc)
		})
	}
}

func TestGoldenRunnerTVAlignment(t *testing.T) {
	root := projectRootFromCwd()
	for _, tc := range tvAlignmentCases() {
		t.Run(tc.Name, func(t *testing.T) {
			assertTVAlignment(t, loadGoldenRunnerTrades(t, root, tc), loadTVTradesInFixtureWindow(t, root, tc), tc)
		})
	}
}
