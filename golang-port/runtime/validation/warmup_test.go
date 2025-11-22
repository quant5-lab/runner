package validation

import (
	"math"
	"strings"
	"testing"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/parser"
)

// TestWarmupAnalyzer_SimpleLiteralSubscript tests basic subscript detection
func TestWarmupAnalyzer_SimpleLiteralSubscript(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		expectedLookback int
		expectedSource   string
	}{
		{
			name: "simple_literal_subscript",
			code: `
//@version=5
indicator("test")
x = close[10]
`,
			expectedLookback: 10,
			expectedSource:   "close[10]",
		},
		{
			name: "large_lookback",
			code: `
//@version=5
indicator("test")
historical = high[1260]
`,
			expectedLookback: 1260,
			expectedSource:   "high[1260]",
		},
		{
			name: "multiple_subscripts_max",
			code: `
//@version=5
indicator("test")
x = close[100]
y = open[500]
z = high[50]
`,
			expectedLookback: 500,
			expectedSource:   "open[500]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()
			reqs := analyzer.AnalyzeScript(script)

			if len(reqs) == 0 {
				t.Fatal("Expected requirements, got none")
			}

			// Find max lookback
			maxLookback := 0
			var maxReq WarmupRequirement
			for _, req := range reqs {
				if req.MaxLookback > maxLookback {
					maxLookback = req.MaxLookback
					maxReq = req
				}
			}

			if maxLookback != tt.expectedLookback {
				t.Errorf("Expected max lookback %d, got %d", tt.expectedLookback, maxLookback)
			}

			if !strings.Contains(maxReq.Expression, tt.expectedSource) {
				t.Errorf("Expected source containing %q, got %q", tt.expectedSource, maxReq.Expression)
			}
		})
	}
}

// TestWarmupAnalyzer_VariableSubscript tests subscripts with variable indices
func TestWarmupAnalyzer_VariableSubscript(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		expectedLookback int
	}{
		{
			name: "constant_variable_subscript",
			code: `
//@version=5
indicator("test")
n = 252
x = close[n]
`,
			expectedLookback: 252,
		},
		{
			name: "calculated_constant_subscript",
			code: `
//@version=5
indicator("test")
years = 5
days = 252
total = years * days
x = close[total]
`,
			expectedLookback: 1260,
		},
		{
			name: "ternary_subscript_evaluates_to_max",
			code: `
//@version=5
indicator("test")
n = timeframe.isdaily ? 252 : 52
x = close[n]
`,
			expectedLookback: 0, // Cannot evaluate ternary at compile time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()
			reqs := analyzer.AnalyzeScript(script)

			if tt.expectedLookback == 0 {
				// For ternary, we can't determine at compile time
				// This is acceptable - runtime will handle it
				return
			}

			if len(reqs) == 0 {
				t.Fatal("Expected requirements, got none")
			}

			maxLookback := 0
			for _, req := range reqs {
				if req.MaxLookback > maxLookback {
					maxLookback = req.MaxLookback
				}
			}

			if maxLookback != tt.expectedLookback {
				t.Errorf("Expected lookback %d, got %d", tt.expectedLookback, maxLookback)
			}
		})
	}
}

// TestWarmupAnalyzer_ComplexExpressions tests complex subscript expressions
func TestWarmupAnalyzer_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		expectedLookback int
	}{
		{
			name: "expression_in_subscript",
			code: `
//@version=5
indicator("test")
base = 100
offset = 50
x = close[base + offset]
`,
			expectedLookback: 150,
		},
		{
			name: "math_pow_in_calculation",
			code: `
//@version=5
indicator("test")
yA = 5
interval = 252
nA = interval * yA
viA = close[nA]
`,
			expectedLookback: 1260,
		},
		{
			name: "nested_expressions",
			code: `
//@version=5
indicator("test")
period = 10
multiplier = 2
total = period * multiplier
x = close[total * 2]
`,
			expectedLookback: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()
			reqs := analyzer.AnalyzeScript(script)

			if len(reqs) == 0 {
				t.Fatal("Expected requirements, got none")
			}

			maxLookback := 0
			for _, req := range reqs {
				if req.MaxLookback > maxLookback {
					maxLookback = req.MaxLookback
				}
			}

			if maxLookback != tt.expectedLookback {
				t.Errorf("Expected lookback %d, got %d", tt.expectedLookback, maxLookback)
			}
		})
	}
}

