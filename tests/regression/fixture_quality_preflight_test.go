package regression

import (
	"path/filepath"
	"testing"
)

// TestFixtureQualityPreflight is the Anchor-7 ratchet for fixture quality.
// It enforces five invariants over every entry in tvAlignmentCases():
//
//  1. MinimumBarCount — fixture must have at least as many bars as the number of
//     reference trades whose entry falls inside the fixture's own time window.
//     A fixture cannot host more closed trades than it has bars.
//  2. FlatBarSensitivity — if the strategy's logic toggles direction on O=H=L=C
//     bar equality, the fixture must contain zero flat bars; proved empirically by
//     running the strategy on a flat-bar-perturbed copy and diffing trade output.
//  3. CrossSymbolContentDistinct — per-symbol fixtures in a cross-symbol family
//     must have distinct OHLCV+time bar content; catches placeholder sets that
//     differ only in metadata (symbol field, timezone string) while bar data is
//     copied verbatim — the authoritative single verdict per family.
//  4. SingleSymbolWindowDisjoint — a single-symbol fixture whose time window
//     contains zero reference trades while the reference CSV is non-empty is a
//     temporal-displacement placeholder; it must be rejected even when it is not
//     part of a cross-symbol family and its bar count satisfies the floor.
//  5. PriceScaleConsistent — a single-symbol fixture whose median close differs
//     from the reference median entry price by >= priceScaleIncompatibilityFactor
//     is a synthetic placeholder on price grounds, independently of bar count.
//     Cross-symbol cases are excluded; CrossSymbolContentDistinct is their authority.
func TestFixtureQualityPreflight(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	cases := tvAlignmentCases()

	t.Run("MinimumBarCount", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				fixturePath := filepath.Join(fixtureDir, tc.Data)
				f, err := loadPreflightFixture(fixturePath)
				if err != nil {
					t.Fatalf("load fixture: %v", err)
				}
				floor, err := inWindowTradeCount(fixturePath, referenceCSVPath(root, tc), tc.Timezone)
				if err != nil {
					t.Fatalf("compute in-window trade floor: %v", err)
				}
				if len(f.Bars) < floor {
					t.Errorf(
						"fixture %s has %d bars but reference CSV contains %d in-window trades:"+
							" replace with real-market data",
						tc.Data, len(f.Bars), floor,
					)
				}
			})
		}
	})

	t.Run("FlatBarSensitivity", func(t *testing.T) {
		assertFlatBarSensitivityProbeBudget(t, cases, fixtureDir)
		for _, tc := range cases {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				t.Parallel()
				f, err := loadPreflightFixture(filepath.Join(fixtureDir, tc.Data))
				if err != nil {
					t.Fatalf("load fixture: %v", err)
				}
				flatIdx := flatBarIndices(f.Bars)
				if len(flatIdx) == 0 {
					return
				}
				if !flatFractionExceedsThreshold(len(flatIdx), len(f.Bars)) {
					return
				}
				if isFlatBarSensitive(t, root, tc, f) {
					t.Errorf(
						"%s: strategy is flat-bar-sensitive; fixture %s has %d flat bars (%.1f%%, policy threshold %.0f%%)"+
							" — flat-bar-induced direction divergence exhausts RunnerOnly budget (cap %d);"+
							" replace fixture with real-market data",
						tc.Name, tc.Data, len(flatIdx),
						float64(len(flatIdx))/float64(len(f.Bars))*100,
						maxPrimaryFlatBarFraction*100, maxRunnerOnly,
					)
				}
			})
		}
	})

	t.Run("CrossSymbolContentDistinct", func(t *testing.T) {
		for _, family := range groupCrossSymbolFamilies(cases) {
			family := family
			t.Run(filepath.Base(family.Strategy), func(t *testing.T) {
				dup, isDup, err := familyDuplicateSymbols(family, fixtureDir)
				if err != nil {
					t.Fatalf("read fixture content: %v", err)
				}
				if isDup {
					t.Errorf(
						"strategy %s: symbols %q and %q have identical OHLCV bar content"+
							" — fixtures are synthetic placeholders; replace with real per-symbol market data",
						family.Strategy, dup[0], dup[1],
					)
				}
			})
		}
	})

	t.Run("SingleSymbolWindowDisjoint", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				fixturePath := filepath.Join(fixtureDir, tc.Data)
				disjoint, err := fixtureWindowDisjointFromReference(
					fixturePath, referenceCSVPath(root, tc), tc.Timezone,
				)
				if err != nil {
					t.Fatalf("window disjoint check: %v", err)
				}
				if disjoint {
					t.Errorf(
						"fixture %s has no reference trades in its time window while the reference CSV"+
							" is non-empty — fixture timestamps are displaced from real instrument history;"+
							" replace with real-market data",
						tc.Data,
					)
				}
			})
		}
	})

	t.Run("PriceScaleConsistent", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			if isCrossSymbolCase(tc, cases) {
				continue
			}
			t.Run(tc.Name, func(t *testing.T) {
				consistent, factor, err := priceScaleConsistent(
					filepath.Join(fixtureDir, tc.Data),
					referenceCSVPath(root, tc),
					tc.Timezone,
				)
				if err != nil {
					t.Fatalf("price scale check: %v", err)
				}
				if !consistent {
					t.Errorf(
						"fixture %s median close is incompatible with reference entry prices for %s"+
							" (scale factor %.0fx >= %.0fx threshold) — fixture is a synthetic placeholder;"+
							" replace with real-market data",
						tc.Data, tc.Symbol, factor, priceScaleIncompatibilityFactor,
					)
				}
			})
		}
	})
}
