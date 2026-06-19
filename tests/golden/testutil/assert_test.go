package testutil

import (
	"math"
	"testing"
)

func TestClosedEquityMismatch(t *testing.T) {
	tests := []struct {
		name           string
		initialCapital float64
		netProfit      float64
		equity         float64
		wantMismatch   bool
	}{
		{"default capital zero profit", 10000, 0, 10000, false},
		{"default capital with profit", 10000, 500, 10500, false},
		{"small capital 1000 at loss", 1000, -20, 980, false},
		{"large capital 100000 with profit", 100000, 2500, 102500, false},
		{"net loss reduces equity correctly", 10000, -300, 9700, false},
		{"zero initial capital", 0, 500, 500, false},
		{"both zero", 0, 0, 0, false},

		{"drift within tolerance passes", 10000, 500, 10500 + 10500*financialRelEps*0.4, false},
		{"drift beyond tolerance fails", 10000, 500, 10500 + 10500*financialRelEps*2, true},

		{"10x equity mismatch capital=1000", 1000, -20, 9980, true},
		{"equity zero when nonzero expected", 10000, 500, 0, true},
		{"equity far above expected", 10000, 0, 50000, true},
		{"equity far below expected", 10000, 0, 1000, true},

		{"NaN equity detected", 10000, 0, math.NaN(), true},
		{"positive Inf equity detected", 10000, 0, math.Inf(1), true},
		{"negative Inf equity detected", 10000, 0, math.Inf(-1), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := closedEquityMismatch(tc.equity, tc.initialCapital, tc.netProfit)
			gotMismatch := msg != ""
			if gotMismatch != tc.wantMismatch {
				t.Errorf("closedEquityMismatch(equity=%.4f, initial=%.0f, profit=%.4f): wantMismatch=%v, got %v (msg=%q)",
					tc.equity, tc.initialCapital, tc.netProfit, tc.wantMismatch, gotMismatch, msg)
			}
		})
	}
}

func TestTradeUnrealizedPnL(t *testing.T) {
	tests := []struct {
		name       string
		direction  string
		entryPrice float64
		size       float64
		markClose  float64
		want       float64
	}{
		{"long in-profit", "long", 100, 10, 120, 200},
		{"long at-loss", "long", 100, 10, 80, -200},
		{"long break-even", "long", 100, 10, 100, 0},

		{"short in-profit", "short", 100, 10, 80, 200},
		{"short at-loss", "short", 100, 10, 120, -200},
		{"short break-even", "short", 100, 10, 100, 0},

		{"fractional crypto lot long", "long", 65000, 0.001, 65100, 0.1},
		{"large price short in-profit", "short", 65000, 0.5, 64000, 500},
		{"unit size long", "long", 200, 1, 250, 50},
		{"large size long", "long", 50, 1000, 55, 5000},

		{"zero size long", "long", 100, 0, 120, 0},
		{"zero size short", "short", 100, 0, 80, 0},

		{"long marked to zero", "long", 100, 10, 0, -1000},
		{"short marked to zero", "short", 100, 10, 0, 1000},

		{"long entered at zero price", "long", 0, 5, 50, 250},

		// Only the exact string "short" flips the sign; all other values treat the position as long.
		{"empty direction treated as long", "", 100, 10, 120, 200},
		{"unknown direction treated as long", "LONG", 100, 10, 120, 200},
		{"mixed-case direction treated as long", "Short", 100, 5, 80, -100},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trade := Trade{Direction: tc.direction, EntryPrice: tc.entryPrice, Size: tc.size}
			got := tradeUnrealizedPnL(trade, tc.markClose)
			if !matchFinancial(got, tc.want) {
				t.Errorf("tradeUnrealizedPnL(%s entry=%.4f size=%.4f mark=%.4f): want %.4f, got %.4f",
					tc.direction, tc.entryPrice, tc.size, tc.markClose, tc.want, got)
			}
		})
	}
}

