package codegen

import (
	"strings"
	"testing"
)

func TestDerivedPriceFormulaGenerator_Generate_HL2(t *testing.T) {
	gen := NewDerivedPriceFormulaGenerator()

	tests := []struct {
		name         string
		highAccess   string
		lowAccess    string
		closeAccess  string
		openAccess   string
		wantContains []string
	}{
		{
			name:        "bar accessor",
			highAccess:  "bar.High",
			lowAccess:   "bar.Low",
			closeAccess: "bar.Close",
			openAccess:  "bar.Open",
			wantContains: []string{
				"bar.High",
				"bar.Low",
				"/ 2",
			},
		},
		{
			name:        "ctx.Data accessor",
			highAccess:  "ctx.Data[i].High",
			lowAccess:   "ctx.Data[i].Low",
			closeAccess: "ctx.Data[i].Close",
			openAccess:  "ctx.Data[i].Open",
			wantContains: []string{
				"ctx.Data[i].High",
				"ctx.Data[i].Low",
				"/ 2",
			},
		},
		{
			name:        "offset accessor",
			highAccess:  "ctx.Data[i-5].High",
			lowAccess:   "ctx.Data[i-5].Low",
			closeAccess: "ctx.Data[i-5].Close",
			openAccess:  "ctx.Data[i-5].Open",
			wantContains: []string{
				"ctx.Data[i-5].High",
				"ctx.Data[i-5].Low",
				"/ 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formula := gen.Generate("hl2", tt.highAccess, tt.lowAccess, tt.closeAccess, tt.openAccess)

			if formula == "" {
				t.Fatal("Generate returned empty string")
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(formula, want) {
					t.Errorf("Expected formula to contain %q, got: %s", want, formula)
				}
			}

			if strings.Contains(formula, tt.closeAccess) {
				t.Errorf("hl2 formula should not contain Close, got: %s", formula)
			}
			if strings.Contains(formula, tt.openAccess) {
				t.Errorf("hl2 formula should not contain Open, got: %s", formula)
			}
		})
	}
}

func TestDerivedPriceFormulaGenerator_Generate_AllPrices(t *testing.T) {
	gen := NewDerivedPriceFormulaGenerator()

	tests := []struct {
		priceName    string
		wantContains []string
		wantDivisor  string
		fieldCount   int
	}{
		{
			priceName:    "hl2",
			wantContains: []string{"HIGH", "LOW"},
			wantDivisor:  "/ 2",
			fieldCount:   2,
		},
		{
			priceName:    "hlc3",
			wantContains: []string{"HIGH", "LOW", "CLOSE"},
			wantDivisor:  "/ 3",
			fieldCount:   3,
		},
		{
			priceName:    "ohlc4",
			wantContains: []string{"OPEN", "HIGH", "LOW", "CLOSE"},
			wantDivisor:  "/ 4",
			fieldCount:   4,
		},
		{
			priceName:    "hlcc4",
			wantContains: []string{"HIGH", "LOW", "CLOSE"},
			wantDivisor:  "/ 4",
			fieldCount:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.priceName, func(t *testing.T) {
			formula := gen.Generate(tt.priceName, "HIGH", "LOW", "CLOSE", "OPEN")

			if formula == "" {
				t.Fatalf("Generate(%s) returned empty string", tt.priceName)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(formula, want) {
					t.Errorf("Expected formula to contain %q, got: %s", want, formula)
				}
			}

			if !strings.Contains(formula, tt.wantDivisor) {
				t.Errorf("Expected formula to contain divisor %q, got: %s", tt.wantDivisor, formula)
			}

			plusCount := strings.Count(formula, "+")
			expectedPlusCount := tt.fieldCount - 1
			if plusCount != expectedPlusCount {
				t.Errorf("Expected %d + operators for %s, got %d in: %s", expectedPlusCount, tt.priceName, plusCount, formula)
			}
		})
	}
}

func TestDerivedPriceFormulaGenerator_Generate_InvalidPrice(t *testing.T) {
	gen := NewDerivedPriceFormulaGenerator()

	invalidPrices := []string{"hl3", "hlc4", "invalid", "", "HL2"}

	for _, price := range invalidPrices {
		t.Run(price, func(t *testing.T) {
			formula := gen.Generate(price, "bar.High", "bar.Low", "bar.Close", "bar.Open")

			if formula != "" {
				t.Errorf("Generate(%s) should return empty for invalid price, got: %s", price, formula)
			}
		})
	}
}

func TestDerivedPriceFormulaGenerator_Generate_FormulaStructure(t *testing.T) {
	gen := NewDerivedPriceFormulaGenerator()

	tests := []struct {
		priceName string
		checkFunc func(string) bool
		desc      string
	}{
		{
			priceName: "hl2",
			checkFunc: func(f string) bool {
				return strings.HasPrefix(f, "((") && strings.HasSuffix(f, ")")
			},
			desc: "should have parentheses wrapper",
		},
		{
			priceName: "hlc3",
			checkFunc: func(f string) bool {
				return strings.Count(f, "+") == 2
			},
			desc: "should have exactly 2 addition operators",
		},
		{
			priceName: "ohlc4",
			checkFunc: func(f string) bool {
				return strings.Count(f, "+") == 3
			},
			desc: "should have exactly 3 addition operators",
		},
		{
			priceName: "hlcc4",
			checkFunc: func(f string) bool {
				return strings.Count(f, "CLOSE") == 2
			},
			desc: "should have Close field twice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.priceName+"_"+tt.desc, func(t *testing.T) {
			formula := gen.Generate(tt.priceName, "HIGH", "LOW", "CLOSE", "OPEN")

			if !tt.checkFunc(formula) {
				t.Errorf("%s formula %s, got: %s", tt.priceName, tt.desc, formula)
			}
		})
	}
}

func TestDerivedPriceFormulaGenerator_Generate_AccessorFlexibility(t *testing.T) {
	gen := NewDerivedPriceFormulaGenerator()

	accessorPatterns := []struct {
		name        string
		highAccess  string
		lowAccess   string
		closeAccess string
		openAccess  string
	}{
		{
			name:        "bar direct",
			highAccess:  "bar.High",
			lowAccess:   "bar.Low",
			closeAccess: "bar.Close",
			openAccess:  "bar.Open",
		},
		{
			name:        "ctx.Data current",
			highAccess:  "highSeries.GetCurrent()",
			lowAccess:   "lowSeries.GetCurrent()",
			closeAccess: "closeSeries.GetCurrent()",
			openAccess:  "openSeries.GetCurrent()",
		},
		{
			name:        "ctx.Data offset",
			highAccess:  "ctx.Data[i-10].High",
			lowAccess:   "ctx.Data[i-10].Low",
			closeAccess: "ctx.Data[i-10].Close",
			openAccess:  "ctx.Data[i-10].Open",
		},
		{
			name:        "complex expression",
			highAccess:  "data[idx+offset].High",
			lowAccess:   "data[idx+offset].Low",
			closeAccess: "data[idx+offset].Close",
			openAccess:  "data[idx+offset].Open",
		},
	}

	for _, pattern := range accessorPatterns {
		t.Run(pattern.name, func(t *testing.T) {
			formula := gen.Generate("hl2", pattern.highAccess, pattern.lowAccess, pattern.closeAccess, pattern.openAccess)

			if !strings.Contains(formula, pattern.highAccess) {
				t.Errorf("Formula should contain high accessor %q, got: %s", pattern.highAccess, formula)
			}
			if !strings.Contains(formula, pattern.lowAccess) {
				t.Errorf("Formula should contain low accessor %q, got: %s", pattern.lowAccess, formula)
			}
		})
	}
}
