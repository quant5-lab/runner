package regression

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/market"
	"github.com/quant5-lab/runner/runtime/strategy"
	goldenutil "github.com/quant5-lab/runner/tests/golden/testutil"
	tvref "github.com/quant5-lab/runner/tests/regression/tv_reference"
)

func TestCycle20260612Close_StructuralDefectsStayClosed(t *testing.T) {
	t.Run("commission does not change default entry quantity", func(t *testing.T) {
		withoutCommission := defaultEntryQty(t, strategy.CommissionPercent, 0)
		withCommission := defaultEntryQty(t, strategy.CommissionPercent, 50)
		if withCommission != withoutCommission {
			t.Fatalf("commission changed default entry qty: no commission %.8f, with commission %.8f", withoutCommission, withCommission)
		}
	})

	t.Run("financial tolerance derives from measured residual with headroom", func(t *testing.T) {
		want := goldenutil.MeasuredSizeResidualBound * goldenutil.FinancialToleranceSafetyMultiple
		if goldenutil.FinancialRelEps != want {
			t.Fatalf("FinancialRelEps = %.10f, want %.10f", goldenutil.FinancialRelEps, want)
		}
		if goldenutil.FinancialToleranceSafetyMultiple < 5 {
			t.Fatalf("FinancialToleranceSafetyMultiple = %.2f, want at least 5", goldenutil.FinancialToleranceSafetyMultiple)
		}
	})

	t.Run("tv size residual is observable", func(t *testing.T) {
		residual := tvref.SizeResidual(1.25, 1.00)
		if residual != 0.25 {
			t.Fatalf("SizeResidual(1.25, 1.00) = %.8f, want 0.25", residual)
		}
	})

	t.Run("metadata qty step overrides exchange fallback", func(t *testing.T) {
		profile := market.ResolveProfileWithMetadata("BTCUSDT", market.SourceMetadata{
			Exchange: "BINANCE",
			QtyStep:  0.125,
		})
		if profile.QtyStep != 0.125 {
			t.Fatalf("profile qty step = %.8f, want metadata qty step 0.12500000", profile.QtyStep)
		}
	})
}

func defaultEntryQty(t *testing.T, commissionType string, commissionValue float64) float64 {
	t.Helper()
	s := strategy.NewStrategy()
	s.Call("qty", 10000)
	s.SetDefaultQty(1000, strategy.QtyTypeCash)
	s.SetCommission(commissionValue, commissionType)
	return s.DefaultEntryQty(50)
}
