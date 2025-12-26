package codegen

import (
	"strings"
	"testing"
)

func TestArrowContextLifecycleManager_AllocateContextVariable(t *testing.T) {
	tests := []struct {
		name         string
		funcName     string
		callCount    int
		wantNames    []string
		wantInstance int
	}{
		{
			name:         "single allocation",
			funcName:     "adx",
			callCount:    1,
			wantNames:    []string{"arrowCtx_adx_1"},
			wantInstance: 1,
		},
		{
			name:         "multiple allocations same function",
			funcName:     "adx",
			callCount:    3,
			wantNames:    []string{"arrowCtx_adx_1", "arrowCtx_adx_2", "arrowCtx_adx_3"},
			wantInstance: 3,
		},
		{
			name:         "different function names",
			funcName:     "dirmov",
			callCount:    2,
			wantNames:    []string{"arrowCtx_dirmov_1", "arrowCtx_dirmov_2"},
			wantInstance: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewArrowContextLifecycleManager()

			gotNames := make([]string, tt.callCount)
			for i := 0; i < tt.callCount; i++ {
				gotNames[i] = manager.AllocateContextVariable(tt.funcName)
			}

			for i, want := range tt.wantNames {
				if gotNames[i] != want {
					t.Errorf("AllocateContextVariable() call %d = %q, want %q", i+1, gotNames[i], want)
				}
			}

			gotInstance := manager.GetInstanceCount(tt.funcName)
			if gotInstance != tt.wantInstance {
				t.Errorf("GetInstanceCount() = %d, want %d", gotInstance, tt.wantInstance)
			}
		})
	}
}

func TestArrowContextLifecycleManager_UniqueNamesAcrossFunctions(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	adx1 := manager.AllocateContextVariable("adx")
	dirmov1 := manager.AllocateContextVariable("dirmov")
	adx2 := manager.AllocateContextVariable("adx")
	dirmov2 := manager.AllocateContextVariable("dirmov")

	expected := map[string]string{
		adx1:    "arrowCtx_adx_1",
		dirmov1: "arrowCtx_dirmov_1",
		adx2:    "arrowCtx_adx_2",
		dirmov2: "arrowCtx_dirmov_2",
	}

	for got, want := range expected {
		if got != want {
			t.Errorf("Expected %q, got %q", want, got)
		}
	}

	if manager.GetInstanceCount("adx") != 2 {
		t.Errorf("adx instance count = %d, want 2", manager.GetInstanceCount("adx"))
	}
	if manager.GetInstanceCount("dirmov") != 2 {
		t.Errorf("dirmov instance count = %d, want 2", manager.GetInstanceCount("dirmov"))
	}
}

func TestArrowContextLifecycleManager_Reset(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	manager.AllocateContextVariable("adx")
	manager.AllocateContextVariable("adx")

	if manager.GetInstanceCount("adx") != 2 {
		t.Fatalf("Before reset: adx count = %d, want 2", manager.GetInstanceCount("adx"))
	}

	manager.Reset()

	if manager.GetInstanceCount("adx") != 0 {
		t.Errorf("After reset: adx count = %d, want 0", manager.GetInstanceCount("adx"))
	}

	name := manager.AllocateContextVariable("adx")
	if name != "arrowCtx_adx_1" {
		t.Errorf("After reset: first allocation = %q, want %q", name, "arrowCtx_adx_1")
	}
}

func TestArrowContextLifecycleManager_NoRedeclaration(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	names := make([]string, 5)
	for i := 0; i < 5; i++ {
		names[i] = manager.AllocateContextVariable("test_func")
	}

	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("Duplicate variable name generated: %q", name)
		}
		seen[name] = true

		if !strings.HasPrefix(name, "arrowCtx_test_func_") {
			t.Errorf("Invalid name format: %q", name)
		}
	}
}

