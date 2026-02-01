package codegen

import (
	"strings"
	"testing"
)

func TestIIFECodeBuilder_WarmupExpressionGeneration(t *testing.T) {
	tests := []struct {
		name             string
		period           PeriodExpression
		baseOffset       int
		expectedWarmup   string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:           "Constant period - zero offset",
			period:         NewConstantPeriod(20),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < 19 { return math.NaN() }",
			shouldContain:  []string{"func() float64", "19", "math.NaN()"},
		},
		{
			name:           "Constant period - with offset",
			period:         NewConstantPeriod(10),
			baseOffset:     2,
			expectedWarmup: "if ctx.BarIndex < 11 { return math.NaN() }",
			shouldContain:  []string{"11"},
		},
		{
			name:           "Constant period - minimum (1)",
			period:         NewConstantPeriod(1),
			baseOffset:     0,
			expectedWarmup: "",
			shouldContain:  []string{"func() float64"},
		},
		{
			name:           "Constant period - large value",
			period:         NewConstantPeriod(200),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < 199 { return math.NaN() }",
			shouldContain:  []string{"199"},
		},
		{
			name:             "Runtime period - zero offset",
			period:           NewRuntimePeriod("length"),
			baseOffset:       0,
			expectedWarmup:   "if ctx.BarIndex < int(length)-1 { return math.NaN() }",
			shouldContain:    []string{"int(length)", "-1"},
			shouldNotContain: []string{"+"},
		},
		{
			name:           "Runtime period - with offset",
			period:         NewRuntimePeriod("period"),
			baseOffset:     1,
			expectedWarmup: "if ctx.BarIndex < int(period)-1+1 { return math.NaN() }",
			shouldContain:  []string{"int(period)", "-1+1"},
		},
		{
			name:           "Runtime period - large offset",
			period:         NewRuntimePeriod("len"),
			baseOffset:     5,
			expectedWarmup: "if ctx.BarIndex < int(len)-1+5 { return math.NaN() }",
			shouldContain:  []string{"int(len)", "-1+5"},
		},
		{
			name:           "Runtime period - single char variable",
			period:         NewRuntimePeriod("n"),
			baseOffset:     0,
			expectedWarmup: "if ctx.BarIndex < int(n)-1 { return math.NaN() }",
			shouldContain:  []string{"int(n)", "-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := NewIIFECodeBuilder().
				WithWarmupCheckPeriodExpression(tt.period, tt.baseOffset).
				WithBody("return 42.0").
				Build()

			if !strings.Contains(code, tt.expectedWarmup) && tt.expectedWarmup != "" {
				t.Errorf("Expected warmup guard %q\nGot code:\n%s", tt.expectedWarmup, code)
			}

			for _, pattern := range tt.shouldContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Expected pattern %q in code:\n%s", pattern, code)
				}
			}

			for _, pattern := range tt.shouldNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Should NOT contain pattern %q in code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestIIFECodeBuilder_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name           string
		period         int
		expectedWarmup string
		allowEmpty     bool
	}{
		{"Legacy constant warmup - 20", 20, "if ctx.BarIndex < 19 { return math.NaN() }", false},
		{"Legacy constant warmup - 14", 14, "if ctx.BarIndex < 13 { return math.NaN() }", false},
		{"Legacy constant warmup - 1", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := NewIIFECodeBuilder().
				WithWarmupCheck(tt.period).
				WithBody("return 42.0").
				Build()

			if tt.allowEmpty && tt.expectedWarmup == "" {
				return
			}

			if !strings.Contains(code, tt.expectedWarmup) {
				t.Errorf("Expected warmup %q in code:\n%s", tt.expectedWarmup, code)
			}
		})
	}
}

func TestIIFECodeBuilder_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		buildFunc     func() string
		shouldContain []string
	}{
		{
			name: "No warmup check - body only",
			buildFunc: func() string {
				return NewIIFECodeBuilder().
					WithBody("return 100.0").
					Build()
			},
			shouldContain: []string{"func() float64", "return 100.0", "}()"},
		},
		{
			name: "Empty body",
			buildFunc: func() string {
				return NewIIFECodeBuilder().
					WithWarmupCheck(10).
					WithBody("").
					Build()
			},
			shouldContain: []string{"func() float64", "if ctx.BarIndex < 9"},
		},
		{
			name: "Complex body expression",
			buildFunc: func() string {
				return NewIIFECodeBuilder().
					WithWarmupCheckPeriodExpression(NewRuntimePeriod("n"), 0).
					WithBody("sum := 0.0; for i := 0; i < 10; i++ { sum += float64(i) }; return sum").
					Build()
			},
			shouldContain: []string{"int(n)-1", "sum := 0.0", "for i := 0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := tt.buildFunc()

			for _, pattern := range tt.shouldContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Expected pattern %q in code:\n%s", pattern, code)
				}
			}
		})
	}
}

type testAccessorWithOffset struct {
	baseOffset int
}

func (t *testAccessorWithOffset) GenerateLoopValueAccess(loopVar string) string {
	return "testValue"
}

func (t *testAccessorWithOffset) GenerateInitialValueAccess(period int) string {
	return "testInitial"
}

func (t *testAccessorWithOffset) GenerateCurrentValueAccess() string {
	return "testCurrent"
}

func (t *testAccessorWithOffset) GetBaseOffset() int {
	return t.baseOffset
}
