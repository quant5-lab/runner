package regression

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/runtime/market"
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

// A per-timeframe fixture written without qtyStep silently quantises positions
// differently from its sibling, masking lot-size errors that only surface when
// a strategy switches to that timeframe.
func TestFixtureMetadata_SameSymbolSameQtyStep(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")

	type fixtureEntry struct {
		dataFile string
		qtyStep  float64
	}
	bySymbol := make(map[string][]fixtureEntry)
	for _, tc := range tvAlignmentCases() {
		path := filepath.Join(fixtureDir, tc.Data)
		qs := fixtureQtyStep(t, path)
		bySymbol[tc.Symbol] = append(bySymbol[tc.Symbol], fixtureEntry{tc.Data, qs})
	}

	for symbol, entries := range bySymbol {
		if len(entries) < 2 {
			continue
		}
		first := entries[0]
		for _, e := range entries[1:] {
			if e.dataFile == first.dataFile {
				continue
			}
			if e.qtyStep != first.qtyStep {
				t.Errorf(
					"symbol %s: qtyStep mismatch across fixtures: %s=%.8g vs %s=%.8g"+
						" — qtyStep is instrument metadata, not bar-resolution metadata;"+
						" all timeframe fixtures for the same symbol must agree",
					symbol, first.dataFile, first.qtyStep, e.dataFile, e.qtyStep,
				)
			}
		}
	}
}

// step <= 0 intentionally degrades to a whole-lot check so Binance's offline
// zero sentinel never silently accepts fractional quantities.
func TestMultipleOfStep_BoundaryValues(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		step  float64
		want  bool
	}{
		{"exact multiple", 3e-05, 1e-05, true},
		{"zero is always a multiple of any step", 0.0, 1e-05, true},
		{"just inside epsilon band", 1e-05 + 9e-10, 1e-05, true},
		{"just outside epsilon band", 1e-05 + 1.1e-9, 1e-05, false},
		{"large value exact", 1.23456, 0.00001, true},
		{"large value with fractional remainder", 1.234565, 0.00001, false},

		{"zero step: whole value is valid", 1.0, 0, true},
		{"zero step: fractional value is invalid", 1.5, 0, false},
		{"negative step: treated as zero, whole value", 2.0, -1, true},
		{"negative step: treated as zero, fractional value", 2.5, -1, false},

		{"step equals value: always a multiple", 0.001, 0.001, true},
		{"half step is not a multiple", 0.0005, 0.001, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := multipleOfStep(tc.value, tc.step); got != tc.want {
				t.Errorf("multipleOfStep(%.16g, %.16g) = %v, want %v",
					tc.value, tc.step, got, tc.want)
			}
		})
	}
}

// A divergence between isStockSymbol and the exchange resolution silently routes
// a Binance symbol through the whole-lot path, bypassing the fixture-level
// qtyStep guard entirely.
func TestIsStockSymbol_AgreesWithExchangeResolution(t *testing.T) {
	for _, tc := range tvAlignmentCases() {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			exchange := market.ResolveExchange(tc.Symbol)
			defaultStep := market.InstrumentQtyStep(exchange)
			exchangeIsWholeLot := defaultStep >= 1

			if isStockSymbol(tc.Symbol) != exchangeIsWholeLot {
				t.Errorf(
					"symbol %q: isStockSymbol=%v but InstrumentQtyStep(%q)=%.8g"+
						" (exchangeIsWholeLot=%v) — isStockSymbol must agree with exchange"+
						" resolution so the qtyStep fixture-metadata guard is never bypassed",
					tc.Symbol, isStockSymbol(tc.Symbol),
					exchange, defaultStep, exchangeIsWholeLot,
				)
			}
		})
	}
}
