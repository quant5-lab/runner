package codegen

import (
	"strings"
	"testing"
)

/*
TestArrowSeriesParameter_CodegenSignature validates that arrow functions whose parameters
are used as series (passed to TA functions, subscripted) receive *series.Series typed
parameters in the generated Go signature, while scalar-used parameters remain float64.

ForwardSeriesBuffer paradigm: series parameters are passed as *series.Series and accessed
via paramSeries.GetCurrent() for the current bar value. Scalar parameters are passed as
float64 and accessed directly.
*/
func TestArrowSeriesParameter_CodegenSignature(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "series-used parameter generates *series.Series signature",
			pine: `
//@version=5
indicator("Test")
mySma(src, len) =>
    ta.sma(src, len)
plot(mySma(close, 14))
`,
			mustContainAll: []string{
				"srcSeries *series.Series",
				"srcSeries.Get(",
			},
			forbiddenPattern: []string{
				"src float64",
			},
			description: "src passed to ta.sma as series → *series.Series parameter type with Get() access inside TA",
		},
		{
			name: "scalar-only parameter remains float64",
			pine: `
//@version=5
indicator("Test")
scale(x, factor) =>
    x * factor
plot(scale(close, 2.0))
`,
			mustContainAll: []string{
				"factor float64",
			},
			forbiddenPattern: []string{
				"factorSeries *series.Series",
			},
			description: "factor used only in multiplication → float64 parameter type",
		},
		{
			name: "mixed: series and scalar parameters in same function",
			pine: `
//@version=5
indicator("Test")
scaledSma(src, len, mult) =>
    ta.sma(src, len) * mult
plot(scaledSma(close, 14, 2.0))
`,
			mustContainAll: []string{
				"srcSeries *series.Series",
				"len float64",
				"mult float64",
			},
			forbiddenPattern: []string{
				"src float64",
				"multSeries *series.Series",
			},
			description: "src is series-typed, len and mult are scalar",
		},
		{
			name: "series parameter accessed via GetCurrent() when used in direct expression",
			pine: `
//@version=5
indicator("Test")
normalize(src, len) =>
    avg = ta.sma(src, len)
    src_val = src
    src_val - avg
plot(normalize(close, 10))
`,
			mustContainAll: []string{
				"srcSeries *series.Series",
				"srcSeries.GetCurrent()",
			},
			forbiddenPattern: []string{
				"src float64",
				"srcSeries.GetCurrent().GetCurrent()",
			},
			description: "series parameter accessed in direct expression resolves to GetCurrent()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/*
TestArrowV3BuiltinAlias validates that Pine v3 identifier aliases resolve to their v4
equivalents and generate correct access code. The alias `n` must behave identically to
`bar_index` in all arrow-function and top-level contexts.

Access patterns:
  - Top-level bar scope:    bar_index → float64(i)       (i = main bar-loop counter)
  - Arrow function scope:   bar_index → float64(ctx.BarIndex)
*/
func TestArrowV3BuiltinAlias(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "n alias resolves to bar_index value in top-level expression",
			pine: `
//@version=5
indicator("Test")
x = n
plot(x)
`,
			mustContainAll: []string{
				"float64(i)",
			},
			forbiddenPattern: []string{
				"var nSeries",
				"nSeries.GetCurrent()",
			},
			description: "Pine v3 n alias resolves to float64(i) (bar_index) at top-level bar scope",
		},
		{
			name: "n alias in arithmetic expression at top-level",
			pine: `
//@version=5
indicator("Test")
x = n + 1
plot(x)
`,
			mustContainAll: []string{
				"float64(i)",
			},
			forbiddenPattern: []string{
				"var nSeries",
				"nSeries.GetCurrent()",
			},
			description: "Pine v3 n in arithmetic resolves via alias to bar_index (float64(i))",
		},
		{
			name: "bar_index and n produce identical access code",
			pine: `
//@version=5
indicator("Test")
a = bar_index
b = n
plot(a + b)
`,
			mustContainAll: []string{
				"float64(i)",
			},
			forbiddenPattern: []string{
				"var nSeries",
				"nSeries.GetCurrent()",
			},
			description: "bar_index and n alias produce identical float64(i) access at top-level",
		},
		{
			name: "n alias in arrow function body",
			pine: `
//@version=5
indicator("Test")
getBar(mult) =>
    n * mult
plot(getBar(2.0))
`,
			mustContainAll: []string{
				"float64(ctx.BarIndex)",
			},
			forbiddenPattern: []string{
				"var nSeries",
				"nSeries.GetCurrent()",
			},
			description: "n alias inside arrow function resolves to float64(ctx.BarIndex)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/*
TestArrowForLoopIdentifierScope validates that all identifier categories — series parameters,
scalar parameters, local variables, loop-modified variables, and outer-scope series — resolve
correctly when accessed from within for-loop bodies inside arrow functions.
*/
func TestArrowForLoopIdentifierScope(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "series parameter subscripted in loop body uses Series.Get()",
			pine: `
//@version=5
indicator("Test")
sumSrc(src, len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + src[i]
    total
plot(sumSrc(close, 5))
`,
			mustContainAll: []string{
				"srcSeries *series.Series",
				"srcSeries.Get(int(float64(i)))",
			},
			forbiddenPattern: []string{
				"src float64",
			},
			description: "series parameter src subscripted in loop body resolves via srcSeries.Get(offset)",
		},
		{
			name: "scalar parameter in loop body accessed directly without series lookup",
			pine: `
//@version=5
indicator("Test")
accumulate(factor, len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + factor
    total
plot(accumulate(1.5, 10))
`,
			mustContainAll: []string{
				"factor float64",
				"totalSeries.Set(",
			},
			forbiddenPattern: []string{
				"factorSeries *series.Series",
			},
			description: "scalar parameter accessed in loop body stays float64, no series lookup",
		},
		{
			name: "outer scope series variable captured and passed as *series.Series parameter",
			pine: `
//@version=5
indicator("Test")
ref = 100.0
clamp(len) =>
    total = 0.0
    for i = 0 to len - 1
        total := total + ref
    total
plot(clamp(5))
`,
			mustContainAll: []string{
				"totalSeries.Set(",
				"refSeries *series.Series",
				"refSeries.GetCurrent()",
			},
			forbiddenPattern: []string{
				"arrowCtx.GetOrCreateSeries(\"ref\")",
				"ref float64",
			},
			description: "outer scope float series variable is captured as *series.Series injection, not a local arrow series",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/*
TestTopLevelForInIdentifierScope validates that for-in loop index and element variables
in top-level (main bar-loop) context resolve correctly, including int→float64 coercion
for index variables used in arithmetic.
*/
func TestTopLevelForInIdentifierScope(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "tuple index used alone resolves to float64-cast",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = i
`,
			mustContainAll: []string{
				"for i, val := range",
				"float64(i)",
			},
			forbiddenPattern: []string{
				"iSeries",
			},
			description: "bare tuple index used alone is float64-cast",
		},
		{
			name: "tuple index in division with float element",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = val / i
`,
			mustContainAll: []string{
				"for i, val := range",
				"float64(i)",
			},
			forbiddenPattern: []string{
				"iSeries",
			},
			description: "tuple index in arithmetic with element receives float64 cast",
		},
		{
			name: "tuple index in subtraction expression",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = val - i
`,
			mustContainAll: []string{
				"float64(i)",
				"val - float64(i)",
			},
			forbiddenPattern: []string{
				"var iSeries",
				"iSeries.GetCurrent()",
			},
			description: "tuple index in subtraction binary expression receives float64 cast",
		},
		{
			name: "element variable accessed without cast",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = val + 1
`,
			mustContainAll: []string{
				"val",
			},
			forbiddenPattern: []string{
				"valSeries",
				"float64(val)",
			},
			description: "element variable from range is already float64, no cast needed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}

/*
TestArrowOuterScopeCapture validates that outer-scope (strategy-level) variables referenced
inside arrow function bodies are correctly captured as injected parameters. Captures are
classified by the storage kind of the outer variable: float/bool series → *series.Series,
string → string, array_series → *series.ArraySeries, scalar inputs → float64.

Shadowing rules: formal parameters and locally declared variables within the arrow function
body suppress capture of any same-named outer variable.
*/
func TestArrowOuterScopeCapture(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "outer series variable injected as *series.Series parameter",
			pine: `
//@version=5
indicator("Test")
price = close
getVal() =>
    price
plot(getVal())
`,
			mustContainAll: []string{
				"priceSeries *series.Series",
				"priceSeries.GetCurrent()",
				"getVal(", // call site includes priceSeries arg
				"priceSeries)",
			},
			forbiddenPattern: []string{
				"price float64",
				"arrowCtx.GetOrCreateSeries(\"price\")",
			},
			description: "outer float series variable is injected as *series.Series, accessed via GetCurrent()",
		},
		{
			name: "outer series variable subscripted in arrow body uses Series.Get()",
			pine: `
//@version=5
indicator("Test")
price = close
prevVal() =>
    price[1]
plot(prevVal())
`,
			mustContainAll: []string{
				"priceSeries *series.Series",
				"priceSeries.Get(int(1))",
				"priceSeries)",
			},
			forbiddenPattern: []string{
				"price float64",
				"priceSeries.GetCurrent()",
			},
			description: "outer series variable subscripted inside arrow body resolves via Get(offset)",
		},
		{
			name: "multiple outer series variables each captured distinctly",
			pine: `
//@version=5
indicator("Test")
src1 = close
src2 = open
diff() =>
    src1 - src2
plot(diff())
`,
			mustContainAll: []string{
				"src1Series *series.Series",
				"src2Series *series.Series",
				"src1Series.GetCurrent()",
				"src2Series.GetCurrent()",
			},
			forbiddenPattern: []string{
				"src1 float64",
				"src2 float64",
			},
			description: "each outer series variable gets its own *series.Series parameter in the signature",
		},
		{
			name: "formal parameter shadows outer series — outer NOT captured as function param",
			pine: `
//@version=5
indicator("Test")
price = close
useParam(price) =>
    price
plot(useParam(close))
`,
			mustContainAll: []string{
				"price float64",
			},
			forbiddenPattern: []string{
				// The outer priceSeries exists in main scope but must NOT appear in the function signature
				"useParam(arrowCtx *context.ArrowContext, priceSeries",
			},
			description: "formal parameter named 'price' shadows the outer series 'price'; outer is not captured as function parameter",
		},
		{
			name: "local variable shadows outer series — outer NOT injected as capture param",
			pine: `
//@version=5
indicator("Test")
ref = close
localShadow() =>
    ref = 42.0
    ref
plot(localShadow())
`,
			mustContainAll: []string{
				"arrowCtx.GetOrCreateSeries(\"ref\")",
			},
			forbiddenPattern: []string{
				// The outer refSeries exists in main scope but must NOT appear in the function signature
				"localShadow(arrowCtx *context.ArrowContext, refSeries",
			},
			description: "locally declared 'ref' shadows outer 'ref'; outer series is not injected as extra capture parameter",
		},
		{
			name: "user-defined function name in outer scope is NOT captured",
			pine: `
//@version=5
indicator("Test")
helper(x) =>
    x * 2.0
caller() =>
    helper(close)
plot(caller())
`,
			mustContainAll: []string{
				"helper(",
			},
			forbiddenPattern: []string{
				"helperSeries *series.Series",
				"helper float64",
				"helper string",
			},
			description: "function-type outer variables are not injectable captures; calls go through the call router",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goCode, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(goCode, pattern) {
					t.Errorf("Missing required pattern: %q\nDescription: %s\nGenerated code:\n%s",
						pattern, tt.description, goCode)
				}
			}

			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(goCode, forbidden) {
					t.Errorf("Found forbidden pattern: %q\nDescription: %s\nGenerated code:\n%s",
						forbidden, tt.description, goCode)
				}
			}
		})
	}
}
