package codegen

import (
	"strings"
	"testing"
)

/* TestArrowForInRangeGeneration validates for-in range clause generation in arrow function context */
func TestArrowForInRangeGeneration(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "single element generates blank index",
			pine: `
//@version=5
indicator("Test")
iterate(arr) =>
    total = 0.0
    for val in arr
        total := total + val
    total
plot(iterate(close))
`,
			mustContainAll:   []string{"for _, val := range"},
			forbiddenPattern: []string{"for val := range"},
			description:      "single form uses _ for unused index variable",
		},
		{
			name: "tuple form generates named index",
			pine: `
//@version=5
indicator("Test")
weighted(arr) =>
    total = 0.0
    for [i, val] in arr
        total := total + val * i
    total
plot(weighted(close))
`,
			mustContainAll:   []string{"for i, val := range"},
			forbiddenPattern: []string{"for _, val := range"},
			description:      "tuple form uses named index variable",
		},
		{
			name: "parameter as collection",
			pine: `
//@version=5
indicator("Test")
process(data) =>
    sum = 0.0
    for elem in data
        sum := sum + elem
    sum
plot(process(close))
`,
			mustContainAll:   []string{"for _, elem := range data"},
			forbiddenPattern: nil,
			description:      "function parameter resolves as scalar in range expression",
		},
		{
			name: "tuple index wrapping in arithmetic",
			pine: `
//@version=5
indicator("Test")
weightedSum(arr) =>
    result = 0.0
    for [idx, val] in arr
        result := result + val * idx
    result
plot(weightedSum(close))
`,
			mustContainAll:   []string{"for idx, val := range", "val * idx"},
			forbiddenPattern: []string{"idxSeries"},
			description:      "tuple index variable used directly in arithmetic expression",
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

/* TestArrowForInVariableTracking validates Series backing for loop-modified variables in arrow context */
func TestArrowForInVariableTracking(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "reassigned accumulator uses series",
			pine: `
//@version=5
indicator("Test")
accumulate(arr) =>
    count = 0.0
    for val in arr
        count := count + 1
    count
plot(accumulate(close))
`,
			mustContainAll:   []string{"countSeries.Set("},
			forbiddenPattern: nil,
			description:      "loop-modified variable stored via Series.Set()",
		},
		{
			name: "all locals get series backing",
			pine: `
//@version=5
indicator("Test")
calc(arr) =>
    multiplier = 2.0
    sum = 0.0
    for val in arr
        sum := sum + val
    sum * multiplier
plot(calc(close))
`,
			mustContainAll:   []string{"multiplierSeries", "sumSeries", "* multiplier)"},
			forbiddenPattern: nil,
			description:      "all arrow-context locals get series backing for bar persistence",
		},
		{
			name: "unmodified local returns scalar access",
			pine: `
//@version=5
indicator("Test")
transform(arr) =>
    factor = 3.0
    total = 0.0
    for val in arr
        total := total + val
    total * factor
plot(transform(close))
`,
			mustContainAll:   []string{"totalSeries", "factorSeries", "* factor)"},
			forbiddenPattern: nil,
			description:      "all locals including unmodified get series backing with scalar variable access in expressions",
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

/* TestArrowForInNestedLoops validates nesting combinations of for-in with traditional for in arrow context */
func TestArrowForInNestedLoops(t *testing.T) {
	tests := []struct {
		name             string
		pine             string
		mustContainAll   []string
		forbiddenPattern []string
		description      string
	}{
		{
			name: "for-in inside traditional for",
			pine: `
//@version=5
indicator("Test")
hybrid(arr, len) =>
    total = 0.0
    for i = 0 to len - 1
        for val in arr
            total := total + val
    total
plot(hybrid(close, 5))
`,
			mustContainAll:   []string{"_to := int(", "for _, val := range"},
			forbiddenPattern: nil,
			description:      "traditional for bounds and for-in range coexist in arrow context",
		},
		{
			name: "traditional for inside for-in",
			pine: `
//@version=5
indicator("Test")
scan(arr) =>
    total = 0.0
    for val in arr
        for j = 0 to 3
            total := total + val
    total
plot(scan(close))
`,
			mustContainAll:   []string{"for _, val := range", "_to := int(3)"},
			forbiddenPattern: nil,
			description:      "for-in body can contain traditional for loops in arrow context",
		},
		{
			name: "for-in inside for-in",
			pine: `
//@version=5
indicator("Test")
nested(outer, inner) =>
    total = 0.0
    for a in outer
        for b in inner
            total := total + a + b
    total
plot(nested(close, open))
`,
			mustContainAll:   []string{"for _, a := range", "for _, b := range"},
			forbiddenPattern: nil,
			description:      "nested for-in loops generate independent range clauses",
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
