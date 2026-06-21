package regression

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTVAlignmentCasesAreUniqueAndComplete(t *testing.T) {
	root := projectRootFromCwd()
	seenNames := map[string]bool{}
	seenInputs := map[string]bool{}
	for _, tc := range tvAlignmentCases() {
		if seenNames[tc.Name] {
			t.Errorf("duplicate TV alignment case %q", tc.Name)
		}
		seenNames[tc.Name] = true

		inputKey := tc.Strategy + "\x00" + tc.Data + "\x00" + tc.Symbol + "\x00" + tc.Timeframe + "\x00" + tc.CSV
		if seenInputs[inputKey] {
			t.Errorf("%s duplicates a TV alignment input tuple", tc.Name)
		}
		seenInputs[inputKey] = true

		if tc.Strategy == "" || tc.Data == "" || tc.Symbol == "" || tc.Timeframe == "" || tc.Golden == "" || tc.CSV == "" {
			t.Errorf("%s has an empty required field: %+v", tc.Name, tc)
		}
		if tc.Tolerance.Time <= 0 || tc.Tolerance.Price <= 0 {
			t.Errorf("%s has invalid tolerances: %+v", tc.Name, tc)
		}
		if tc.Discrepancy.RunnerOnly < 0 || tc.Discrepancy.TVOnly < 0 {
			t.Errorf("%s has invalid discrepancy boundary: %+v", tc.Name, tc)
		}
		for label, path := range map[string]string{
			"strategy": filepath.Join(root, "strategies", tc.Strategy),
			"data":     filepath.Join(root, "tests", "golden", "fixtures", "data", tc.Data),
			"golden":   filepath.Join(root, "tests", "golden", "fixtures", "expected", tc.Golden),
			"csv":      filepath.Join(root, "tests", "regression", "tv_reference", "fixtures", tc.CSV),
		} {
			if _, err := os.Stat(path); err != nil {
				t.Errorf("%s %s path invalid: %v", tc.Name, label, err)
			}
		}
	}
}

func TestTVAlignmentCasesStayWithinPolicyCaps(t *testing.T) {
	const maxTimeTolerance = 2 * time.Hour
	const maxPriceTolerance = 2.00
	const maxRunnerOnly = 2
	const maxTVOnly = 5

	for _, tc := range tvAlignmentCases() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Tolerance.Time > maxTimeTolerance {
				t.Errorf("time tolerance %s exceeds policy cap %s", tc.Tolerance.Time, maxTimeTolerance)
			}
			if tc.Tolerance.Price > maxPriceTolerance {
				t.Errorf("price tolerance %.4f exceeds policy cap %.4f", tc.Tolerance.Price, maxPriceTolerance)
			}
			if tc.Discrepancy.RunnerOnly > maxRunnerOnly {
				t.Errorf("runner-only boundary %d exceeds policy cap %d", tc.Discrepancy.RunnerOnly, maxRunnerOnly)
			}
			if tc.Discrepancy.TVOnly > maxTVOnly {
				t.Errorf("tv-only boundary %d exceeds policy cap %d", tc.Discrepancy.TVOnly, maxTVOnly)
			}
		})
	}
}

func TestTVAlignmentCasesUseSharedPolicyObjects(t *testing.T) {
	for _, tc := range tvAlignmentCases() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Tolerance == (tvAlignmentTolerance{}) {
				t.Fatalf("missing tolerance policy")
			}
			if tc.Discrepancy.RunnerOnly == 0 && tc.Discrepancy.TVOnly > tc.Discrepancy.FixtureEndOpen {
				t.Fatalf("tv-only discrepancy without runner-only boundary needs an explicit case review")
			}
			if tc.Discrepancy.RunnerOnly > 0 && tc.Discrepancy.TVOnly == 0 {
				t.Fatalf("runner-only discrepancy without tv-only boundary needs an explicit case review")
			}
			if tc.PnLDiscrepancy < 0 {
				t.Fatalf("PnLDiscrepancy must be non-negative, got %d", tc.PnLDiscrepancy)
			}
		})
	}
}

func TestExactAlignmentCases_NeverIncludesSkippedPnL(t *testing.T) {
	for _, tc := range exactAlignmentCases() {
		if tc.SkipPnLRatchetReason != "" {
			t.Errorf("%s: exactAlignmentCases must not include cases with a SkipPnLRatchetReason", tc.Name)
		}
	}
}

// TestTVAlignmentCases_SomeAssertSize guards against the size ratchet becoming
// universally bypassed via SkipSizeRatchetReason on every case.
func TestTVAlignmentCases_SomeAssertSize(t *testing.T) {
	checked := 0
	for _, tc := range tvAlignmentCases() {
		if tc.SkipSizeRatchetReason == "" {
			checked++
		}
	}
	if checked == 0 {
		t.Error("every TV alignment case has SkipSizeRatchetReason set — at least one case must assert position size parity")
	}
}

// TestTVAlignmentCases_SomeSkipSizeRatchet confirms SkipSizeRatchetReason is
// load-bearing: removing all uses would require deleting the field and this test.
func TestTVAlignmentCases_SomeSkipSizeRatchet(t *testing.T) {
	skipped := 0
	for _, tc := range tvAlignmentCases() {
		if tc.SkipSizeRatchetReason != "" {
			skipped++
		}
	}
	if skipped == 0 {
		t.Error("no TV alignment case sets SkipSizeRatchetReason — field is unreachable; remove it or add a case that requires it")
	}
}

func TestTVAlignmentCases_InitialCapitalNonZero(t *testing.T) {
	for _, tc := range tvAlignmentCases() {
		if tc.InitialCapital == 0 {
			t.Errorf("case %q has InitialCapital=0; PnL ratchet divides by equity which is derived from InitialCapital — zero causes division by zero", tc.Name)
		}
	}
}

func TestTVAlignmentCases_SomePnLRatchetSkipped(t *testing.T) {
	skipped := 0
	for _, tc := range tvAlignmentCases() {
		if tc.SkipPnLRatchetReason != "" {
			skipped++
		}
	}
	if skipped == 0 {
		t.Error("no TV alignment case sets SkipPnLRatchetReason — field is unreachable; remove it or add a case that requires it")
	}
}

func TestSkippedPnLCases_ExcludedFromExactAlignment(t *testing.T) {
	exactSet := make(map[string]bool)
	for _, tc := range exactAlignmentCases() {
		exactSet[tc.Name] = true
	}
	for _, tc := range tvAlignmentCases() {
		if tc.SkipPnLRatchetReason != "" && exactSet[tc.Name] {
			t.Errorf("case %q has SkipPnLRatchetReason but appears in exactAlignmentCases — the filter must exclude it from PnL comparison", tc.Name)
		}
	}
}
