package codegen

import (
	"os"
	"strings"
	"testing"
)

/* TestMathFunctions_VariableAssignment validates math functions in top-level variable assignments */
func TestMathFunctions_VariableAssignment(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "unprefixed pow",
			pine: `
//@version=5
indicator("Test")
x = pow(close, 2.0)
`,
			mustContain: []string{"math.Pow(bar.Close, 2)"},
		},
		{
			name: "prefixed math.pow",
			pine: `
//@version=5
indicator("Test")
x = math.pow(close, 2.0)
`,
			mustContain: []string{"math.Pow(bar.Close, 2)"},
		},
		{
			name: "unprefixed abs",
			pine: `
//@version=5
indicator("Test")
x = abs(close)
`,
			mustContain: []string{"math.Abs(bar.Close)"},
		},
		{
			name: "prefixed math.abs",
			pine: `
//@version=5
indicator("Test")
x = math.abs(close)
`,
			mustContain: []string{"math.Abs(bar.Close)"},
		},
		{
			name: "unprefixed sqrt",
			pine: `
//@version=5
indicator("Test")
x = sqrt(close)
`,
			mustContain: []string{"math.Sqrt(bar.Close)"},
		},
		{
			name: "unprefixed max",
			pine: `
//@version=5
indicator("Test")
x = max(high, low)
`,
			mustContain: []string{"math.Max(bar.High, bar.Low)"},
		},
		{
			name: "unprefixed min",
			pine: `
//@version=5
indicator("Test")
x = min(high, low)
`,
			mustContain: []string{"math.Min(bar.High, bar.Low)"},
		},
		{
			name: "unprefixed log",
			pine: `
//@version=5
indicator("Test")
x = log(close)
`,
			mustContain: []string{"math.Log(bar.Close)"},
		},
		{
			name: "unprefixed log10",
			pine: `
//@version=5
indicator("Test")
x = log10(close)
`,
			mustContain: []string{"math.Log10(bar.Close)"},
		},
		{
			name: "unprefixed exp",
			pine: `
//@version=5
indicator("Test")
x = exp(close)
`,
			mustContain: []string{"math.Exp(bar.Close)"},
		},
		{
			name: "unprefixed floor",
			pine: `
//@version=5
indicator("Test")
x = floor(close)
`,
			mustContain: []string{"math.Floor(bar.Close)"},
		},
		{
			name: "unprefixed ceil",
			pine: `
//@version=5
indicator("Test")
x = ceil(close)
`,
			mustContain: []string{"math.Ceil(bar.Close)"},
		},
		{
			name: "unprefixed round",
			pine: `
//@version=5
indicator("Test")
x = round(close)
`,
			mustContain: []string{"math.Round(bar.Close)"},
		},
		{
			name: "unprefixed sign",
			pine: `
//@version=5
indicator("Test")
x = sign(close)
`,
			mustContain: []string{"v := bar.Close; if v > 0 { return 1 }"},
		},
		{
			name: "unprefixed sin",
			pine: `
//@version=5
indicator("Test")
x = sin(close)
`,
			mustContain: []string{"math.Sin(bar.Close)"},
		},
		{
			name: "unprefixed cos",
			pine: `
//@version=5
indicator("Test")
x = cos(close)
`,
			mustContain: []string{"math.Cos(bar.Close)"},
		},
		{
			name: "unprefixed tan",
			pine: `
//@version=5
indicator("Test")
x = tan(close)
`,
			mustContain: []string{"math.Tan(bar.Close)"},
		},
		{
			name: "unprefixed asin",
			pine: `
//@version=5
indicator("Test")
x = asin(close)
`,
			mustContain: []string{"math.Asin(bar.Close)"},
		},
		{
			name: "unprefixed acos",
			pine: `
//@version=5
indicator("Test")
x = acos(close)
`,
			mustContain: []string{"math.Acos(bar.Close)"},
		},
		{
			name: "unprefixed atan",
			pine: `
//@version=5
indicator("Test")
x = atan(close)
`,
			mustContain: []string{"math.Atan(bar.Close)"},
		},
		{
			name: "multiple math functions in one script",
			pine: `
//@version=5
indicator("Test")
a = abs(close - open)
b = sqrt(close)
c = max(high, low)
d = pow(close, 2.0)
`,
			mustContain: []string{
				"math.Abs(",
				"math.Sqrt(bar.Close)",
				"math.Max(bar.High, bar.Low)",
				"math.Pow(bar.Close, 2)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}

			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO placeholder\nGenerated code:\n%s", code)
			}
		})
	}
}

