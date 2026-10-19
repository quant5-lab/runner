package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryExpressionGapScenarios covers the distinct positions where an unknown
// function call can appear in a variable-initializer expression and must degrade
// to a NaN stub while reporting the gap.
var binaryExpressionGapScenarios = []featureGapExpressionScenario{
	{
		name: "unknown function as right-hand binary operand",
		script: `//@version=5
indicator("Binary Expression Gap RHS")
x = close + binExprRhsGapFunc(close)
plot(x, "x")
`,
		wantGapFragment: "binExprRhsGapFunc",
	},
	{
		name: "unknown function as left-hand binary operand",
		script: `//@version=5
indicator("Binary Expression Gap LHS")
x = binExprLhsGapFunc(open) + close
plot(x, "x")
`,
		wantGapFragment: "binExprLhsGapFunc",
	},
	{
		name: "unknown namespaced function in multiplication",
		script: `//@version=5
indicator("Binary Expression Gap Namespaced")
x = close * ns.binExprNsGapFunc(high)
plot(x, "x")
`,
		wantGapFragment: "ns.binExprNsGapFunc",
	},
	{
		name: "unknown function in chained arithmetic (three operands)",
		script: `//@version=5
indicator("Binary Expression Gap Chained")
x = close + binChainGapFunc(open) * high
plot(x, "x")
`,
		wantGapFragment: "binChainGapFunc",
	},
}

// TestFeatureGaps_BinaryExpression_CodegenSucceeds verifies that pine-gen exits 0
// and reports the gap on stderr when an unknown function appears as a binary
// expression operand in a variable initializer. The generated binary must compile
// and run without crashing.
func TestFeatureGaps_BinaryExpression_CodegenSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	for _, sc := range binaryExpressionGapScenarios {
		t.Run(sc.name, func(t *testing.T) {
			testDir := t.TempDir()

			strategyPath := filepath.Join(testDir, "strategy.pine")
			if err := os.WriteFile(strategyPath, []byte(sc.script), 0644); err != nil {
				t.Fatal(err)
			}

			genStdout, genStderr := runPineGen(t, strategyPath, testDir, projectRoot)

			if !strings.Contains(genStderr, "WARNING:") {
				t.Errorf("pine-gen stderr missing WARNING: prefix\nstderr: %s\nstdout: %s", genStderr, genStdout)
			}
			if !strings.Contains(genStderr, sc.wantGapFragment) {
				t.Errorf("pine-gen stderr missing gap %q\nstderr: %s\nstdout: %s",
					sc.wantGapFragment, genStderr, genStdout)
			}

			generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
			exePath := buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)

			dataPath := filepath.Join(testDir, "data.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			resultPath := filepath.Join(testDir, "result.json")
			runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
			if out, err := runCmd.CombinedOutput(); err != nil {
				t.Fatalf("generated binary must not crash with unknown binary-operand stub: %v\noutput: %s", err, out)
			}

			if _, err := os.Stat(resultPath); os.IsNotExist(err) {
				t.Error("result.json not written — binary did not complete successfully")
			}
		})
	}
}

// TestFeatureGaps_BinaryExpression_MultipleGapsAllReported verifies that when
// multiple distinct unknown functions appear in one expression, every gap name
// appears in the warning output (not just the first one).
func TestFeatureGaps_BinaryExpression_MultipleGapsAllReported(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
indicator("Binary Expression Multiple Gaps")
x = binMultiGapA(close) + binMultiGapB(open)
plot(x, "x")
`
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	testDir := t.TempDir()

	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	genStdout, genStderr := runPineGen(t, strategyPath, testDir, projectRoot)

	for _, gap := range []string{"binMultiGapA", "binMultiGapB"} {
		if !strings.Contains(genStderr, gap) {
			t.Errorf("pine-gen stderr missing gap %q\nstderr: %s\nstdout: %s", gap, genStderr, genStdout)
		}
	}

	generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
	exePath := buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)

	dataPath := filepath.Join(testDir, "data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(20, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	if out, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("generated binary must not crash: %v\noutput: %s", err, out)
	}
}

// TestFeatureGaps_BinaryExpression_NoWarningForKnownFunctions is the negative
// case: binary expressions using only OHLCV builtins and implemented math must
// not emit any WARNING on stderr.
func TestFeatureGaps_BinaryExpression_NoWarningForKnownFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
indicator("Binary No Gaps")
x = close + open
y = high - low
z = x * y
plot(z, "z")
`
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	testDir := t.TempDir()

	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	_, genStderr := runPineGen(t, strategyPath, testDir, projectRoot)

	if strings.Contains(genStderr, "WARNING:") {
		t.Errorf("pine-gen must not emit WARNING for binary expressions using only known sources\nstderr: %s", genStderr)
	}
}