func TestArrowContextLifecycleManager_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		funcName     string
		allocations  int
		expectPrefix string
		expectCount  int
	}{
		{
			name:         "zero allocations",
			funcName:     "unused",
			allocations:  0,
			expectPrefix: "",
			expectCount:  0,
		},
		{
			name:         "single character function name",
			funcName:     "a",
			allocations:  2,
			expectPrefix: "arrowCtx_a_",
			expectCount:  2,
		},
		{
			name:         "function name with underscores",
			funcName:     "calc_moving_avg",
			allocations:  3,
			expectPrefix: "arrowCtx_calc_moving_avg_",
			expectCount:  3,
		},
		{
			name:         "function name with numbers",
			funcName:     "func123",
			allocations:  2,
			expectPrefix: "arrowCtx_func123_",
			expectCount:  2,
		},
		{
			name:         "very long function name",
			funcName:     "calculateExponentialMovingAverageWithVolatilityAdjustment",
			allocations:  2,
			expectPrefix: "arrowCtx_calculateExponentialMovingAverageWithVolatilityAdjustment_",
			expectCount:  2,
		},
		{
			name:         "many sequential allocations",
			funcName:     "repeated",
			allocations:  50,
			expectPrefix: "arrowCtx_repeated_",
			expectCount:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewArrowContextLifecycleManager()

			for i := 0; i < tt.allocations; i++ {
				name := manager.AllocateContextVariable(tt.funcName)
				if tt.allocations > 0 && !strings.HasPrefix(name, tt.expectPrefix) {
					t.Errorf("Allocation %d: expected prefix %q, got %q", i+1, tt.expectPrefix, name)
				}
			}

			if got := manager.GetInstanceCount(tt.funcName); got != tt.expectCount {
				t.Errorf("GetInstanceCount() = %d, want %d", got, tt.expectCount)
			}
		})
	}
}

func TestArrowContextLifecycleManager_StateTransitions(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	manager.AllocateContextVariable("func1")
	manager.AllocateContextVariable("func2")

	manager.Reset()

	name1 := manager.AllocateContextVariable("func1")
	if name1 != "arrowCtx_func1_1" {
		t.Errorf("After reset: expected arrowCtx_func1_1, got %q", name1)
	}

	manager.Reset()
	manager.Reset()

	name2 := manager.AllocateContextVariable("func1")
	if name2 != "arrowCtx_func1_1" {
		t.Errorf("After multiple resets: expected arrowCtx_func1_1, got %q", name2)
	}
}

func TestArrowContextLifecycleManager_BoundaryConditions(t *testing.T) {
	t.Run("uninitialized state query", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()
		count := manager.GetInstanceCount("never_allocated")
		if count != 0 {
			t.Errorf("Unallocated function count = %d, want 0", count)
		}
	})

	t.Run("large counter values", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()
		for i := 0; i < 1000; i++ {
			name := manager.AllocateContextVariable("stress_test")
			if !strings.HasPrefix(name, "arrowCtx_stress_test_") {
				t.Errorf("Allocation %d: invalid format %q", i+1, name)
				break
			}
		}
		if manager.GetInstanceCount("stress_test") != 1000 {
			t.Errorf("After 1000 allocations: count = %d, want 1000", manager.GetInstanceCount("stress_test"))
		}
	})

	t.Run("reset mid-sequence", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()
		manager.AllocateContextVariable("partial")
		manager.AllocateContextVariable("partial")
		manager.AllocateContextVariable("partial")

		manager.Reset()

		manager.AllocateContextVariable("other")
		name := manager.AllocateContextVariable("partial")

		if name != "arrowCtx_partial_1" {
			t.Errorf("After mid-sequence reset: expected arrowCtx_partial_1, got %q", name)
		}
	})
}

func TestArrowContextLifecycleManager_CaseSensitivity(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	lower := manager.AllocateContextVariable("func")
	upper := manager.AllocateContextVariable("FUNC")
	mixed := manager.AllocateContextVariable("Func")

	if lower == upper || lower == mixed || upper == mixed {
		t.Error("Function names should be case-sensitive")
	}

	if !strings.Contains(lower, "func") {
		t.Errorf("Expected lowercase 'func', got %q", lower)
	}
	if !strings.Contains(upper, "FUNC") {
		t.Errorf("Expected uppercase 'FUNC', got %q", upper)
	}
	if !strings.Contains(mixed, "Func") {
		t.Errorf("Expected mixed case 'Func', got %q", mixed)
	}
}

