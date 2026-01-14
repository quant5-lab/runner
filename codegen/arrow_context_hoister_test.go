package codegen

import (
	"strings"
	"testing"
)

func TestArrowContextHoister_EmptyCallSites(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	code := hoister.GeneratePreLoopDeclarations([]ArrowCallSite{})

	if code != "" {
		t.Errorf("Expected empty code for no call sites, got %q", code)
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
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, code)
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

	expectedLines := []string{
		"\tarrowCtx_adx_1 := context.NewArrowContext(ctx)",
		"\tarrowCtx_rma_1 := context.NewArrowContext(ctx)",
		"\tarrowCtx_ema_1 := context.NewArrowContext(ctx)",
	}

	for _, line := range expectedLines {
		if !strings.Contains(code, line) {
			t.Errorf("Expected code to contain:\n%s\n\nGot:\n%s", line, code)
		}
	}
}

func TestArrowContextHoister_IndentationVariations(t *testing.T) {
	tests := []struct {
		name        string
		indentation string
		expectStart string
	}{
		{
			name:        "no indentation",
			indentation: "",
			expectStart: "arrowCtx_func_1 := context.NewArrowContext(ctx)",
		},
		{
			name:        "single tab",
			indentation: "\t",
			expectStart: "\tarrowCtx_func_1 := context.NewArrowContext(ctx)",
		},
		{
			name:        "two tabs",
			indentation: "\t\t",
			expectStart: "\t\tarrowCtx_func_1 := context.NewArrowContext(ctx)",
		},
		{
			name:        "four spaces",
			indentation: "    ",
			expectStart: "    arrowCtx_func_1 := context.NewArrowContext(ctx)",
		},
		{
			name:        "eight spaces",
			indentation: "        ",
			expectStart: "        arrowCtx_func_1 := context.NewArrowContext(ctx)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hoister := NewArrowContextHoister(tt.indentation)
			sites := []ArrowCallSite{
				{FunctionName: "func", CallIndex: 1, ContextVar: "arrowCtx_func_1"},
			}

			code := hoister.GeneratePreLoopDeclarations(sites)

			if !strings.HasPrefix(code, tt.expectStart) {
				t.Errorf("Expected code to start with:\n%q\n\nGot:\n%q", tt.expectStart, code)
			}
		})
	}
}

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
		t.Fatalf("Expected 3 lines of code, got %d", len(lines))
	}

	expectedOrder := []string{"arrowCtx_first_1", "arrowCtx_second_1", "arrowCtx_third_1"}
	for i, expectedVar := range expectedOrder {
		if !strings.Contains(lines[i], expectedVar) {
			t.Errorf("Line %d: expected to contain %q, got %q", i, expectedVar, lines[i])
		}
	}
}

func TestArrowContextHoister_SameFunctionMultipleInstances(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "calc", CallIndex: 1, ContextVar: "arrowCtx_calc_1"},
		{FunctionName: "calc", CallIndex: 2, ContextVar: "arrowCtx_calc_2"},
		{FunctionName: "calc", CallIndex: 3, ContextVar: "arrowCtx_calc_3"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	expectedVars := []string{"arrowCtx_calc_1", "arrowCtx_calc_2", "arrowCtx_calc_3"}
	for _, varName := range expectedVars {
		if !strings.Contains(code, varName) {
			t.Errorf("Expected code to contain %q\n\nGot:\n%s", varName, code)
		}
	}

	lines := strings.Split(strings.TrimSpace(code), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 distinct declarations, got %d", len(lines))
	}
}

func TestArrowContextHoister_CodeFormat(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "test", CallIndex: 1, ContextVar: "arrowCtx_test_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	if !strings.Contains(code, ":=") {
		t.Error("Expected short variable declaration (:=)")
	}
	if !strings.Contains(code, "context.NewArrowContext") {
		t.Error("Expected context.NewArrowContext constructor call")
	}
	if !strings.Contains(code, "(ctx)") {
		t.Error("Expected ctx parameter to NewArrowContext")
	}
	if !strings.HasSuffix(code, "\n") {
		t.Error("Expected code to end with newline")
	}
}