func TestOpenEquityMismatch(t *testing.T) {
	const (
		initial   = 10000.0
		markClose = 120.0
	)

	const (
		longEntry  = 100.0
		longSize   = 10.0
		shortEntry = 150.0
		shortSize  = 5.0

		longUnrealized  = (markClose - longEntry) * longSize   // 200
		shortUnrealized = (shortEntry - markClose) * shortSize // 150
	)

	longTrade := Trade{Direction: "long", EntryPrice: longEntry, Size: longSize}
	shortTrade := Trade{Direction: "short", EntryPrice: shortEntry, Size: shortSize}

	tests := []struct {
		name           string
		equity         float64
		initialCapital float64
		netProfit      float64
		openTrades     []Trade
		markClose      float64
		wantMismatch   bool
	}{
		{
			"single long correct equity passes",
			initial + 0 + longUnrealized, initial, 0,
			[]Trade{longTrade}, markClose, false,
		},
		{
			"single long missing unrealized detected",
			initial, initial, 0,
			[]Trade{longTrade}, markClose, true,
		},
		{
			"single long wrong sign (negative unrealized) detected",
			initial - longUnrealized, initial, 0,
			[]Trade{longTrade}, markClose, true,
		},

		{
			"single short correct equity passes",
			initial + 0 + shortUnrealized, initial, 0,
			[]Trade{shortTrade}, markClose, false,
		},
		{
			"single short wrong side (long formula) detected",
			initial + (markClose-shortEntry)*shortSize, initial, 0,
			[]Trade{shortTrade}, markClose, true,
		},

		{
			"long+short mix correct equity passes",
			initial + 500 + longUnrealized + shortUnrealized, initial, 500,
			[]Trade{longTrade, shortTrade}, markClose, false,
		},
		{
			"long+short mix missing unrealized detected",
			initial + 500, initial, 500,
			[]Trade{longTrade, shortTrade}, markClose, true,
		},

		{
			"negative closed PL with open position correct",
			initial + (-300) + longUnrealized, initial, -300,
			[]Trade{longTrade}, markClose, false,
		},
		{
			"negative closed PL with open position wrong equity detected",
			initial + longUnrealized, initial, -300,
			[]Trade{longTrade}, markClose, true,
		},

		{
			"empty open trades zero unrealized passes when equity matches",
			initial + 500, initial, 500,
			[]Trade{}, markClose, false,
		},
		{
			"empty open trades wrong equity detected",
			initial + 9999, initial, 500,
			[]Trade{}, markClose, true,
		},

		{
			"drift within financial tolerance passes",
			(initial + longUnrealized) * (1 + financialRelEps*0.4), initial, 0,
			[]Trade{longTrade}, markClose, false,
		},
		{
			"drift beyond financial tolerance fails",
			(initial + longUnrealized) * (1 + financialRelEps*2), initial, 0,
			[]Trade{longTrade}, markClose, true,
		},

		{"wildly wrong equity detected", 50000, initial, 0, []Trade{longTrade}, markClose, true},
		{"zero equity detected", 0, initial, 0, []Trade{longTrade}, markClose, true},
		{"NaN equity detected", math.NaN(), initial, 0, []Trade{longTrade}, markClose, true},
		{"positive Inf equity detected", math.Inf(1), initial, 0, []Trade{longTrade}, markClose, true},
		{"negative Inf equity detected", math.Inf(-1), initial, 0, []Trade{longTrade}, markClose, true},

		{
			"10x scale mismatch small capital detected",
			9980, 1000, -20,
			[]Trade{}, markClose, true,
		},
		{
			"correct small capital equity passes",
			980, 1000, -20,
			[]Trade{}, markClose, false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := openEquityMismatch(tc.equity, tc.initialCapital, tc.netProfit, tc.openTrades, tc.markClose)
			gotMismatch := msg != ""
			if gotMismatch != tc.wantMismatch {
				t.Errorf("openEquityMismatch: wantMismatch=%v got %v (msg=%q)",
					tc.wantMismatch, gotMismatch, msg)
			}
		})
	}
}

func TestOpenEquityCheckRequired(t *testing.T) {
	tests := []struct {
		name       string
		fromRunner bool
		markClose  float64
		want       bool
	}{
		{"runner, positive mark", true, 120.0, true},
		{"runner, zero mark", true, 0.0, true},
		{"runner, negative mark", true, -1.0, true},
		{"runner, very small positive mark", true, 1e-15, true},

		{"synthetic, positive mark", false, 120.0, true},
		{"synthetic, negative mark", false, -1.0, true},
		{"synthetic, very small positive mark", false, 1e-15, true},

		{"synthetic, zero mark", false, 0.0, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := openEquityCheckRequired(tc.fromRunner, tc.markClose)
			if got != tc.want {
				t.Errorf("openEquityCheckRequired(fromRunner=%v, markClose=%v): want %v, got %v",
					tc.fromRunner, tc.markClose, tc.want, got)
			}
		})
	}
}

