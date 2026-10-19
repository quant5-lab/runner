package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// taSourceGapScenarios exercises the TA source argument unknown-call fallback path,
// distinct from expression-position unknown calls covered in feature_gaps_expression_position_test.go.
var taSourceGapScenarios = []featureGapExpressionScenario{
	{
		name: "unknown function as ta.sma source",
		script: `//@version=5
indicator("TA Source Gap SMA")
x = ta.sma(taSourceGapFunc(close), 14)
plot(x, "x")
`,
		wantGapFragment: "taSourceGapFunc",
	},
	{
		name: "unknown function as ta.ema source",
		script: `//@version=5
indicator("TA Source Gap EMA")
x = ta.ema(taEmaSourceGap(close), 7)
plot(x, "x")
`,
		wantGapFragment: "taEmaSourceGap",
	},
	{
		name: "unknown function inside binary expression as ta.rsi source",
		script: `//@version=5
indicator("TA Source Gap RSI Binary")
x = ta.rsi(taBinarySourceGap(close) + close, 14)
plot(x, "x")
`,
		wantGapFragment: "taBinarySourceGap",
	},
	{
		name: "namespaced unknown function as ta.sma source",
		script: `//@version=5
indicator("TA Source Gap Namespaced")
x = ta.sma(ns.taNamespacedGap(close), 14)
plot(x, "x")
`,
		wantGapFragment: "ns.taNamespacedGap",
	},
}

// TestFeatureGaps_TASourceArgument_CodegenSucceeds verifies that pine-gen exits 0
// and reports the gap on stderr when an unknown function is nested as the source
// argument to any TA indicator. The generated binary must compile and run without
// crashing.
func TestFeatureGaps_TASourceArgument_CodegenSucceeds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	for _, sc := range taSourceGapScenarios {
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
				t.Fatalf("generated binary must not crash with unknown TA source stub: %v\noutput: %s", err, out)
			}

			if _, err := os.Stat(resultPath); os.IsNotExist(err) {
				t.Error("result.json not written — binary did not complete successfully")
			}
		})
	}
}

// TestFeatureGaps_TASourceArgument_KnownSourceNoWarning is the negative case:
// a script using only fully-implemented TA sources must not emit any WARNING.
func TestFeatureGaps_TASourceArgument_KnownSourceNoWarning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
indicator("TA Known Source No Gaps")
x = ta.sma(close, 14)
y = ta.ema(ta.sma(close, 7), 3)
z = ta.rsi(close - open, 14)
plot(x + y + z, "combined")
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
		t.Errorf("pine-gen must not emit WARNING for fully-implemented TA sources\nstderr: %s", genStderr)
	}
}
