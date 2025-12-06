package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

// TestBooleanLiterals_InTernary_Codegen ensures true/false generate numeric values (1.0/0.0)
func TestBooleanLiterals_InTernary_Codegen(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		mustHave []string
		mustNot  []string
	}{
		{
			name: "false and true in ternary",
			script: `//@version=5
indicator("Test")
x = na(close) ? false : true`,
			mustHave: []string{
				"return 0.0", // false → 0.0
				"return 1.0", // true → 1.0
			},
			mustNot: []string{
				"falseSeries",
				"trueSeries",
				"GetCurrent",
			},
		},
		{
			name: "multiple variables with boolean ternaries",
			script: `//@version=5
indicator("Test")
a = na(close) ? false : true
b = close > 100 ? true : false`,
			mustHave: []string{
				"aSeries.Set(func() float64",
				"bSeries.Set(func() float64",
				"return 0.0",
				"return 1.0",
			},
			mustNot: []string{
				"falseSeries.GetCurrent()",
				"trueSeries.GetCurrent()",
			},
		},
		{
			name: "session time pattern (BB7 regression)",
			script: `//@version=4
study(title="Test", overlay=true)
entry_time = input("0950-1345", title="Entry Time", type=input.session)
session_open = na(time(timeframe.period, entry_time)) ? false : true`,
			mustHave: []string{
				"session_openSeries.Set(func() float64",
				"math.IsNaN(session.TimeFunc",
				"return 0.0", // false
				"return 1.0", // true
			},
			mustNot: []string{
				"falseSeries.GetCurrent()",
				"trueSeries.GetCurrent()",
				"undefined: falseSeries",
				"undefined: trueSeries",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			parseResult, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(parseResult)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			// Check required patterns
			for _, pattern := range tt.mustHave {
				if !strings.Contains(result.FunctionBody, pattern) {
					t.Errorf("Expected pattern %q not found in generated code", pattern)
				}
			}

			// Check forbidden patterns
			for _, pattern := range tt.mustNot {
				if strings.Contains(result.FunctionBody, pattern) {
					t.Errorf("Forbidden pattern %q found in generated code (REGRESSION)", pattern)
				}
			}
		})
	}
}

// TestBooleanLiterals_NotConfusedWithIdentifiers ensures parser disambiguation
func TestBooleanLiterals_NotConfusedWithIdentifiers(t *testing.T) {
	script := `//@version=5
indicator("Test")
// These should be boolean Literals
a = true
b = false
c = true ? 1 : 0
d = false ? 1 : 0
// User-defined variable (should use Series)
myvar = close
e = myvar ? 1 : 0`

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}

	// Booleans should generate 1.00 or 0.00
	requiredPatterns := []string{
		"aSeries.Set(1.00)", // a = true
		"bSeries.Set(0.00)", // b = false
		"myvarSeries.Set(",  // myvar uses Series (not boolean literal)
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(result.FunctionBody, pattern) {
			t.Errorf("Expected pattern %q not found", pattern)
		}
	}

	// Should NOT have these patterns
	forbiddenPatterns := []string{
		"trueSeries.GetCurrent()",
		"falseSeries.GetCurrent()",
		"undefined: trueSeries",
		"undefined: falseSeries",
	}

	for _, pattern := range forbiddenPatterns {
		if strings.Contains(result.FunctionBody, pattern) {
			t.Errorf("REGRESSION: Forbidden pattern %q found", pattern)
		}
	}
}