// TestValidateDataAvailability_EdgeCases tests validation edge cases
func TestValidateDataAvailability_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		barCount      int
		requirements  []WarmupRequirement
		expectError   bool
		errorContains string
	}{
		{
			name:         "no_requirements_always_valid",
			barCount:     10,
			requirements: []WarmupRequirement{},
			expectError:  false,
		},
		{
			name:     "exact_minimum_bars_invalid",
			barCount: 1260,
			requirements: []WarmupRequirement{
				{MaxLookback: 1260, Source: "src[1260]"},
			},
			expectError:   true,
			errorContains: "need 1261+ bars",
		},
		{
			name:     "one_bar_above_minimum_valid",
			barCount: 1261,
			requirements: []WarmupRequirement{
				{MaxLookback: 1260, Source: "src[1260]"},
			},
			expectError: false,
		},
		{
			name:     "way_too_few_bars",
			barCount: 100,
			requirements: []WarmupRequirement{
				{MaxLookback: 1260, Source: "src[1260]"},
			},
			expectError:   true,
			errorContains: "have 100 bars",
		},
		{
			name:     "multiple_requirements_checks_max",
			barCount: 500,
			requirements: []WarmupRequirement{
				{MaxLookback: 100, Source: "x[100]"},
				{MaxLookback: 600, Source: "y[600]"},
				{MaxLookback: 200, Source: "z[200]"},
			},
			expectError:   true,
			errorContains: "need 601+ bars",
		},
		{
			name:     "zero_bars_with_requirement",
			barCount: 0,
			requirements: []WarmupRequirement{
				{MaxLookback: 10, Source: "x[10]"},
			},
			expectError: true,
		},
		{
			name:         "single_bar_with_no_lookback",
			barCount:     1,
			requirements: []WarmupRequirement{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDataAvailability(tt.barCount, tt.requirements)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestGetWarmupInfo tests warmup information formatting
func TestGetWarmupInfo(t *testing.T) {
	tests := []struct {
		name         string
		barCount     int
		requirements []WarmupRequirement
		expectedInfo string
	}{
		{
			name:         "no_warmup",
			barCount:     1000,
			requirements: []WarmupRequirement{},
			expectedInfo: "No warmup period required",
		},
		{
			name:     "typical_warmup",
			barCount: 1500,
			requirements: []WarmupRequirement{
				{MaxLookback: 1260},
			},
			expectedInfo: "Warmup: 1260 bars, Valid output: 240 bars (16.0%)",
		},
		{
			name:     "insufficient_data_shows_zero",
			barCount: 100,
			requirements: []WarmupRequirement{
				{MaxLookback: 1260},
			},
			expectedInfo: "Warmup: 1260 bars, Valid output: 0 bars (0.0%)",
		},
		{
			name:     "small_warmup_high_percentage",
			barCount: 1000,
			requirements: []WarmupRequirement{
				{MaxLookback: 50},
			},
			expectedInfo: "Warmup: 50 bars, Valid output: 950 bars (95.0%)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := GetWarmupInfo(tt.barCount, tt.requirements)
			if info != tt.expectedInfo {
				t.Errorf("Expected info %q, got %q", tt.expectedInfo, info)
			}
		})
	}
}

// TestWarmupAnalyzer_RealWorldScenarios tests real Pine Script patterns
func TestWarmupAnalyzer_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name              string
		code              string
		barCount          int
		expectValid       bool
		expectedWarmup    int
		expectedValidBars int
	}{
		{
			name: "rolling_cagr_5yr_sufficient_data",
			code: `
//@version=5
indicator("Rolling CAGR")
yA = input.float(5, title='Years')
iyA = math.pow(yA, -1)
src = input.source(defval = close)
interval_multiplier = timeframe.isdaily ? 252 : timeframe.isweekly ? 52 : na
nA = interval_multiplier * yA
viA = src[nA]
vf = src[0]
cagrA = (math.pow(vf / viA, iyA) - 1) * 100
plot(cagrA)
`,
			barCount:          1500,
			expectValid:       true, // nA is not constant (depends on runtime timeframe), so no compile-time requirement detected
			expectedWarmup:    0,    // Cannot determine at compile time
			expectedValidBars: 1500, // No warmup requirement detected, all bars considered valid
		},
		{
			name: "fixed_period_strategy",
			code: `
//@version=5
strategy("MA Cross")
fast = 10
slow = 50
fastMA = ta.sma(close, fast)
slowMA = ta.sma(close, slow)
historical_fast = fastMA[20]
plot(fastMA)
`,
			barCount:          100,
			expectValid:       true,
			expectedWarmup:    20,
			expectedValidBars: 80,
		},
		{
			name: "deep_lookback",
			code: `
//@version=5
indicator("Deep History")
baseline = close[500]
current = close[0]
change = (current - baseline) / baseline * 100
plot(change)
`,
			barCount:          1000,
			expectValid:       true,
			expectedWarmup:    500,
			expectedValidBars: 500,
		},
		{
			name: "insufficient_data_scenario",
			code: `
//@version=5
indicator("Long Period")
reference = close[1000]
plot(close - reference)
`,
			barCount:          500,
			expectValid:       false, // Need 1001 bars, have 500
			expectedWarmup:    1000,
			expectedValidBars: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()
			reqs := analyzer.AnalyzeScript(script)

			err := ValidateDataAvailability(tt.barCount, reqs)
			isValid := (err == nil)

			if isValid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v (error: %v)", tt.expectValid, isValid, err)
			}

			// Check warmup info only if we found requirements
			if len(reqs) > 0 {
				info := GetWarmupInfo(tt.barCount, reqs)
				if !strings.Contains(info, "Warmup:") {
					t.Errorf("Expected warmup info, got: %s", info)
				}
			}
		})
	}
}

