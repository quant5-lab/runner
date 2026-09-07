package regression

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/strategy"
)

/*
TestStrategyPyramidingCap_TVSemantics locks the TradingView pyramiding contract:
pyramiding=N caps simultaneous same-direction entries at N. The N+1-th entry on
the same direction (with the first N still open) must be rejected.

The historical bug used `sameDirectionCount > pyramiding` which silently doubled
the open position count when a signal fired on consecutive bars at pyramiding=1
(observed empirically as 264 trades vs TV's 122 on the Hull SBERP 1h golden).

Each case opens entries on successive bars and asserts the final open-trade count
on the bar AFTER the last entry attempt — by then all pending orders have filled
and the cap is fully exercised.
*/
func TestStrategyPyramidingCap_TVSemantics(t *testing.T) {
	type entryAttempt struct {
		id        string
		direction string
	}

	tests := []struct {
		name          string
		pyramiding    int
		attempts      []entryAttempt
		wantOpenCount int
	}{
		{
			// The canonical Hull regression: pyramiding=1 with two same-direction
			// entries on consecutive bars. Pre-fix this opened 2 trades; per TV
			// the 2nd must be blocked while the 1st is still open.
			name:          "pyramiding_1_blocks_second_long_entry",
			pyramiding:    1,
			attempts:      []entryAttempt{{"e1", strategy.Long}, {"e2", strategy.Long}},
			wantOpenCount: 1,
		},
		{
			name:          "pyramiding_1_blocks_second_short_entry",
			pyramiding:    1,
			attempts:      []entryAttempt{{"s1", strategy.Short}, {"s2", strategy.Short}},
			wantOpenCount: 1,
		},
		{
			// Pyramiding=2: two entries land, third is rejected.
			name:          "pyramiding_2_allows_two_blocks_third",
			pyramiding:    2,
			attempts:      []entryAttempt{{"e1", strategy.Long}, {"e2", strategy.Long}, {"e3", strategy.Long}},
			wantOpenCount: 2,
		},
		{
			// Pyramiding=3: three entries land, fourth is rejected.
			name:          "pyramiding_3_allows_three_blocks_fourth",
			pyramiding:    3,
			attempts:      []entryAttempt{{"e1", strategy.Long}, {"e2", strategy.Long}, {"e3", strategy.Long}, {"e4", strategy.Long}},
			wantOpenCount: 3,
		},
		{
			// Pyramiding=0 is the Pine default: one entry per direction is permitted.
			// Locks the special-case handling that prevents pyramiding=0 from blocking
			// the first entry too.
			name:          "pyramiding_0_allows_one_blocks_second",
			pyramiding:    0,
			attempts:      []entryAttempt{{"e1", strategy.Long}, {"e2", strategy.Long}},
			wantOpenCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := strategy.NewStrategy()
			s.CallWithPyramiding("Test", 10000, tt.pyramiding)

			bar, ts := 0, int64(1000)
			for _, a := range tt.attempts {
				if err := s.Entry(a.id, a.direction, 10, ""); err != nil {
					t.Fatalf("Entry(%q,%q) returned error: %v", a.id, a.direction, err)
				}
				bar++
				ts += 1000
				s.OnBarUpdate(bar, 100, ts)
			}
			// One more bar to drain any pending order created on the last attempt.
			bar++
			ts += 1000
			s.OnBarUpdate(bar, 100, ts)

			gotOpen := len(s.GetTradeHistory().GetOpenTrades())
			if gotOpen != tt.wantOpenCount {
				t.Errorf("open trades after pyramiding=%d with %d attempts: got %d, want %d",
					tt.pyramiding, len(tt.attempts), gotOpen, tt.wantOpenCount)
			}
		})
	}
}

/*
TestStrategyPyramidingCap_ConsecutiveBarSignals reproduces the Hull empirical
fingerprint directly: a signal that fires on two consecutive bars at pyramiding=1
must produce exactly one open trade (the second signal is blocked) — not two open
trades 1 bar apart.
*/
func TestStrategyPyramidingCap_ConsecutiveBarSignals(t *testing.T) {
	s := strategy.NewStrategy()
	s.CallWithPyramiding("Test", 10000, 1)

	// Bar 0: signal fires → entry "a" registered (will fill on bar 1 open).
	s.Entry("a", strategy.Long, 10, "")
	s.OnBarUpdate(1, 100, 1000)

	// Bar 1: same signal fires the very next bar. Pre-fix: this is allowed because
	// `1 > 1` is false → 2 open trades. Post-fix: blocked because count >= cap.
	s.Entry("a", strategy.Long, 10, "")
	s.OnBarUpdate(2, 101, 2000)

	open := s.GetTradeHistory().GetOpenTrades()
	if len(open) != 1 {
		t.Fatalf("consecutive-bar signals at pyramiding=1: expected 1 open trade, got %d", len(open))
	}
}
