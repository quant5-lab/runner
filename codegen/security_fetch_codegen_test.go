package codegen

import (
	"fmt"
	"strings"
	"testing"
)

func TestFetchDataBlock_LimitBlockComputesWarmupBound(t *testing.T) {
	const warmup = 75
	code := fetchDataBlock("v", "ctx.Symbol", `"1D"`, warmup)

	required := []string{
		"v_limit := len(ctx.Data)",
		"secTimeframeSeconds != baseTimeframeSeconds",
		fmt.Sprintf("detectedWarmup := %d", warmup),
		"fixedMinimumWarmup := 500",
		"requiredWarmup",
		"v_limit = baseSecurityBars + requiredWarmup",
	}
	for _, frag := range required {
		if !strings.Contains(code, frag) {
			t.Errorf("limit block missing %q\ncode:\n%s", frag, code)
		}
	}
}

func TestFetchDataBlock_WarmupBarsEmbeddedAsIntegerLiteral(t *testing.T) {
	for _, warmup := range []int{0, 50, 100, 250, 500} {
		code := fetchDataBlock("v", "ctx.Symbol", `"4h"`, warmup)
		literal := fmt.Sprintf("detectedWarmup := %d", warmup)
		if !strings.Contains(code, literal) {
			t.Errorf("warmup=%d: expected %q in generated code:\n%s", warmup, literal, code)
		}
	}
}

func TestFetchDataBlock_OutputVarsDefaultToPrimaryContext(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"4h"`, 50)

	vars := []string{
		"var v_data []context.OHLCV",
		"v_tz := ctx.Timezone",
		"v_refsession := ctx.ReferenceSession",
		"v_anchor := ctx.PeriodAnchor",
	}
	for _, v := range vars {
		if !strings.Contains(code, v) {
			t.Errorf("output var declaration missing %q\ncode:\n%s", v, code)
		}
	}
}

// Both static and runtime symbolCode are tested so neither resolution mode can
// silently break the same-symbol guard.
func TestFetchDataBlock_AggregateFallbackGuardCondition(t *testing.T) {
	cases := []struct {
		name       string
		symbolCode string
		wantGuard  string
	}{
		{
			name:       "static_symbol_literal",
			symbolCode: `"AAPL"`,
			wantGuard:  `"AAPL" == ctx.Symbol`,
		},
		{
			name:       "runtime_ctx_symbol",
			symbolCode: "ctx.Symbol",
			wantGuard:  "ctx.Symbol == ctx.Symbol",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := fetchDataBlock("v", tc.symbolCode, `"4h"`, 50)

			if !strings.Contains(code, "secTimeframeSeconds > baseTimeframeSeconds") {
				t.Errorf("fallback guard missing coarser-TF condition:\n%s", code)
			}
			if !strings.Contains(code, tc.wantGuard) {
				t.Errorf("fallback guard missing same-symbol condition %q:\n%s", tc.wantGuard, code)
			}
		})
	}
}

// ctx.PeriodAnchor is already derived from observed session opens; the fallback
// must not introduce a second derivation.
func TestFetchDataBlock_AggregateFallbackCallsAggregateWithPrimaryAnchor(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"4h"`, 50)

	if !strings.Contains(code, "context.AggregateToCoarserPeriod(ctx.Data") {
		t.Errorf("fallback path must call AggregateToCoarserPeriod with ctx.Data:\n%s", code)
	}
	if !strings.Contains(code, "ctx.PeriodAnchor") {
		t.Errorf("fallback path must pass ctx.PeriodAnchor to AggregateToCoarserPeriod:\n%s", code)
	}
}

// Continuing with empty data for a foreign symbol would produce silently wrong output.
func TestFetchDataBlock_MissingFixtureForForeignSymbolHardExits(t *testing.T) {
	code := fetchDataBlock("v", `"EURUSD"`, `"1D"`, 50)

	if !strings.Contains(code, "os.Exit(1)") {
		t.Errorf("hard-exit path must call os.Exit(1):\n%s", code)
	}
	if !strings.Contains(code, "Failed to fetch") {
		t.Errorf("hard-exit path must print a 'Failed to fetch' message to stderr:\n%s", code)
	}
}

