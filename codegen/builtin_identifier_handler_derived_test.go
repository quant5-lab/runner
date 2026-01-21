package codegen

import (
	"strings"
	"testing"
)

func TestBuiltinIdentifierHandler_DerivedPrices_IsBuiltinSeriesIdentifier(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"hl2 is builtin", "hl2", true},
		{"hlc3 is builtin", "hlc3", true},
		{"ohlc4 is builtin", "ohlc4", true},
		{"hlcc4 is builtin", "hlcc4", true},
		{"HL2 uppercase not builtin", "HL2", false},
		{"hl3 invalid not builtin", "hl3", false},
		{"derived not builtin", "derived", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.IsBuiltinSeriesIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("IsBuiltinSeriesIdentifier(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_DerivedPrices_GenerateCurrentBarAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		priceName    string
		wantContains []string
		wantFormula  string
	}{
		{
			name:      "hl2 current bar",
			priceName: "hl2",
			wantContains: []string{
				"bar.High",
				"bar.Low",
				"/ 2",
			},
			wantFormula: "((bar.High + bar.Low) / 2)",
		},
		{
			name:      "hlc3 current bar",
			priceName: "hlc3",
			wantContains: []string{
				"bar.High",
				"bar.Low",
				"bar.Close",
				"/ 3",
			},
			wantFormula: "((bar.High + bar.Low + bar.Close) / 3)",
		},
		{
			name:      "ohlc4 current bar",
			priceName: "ohlc4",
			wantContains: []string{
				"bar.Open",
				"bar.High",
				"bar.Low",
				"bar.Close",
				"/ 4",
			},
			wantFormula: "((bar.Open + bar.High + bar.Low + bar.Close) / 4)",
		},
		{
			name:      "hlcc4 current bar",
			priceName: "hlcc4",
			wantContains: []string{
				"bar.High",
				"bar.Low",
				"bar.Close",
				"/ 4",
			},
			wantFormula: "((bar.High + bar.Low + bar.Close + bar.Close) / 4)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handler.GenerateCurrentBarAccess(tt.priceName)

			if code == "" {
				t.Fatal("GenerateCurrentBarAccess returned empty string")
			}

			if code != tt.wantFormula {
				t.Errorf("GenerateCurrentBarAccess(%s) = %s, want %s", tt.priceName, code, tt.wantFormula)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected code to contain %q, got: %s", want, code)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_DerivedPrices_GenerateSecurityContextAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		priceName    string
		wantContains []string
	}{
		{
			name:      "hl2 in security context",
			priceName: "hl2",
			wantContains: []string{
				"ctx.Data[ctx.BarIndex].High",
				"ctx.Data[ctx.BarIndex].Low",
				"/ 2",
			},
		},
		{
			name:      "hlc3 in security context",
			priceName: "hlc3",
			wantContains: []string{
				"ctx.Data[ctx.BarIndex].High",
				"ctx.Data[ctx.BarIndex].Low",
				"ctx.Data[ctx.BarIndex].Close",
				"/ 3",
			},
		},
		{
			name:      "ohlc4 in security context",
			priceName: "ohlc4",
			wantContains: []string{
				"ctx.Data[ctx.BarIndex].Open",
				"ctx.Data[ctx.BarIndex].High",
				"ctx.Data[ctx.BarIndex].Low",
				"ctx.Data[ctx.BarIndex].Close",
				"/ 4",
			},
		},
		{
			name:      "hlcc4 in security context",
			priceName: "hlcc4",
			wantContains: []string{
				"ctx.Data[ctx.BarIndex].High",
				"ctx.Data[ctx.BarIndex].Low",
				"ctx.Data[ctx.BarIndex].Close",
				"/ 4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handler.GenerateSecurityContextAccess(tt.priceName)

			if code == "" {
				t.Fatal("GenerateSecurityContextAccess returned empty string")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected code to contain %q, got: %s", want, code)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_DerivedPrices_GenerateHistoricalAccess(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name         string
		priceName    string
		offset       int
		wantContains []string
	}{
		{
			name:      "hl2[1] offset 1",
			priceName: "hl2",
			offset:    1,
			wantContains: []string{
				"func() float64",
				"if i-1 >= 0",
				"ctx.Data[i-1].High",
				"ctx.Data[i-1].Low",
				"/ 2",
				"math.NaN()",
			},
		},
		{
			name:      "hlc3[2] offset 2",
			priceName: "hlc3",
			offset:    2,
			wantContains: []string{
				"func() float64",
				"if i-2 >= 0",
				"ctx.Data[i-2].High",
				"ctx.Data[i-2].Low",
				"ctx.Data[i-2].Close",
				"/ 3",
				"math.NaN()",
			},
		},
		{
			name:      "ohlc4[1] offset 1",
			priceName: "ohlc4",
			offset:    1,
			wantContains: []string{
				"func() float64",
				"if i-1 >= 0",
				"ctx.Data[i-1].Open",
				"ctx.Data[i-1].High",
				"ctx.Data[i-1].Low",
				"ctx.Data[i-1].Close",
				"/ 4",
				"math.NaN()",
			},
		},
		{
			name:      "hlcc4[3] offset 3",
			priceName: "hlcc4",
			offset:    3,
			wantContains: []string{
				"func() float64",
				"if i-3 >= 0",
				"ctx.Data[i-3].High",
				"ctx.Data[i-3].Low",
				"ctx.Data[i-3].Close",
				"/ 4",
				"math.NaN()",
			},
		},
		{
			name:      "hl2[10] large offset",
			priceName: "hl2",
			offset:    10,
			wantContains: []string{
				"func() float64",
				"if i-10 >= 0",
				"ctx.Data[i-10].High",
				"ctx.Data[i-10].Low",
				"math.NaN()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handler.GenerateHistoricalAccess(tt.priceName, tt.offset)

			if code == "" {
				t.Fatal("GenerateHistoricalAccess returned empty string")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("Expected code to contain %q, got: %s", want, code)
				}
			}
		})
	}
}

func TestBuiltinIdentifierHandler_DerivedPrices_ConsistencyAcrossContexts(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	derivedPrices := []string{"hl2", "hlc3", "ohlc4", "hlcc4"}

	for _, price := range derivedPrices {
		t.Run(price, func(t *testing.T) {
			currentBar := handler.GenerateCurrentBarAccess(price)
			securityCtx := handler.GenerateSecurityContextAccess(price)
			historical := handler.GenerateHistoricalAccess(price, 1)

			if currentBar == "" {
				t.Errorf("GenerateCurrentBarAccess(%s) returned empty", price)
			}
			if securityCtx == "" {
				t.Errorf("GenerateSecurityContextAccess(%s) returned empty", price)
			}
			if historical == "" {
				t.Errorf("GenerateHistoricalAccess(%s, 1) returned empty", price)
			}

			if !strings.Contains(currentBar, "bar.High") && !strings.Contains(currentBar, "bar.Low") {
				t.Errorf("Current bar access for %s should use bar accessor", price)
			}
			if !strings.Contains(securityCtx, "ctx.Data[ctx.BarIndex]") {
				t.Errorf("Security context access for %s should use ctx.Data[ctx.BarIndex]", price)
			}
			if !strings.Contains(historical, "ctx.Data[i-1]") {
				t.Errorf("Historical access for %s should use ctx.Data[i-offset]", price)
			}
		})
	}
}

func TestBuiltinIdentifierHandler_DerivedPrices_FormulaStructure(t *testing.T) {
	handler := NewBuiltinIdentifierHandler()

	tests := []struct {
		name          string
		priceName     string
		fieldCount    int
		divisor       string
		requiredField string
	}{
		{"hl2 has 2 fields", "hl2", 2, "/ 2", "High"},
		{"hlc3 has 3 fields", "hlc3", 3, "/ 3", "Close"},
		{"ohlc4 has 4 fields", "ohlc4", 4, "/ 4", "Open"},
		{"hlcc4 has 4 fields with Close twice", "hlcc4", 4, "/ 4", "Close"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := handler.GenerateCurrentBarAccess(tt.priceName)

			if !strings.Contains(code, tt.divisor) {
				t.Errorf("Formula for %s should contain divisor %s, got: %s", tt.priceName, tt.divisor, code)
			}

			if !strings.Contains(code, tt.requiredField) {
				t.Errorf("Formula for %s should contain field %s, got: %s", tt.priceName, tt.requiredField, code)
			}

			fieldParts := strings.Count(code, "bar.")
			if fieldParts != tt.fieldCount {
				t.Errorf("Formula for %s should have %d bar. references, got %d in: %s", tt.priceName, tt.fieldCount, fieldParts, code)
			}
		})
	}
}
