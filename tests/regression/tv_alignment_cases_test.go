package regression

import (
	"os"
	"path/filepath"
	"testing"
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
	for _, tc := range tvAlignmentCases() {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Tolerance.Time > maxTimeTolerance {
				t.Errorf("time tolerance %s exceeds policy cap %s", tc.Tolerance.Time, maxTimeTolerance)
			}
			if tc.Tolerance.Price > maxPriceTolerance {
				t.Errorf("price tolerance %.4f exceeds policy cap %.4f", tc.Tolerance.Price, maxPriceTolerance)
			}
			if msg := runnerOnlyCapViolation(tc.Name, tc.Discrepancy); msg != "" {
				t.Errorf("%s", msg)
			}
			if msg := tvOnlyCapViolation(tc.Name, tc.Discrepancy); msg != "" {
				t.Errorf("%s", msg)
			}
			if msg := exportHorizonRunnerOnlyCapViolation(tc.Name, tc.Discrepancy); msg != "" {
				t.Errorf("%s", msg)
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
			// Attribution via FixtureEndOpen/FixtureStartWarmup is required only when TVOnly
			// exceeds the policy cap — within-cap TVOnly without a RunnerOnly boundary is acceptable
			// since the cap already limits accumulation and the case has been reviewed.
			if tc.Discrepancy.RunnerOnly == 0 &&
				tc.Discrepancy.TVOnly > maxTVOnly &&
				tc.Discrepancy.TVOnly > tc.Discrepancy.FixtureEndOpen+tc.Discrepancy.FixtureStartWarmup {
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

func TestTVAlignmentCases_CapEscalationRatchet(t *testing.T) {
	const (
		maxTVOnlyEscalations     = 1 // operator-approved: UtPlus SBERP
		maxRunnerOnlyEscalations = 0 // none currently operator-approved
	)
	nTVOnly, nRunnerOnly := 0, 0
	for _, tc := range tvAlignmentCases() {
		if tc.Discrepancy.TVOnlyCapEscalated {
			nTVOnly++
		}
		if tc.Discrepancy.RunnerOnlyCapEscalated {
			nRunnerOnly++
		}
	}
	if nTVOnly > maxTVOnlyEscalations {
		t.Errorf("TVOnlyCapEscalated count is %d (ceiling %d) — each additional operator-approved escalation must be individually justified", nTVOnly, maxTVOnlyEscalations)
	}
	if nRunnerOnly > maxRunnerOnlyEscalations {
		t.Errorf("RunnerOnlyCapEscalated count is %d (ceiling %d) — each additional operator-approved escalation must be individually justified", nRunnerOnly, maxRunnerOnlyEscalations)
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

// TestTVAlignmentCases_ExportHorizonCapEscalationRatchet bounds the number of
// operator-approved ExportHorizonRunnerOnly cap escalations. ExportHorizonRunnerOnly
// uses the same maxRunnerOnly cap as RunnerOnly and each escalation must be individually
// justified — this ratchet prevents silent accumulation.
func TestTVAlignmentCases_ExportHorizonCapEscalationRatchet(t *testing.T) {
	const maxEscalations = 0 // none currently operator-approved
	n := 0
	for _, tc := range tvAlignmentCases() {
		if tc.Discrepancy.ExportHorizonRunnerOnlyCapEscalated {
			n++
		}
	}
	if n > maxEscalations {
		t.Errorf("ExportHorizonRunnerOnlyCapEscalated count is %d (ceiling %d) — each additional operator-approved escalation must be individually justified", n, maxEscalations)
	}
}

// TestTVAlignmentCases_ExemptionSumBoundedByTVOnly enforces that the sum of all
// boundary-attribution exemptions never exceeds the total TVOnly count for a case.
// FixtureEndOpen and FixtureStartWarmup together explain a subset of TV-only trades;
// declaring more exemptions than there are TV-only trades is logically impossible and
// indicates a miscounted entry.
func TestTVAlignmentCases_ExemptionSumBoundedByTVOnly(t *testing.T) {
	for _, tc := range tvAlignmentCases() {
		d := tc.Discrepancy
		sum := d.FixtureEndOpen + d.FixtureStartWarmup
		if sum > d.TVOnly {
			t.Errorf(
				"%s: FixtureEndOpen(%d)+FixtureStartWarmup(%d)=%d exceeds TVOnly(%d) — "+
					"declared exemptions cannot explain more TV-only trades than exist",
				tc.Name, d.FixtureEndOpen, d.FixtureStartWarmup, sum, d.TVOnly,
			)
		}
	}
}

// TestTVAlignmentCases_ExemptionBoundaryFieldsAreNonNegative extends the
// RunnerOnly/TVOnly non-negativity check (already in TestTVAlignmentCasesAreUniqueAndComplete)
// to every exemption field in tvAlignmentDiscrepancy.
func TestTVAlignmentCases_ExemptionBoundaryFieldsAreNonNegative(t *testing.T) {
	for _, tc := range tvAlignmentCases() {
		d := tc.Discrepancy
		if d.FixtureEndOpen < 0 {
			t.Errorf("%s: FixtureEndOpen=%d must be non-negative", tc.Name, d.FixtureEndOpen)
		}
		if d.FixtureStartWarmup < 0 {
			t.Errorf("%s: FixtureStartWarmup=%d must be non-negative", tc.Name, d.FixtureStartWarmup)
		}
		if d.ExportHorizonRunnerOnly < 0 {
			t.Errorf("%s: ExportHorizonRunnerOnly=%d must be non-negative", tc.Name, d.ExportHorizonRunnerOnly)
		}
	}
}

// exemptionBudget holds the approved bounds for one exemption field in
// tvAlignmentDiscrepancy. Keeping all three bounds together means adding a new
// exemption field requires exactly one new row in exemptionBudgets.
type exemptionBudget struct {
	name         string
	get          func(tvAlignmentDiscrepancy) int
	minCasesUsed int    // liveness floor: the field must appear in at least this many cases
	maxCasesUsed int    // case-count ceiling: prevents new cases from silently adding the exemption
	maxTotal     int    // magnitude ceiling: prevents existing cases from silently inflating their value
	justifyHint  string // operator guidance surfaced in the error when a ceiling is exceeded
}

// exemptionBudgets is the single source of truth for every exemption field's
// approved bounds — a new exemption field requires exactly one new row here.
var exemptionBudgets = []exemptionBudget{
	{
		name:         "FixtureEndOpen",
		get:          func(d tvAlignmentDiscrepancy) int { return d.FixtureEndOpen },
		minCasesUsed: 1,
		maxCasesUsed: 4, // BB+RSI BTCUSDT, BB+RSI AAPL, BB7 BTCUSDT, Moon BTCUSDT
		maxTotal:     4, // each of the 4 cases contributes exactly 1
		justifyHint:  "verify the runner genuinely holds an open position at fixture end",
	},
	{
		name:         "FixtureStartWarmup",
		get:          func(d tvAlignmentDiscrepancy) int { return d.FixtureStartWarmup },
		minCasesUsed: 1,
		maxCasesUsed: 1, // Moon BTCUSDT: 1 TV trade precedes the BTCUSDT-M.json fixture start
		maxTotal:     1, // Moon BTCUSDT=1
		justifyHint:  "verify TV history genuinely predates the available fixture data",
	},
	{
		name:         "ExportHorizonRunnerOnly",
		get:          func(d tvAlignmentDiscrepancy) int { return d.ExportHorizonRunnerOnly },
		minCasesUsed: 1,
		maxCasesUsed: 4,  // Hull SBERP=1, UtPlus SBERP=12, BB+RSI SBERP=1, Aostoch SBERP=1
		maxTotal:     15, // Hull=1 + UtPlus=12 + BB+RSI=1 + Aostoch=1
		justifyHint:  "verify the reference CSV is provably truncated before the fixture end; if TV genuinely stopped trading, remove ExportHorizonRunnerOnly",
	},
}

// TestTVAlignmentCases_ExemptionFieldCaseCountsBounded enforces a two-sided
// case-count ratchet on every exemption field: the field must appear in at least
// minCasesUsed cases (liveness) and at most maxCasesUsed cases (ceiling). The
// per-case behavioral guards in tv_alignment_content_test.go cover the orthogonal
// per-value dimension; this test covers the cross-case count dimension.
func TestTVAlignmentCases_ExemptionFieldCaseCountsBounded(t *testing.T) {
	cases := tvAlignmentCases()
	for _, b := range exemptionBudgets {
		b := b
		t.Run(b.name, func(t *testing.T) {
			n := 0
			for _, tc := range cases {
				if b.get(tc.Discrepancy) > 0 {
					n++
				}
			}
			if n < b.minCasesUsed {
				t.Errorf(
					"%s appears in %d case(s) (minimum %d) — "+
						"field and its guard clause may be dead code; "+
						"add a case that requires it or remove the field",
					b.name, n, b.minCasesUsed,
				)
			}
			if n > b.maxCasesUsed {
				t.Errorf(
					"%s appears in %d case(s) (ceiling %d) — "+
						"each additional exemption must be individually justified; %s",
					b.name, n, b.maxCasesUsed, b.justifyHint,
				)
			}
		})
	}
}

// TestTVAlignmentCases_ExemptionFieldTotalsBounded closes the inflation gap that
// TestTVAlignmentCases_ExemptionFieldCaseCountsBounded leaves open: a case-count
// ceiling prevents new cases from being added, but an existing case can silently
// bump its own value; the magnitude ceiling here catches that.
func TestTVAlignmentCases_ExemptionFieldTotalsBounded(t *testing.T) {
	cases := tvAlignmentCases()
	for _, b := range exemptionBudgets {
		b := b
		t.Run(b.name, func(t *testing.T) {
			total := 0
			for _, tc := range cases {
				total += b.get(tc.Discrepancy)
			}
			if total > b.maxTotal {
				t.Errorf(
					"total %s across all cases is %d (ceiling %d) — "+
						"each additional exempted trade must be individually justified; %s",
					b.name, total, b.maxTotal, b.justifyHint,
				)
			}
		})
	}
}
