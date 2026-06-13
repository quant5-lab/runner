package testutil

import (
	"math"
	"strings"
	"testing"
)

func baselineTrade() Trade {
	return Trade{
		EntryID:      "Long",
		EntryBar:     10,
		EntryTime:    1_000_000,
		EntryPrice:   100.00,
		EntryComment: "entry",
		ExitBar:      20,
		ExitTime:     2_000_000,
		ExitPrice:    110.00,
		ExitComment:  "exit",
		Size:         1.0,
		Profit:       10.00,
		Direction:    "long",
	}
}

func baselineResult() *StrategyResult {
	tr := baselineTrade()
	return &StrategyResult{
		Trades:      []Trade{tr},
		OpenTrades:  []Trade{},
		TotalTrades: 1,
		Equity:      10000.00,
		NetProfit:   10.00,
		Plots:       map[string][]float64{},
	}
}

func TestFloatWithin(t *testing.T) {
	tests := []struct {
		name      string
		a, b      float64
		tolerance float64
		want      bool
	}{
		{"equal values", 1.0, 1.0, 0.01, true},
		{"within tolerance", 1.005, 1.0, 0.01, true},
		{"at exact tolerance boundary", 0.125, 0.0, 0.125, true},
		{"just outside tolerance", 1.011, 1.0, 0.01, false},
		{"both NaN", math.NaN(), math.NaN(), 0.01, true},
		{"a NaN b number", math.NaN(), 1.0, 0.01, false},
		{"a number b NaN", 1.0, math.NaN(), 0.01, false},
		{"both positive Inf", math.Inf(1), math.Inf(1), 0.01, true},
		{"both negative Inf", math.Inf(-1), math.Inf(-1), 0.01, true},
		{"positive Inf vs negative Inf", math.Inf(1), math.Inf(-1), 0.01, false},
		{"zero tolerance exact match", 5.0, 5.0, 0.0, true},
		{"zero tolerance any difference", 5.0, 5.0 + 1e-15, 0.0, false},
		{"large values within tolerance", 100000.0, 100000.005, 0.01, true},
		{"negative values within tolerance", -50.0, -50.005, 0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := floatWithin(tt.a, tt.b, tt.tolerance); got != tt.want {
				t.Errorf("floatWithin(%v, %v, %v) = %v, want %v",
					tt.a, tt.b, tt.tolerance, got, tt.want)
			}
		})
	}
}

func TestCompareResults_EqualResults(t *testing.T) {
	if err := CompareResults(baselineResult(), baselineResult()); err != nil {
		t.Errorf("expected no error for identical results, got: %v", err)
	}
}

func TestCompareResults_EmptyResults(t *testing.T) {
	empty := &StrategyResult{Plots: map[string][]float64{}}
	if err := CompareResults(empty, empty); err != nil {
		t.Errorf("expected no error for empty results, got: %v", err)
	}
}

func TestCompareResults_NilResults(t *testing.T) {
	nilResult := &StrategyResult{}
	if err := CompareResults(nilResult, nilResult); err != nil {
		t.Errorf("expected no error for nil-field results, got: %v", err)
	}
}

