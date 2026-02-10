package codegen

import (
	"strings"
	"testing"
)

/* TestBreakContinueStatementCodegen validates break/continue emission across all loop types in statement context */
func TestBreakContinueStatementCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "break in for loop",
			pine: `
//@version=5
indicator("Test")
findFirst(len) =>
    result = 0.0
    for i = 0 to len - 1
        result := i
        break
    result
plot(findFirst(10))
`,
			mustContainAll:   []string{"break"},
			forbiddenPattern: nil,
			description:      "break keyword emitted inside traditional for body",
		},
		{
			name: "continue in for loop",
			pine: `
//@version=5
indicator("Test")
skipOdd(len) =>
    sum = 0.0
    for i = 0 to len - 1
        continue
        sum := sum + i
    sum
plot(skipOdd(10))
`,
			mustContainAll:   []string{"continue"},
			forbiddenPattern: nil,
			description:      "continue keyword emitted inside traditional for body",
		},
		{
			name: "break in for-in single form",
			pine: `
//@version=5
indicator("Test")
first(arr) =>
    result = 0.0
    for val in arr
        result := val
        break
    result
plot(first(close))
`,
			mustContainAll:   []string{"for _, val := range", "break"},
			forbiddenPattern: nil,
			description:      "break emitted inside for-in single form body",
		},
		{
			name: "continue in for-in tuple form",
			pine: `
//@version=5
indicator("Test")
skipFirst(arr) =>
    sum = 0.0
    for [i, val] in arr
        if i < 1
            continue
        sum := sum + val
    sum
plot(skipFirst(close))
`,
			mustContainAll:   []string{"for i, val := range", "continue"},
			forbiddenPattern: nil,
			description:      "continue emitted inside for-in tuple form body",
		},
		{
			name: "conditional break inside if block",
			pine: `
//@version=5
indicator("Test")
search(len) =>
    result = -1.0
    for i = 0 to len - 1
        if i > 5
            result := i
            break
    result
plot(search(20))
`,
			mustContainAll:   []string{"break", "> 5"},
			forbiddenPattern: nil,
			description:      "break inside if block within for loop body",
		},
		{
			name: "conditional break in for-in",
			pine: `
//@version=5
indicator("Test")
findBig(arr) =>
    result = 0.0
    for val in arr
        if val > 100
            result := val
            break
    result
plot(findBig(close))
`,
			mustContainAll:   []string{"for _, val := range", "break"},
			forbiddenPattern: nil,
			description:      "break inside if in for-in body",
		},
		{
			name: "both break and continue in same loop",
			pine: `
//@version=5
indicator("Test")
process(len) =>
    sum = 0.0
    for i = 0 to len - 1
        if i < 2
            continue
        if i > 7
            break
        sum := sum + i
    sum
plot(process(20))
`,
			mustContainAll:   []string{"break", "continue"},
			forbiddenPattern: nil,
			description:      "break and continue coexist in same loop body",
		},
		{
			name: "break in inner for-in within outer for",
			pine: `
//@version=5
indicator("Test")
scan(arr, len) =>
    total = 0.0
    for i = 0 to len - 1
        for val in arr
            total := total + val
            break
    total
plot(scan(close, 5))
`,
			mustContainAll:   []string{"_to := int(", "for _, val := range", "break"},
			forbiddenPattern: nil,
			description:      "break in nested inner loop only breaks inner scope",
		},
		{
			name: "continue in inner for within outer for-in",
			pine: `
//@version=5
indicator("Test")
process(arr) =>
    total = 0.0
    for val in arr
        for j = 0 to 3
            if j < 2
                continue
            total := total + val
    total
plot(process(close))
`,
			mustContainAll:   []string{"for _, val := range", "continue"},
			forbiddenPattern: nil,
			description:      "continue in inner traditional for within outer for-in",
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

/* TestBreakContinueExpressionCodegen validates break/continue emission inside IIFE expression loops */
func TestBreakContinueExpressionCodegen(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "break in for-loop IIFE",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    if i > 5
        break
    i
plot(x)
`,
			mustContainAll:   []string{"func() float64", "break", "__result"},
			forbiddenPattern: nil,
			description:      "break emitted inside IIFE-wrapped for expression",
		},
		{
			name: "continue in for-loop IIFE",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 9
    if i < 3
        continue
    i
plot(x)
`,
			mustContainAll:   []string{"func() float64", "continue", "__result"},
			forbiddenPattern: nil,
			description:      "continue emitted inside IIFE-wrapped for expression",
		},
		{
			name: "break in for-in IIFE",
			pine: `
//@version=5
indicator("Test")
x = for val in close
    if val > 100
        break
    val
plot(x)
`,
			mustContainAll:   []string{"func() float64", "break", "for _, val := range"},
			forbiddenPattern: nil,
			description:      "break emitted inside IIFE-wrapped for-in expression",
		},
		{
			name: "continue in for-in IIFE",
			pine: `
//@version=5
indicator("Test")
x = for val in close
    if val < 50
        continue
    val
plot(x)
`,
			mustContainAll:   []string{"func() float64", "continue", "for _, val := range"},
			forbiddenPattern: nil,
			description:      "continue emitted inside IIFE-wrapped for-in expression",
		},
		{
			name: "both keywords in for IIFE",
			pine: `
//@version=5
indicator("Test")
x = for i = 0 to 20
    if i < 3
        continue
    if i > 15
        break
    i
plot(x)
`,
			mustContainAll:   []string{"break", "continue", "__result"},
			forbiddenPattern: nil,
			description:      "break and continue coexist inside IIFE expression loop",
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
