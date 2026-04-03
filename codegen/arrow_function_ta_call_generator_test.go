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
	gen := newTestArrowTAGenerator(g)

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
		name                 string
		call                 *ast.CallExpression
		expectError          bool
		expectedPeriod       int
		expectRuntimePeriod  bool
		runtimeVariableName  string
		expectComputedPeriod bool
		computedContains     string
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
			name: "parameter period - creates runtime period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "sma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "period"},
				},
			},
			expectError:         false,
			expectRuntimePeriod: true,
			runtimeVariableName: "period",
		},
		{
			name: "binary expression period - creates computed period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.BinaryExpression{
						Left:     &ast.Identifier{Name: "period"},
						Operator: "/",
						Right:    &ast.Literal{Value: 2.0},
					},
				},
			},
			expectError:          false,
			expectComputedPeriod: true,
			computedContains:     "period",
		},
		{
			name: "call expression period - creates computed period",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "wma"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "math.round"},
						Arguments: []ast.Expression{&ast.Identifier{Name: "period"}},
					},
				},
			},
			expectError:          false,
			expectComputedPeriod: true,
			computedContains:     "period",
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
			gen := newTestArrowTAGenerator(g)

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

			if tt.expectComputedPeriod {
				computedPeriod, ok := period.(*ComputedPeriod)
				if !ok {
					t.Errorf("Expected ComputedPeriod, got %T", period)
					return
				}
				if !strings.Contains(computedPeriod.AsGoExpr(), tt.computedContains) {
					t.Errorf("ComputedPeriod expression %q should contain %q", computedPeriod.AsGoExpr(), tt.computedContains)
				}
				if computedPeriod.IsConstant() {
					t.Error("ComputedPeriod must not be constant")
				}
			} else if tt.expectRuntimePeriod {
				runtimePeriod, ok := period.(*RuntimePeriod)
				if !ok {
					t.Errorf("Expected RuntimePeriod, got %T", period)
					return
				}
				if runtimePeriod.variableName != tt.runtimeVariableName {
					t.Errorf("RuntimePeriod variable = %q, want %q", runtimePeriod.variableName, tt.runtimeVariableName)
				}
			} else {
				constPeriod, ok := period.(*ConstantPeriod)
				if !ok {
					t.Errorf("Expected ConstantPeriod, got %T", period)
					return
				}
				if constPeriod.value != tt.expectedPeriod {
					t.Errorf("Period = %d, want %d", constPeriod.value, tt.expectedPeriod)
				}
			}
		})
	}
}