func TestCompareResults_TradeCountMismatch(t *testing.T) {
	expected := baselineResult()
	actual := baselineResult()
	actual.Trades = append(actual.Trades, baselineTrade())

	err := CompareResults(expected, actual)
	if err == nil {
		t.Fatal("expected error for trade count mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "trades count") {
		t.Errorf("expected error to mention 'trades count', got %q", err.Error())
	}
}

func TestCompareResults_TradeFieldMismatch(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Trade)
		wantErr string
	}{
		{"entryId mismatch", func(tr *Trade) { tr.EntryID = "Short" }, "entryId"},
		{"entryBar mismatch", func(tr *Trade) { tr.EntryBar = 99 }, "entryBar"},
		{"entryTime mismatch", func(tr *Trade) { tr.EntryTime = 9_999_999 }, "entryTime"},
		{"entryPrice within tolerance passes", func(tr *Trade) { tr.EntryPrice += 100.0 * priceRelEps * 0.5 }, ""},
		{"entryPrice outside tolerance fails", func(tr *Trade) { tr.EntryPrice += 100.0 * priceRelEps * 2 }, "entryPrice"},
		{"entryComment mismatch", func(tr *Trade) { tr.EntryComment = "other" }, "entryComment"},
		{"exitBar mismatch", func(tr *Trade) { tr.ExitBar = 99 }, "exitBar"},
		{"exitTime mismatch", func(tr *Trade) { tr.ExitTime = 9_999_999 }, "exitTime"},
		{"exitPrice within tolerance passes", func(tr *Trade) { tr.ExitPrice += 110.0 * priceRelEps * 0.5 }, ""},
		{"exitPrice outside tolerance fails", func(tr *Trade) { tr.ExitPrice += 110.0 * priceRelEps * 2 }, "exitPrice"},
		{"exitComment mismatch", func(tr *Trade) { tr.ExitComment = "other" }, "exitComment"},
		{"size within tolerance passes", func(tr *Trade) { tr.Size += 1.0 * financialRelEps * 0.5 }, ""},
		{"size outside tolerance fails", func(tr *Trade) { tr.Size += 1.0 * financialRelEps * 2 }, "size"},
		{"profit within tolerance passes", func(tr *Trade) { tr.Profit += 10.0 * financialRelEps * 0.5 }, ""},
		{"profit outside tolerance fails", func(tr *Trade) { tr.Profit += 10.0 * financialRelEps * 2 }, "profit"},
		{"direction mismatch", func(tr *Trade) { tr.Direction = "short" }, "direction"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := baselineResult()
			actual := baselineResult()
			tt.mutate(&actual.Trades[0])

			err := CompareResults(expected, actual)
			assertErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestCompareResults_OpenTradesMismatch(t *testing.T) {
	makeResultWithOpenTrade := func() *StrategyResult {
		r := baselineResult()
		r.OpenTrades = []Trade{baselineTrade()}
		return r
	}

	tests := []struct {
		name    string
		mutate  func(*StrategyResult)
		wantErr string
	}{
		{
			"count mismatch labels openTrades",
			func(r *StrategyResult) { r.OpenTrades = append(r.OpenTrades, baselineTrade()) },
			"openTrades count",
		},
		{
			"field mismatch labels openTrades",
			func(r *StrategyResult) { r.OpenTrades[0].EntryBar = 99 },
			"openTrades[0].entryBar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := makeResultWithOpenTrade()
			actual := makeResultWithOpenTrade()
			tt.mutate(actual)

			err := CompareResults(expected, actual)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCompareResults_ScalarMismatch(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*StrategyResult)
		wantErr string
	}{
		{"totalTrades mismatch", func(r *StrategyResult) { r.TotalTrades = 99 }, "totalTrades"},
		{"equity within tolerance passes", func(r *StrategyResult) { r.Equity += 10000.0 * financialRelEps * 0.5 }, ""},
		{"equity outside tolerance fails", func(r *StrategyResult) { r.Equity += 10000.0 * financialRelEps * 2 }, "equity"},
		{"netProfit within tolerance passes", func(r *StrategyResult) { r.NetProfit += 10.0 * financialRelEps * 0.5 }, ""},
		{"netProfit outside tolerance fails", func(r *StrategyResult) { r.NetProfit += 10.0 * financialRelEps * 2 }, "netProfit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := baselineResult()
			actual := baselineResult()
			tt.mutate(actual)

			assertErrorContains(t, CompareResults(expected, actual), tt.wantErr)
		})
	}
}

func TestComparePlots(t *testing.T) {
	tests := []struct {
		name     string
		expected map[string][]float64
		actual   map[string][]float64
		wantErr  string
	}{
		{
			"equal plots",
			map[string][]float64{"rsi": {50.0, 60.0}},
			map[string][]float64{"rsi": {50.0, 60.0}},
			"",
		},
		{
			"nil vs nil",
			nil,
			nil,
			"",
		},
		{
			"nil vs empty map",
			nil,
			map[string][]float64{},
			"",
		},
		{
			"series count mismatch",
			map[string][]float64{"rsi": {50.0}},
			map[string][]float64{"rsi": {50.0}, "macd": {1.0}},
			"plots:",
		},
		{
			"series missing in actual",
			map[string][]float64{"rsi": {50.0}},
			map[string][]float64{"macd": {50.0}},
			`"rsi" missing`,
		},
		{
			"series length mismatch",
			map[string][]float64{"rsi": {50.0, 60.0}},
			map[string][]float64{"rsi": {50.0}},
			`plots["rsi"]`,
		},
		{
			"value within plot tolerance passes",
			map[string][]float64{"rsi": {50.0}},
			map[string][]float64{"rsi": {50.0 + plotAbsEps*0.5}},
			"",
		},
		{
			"value outside plot tolerance fails",
			map[string][]float64{"rsi": {50.0}},
			map[string][]float64{"rsi": {50.0 + plotAbsEps + 1e-7}},
			`plots["rsi"][0]`,
		},
		{
			"both NaN values pass",
			map[string][]float64{"rsi": {math.NaN()}},
			map[string][]float64{"rsi": {math.NaN()}},
			"",
		},
		{
			"NaN vs number fails",
			map[string][]float64{"rsi": {math.NaN()}},
			map[string][]float64{"rsi": {50.0}},
			`plots["rsi"][0]`,
		},
		{
			"multiple series all equal",
			map[string][]float64{"rsi": {50.0}, "macd": {1.5}},
			map[string][]float64{"rsi": {50.0}, "macd": {1.5}},
			"",
		},
		{
			"second series value mismatch",
			map[string][]float64{"macd": {1.5}, "rsi": {50.0}},
			map[string][]float64{"macd": {1.5}, "rsi": {60.0}},
			`plots["rsi"][0]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrorContains(t, comparePlots(tt.expected, tt.actual), tt.wantErr)
		})
	}
}

func TestResultsEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b *StrategyResult
		want bool
	}{
		{"identical results", baselineResult(), baselineResult(), true},
		{"empty results", &StrategyResult{}, &StrategyResult{}, true},
		{
			"different trade bar",
			baselineResult(),
			func() *StrategyResult {
				r := baselineResult()
				r.Trades[0].EntryBar = 99
				return r
			}(),
			false,
		},
		{
			"different equity",
			baselineResult(),
			func() *StrategyResult {
				r := baselineResult()
				r.Equity = 99999.00
				return r
			}(),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResultsEqual(tt.a, tt.b); got != tt.want {
				t.Errorf("ResultsEqual = %v, want %v", got, tt.want)
			}
		})
	}
}

func assertErrorContains(t *testing.T, err error, wantSubstr string) {
	t.Helper()
	if wantSubstr == "" {
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		return
	}
	if err == nil {
		t.Errorf("expected error containing %q, got nil", wantSubstr)
		return
	}
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("expected error containing %q, got %q", wantSubstr, err.Error())
	}
}
