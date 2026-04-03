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
			name:           "SMA - computed period",
			generator:      &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewComputedPeriod("(nSeries.GetCurrent() / 2)"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int((nSeries.GetCurrent() / 2))-1 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int((nSeries.GetCurrent() / 2)); j++"},
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
			name:           "WMA - computed period",
			generator:      &WMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:         NewComputedPeriod("math.Round(math.Sqrt(nSeries.GetCurrent()))"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(math.Round(math.Sqrt(nSeries.GetCurrent())))-1 { return math.NaN() }",
			shouldContain:  []string{"for j := 0; j < int(math.Round(math.Sqrt(nSeries.GetCurrent()))); j++"},
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
		{
			name:          "EMA - computed period",
			generator:     &EMAIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()},
			period:        NewComputedPeriod("(nSeries.GetCurrent() / 2)"),
			shouldContain: []string{"alpha", "ema", "func() float64"},
		},
		{
			name:          "RMA - computed period",
			generator:     &RMAIIFEGenerator{namingStrategy: series_naming.NewStatefulIndicatorNamer()},
			period:        NewComputedPeriod("math.Round(math.Sqrt(nSeries.GetCurrent()))"),
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
		{
			name:       "SMA - computed period",
			generator:  &SMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewComputedPeriod("(nSeries.GetCurrent() / 2)"),
			baseOffset: 0,
		},
		{
			name:       "Highest - computed period",
			generator:  &HighestIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()},
			period:     NewComputedPeriod("math.Round(nSeries.GetCurrent())"),
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

/* WMA weight must use float64 arithmetic to avoid Go type mismatch between float64 period and int loop counter */
func TestWMAIIFEGenerator_TypeSafeWeightCalculation(t *testing.T) {
	tests := []struct {
		name            string
		period          PeriodExpression
		expectedWeight  string
		forbiddenWeight string
	}{
		{
			name:            "Constant period uses float64 subtraction",
			period:          NewConstantPeriod(5),
			expectedWeight:  "weight := float64(5) - float64(j)",
			forbiddenWeight: "float64(5 - j)",
		},
		{
			name:            "Runtime period uses float64 subtraction",
			period:          NewRuntimePeriod("period"),
			expectedWeight:  "weight := float64(period) - float64(j)",
			forbiddenWeight: "float64(period - j)",
		},
		{
			name:            "Computed period uses float64 subtraction",
			period:          NewComputedPeriod("(nSeries.GetCurrent() / 2)"),
			expectedWeight:  "weight := float64((nSeries.GetCurrent() / 2)) - float64(j)",
			forbiddenWeight: "float64((nSeries.GetCurrent() / 2) - j)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &WMAIIFEGenerator{namingStrategy: series_naming.NewWindowBasedNamer()}
			accessor := &testAccessorWithOffset{baseOffset: 0}
			code := gen.Generate(accessor, tt.period, "test_hash")

			if !strings.Contains(code, tt.expectedWeight) {
				t.Errorf("Expected weight pattern: %q\nGenerated:\n%s", tt.expectedWeight, code)
			}

			if strings.Contains(code, tt.forbiddenWeight) {
				t.Errorf("Forbidden weight pattern found: %q\nGenerated:\n%s", tt.forbiddenWeight, code)
			}
		})
	}
}