/* TestMathFunctions_PlotExpression validates math functions inside plot() calls */
func TestMathFunctions_PlotExpression(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "plot unprefixed abs",
			pine: `
//@version=5
indicator("Test")
plot(abs(close))
`,
			mustContain: []string{"math.Abs(bar.Close)"},
		},
		{
			name: "plot prefixed math.abs",
			pine: `
//@version=5
indicator("Test")
plot(math.abs(close))
`,
			mustContain: []string{"math.Abs(bar.Close)"},
		},
		{
			name: "plot unprefixed sqrt",
			pine: `
//@version=5
indicator("Test")
plot(sqrt(close))
`,
			mustContain: []string{"math.Sqrt(bar.Close)"},
		},
		{
			name: "plot unprefixed pow",
			pine: `
//@version=5
indicator("Test")
plot(pow(close, 2.0))
`,
			mustContain: []string{"math.Pow(bar.Close, 2)"},
		},
		{
			name: "plot unprefixed max",
			pine: `
//@version=5
indicator("Test")
plot(max(high, low))
`,
			mustContain: []string{"math.Max(bar.High, bar.Low)"},
		},
		{
			name: "plot unprefixed sign",
			pine: `
//@version=5
indicator("Test")
plot(sign(close))
`,
			mustContain: []string{"v := bar.Close; if v > 0 { return 1 }"},
		},
		{
			name: "plot unprefixed log10",
			pine: `
//@version=5
indicator("Test")
plot(log10(close))
`,
			mustContain: []string{"math.Log10(bar.Close)"},
		},
		{
			name: "plot unprefixed sin",
			pine: `
//@version=5
indicator("Test")
plot(sin(close))
`,
			mustContain: []string{"math.Sin(bar.Close)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}

			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO placeholder\nGenerated code:\n%s", code)
			}
		})
	}
}

/* TestMathFunctions_ArrowFunctionBody validates math functions inside user-defined arrow functions */
func TestMathFunctions_ArrowFunctionBody(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		mustContain []string
	}{
		{
			name: "arrow with unprefixed pow",
			pine: `
//@version=5
indicator("Test")
f(a, b) =>
    pow(a, b)
x = f(close, 2.0)
`,
			mustContain: []string{"math.Pow("},
		},
		{
			name: "arrow with unprefixed abs",
			pine: `
//@version=5
indicator("Test")
f(x) =>
    abs(x)
y = f(close)
`,
			mustContain: []string{"math.Abs("},
		},
		{
			name: "arrow with prefixed math.sqrt",
			pine: `
//@version=5
indicator("Test")
f(x) =>
    math.sqrt(x)
y = f(close)
`,
			mustContain: []string{"math.Sqrt("},
		},
		{
			name: "arrow with unprefixed max",
			pine: `
//@version=5
indicator("Test")
f(a, b) =>
    max(a, b)
y = f(high, low)
`,
			mustContain: []string{"math.Max("},
		},
		{
			name: "arrow with multiple math calls",
			pine: `
//@version=5
indicator("Test")
f(x, y) =>
    pow(abs(x), sqrt(y))
z = f(close, 4.0)
`,
			mustContain: []string{"math.Pow(", "math.Abs(", "math.Sqrt("},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}

			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO placeholder\nGenerated code:\n%s", code)
			}
		})
	}
}

/* TestMathFunctions_PrefixSymmetry validates prefixed and unprefixed forms produce identical output */
func TestMathFunctions_PrefixSymmetry(t *testing.T) {
	tests := []struct {
		name       string
		unprefixed string
		prefixed   string
		expected   string
	}{
		{
			name:       "pow symmetry",
			unprefixed: "x = pow(close, 2.0)",
			prefixed:   "x = math.pow(close, 2.0)",
			expected:   "math.Pow(bar.Close, 2)",
		},
		{
			name:       "abs symmetry",
			unprefixed: "x = abs(close)",
			prefixed:   "x = math.abs(close)",
			expected:   "math.Abs(bar.Close)",
		},
		{
			name:       "sqrt symmetry",
			unprefixed: "x = sqrt(close)",
			prefixed:   "x = math.sqrt(close)",
			expected:   "math.Sqrt(bar.Close)",
		},
		{
			name:       "max symmetry",
			unprefixed: "x = max(high, low)",
			prefixed:   "x = math.max(high, low)",
			expected:   "math.Max(bar.High, bar.Low)",
		},
		{
			name:       "log symmetry",
			unprefixed: "x = log(close)",
			prefixed:   "x = math.log(close)",
			expected:   "math.Log(bar.Close)",
		},
		{
			name:       "sign symmetry",
			unprefixed: "x = sign(close)",
			prefixed:   "x = math.sign(close)",
			expected:   "v := bar.Close; if v > 0 { return 1 }",
		},
	}

	wrap := func(stmt string) string {
		return "//@version=5\nindicator(\"Test\")\n" + stmt + "\n"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unprefixedCode, err := compilePineScript(wrap(tt.unprefixed))
			if err != nil {
				t.Fatalf("Unprefixed compilation failed: %v", err)
			}

			prefixedCode, err := compilePineScript(wrap(tt.prefixed))
			if err != nil {
				t.Fatalf("Prefixed compilation failed: %v", err)
			}

			if !strings.Contains(unprefixedCode, tt.expected) {
				t.Errorf("Unprefixed form missing expected output: %q\nGenerated:\n%s", tt.expected, unprefixedCode)
			}

			if !strings.Contains(prefixedCode, tt.expected) {
				t.Errorf("Prefixed form missing expected output: %q\nGenerated:\n%s", tt.expected, prefixedCode)
			}
		})
	}
}

