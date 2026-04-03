package codegen

import (
	"testing"
)

/* TestHandlerBareNameParity validates all handlers accept both namespaced and bare function names
 *
 * PineScript v5 uses namespaced form (ta.sma), v4 uses bare form (sma)
 *
 * All handlers must support both for backward compatibility
 */
func TestHandlerBareNameParity(t *testing.T) {
	/* All 21 handler-registered functions */
	handlers := []struct {
		name       string
		handler    TAFunctionHandler
		namespaced string
		bare       string
	}{
		{"SMA", &SMAHandler{}, "ta.sma", "sma"},
		{"EMA", &EMAHandler{}, "ta.ema", "ema"},
		{"STDEV", &STDEVHandler{}, "ta.stdev", "stdev"},
		{"WMA", &WMAHandler{}, "ta.wma", "wma"},
		{"DEV", &DEVHandler{}, "ta.dev", "dev"},
		{"ATR", &ATRHandler{}, "ta.atr", "atr"},
		{"RMA", &RMAHandler{}, "ta.rma", "rma"},
		{"RSI", &RSIHandler{}, "ta.rsi", "rsi"},
		{"Change", &ChangeHandler{}, "ta.change", "change"},
		{"PivotHigh", &PivotHighHandler{}, "ta.pivothigh", "pivothigh"},
		{"PivotLow", &PivotLowHandler{}, "ta.pivotlow", "pivotlow"},
		{"Crossover", &CrossoverHandler{}, "ta.crossover", "crossover"},
		{"Crossunder", &CrossunderHandler{}, "ta.crossunder", "crossunder"},
		{"Fixnan", &FixnanHandler{}, "fixnan", "fixnan"},
		{"Sum", &SumHandler{}, "sum", "sum"},
		{"Valuewhen", &ValuewhenHandler{}, "ta.valuewhen", "valuewhen"},
		{"Highest", &HighestHandler{}, "ta.highest", "highest"},
		{"Lowest", &LowestHandler{}, "ta.lowest", "lowest"},
		{"Linreg", &LinregHandler{}, "ta.linreg", "linreg"},
		{"BarsSince", &BarsSinceHandler{}, "ta.barssince", "barssince"},
		{"MFI", &MFIHandler{}, "ta.mfi", "mfi"},
		{"Rising", &RisingHandler{}, "ta.rising", "rising"},
		{"Falling", &FallingHandler{}, "ta.falling", "falling"},
		{"Cross", &CrossHandler{}, "ta.cross", "cross"},
		{"Highestbars", &HighestbarsHandler{}, "ta.highestbars", "highestbars"},
		{"Lowestbars", &LowestbarsHandler{}, "ta.lowestbars", "lowestbars"},
		{"Mom", &MomHandler{}, "ta.mom", "mom"},
		{"Roc", &RocHandler{}, "ta.roc", "roc"},
		{"Cmo", &CmoHandler{}, "ta.cmo", "cmo"},
		{"Wpr", &WprHandler{}, "ta.wpr", "wpr"},
	}

	for _, h := range handlers {
		t.Run(h.name+" namespaced", func(t *testing.T) {
			if !h.handler.CanHandle(h.namespaced) {
				t.Errorf("%s handler must accept namespaced form %q", h.name, h.namespaced)
			}
		})

		t.Run(h.name+" bare", func(t *testing.T) {
			if !h.handler.CanHandle(h.bare) {
				t.Errorf("%s handler must accept bare form %q (v4 compatibility)", h.name, h.bare)
			}
		})

		/* Negative test: should reject other function names */
		t.Run(h.name+" rejects other functions", func(t *testing.T) {
			otherFunctions := []string{"ta.sma", "ta.ema", "ta.rsi", "ta.unknown"}
			for _, other := range otherFunctions {
				if other != h.namespaced && h.handler.CanHandle(other) {
					t.Errorf("%s handler incorrectly accepts %q", h.name, other)
				}
			}
		})
	}
}

/* TestHandlerBareNameParity_EdgeCases validates special cases in bare name handling
 *
 * Edge cases:
 * - Functions without "ta." namespace (fixnan, sum)
 * - Functions with identical bare/namespaced forms
 * - Math namespace functions (sum can be "sum" or "math.sum")
 */
func TestHandlerBareNameParity_EdgeCases(t *testing.T) {
	t.Run("Fixnan has no ta. prefix", func(t *testing.T) {
		handler := &FixnanHandler{}
		if !handler.CanHandle("fixnan") {
			t.Error("FixnanHandler must accept 'fixnan' (no ta. prefix)")
		}
		if handler.CanHandle("ta.fixnan") {
			t.Error("FixnanHandler should not accept 'ta.fixnan' (invalid namespace)")
		}
	})

	t.Run("Sum accepts both namespaces", func(t *testing.T) {
		handler := &SumHandler{}
		if !handler.CanHandle("sum") {
			t.Error("SumHandler must accept 'sum'")
		}
		if !handler.CanHandle("math.sum") {
			t.Error("SumHandler must accept 'math.sum'")
		}
		if handler.CanHandle("ta.sum") {
			t.Error("SumHandler should not accept 'ta.sum' (wrong namespace)")
		}
	})

	t.Run("All handlers reject empty string", func(t *testing.T) {
		handlers := []TAFunctionHandler{
			&SMAHandler{}, &EMAHandler{}, &RSIHandler{}, &MFIHandler{},
			&PivotHighHandler{}, &CrossoverHandler{},
		}

		for _, h := range handlers {
			if h.CanHandle("") {
				t.Errorf("Handler %T incorrectly accepts empty string", h)
			}
		}
	})
}