func TestArrowContextLifecycleManager_MarkAsHoisted(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contextVar := "arrowCtx_adx_1"

	if manager.IsHoisted(contextVar) {
		t.Error("Context should not be hoisted before marking")
	}

	manager.MarkAsHoisted(contextVar)

	if !manager.IsHoisted(contextVar) {
		t.Error("Context should be hoisted after marking")
	}
}

func TestArrowContextLifecycleManager_IsHoisted_DefaultState(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contexts := []string{
		"arrowCtx_adx_1",
		"arrowCtx_rma_1",
		"arrowCtx_ema_1",
		"arrowCtx_dirmov_1",
	}

	for _, ctx := range contexts {
		if manager.IsHoisted(ctx) {
			t.Errorf("Context %q should not be hoisted by default", ctx)
		}
	}
}

func TestArrowContextLifecycleManager_HoistingMultipleContexts(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contexts := []string{
		"arrowCtx_adx_1",
		"arrowCtx_adx_2",
		"arrowCtx_rma_1",
		"arrowCtx_ema_1",
	}

	for _, ctx := range contexts {
		manager.MarkAsHoisted(ctx)
	}

	for _, ctx := range contexts {
		if !manager.IsHoisted(ctx) {
			t.Errorf("Context %q should be hoisted", ctx)
		}
	}
}

func TestArrowContextLifecycleManager_HoistingIndependence(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	manager.MarkAsHoisted("arrowCtx_adx_1")
	manager.MarkAsHoisted("arrowCtx_adx_3")

	if !manager.IsHoisted("arrowCtx_adx_1") {
		t.Error("arrowCtx_adx_1 should be hoisted")
	}

	if manager.IsHoisted("arrowCtx_adx_2") {
		t.Error("arrowCtx_adx_2 should NOT be hoisted (not marked)")
	}

	if !manager.IsHoisted("arrowCtx_adx_3") {
		t.Error("arrowCtx_adx_3 should be hoisted")
	}
}

func TestArrowContextLifecycleManager_ResetClearsHoistedState(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contexts := []string{
		"arrowCtx_func1_1",
		"arrowCtx_func2_1",
		"arrowCtx_func3_1",
	}

	for _, ctx := range contexts {
		manager.MarkAsHoisted(ctx)
	}

	for _, ctx := range contexts {
		if !manager.IsHoisted(ctx) {
			t.Fatalf("Context %q should be hoisted before reset", ctx)
		}
	}

	manager.Reset()

	for _, ctx := range contexts {
		if manager.IsHoisted(ctx) {
			t.Errorf("Context %q should NOT be hoisted after reset", ctx)
		}
	}
}

func TestArrowContextLifecycleManager_HoistingDuplicateMarking(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contextVar := "arrowCtx_test_1"

	manager.MarkAsHoisted(contextVar)
	manager.MarkAsHoisted(contextVar)
	manager.MarkAsHoisted(contextVar)

	if !manager.IsHoisted(contextVar) {
		t.Error("Context should remain hoisted after duplicate marking")
	}
}

func TestArrowContextLifecycleManager_HoistingWithAllocation(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	ctx1 := manager.AllocateContextVariable("adx")
	ctx2 := manager.AllocateContextVariable("adx")

	manager.MarkAsHoisted(ctx1)

	if !manager.IsHoisted(ctx1) {
		t.Errorf("Context %q should be hoisted", ctx1)
	}

	if manager.IsHoisted(ctx2) {
		t.Errorf("Context %q should NOT be hoisted (not marked)", ctx2)
	}

	if manager.GetInstanceCount("adx") != 2 {
		t.Errorf("Expected 2 adx instances, got %d", manager.GetInstanceCount("adx"))
	}
}

func TestArrowContextLifecycleManager_HoistingCaseSensitivity(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	contexts := []string{
		"arrowCtx_func_1",
		"arrowCtx_Func_1",
		"arrowCtx_FUNC_1",
	}

	manager.MarkAsHoisted(contexts[0])

	if !manager.IsHoisted(contexts[0]) {
		t.Errorf("Context %q should be hoisted", contexts[0])
	}

	if manager.IsHoisted(contexts[1]) {
		t.Errorf("Context %q should NOT be hoisted (different case)", contexts[1])
	}

	if manager.IsHoisted(contexts[2]) {
		t.Errorf("Context %q should NOT be hoisted (different case)", contexts[2])
	}
}

