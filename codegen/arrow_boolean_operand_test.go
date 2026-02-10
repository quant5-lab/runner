package codegen

import (
	"strings"
	"testing"
)

/*
TestArrowFunction_IdentifierTruthiness validates float64→bool conversion in arrow function contexts.

PineScript stores all values as float64. When identifiers appear in boolean contexts
(logical operators, unary not), they need truthiness conversion via != 0.
Comparison expressions are already boolean and skip conversion.
*/
func TestArrowFunction_IdentifierTruthiness(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
	}{
		{
			name: "not with comparison operand preserves expression",
			pine: `
//@version=5
indicator("Test")
check(x) =>
    result = not(x > 10)
    result
plot(check(close))
`,
			mustContainAll: []string{
				"!(x > 10)",
			},
		},
		{
			name: "logical AND converts identifier operands to truthiness",
			pine: `
//@version=5
indicator("Test")
check(x, y) =>
    a = x > 10
    b = y < 5
    result = a and b
    result
plot(check(close, open))
`,
			mustContainAll: []string{
				"((a != 0) && (b != 0))",
			},
		},
		{
			name: "logical OR converts identifier operands to truthiness",
			pine: `
//@version=5
indicator("Test")
check(x, y) =>
    a = x > 10
    b = y < 5
    result = a or b
    result
plot(check(close, open))
`,
			mustContainAll: []string{
				"((a != 0) || (b != 0))",
			},
		},
		{
			name: "comparison operands in logical skip truthiness conversion",
			pine: `
//@version=5
indicator("Test")
check(x, y) =>
    result = (x > 10) and (y < 5)
    result
plot(check(close, open))
`,
			mustContainAll: []string{
				"((x > 10) && (y < 5))",
			},
			forbiddenPattern: []string{
				"!= 0",
			},
		},
		{
			name: "not with identifier operand converts to truthiness",
			pine: `
//@version=5
indicator("Test")
check(x) =>
    flag = x > 10
    result = not flag
    result
plot(check(close))
`,
			mustContainAll: []string{
				"!(flag != 0)",
			},
		},
		{
			name: "ternary test with identifier converts to truthiness",
			pine: `
//@version=5
indicator("Test")
pick(x) =>
    condition = x > 50
    result = condition ? 1 : 0
    result
plot(pick(close))
`,
			mustContainAll: []string{
				"(condition != 0)",
			},
		},
		{
			name: "ternary test with comparison preserves expression",
			pine: `
//@version=5
indicator("Test")
pick(x) =>
    result = x > 50 ? 1 : 0
    result
plot(pick(close))
`,
			mustContainAll: []string{
				"if (x > 50)",
			},
			forbiddenPattern: []string{
				"!= 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated:\n%s", pattern, code)
				}
			}
			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("found forbidden pattern: %q\ngenerated:\n%s", forbidden, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_ReturnTypeCoercion validates bool→float64 conversion at arrow function return points.

Arrow functions return float64. Boolean expressions (comparisons, logical) must be converted
to 1.0/0.0 at the return point. Arithmetic expressions pass through unchanged.
*/
func TestArrowFunction_ReturnTypeCoercion(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
	}{
		{
			name: "comparison return converts to float64 branch",
			pine: `
//@version=5
indicator("Test")
isBull() =>
    close > open
plot(isBull())
`,
			mustContainAll: []string{
				"if (bar.Close > bar.Open) { return 1.0 }",
				"return 0.0",
			},
			forbiddenPattern: []string{
				"return (bar.Close > bar.Open)\n",
			},
		},
		{
			name: "logical expression return converts to float64 branch",
			pine: `
//@version=5
indicator("Test")
checkBoth(x, y) =>
    x > 0 and y > 0
plot(checkBoth(close, open))
`,
			mustContainAll: []string{
				"if ((x > 0) && (y > 0)) { return 1.0 }",
				"return 0.0",
			},
		},
		{
			name: "arithmetic return passes through unchanged",
			pine: `
//@version=5
indicator("Test")
double(v) =>
    v * 2
plot(double(close))
`,
			mustContainAll: []string{
				"return (v * 2)",
			},
			forbiddenPattern: []string{
				"return 1.0",
				"return 0.0",
			},
		},
		{
			name: "unary not return converts to float64 branch",
			pine: `
//@version=5
indicator("Test")
invert(x) =>
    not(x > 0)
plot(invert(close))
`,
			mustContainAll: []string{
				"if !(x > 0) { return 1.0 }",
				"return 0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated:\n%s", pattern, code)
				}
			}
			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("found forbidden pattern: %q\ngenerated:\n%s", forbidden, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_VariableDeclarationCoercion validates bool→float64 at variable assignment.

Local variables in arrow functions store into Series.Set(float64). Boolean expressions
must be coerced to float64 at declaration point via IIFE wrapping.
*/
func TestArrowFunction_VariableDeclarationCoercion(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
	}{
		{
			name: "comparison variable gets IIFE wrap",
			pine: `
//@version=5
indicator("Test")
calc(x) =>
    above = x > 100
    above
plot(calc(close))
`,
			mustContainAll: []string{
				"above := func() float64 { if (x > 100) { return 1.0 } else { return 0.0 } }()",
			},
		},
		{
			name: "logical variable gets IIFE wrap",
			pine: `
//@version=5
indicator("Test")
calc(x, y) =>
    valid = x > 0 and y > 0
    valid
plot(calc(close, open))
`,
			mustContainAll: []string{
				"valid := func() float64 { if ((x > 0) && (y > 0)) { return 1.0 } else { return 0.0 } }()",
			},
		},
		{
			name: "arithmetic variable passes through without IIFE",
			pine: `
//@version=5
indicator("Test")
calc(x) =>
    diff = x - 100
    diff
plot(calc(close))
`,
			mustContainAll: []string{
				"diff := (x - 100)",
			},
			forbiddenPattern: []string{
				"func() float64 { if",
			},
		},
		{
			name: "ternary variable gets IIFE for branches",
			pine: `
//@version=5
indicator("Test")
calc(x) =>
    clamped = x > 100 ? 100 : x
    clamped
plot(calc(close))
`,
			mustContainAll: []string{
				"clamped := func() float64 { if (x > 100) { return 100 } else { return x } }()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated:\n%s", pattern, code)
				}
			}
			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("found forbidden pattern: %q\ngenerated:\n%s", forbidden, code)
				}
			}
		})
	}
}

