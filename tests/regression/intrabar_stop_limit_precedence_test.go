package regression

import (
	"math"
	"testing"

	"github.com/quant5-lab/runner/runtime/strategy"
)

/*
TestIntrabarStopLimitPrecedence_StopWins ratchets the TradingView conservative-fill
rule: when both stop and limit levels are breached within the same bar (each level
falls within [barLow, barHigh]), the stop fills first regardless of bar OHLC
ordering. This matches the TV broker emulator under default settings
(calc_on_every_tick=false, no intrabar tick data): without sub-bar resolution the
worse-for-trader outcome is assumed.

Previously a proximity-to-open heuristic returned whichever level was nearer to
the segment start of the assumed intrabar path. That selection diverged from
TradingView on bars where both levels breached but the limit was nearer to the
open — most visibly in TestSupertrend_BTCUSDT_Hourly trade #44, whose expected
exit was the stop at 87119.99 but the runner picked the limit at 87164.54.

Symbol-agnostic shape: the test drives only PendingExitManager's public surface
with constructed OHLC bars, so it is immune to strategy-template changes and
will only break if the precedence rule is reverted.
*/
func TestIntrabarStopLimitPrecedence_StopWins(t *testing.T) {
	const registrationBar = 0
	const checkBar = 1 // satisfies IsEligible: currentBar > FirstRegisteredBar

	cases := []struct {
		name       string
		direction  string
		stopLevel  float64
		limitLevel float64
		barOpen    float64
		barHigh    float64
		barLow     float64
		wantPrice  float64
	}{
		// Long: stop is below entry (filled on a drop), limit is above (filled on a rise).
		// Both inside [low, high] → both breach within the bar.
		{
			name:      "long_open_near_high_both_breach_stop_wins",
			direction: strategy.Long,
			stopLevel: 95, limitLevel: 110,
			barOpen: 112, barHigh: 115, barLow: 90,
			wantPrice: 95,
		},
		{
			name:      "long_open_near_low_both_breach_stop_wins",
			direction: strategy.Long,
			stopLevel: 95, limitLevel: 110,
			barOpen: 93, barHigh: 115, barLow: 90,
			wantPrice: 95,
		},
		{
			name:      "long_open_equidistant_both_breach_stop_wins",
			direction: strategy.Long,
			stopLevel: 95, limitLevel: 110,
			barOpen: 102.5, barHigh: 115, barLow: 90,
			wantPrice: 95,
		},
		// Short: stop is above entry, limit is below. Both inside [low, high].
		{
			name:      "short_open_near_high_both_breach_stop_wins",
			direction: strategy.Short,
			stopLevel: 105, limitLevel: 90,
			barOpen: 108, barHigh: 110, barLow: 85,
			wantPrice: 105,
		},
		{
			name:      "short_open_near_low_both_breach_stop_wins",
			direction: strategy.Short,
			stopLevel: 105, limitLevel: 90,
			barOpen: 87, barHigh: 110, barLow: 85,
			wantPrice: 105,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pem := strategy.NewPendingExitManager()
			pem.RegisterExit("exit1", "entry1", tc.stopLevel, tc.limitLevel, registrationBar, "")

			exits := pem.GetExitsForEntry("entry1")
			if len(exits) != 1 {
				t.Fatalf("expected 1 exit, got %d", len(exits))
			}

			trade := strategy.Trade{EntryID: "entry1", Direction: tc.direction}
			triggered, price, kind := pem.CheckExitTriggered(
				exits[0], trade, checkBar, tc.barOpen, tc.barHigh, tc.barLow,
			)
			if !triggered {
				t.Fatalf("expected trigger on bar where both levels breach; got triggered=false")
			}
			if kind != "stop" {
				t.Errorf("expected fill kind=\"stop\" (conservative-fill), got %q", kind)
			}
			if price != tc.wantPrice {
				t.Errorf("expected fill price %v (stop), got %v", tc.wantPrice, price)
			}
		})
	}
}

/*
TestIntrabarStopLimitPrecedence_SingleSideBreachUnaffected guards the negative
case: the conservative-fill rule applies ONLY when both levels breach. When
just one side breaches, the breached side fills at its own level — the stop
precedence must not over-fire and shadow a clean limit-only or stop-only
trigger.
*/
func TestIntrabarStopLimitPrecedence_SingleSideBreachUnaffected(t *testing.T) {
	const registrationBar = 0
	const checkBar = 1
	nan := math.NaN()

	cases := []struct {
		name       string
		direction  string
		stopLevel  float64
		limitLevel float64
		barHigh    float64
		barLow     float64
		wantPrice  float64
		wantKind   string
	}{
		{
			name:      "long_only_stop_breaches",
			direction: strategy.Long,
			stopLevel: 95, limitLevel: nan,
			barHigh: 105, barLow: 90,
			wantPrice: 95, wantKind: "stop",
		},
		{
			name:      "long_only_limit_breaches",
			direction: strategy.Long,
			stopLevel: nan, limitLevel: 110,
			barHigh: 115, barLow: 100,
			wantPrice: 110, wantKind: "limit",
		},
		{
			name:      "short_only_stop_breaches",
			direction: strategy.Short,
			stopLevel: 105, limitLevel: nan,
			barHigh: 110, barLow: 95,
			wantPrice: 105, wantKind: "stop",
		},
		{
			name:      "short_only_limit_breaches",
			direction: strategy.Short,
			stopLevel: nan, limitLevel: 90,
			barHigh: 100, barLow: 85,
			wantPrice: 90, wantKind: "limit",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pem := strategy.NewPendingExitManager()
			pem.RegisterExit("exit1", "entry1", tc.stopLevel, tc.limitLevel, registrationBar, "")
			exits := pem.GetExitsForEntry("entry1")
			trade := strategy.Trade{EntryID: "entry1", Direction: tc.direction}
			// barOpen value cannot affect single-side outcomes; pick midpoint.
			barOpen := (tc.barHigh + tc.barLow) / 2
			triggered, price, kind := pem.CheckExitTriggered(
				exits[0], trade, checkBar, barOpen, tc.barHigh, tc.barLow,
			)
			if !triggered {
				t.Fatalf("expected trigger; got triggered=false")
			}
			if kind != tc.wantKind {
				t.Errorf("kind: want %q, got %q", tc.wantKind, kind)
			}
			if price != tc.wantPrice {
				t.Errorf("price: want %v, got %v", tc.wantPrice, price)
			}
		})
	}
}
