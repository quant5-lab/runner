package codegen

import (
	"strings"
	"testing"
)

// The mode dispatch switch is emitted for any script triggering the security evaluator;
// varying the script does not add coverage for the switch structure itself.
const securityWithUserVarInExpr = `//@version=4
strategy("Test", overlay=true)
src = ema(close, 10)
result = security(syminfo.tickerid, "1D", sma(src, 20))
plot(result)`

func TestBarMapperSetup_EmitsModeSwitchCoveringAllMappingModes(t *testing.T) {
	code := parseScript(t, securityWithUserVarInExpr)

	cases := []struct {
		desc string
		want string
	}{
		{"mode dispatch switch", "securityBarMapper.Mode()"},
		{"ModeIdentity branch label", "case request.ModeIdentity:"},
		{"ModeTransformed branch label", "case request.ModeTransformed:"},
		{"default branch for range-based modes", "default:"},
	}

	for _, c := range cases {
		if !strings.Contains(code, c.want) {
			t.Errorf("bar mapper setup missing %s: %q not found", c.desc, c.want)
		}
	}
}

func TestBarMapperSetup_IdentityBranch_MapsEachSecurityBarToItselfByIndex(t *testing.T) {
	code := parseScript(t, securityWithUserVarInExpr)

	checks := []struct {
		desc string
		want string
	}{
		{"iterates security context bars", "for i := range secCtx.Data"},
		{"maps index to itself (i→i)", "barMapper.SetMapping(i, i)"},
	}

	for _, c := range checks {
		if !strings.Contains(code, c.want) {
			t.Errorf("ModeIdentity branch: %s — %q not found in generated code", c.desc, c.want)
		}
	}
}

func TestBarMapperSetup_TransformedBranch_BuildsReverseMappingFromMainToSynthetic(t *testing.T) {
	code := parseScript(t, securityWithUserVarInExpr)

	checks := []struct {
		desc string
		want string
	}{
		{"iterates mainToSynthetic slice", "for mainIdx, secIdx := range securityBarMapper.MainToSynthetic()"},
		{"guards against pre-formation sentinel -1", "if secIdx >= 0"},
		{"stores reverse mapping secIdx→mainIdx", "barMapper.SetMapping(secIdx, mainIdx)"},
	}

	for _, c := range checks {
		if !strings.Contains(code, c.want) {
			t.Errorf("ModeTransformed branch: %s — %q not found in generated code", c.desc, c.want)
		}
	}
}

func TestBarMapperSetup_DefaultBranch_PopulatesFromRangesForCrossTimeframeModes(t *testing.T) {
	code := parseScript(t, securityWithUserVarInExpr)

	checks := []struct {
		desc string
		want string
	}{
		{"iterates bar ranges", "for _, rr := range securityBarMapper.GetRanges()"},
		{"guards against unmatched range sentinel", "rr.StartHourlyIndex >= 0"},
		{"maps security bar to first lower-TF bar in its period", "barMapper.SetMapping(rr.DailyBarIndex, rr.StartHourlyIndex)"},
	}

	for _, c := range checks {
		if !strings.Contains(code, c.want) {
			t.Errorf("default branch: %s — %q not found in generated code", c.desc, c.want)
		}
	}
}