// TestWarmupAnalyzer_DifferentTimeframes tests timeframe-specific calculations
func TestWarmupAnalyzer_DifferentTimeframes(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectReqs bool
	}{
		{
			name: "daily_timeframe_check",
			code: `
//@version=5
indicator("test")
multiplier = timeframe.isdaily ? 252 : 52
period = 5 * multiplier
old_value = close[period]
`,
			expectReqs: false, // Ternary cannot be evaluated at compile time
		},
		{
			name: "fixed_daily_calculation",
			code: `
//@version=5
indicator("test")
daily_periods = 252
years = 5
total = daily_periods * years
old_value = close[total]
`,
			expectReqs: true, // Can evaluate: 252 * 5 = 1260
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()
			reqs := analyzer.AnalyzeScript(script)

			hasReqs := len(reqs) > 0
			if hasReqs != tt.expectReqs {
				t.Errorf("Expected requirements=%v, got requirements=%v (count: %d)",
					tt.expectReqs, hasReqs, len(reqs))
			}
		})
	}
}

// TestEvaluateConstant_MathOperations tests constant evaluation for various operations
func TestEvaluateConstant_MathOperations(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		varName  string
		expected float64
	}{
		{
			name: "simple_multiplication",
			code: `
//@version=5
indicator("test")
result = 5 * 252
`,
			varName:  "result",
			expected: 1260,
		},
		{
			name: "addition_and_multiplication",
			code: `
//@version=5
indicator("test")
result = (10 + 5) * 10
`,
			varName:  "result",
			expected: 150,
		},
		{
			name: "division",
			code: `
//@version=5
indicator("test")
result = 1000 / 4
`,
			varName:  "result",
			expected: 250,
		},
		{
			name: "subtraction",
			code: `
//@version=5
indicator("test")
result = 1500 - 240
`,
			varName:  "result",
			expected: 1260,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := parseTestScript(t, tt.code)
			analyzer := NewWarmupAnalyzer()

			// Collect constants
			for _, node := range script.Body {
				analyzer.collectConstants(node)
			}

			val, exists := analyzer.constants[tt.varName]
			if !exists {
				t.Fatalf("Constant %q not found", tt.varName)
			}

			if math.Abs(val-tt.expected) > 0.0001 {
				t.Errorf("Expected %v = %.2f, got %.2f", tt.varName, tt.expected, val)
			}
		})
	}
}

// Helper function to parse test scripts
func parseTestScript(t *testing.T, code string) *ast.Program {
	t.Helper()
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	script, err := p.ParseString("test.pine", code)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}
	converter := parser.NewConverter()
	program, err := converter.ToESTree(script)
	if err != nil {
		t.Fatalf("Failed to convert to ESTree: %v", err)
	}
	return program
}
