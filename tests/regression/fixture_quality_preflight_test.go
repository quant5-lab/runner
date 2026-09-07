package regression

import (
	"path/filepath"
	"testing"
	"time"
)

// TestFixtureQualityPreflight is the Anchor-7 ratchet for fixture quality.
// It enforces five invariants over every entry in tvAlignmentCases():
//
//  1. MinimumBarCount — fixture must have at least as many bars as the number of
//     reference trades whose entry falls inside the fixture's own time window.
//     A fixture cannot host more closed trades than it has bars.
//  2. CrossSymbolContentDistinct — per-symbol fixtures in a cross-symbol family
//     must have distinct OHLCV+time bar content; catches placeholder sets that
//     differ only in metadata (symbol field, timezone string) while bar data is
//     copied verbatim — the authoritative single verdict per family.
//  3. SingleSymbolWindowDisjoint — a single-symbol fixture whose time window
//     contains zero reference trades while the reference CSV is non-empty is a
//     temporal-displacement placeholder; it must be rejected even when it is not
//     part of a cross-symbol family and its bar count satisfies the floor.
//  4. PriceScaleConsistent — a single-symbol fixture whose median close differs
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

// TestPreflightDropIntegrityAudit is a compile-time + runtime canary ensuring
// every mandatory fixture-quality guard function remains present and callable.
// Each sub-test compiles against the guard function it names; a missing or
// renamed function fails to compile, catching accidental guard removal before
// any runtime assertion is reached.
func TestPreflightDropIntegrityAudit(t *testing.T) {
	root := projectRootFromCwd()
	fixtureDir := filepath.Join(root, "tests", "golden", "fixtures", "data")
	cases := tvAlignmentCases()
	if len(cases) == 0 {
		t.Fatal("tvAlignmentCases returned no cases — audit has nothing to check")
	}

	// Identify a single-symbol case for guards that need one.
	var singleCase tvAlignmentCase
	for _, c := range cases {
		if !isCrossSymbolCase(c, cases) {
			singleCase = c
			break
		}
	}
	if singleCase.Data == "" {
		t.Fatal("no single-symbol case found in registry — CrossSymbolContentDistinct and PriceScaleConsistent cannot be audited")
	}

	t.Run("MinimumBarCount", func(t *testing.T) {
		count, err := inWindowTradeCount(
			filepath.Join(fixtureDir, singleCase.Data),
			referenceCSVPath(root, singleCase),
			singleCase.Timezone,
		)
		if err != nil {
			t.Fatalf("inWindowTradeCount: %v", err)
		}
		if count < 0 {
			t.Errorf("inWindowTradeCount returned %d; must be ≥ 0", count)
		}
	})

	t.Run("CrossSymbolContentDistinct", func(t *testing.T) {
		families := groupCrossSymbolFamilies(cases)
		if len(families) == 0 {
			t.Skip("no cross-symbol families in registry — guard trivially satisfied")
		}
		for _, fam := range families {
			_, _, err := familyDuplicateSymbols(fam, fixtureDir)
			if err != nil {
				t.Fatalf("familyDuplicateSymbols(%s): %v", fam.Strategy, err)
			}
			break // one family is enough to prove the guard is callable
		}
	})

	t.Run("SingleSymbolWindowDisjoint", func(t *testing.T) {
		_, err := fixtureWindowDisjointFromReference(
			filepath.Join(fixtureDir, singleCase.Data),
			referenceCSVPath(root, singleCase),
			singleCase.Timezone,
		)
		if err != nil {
			t.Fatalf("fixtureWindowDisjointFromReference: %v", err)
		}
	})

	t.Run("PriceScaleConsistent", func(t *testing.T) {
		_, _, err := priceScaleConsistent(
			filepath.Join(fixtureDir, singleCase.Data),
			referenceCSVPath(root, singleCase),
			singleCase.Timezone,
		)
		if err != nil {
			t.Fatalf("priceScaleConsistent: %v", err)
		}
	})

	t.Run("PolicyCapConstants", func(t *testing.T) {
		// Removing or renaming these constants causes a compile failure; changing their
		// values fails here at runtime.
		if maxRunnerOnly != 2 {
			t.Errorf("maxRunnerOnly = %d; anti-scope mandates 2", maxRunnerOnly)
		}
		if maxTVOnly != 4 {
			t.Errorf("maxTVOnly = %d; anti-scope mandates 4", maxTVOnly)
		}
		if maxTimeTolerance != 2*time.Hour {
			t.Errorf("maxTimeTolerance = %v; anti-scope mandates 2h", maxTimeTolerance)
		}
		if maxPriceTolerance != 2.00 {
			t.Errorf("maxPriceTolerance = %.2f; anti-scope mandates 2.00", maxPriceTolerance)
		}
	})
}
