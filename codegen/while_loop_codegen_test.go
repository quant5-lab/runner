package codegen

import (
	"strings"
	"testing"
)

func TestWhileLoopCodegenStructure(t *testing.T) {
	tests := []struct {
		name           string
		pine           string
		mustContainAll []string
		forbidden      []string
	}{
		{
			name: "basic while generates for-loop with guard",
			pine: `
//@version=5
indicator("Test")
sum = 0.0
i = 1
while i <= 10
    sum := sum + i
    i := i + 1
plot(sum)
`,
			mustContainAll: []string{
				"__whileGuard := 0",
				"__whileGuard++",
				"100000",
				"for ",
			},
		},
		{
			name: "while scope isolation wraps in block",
			pine: `
//@version=5
indicator("Test")
x = 0.0
while x < 5
    x := x + 1
plot(x)
`,
			mustContainAll: []string{
				"__whileGuard := 0",
				"for ",
				"break",
			},
		},
		{
			name: "while with break preserves break keyword",
			pine: `
//@version=5
indicator("Test")
sum = 0.0
i = 1
while i <= 100
    if i > 10
        break
    sum := sum + i
    i := i + 1
plot(sum)
`,
			mustContainAll: []string{
				"break",
				"__whileGuard := 0",
				"for ",
			},
		},
		{
			name: "while with continue preserves continue keyword",
			pine: `
//@version=5
indicator("Test")
sum = 0.0
i = 0
while i < 10
    i := i + 1
    if i == 5
        continue
    sum := sum + i
plot(sum)
`,
			mustContainAll: []string{
				"continue",
				"__whileGuard := 0",
				"for ",
			},
		},
		{
			name: "guard absent from for-loop",
			pine: `
//@version=5
indicator("Test")
sum = 0.0
for i = 1 to 10
    sum := sum + i
plot(sum)
`,
			forbidden: []string{
				"__whileGuard",
			},
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
					t.Errorf("Missing required pattern: %q\nGenerated code:\n%s", pattern, goCode)
				}
			}

			for _, pattern := range tt.forbidden {
				if strings.Contains(goCode, pattern) {
					t.Errorf("Found forbidden pattern: %q\nGenerated code:\n%s", pattern, goCode)
				}
			}
		})
	}
}
