package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/parser"
)

/* TestArrowFunctionCodegen_SignatureGeneration validates function signature construction */
func TestArrowFunctionCodegen_SignatureGeneration(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		params      []ast.Identifier
		expectedSig string
	}{
		{
			name:        "zero parameters",
			funcName:    "simple",
			params:      []ast.Identifier{},
			expectedSig: "func simple(arrowCtx *context.ArrowContext)",
		},
		{
			name:     "single parameter",
			funcName: "getValue",
			params: []ast.Identifier{
				{Name: "period"},
			},
			expectedSig: "func getValue(arrowCtx *context.ArrowContext, period float64)",
		},
		{
			name:     "multiple parameters",
			funcName: "calculate",
			params: []ast.Identifier{
				{Name: "len"},
				{Name: "mult"},
			},
			expectedSig: "func calculate(arrowCtx *context.ArrowContext, len float64, mult float64)",
		},
		{
			name:     "parameter name preservation",
			funcName: "custom",
			params: []ast.Identifier{
				{Name: "myLength"},
				{Name: "myMultiplier"},
				{Name: "threshold"},
			},
			expectedSig: "func custom(arrowCtx *context.ArrowContext, myLength float64, myMultiplier float64, threshold float64)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			afc := NewArrowFunctionCodegen(gen)

			arrowFunc := &ast.ArrowFunctionExpression{
				Params: tt.params,
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.Literal{Value: 1.0},
					},
				},
			}

			analyzer := NewParameterUsageAnalyzer()
			paramTypes := analyzer.AnalyzeArrowFunction(arrowFunc)

			signature, _, err := afc.analyzeAndGenerateSignature(tt.funcName, arrowFunc, paramTypes)
			if err != nil {
				t.Fatalf("analyzeAndGenerateSignature() error: %v", err)
			}

			if signature != tt.expectedSig {
				t.Errorf("Signature mismatch:\nGot:  %q\nWant: %q", signature, tt.expectedSig)
			}
		})
	}
}

/* TestArrowFunctionCodegen_ReturnTypeInference validates return type detection */
func TestArrowFunctionCodegen_ReturnTypeInference(t *testing.T) {
	tests := []struct {
		name         string
		body         []ast.Node
		expectedType string
	}{
		{
			name: "single value return from expression",
			body: []ast.Node{
				&ast.ExpressionStatement{
					Expression: &ast.Literal{Value: 42.0},
				},
			},
			expectedType: "float64",
		},
		{
			name: "single value return from variable",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID:   &ast.Identifier{Name: "result"},
							Init: &ast.Literal{Value: 10.0},
						},
					},
				},
			},
			expectedType: "float64",
		},
		{
			name: "tuple return from array pattern",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.ArrayPattern{
								Elements: []ast.Identifier{
									{Name: "a"},
									{Name: "b"},
								},
							},
							Init: &ast.CallExpression{
								Callee: &ast.Identifier{Name: "ta.rma"},
							},
						},
					},
				},
			},
			expectedType: "(float64, float64)",
		},
		{
			name: "tuple return with three values",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.ArrayPattern{
								Elements: []ast.Identifier{
									{Name: "x"},
									{Name: "y"},
									{Name: "z"},
								},
							},
							Init: &ast.Literal{Value: nil},
						},
					},
				},
			},
			expectedType: "(float64, float64, float64)",
		},
		{
			name: "single element array pattern",
			body: []ast.Node{
				&ast.VariableDeclaration{
					Declarations: []ast.VariableDeclarator{
						{
							ID: &ast.ArrayPattern{
								Elements: []ast.Identifier{
									{Name: "value"},
								},
							},
							Init: &ast.Literal{Value: nil},
						},
					},
				},
			},
			expectedType: "float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			afc := NewArrowFunctionCodegen(gen)

			arrowFunc := &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body:   tt.body,
			}

			returnType, err := afc.inferReturnType(arrowFunc)
			if err != nil {
				t.Fatalf("inferReturnType() error: %v", err)
			}

			if returnType != tt.expectedType {
				t.Errorf("Return type mismatch:\nGot:  %q\nWant: %q", returnType, tt.expectedType)
			}
		})
	}
}

