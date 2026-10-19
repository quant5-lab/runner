package codegen

import (
	"strings"
	"testing"
)

func TestMapperInitBlock_DispatchChainBuilderCalls(t *testing.T) {
	code := mapperInitBlock("sec", false)

	cases := []struct {
		name    string
		builder string
	}{
		{"finer_secondary", "BuildMappingForUpscaling"},
		{"same_tf_identity", "BuildIdentityMapping"},
		{"intraday_coarser", "BuildMappingByTimestamp"},
		{"calendar_coarser", "BuildMappingWithDateFilter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(code, tc.builder) {
				t.Errorf("dispatch arm %q: builder %q not found in:\n%s", tc.name, tc.builder, code)
			}
		})
	}
}

func TestMapperInitBlock_DispatchChainConditions(t *testing.T) {
	code := mapperInitBlock("sec", false)

	cases := []struct {
		name      string
		condition string
	}{
		{"finer_secondary", "secTimeframeSeconds < baseTimeframeSeconds"},
		{"same_tf", "secTimeframeSeconds == baseTimeframeSeconds"},
		{"intraday_coarser", "secTimeframeSeconds < 86400"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(code, tc.condition) {
				t.Errorf("dispatch condition %q: %q not found in:\n%s", tc.name, tc.condition, code)
			}
		})
	}
}

// An off-by-one at the 86400 s boundary would mis-classify daily bars as intraday
// or omit the date-filter for daily+ secondaries.
func TestMapperInitBlock_IntradayBoundaryIsStrictlyOneDayInSeconds(t *testing.T) {
	code := mapperInitBlock("sec", false)

	if !strings.Contains(code, "< 86400") {
		t.Errorf("intraday boundary must use strict less-than (< 86400); got:\n%s", code)
	}
}

func TestMapperInitBlock_VariableBarCountSelectsTransformBuilder(t *testing.T) {
	identity := mapperInitBlock("sec", false)
	transform := mapperInitBlock("sec", true)

	t.Run("false_emits_identity", func(t *testing.T) {
		if !strings.Contains(identity, "BuildIdentityMapping") {
			t.Errorf("variableBarCount=false must emit BuildIdentityMapping:\n%s", identity)
		}
		if strings.Contains(identity, "BuildMappingFromTransform") {
			t.Errorf("variableBarCount=false must not emit BuildMappingFromTransform:\n%s", identity)
		}
	})
	t.Run("true_emits_transform", func(t *testing.T) {
		if !strings.Contains(transform, "BuildMappingFromTransform") {
			t.Errorf("variableBarCount=true must emit BuildMappingFromTransform:\n%s", transform)
		}
		if strings.Contains(transform, "BuildIdentityMapping") {
			t.Errorf("variableBarCount=true must not emit BuildIdentityMapping:\n%s", transform)
		}
	})
}

func TestMapperInitBlock_VarNamePropagatedToMapperAndCtx(t *testing.T) {
	const varName = "sberp_4h"
	code := mapperInitBlock(varName, false)

	for _, fragment := range []string{
		varName + "_mapper := request.NewSecurityBarMapper()",
		varName + "_ctx.Data",
		varName + "_mapper.",
	} {
		if !strings.Contains(code, fragment) {
			t.Errorf("expected %q in mapperInitBlock output:\n%s", fragment, code)
		}
	}
}

func TestMapperInitBlock_CrossContaminationAbsent(t *testing.T) {
	code := mapperInitBlock("sec", false)

	builders := []string{
		"BuildMappingForUpscaling",
		"BuildIdentityMapping",
		"BuildMappingByTimestamp",
		"BuildMappingWithDateFilter",
	}
	for _, b := range builders {
		count := strings.Count(code, b)
		if count != 1 {
			t.Errorf("builder %q: expected exactly 1 occurrence, got %d:\n%s", b, count, code)
		}
	}
}
