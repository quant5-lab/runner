package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

/*
	 TestArrowFunctionFixnan_Integration validates fixnan() generation in arro			result, 			result, err := GenerateStrategyCodeFromAST(program)
				if err != nil {
					t.Fatalf("Generation failed: %v", err)
				}

				code := result.FunctionBody

				for _, pattern := range tt.mainMustContain {
					if !strings.Contains(code, pattern) {
						t.Errorf("Main context code missing pattern %q", pattern)
					}
				}ateStrategyCodeFromAST(program)
				if err != nil {
					t.Fatalf("Generation failed: %v", err)
				}

				code := result.FunctionBody

				for _, pattern := range tt.mainMustContain {
					if !strings.Contains(code, pattern) {
						t.Errorf("Main context code missing pattern %q", pattern)
					}
				}
*/
func TestArrowFunctionFixnan_Integration(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "fixnan with OHLCV field direct return",
			source: `
safeDivClose(denominator) =>
    fixnan(close / denominator)

value = safeDivClose(volume)
`,
			mustContain: []string{
				"func safeDivClose(arrowCtx *context.ArrowContext, denominator float64) float64",
				"func() float64",
				"val :=",
				"if math.IsNaN(val) { return 0.0 }",
				"return val",
			},
			mustNotContain: []string{
				"fixnanState",
				"selfSeries",
				".Position()",
				"for j :=",
			},
		},
		{
			name: "fixnan with complex arithmetic expression in variable",
			source: `
momentum(len) =>
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(close, len) / truerange)
    plus

result = momentum(14)
`,
			mustContain: []string{
				"func momentum(arrowCtx *context.ArrowContext, len float64) float64",
				"plusSeries := arrowCtx.GetOrCreateSeries(\"plus\")",
				"plusSeries.Set(",
				"func() float64",
				"val :=",
				"if math.IsNaN(val) { return 0.0 }",
				"return plusSeries.GetCurrent()",
			},
			mustNotContain: []string{
				"fixnanState",
				"lastValidValue",
			},
		},
		{
			name: "fixnan in tuple return with multiple calls",
			source: `
dirmov(len) =>
    up = change(high)
    down = change(low)
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(up, len) / truerange)
    minus = fixnan(100 * rma(down, len) / truerange)
    [plus, minus]

[x, y] = dirmov(5)
`,
			mustContain: []string{
				"func dirmov(arrowCtx *context.ArrowContext, len float64) (float64, float64)",
				"plusSeries := arrowCtx.GetOrCreateSeries(\"plus\")",
				"minusSeries := arrowCtx.GetOrCreateSeries(\"minus\")",
				"plusSeries.Set(",
				"minusSeries.Set(",
				"return plusSeries.GetCurrent(), minusSeries.GetCurrent()",
			},
			mustNotContain: []string{
				"fixnanState_plus",
				"fixnanState_minus",
			},
		},
		{
			name: "fixnan with ta prefix",
			source: `
indicator() =>
    ta.fixnan(close / volume)

result = indicator()
`,
			mustContain: []string{
				"func indicator(arrowCtx *context.ArrowContext) float64",
				"func() float64",
				"if math.IsNaN(val) { return 0.0 }",
			},
		},
		{
			name: "fixnan with arrow function parameter",
			source: `
processor(src) =>
    fixnan(src * 2.0)

output = processor(close)
`,
			mustContain: []string{
				"func processor(arrowCtx *context.ArrowContext, src float64) float64",
				"func() float64",
				"return val",
			},
		},
		{
			name: "multiple fixnan calls in sequence",
			source: `
normalizer(a, b, c) =>
    na = fixnan(a)
    nb = fixnan(b)
    nc = fixnan(c)
    na + nb + nc

result = normalizer(close, high, low)
`,
			mustContain: []string{
				"func normalizer(arrowCtx *context.ArrowContext, a float64, b float64, c float64) float64",
				"naSeries := arrowCtx.GetOrCreateSeries(\"na\")",
				"nbSeries := arrowCtx.GetOrCreateSeries(\"nb\")",
				"ncSeries := arrowCtx.GetOrCreateSeries(\"nc\")",
			},
		},
		{
			name: "fixnan with nested TA function calls",
			source: `
composite(len) =>
    avg = sma(close, len)
    fixnan(avg / ema(volume, len))

result = composite(20)
`,
			mustContain: []string{
				"func composite(arrowCtx *context.ArrowContext, len float64) float64",
				"avgSeries := arrowCtx.GetOrCreateSeries(\"avg\")",
				"func() float64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
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
				t.Fatalf("Generation failed: %v", err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Generated code missing pattern %q\nCode:\n%s", pattern, code)
				}
			}

			for _, pattern := range tt.mustNotContain {
				if strings.Contains(code, pattern) {
					t.Errorf("Generated code should not contain pattern %q\nCode:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestArrowFunctionFixnan_EdgeCases validates boundary conditions and error handling */
func TestArrowFunctionFixnan_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
		errorMsg    string
		mustContain []string
	}{
		{
			name: "fixnan with OHLCV field - valid",
			source: `
validOHLCV() =>
    fixnan(close)

result = validOHLCV()
`,
			expectError: false,
			mustContain: []string{
				"func validOHLCV(arrowCtx *context.ArrowContext) float64",
				"func() float64",
			},
		},
		{
			name: "fixnan with parameter - valid",
			source: `
validParam(x) =>
    fixnan(x)

result = validParam(close)
`,
			expectError: false,
			mustContain: []string{
				"func validParam(arrowCtx *context.ArrowContext, x float64) float64",
				"val := x",
			},
		},
		{
			name: "fixnan with TA function result - valid",
			source: `
validTA() =>
    fixnan(sma(close, 14))

result = validTA()
`,
			expectError: false,
			mustContain: []string{
				"func validTA(arrowCtx *context.ArrowContext) float64",
			},
		},
		{
			name: "fixnan in deeply nested expression - valid",
			source: `
nested() =>
    a = fixnan(close)
    b = fixnan(high)
    fixnan(a + b)

result = nested()
`,
			expectError: false,
			mustContain: []string{
				"aSeries := arrowCtx.GetOrCreateSeries(\"a\")",
				"bSeries := arrowCtx.GetOrCreateSeries(\"b\")",
				"func() float64",
			},
		},
		{
			name: "fixnan with long variable name - valid",
			source: `
longVarName() =>
    veryLongVariableNameForTestingPurposesOnly = fixnan(close)
    veryLongVariableNameForTestingPurposesOnly

result = longVarName()
`,
			expectError: false,
			mustContain: []string{
				"veryLongVariableNameForTestingPurposesOnlySeries := arrowCtx.GetOrCreateSeries(\"veryLongVariableNameForTestingPurposesOnly\")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Parse failed unexpectedly: %v", err)
				}
				return
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Conversion failed unexpectedly: %v", err)
				}
				return
			}

			result, err := GenerateStrategyCodeFromAST(program)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Generation failed unexpectedly: %v", err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Generated code missing pattern %q\nCode:\n%s", pattern, code)
				}
			}
		})
	}
}

