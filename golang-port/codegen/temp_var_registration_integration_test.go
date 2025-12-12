package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

// TestTempVarRegistration_TAFunctionsOnly verifies temp var declarations for TA functions
func TestTempVarRegistration_TAFunctionsOnly(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		expectedDecl      string
		expectedSeriesVar string
	}{
		{
			name: "sma generates temp var",
			script: `//@version=5
indicator("Test")
daily_close = request.security(syminfo.tickerid, "D", sma(close, 20))
`,
			expectedDecl:      "var sma_",
			expectedSeriesVar: "Series",
		},
		{
			name: "ema generates temp var",
			script: `//@version=5
indicator("Test")
daily_ema = request.security(syminfo.tickerid, "D", ema(close, 21))
`,
			expectedDecl:      "var ema_",
			expectedSeriesVar: "Series",
		},
		{
			name: "nested ta functions generate multiple temp vars",
			script: `//@version=5
indicator("Test")
daily_rma = request.security(syminfo.tickerid, "D", rma(sma(close, 10), 20))
`,
			expectedDecl:      "var sma_",
			expectedSeriesVar: "Series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			if !strings.Contains(result.FunctionBody, tt.expectedDecl) {
				t.Errorf("Expected temp var declaration %q not found in:\n%s", tt.expectedDecl, result.FunctionBody)
			}

			if !strings.Contains(result.FunctionBody, tt.expectedSeriesVar) {
				t.Errorf("Expected Series variable %q not found in:\n%s", tt.expectedSeriesVar, result.FunctionBody)
			}
		})
	}
}

// TestTempVarRegistration_MathFunctionsOnly verifies temp var declarations for math functions without TA
func TestTempVarRegistration_MathFunctionsOnly(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		expectedDecl      string
		expectedSeriesVar string
	}{
		{
			name: "max with constants does not generate temp var",
			script: `//@version=5
indicator("Test")
daily_max = request.security(syminfo.tickerid, "D", math.max(10, 20))
`,
			expectedDecl:      "",
			expectedSeriesVar: "",
		},
		{
			name: "min with constants does not generate temp var",
			script: `//@version=5
indicator("Test")
daily_min = request.security(syminfo.tickerid, "D", math.min(5, 15))
`,
			expectedDecl:      "",
			expectedSeriesVar: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			// Verify math function temp vars NOT created for constant-only expressions
			if tt.expectedDecl == "" {
				if strings.Contains(result.FunctionBody, "var math_max") || strings.Contains(result.FunctionBody, "var math_min") {
					t.Errorf("Unexpected math temp var declaration found in:\n%s", result.FunctionBody)
				}
			}
		})
	}
}

// TestTempVarRegistration_MathWithTANested verifies temp var declarations for math functions with TA dependencies
func TestTempVarRegistration_MathWithTANested(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		expectedMathDecl  string
		expectedTADecl    string
		expectedSeriesVar string
	}{
		{
			name: "rma with max and change generates multiple temp vars",
			script: `//@version=5
indicator("Test")
daily_rma = request.security(syminfo.tickerid, "D", ta.rma(math.max(ta.change(close), 0), 9))
`,
			expectedMathDecl:  "var math_max_",
			expectedTADecl:    "var ta_change_",
			expectedSeriesVar: "Series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			if !strings.Contains(result.FunctionBody, tt.expectedMathDecl) {
				t.Errorf("Expected math temp var %q not found in:\n%s", tt.expectedMathDecl, result.FunctionBody)
			}

			if !strings.Contains(result.FunctionBody, tt.expectedTADecl) {
				t.Errorf("Expected TA temp var %q not found in:\n%s", tt.expectedTADecl, result.FunctionBody)
			}

			if !strings.Contains(result.FunctionBody, tt.expectedSeriesVar) {
				t.Errorf("Expected Series variable %q not found in:\n%s", tt.expectedSeriesVar, result.FunctionBody)
			}
		})
	}
}

// TestTempVarRegistration_ComplexNested verifies temp var declarations for deeply nested expressions
func TestTempVarRegistration_ComplexNested(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		expectedDecl []string
	}{
		{
			name: "triple nested ta functions",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.rma(ta.sma(ta.ema(close, 10), 20), 30))
`,
			expectedDecl: []string{"var ta_ema_", "var ta_sma_", "var ta_rma_"},
		},
		{
			name: "nested math and ta combination",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.rma(math.max(ta.change(close), 0), 9))
`,
			expectedDecl: []string{"var ta_change_", "var math_max_", "var ta_rma_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			for _, expectedDecl := range tt.expectedDecl {
				if !strings.Contains(result.FunctionBody, expectedDecl) {
					t.Errorf("Expected temp var %q not found in:\n%s", expectedDecl, result.FunctionBody)
				}
			}
		})
	}
}

// TestTempVarRegistration_EdgeCases verifies edge cases for temp var registration
func TestTempVarRegistration_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		script       string
		expectedDecl string
		notExpected  string
	}{
		{
			name: "ta function in arithmetic",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", ta.sma(close, 20) * 2)
`,
			expectedDecl: "var ta_sma_",
			notExpected:  "",
		},
		{
			name: "math function without ta dependencies",
			script: `//@version=5
indicator("Test")
daily = request.security(syminfo.tickerid, "D", math.abs(close))
`,
			expectedDecl: "",
			notExpected:  "var math_abs_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.script))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Codegen failed: %v", err)
			}

			if tt.expectedDecl != "" && !strings.Contains(result.FunctionBody, tt.expectedDecl) {
				t.Errorf("Expected temp var %q not found in:\n%s", tt.expectedDecl, result.FunctionBody)
			}

			if tt.notExpected != "" && strings.Contains(result.FunctionBody, tt.notExpected) {
				t.Errorf("Unexpected temp var %q found in:\n%s", tt.notExpected, result.FunctionBody)
			}
		})
	}
}
