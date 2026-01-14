package codegen

import (
	"strings"
	"testing"
)

func TestWindowFunctions_LoopBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTAIIFEGenerator
		period    int
	}{
		{"lowest period 1", &LowestIIFEGenerator{}, 1},
		{"lowest period 2", &LowestIIFEGenerator{}, 2},
		{"lowest period 10", &LowestIIFEGenerator{}, 10},
		{"lowest period 100", &LowestIIFEGenerator{}, 100},
		{"highest period 1", &HighestIIFEGenerator{}, 1},
		{"highest period 2", &HighestIIFEGenerator{}, 2},
		{"highest period 10", &HighestIIFEGenerator{}, 10},
		{"highest period 100", &HighestIIFEGenerator{}, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[i-period+1]",
				loopAccess: "data[i-j]",
			}

			period := &ConstantPeriod{value: tt.period}
			code := tt.generator.Generate(accessor, period, "test")

			if !strings.Contains(code, "j >= 0") {
				t.Errorf("Loop must use 'j >= 0' to include current bar\nPeriod: %d\nGenerated: %s", tt.period, code)
			}

			if strings.Contains(code, "j > 0") {
				t.Errorf("Loop incorrectly uses 'j > 0' which excludes current bar\nPeriod: %d\nGenerated: %s", tt.period, code)
			}

			expectedStart := strings.Contains(code, "j := "+intToString(tt.period-1))
			if !expectedStart {
				t.Errorf("Loop should start at j=%d for period %d\nGenerated: %s", tt.period-1, tt.period, code)
			}
		})
	}
}

func TestWindowFunctions_WarmupCheck(t *testing.T) {
	tests := []struct {
		name              string
		generator         InlineTAIIFEGenerator
		period            int
		checkBar          int
		expectWarmupCheck bool
	}{
		{"lowest single bar", &LowestIIFEGenerator{}, 1, 0, false},
		{"lowest two bars", &LowestIIFEGenerator{}, 2, 1, true},
		{"lowest ten bars", &LowestIIFEGenerator{}, 10, 9, true},
		{"highest single bar", &HighestIIFEGenerator{}, 1, 0, false},
		{"highest two bars", &HighestIIFEGenerator{}, 2, 1, true},
		{"highest ten bars", &HighestIIFEGenerator{}, 10, 9, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			period := &ConstantPeriod{value: tt.period}
			code := tt.generator.Generate(accessor, period, "test")

			hasWarmupCheck := strings.Contains(code, "ctx.BarIndex <")

			if tt.expectWarmupCheck && !hasWarmupCheck {
				t.Errorf("Missing warmup check for period %d\nGenerated: %s", tt.period, code)
			}

			if !tt.expectWarmupCheck && hasWarmupCheck {
				t.Errorf("Period %d should not have warmup check (no bars needed before current)\nGenerated: %s", tt.period, code)
			}

			if tt.expectWarmupCheck {
				expectedCheck := "ctx.BarIndex < " + intToString(tt.checkBar)
				if !strings.Contains(code, expectedCheck) {
					t.Errorf("Warmup check should be 'ctx.BarIndex < %d' for period %d\nGenerated: %s", tt.checkBar, tt.period, code)
				}
			}
		})
	}
}

func TestWindowFunctions_InitialValueAccess(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTAIIFEGenerator
		period    int
	}{
		{"lowest uses initial value", &LowestIIFEGenerator{}, 5},
		{"highest uses initial value", &HighestIIFEGenerator{}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initValue := "INITIAL_ACCESS"
			loopValue := "LOOP_ACCESS"

			accessor := &mockWindowAccessor{
				initAccess: initValue,
				loopAccess: loopValue,
			}

			period := &ConstantPeriod{value: tt.period}
			code := tt.generator.Generate(accessor, period, "test")

			if !strings.Contains(code, initValue) {
				t.Errorf("Missing initial value access\nExpected: %s\nGenerated: %s", initValue, code)
			}

			if !strings.Contains(code, loopValue) {
				t.Errorf("Missing loop value access\nExpected: %s\nGenerated: %s", loopValue, code)
			}
		})
	}
}

