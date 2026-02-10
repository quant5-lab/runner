package codegen

import (
	"strings"
	"testing"
)

/* TestForInStatementCodegen validates top-level for-in code generation across all forms and contexts */
func TestForInStatementCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "single element range syntax",
			pine: `
//@version=5
strategy("Test")
for val in close
    x = val + 1
`,
			mustContainAll: []string{
				"for _, val := range",
			},
			forbiddenPattern: []string{
				"for val := range",
			},
			description: "single form uses blank index in range clause",
		},
		{
			name: "tuple element range syntax",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = val + i
`,
			mustContainAll: []string{
				"for i, val := range",
			},
			forbiddenPattern: []string{
				"for _, val := range",
			},
			description: "tuple form uses named index variable",
		},
		{
			name: "body variables stay loop-local",
			pine: `
//@version=5
strategy("Test")
for val in close
    x = val + 1
`,
			mustContainAll: []string{
				"for _, val := range",
			},
			forbiddenPattern: []string{
				"xSeries",
				"valSeries",
			},
			description: "variables declared inside top-level for-in remain loop-local scalars",
		},
		{
			name: "tuple index wrapped as float64",
			pine: `
//@version=5
strategy("Test")
for [i, val] in close
    x = val + i
`,
			mustContainAll: []string{
				"for i, val := range",
				"float64(i)",
			},
			forbiddenPattern: []string{
				"iSeries",
				"valSeries",
			},
			description: "int index from range wrapped in float64() for arithmetic",
		},
		{
			name: "collection from builtin series",
			pine: `
//@version=5
strategy("Test")
for val in close
    x = val
`,
			mustContainAll: []string{
				"range",
				"val",
			},
			forbiddenPattern: nil,
			description:      "builtin series used as collection expression",
		},
		{
			name: "for-in nested inside traditional for",
			pine: `
//@version=5
strategy("Test")
for i = 0 to 5
    for val in close
        x = val
`,
			mustContainAll: []string{
				"for _, val := range",
				"_to := int(5)",
			},
			forbiddenPattern: nil,
			description:      "for-in inside traditional for generates both loop headers",
		},
		{
			name: "traditional for nested inside for-in",
			pine: `
//@version=5
strategy("Test")
for val in close
    for j = 0 to 3
        x = j
`,
			mustContainAll: []string{
				"for _, val := range",
				"_to := int(3)",
			},
			forbiddenPattern: nil,
			description:      "traditional for inside for-in body generates correct bounds",
		},
		{
			name: "sequential for-in statements",
			pine: `
//@version=5
strategy("Test")
for a in close
    x = a
for b in open
    y = b
`,
			mustContainAll: []string{
				"for _, a := range",
				"for _, b := range",
			},
			forbiddenPattern: nil,
			description:      "multiple sequential for-in loops generate independent range clauses",
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