func TestFetchDataBlock_NormalPathNormalizesWithSecondaryMetadata(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"1D"`, 50)

	required := []string{
		"market.CompleteSecondaryMetadata(",
		"v_marketData.SourceMetadata",
		"market.SecondaryContextDefaults{Timezone: ctx.Timezone, ReferenceSession: ctx.ReferenceSession}",
		"market.NormalizeBarsWithMetadataE",
		"v_normErr",
		"v_data = v_normalized",
	}
	for _, frag := range required {
		if !strings.Contains(code, frag) {
			t.Errorf("normalize path missing %q\ncode:\n%s", frag, code)
		}
	}
}

// An un-normalizable fixture would produce misaligned timestamps corrupting all downstream calculations.
func TestFetchDataBlock_NormalizationErrorHardExits(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"1D"`, 50)

	if !strings.Contains(code, "v_normErr != nil") {
		t.Errorf("normalize error check missing:\n%s", code)
	}
	// os.Exit(1) appears at least twice: normalize error AND the fetch-fail else branch.
	if strings.Count(code, "os.Exit(1)") < 2 {
		t.Errorf("expected os.Exit(1) in both fallback and normalize error paths:\n%s", code)
	}
}

// Multi-exchange strategies require each secondary to use its own session-open
// anchor (NYSE 09:30 ET, MOEX 07:00 MSK) independently of the primary.
func TestFetchDataBlock_SecondaryAnchorDerivedFromSecondaryProfileAndBars(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"1D"`, 50)

	anchor := "v_anchor = market.DeriveSessionAnchor(v_profile, v_normalized)"
	if !strings.Contains(code, anchor) {
		t.Errorf("normalize path must derive secondary anchor from v_profile and v_normalized:\n%s", code)
	}

	// The derivation must occur AFTER the default declaration v_anchor := ctx.PeriodAnchor
	// so that the aggregate-fallback path (which never enters the else block) still
	// keeps the primary anchor as its default.
	defaultDecl := "v_anchor := ctx.PeriodAnchor"
	defaultIdx := strings.Index(code, defaultDecl)
	overrideIdx := strings.Index(code, anchor)
	if defaultIdx < 0 {
		t.Fatalf("default anchor declaration %q not found", defaultDecl)
	}
	if overrideIdx < 0 {
		t.Fatalf("anchor override %q not found", anchor)
	}
	if overrideIdx <= defaultIdx {
		t.Errorf("anchor override must appear after default declaration; defaultIdx=%d overrideIdx=%d", defaultIdx, overrideIdx)
	}
}

func TestFetchDataBlock_TrimBlockSlicesToComputedLimit(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"4h"`, 50)

	if !strings.Contains(code, "v_limit > 0 && v_limit < len(v_data)") {
		t.Errorf("trim guard missing:\n%s", code)
	}
	if !strings.Contains(code, "v_data = v_data[len(v_data)-v_limit:]") {
		t.Errorf("trim slice expression missing:\n%s", code)
	}
}

func TestFetchDataBlock_TrimOccursAfterAcquisition(t *testing.T) {
	code := fetchDataBlock("v", "ctx.Symbol", `"4h"`, 50)

	fallbackIdx := strings.Index(code, "context.AggregateToCoarserPeriod")
	trimIdx := strings.Index(code, "v_data = v_data[len(v_data)-v_limit:]")
	if fallbackIdx < 0 || trimIdx < 0 {
		t.Fatal("expected both aggregate fallback and trim slice in generated code")
	}
	if trimIdx <= fallbackIdx {
		t.Errorf("trim must appear after aggregate fallback; fallbackIdx=%d trimIdx=%d", fallbackIdx, trimIdx)
	}
}

// Multiple feeds in the same binary require each feed's variables to be namespaced by varName prefix.
func TestFetchDataBlock_VarNamePropagatedToAllEmittedIdentifiers(t *testing.T) {
	const varName = "aapl_4h"
	code := fetchDataBlock(varName, "ctx.Symbol", `"4h"`, 50)

	identifiers := []string{
		varName + "_limit",
		varName + "_data",
		varName + "_tz",
		varName + "_refsession",
		varName + "_anchor",
		varName + "_marketData",
		varName + "_err",
		varName + "_normalized",
		varName + "_profile",
		varName + "_normErr",
		varName + "_metadata",
	}
	for _, id := range identifiers {
		if !strings.Contains(code, id) {
			t.Errorf("expected identifier %q in fetchDataBlock output:\n%s", id, code)
		}
	}
}
