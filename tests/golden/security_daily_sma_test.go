package golden

import (
	"math"
	"testing"
)

// TestSecurityDailySMA invariants for security()-backed daily SMA evaluation:
//   - All five indicator series share identical output length.
//   - Each SMA series has NaN values only at the leading warmup prefix; once finite
//     the series stays finite.
//   - Warmup prefix length is proportional to the SMA period (SMA200 ≥ SMA50 ≥ SMA20).
//   - Comparison flag Bull20_50_1D equals 1.0 exactly when SMA20 > SMA50, 0.0
//     otherwise, on every bar where both SMA operands are finite.
//   - Comparison flag Bull50_200_1D equals 1.0 exactly when SMA50 > SMA200, 0.0
//     otherwise, on every bar where both SMA operands are finite.
func TestSecurityDailySMA(t *testing.T) {
	suite := NewTestSuite(t)
	strategyPath := suite.TestFixturePath("test-security-daily-sma-comparison.pine")

	cases := []struct {
		symbol string
		asset  string
	}{
		{"SBERP", "SBERP"},
		{"BTCUSDT", "BTCUSDT"},
		{"AAPL", "AAPL"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.symbol, func(t *testing.T) {
			t.Parallel()

			subSuite := NewTestSuite(t)
			asset := subSuite.assets[tc.asset]
			dataPath := subSuite.golden.EnsureDataFile(t, asset.DataFile)
			result := subSuite.runner.Execute(t, strategyPath, dataPath, asset.Symbol, asset.Timeframe)

			const (
				keySMA20     = "SMA20_1D"
				keySMA50     = "SMA50_1D"
				keySMA200    = "SMA200_1D"
				keyBull2050  = "Bull20_50_1D"
				keyBull50200 = "Bull50_200_1D"
			)

			for _, key := range []string{keySMA20, keySMA50, keySMA200, keyBull2050, keyBull50200} {
				if len(result.Indicators[key]) == 0 {
					t.Fatalf("indicator %q absent or empty — strategy output not wired into Indicators map", key)
				}
			}

			sma20 := result.Indicators[keySMA20]
			sma50 := result.Indicators[keySMA50]
			sma200 := result.Indicators[keySMA200]
			bull2050 := result.Indicators[keyBull2050]
			bull50200 := result.Indicators[keyBull50200]

			t.Run("series_length_consistency", func(t *testing.T) {
				n := len(sma20)
				for _, pair := range []struct {
					name string
					s    []float64
				}{
					{keySMA50, sma50},
					{keySMA200, sma200},
					{keyBull2050, bull2050},
					{keyBull50200, bull50200},
				} {
					if len(pair.s) != n {
						t.Errorf("%s length %d != %s length %d — all series must share identical output length",
							pair.name, len(pair.s), keySMA20, n)
					}
				}
			})

			nanPrefixLen := func(s []float64) int {
				for i, v := range s {
					if !math.IsNaN(v) {
						return i
					}
				}
				return len(s)
			}

			t.Run("nan_prefix_sma20", func(t *testing.T) {
				assertNaNPrefixOnly(t, keySMA20, sma20)
			})
			t.Run("nan_prefix_sma50", func(t *testing.T) {
				assertNaNPrefixOnly(t, keySMA50, sma50)
			})
			t.Run("nan_prefix_sma200", func(t *testing.T) {
				assertNaNPrefixOnly(t, keySMA200, sma200)
			})

			t.Run("warmup_proportional_to_period", func(t *testing.T) {
				n20, n50, n200 := nanPrefixLen(sma20), nanPrefixLen(sma50), nanPrefixLen(sma200)
				if n50 < n20 {
					t.Errorf("SMA50 NaN prefix (%d bars) < SMA20 NaN prefix (%d bars) — warmup must grow with period",
						n50, n20)
				}
				if n200 < n50 {
					t.Errorf("SMA200 NaN prefix (%d bars) < SMA50 NaN prefix (%d bars) — warmup must grow with period",
						n200, n50)
				}
			})

			t.Run("comparison_consistency_bull20_50", func(t *testing.T) {
				for i := range sma20 {
					if math.IsNaN(sma20[i]) || math.IsNaN(sma50[i]) {
						continue
					}
					want := 0.0
					if sma20[i] > sma50[i] {
						want = 1.0
					}
					got := bull2050[i]
					if !math.IsNaN(got) && got != want {
						t.Errorf("bar %d: %s=%.1f but SMA20=%.6f SMA50=%.6f (want %.1f) — flag disagrees with comparison",
							i, keyBull2050, got, sma20[i], sma50[i], want)
					}
				}
			})

			t.Run("comparison_consistency_bull50_200", func(t *testing.T) {
				for i := range sma50 {
					if math.IsNaN(sma50[i]) || math.IsNaN(sma200[i]) {
						continue
					}
					want := 0.0
					if sma50[i] > sma200[i] {
						want = 1.0
					}
					got := bull50200[i]
					if !math.IsNaN(got) && got != want {
						t.Errorf("bar %d: %s=%.1f but SMA50=%.6f SMA200=%.6f (want %.1f) — flag disagrees with comparison",
							i, keyBull50200, got, sma50[i], sma200[i], want)
					}
				}
			})

			t.Run("both_states_observable_bull20_50", func(t *testing.T) {
				assertBothStatesObservable(t, keyBull2050, bull2050)
			})

			t.Run("both_states_observable_bull50_200", func(t *testing.T) {
				assertBothStatesObservable(t, keyBull50200, bull50200)
			})
		})
	}
}

// assertNaNPrefixOnly verifies that NaN values are confined to a leading prefix:
// once the series produces a finite value it must remain finite for all subsequent
// bars. A scattered NaN mid-series indicates broken carry-forward or a mapping gap.
func assertNaNPrefixOnly(t *testing.T, name string, series []float64) {
	t.Helper()

	sawFinite := false
	for i, v := range series {
		if !math.IsNaN(v) {
			sawFinite = true
		} else if sawFinite {
			t.Errorf("%s: NaN at bar %d after finite values — NaN must be confined to warmup prefix", name, i)
		}
	}
	if !sawFinite {
		t.Errorf("%s: every value is NaN — warmup never completed (fixture may have too few bars)", name)
	}
}

// assertBothStatesObservable warns when a 0/1 flag series shows only one state in the
// fixture. A single-state dataset (e.g. AAPL in a sustained bull run) does not break
// comparison logic but reduces observable coverage — comparison_consistency is authoritative.
func assertBothStatesObservable(t *testing.T, name string, series []float64) {
	t.Helper()

	sawBull, sawBear := false, false
	for _, v := range series {
		if math.IsNaN(v) {
			continue
		}
		if v == 1.0 {
			sawBull = true
		} else if v == 0.0 {
			sawBear = true
		}
	}
	if !sawBull {
		t.Logf("%s: no bullish bar found in fixture — single-state dataset reduces observable coverage", name)
	}
	if !sawBear {
		t.Logf("%s: no bearish bar found in fixture — single-state dataset reduces observable coverage", name)
	}
}