func TestValidateEquityConsistency_ClosedStrategy(t *testing.T) {
	tests := []struct {
		name           string
		initialCapital float64
		netProfit      float64
		equity         float64
	}{
		{"zero profit at default capital", 10000, 0, 10000},
		{"profitable run at default capital", 10000, 1500, 11500},
		{"losing run at default capital", 10000, -300, 9700},
		{"small capital base", 1000, -20, 980},
		{"large capital base", 100000, 2500, 102500},
		{"zero initial capital with profit", 0, 500, 500},
		{"equity within financial tolerance", 10000, 500, 10500 + 10500*financialRelEps*0.4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ValidateEquityConsistency(t, &StrategyResult{
				InitialCapital: tc.initialCapital,
				NetProfit:      tc.netProfit,
				Equity:         tc.equity,
			})
		})
	}
}

func TestValidateEquityConsistency_OpenPositionEquityVerified(t *testing.T) {
	const (
		initialCapital = 10000.0
		closedPL       = 200.0
		markClose      = 120.0
	)

	t.Run("single long position", func(t *testing.T) {
		open := Trade{EntryID: "E1", Direction: "long", EntryPrice: 100, Size: 10}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      closedPL,
			Equity:         initialCapital + closedPL + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			OpenTrades:     []Trade{open},
		})
	})

	t.Run("single short position", func(t *testing.T) {
		open := Trade{EntryID: "E2", Direction: "short", EntryPrice: 150, Size: 3}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      closedPL,
			Equity:         initialCapital + closedPL + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			OpenTrades:     []Trade{open},
		})
	})

	t.Run("long and short mix", func(t *testing.T) {
		openTrades := []Trade{
			{EntryID: "E1", Direction: "long", EntryPrice: 100, Size: 10},
			{EntryID: "E2", Direction: "short", EntryPrice: 150, Size: 3},
		}
		unrealized := tradeUnrealizedPnL(openTrades[0], markClose) + tradeUnrealizedPnL(openTrades[1], markClose)
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      closedPL,
			Equity:         initialCapital + closedPL + unrealized,
			MarkClose:      markClose,
			OpenTrades:     openTrades,
		})
	})

	t.Run("negative closed PL", func(t *testing.T) {
		const lossedPL = -500.0
		open := Trade{EntryID: "E1", Direction: "long", EntryPrice: 100, Size: 5}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      lossedPL,
			Equity:         initialCapital + lossedPL + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			OpenTrades:     []Trade{open},
		})
	})
}

func TestValidateEquityConsistency_RunnerProvenanceAsserts(t *testing.T) {
	t.Run("runner result with zero mark close passes when equity is mark-consistent", func(t *testing.T) {
		const (
			initialCapital = 10000.0
			netProfit      = 0.0
			markClose      = 0.0
		)
		open := Trade{Direction: "long", EntryPrice: 100, Size: 10}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      netProfit,
			Equity:         initialCapital + netProfit + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			FromRunner:     true,
			OpenTrades:     []Trade{open},
		})
	})

	t.Run("runner result with positive mark close passes when equity is mark-consistent", func(t *testing.T) {
		const (
			initialCapital = 10000.0
			netProfit      = 500.0
			markClose      = 120.0
		)
		open := Trade{Direction: "short", EntryPrice: 150, Size: 5}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      netProfit,
			Equity:         initialCapital + netProfit + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			FromRunner:     true,
			OpenTrades:     []Trade{open},
		})
	})

	t.Run("runner result with negative mark close passes when equity is mark-consistent", func(t *testing.T) {
		const (
			initialCapital = 5000.0
			netProfit      = -100.0
			markClose      = -1.0
		)
		open := Trade{Direction: "short", EntryPrice: 0, Size: 3}
		ValidateEquityConsistency(t, &StrategyResult{
			InitialCapital: initialCapital,
			NetProfit:      netProfit,
			Equity:         initialCapital + netProfit + tradeUnrealizedPnL(open, markClose),
			MarkClose:      markClose,
			FromRunner:     true,
			OpenTrades:     []Trade{open},
		})
	})
}

func TestValidateEquityConsistency_NoMarkPriceDegradesGracefully(t *testing.T) {
	ValidateEquityConsistency(t, &StrategyResult{
		InitialCapital: 10000,
		NetProfit:      0,
		Equity:         50000, // wildly wrong but no mark price to assert against
		MarkClose:      0,
		OpenTrades:     []Trade{{EntryID: "open", Direction: "long", EntryPrice: 100}},
	})
}
