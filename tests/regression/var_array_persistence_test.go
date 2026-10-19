//go:build integration

package regression

import (
	"os"
	"strings"
	"testing"
)

// TestVarArrayPersistence covers all float[] and line[] (drawing-type float[]) var
// declarations: codegen succeeds, binary builds, runs cleanly, and the generated
// carry-forward uses the ArraySeries suffix rather than the plain Series suffix.
func TestVarArrayPersistence(t *testing.T) {
	projectRoot := projectRootFromCwd()

	tests := []struct {
		name        string
		pine        string
		mustContain string // pattern in generated Go source
		mustAvoid   string // pattern that would indicate wrong suffix
	}{
		{
			name: "float array carry-forward uses ArraySeries suffix",
			pine: `//@version=5
strategy("Float Array Persistence")
var float[] levels = array.new_float()
array.push(levels, close)
plot(array.size(levels))
`,
			mustAvoid:   "levelsSeries.Set(levelsSeries.Get(1))",
			mustContain: "levelsArraySeries",
		},
		{
			name: "line array carry-forward uses ArraySeries suffix",
			pine: `//@version=5
strategy("Line Array Persistence")
var line[] lines = array.new_line()
plot(array.size(lines))
`,
			mustAvoid:   "linesSeries.Set(linesSeries.Get(1))",
			mustContain: "linesArraySeries",
		},
		{
			name: "multiple array types in same strategy",
			pine: `//@version=5
strategy("Multi Array")
var float[] prices = array.new_float()
var line[] markers = array.new_line()
array.push(prices, close)
plot(array.size(prices))
`,
			mustContain: "pricesArraySeries",
			mustAvoid:   "pricesSeries.Set(pricesSeries.Get(1))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			result, ok := codegenAndBuild(t, tmpDir, "array_persistence", tt.pine, projectRoot)
			if !ok {
				t.Fatal("codegen+build failed")
			}

			generated, err := os.ReadFile(result.GeneratedPath)
			if err != nil {
				t.Fatalf("read generated: %v", err)
			}
			code := string(generated)

			if tt.mustContain != "" && !strings.Contains(code, tt.mustContain) {
				t.Errorf("generated code missing %q", tt.mustContain)
			}
			if tt.mustAvoid != "" && strings.Contains(code, tt.mustAvoid) {
				t.Errorf("generated code contains wrong pattern %q", tt.mustAvoid)
			}

			dataPath := tmpDir + "/data.json"
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 3600)), 0644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			exitCode, stderr := runBinary(result.BinaryPath, dataPath, tmpDir+"/out.json")
			if containsRawPanic(stderr) {
				t.Fatalf("raw panic:\n%s", stderr)
			}
			if exitCode != 0 && exitCode != 2 {
				t.Fatalf("unexpected exit %d\nstderr: %s", exitCode, stderr)
			}
		})
	}
}