/* TestArrowFunctionFixnan_RealWorldPatterns validates production use cases */
func TestArrowFunctionFixnan_RealWorldPatterns(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		description string
		mustContain []string
	}{
		{
			name: "DMI calculation pattern - simplified",
			source: `
dirmov(len) =>
    up = change(high)
    down = -change(low)
    truerange = rma(tr, len)
    upMA = rma(up, len)
    downMA = rma(down, len)
    plus = fixnan(100 * upMA / truerange)
    minus = fixnan(100 * downMA / truerange)
    [plus, minus]

[p, m] = dirmov(18)
`,
			description: "DMI indicator with fixnan for safe division",
			mustContain: []string{
				"func dirmov(arrowCtx *context.ArrowContext, len float64) (float64, float64)",
				"plusSeries := arrowCtx.GetOrCreateSeries(\"plus\")",
				"minusSeries := arrowCtx.GetOrCreateSeries(\"minus\")",
				"if math.IsNaN(val) { return 0.0 }",
				"return plusSeries.GetCurrent(), minusSeries.GetCurrent()",
			},
		},
		{
			name: "Safe division wrapper",
			source: `
safeDiv(numerator, denominator) =>
    fixnan(numerator / denominator)

ratio = safeDiv(close, volume)
`,
			description: "Common pattern for avoiding NaN in division",
			mustContain: []string{
				"func safeDiv(arrowCtx *context.ArrowContext, numerator float64, denominator float64) float64",
				"func() float64",
				"return val",
			},
		},
		{
			name: "Indicator normalization",
			source: `
normalize(value, baseline) =>
    ratio = value / baseline
    fixnan(ratio * 100)

normalized = normalize(close, sma(close, 200))
`,
			description: "Normalize indicator values with fixnan",
			mustContain: []string{
				"func normalize(arrowCtx *context.ArrowContext, value float64, baseline float64) float64",
				"ratioSeries := arrowCtx.GetOrCreateSeries(\"ratio\")",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			script, err := p.ParseBytes("test.pine", []byte(tt.source))
			if err != nil {
				t.Fatalf("Parse failed for %s: %v", tt.description, err)
			}

			converter := parser.NewConverter()
			program, err := converter.ToESTree(script)
			if err != nil {
				t.Fatalf("Conversion failed for %s: %v", tt.description, err)
			}

			result, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("Generation failed for %s: %v", tt.description, err)
			}

			code := result.UserDefinedFunctions

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("%s: Generated code missing pattern %q\nCode:\n%s",
						tt.description, pattern, code)
				}
			}

			if strings.Contains(code, "fixnanState") {
				t.Errorf("%s: Arrow functions must not generate stateful fixnan code", tt.description)
			}
		})
	}
}