/* TestArrowFunctionTACallGenerator_ExtractChangeArguments validates change() argument parsing */
func TestArrowFunctionTACallGenerator_ExtractChangeArguments(t *testing.T) {
	tests := []struct {
		name                string
		call                *ast.CallExpression
		expectError         bool
		expectedOffset      int
		expectRuntimeOffset bool
		runtimeVariableName string
	}{
		{
			name: "change with source only",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
				},
			},
			expectError:    false,
			expectedOffset: 1,
		},
		{
			name: "change with source and offset",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
					&ast.Literal{Value: 3.0},
				},
			},
			expectError:    false,
			expectedOffset: 3,
		},
		{
			name: "ta.change with source and offset",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "low"},
					&ast.Literal{Value: 5.0},
				},
			},
			expectError:    false,
			expectedOffset: 5,
		},
		{
			name: "change with integer offset",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: int(10)},
				},
			},
			expectError:    false,
			expectedOffset: 10,
		},
		{
			name: "change without arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "change with parameter as offset",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Identifier{Name: "period"},
				},
			},
			expectError:         false,
			expectRuntimeOffset: true,
			runtimeVariableName: "period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			g.variables["period"] = "float"
			gen := newTestArrowTAGenerator(g)

			accessor, offset, err := gen.extractChangeArguments(tt.call)

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

			/* Validate offset type and value */
			if tt.expectRuntimeOffset {
				runtimeOffset, ok := offset.(*RuntimePeriod)
				if !ok {
					t.Errorf("Expected RuntimePeriod offset, got %T", offset)
					return
				}
				if runtimeOffset.variableName != tt.runtimeVariableName {
					t.Errorf("RuntimePeriod variable = %q, want %q", runtimeOffset.variableName, tt.runtimeVariableName)
				}
			} else {
				constOffset, ok := offset.(*ConstantPeriod)
				if !ok {
					t.Errorf("Expected ConstantPeriod offset, got %T", offset)
					return
				}
				if constOffset.value != tt.expectedOffset {
					t.Errorf("Offset = %d, want %d", constOffset.value, tt.expectedOffset)
				}
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
				/* Arrow functions use ctx.Data pattern - no access to main scope Series variables */
				if !strings.Contains(code, "ctx.Data[ctx.BarIndex-j].Close") {
					t.Errorf("Expected ctx.Data pattern, got %s", code)
				}
			},
		},
		{
			name: "OHLCV high",
			expr: &ast.Identifier{Name: "high"},
			checkType: func(t *testing.T, gen AccessGenerator) {
				code := gen.GenerateLoopValueAccess("j")
				/* Arrow functions use ctx.Data pattern - no access to main scope Series variables */
				if !strings.Contains(code, "ctx.Data[ctx.BarIndex-j].High") {
					t.Errorf("Expected ctx.Data pattern, got %s", code)
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
				/* Arrow functions use ctx.Data pattern - no access to main scope Series variables */
				if !strings.Contains(code, "ctx.Data[ctx.BarIndex-j].Close") {
					t.Errorf("Expected ctx.Data pattern, got %s", code)
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
			gen := newTestArrowTAGenerator(g)

			accessor, err := gen.accessorFactory.CreateAccessorForExpression(tt.expr)

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
			gen := newTestArrowTAGenerator(g)

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
		name                 string
		expr                 ast.Expression
		variables            map[string]string
		expected             int
		expectRuntimePeriod  bool
		runtimeVariableName  string
		expectComputedPeriod bool
		expectError          bool
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
			name:                "parameter identifier - creates runtime period",
			expr:                &ast.Identifier{Name: "len"},
			variables:           map[string]string{"len": "float"},
			expectRuntimePeriod: true,
			runtimeVariableName: "len",
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
			name: "binary expression - creates computed period",
			expr: &ast.BinaryExpression{
				Left:     &ast.Identifier{Name: "len"},
				Operator: "/",
				Right:    &ast.Literal{Value: 2.0},
			},
			variables:            map[string]string{"len": "float"},
			expectComputedPeriod: true,
		},
		{
			name: "call expression - creates computed period",
			expr: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "math.round"},
				Arguments: []ast.Expression{
					&ast.CallExpression{
						Callee:    &ast.Identifier{Name: "math.sqrt"},
						Arguments: []ast.Expression{&ast.Identifier{Name: "len"}},
					},
				},
			},
			variables:            map[string]string{"len": "float"},
			expectComputedPeriod: true,
		},
		{
			name:        "malformed binary expression - nil operands",
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
			gen := newTestArrowTAGenerator(g)

			period, err := gen.extractPeriodExpression(tt.expr)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			/* Validate period type and value */
			if tt.expectComputedPeriod {
				computedPeriod, ok := period.(*ComputedPeriod)
				if !ok {
					t.Errorf("Expected ComputedPeriod, got %T", period)
					return
				}
				if computedPeriod.IsConstant() {
					t.Error("ComputedPeriod must not be constant")
				}
				if computedPeriod.AsInt() != -1 {
					t.Errorf("ComputedPeriod.AsInt() = %d, want -1", computedPeriod.AsInt())
				}
			} else if tt.expectRuntimePeriod {
				runtimePeriod, ok := period.(*RuntimePeriod)
				if !ok {
					t.Errorf("Expected RuntimePeriod, got %T", period)
					return
				}
				if runtimePeriod.variableName != tt.runtimeVariableName {
					t.Errorf("RuntimePeriod variable = %q, want %q", runtimePeriod.variableName, tt.runtimeVariableName)
				}
			} else {
				constPeriod, ok := period.(*ConstantPeriod)
				if !ok {
					t.Errorf("Expected ConstantPeriod, got %T", period)
					return
				}
				if constPeriod.value != tt.expected {
					t.Errorf("Period = %d, want %d", constPeriod.value, tt.expected)
				}
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
		{
			name: "change without arguments",
			call: &ast.CallExpression{
				Callee:    &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{},
			},
			expectError: true,
		},
		{
			name: "change with ta prefix and valid args",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "ta.change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "close"},
					&ast.Literal{Value: 2.0},
				},
			},
			expectError: false,
		},
		{
			name: "change with only source argument",
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "high"},
				},
			},
			expectError: false,
		},
		{
			name: "change with runtime parameter offset",
			setup: func(g *generator) {
				g.variables["offset"] = "float"
			},
			call: &ast.CallExpression{
				Callee: &ast.Identifier{Name: "change"},
				Arguments: []ast.Expression{
					&ast.Identifier{Name: "low"},
					&ast.Identifier{Name: "offset"},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newTestGenerator()
			if tt.setup != nil {
				tt.setup(g)
			}
			gen := newTestArrowTAGenerator(g)

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
				"func indicator(arrowCtx *context.ArrowContext, period float64) float64",
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
				"func bands(arrowCtx *context.ArrowContext, len float64, mult float64) float64",
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
				"func smoothed(arrowCtx *context.ArrowContext, srcSeries *series.Series, len float64) float64",
				"alpha",
				"ema",
				"srcSeries.Get(",
				"arrowCtx_smoothed_1 := context.NewArrowContext(ctx)",
				"smoothed(arrowCtx_smoothed_1, closeSeries, 14.0)",
			},
		},
		{
			name: "arrow function with computed period (Hull MA pattern)",
			script: `//@version=5
indicator("Test")
hull(src, n) =>
    h = ta.wma(src, n / 2)
    f = ta.wma(src, n)
    d = 2 * h - f
    ta.wma(d, math.round(math.sqrt(n)))
result = hull(close, 16)
plot(result)`,
			mustContain: []string{
				"func hull(arrowCtx *context.ArrowContext",
				"(n / 2)",
			},
			mustNotContain: []string{
				"_wma_20_",
				"unsupported period",
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
				if !strings.Contains(code.UserDefinedFunctions+code.FunctionBody, want) {
					t.Errorf("Missing %q in generated code", want)
				}
			}

			for _, notWant := range tt.mustNotContain {
				if strings.Contains(code.UserDefinedFunctions+code.FunctionBody, notWant) {
					t.Errorf("Unexpected %q in generated code", notWant)
				}
			}
		})
	}
}
