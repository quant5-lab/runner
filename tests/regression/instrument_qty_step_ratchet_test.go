package regression

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	goldenutil "github.com/quant5-lab/runner/tests/golden/testutil"
)

func TestQtyLotFloor_StockExchangesProduceWholeLotSizes(t *testing.T) {
	root := projectRootFromCwd()
	runner := goldenutil.NewStrategyRunner(t)

	for _, tc := range tvAlignmentCases() {
		tc := tc
		if isStockSymbol(tc.Symbol) {
			t.Run(tc.Name, func(t *testing.T) {
				result := runner.Execute(
					t,
					filepath.Join(root, "strategies", tc.Strategy),
					filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data),
					tc.Symbol,
					tc.Timeframe,
				)
				assertAllWholeLots(t, result.Trades)
				assertAllWholeLots(t, result.OpenTrades)
			})
		}
	}
}

func TestQtyLotFloor_BinanceUsesFixtureQtyStep(t *testing.T) {
	root := projectRootFromCwd()
	runner := goldenutil.NewStrategyRunner(t)

	covered := 0
	for _, tc := range tvAlignmentCases() {
		tc := tc
		if !isStockSymbol(tc.Symbol) {
			t.Run(tc.Name, func(t *testing.T) {
				dataPath := filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data)
				qtyStep := fixtureQtyStep(t, dataPath)
				if qtyStep <= 0 {
					t.Fatalf("%s has no positive qtyStep metadata", tc.Data)
				}

				result := runner.Execute(
					t,
					filepath.Join(root, "strategies", tc.Strategy),
					dataPath,
					tc.Symbol,
					tc.Timeframe,
				)

				all := append(result.Trades, result.OpenTrades...)
				if len(all) == 0 {
					t.Fatalf("no trades produced for %s: cannot verify qty step", tc.Symbol)
				}

				for _, tr := range all {
					if !multipleOfStep(tr.Size, qtyStep) {
						t.Errorf("%s size %.16f is not a multiple of qtyStep %.8f", tc.Symbol, tr.Size, qtyStep)
					}
				}
			})
			covered++
		}
	}
	if covered == 0 {
		t.Skip("no non-stock TV alignment case present")
	}
}

func isStockSymbol(symbol string) bool {
	switch symbol {
	case "SBERP", "SBER", "CNRU", "AAPL", "NVDA", "TSLA":
		return true
	default:
		return false
	}
}

func assertAllWholeLots(t *testing.T, trades []goldenutil.Trade) {
	t.Helper()
	for i, tr := range trades {
		if tr.Size != math.Floor(tr.Size) {
			t.Errorf("trades[%d].size = %.10f: not a whole lot (stock exchange requires integer position size)", i, tr.Size)
		}
	}
}

func multipleOfStep(value, step float64) bool {
	if step <= 0 {
		return value == math.Floor(value)
	}
	nearest := math.Round(value/step) * step
	return math.Abs(value-nearest) < 1e-9
}

func fixtureQtyStep(t *testing.T, path string) float64 {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	var envelope struct {
		QtyStep float64 `json:"qtyStep"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("parse fixture %s: %v", path, err)
	}
	return envelope.QtyStep
}