func TestArrowContextLifecycleManager_HoistingStateIntegration(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	ctx1 := manager.AllocateContextVariable("adx")
	ctx2 := manager.AllocateContextVariable("rma")
	ctx3 := manager.AllocateContextVariable("adx")

	manager.MarkAsHoisted(ctx1)
	manager.MarkAsHoisted(ctx3)

	hoistedCount := 0
	nonHoistedCount := 0

	for _, ctx := range []string{ctx1, ctx2, ctx3} {
		if manager.IsHoisted(ctx) {
			hoistedCount++
		} else {
			nonHoistedCount++
		}
	}

	if hoistedCount != 2 {
		t.Errorf("Expected 2 hoisted contexts, got %d", hoistedCount)
	}

	if nonHoistedCount != 1 {
		t.Errorf("Expected 1 non-hoisted context, got %d", nonHoistedCount)
	}

	if manager.GetInstanceCount("adx") != 2 {
		t.Errorf("Expected 2 adx instances, got %d", manager.GetInstanceCount("adx"))
	}

	if manager.GetInstanceCount("rma") != 1 {
		t.Errorf("Expected 1 rma instance, got %d", manager.GetInstanceCount("rma"))
	}
}

func TestArrowContextLifecycleManager_HoistingWithResetAndReallocation(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	ctx1 := manager.AllocateContextVariable("func")
	manager.MarkAsHoisted(ctx1)

	if !manager.IsHoisted(ctx1) {
		t.Fatal("Context should be hoisted before reset")
	}

	manager.Reset()

	if manager.IsHoisted(ctx1) {
		t.Error("Context should NOT be hoisted after reset")
	}

	ctx2 := manager.AllocateContextVariable("func")

	if ctx1 != ctx2 {
		t.Logf("Note: After reset, same function gets same context name: %q == %q", ctx1, ctx2)
	}

	manager.MarkAsHoisted(ctx2)

	if !manager.IsHoisted(ctx2) {
		t.Error("Re-allocated context should be hoisted after marking")
	}
}

func TestArrowContextLifecycleManager_HoistingBoundaryConditions(t *testing.T) {
	t.Run("empty context name", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()
		manager.MarkAsHoisted("")

		if !manager.IsHoisted("") {
			t.Error("Empty string should be markable as hoisted")
		}
	})

	t.Run("very long context name", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()
		longName := "arrowCtx_" + strings.Repeat("veryLongFunctionName", 10) + "_1"

		manager.MarkAsHoisted(longName)

		if !manager.IsHoisted(longName) {
			t.Error("Long context name should be markable as hoisted")
		}
	})

	t.Run("many hoisted contexts", func(t *testing.T) {
		manager := NewArrowContextLifecycleManager()

		for i := 1; i <= 1000; i++ {
			ctxName := "arrowCtx_func_" + string(rune('0'+i%10))
			manager.MarkAsHoisted(ctxName)
		}

		for i := 1; i <= 1000; i++ {
			ctxName := "arrowCtx_func_" + string(rune('0'+i%10))
			if !manager.IsHoisted(ctxName) {
				t.Errorf("Context %q should be hoisted", ctxName)
				break
			}
		}
	})
}

func TestArrowContextLifecycleManager_HoistingPreventsDuplicateAllocation(t *testing.T) {
	manager := NewArrowContextLifecycleManager()

	ctx1 := manager.AllocateContextVariable("adx")

	manager.MarkAsHoisted(ctx1)

	if !manager.IsHoisted(ctx1) {
		t.Fatal("Context should be hoisted")
	}

	ctx2 := manager.AllocateContextVariable("adx")

	if ctx1 == ctx2 {
		t.Error("Second allocation should produce different context name (hoisting doesn't prevent allocation)")
	}

	if manager.GetInstanceCount("adx") != 2 {
		t.Errorf("Expected 2 instances (hoisting is metadata, not allocation control), got %d", manager.GetInstanceCount("adx"))
	}
}
