package codegen

import (
	"strings"
	"testing"
)

func TestArrowContextHoister_EmptyCallSites(t *testing.T) {
	tests := []struct {
		name  string
		sites []ArrowCallSite
	}{
		{"nil_slice", nil},
		{"empty_slice", []ArrowCallSite{}},
	}

	hoister := NewArrowContextHoister("\t")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := hoister.GeneratePreLoopDeclarations(tt.sites)
			if code != "" {
				t.Errorf("expected empty code, got %q", code)
			}
		})
	}
}

func TestArrowContextHoister_SingleCallSite(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "adx", CallIndex: 1, ContextVar: "arrowCtx_adx_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	expected := "\tarrowCtx_adx_1 := context.NewArrowContext(ctx)\n"
	if code != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, code)
	}
}

func TestArrowContextHoister_MultipleCallSites(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "adx", CallIndex: 1, ContextVar: "arrowCtx_adx_1"},
		{FunctionName: "rma", CallIndex: 1, ContextVar: "arrowCtx_rma_1"},
		{FunctionName: "ema", CallIndex: 1, ContextVar: "arrowCtx_ema_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	expectedVars := []string{"arrowCtx_adx_1", "arrowCtx_rma_1", "arrowCtx_ema_1"}
	for _, v := range expectedVars {
		if !strings.Contains(code, v) {
			t.Errorf("missing variable %q in output:\n%s", v, code)
		}
	}
}

/* TestArrowContextHoister_IndentationVariations verifies all indentation styles */
func TestArrowContextHoister_IndentationVariations(t *testing.T) {
	tests := []struct {
		name        string
		indentation string
	}{
		{"no_indentation", ""},
		{"single_tab", "\t"},
		{"double_tab", "\t\t"},
		{"four_spaces", "    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hoister := NewArrowContextHoister(tt.indentation)
			sites := []ArrowCallSite{
				{FunctionName: "f", CallIndex: 1, ContextVar: "arrowCtx_f_1"},
			}

			code := hoister.GeneratePreLoopDeclarations(sites)

			expected := tt.indentation + "arrowCtx_f_1 := context.NewArrowContext(ctx)"
			if !strings.HasPrefix(code, expected) {
				t.Errorf("expected prefix %q, got %q", expected, code)
			}
		})
	}
}

/* TestArrowContextHoister_OrderPreservation verifies declaration order matches input order */
func TestArrowContextHoister_OrderPreservation(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "first", CallIndex: 1, ContextVar: "arrowCtx_first_1"},
		{FunctionName: "second", CallIndex: 1, ContextVar: "arrowCtx_second_1"},
		{FunctionName: "third", CallIndex: 1, ContextVar: "arrowCtx_third_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)
	lines := strings.Split(strings.TrimSpace(code), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	expectedOrder := []string{"arrowCtx_first_1", "arrowCtx_second_1", "arrowCtx_third_1"}
	for i, expectedVar := range expectedOrder {
		if !strings.Contains(lines[i], expectedVar) {
			t.Errorf("line %d: expected %q, got %q", i, expectedVar, lines[i])
		}
	}
}

/* TestArrowContextHoister_SameFunctionMultipleInstances verifies unique context per call site */
func TestArrowContextHoister_SameFunctionMultipleInstances(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "calc", CallIndex: 1, ContextVar: "arrowCtx_calc_1"},
		{FunctionName: "calc", CallIndex: 2, ContextVar: "arrowCtx_calc_2"},
		{FunctionName: "calc", CallIndex: 3, ContextVar: "arrowCtx_calc_3"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	for _, varName := range []string{"arrowCtx_calc_1", "arrowCtx_calc_2", "arrowCtx_calc_3"} {
		count := strings.Count(code, varName)
		if count != 1 {
			t.Errorf("variable %q appears %d times, expected 1", varName, count)
		}
	}
}

/* TestArrowContextHoister_CodeFormat verifies structural properties of generated code */
func TestArrowContextHoister_CodeFormat(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "test", CallIndex: 1, ContextVar: "arrowCtx_test_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	checks := []struct {
		desc    string
		pattern string
	}{
		{"short variable declaration", ":="},
		{"constructor call", "context.NewArrowContext"},
		{"ctx parameter", "(ctx)"},
	}

	for _, c := range checks {
		if !strings.Contains(code, c.pattern) {
			t.Errorf("missing %s (%q) in output:\n%s", c.desc, c.pattern, code)
		}
	}

	if !strings.HasSuffix(code, "\n") {
		t.Error("expected trailing newline")
	}
}

