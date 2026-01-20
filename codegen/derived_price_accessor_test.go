package codegen

import (
	"strings"
	"testing"
)

func TestDerivedPriceAccessor_HL2(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hl2", 0)

	t.Run("GenerateLoopValueAccess", func(t *testing.T) {
		code := accessor.GenerateLoopValueAccess("j")
		if !strings.Contains(code, "ctx.Data[ctx.BarIndex-j].High") {
			t.Errorf("Expected hl2 loop access to contain High field, got: %s", code)
		}
		if !strings.Contains(code, "ctx.Data[ctx.BarIndex-j].Low") {
			t.Errorf("Expected hl2 loop access to contain Low field, got: %s", code)
		}
		if !strings.Contains(code, "/ 2") {
			t.Errorf("Expected hl2 to divide by 2, got: %s", code)
		}
	})

	t.Run("GenerateInitialValueAccess", func(t *testing.T) {
		code := accessor.GenerateInitialValueAccess(14)
		if !strings.Contains(code, "ctx.BarIndex-13") {
			t.Errorf("Expected initial value at offset 13 (period-1), got: %s", code)
		}
	})

	t.Run("GenerateCurrentValueAccess", func(t *testing.T) {
		code := accessor.GenerateCurrentValueAccess()
		if !strings.Contains(code, "ctx.Data[ctx.BarIndex]") {
			t.Errorf("Expected current bar access, got: %s", code)
		}
	})
}

func TestDerivedPriceAccessor_HLC3(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hlc3", 0)

	code := accessor.GenerateCurrentValueAccess()
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].High") {
		t.Errorf("Expected hlc3 to contain High, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Low") {
		t.Errorf("Expected hlc3 to contain Low, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Close") {
		t.Errorf("Expected hlc3 to contain Close, got: %s", code)
	}
	if !strings.Contains(code, "/ 3") {
		t.Errorf("Expected hlc3 to divide by 3, got: %s", code)
	}
}

func TestDerivedPriceAccessor_OHLC4(t *testing.T) {
	accessor := NewDerivedPriceAccessor("ohlc4", 0)

	code := accessor.GenerateCurrentValueAccess()
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Open") {
		t.Errorf("Expected ohlc4 to contain Open, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].High") {
		t.Errorf("Expected ohlc4 to contain High, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Low") {
		t.Errorf("Expected ohlc4 to contain Low, got: %s", code)
	}
	if !strings.Contains(code, "ctx.Data[ctx.BarIndex].Close") {
		t.Errorf("Expected ohlc4 to contain Close, got: %s", code)
	}
	if !strings.Contains(code, "/ 4") {
		t.Errorf("Expected ohlc4 to divide by 4, got: %s", code)
	}
}

func TestDerivedPriceAccessor_HLCC4(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hlcc4", 0)

	code := accessor.GenerateCurrentValueAccess()
	// hlcc4 has Close twice
	closeCount := strings.Count(code, "Close")
	if closeCount != 2 {
		t.Errorf("Expected hlcc4 to contain Close twice, got %d occurrences in: %s", closeCount, code)
	}
}

func TestDerivedPriceAccessor_WithBaseOffset(t *testing.T) {
	tests := []struct {
		name       string
		baseOffset int
		wantInLoop string
	}{
		{"no offset", 0, "ctx.BarIndex-j"},
		{"offset 1", 1, "ctx.BarIndex-(j+1)"},
		{"offset 2", 2, "ctx.BarIndex-(j+2)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewDerivedPriceAccessor("hl2", tt.baseOffset)
			code := accessor.GenerateLoopValueAccess("j")
			if !strings.Contains(code, tt.wantInLoop) {
				t.Errorf("Expected loop access with %q, got: %s", tt.wantInLoop, code)
			}
		})
	}
}

func TestDerivedPriceAccessor_InitialValueWithBaseOffset(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hl2", 1)
	code := accessor.GenerateInitialValueAccess(14)
	// period-1 + baseOffset = 13 + 1 = 14
	if !strings.Contains(code, "ctx.BarIndex-14") {
		t.Errorf("Expected initial value at offset 14 (period-1 + baseOffset), got: %s", code)
	}
}

func TestDerivedPriceAccessor_CurrentValueWithBaseOffset(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hl2", 2)
	code := accessor.GenerateCurrentValueAccess()
	if !strings.Contains(code, "ctx.BarIndex-2") {
		t.Errorf("Expected current value at offset 2 (baseOffset), got: %s", code)
	}
}

func TestDerivedPriceAccessor_GetBaseOffset(t *testing.T) {
	tests := []int{0, 1, 5, 10}
	for _, offset := range tests {
		accessor := NewDerivedPriceAccessor("hl2", offset)
		if got := accessor.GetBaseOffset(); got != offset {
			t.Errorf("GetBaseOffset() = %d, want %d", got, offset)
		}
	}
}

func TestDerivedPriceAccessor_GetPreamble(t *testing.T) {
	accessor := NewDerivedPriceAccessor("hl2", 0)
	if preamble := accessor.GetPreamble(); preamble != "" {
		t.Errorf("Expected empty preamble, got: %s", preamble)
	}
}

func TestDerivedPriceAccessor_UnknownPrice(t *testing.T) {
	accessor := NewDerivedPriceAccessor("unknown", 0)
	code := accessor.GenerateCurrentValueAccess()
	if code != "math.NaN()" {
		t.Errorf("Expected math.NaN() for unknown price, got: %s", code)
	}
}
