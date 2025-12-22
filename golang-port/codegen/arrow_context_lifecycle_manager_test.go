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
