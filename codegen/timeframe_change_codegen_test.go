package codegen

import (
	"strings"
	"testing"
)

func TestBuildTimeframeChangeIIFE_StructuralProperties(t *testing.T) {
	tests := []struct {
		name        string
		tfExpr      string
		style       timeframeChangeStyle
		wantContain []string
		wantAbsent  []string
	}{
		{
			"float64_style",
			`"1D"`,
			timeframeChangeFloat,
			[]string{
				"func() float64",
				`tf := "1D"`,
				"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex].Time, tf, ctx.PeriodAnchor)",
				"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex-1].Time, tf, ctx.PeriodAnchor)",
				"if ctx.BarIndex == 0 { return 1 }",
				"if currAligned != prevAligned { return 1 }; return 0",
			},
			[]string{"func() bool"},
		},
		{
			"bool_style",
			`"1D"`,
			timeframeChangeBool,
			[]string{
				"func() bool",
				`tf := "1D"`,
				"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex].Time, tf, ctx.PeriodAnchor)",
				"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex-1].Time, tf, ctx.PeriodAnchor)",
				"if ctx.BarIndex == 0 { return true }",
				"return currAligned != prevAligned",
			},
			[]string{"func() float64"},
		},
		{
			"variable_expression",
			"myTimeframe",
			timeframeChangeFloat,
			[]string{"tf := myTimeframe"},
			nil,
		},
		{
			"member_expression",
			"ctx.Timeframe",
			timeframeChangeBool,
			[]string{"tf := ctx.Timeframe"},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := buildTimeframeChangeIIFE(tt.tfExpr, tt.style)

			for _, want := range tt.wantContain {
				if !strings.Contains(code, want) {
					t.Errorf("missing %q in:\n%s", want, code)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(code, absent) {
					t.Errorf("unexpected %q in:\n%s", absent, code)
				}
			}
		})
	}
}

func TestBuildTimeframeChangeIIFE_EnclosedAsIIFE(t *testing.T) {
	tests := []struct {
		name  string
		style timeframeChangeStyle
	}{
		{"float64", timeframeChangeFloat},
		{"bool", timeframeChangeBool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := buildTimeframeChangeIIFE(`"1W"`, tt.style)

			if !strings.HasPrefix(code, "(func()") {
				t.Errorf("IIFE must start with (func(), got: %s", code[:30])
			}
			if !strings.HasSuffix(code, "}())") {
				t.Errorf("IIFE must end with }()), got: %s", code[len(code)-30:])
			}
		})
	}
}

func TestBuildTimeframeChangeIIFE_StyleConsistency(t *testing.T) {
	floatCode := buildTimeframeChangeIIFE(`"1D"`, timeframeChangeFloat)
	boolCode := buildTimeframeChangeIIFE(`"1D"`, timeframeChangeBool)

	sharedFragments := []string{
		"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex].Time, tf, ctx.PeriodAnchor)",
		"context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex-1].Time, tf, ctx.PeriodAnchor)",
		"ctx.BarIndex == 0",
		`tf := "1D"`,
	}

	for _, fragment := range sharedFragments {
		if !strings.Contains(floatCode, fragment) {
			t.Errorf("float64 style missing shared fragment: %q", fragment)
		}
		if !strings.Contains(boolCode, fragment) {
			t.Errorf("bool style missing shared fragment: %q", fragment)
		}
	}
}