func TestWindowFunctions_ComparisonLogic(t *testing.T) {
	tests := []struct {
		name       string
		generator  InlineTAIIFEGenerator
		varName    string
		comparison string
	}{
		{"lowest uses less than", &LowestIIFEGenerator{}, "lowest", "v < lowest"},
		{"highest uses greater than", &HighestIIFEGenerator{}, "highest", "v > highest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			period := &ConstantPeriod{value: 5}
			code := tt.generator.Generate(accessor, period, "test")

			expectedDecl := tt.varName + " :="
			if !strings.Contains(code, expectedDecl) {
				t.Errorf("Missing %s variable declaration\nGenerated: %s", tt.varName, code)
			}

			if !strings.Contains(code, tt.comparison) {
				t.Errorf("Missing comparison '%s'\nGenerated: %s", tt.comparison, code)
			}

			expectedReturn := "return " + tt.varName
			if !strings.Contains(code, expectedReturn) {
				t.Errorf("Missing return statement for %s\nGenerated: %s", tt.varName, code)
			}
		})
	}
}

func TestWindowFunctions_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTAIIFEGenerator
		period    int
		shouldErr bool
	}{
		{"lowest period 1 (current bar only)", &LowestIIFEGenerator{}, 1, false},
		{"highest period 1 (current bar only)", &HighestIIFEGenerator{}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			period := &ConstantPeriod{value: tt.period}
			code := tt.generator.Generate(accessor, period, "test")

			if code == "" {
				if !tt.shouldErr {
					t.Error("Expected code generation, got empty string")
				}
				return
			}

			if tt.period == 1 {
				if !strings.Contains(code, "j := 0") {
					t.Errorf("Period 1 should start loop at j=0\nGenerated: %s", code)
				}
				if !strings.Contains(code, "j >= 0") {
					t.Errorf("Period 1 must include j=0 iteration\nGenerated: %s", code)
				}
			}
		})
	}
}

func TestWindowFunctions_CodeStructure(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTAIIFEGenerator
	}{
		{"lowest structure", &LowestIIFEGenerator{}},
		{"highest structure", &HighestIIFEGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			period := &ConstantPeriod{value: 10}
			code := tt.generator.Generate(accessor, period, "test")

			if !strings.Contains(code, "func()") {
				t.Error("Generated code should be IIFE with func() wrapper")
			}

			if !strings.Contains(code, "return") {
				t.Error("Generated code should have return statement")
			}

			if !strings.Contains(code, "for j :=") {
				t.Error("Generated code should have for loop with j variable")
			}

			if !strings.Contains(code, "if ctx.BarIndex <") {
				t.Error("Generated code should have warmup check for period 10")
			}

			if !strings.Contains(code, "math.NaN()") {
				t.Error("Generated code should return NaN before warmup period")
			}
		})
	}
}

func TestWindowFunctions_PineScriptSemantics(t *testing.T) {
	t.Run("lowest semantics", func(t *testing.T) {
		/* PineScript lowest(source, length) returns minimum over length bars including current */
		t.Log("lowest(2) window includes current bar")
		t.Log("Example: bars [605.8, 604.2] at indices [i-1, i] → minimum = 604.2")
		t.Log("Loop: j=1 (bar[i-1]), j=0 (bar[i])")
	})

	t.Run("highest semantics", func(t *testing.T) {
		/* PineScript highest(source, length) returns maximum over length bars including current */
		t.Log("highest(2) window includes current bar")
		t.Log("Example: bars [610.0, 615.0] at indices [i-1, i] → maximum = 615.0")
		t.Log("Loop: j=1 (bar[i-1]), j=0 (bar[i])")
	})
}

type mockWindowAccessor struct {
	initAccess string
	loopAccess string
}

func (m *mockWindowAccessor) GenerateInitialValueAccess(period int) string {
	return m.initAccess
}

func (m *mockWindowAccessor) GenerateLoopValueAccess(loopVar string) string {
	return m.loopAccess
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}
