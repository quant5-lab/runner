package codegen

import (
	"strings"
	"testing"
)

/* TestLoopExpressionResultTracking validates __result assignment for all expression types in IIFE loops */
func TestLoopExpressionResultTracking(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "bare identifier result in for",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    i
plot(x)
`,
			mustContainAll:   []string{"__result = float64(i)", "func() float64"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "bare identifier as last body statement assigns to __result",
		},
		{
			name: "binary expression result in for",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    i * 2
plot(x)
`,
			mustContainAll:   []string{"__result = float64(", "func() float64"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "binary expression as last body statement assigns to __result",
		},
		{
			name: "bare identifier result in for-in",
			pine: `
//@version=5
indicator("Test")
x = for val in close
    val
plot(x)
`,
			mustContainAll:   []string{"__result = float64(val)", "for _, val := range"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "bare identifier result in for-in IIFE",
		},
		{
			name: "binary expression result in for-in",
			pine: `
//@version=5
indicator("Test")
x = for val in close
    val + 1
plot(x)
`,
			mustContainAll:   []string{"__result = float64(", "for _, val := range"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "binary expression result in for-in IIFE",
		},
		{
			name: "no expression result falls back to zero",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    a = i + 1
plot(x)
`,
			mustContainAll:   []string{"__result = 0.0", "return __result"},
			forbiddenPattern: nil,
			description:      "non-expression last statement falls back to __result = 0.0",
		},
		{
			name: "result after multi-statement body in for",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    a = i + 1
    i
plot(x)
`,
			mustContainAll:   []string{"__result = float64(i)", "return __result"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "only last expression statement becomes __result, preceding statements normal",
		},
		{
			name: "result after multi-statement body in for-in",
			pine: `
//@version=5
indicator("Test")
x = for val in close
    temp = val * 2
    val
plot(x)
`,
			mustContainAll:   []string{"__result = float64(val)", "return __result"},
			forbiddenPattern: []string{"__result = 0.0"},
			description:      "for-in multi-statement body with last expression as result",
		},
		{
			name: "IIFE wrapping with var __result declaration",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 5
    i
plot(x)
`,
			mustContainAll:   []string{"var __result float64", "return __result", "}())"},
			forbiddenPattern: nil,
			description:      "IIFE wraps with __result declaration, return, and closure call",
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