/* TestArrowContextHoister_SecurityBridge verifies security bridge generation behavior */
func TestArrowContextHoister_SecurityBridge(t *testing.T) {
	tests := []struct {
		name            string
		sites           []ArrowCallSite
		wantBridgeFor   []string
		wantNoBridgeFor []string
	}{
		{
			name: "single_security_function",
			sites: []ArrowCallSite{
				{FunctionName: "getHTF", CallIndex: 1, ContextVar: "arrowCtx_getHTF_1", NeedsSecurity: true},
			},
			wantBridgeFor: []string{"arrowCtx_getHTF_1"},
		},
		{
			name: "multiple_security_functions",
			sites: []ArrowCallSite{
				{FunctionName: "getDaily", CallIndex: 1, ContextVar: "arrowCtx_getDaily_1", NeedsSecurity: true},
				{FunctionName: "getWeekly", CallIndex: 1, ContextVar: "arrowCtx_getWeekly_1", NeedsSecurity: true},
			},
			wantBridgeFor: []string{"arrowCtx_getDaily_1", "arrowCtx_getWeekly_1"},
		},
		{
			name: "mixed_security_and_non_security",
			sites: []ArrowCallSite{
				{FunctionName: "calcSMA", CallIndex: 1, ContextVar: "arrowCtx_calcSMA_1", NeedsSecurity: false},
				{FunctionName: "getHTF", CallIndex: 1, ContextVar: "arrowCtx_getHTF_1", NeedsSecurity: true},
				{FunctionName: "calcEMA", CallIndex: 1, ContextVar: "arrowCtx_calcEMA_1", NeedsSecurity: false},
			},
			wantBridgeFor:   []string{"arrowCtx_getHTF_1"},
			wantNoBridgeFor: []string{"arrowCtx_calcSMA_1", "arrowCtx_calcEMA_1"},
		},
		{
			name: "all_non_security",
			sites: []ArrowCallSite{
				{FunctionName: "a", CallIndex: 1, ContextVar: "arrowCtx_a_1", NeedsSecurity: false},
				{FunctionName: "b", CallIndex: 1, ContextVar: "arrowCtx_b_1", NeedsSecurity: false},
			},
			wantNoBridgeFor: []string{"arrowCtx_a_1", "arrowCtx_b_1"},
		},
	}

	hoister := NewArrowContextHoister("\t")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := hoister.GeneratePreLoopDeclarations(tt.sites)

			for _, varName := range tt.wantBridgeFor {
				bridgePattern := varName + ".SecurityContexts"
				if !strings.Contains(code, bridgePattern) {
					t.Errorf("expected security bridge for %s.\nGot:\n%s", varName, code)
				}

				mapperPattern := varName + ".SetBarMapper"
				if !strings.Contains(code, mapperPattern) {
					t.Errorf("expected SetBarMapper for %s.\nGot:\n%s", varName, code)
				}
			}

			for _, varName := range tt.wantNoBridgeFor {
				bridgePattern := varName + ".SecurityContexts"
				if strings.Contains(code, bridgePattern) {
					t.Errorf("non-security %s should NOT have bridge.\nGot:\n%s", varName, code)
				}
			}
		})
	}
}

/* TestArrowContextHoister_SecurityBridgeCodeStructure verifies generated bridge code format */
func TestArrowContextHoister_SecurityBridgeCodeStructure(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "getHTF", CallIndex: 1, ContextVar: "arrowCtx_getHTF_1", NeedsSecurity: true},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	requiredElements := []string{
		"context.NewArrowContext(ctx)",
		".SecurityContexts = securityContexts",
		"for secKey, mapper := range securityBarMappers",
		".SetBarMapper(secKey, mapper)",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(code, elem) {
			t.Errorf("missing structural element %q.\nGot:\n%s", elem, code)
		}
	}
}
