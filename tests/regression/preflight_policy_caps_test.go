package regression

import (
	"testing"
	"time"
)

// Policy cap ceilings shared across every TV alignment ratchet and fixture-quality
// preflight. Defining them here — once, at package scope — means a cap change
// propagates automatically to all consumers without separate magic-number hunts.
//
// Values are fixed by the anti-scope section of .github/docs/TODO.md.
// Elevation above any cap requires explicit operator approval.
const (
	maxTimeTolerance  = 2 * time.Hour
	maxPriceTolerance = 2.00
	maxRunnerOnly     = 2
	maxTVOnly         = 4

	// maxPrimaryFlatBarFraction: flat O=H=L=C bars in a primary fixture degrade
	// direction-toggle strategies, consuming RunnerOnly headroom toward the cap.
	maxPrimaryFlatBarFraction = 0.05
)

// TestPolicyCapConstants_AntiScopeAligned pins every cap to its authoritative
// value from the anti-scope section of .github/docs/TODO.md. A silent bump to
// any constant fails this test, forcing the contributor to locate and justify
// the escalation path before the change can land.
func TestPolicyCapConstants_AntiScopeAligned(t *testing.T) {
	if maxTimeTolerance != 2*time.Hour {
		t.Errorf("maxTimeTolerance = %v; anti-scope cap is 2h — escalation requires operator approval", maxTimeTolerance)
	}
	if maxPriceTolerance != 2.00 {
		t.Errorf("maxPriceTolerance = %.2f; anti-scope cap is 2.00 — escalation requires operator approval", maxPriceTolerance)
	}
	if maxRunnerOnly != 2 {
		t.Errorf("maxRunnerOnly = %d; anti-scope cap is 2 — escalation requires operator approval", maxRunnerOnly)
	}
	if maxTVOnly != 4 {
		t.Errorf("maxTVOnly = %d; anti-scope cap is 4 — escalation requires operator approval", maxTVOnly)
	}
	if maxPrimaryFlatBarFraction != 0.05 {
		t.Errorf("maxPrimaryFlatBarFraction = %.4f; fixture-quality threshold is 0.05 — escalation requires operator approval", maxPrimaryFlatBarFraction)
	}
}

// TestPolicyCap_TolerancesAreStrictlyPositive guards against zero or negative
// tolerance values, which would make every wired alignment case a policy violation
// regardless of how well the runner matches TradingView.
func TestPolicyCap_TolerancesAreStrictlyPositive(t *testing.T) {
	if maxTimeTolerance <= 0 {
		t.Errorf("maxTimeTolerance = %v; must be > 0 — zero time tolerance rejects all timestamp matches", maxTimeTolerance)
	}
	if maxPriceTolerance <= 0 {
		t.Errorf("maxPriceTolerance = %.4f; must be > 0 — zero price tolerance rejects all price matches", maxPriceTolerance)
	}
}

// TestPolicyCap_DiscrepancyCapsAreStrictlyPositive guards against zero or
// negative discrepancy caps. A cap of zero forbids any discrepancy at all,
// making every registry entry that wires a non-zero RunnerOnly or TVOnly boundary
// an immediate policy violation and breaking the entire alignment registry.
func TestPolicyCap_DiscrepancyCapsAreStrictlyPositive(t *testing.T) {
	if maxRunnerOnly <= 0 {
		t.Errorf("maxRunnerOnly = %d; must be > 0 — zero cap forbids all runner-only discrepancies, breaking the alignment registry", maxRunnerOnly)
	}
	if maxTVOnly <= 0 {
		t.Errorf("maxTVOnly = %d; must be > 0 — zero cap forbids all TV-only discrepancies, breaking the alignment registry", maxTVOnly)
	}
}

// TestPolicyCap_TVOnlyExceedsRunnerOnly enforces the structural asymmetry of the
// policy: TV-only trades arise legitimately from fixture-end boundary effects and
// pre-window warmup history; runner-only trades signal spurious output the runner
// should not produce, so the TVOnly budget must always exceed the RunnerOnly budget.
func TestPolicyCap_TVOnlyExceedsRunnerOnly(t *testing.T) {
	if maxTVOnly <= maxRunnerOnly {
		t.Errorf("maxTVOnly (%d) must exceed maxRunnerOnly (%d); TVOnly budget is always more permissive than RunnerOnly budget", maxTVOnly, maxRunnerOnly)
	}
}

// TestPolicyCap_FlatBarFractionIsStrictlyWithinUnitInterval guards both boundary
// degenerate cases: a fraction of 0 would reject every fixture that contains even
// a single flat bar, including currently-aligning strategies whose logic is not
// flat-bar-sensitive; a fraction of 1 accepts any fixture regardless of flat-bar
// density, making the threshold vacuous and the preflight ineffective.
func TestPolicyCap_FlatBarFractionIsStrictlyWithinUnitInterval(t *testing.T) {
	if maxPrimaryFlatBarFraction <= 0 {
		t.Errorf("maxPrimaryFlatBarFraction = %.4f; must be > 0 — zero threshold rejects every flat-bar-bearing fixture unconditionally, including non-sensitive strategies", maxPrimaryFlatBarFraction)
	}
	if maxPrimaryFlatBarFraction >= 1 {
		t.Errorf("maxPrimaryFlatBarFraction = %.4f; must be < 1 — threshold of 1.0 accepts any fixture regardless of flat-bar density, making the preflight vacuous", maxPrimaryFlatBarFraction)
	}
}

// TestTVOnlyCapViolation exercises all eight boundary combinations of the
// TVOnly cap-enforcement algorithm — independently of any wired alignment case
// in the registry — so the policy rules are verified as pure logic.
func TestTVOnlyCapViolation(t *testing.T) {
	cases := []struct {
		name          string
		d             tvAlignmentDiscrepancy
		wantViolation bool
	}{
		{name: "zero_tvonly_no_flag", d: tvAlignmentDiscrepancy{TVOnly: 0}, wantViolation: false},
		{name: "below_cap_no_flag", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly - 1}, wantViolation: false},
		{name: "at_cap_no_flag", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly}, wantViolation: false},
		{name: "one_above_cap_no_flag", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly + 1}, wantViolation: true},
		{name: "well_above_cap_no_flag", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly + 10}, wantViolation: true},
		{name: "above_cap_with_escalation_approved", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly + 1, TVOnlyCapEscalated: true}, wantViolation: false},
		{name: "large_escalation_approved", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly + 10, TVOnlyCapEscalated: true}, wantViolation: false},
		{name: "dead_escalation_at_cap", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly, TVOnlyCapEscalated: true}, wantViolation: true},
		{name: "dead_escalation_below_cap", d: tvAlignmentDiscrepancy{TVOnly: maxTVOnly - 1, TVOnlyCapEscalated: true}, wantViolation: true},
		{name: "dead_escalation_zero", d: tvAlignmentDiscrepancy{TVOnly: 0, TVOnlyCapEscalated: true}, wantViolation: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			msg := tvOnlyCapViolation(tc.name, tc.d)
			gotViolation := msg != ""
			if gotViolation != tc.wantViolation {
				t.Errorf("tvOnlyCapViolation(%+v) violation=%v, want %v (msg=%q)",
					tc.d, gotViolation, tc.wantViolation, msg)
			}
		})
	}
}
