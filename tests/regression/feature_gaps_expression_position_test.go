package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// featureGapExpressionScenario describes a Pine script that uses an unknown
// function in expression position (if condition, binary operand, etc.).
type featureGapExpressionScenario struct {
	name            string
	script          string
	wantGapFragment string // function name expected in pine-gen stderr WARNING
}

var featureGapExpressionScenarios = []featureGapExpressionScenario{
	{
		name: "unknown function as if-condition",
		script: `//@version=5
indicator("Feature Gap If Condition")
result = 0.0
if ifCondGapFunc(close)
    result := 1.0
plot(result, "result")
`,
		wantGapFragment: "ifCondGapFunc",
	},
	{
		name: "unknown function as binary operand in if-condition",
		script: `//@version=5
indicator("Feature Gap Binary Operand")
result = 0.0
if binaryGapFunc(close) > 0.0
    result := 1.0
plot(result, "result")
`,
		wantGapFragment: "binaryGapFunc",
	},
	{
		name: "unknown function as arrow expression return value",
		script: `//@version=5
indicator("Feature Gap Arrow Body")
f(src) => arrowGapFunc(src)
plot(f(close), "f")
`,
		wantGapFragment: "arrowGapFunc",
	},
	{
		name: "unknown function as operand in arrow arithmetic expression",
		script: `//@version=5
indicator("Feature Gap Arrow Arithmetic")
f(src) => arrowArithGapFunc(src) + src
plot(f(close), "f")
`,
		wantGapFragment: "arrowArithGapFunc",
	},
	{
		name: "unknown function as direct plot() argument",
		script: `//@version=5
indicator("Feature Gap Plot Argument")
plot(plotArgGapFunc(close), "x")
`,
		wantGapFragment: "plotArgGapFunc",
	},
}

// TestFeatureGaps_ExpressionPosition_CodegenSucceeds verifies that pine-gen
// exits 0 and reports the gap on stderr when an unknown function appears in
// expression position.  The generated binary must also compile and run cleanly.
func TestFeatureGaps_ExpressionPosition_CodegenSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	for _, sc := range featureGapExpressionScenarios {
		t.Run(sc.name, func(t *testing.T) {
			testDir := t.TempDir()

			strategyPath := filepath.Join(testDir, "strategy.pine")
			if err := os.WriteFile(strategyPath, []byte(sc.script), 0644); err != nil {
				t.Fatal(err)
			}

			genStdout, genStderr := runPineGen(t, strategyPath, testDir, projectRoot)

			if !strings.Contains(genStderr, "WARNING:") {
				t.Errorf("pine-gen stderr missing WARNING: prefix\nstderr: %s", genStderr)
			}
			if !strings.Contains(genStderr, sc.wantGapFragment) {
				t.Errorf("pine-gen stderr missing gap %q\nstderr: %s\nstdout: %s",
					sc.wantGapFragment, genStderr, genStdout)
			}

			generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
			exePath := buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)

			dataPath := filepath.Join(testDir, "data.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			resultPath := filepath.Join(testDir, "result.json")
			runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
			if out, err := runCmd.CombinedOutput(); err != nil {
				t.Fatalf("generated binary must not crash with unknown function stub: %v\noutput: %s", err, out)
			}

			if _, err := os.Stat(resultPath); os.IsNotExist(err) {
				t.Error("result.json not written — binary did not complete successfully")
			}
		})
	}
}

// TestFeatureGaps_ExpressionPosition_NoWarningForKnownFunctions is the negative
// case: a script that uses only fully-implemented functions must not emit any
// WARNING on stderr and must produce a binary that runs cleanly.
func TestFeatureGaps_ExpressionPosition_NoWarningForKnownFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
indicator("No Gaps")
x = ta.sma(close, 14)
y = ta.ema(close, 7)
result = 0.0
if x > y
    result := 1.0
plot(result, "signal")
plot(math.abs(close - open), "range")
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
		t.Errorf("pine-gen must not emit WARNING for fully-implemented script\nstderr: %s", genStderr)
	}
}