/* TestArrowFunctionCodegen_ParameterHandling validates parameter registration and usage */
func TestArrowFunctionCodegen_ParameterHandling(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
	}{
		{
			name: "parameter used in binary expression",
			script: `//@version=4
study("Test")
calc(len) =>
    close * len
result = calc(5)
plot(result)`,
			mustContain: []string{
				"func calc(arrowCtx *context.ArrowContext, len float64) float64",
				"return",
			},
		},
		{
			name: "parameter passed to TA function",
			script: `//@version=4
study("Test")
avg(period) =>
    sma(close, period)
result = avg(14)
plot(result)`,
			mustContain: []string{
				"func avg(arrowCtx *context.ArrowContext, period float64) float64",
				"func() float64",
				"sum",
			},
		},
		{
			name: "multiple parameters in computation",
			script: `//@version=4
study("Test")
band(len, mult) =>
    sma(close, len) + stdev(close, len) * mult
upper = band(20, 2)
plot(upper)`,
			mustContain: []string{
				"func band(arrowCtx *context.ArrowContext, len float64, mult float64) float64",
				"return",
			},
		},
		{
			name: "parameter in conditional expression",
			script: `//@version=4
study("Test")
signal(threshold) =>
    close > threshold ? 1 : 0
s = signal(100)
plot(s)`,
			mustContain: []string{
				"func signal(arrowCtx *context.ArrowContext, threshold float64) float64",
				"if",
				"threshold",
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
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q in generated code", want)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_BodyGeneration validates statement generation in function body */
func TestArrowFunctionCodegen_BodyGeneration(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "single statement body",
			script: `//@version=4
study("Test")
double(x) =>
    x * 2
result = double(close)
plot(result)`,
			mustContain: []string{
				"func double(arrowCtx *context.ArrowContext, x float64) float64",
				"return",
				"x * 2",
			},
			mustNotContain: []string{
				"undefined",
			},
		},
		{
			name: "multi-statement body with variable",
			script: `//@version=4
study("Test")
compute(len) =>
   
    avg = sma(close, len)
    dev = stdev(close, len)
    avg + dev
result = compute(20)
plot(result)`,
			mustContain: []string{
				"func compute(arrowCtx *context.ArrowContext, len float64) float64",
				"avg :=",
				"dev :=",
				"return",
			},
			mustNotContain: []string{
				"undefined",
			},
		},
		{
			name: "body with TA function calls",
			script: `//@version=4
study("Test")
indicator(period) =>
   
    ma = sma(close, period)
    upper = ma + stdev(close, period) * 2
    upper
signal = indicator(14)
plot(signal)`,
			mustContain: []string{
				"func indicator(arrowCtx *context.ArrowContext, period float64) float64",
				"func() float64",
				"sum",
			},
			mustNotContain: []string{
				"undefined",
			},
		},
		{
			name: "body with conditional logic",
			script: `//@version=4
study("Test")
check(threshold) =>
   
    value = close > open ? 1 : -1
    value * threshold
result = check(2.0)
plot(result)`,
			mustContain: []string{
				"func check(arrowCtx *context.ArrowContext, threshold float64) float64",
				"if",
				"return",
			},
			mustNotContain: []string{
				"undefined",
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
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q in generated code", want)
				}
			}

			for _, notWant := range tt.mustNotContain {
				if strings.Contains(code.UserDefinedFunctions, notWant) {
					t.Errorf("Found unexpected pattern %q in generated code", notWant)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_TupleReturns validates multi-value return generation */
func TestArrowFunctionCodegen_TupleReturns(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
	}{
		{
			name: "two-element tuple return",
			script: `//@version=4
study("Test")
pair() =>
    a = 1.0
    b = 2.0
    [a, b]
[x, y] = pair()
plot(x)`,
			mustContain: []string{
				"func pair(arrowCtx *context.ArrowContext) (float64, float64)",
				"return",
			},
		},
		{
			name: "tuple return with computation",
			script: `//@version=4
study("Test")
bounds(len) =>
    avg = sma(close, len)
    dev = stdev(close, len)
    lower = avg - dev
    upper = avg + dev
    [lower, upper]
[lower, upper] = bounds(20)
plot(lower)`,
			mustContain: []string{
				"func bounds(arrowCtx *context.ArrowContext, len float64) (float64, float64)",
				"return",
			},
		},
		{
			name: "tuple with intermediate variables",
			script: `//@version=4
study("Test")
minmax(len) =>
    h = highest(len)
    l = lowest(len)
    [l, h]
[min, max] = minmax(10)
plot(min)`,
			mustContain: []string{
				"func minmax(arrowCtx *context.ArrowContext, len float64) (float64, float64)",
				"h :=",
				"l :=",
				"return l, h",
			},
		},
		{
			name: "three-element tuple",
			script: `//@version=4
study("Test")
triple() =>
    o = open
    h = high
    l = low
    [o, h, l]
[o, h, l] = triple()
plot(o)`,
			mustContain: []string{
				"func triple(arrowCtx *context.ArrowContext) (float64, float64, float64)",
				"return",
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
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q in generated code", want)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_EdgeCases validates error handling and boundary conditions */
func TestArrowFunctionCodegen_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		arrowFunc   *ast.ArrowFunctionExpression
		expectError bool
		errorSubstr string
	}{
		{
			name: "empty body",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body:   []ast.Node{},
			},
			expectError: true,
			errorSubstr: "empty body",
		},
		{
			name: "expression statement return",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.BinaryExpression{
							Left:     &ast.Literal{Value: 1},
							Operator: "+",
							Right:    &ast.Literal{Value: 2},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid single expression",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{{Name: "x"}},
				Body: []ast.Node{
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "x"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid tuple with array pattern",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{
							{
								ID: &ast.ArrayPattern{
									Elements: []ast.Identifier{
										{Name: "a"},
										{Name: "b"},
									},
								},
								Init: &ast.Literal{Value: nil},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty variable declarator",
			arrowFunc: &ast.ArrowFunctionExpression{
				Params: []ast.Identifier{},
				Body: []ast.Node{
					&ast.VariableDeclaration{
						Declarations: []ast.VariableDeclarator{},
					},
				},
			},
			expectError: true,
			errorSubstr: "empty variable declaration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGenerator()
			afc := NewArrowFunctionCodegen(gen)

			_, err := afc.Generate("testFunc", tt.arrowFunc)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Error %q does not contain substring %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_IntegrationWithVariableDeclaration validates full codegen flow */
func TestArrowFunctionCodegen_IntegrationWithVariableDeclaration(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "arrow function defined and called",
			script: `//@version=4
study("Test")
double(x) =>
    x * 2
result = double(close)
plot(result)`,
			mustContain: []string{
				"func double(arrowCtx *context.ArrowContext, x float64) float64",
				"return",
			},
			mustNotContain: []string{
				"not yet implemented",
				"undefined",
			},
		},
		{
			name: "multiple arrow functions",
			script: `//@version=4
study("Test")
add(a, b) =>
    a + b
multiply(x, y) =>
    x * y
result = add(multiply(close, 2), 10)
plot(result)`,
			mustContain: []string{
				"func add(arrowCtx *context.ArrowContext, a float64, b float64) float64",
				"func multiply(arrowCtx *context.ArrowContext, x float64, y float64) float64",
			},
			mustNotContain: []string{
				"not yet implemented",
			},
		},
		{
			name: "arrow function with tuple used in assignment",
			script: `//@version=4
study("Test")
range(len) =>
    [highest(len), lowest(len)]
[h, l] = range(10)
plot(h - l)`,
			mustContain: []string{
				"func range(arrowCtx *context.ArrowContext, len float64) (float64, float64)",
				"return",
			},
			mustNotContain: []string{
				"not yet implemented",
			},
		},
		{
			name: "arrow function before and after usage",
			script: `//@version=4
study("Test")
helper(n) =>
    n * 2
value = helper(5)
another(m) =>
    m + 1
plot(value + another(3))`,
			mustContain: []string{
				"func helper(arrowCtx *context.ArrowContext, n float64) float64",
				"func another(arrowCtx *context.ArrowContext, m float64) float64",
			},
			mustNotContain: []string{
				"not yet implemented",
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
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q in generated code", want)
				}
			}

			for _, notWant := range tt.mustNotContain {
				if strings.Contains(code.UserDefinedFunctions, notWant) {
					t.Errorf("Found unexpected pattern %q in generated code", notWant)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_ParameterShadowing validates parameter scope isolation */
func TestArrowFunctionCodegen_ParameterShadowing(t *testing.T) {
	script := `//@version=4
study("Test")
len = 10
myFunc(len) =>
    sma(close, len)
result = myFunc(20)
plot(result)
`

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

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
	}

	if !strings.Contains(code.UserDefinedFunctions, "func myFunc(arrowCtx *context.ArrowContext, len float64)") {
		t.Error("Function should have len parameter")
	}

	if !strings.Contains(code.FunctionBody, "myFunc(ctx, 20.0)") {
		t.Error("Function should be called with 20.0, not global len")
	}
}

/* TestArrowFunctionCodegen_ExpressionTypes validates various expression handling */
func TestArrowFunctionCodegen_ExpressionTypes(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		mustContain []string
	}{
		{
			name: "literal return",
			script: `//@version=4
study("Test")
constant() =>
    42.0
result = constant()
plot(result)`,
			mustContain: []string{
				"return 42",
			},
		},
		{
			name: "identifier return",
			script: `//@version=4
study("Test")
getClose() =>
    close
result = getClose()
plot(result)`,
			mustContain: []string{
				"return close",
			},
		},
		{
			name: "binary expression return",
			script: `//@version=4
study("Test")
diff() =>
    close - open
result = diff()
plot(result)`,
			mustContain: []string{
				"return",
				"-",
			},
		},
		{
			name: "call expression return",
			script: `//@version=4
study("Test")
average(len) =>
    sma(close, len)
result = average(14)
plot(result)`,
			mustContain: []string{
				"func() float64",
				"return",
			},
		},
		{
			name: "member expression return",
			script: `//@version=4
study("Test")
getEquity() =>
    strategy.equity
result = getEquity()
plot(result)`,
			mustContain: []string{
				"return",
				"strategy",
			},
		},
		{
			name: "conditional expression return",
			script: `//@version=4
study("Test")
signal() =>
    close > open ? 1 : -1
result = signal()
plot(result)`,
			mustContain: []string{
				"if",
				"return",
			},
		},
		{
			name: "unary expression return",
			script: `//@version=4
study("Test")
negate(x) =>
    -x
result = negate(close)
plot(result)`,
			mustContain: []string{
				"return",
				"-",
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
				if !strings.Contains(code.UserDefinedFunctions, want) {
					t.Errorf("Missing pattern %q in generated code", want)
				}
			}
		})
	}
}

/* TestArrowFunctionCodegen_FunctionCallWithParameters validates invocation generation */
func TestArrowFunctionCodegen_FunctionCallWithParameters(t *testing.T) {
	script := `//@version=4
study("Test")
calc(multiplier, offset) =>
    close * multiplier + offset
result = calc(2.0, 10.0)
plot(result)
`

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

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error: %v", err)
	}

	if !strings.Contains(code.UserDefinedFunctions, "func calc(arrowCtx *context.ArrowContext, multiplier float64, offset float64)") {
		t.Error("Function signature incorrect")
	}

	if !strings.Contains(code.FunctionBody, "calc(ctx, 2.0, 10.0)") {
		t.Error("Function invocation should include ctx and arguments")
	}
}