/* TestMathFunctions_FixtureCompilation validates .pine fixture files compile through full codegen pipeline */
func TestMathFunctions_FixtureCompilation(t *testing.T) {
	fixturesDir := "../e2e/fixtures/strategies"

	tests := []struct {
		name        string
		fixture     string
		mustContain []string
	}{
		{
			name:    "unprefixed math functions",
			fixture: "test-math-unprefixed.pine",
			mustContain: []string{
				"math.Abs(", "math.Sqrt(", "math.Floor(", "math.Ceil(",
				"math.Round(", "math.Log(", "math.Log10(", "math.Exp(",
				"math.Pow(", "math.Max(", "math.Min(",
				"math.Sin(", "math.Cos(", "math.Tan(",
				"math.Asin(", "math.Acos(", "math.Atan(",
			},
		},
		{
			name:    "prefixed math functions",
			fixture: "test-math-prefixed.pine",
			mustContain: []string{
				"math.Abs(", "math.Sqrt(", "math.Floor(", "math.Ceil(",
				"math.Round(", "math.Log(", "math.Log10(", "math.Exp(",
				"math.Pow(", "math.Max(", "math.Min(",
				"math.Sin(", "math.Cos(", "math.Tan(",
				"math.Asin(", "math.Acos(", "math.Atan(",
			},
		},
		{
			name:    "math in plot expressions",
			fixture: "test-math-plot-inline.pine",
			mustContain: []string{
				"math.Abs(", "math.Sqrt(", "math.Pow(",
				"math.Max(", "math.Min(", "math.Log(",
				"math.Floor(", "math.Ceil(",
			},
		},
		{
			name:    "math in arrow functions",
			fixture: "test-math-arrow.pine",
			mustContain: []string{
				"math.Pow(", "math.Abs(", "math.Sqrt(", "math.Max(",
				"func myPow(", "func myAbs(", "func mySqrt(", "func myMax(",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(fixturesDir + "/" + tt.fixture)
			if err != nil {
				t.Fatalf("Failed to read fixture %s: %v", tt.fixture, err)
			}

			code, err := compilePineScript(string(content))
			if err != nil {
				t.Fatalf("Compilation failed for %s: %v", tt.fixture, err)
			}

			for _, pattern := range tt.mustContain {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing expected pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}

			if strings.Contains(code, "TODO") {
				t.Errorf("Generated code contains TODO placeholder\nGenerated code:\n%s", code)
			}
		})
	}
}

/* TestMathFunctions_FixtureSymmetry validates unprefixed and prefixed fixtures produce identical math calls */
func TestMathFunctions_FixtureSymmetry(t *testing.T) {
	fixturesDir := "../e2e/fixtures/strategies"

	unprefixedContent, err := os.ReadFile(fixturesDir + "/test-math-unprefixed.pine")
	if err != nil {
		t.Fatalf("Failed to read unprefixed fixture: %v", err)
	}
	prefixedContent, err := os.ReadFile(fixturesDir + "/test-math-prefixed.pine")
	if err != nil {
		t.Fatalf("Failed to read prefixed fixture: %v", err)
	}

	unprefixedCode, err := compilePineScript(string(unprefixedContent))
	if err != nil {
		t.Fatalf("Unprefixed fixture compilation failed: %v", err)
	}
	prefixedCode, err := compilePineScript(string(prefixedContent))
	if err != nil {
		t.Fatalf("Prefixed fixture compilation failed: %v", err)
	}

	/* Both fixtures declare identical variables (a-r) with identical math operations.
	   The generated Series.Set() calls must match regardless of prefix. */
	mathCalls := []string{
		"math.Abs(", "math.Sqrt(", "math.Floor(", "math.Ceil(",
		"math.Round(", "math.Log(", "math.Log10(", "math.Exp(",
		"math.Pow(", "math.Max(", "math.Min(",
		"math.Sin(", "math.Cos(", "math.Tan(",
		"math.Asin(", "math.Acos(", "math.Atan(",
	}

	for _, call := range mathCalls {
		inUnprefixed := strings.Contains(unprefixedCode, call)
		inPrefixed := strings.Contains(prefixedCode, call)

		if inUnprefixed != inPrefixed {
			t.Errorf("Symmetry broken for %s: unprefixed=%v, prefixed=%v", call, inUnprefixed, inPrefixed)
		}
		if !inUnprefixed {
			t.Errorf("Missing %s in unprefixed fixture output", call)
		}
	}
}
