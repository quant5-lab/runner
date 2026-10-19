package market

import "testing"

// TestSessionOpenMinute_CuratedTablePinnedToTVReference locks each known
// exchange's regular-session open minute to its declared TradingView reference
// open. DeriveSessionAnchor (period_anchor_resolver.go) routes every known
// exchange to this curated value with NO observation fallback, so a silent drift
// in the schedule table would mis-tile multi-hour period grids on security()
// series without any other guard catching it. This ratchet fails loudly on such
// a drift. Known exchanges have no observation fallback, so the curated table is
// the sole source of truth for their session-open anchor; this ratchet pins it
// to the reference open so a bad table entry cannot pass silently.
func TestSessionOpenMinute_CuratedTablePinnedToTVReference(t *testing.T) {
	cases := []struct {
		name     string
		exchange Exchange
		wantMin  int
		tvOpen   string
	}{
		{"MOEX", ExchangeMOEX, 7 * 60, "07:00 MSK weekday reference open"},
		{"NYSE", ExchangeNYSE, 9*60 + 30, "09:30 ET regular open"},
		{"BINANCE", ExchangeBinance, 0, "00:00 (24/7 crypto midnight anchor)"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			got := sessionOpenMinuteFor(c.exchange, ReferenceSessionRegular)
			if got != c.wantMin {
				t.Errorf("sessionOpenMinuteFor(%s, Regular) = %d min; want %d (%s). "+
					"The curated schedule table drifted from the TV reference open; "+
					"since known exchanges have no observation fallback, a wrong value "+
					"silently mis-tiles multi-hour security() grids.",
					c.exchange, got, c.wantMin, c.tvOpen)
			}
		})
	}
}
