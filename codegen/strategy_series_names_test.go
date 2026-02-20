package codegen

import "testing"

func TestStrategySeriesBindings_NoEmptyFields(t *testing.T) {
	for _, b := range strategySeriesBindings() {
		t.Run(b.varName, func(t *testing.T) {
			if b.varName == "" {
				t.Error("varName must not be empty")
			}
			if b.accessor == "" {
				t.Error("accessor must not be empty")
			}
		})
	}
}

func TestStrategySeriesBindings_NoDuplicateVarNames(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range strategySeriesBindings() {
		if seen[b.varName] {
			t.Errorf("duplicate varName: %q", b.varName)
		}
		seen[b.varName] = true
	}
}

func TestStrategySeriesBindings_NoDuplicateAccessors(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range strategySeriesBindings() {
		if seen[b.accessor] {
			t.Errorf("duplicate accessor: %q", b.accessor)
		}
		seen[b.accessor] = true
	}
}

func TestStrategySeriesBindings_VarNamesHaveStrategyPrefix(t *testing.T) {
	for _, b := range strategySeriesBindings() {
		t.Run(b.varName, func(t *testing.T) {
			if len(b.varName) < 8 || b.varName[:8] != "strategy" {
				t.Errorf("varName %q must start with 'strategy'", b.varName)
			}
		})
	}
}

func TestStrategySeriesBindings_VarNamesHaveSeriesSuffix(t *testing.T) {
	suffix := "Series"
	for _, b := range strategySeriesBindings() {
		t.Run(b.varName, func(t *testing.T) {
			n := len(b.varName)
			if n < len(suffix) || b.varName[n-len(suffix):] != suffix {
				t.Errorf("varName %q must end with 'Series'", b.varName)
			}
		})
	}
}

func TestStrategySeriesBindings_CountMatchesConstants(t *testing.T) {
	/* 20 strategy.* properties mapped to series */
	const expectedCount = 20
	bindings := strategySeriesBindings()
	if len(bindings) != expectedCount {
		t.Errorf("strategySeriesBindings() returned %d bindings, want %d", len(bindings), expectedCount)
	}
}

func TestStrategySeriesBindings_ConstantsMatchBindings(t *testing.T) {
	constants := []string{
		StrategyPositionAvgPriceSeriesName,
		StrategyPositionSizeSeriesName,
		StrategyEquitySeriesName,
		StrategyNetProfitSeriesName,
		StrategyClosedTradesSeriesName,
		StrategyInitialCapitalSeriesName,
		StrategyGrossProfitSeriesName,
		StrategyGrossLossSeriesName,
		StrategyWinTradesSeriesName,
		StrategyLossTradesSeriesName,
		StrategyEvenTradesSeriesName,
		StrategyOpenProfitSeriesName,
		StrategyOpenTradesSeriesName,
		StrategyAvgTradeSeriesName,
		StrategyAvgWinningTradeSeriesName,
		StrategyAvgLosingTradeSeriesName,
		StrategyMaxDrawdownSeriesName,
		StrategyMaxRunupSeriesName,
		StrategyMaxDrawdownPctSeriesName,
		StrategyMaxRunupPctSeriesName,
	}

	inBindings := map[string]bool{}
	for _, b := range strategySeriesBindings() {
		inBindings[b.varName] = true
	}

	for _, c := range constants {
		t.Run(c, func(t *testing.T) {
			if !inBindings[c] {
				t.Errorf("constant %q not present in strategySeriesBindings()", c)
			}
		})
	}
}
