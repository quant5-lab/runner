package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func parseAndGenerate(t *testing.T, script string) string {
	t.Helper()
	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	parsed, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	program, err := parser.NewConverter().ToESTree(parsed)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}
	result, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("Codegen failed: %v", err)
	}
	return result.FunctionBody
}

// TestTempVarRegistration_TAFunctionsInSecurity verifies that TA calls inside
// request.security() are evaluated at runtime via the bar evaluator and are NOT
// hoisted as temp var declarations into the main bar loop.
func TestTempVarRegistration_TAFunctionsInSecurity(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		notExpected []string // must NOT appear in main bar loop
	}{
		{
			name: "sma inside security uses bar evaluator not temp var",
			script: `//@version=5
indicator("Test")
daily_close = request.security(syminfo.tickerid, "D", sma(close, 20))
`,
			notExpected: []string{"var sma_"},
		},
		{
			name: "ema inside security uses bar evaluator not temp var",
			script: `//@version=5
indicator("Test")
daily_ema = request.security(syminfo.tickerid, "D", ema(close, 21))
`,
			notExpected: []string{"var ema_"},
		},
		{
			name: "nested ta inside security uses bar evaluator not temp vars",
			script: `//@version=5
indicator("Test")
daily_rma = request.security(syminfo.tickerid, "D", rma(sma(close, 10), 20))
`,
			notExpected: []string{"var sma_", "var rma_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := parseAndGenerate(t, tt.script)

			for _, bad := range tt.notExpected {
				if strings.Contains(code, bad) {
					t.Errorf("TA inside security() must not hoist %q to main loop; found in:\n%s", bad, code)
				}
			}
			if !strings.Contains(code, "EvaluateAtBar") {
				t.Errorf("Expected runtime bar evaluator (EvaluateAtBar) for TA inside security()")
			}
			if !strings.Contains(code, "Series") {
				t.Errorf("Expected result series declaration in generated code")
			}
		})
	}
}

// TestTempVarRegistration_MathFunctionsOnly verifies temp var declarations for math functions without TA
func TestTempVarRegistration_MathFunctionsOnly(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "max with constants does not generate temp var",
			script: `//@version=5
indicator("Test")
daily_max = request.security(syminfo.tickerid, "D", math.max(10, 20))
`,
		},
		{
			name: "min with constants does not generate temp var",
			script: `//@version=5
indicator("Test")
daily_min = request.security(syminfo.tickerid, "D", math.min(5, 15))
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := parseAndGenerate(t, tt.script)
			if strings.Contains(code, "var math_max") || strings.Contains(code, "var math_min") {
				t.Errorf("Unexpected math temp var declaration found in:\n%s", code)
			}
		})
	}
}

// TestTempVarRegistration_MathWithTANested verifies that deeply nested TA+math expressions
// inside request.security() are not hoisted to the main bar loop.
func TestTempVarRegistration_MathWithTANested(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		notExpected []string
	}{
		{
			name: "rma with nested math and change inside security uses bar evaluator",
			script: `//@version=5
indicator("Test")
daily_rma = request.security(syminfo.tickerid, "D", ta.rma(math.max(ta.change(close), 0), 9))
`,
			notExpected: []string{"var math_max_", "var ta_change_", "var ta_rma_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := parseAndGenerate(t, tt.script)

			for _, bad := range tt.notExpected {
				if strings.Contains(code, bad) {
					t.Errorf("Nested TA/math inside security() must not hoist %q to main loop; found in:\n%s", bad, code)
				}
			}
			if !strings.Contains(code, "EvaluateAtBar") {
				t.Errorf("Expected runtime bar evaluator (EvaluateAtBar) for nested TA inside security()")
			}
		})
	}
}

// TestTempVarRegistration_ComplexNested verifies that deeply nested TA expressions
// inside request.security() are evaluated via the bar evaluator, not inlined.
func TestTempVarRegistration_ComplexNested(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		notExpected []string
	}{
		{
			name: "triple nested ta functions inside security use bar evaluator",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.rma(ta.sma(ta.ema(close, 10), 20), 30))
`,
			notExpected: []string{"var ta_ema_", "var ta_sma_", "var ta_rma_"},
		},
		{
			name: "nested math and ta combination inside security use bar evaluator",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.rma(math.max(ta.change(close), 0), 9))
`,
			notExpected: []string{"var ta_change_", "var math_max_", "var ta_rma_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := parseAndGenerate(t, tt.script)

			for _, bad := range tt.notExpected {
				if strings.Contains(code, bad) {
					t.Errorf("Complex nested TA inside security() must not hoist %q to main loop; found in:\n%s", bad, code)
				}
			}
			if !strings.Contains(code, "EvaluateAtBar") {
				t.Errorf("Expected runtime bar evaluator (EvaluateAtBar) for nested TA inside security()")
			}
		})
	}
}

// TestTempVarRegistration_EdgeCases verifies edge cases for temp var registration
func TestTempVarRegistration_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		notExpected string
		mustHave    string
	}{
		{
			name: "ta function in arithmetic inside security uses bar evaluator",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.sma(close, 20) * 2)
`,
			notExpected: "var ta_sma_",
			mustHave:    "EvaluateAtBar",
		},
		{
			name: "math function without ta dependencies does not generate temp var",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", math.abs(close))
`,
			notExpected: "var math_abs_",
			mustHave:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := parseAndGenerate(t, tt.script)

			if tt.notExpected != "" && strings.Contains(code, tt.notExpected) {
				t.Errorf("Must not hoist %q to main loop; found in:\n%s", tt.notExpected, code)
			}
			if tt.mustHave != "" && !strings.Contains(code, tt.mustHave) {
				t.Errorf("Expected %q in generated code:\n%s", tt.mustHave, code)
			}
		})
	}
}
