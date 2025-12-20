package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

/* TestArrowFunctionTACallGenerator_CanHandle validates TA function recognition */
func TestArrowFunctionTACallGenerator_CanHandle(t *testing.T) {
	g := newTestGenerator()
	gen := NewArrowFunctionTACallGenerator(g)

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		// TA functions with ta. prefix
		{"ta.sma recognized", "ta.sma", true},
		{"ta.ema recognized", "ta.ema", true},
		{"ta.stdev recognized", "ta.stdev", true},
		{"ta.rma recognized", "ta.rma", true},
		{"ta.wma recognized", "ta.wma", true},

		// TA functions without ta. prefix (PineScript v4 compatibility)
		{"sma without prefix", "sma", true},
		{"ema without prefix", "ema", true},
		{"stdev without prefix", "stdev", true},
		{"rma without prefix", "rma", true},
		{"wma without prefix", "wma", true},

		// Non-TA functions
		{"user function", "myFunc", false},
		{"strategy function", "strategy.entry", false},
		{"plot function", "plot", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gen.iifeRegistry.IsSupported(tt.funcName)
			if got != tt.want {
				t.Errorf("IsSupported(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_ArgumentExtraction validates argument parsing */
func TestArrowFunctionTACallGenerator_ArgumentExtraction(t *testing.T) {
	tests := []struct {
		name           string
		call           *ast.CallExpression
		expectError    bool
		expectedPeriod int
	}{
		{
			name: "literal arguments",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			expectError:    false,
			expectedPeriod: 20,
		},
		{
			name: "integer period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: int(14)},
				},
			},
			expectError:    false,
			expectedPeriod: 14,
		},
		{
			name: "series source",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "mySeries"},
					&ast.Literal{Value: 50.0},
				},
			},
			expectError:    false,
			expectedPeriod: 50,
		},
		{
			name: "parameter period - uses default",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "period"},
				},
			},
			expectError:    false,
			expectedPeriod: 20,
		},
		{
			name: "insufficient arguments",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expectError: true,
		},
		{
			name: "no arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["period"] = "float"
			gen := NewArrowFunctionTACallGenerator(g)

			funcName := extractCallFunctionName(tt.call)
			accessor, period, err := gen.extractTAArguments(funcName, tt.call)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if accessor == nil {
				t.Error("Expected accessor, got nil")
			}

			if period != tt.expectedPeriod {
				t.Errorf("Period = %d, want %d", period, tt.expectedPeriod)
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_SourceClassification validates source type detection */
func TestArrowFunctionTACallGenerator_SourceClassification(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		variables   map[string]string
		expectError bool
		checkType   func(*testing.T, AccessGenerator)
	}{
		{
			name: "OHLCV close",
			expr: &ast.Identifier{Name: "close"},
			checkType: func(t *testing.T, gen AccessGenerator) {
				code := gen.GenerateLoopValueAccess("j")
				if !strings.Contains(code, "Close") {
					t.Errorf("Expected Close field, got %s", code)
				}
			},
		},
		{
			name: "OHLCV high",
			expr: &ast.Identifier{Name: "high"},
			checkType: func(t *testing.T, gen AccessGenerator) {
				code := gen.GenerateLoopValueAccess("j")
				if !strings.Contains(code, "High") {
					t.Errorf("Expected High field, got %s", code)
				}
			},
		},
		{
			name: "Series variable",
			expr: &ast.Identifier{Name: "mySeries"},
			checkType: func(t *testing.T, gen AccessGenerator) {
				code := gen.GenerateLoopValueAccess("j")
				if !strings.Contains(code, "mySeries") {
					t.Errorf("Expected mySeries, got %s", code)
				}
			},
		},
		{
			name:      "Arrow function parameter",
			expr:      &ast.Identifier{Name: "myParam"},
			variables: map[string]string{"myParam": "float"},
			checkType: func(t *testing.T, gen AccessGenerator) {
				if _, ok := gen.(*ArrowFunctionParameterAccessor); !ok {
					t.Errorf("Expected ArrowFunctionParameterAccessor, got %T", gen)
				}
			},
		},
		{
			name: "MemberExpression ctx.close",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "ctx"},
				Property: &ast.Identifier{Name: "close"},
			},
			checkType: func(t *testing.T, gen AccessGenerator) {
				code := gen.GenerateLoopValueAccess("j")
				if !strings.Contains(code, "Close") {
					t.Errorf("Expected Close field, got %s", code)
				}
			},
		},
		{
			name: "Invalid MemberExpression",
			expr: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "invalid"},
				Property: &ast.Identifier{Name: "field"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			if tt.variables != nil {
				for k, v := range tt.variables {
					g.variables[k] = v
				}
			}
			gen := NewArrowFunctionTACallGenerator(g)

			accessor, err := gen.createAccessorFromExpression(tt.expr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if accessor == nil {
				t.Fatal("Expected accessor, got nil")
			}

			if tt.checkType != nil {
				tt.checkType(t, accessor)
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_IIFEGeneration validates IIFE code output */
func TestArrowFunctionTACallGenerator_IIFEGeneration(t *testing.T) {
	tests := []struct {
		name        string
		call        *ast.CallExpression
		mustContain []string
	}{
		{
			name: "SMA generates IIFE",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			mustContain: []string{
				"func() float64",
				"ctx.BarIndex",
				"math.NaN()",
				"sum",
				"return",
			},
		},
		{
			name: "EMA generates IIFE",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ema"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 14.0},
				},
			},
			mustContain: []string{
				"func() float64",
				"alpha",
				"ema",
				"return",
			},
		},
		{
			name: "STDEV generates IIFE",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "stdev"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 10.0},
				},
			},
			mustContain: []string{
				"func() float64",
				"mean",
				"variance",
				"math.Sqrt",
				"return",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			gen := NewArrowFunctionTACallGenerator(g)

			code, err := gen.Generate(tt.call)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}

			for _, want := range tt.mustContain {
				if !strings.Contains(code, want) {
					t.Errorf("Missing %q in generated code:\n%s", want, code)
				}
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_PeriodExtraction validates period value parsing */
func TestArrowFunctionTACallGenerator_PeriodExtraction(t *testing.T) {
	tests := []struct {
		name        string
		expr        ast.Expression
		variables   map[string]string
		expected    int
		expectError bool
	}{
		{
			name:     "float literal",
			expr:     &ast.Literal{Value: 20.0},
			expected: 20,
		},
		{
			name:     "integer literal",
			expr:     &ast.Literal{Value: int(15)},
			expected: 15,
		},
		{
			name:     "large period",
			expr:     &ast.Literal{Value: 200.0},
			expected: 200,
		},
		{
			name:     "minimum period",
			expr:     &ast.Literal{Value: 1.0},
			expected: 1,
		},
		{
			name:      "parameter identifier - uses default",
			expr:      &ast.Identifier{Name: "len"},
			variables: map[string]string{"len": "float"},
			expected:  20,
		},
		{
			name:        "global constant identifier",
			expr:        &ast.Identifier{Name: "unknown"},
			variables:   map[string]string{},
			expectError: true,
		},
		{
			name:        "string literal",
			expr:        &ast.Literal{Value: "not_a_number"},
			expectError: true,
		},
		{
			name:        "unsupported expression type",
			expr:        &ast.BinaryExpression{Operator: "+"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			if tt.variables != nil {
				for k, v := range tt.variables {
					g.variables[k] = v
				}
			}
			gen := NewArrowFunctionTACallGenerator(g)

			period, err := gen.extractPeriodValue(tt.expr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if period != tt.expected {
				t.Errorf("Period = %d, want %d", period, tt.expected)
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_EdgeCases validates boundary conditions */
func TestArrowFunctionTACallGenerator_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*generator)
		call        *ast.CallExpression
		expectError bool
	}{
		{
			name: "unsupported TA function",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.unsupported"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 20.0},
				},
			},
			expectError: true,
		},
		{
			name: "missing source argument",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "nil call expression",
			call: nil,
		},
		{
			name: "nested series access",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.MemberExpression{
						Object:   &ast.Identifier{Name: "nested"},
						Property: &ast.Identifier{Name: "close"},
					},
					&ast.Literal{Value: 10.0},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			if tt.setup != nil {
				tt.setup(g)
			}
			gen := NewArrowFunctionTACallGenerator(g)

			if tt.call == nil {
				return
			}

			_, err := gen.Generate(tt.call)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

/* TestArrowFunctionParameterAccessor_CodeGeneration validates parameter accessor output */
func TestArrowFunctionParameterAccessor_CodeGeneration(t *testing.T) {
	tests := []struct {
		name            string
		paramName       string
		loopVar         string
		period          int
		expectedLoop    string
		expectedInitial string
	}{
		{
			name:            "standard parameter",
			paramName:       "length",
			loopVar:         "j",
			period:          20,
			expectedLoop:    "lengthSeries.Get(j)",
			expectedInitial: "lengthSeries.Get(20-1)",
		},
		{
			name:            "different loop variable",
			paramName:       "period",
			loopVar:         "i",
			period:          10,
			expectedLoop:    "periodSeries.Get(i)",
			expectedInitial: "periodSeries.Get(10-1)",
		},
		{
			name:            "single period",
			paramName:       "len",
			loopVar:         "k",
			period:          1,
			expectedLoop:    "lenSeries.Get(k)",
			expectedInitial: "lenSeries.Get(1-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessor := NewArrowFunctionParameterAccessor(tt.paramName)

			loopCode := accessor.GenerateLoopValueAccess(tt.loopVar)
			if loopCode != tt.expectedLoop {
				t.Errorf("Loop access = %q, want %q", loopCode, tt.expectedLoop)
			}

			initialCode := accessor.GenerateInitialValueAccess(tt.period)
			if initialCode != tt.expectedInitial {
				t.Errorf("Initial access = %q, want %q", initialCode, tt.expectedInitial)
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_Integration validates full compilation flow */
func TestArrowFunctionTACallGenerator_Integration(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "arrow function with SMA",
			script: `//@version=4
study("Test")
indicator(period) =>
    sma(close, period)
result = indicator(20)
plot(result)`,
			mustContain: []string{
				"func indicator(ctx *Context, period float64) float64",
				"func() float64",
				"sum",
				"return",
			},
		},
		{
			name: "arrow function with multiple TA calls",
			script: `//@version=4
study("Test")
bands(len, mult) =>
    avg = sma(close, len)
    dev = stdev(close, len)
    avg + dev * mult
upper = bands(20, 2)
plot(upper)`,
			mustContain: []string{
				"func bands(ctx *Context, len float64, mult float64) float64",
				"func() float64",
				"sum",
				"variance",
			},
		},
		{
			name: "arrow function with series source",
			script: `//@version=4
study("Test")
smoothed(src, len) =>
    ema(src, len)
result = smoothed(close, 14)
plot(result)`,
			mustContain: []string{
				"func smoothed(ctx *Context, srcSeries *series.Series, len float64) float64",
				"alpha",
				"ema",
				"srcSeries.Get(",
				"smoothed(ctx, closeSeries, 14.0)",
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

			code, err := GenerateStrategyCodeFromAST(program)
			if err != nil {
				t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
			}

			for _, want := range tt.mustContain {
				if !strings.Contains(code.FunctionBody, want) {
					t.Errorf("Missing %q in generated code", want)
				}
			}

			for _, notWant := range tt.mustNotContain {
				if strings.Contains(code.FunctionBody, notWant) {
					t.Errorf("Unexpected %q in generated code", notWant)
				}
			}
		})
	}
}
