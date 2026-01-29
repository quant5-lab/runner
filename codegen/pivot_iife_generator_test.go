package codegen

import (
	"strings"
	"testing"
)

func TestPivotIIFEGenerator_DualPeriodBoundaries(t *testing.T) {
	tests := []struct {
		name            string
		generator       InlineTADualPeriodGenerator
		leftPeriod      PeriodExpression
		rightPeriod     PeriodExpression
		wantTotalWindow int
	}{
		{"pivothigh constant equal small", &PivotHighIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(1), 3},
		{"pivothigh constant equal medium", &PivotHighIIFEGenerator{}, NewConstantPeriod(5), NewConstantPeriod(5), 11},
		{"pivothigh constant equal large", &PivotHighIIFEGenerator{}, NewConstantPeriod(10), NewConstantPeriod(10), 21},
		{"pivothigh constant asymmetric left", &PivotHighIIFEGenerator{}, NewConstantPeriod(10), NewConstantPeriod(2), 13},
		{"pivothigh constant asymmetric right", &PivotHighIIFEGenerator{}, NewConstantPeriod(2), NewConstantPeriod(10), 13},
		{"pivotlow constant equal small", &PivotLowIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(1), 3},
		{"pivotlow constant equal medium", &PivotLowIIFEGenerator{}, NewConstantPeriod(5), NewConstantPeriod(5), 11},
		{"pivotlow constant equal large", &PivotLowIIFEGenerator{}, NewConstantPeriod(10), NewConstantPeriod(10), 21},
		{"pivotlow constant asymmetric left", &PivotLowIIFEGenerator{}, NewConstantPeriod(10), NewConstantPeriod(2), 13},
		{"pivotlow constant asymmetric right", &PivotLowIIFEGenerator{}, NewConstantPeriod(2), NewConstantPeriod(10), 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[i]",
				loopAccess: "data[i-j]",
			}

			code := tt.generator.GenerateDualPeriod(accessor, tt.leftPeriod, tt.rightPeriod, "test")

			expectedWarmup := tt.wantTotalWindow - 1
			expectedCheck := "ctx.BarIndex < " + intToString(expectedWarmup)
			if !strings.Contains(code, expectedCheck) {
				t.Errorf("Missing warmup check for window %d\nExpected: %s\nGenerated: %s",
					tt.wantTotalWindow, expectedCheck, code)
			}
		})
	}
}

func TestPivotIIFEGenerator_ComparisonLogic(t *testing.T) {
	tests := []struct {
		name                string
		generator           InlineTADualPeriodGenerator
		wantLeftComparison  string
		wantRightComparison string
	}{
		{
			name:                "pivothigh uses >= for rejection",
			generator:           &PivotHighIIFEGenerator{},
			wantLeftComparison:  ">= centerValue",
			wantRightComparison: ">= centerValue",
		},
		{
			name:                "pivotlow uses <= for rejection",
			generator:           &PivotLowIIFEGenerator{},
			wantLeftComparison:  "<= centerValue",
			wantRightComparison: "<= centerValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := &ConstantPeriod{value: 3}
			rightPeriod := &ConstantPeriod{value: 3}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.Contains(code, tt.wantLeftComparison) {
				t.Errorf("Missing left window comparison logic\nExpected substring: %s\nGenerated: %s",
					tt.wantLeftComparison, code)
			}

			if !strings.Contains(code, tt.wantRightComparison) {
				t.Errorf("Missing right window comparison logic\nExpected substring: %s\nGenerated: %s",
					tt.wantRightComparison, code)
			}
		})
	}
}

func TestPivotIIFEGenerator_CenterValueAccess(t *testing.T) {
	tests := []struct {
		name        string
		generator   InlineTADualPeriodGenerator
		leftPeriod  int
		rightPeriod int
	}{
		{"pivothigh symmetric", &PivotHighIIFEGenerator{}, 5, 5},
		{"pivothigh asymmetric", &PivotHighIIFEGenerator{}, 3, 7},
		{"pivotlow symmetric", &PivotLowIIFEGenerator{}, 5, 5},
		{"pivotlow asymmetric", &PivotLowIIFEGenerator{}, 7, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			centerMarker := "CENTER_VALUE_ACCESS"
			accessor := &mockDualPeriodAccessor{
				centerAccess: centerMarker,
			}

			leftPeriod := &ConstantPeriod{value: tt.leftPeriod}
			rightPeriod := &ConstantPeriod{value: tt.rightPeriod}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.Contains(code, centerMarker) {
				t.Errorf("Missing center value access\nExpected: %s\nGenerated: %s", centerMarker, code)
			}

			if !strings.Contains(code, "centerValue") {
				t.Error("Generated code should use 'centerValue' variable")
			}
		})
	}
}

