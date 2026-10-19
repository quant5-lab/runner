package regression

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// unknownFunctionCodegenScenario describes a Pine script whose unknown function
// hits a specific codegen emission path and asserts the generated source contains
// a featuregap.Record call at that site.
type unknownFunctionCodegenScenario struct {
	name          string
	script        string
	wantSourceTag string
}

// unknownFunctionCodegenScenarios covers the codegen paths where an unknown
// function call can appear in a variable-assignment context. Each scenario
// hits a distinct site inside generator.generateVariableFromCall.
//
// Note: expression-position paths (plot argument, arrow body, binary operand)
// are covered at runtime level by feature_gaps_expression_position_test.go and
// feature_gaps_binary_expression_test.go. Only paths with source-level contract
// differences from those scenarios are listed here.
var unknownFunctionCodegenScenarios = []unknownFunctionCodegenScenario{
	{
		name: "variable-initializer unknown function emits featuregap.Record",
		script: `//@version=5
indicator("Variable Init Unknown")
x = mystery_var_init_func(close)
plot(x, "x")
`,
		wantSourceTag: `featuregap.Record("mystery_var_init_func"`,
	},
	{
		name: "namespaced variable-initializer unknown function emits featuregap.Record",
		script: `//@version=5
indicator("Variable Init Namespaced Unknown")
x = ns.mystery_ns_func(close)
plot(x, "x")
`,
		wantSourceTag: `featuregap.Record("ns.mystery_ns_func"`,
	},
}

// TestCodegen_UnknownFunctionVariableInit_EmitsFeaturegapRecord verifies that
// pine-gen generates a featuregap.Record call (not a bare math.NaN literal) when
// an unknown function appears in a variable-assignment initializer position.
// The generated binary must build and run without error.
func TestCodegen_UnknownFunctionVariableInit_EmitsFeaturegapRecord(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := projectRootFromCwd()

	for _, sc := range unknownFunctionCodegenScenarios {
		t.Run(sc.name, func(t *testing.T) {
			testDir := t.TempDir()
			strategyPath := filepath.Join(testDir, "strategy.pine")
			if err := os.WriteFile(strategyPath, []byte(sc.script), 0644); err != nil {
				t.Fatal(err)
			}

			genStdout, _ := runPineGen(t, strategyPath, testDir, projectRoot)
			generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
			src, err := os.ReadFile(generatedFile)
			if err != nil {
				t.Fatalf("read generated source: %v", err)
			}

			if !strings.Contains(string(src), sc.wantSourceTag) {
				t.Errorf("generated source missing %q\nexcerpt: %s",
					sc.wantSourceTag, truncate(string(src), 4000))
			}

			if strings.Contains(string(src), "math.NaN() // TODO:") {
				t.Errorf("generated source still contains orphan bare math.NaN() TODO pattern")
			}

			buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)
		})
	}
}