/*
TestArrowFunction_ConditionalBranchCoercion validates ternary branch bool→float64 coercion.

When ternary branches produce boolean expressions, they must be coerced to float64
so the overall IIFE returns float64 consistently.
*/
func TestArrowFunction_ConditionalBranchCoercion(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
	}{
		{
			name: "bool literal branches become 1.0 and 0.0",
			pine: `
//@version=5
indicator("Test")
flag(x) =>
    result = x > 100 ? true : false
    result
plot(flag(close))
`,
			mustContainAll: []string{
				"return 1.0",
				"return 0.0",
			},
			forbiddenPattern: []string{
				"return true",
				"return false",
			},
		},
		{
			name: "numeric branches pass through",
			pine: `
//@version=5
indicator("Test")
clamp(x) =>
    result = x > 100 ? 100 : x
    result
plot(clamp(close))
`,
			mustContainAll: []string{
				"return 100",
				"return x",
			},
		},
		{
			name: "comparison branch gets nested IIFE",
			pine: `
//@version=5
indicator("Test")
nested(x, y) =>
    result = x > 50 ? (y > 50) : false
    result
plot(nested(close, open))
`,
			mustContainAll: []string{
				"if (x > 50)",
				"func() float64 { if (y > 50) { return 1.0 } else { return 0.0 } }()",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("compilation failed: %v", err)
			}
			for _, pattern := range tt.mustContainAll {
				if !strings.Contains(code, pattern) {
					t.Errorf("missing expected pattern: %q\ngenerated:\n%s", pattern, code)
				}
			}
			for _, forbidden := range tt.forbiddenPattern {
				if strings.Contains(code, forbidden) {
					t.Errorf("found forbidden pattern: %q\ngenerated:\n%s", forbidden, code)
				}
			}
		})
	}
}