func TestPivotIIFEGenerator_NaNHandling(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTADualPeriodGenerator
	}{
		{"pivothigh NaN check", &PivotHighIIFEGenerator{}},
		{"pivotlow NaN check", &PivotLowIIFEGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := &ConstantPeriod{value: 5}
			rightPeriod := &ConstantPeriod{value: 5}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.Contains(code, "math.IsNaN(centerValue)") {
				t.Error("Missing NaN check for centerValue")
			}

			if !strings.Contains(code, "return math.NaN()") {
				t.Error("Missing NaN return for invalid center value")
			}

			nanReturnCount := strings.Count(code, "math.NaN()")
			if nanReturnCount < 2 {
				t.Errorf("Should have at least 2 NaN returns (early exit + no pivot), got %d", nanReturnCount)
			}
		})
	}
}

func TestPivotIIFEGenerator_PivotDetectionFlow(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTADualPeriodGenerator
	}{
		{"pivothigh detection flow", &PivotHighIIFEGenerator{}},
		{"pivotlow detection flow", &PivotLowIIFEGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := &ConstantPeriod{value: 3}
			rightPeriod := &ConstantPeriod{value: 3}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.Contains(code, "isPivot := true") {
				t.Error("Missing isPivot initialization")
			}

			if !strings.Contains(code, "isPivot = false") {
				t.Error("Missing isPivot rejection logic")
			}

			if !strings.Contains(code, "if isPivot") {
				t.Error("Missing isPivot conditional check")
			}

			if !strings.Contains(code, "return centerValue") {
				t.Error("Missing centerValue return on pivot detection")
			}
		})
	}
}

func TestPivotIIFEGenerator_WindowScanning(t *testing.T) {
	tests := []struct {
		name           string
		generator      InlineTADualPeriodGenerator
		leftPeriod     int
		rightPeriod    int
		wantLeftScans  int
		wantRightScans int
	}{
		{"pivothigh 1-1", &PivotHighIIFEGenerator{}, 1, 1, 1, 1},
		{"pivothigh 3-3", &PivotHighIIFEGenerator{}, 3, 3, 3, 3},
		{"pivothigh 5-2", &PivotHighIIFEGenerator{}, 5, 2, 5, 2},
		{"pivothigh 2-5", &PivotHighIIFEGenerator{}, 2, 5, 2, 5},
		{"pivotlow 1-1", &PivotLowIIFEGenerator{}, 1, 1, 1, 1},
		{"pivotlow 3-3", &PivotLowIIFEGenerator{}, 3, 3, 3, 3},
		{"pivotlow 5-2", &PivotLowIIFEGenerator{}, 5, 2, 5, 2},
		{"pivotlow 2-5", &PivotLowIIFEGenerator{}, 2, 5, 2, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := &ConstantPeriod{value: tt.leftPeriod}
			rightPeriod := &ConstantPeriod{value: tt.rightPeriod}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			leftValCount := strings.Count(code, "leftVal")
			if leftValCount < tt.wantLeftScans {
				t.Errorf("Expected at least %d leftVal accesses, got %d", tt.wantLeftScans, leftValCount)
			}

			rightValCount := strings.Count(code, "rightVal")
			if rightValCount < tt.wantRightScans {
				t.Errorf("Expected at least %d rightVal accesses, got %d", tt.wantRightScans, rightValCount)
			}
		})
	}
}

