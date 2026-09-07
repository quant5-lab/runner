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

func TestDiscrepancyCapViolation(t *testing.T) {
	boundaries := []struct {
		name      string
		cap       int
		make      func(actual int, escalated bool) tvAlignmentDiscrepancy
		validate  func(string, tvAlignmentDiscrepancy) string
		fieldName string
	}{
		{
			name: "runner-only",
			cap:  maxRunnerOnly,
			make: func(actual int, escalated bool) tvAlignmentDiscrepancy {
				return tvAlignmentDiscrepancy{RunnerOnly: actual, RunnerOnlyCapEscalated: escalated}
			},
			validate:  runnerOnlyCapViolation,
			fieldName: "RunnerOnly",
		},
		{
			name: "tv-only",
			cap:  maxTVOnly,
			make: func(actual int, escalated bool) tvAlignmentDiscrepancy {
				return tvAlignmentDiscrepancy{TVOnly: actual, TVOnlyCapEscalated: escalated}
			},
			validate:  tvOnlyCapViolation,
			fieldName: "TVOnly",
		},
		{
			name: "export-horizon-runner-only",
			cap:  maxRunnerOnly,
			make: func(actual int, escalated bool) tvAlignmentDiscrepancy {
				return tvAlignmentDiscrepancy{ExportHorizonRunnerOnly: actual, ExportHorizonRunnerOnlyCapEscalated: escalated}
			},
			validate:  exportHorizonRunnerOnlyCapViolation,
			fieldName: "ExportHorizonRunnerOnly",
		},
	}

	states := []struct {
		name          string
		actualOffset  int
		escalated     bool
		wantViolation bool
	}{
		{name: "zero_without_flag", actualOffset: -999, escalated: false, wantViolation: false},
		{name: "below_cap_without_flag", actualOffset: -1, escalated: false, wantViolation: false},
		{name: "at_cap_without_flag", actualOffset: 0, escalated: false, wantViolation: false},
		{name: "above_cap_without_flag", actualOffset: 1, escalated: false, wantViolation: true},
		{name: "far_above_cap_without_flag", actualOffset: 10, escalated: false, wantViolation: true},
		{name: "above_cap_with_flag", actualOffset: 1, escalated: true, wantViolation: false},
		{name: "far_above_cap_with_flag", actualOffset: 10, escalated: true, wantViolation: false},
		{name: "at_cap_with_dead_flag", actualOffset: 0, escalated: true, wantViolation: true},
		{name: "below_cap_with_dead_flag", actualOffset: -1, escalated: true, wantViolation: true},
		{name: "zero_with_dead_flag", actualOffset: -999, escalated: true, wantViolation: true},
	}

	for _, boundary := range boundaries {
		boundary := boundary
		t.Run(boundary.name, func(t *testing.T) {
			for _, state := range states {
				state := state
				t.Run(state.name, func(t *testing.T) {
					actual := boundary.cap + state.actualOffset
					if actual < 0 {
						actual = 0
					}
					msg := boundary.validate(state.name, boundary.make(actual, state.escalated))
					gotViolation := msg != ""
					if gotViolation != state.wantViolation {
						t.Errorf("%s cap violation=%v, want %v for actual=%d escalated=%v (msg=%q)",
							boundary.fieldName, gotViolation, state.wantViolation, actual, state.escalated, msg)
					}
				})
			}
		})
	}
}
