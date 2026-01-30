package codegen

import (
	"strings"
	"testing"
)

type mockLinregAccessor struct {
	loopAccess string
}

func (m *mockLinregAccessor) GenerateLoopValueAccess(loopVar string) string {
	return strings.ReplaceAll(m.loopAccess, "j", loopVar)
}

func (m *mockLinregAccessor) GenerateInitialValueAccess(period int) string {
	return "initialValue"
}

func (m *mockLinregAccessor) GenerateCurrentValueAccess() string {
	return "currentValue"
}

func (m *mockLinregAccessor) GetBaseOffset() int {
	return 0
}

func TestLinregIIFEGenerator_PeriodVariations(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	tests := []struct {
		name        string
		period      int
		wantLoop    string
		wantWarmup  int
		wantFormula string
	}{
		{
			name:        "period_1",
			period:      1,
			wantLoop:    "for j := 0; j < 1",
			wantWarmup:  0,
			wantFormula: "float64(0)",
		},
		{
			name:        "period_5",
			period:      5,
			wantLoop:    "for j := 0; j < 5",
			wantWarmup:  4,
			wantFormula: "float64(4)",
		},
		{
			name:        "period_14",
			period:      14,
			wantLoop:    "for j := 0; j < 14",
			wantWarmup:  13,
			wantFormula: "float64(13)",
		},
		{
			name:        "period_50",
			period:      50,
			wantLoop:    "for j := 0; j < 50",
			wantWarmup:  49,
			wantFormula: "float64(49)",
		},
		{
			name:        "period_100",
			period:      100,
			wantLoop:    "for j := 0; j < 100",
			wantWarmup:  99,
			wantFormula: "float64(99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
			periodExpr := NewConstantPeriod(tt.period)

			code, ok := registry.Generate("ta.linreg", accessor, periodExpr, "test")
			if !ok {
				t.Fatal("IIFE generation failed")
			}

			if !strings.Contains(code, tt.wantLoop) {
				t.Errorf("Generated code missing loop: %q\nGot: %s", tt.wantLoop, code)
			}

			if !strings.Contains(code, tt.wantFormula) {
				t.Errorf("Generated code missing formula multiplier: %q", tt.wantFormula)
			}
		})
	}
}

/* TestLinregIIFEGenerator_AccessorTypes validates generation with different accessor types */
func TestLinregIIFEGenerator_AccessorTypes(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	period := 10

	tests := []struct {
		name     string
		accessor AccessGenerator
		wantY    string
	}{
		{
			name:     "bar_field_close",
			accessor: &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"},
			wantY:    "ctx.Data[i-10 - j - 1].Close",
		},
		{
			name:     "bar_field_high",
			accessor: &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].High"},
			wantY:    "ctx.Data[i-10 - j - 1].High",
		},
		{
			name:     "bar_field_low",
			accessor: &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Low"},
			wantY:    "ctx.Data[i-10 - j - 1].Low",
		},
		{
			name:     "bar_field_open",
			accessor: &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Open"},
			wantY:    "ctx.Data[i-10 - j - 1].Open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := registry.Generate("ta.linreg", tt.accessor, NewConstantPeriod(period), "test")
			if !ok {
				t.Fatal("IIFE generation failed")
			}

			if !strings.Contains(code, "y :=") {
				t.Error("Generated code missing y variable assignment")
			}

			if !strings.Contains(code, "ctx.Data[i-") {
				t.Error("Generated code should use ctx.Data for bar field access")
			}
		})
	}
}

func TestLinregIIFEGenerator_AlgorithmStructure(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
	period := 20

	code, ok := registry.Generate("ta.linreg", accessor, NewConstantPeriod(period), "test")
	if !ok {
		t.Fatal("IIFE generation failed")
	}

	t.Run("initializes_accumulators", func(t *testing.T) {
		accumulators := []string{"sumX := 0.0", "sumY := 0.0", "sumXY := 0.0", "sumX2 := 0.0"}
		for _, acc := range accumulators {
			if !strings.Contains(code, acc) {
				t.Errorf("Missing accumulator initialization: %q", acc)
			}
		}
	})

	t.Run("accumulates_in_loop", func(t *testing.T) {
		accumulations := []string{"sumX += x", "sumY += y", "sumXY += x * y", "sumX2 += x * x"}
		for _, acc := range accumulations {
			if !strings.Contains(code, acc) {
				t.Errorf("Missing accumulation: %q", acc)
			}
		}
	})

	t.Run("calculates_slope_correctly", func(t *testing.T) {
		slopeFormula := "(n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX)"
		if !strings.Contains(code, slopeFormula) {
			t.Error("Slope formula incorrect or missing")
		}
	})

	t.Run("calculates_intercept_correctly", func(t *testing.T) {
		interceptFormula := "(sumY - slope * sumX) / n"
		if !strings.Contains(code, interceptFormula) {
			t.Error("Intercept formula incorrect or missing")
		}
	})

	t.Run("returns_linreg_value", func(t *testing.T) {
		if !strings.Contains(code, "return intercept + slope *") {
			t.Error("Missing linreg return formula")
		}
	})
}