func TestArrowContextHoister_NoCodeInjection(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "func'; DROP TABLE users; --", CallIndex: 1, ContextVar: "arrowCtx_malicious_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	if strings.Contains(code, "DROP TABLE") {
		t.Error("Code injection vulnerability: malicious function name included in output")
	}

	if !strings.Contains(code, "arrowCtx_malicious_1") {
		t.Error("Expected sanitized context variable name")
	}
}

func TestArrowContextHoister_UniqueDeclarations(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "func1", CallIndex: 1, ContextVar: "arrowCtx_func1_1"},
		{FunctionName: "func2", CallIndex: 1, ContextVar: "arrowCtx_func2_1"},
		{FunctionName: "func3", CallIndex: 1, ContextVar: "arrowCtx_func3_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	lines := strings.Split(strings.TrimSpace(code), "\n")
	seen := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if seen[trimmed] {
			t.Errorf("Duplicate declaration found: %q", trimmed)
		}
		seen[trimmed] = true
	}
}

func TestArrowContextHoister_LargeScaleGeneration(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{}

	for i := 1; i <= 100; i++ {
		sites = append(sites, ArrowCallSite{
			FunctionName: "func",
			CallIndex:    i,
			ContextVar:   "arrowCtx_func_" + string(rune('0'+i%10)),
		})
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	lines := strings.Split(strings.TrimSpace(code), "\n")
	if len(lines) != 100 {
		t.Errorf("Expected 100 declarations, got %d", len(lines))
	}

	for i, line := range lines {
		if !strings.Contains(line, ":=") {
			t.Errorf("Line %d missing declaration operator: %q", i, line)
		}
		if !strings.Contains(line, "context.NewArrowContext(ctx)") {
			t.Errorf("Line %d missing constructor call: %q", i, line)
		}
	}
}

func TestArrowContextHoister_ConsistentFormatting(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "a", CallIndex: 1, ContextVar: "arrowCtx_a_1"},
		{FunctionName: "b", CallIndex: 1, ContextVar: "arrowCtx_b_1"},
		{FunctionName: "c", CallIndex: 1, ContextVar: "arrowCtx_c_1"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)
	lines := strings.Split(strings.TrimRight(code, "\n"), "\n")

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	expectedIndent := "\t"
	for i, line := range lines {
		if !strings.HasPrefix(line, expectedIndent) {
			t.Errorf("Line %d does not have expected indentation, got: %q", i, line)
		}

		if !strings.Contains(line, ":=") {
			t.Errorf("Line %d missing declaration operator", i)
		}

		if !strings.Contains(line, "context.NewArrowContext(ctx)") {
			t.Errorf("Line %d missing constructor call", i)
		}
	}
}

func TestArrowContextHoister_NilCallSitesList(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	code := hoister.GeneratePreLoopDeclarations(nil)

	if code != "" {
		t.Errorf("Expected empty code for nil call sites, got %q", code)
	}
}

func TestArrowContextHoister_ContextVariableUniqueness(t *testing.T) {
	hoister := NewArrowContextHoister("\t")
	sites := []ArrowCallSite{
		{FunctionName: "adx", CallIndex: 1, ContextVar: "arrowCtx_adx_1"},
		{FunctionName: "adx", CallIndex: 2, ContextVar: "arrowCtx_adx_2"},
		{FunctionName: "rma", CallIndex: 1, ContextVar: "arrowCtx_rma_1"},
		{FunctionName: "rma", CallIndex: 2, ContextVar: "arrowCtx_rma_2"},
	}

	code := hoister.GeneratePreLoopDeclarations(sites)

	contextVars := []string{"arrowCtx_adx_1", "arrowCtx_adx_2", "arrowCtx_rma_1", "arrowCtx_rma_2"}
	varCounts := make(map[string]int)

	for _, varName := range contextVars {
		count := strings.Count(code, varName)
		varCounts[varName] = count

		if count != 1 {
			t.Errorf("Context variable %q appears %d times, expected 1", varName, count)
		}
	}
}