func TestPivotIIFEGenerator_IIFEStructure(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTADualPeriodGenerator
	}{
		{"pivothigh structure", &PivotHighIIFEGenerator{}},
		{"pivotlow structure", &PivotLowIIFEGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := &ConstantPeriod{value: 5}
			rightPeriod := &ConstantPeriod{value: 5}
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.HasPrefix(code, "func() float64") {
				t.Error("Code should start with 'func() float64'")
			}

			if !strings.HasSuffix(code, "}()") {
				t.Error("Code should end with '}()' for IIFE pattern")
			}

			openBraces := strings.Count(code, "{")
			closeBraces := strings.Count(code, "}")
			if openBraces != closeBraces {
				t.Errorf("Mismatched braces: %d open, %d close", openBraces, closeBraces)
			}
		})
	}
}

func TestPivotIIFEGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		generator   InlineTADualPeriodGenerator
		leftPeriod  PeriodExpression
		rightPeriod PeriodExpression
		wantWindow  int
	}{
		{"pivothigh constant minimal 1-1", &PivotHighIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(1), 3},
		{"pivothigh constant single left", &PivotHighIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(5), 7},
		{"pivothigh constant single right", &PivotHighIIFEGenerator{}, NewConstantPeriod(5), NewConstantPeriod(1), 7},
		{"pivothigh constant large asymmetric", &PivotHighIIFEGenerator{}, NewConstantPeriod(20), NewConstantPeriod(2), 23},
		{"pivothigh runtime minimal 1-1", &PivotHighIIFEGenerator{}, NewRuntimePeriod("l1"), NewRuntimePeriod("r1"), -1},
		{"pivothigh runtime asymmetric", &PivotHighIIFEGenerator{}, NewRuntimePeriod("l20"), NewRuntimePeriod("r2"), -1},
		{"pivotlow constant minimal 1-1", &PivotLowIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(1), 3},
		{"pivotlow constant single left", &PivotLowIIFEGenerator{}, NewConstantPeriod(1), NewConstantPeriod(5), 7},
		{"pivotlow constant single right", &PivotLowIIFEGenerator{}, NewConstantPeriod(5), NewConstantPeriod(1), 7},
		{"pivotlow constant large asymmetric", &PivotLowIIFEGenerator{}, NewConstantPeriod(2), NewConstantPeriod(20), 23},
		{"pivotlow runtime minimal 1-1", &PivotLowIIFEGenerator{}, NewRuntimePeriod("l1"), NewRuntimePeriod("r1"), -1},
		{"pivotlow runtime asymmetric", &PivotLowIIFEGenerator{}, NewRuntimePeriod("l2"), NewRuntimePeriod("r20"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			code := tt.generator.GenerateDualPeriod(accessor, tt.leftPeriod, tt.rightPeriod, "test")

			if code == "" {
				t.Fatal("Generated code is empty")
			}

			if tt.wantWindow > 0 {
				expectedWarmup := tt.wantWindow - 1
				expectedCheck := "ctx.BarIndex < " + intToString(expectedWarmup)
				if !strings.Contains(code, expectedCheck) {
					t.Errorf("Warmup check incorrect for constant window %d\nExpected: %s\nGenerated: %s",
						tt.wantWindow, expectedCheck, code)
				}
			} else {
				if !strings.Contains(code, "totalWindow := leftBars + rightBars + 1") {
					t.Error("Runtime periods should calculate totalWindow dynamically")
				}
				if !strings.Contains(code, "ctx.BarIndex < totalWindow - 1") {
					t.Error("Runtime periods should use dynamic warmup check (format: 'ctx.BarIndex < totalWindow - 1')")
				}
			}
		})
	}
}

type mockDualPeriodAccessor struct {
	centerAccess string
}

func (m *mockDualPeriodAccessor) GenerateLoopValueAccess(loopVar string) string {
	if loopVar == intToString(5) {
		return m.centerAccess
	}
	return "data[" + loopVar + "]"
}

func (m *mockDualPeriodAccessor) GenerateInitialValueAccess(period int) string {
	return "data[0]"
}

func (m *mockDualPeriodAccessor) GenerateCurrentValueAccess() string {
	return "data[current]"
}

func (m *mockDualPeriodAccessor) GetBaseOffset() int {
	return 0
}