/* TestArrowFunctionFixnan_ConsistencyWithMainContext validates behavior alignment */
func TestArrowFunctionFixnan_ConsistencyWithMainContext(t *testing.T) {
	tests := []struct {
		name                string
		arrowFunctionSource string
		mainContextSource   string
		arrowMustContain    []string
		arrowMustNotContain []string
		mainMustContain     []string
	}{
		{
			name: "fixnan behavior differs by context",
			arrowFunctionSource: `
arrowFixnan(x) =>
    fixnan(x / 10.0)

result = arrowFixnan(close)
`,
			mainContextSource: `
value = fixnan(close / 10.0)
`,
			arrowMustContain: []string{
				"func() float64",
				"if math.IsNaN(val) { return 0.0 }",
			},
			arrowMustNotContain: []string{
				"fixnanState",
			},
			mainMustContain: []string{
				"value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := parser.NewParser()
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			t.Run("arrow_function_context", func(t *testing.T) {
				script, err := p.ParseBytes("test.pine", []byte(tt.arrowFunctionSource))
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
					t.Fatalf("Generation failed: %v", err)
				}

				code := result.UserDefinedFunctions

				for _, pattern := range tt.arrowMustContain {
					if !strings.Contains(code, pattern) {
						t.Errorf("Arrow function code missing pattern %q", pattern)
					}
				}

				for _, pattern := range tt.arrowMustNotContain {
					if strings.Contains(code, pattern) {
						t.Errorf("Arrow function code should not contain pattern %q", pattern)
					}
				}
			})

			t.Run("main_context", func(t *testing.T) {
				script, err := p.ParseBytes("test.pine", []byte(tt.mainContextSource))
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
					t.Fatalf("Generation failed: %v", err)
				}

				code := result.FunctionBody

				for _, pattern := range tt.mainMustContain {
					if !strings.Contains(code, pattern) {
						t.Errorf("Main context code missing pattern %q", pattern)
					}
				}
			})
		})
	}
}