func TestLinregIIFEGenerator_WarmupEdgeCases(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}

	tests := []struct {
		name        string
		period      int
		wantWarmup  string
		needsWarmup bool
	}{
		{
			name:        "period_1_minimal_warmup",
			period:      1,
			wantWarmup:  "ctx.BarIndex < 0",
			needsWarmup: false,
		},
		{
			name:        "period_2",
			period:      2,
			wantWarmup:  "ctx.BarIndex < 1",
			needsWarmup: true,
		},
		{
			name:        "period_10",
			period:      10,
			wantWarmup:  "ctx.BarIndex < 9",
			needsWarmup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := registry.Generate("ta.linreg", accessor, NewConstantPeriod(tt.period), "test")
			if !ok {
				t.Fatal("IIFE generation failed")
			}

			if tt.needsWarmup {
				if !strings.Contains(code, "math.NaN()") {
					t.Error("Generated code should return NaN for warmup period")
				}
				if !strings.Contains(code, "ctx.BarIndex <") {
					t.Error("Generated code missing warmup boundary check")
				}
			}
		})
	}
}

func TestLinregIIFEGenerator_CodeConsistency(t *testing.T) {
	registry1 := NewInlineTAIIFERegistry()
	registry2 := NewInlineTAIIFERegistry()

	accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
	period := 14

	code1, ok1 := registry1.Generate("ta.linreg", accessor, NewConstantPeriod(period), "test")
	if !ok1 {
		t.Fatal("First generation failed")
	}

	code2, ok2 := registry2.Generate("ta.linreg", accessor, NewConstantPeriod(period), "test")
	if !ok2 {
		t.Fatal("Second generation failed")
	}

	if code1 != code2 {
		t.Error("IIFE generation should be deterministic - same inputs should produce identical output")
	}
}

func TestLinregIIFEGenerator_Integration(t *testing.T) {
	registry := NewInlineTAIIFERegistry()

	t.Run("supports_ta_dot_linreg", func(t *testing.T) {
		accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
		code, ok := registry.Generate("ta.linreg", accessor, NewConstantPeriod(10), "test")
		if !ok {
			t.Error("Registry should support 'ta.linreg'")
		}
		if code == "" {
			t.Error("Generated code should not be empty")
		}
	})

	t.Run("supports_linreg_v4_syntax", func(t *testing.T) {
		accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
		code, ok := registry.Generate("linreg", accessor, NewConstantPeriod(10), "test")
		if !ok {
			t.Error("Registry should support 'linreg' (Pine v4 syntax)")
		}
		if code == "" {
			t.Error("Generated code should not be empty")
		}
	})

	t.Run("does_not_support_invalid_function", func(t *testing.T) {
		accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
		_, ok := registry.Generate("ta.invalid", accessor, NewConstantPeriod(10), "test")
		if ok {
			t.Error("Registry should not support 'ta.invalid'")
		}
	})
}

func TestLinregIIFEGenerator_LoopIndexing(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}
	period := 5

	code, ok := registry.Generate("ta.linreg", accessor, NewConstantPeriod(period), "test")
	if !ok {
		t.Fatal("IIFE generation failed")
	}

	t.Run("loop_iterates_forward", func(t *testing.T) {
		if !strings.Contains(code, "for j := 0; j <") {
			t.Error("Loop should iterate forward from 0")
		}
	})

	t.Run("x_coordinate_increases", func(t *testing.T) {
		if !strings.Contains(code, "x := float64(j)") {
			t.Error("X coordinate should be simple loop index")
		}
	})

	t.Run("y_accesses_historical_data", func(t *testing.T) {
		if !strings.Contains(code, "ctx.Data[i-") {
			t.Error("Y coordinate should access historical bar data")
		}
		if !strings.Contains(code, "- j - 1") {
			t.Error("Y indexing should account for backward window traversal")
		}
	})
}

func TestLinregIIFEGenerator_FormulaCorrectness(t *testing.T) {
	registry := NewInlineTAIIFERegistry()
	accessor := &mockLinregAccessor{loopAccess: "ctx.Data[i-period - j - 1].Close"}

	tests := []struct {
		period       int
		wantMultiple int
	}{
		{period: 1, wantMultiple: 0},
		{period: 5, wantMultiple: 4},
		{period: 10, wantMultiple: 9},
		{period: 20, wantMultiple: 19},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			code, ok := registry.Generate("ta.linreg", accessor, NewConstantPeriod(tt.period), "test")
			if !ok {
				t.Fatal("IIFE generation failed")
			}

			if !strings.Contains(code, "intercept + slope * float64(") {
				t.Error("Return formula should be: intercept + slope * float64(length-1)")
			}
		})
	}
}