func TestPivotIIFEGenerator_PeriodModeStrategy(t *testing.T) {
	tests := []struct {
		name         string
		generator    InlineTADualPeriodGenerator
		leftPeriod   PeriodExpression
		rightPeriod  PeriodExpression
		wantUnrolled bool
	}{
		{"pivothigh constant-constant uses unrolled", &PivotHighIIFEGenerator{}, NewConstantPeriod(5), NewConstantPeriod(5), true},
		{"pivothigh runtime-runtime uses runtime loop", &PivotHighIIFEGenerator{}, NewRuntimePeriod("leftLen"), NewRuntimePeriod("rightLen"), false},
		{"pivothigh constant-runtime uses runtime loop", &PivotHighIIFEGenerator{}, NewConstantPeriod(5), NewRuntimePeriod("rightLen"), false},
		{"pivothigh runtime-constant uses runtime loop", &PivotHighIIFEGenerator{}, NewRuntimePeriod("leftLen"), NewConstantPeriod(5), false},
		{"pivotlow constant-constant uses unrolled", &PivotLowIIFEGenerator{}, NewConstantPeriod(3), NewConstantPeriod(3), true},
		{"pivotlow runtime-runtime uses runtime loop", &PivotLowIIFEGenerator{}, NewRuntimePeriod("leftLen"), NewRuntimePeriod("rightLen"), false},
		{"pivotlow constant-runtime uses runtime loop", &PivotLowIIFEGenerator{}, NewConstantPeriod(3), NewRuntimePeriod("rightLen"), false},
		{"pivotlow runtime-constant uses runtime loop", &PivotLowIIFEGenerator{}, NewRuntimePeriod("leftLen"), NewConstantPeriod(3), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			code := tt.generator.GenerateDualPeriod(accessor, tt.leftPeriod, tt.rightPeriod, "test")

			if code == "" {
				t.Fatal("Generated code is empty")
			}

			hasUnrolledLoop := !strings.Contains(code, "for j :=")
			hasRuntimeLoop := strings.Contains(code, "for j :=")

			if tt.wantUnrolled && !hasUnrolledLoop {
				t.Errorf("Expected unrolled loop generation (no 'for j :='), but found runtime loop")
			}

			if !tt.wantUnrolled && !hasRuntimeLoop {
				t.Errorf("Expected runtime loop generation (with 'for j :='), but found unrolled loop")
			}

			if !tt.wantUnrolled {
				if !strings.Contains(code, "leftBars") {
					t.Error("Runtime loop should have leftBars variable")
				}
				if !strings.Contains(code, "rightBars") {
					t.Error("Runtime loop should have rightBars variable")
				}
			}
		})
	}
}

func TestPivotIIFEGenerator_RuntimePeriodWarmup(t *testing.T) {
	tests := []struct {
		name      string
		generator InlineTADualPeriodGenerator
	}{
		{"pivothigh runtime warmup", &PivotHighIIFEGenerator{}},
		{"pivotlow runtime warmup", &PivotLowIIFEGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockWindowAccessor{
				initAccess: "data[0]",
				loopAccess: "data[j]",
			}

			leftPeriod := NewRuntimePeriod("leftLen")
			rightPeriod := NewRuntimePeriod("rightLen")
			code := tt.generator.GenerateDualPeriod(accessor, leftPeriod, rightPeriod, "test")

			if !strings.Contains(code, "totalWindow := leftBars + rightBars + 1") {
				t.Error("Runtime periods should calculate totalWindow dynamically")
			}

			if !strings.Contains(code, "ctx.BarIndex < totalWindow - 1") {
				t.Error("Runtime periods should use dynamic warmup check (format: 'ctx.BarIndex < totalWindow - 1')")
			}
		})
	}
}

func TestPivotIIFEGenerator_PineScriptSemantics(t *testing.T) {
	t.Run("pivothigh semantics", func(t *testing.T) {
		t.Log("ta.pivothigh(5, 5) examines 11-bar window: 5 left, center, 5 right")
		t.Log("Center value returned if all left bars <= center AND all right bars <= center")
		t.Log("Requires leftbars + rightbars bars before center (warmup = 10 for 5-5)")
	})

	t.Run("pivotlow semantics", func(t *testing.T) {
		t.Log("ta.pivotlow(3, 3) examines 7-bar window: 3 left, center, 3 right")
		t.Log("Center value returned if all left bars >= center AND all right bars >= center")
		t.Log("Requires leftbars + rightbars bars before center (warmup = 6 for 3-3)")
	})
}
