package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

func TestIIFEGenerators_WarmupGuardGeneration(t *testing.T) {
	tests := []struct {
		name           string
		generator      InlineTAIIFEGenerator
		period         PeriodExpression
		baseOffset     int
		expectedWarmup string
		shouldContain  []string
	}{
		{
			name:           "SMA - constant period",
			generator:      &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewConstantPeriod(20),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < 19 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < 20; j++"},
		},
		{
			name:           "SMA - runtime period",
			generator:      &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("length"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int(length); j++"},
		},
		{
			name:           "WMA - runtime period",
			generator:      &WMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("period"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(period)-1 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int(period); j++"},
		},
		{
			name:           "STDEV - runtime period",
			generator:      &STDEVIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("len"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(len)-1 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int(len); j++", "variance"},
		},
		{
			name:           "Highest - runtime period",
			generator:      &HighestIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("bars"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(bars)-1 { return math.NaN() }",
			shouldContain:  []string{"periodVal := int(bars)", "if v > highest"},
		},
		{
			name:           "Lowest - runtime period",
			generator:      &LowestIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("bars"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(bars)-1 { return math.NaN() }",
			shouldContain:  []string{"periodVal := int(bars)", "if v < lowest"},
		},
		{
			name:           "Linreg - runtime period",
			generator:      &LinregIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("length"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			shouldContain:  []string{"periodVal := int(length)", "slope", "intercept"},
		},
		{
			name:           "SMA - runtime period with offset",
			generator:      &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewRuntimePeriod("n"),
			baseOffset:     2,
			expectedWarmup: "if ctx.BarIndex < int(n)-1+2 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int(n); j++"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &testAccessorWithOffset{baseOffset: tt.baseOffset}
			code := tt.generator.Generate(accessor, tt.period, "test_hash")

			if !strings.Contains(code, tt.expectedWarmup) {
				t.Errorf("Expected warmup: %s\nGenerated code:\n%s", tt.expectedWarmup, code)
			}

			for _, pattern := range tt.shouldContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestStatefulIndicatorGenerators(t *testing.T) {
	tests := []struct {
		name          string
		generator     InlineTAIIFEGenerator
		period        PeriodExpression
		shouldContain []string
	}{
		{
			name:          "EMA - constant period",
			generator:     &EMAIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()},
			period:        NewConstantPeriod(14),
			shouldContain: []string{"alpha", "ema", "func() float64"},
		},
		{
			name:          "EMA - runtime period",
			generator:     &EMAIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()},
			period:        NewRuntimePeriod("length"),
			shouldContain: []string{"alpha", "ema", "func() float64"},
		},
		{
			name:          "RMA - runtime period",
			generator:     &RMAIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()},
			period:        NewRuntimePeriod("period"),
			shouldContain: []string{"alpha", "rma", "func() float64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &testAccessorWithOffset{baseOffset: 0}
			code := tt.generator.Generate(accessor, tt.period, "test_hash")

			for _, pattern := range tt.shouldContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestIIFEGenerators_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		generator  InlineTAIIFEGenerator
		period     PeriodExpression
		baseOffset int
		skipWarmup bool
	}{
		{
			name:       "SMA - minimum period (1)",
			generator:  &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewConstantPeriod(1),
			baseOffset: 0,
			skipWarmup: true,
		},
		{
			name:       "SMA - large period",
			generator:  &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewConstantPeriod(200),
			baseOffset: 0,
		},
		{
			name:       "SMA - large base offset",
			generator:  &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewRuntimePeriod("n"),
			baseOffset: 10,
		},
		{
			name:       "Highest - single character variable",
			generator:  &HighestIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewRuntimePeriod("n"),
			baseOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := &testAccessorWithOffset{baseOffset: tt.baseOffset}
			code := tt.generator.Generate(accessor, tt.period, "test_hash")

			if !strings.Contains(code, "func() float64") {
				t.Errorf("Missing IIFE wrapper\nCode:\n%s", code)
			}

			if !strings.Contains(code, "return") {
				t.Errorf("Missing return statement\nCode:\n%s", code)
			}

			if !tt.skipWarmup && !strings.Contains(code, "ctx.BarIndex") {
				t.Errorf("Missing warmup check\nCode:\n%s", code)
			}
		})
	}
}
