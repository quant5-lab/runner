package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/* bar_index built-in variable integration tests */

type PlotData struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

type Plot struct {
	Data []PlotData `json:"data"`
}

type ChartOutput struct {
	Indicators map[string]Plot `json:"indicators"`
}

func TestBarIndexBasic(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Basic", overlay=false)
barIdx = bar_index
plot(barIdx, "Bar Index")
`

	output := runPineScript(t, "bar-index-basic", pineScript)

	barIndexVals := extractPlotValues(t, output, "Bar Index")

	/* Expect: Sequential integers 0, 1, 2, 3... */
	if len(barIndexVals) < 10 {
		t.Fatal("Expected at least 10 bars")
	}

	if barIndexVals[0] != 0 {
		t.Errorf("bar_index[0] = %f, want 0", barIndexVals[0])
	}

	if barIndexVals[1] != 1 {
		t.Errorf("bar_index[1] = %f, want 1", barIndexVals[1])
	}

	/* Validate sequence integrity */
	for i := 0; i < minInt(len(barIndexVals), 100); i++ {
		if barIndexVals[i] != float64(i) {
			t.Errorf("bar_index[%d] = %f, want %d", i, barIndexVals[i], i)
		}
	}

	t.Logf("✅ bar_index sequence validated: 0 to %d", len(barIndexVals)-1)
}

func TestBarIndexModulo(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Modulo", overlay=false)
mod5 = bar_index % 5
mod20 = bar_index % 20
plot(mod5, "Mod 5")
plot(mod20, "Mod 20")
`

	output := runPineScript(t, "bar-index-modulo", pineScript)

	mod5 := extractPlotValues(t, output, "Mod 5")
	mod20 := extractPlotValues(t, output, "Mod 20")

	/* Validate mod 5 cycles: 0,1,2,3,4,0,1... */
	if len(mod5) < 4 {
		t.Fatal("Not enough data points for mod 5 validation")
	}

	if mod5[0] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 0")
	}

	if len(mod5) > 5 && mod5[5] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 5")
	}

	if len(mod5) > 10 && mod5[10] != 0 {
		t.Error("Mod 5 pattern incorrect at bar 10")
	}

	if mod5[3] != 3 {
		t.Error("Mod 5 pattern incorrect at offset 3")
	}

	if len(mod5) > 8 && mod5[8] != 3 {
		t.Error("Mod 5 pattern incorrect at bar 8")
	}

	/* Validate mod 20 hits 0 at bars 0, 20, 40... */
	if len(mod20) > 0 && mod20[0] != 0 {
		t.Error("Mod 20 pattern incorrect at bar 0")
	}

	if len(mod20) > 40 {
		if mod20[20] != 0 || mod20[40] != 0 {
			t.Error("Mod 20 pattern incorrect at multiples of 20")
		}
	}

	t.Log("✅ bar_index modulo operations validated")
}

func TestBarIndexSecurity(t *testing.T) {
	t.Skip("Security function not implemented - see e2e/fixtures/strategies/test-bar-index-security.pine.skip")

	pineScript := `//@version=5
indicator("bar_index Security", overlay=false)

// CRITICAL: bb9 bug pattern
secBarIndex = security(syminfo.tickerid, "1D", bar_index)
secMod20 = security(syminfo.tickerid, "1D", (bar_index % 20) == 0)

plot(secBarIndex, "Security Bar Index")
plot(secMod20 ? 1 : 0, "Security Mod 20")
`

	output := runPineScript(t, "bar-index-security", pineScript)

	secBarIndex := extractPlotValues(t, output, "Security Bar Index")
	secMod20 := extractPlotValues(t, output, "Security Mod 20")

	/* CRITICAL: security() bar_index must not be NaN */
	for i, val := range secBarIndex {
		if val != val { // NaN check
			t.Errorf("CRITICAL: bar_index in security() is NaN at index %d", i)
		}
	}

	/* CRITICAL: Mod 20 condition must work */
	if len(secMod20) == 0 {
		t.Error("CRITICAL: No values from security() bar_index modulo")
	}

	t.Log("✅ bar_index in security() context validated (bb9 pattern)")
}

func TestBarIndexConditional(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Conditional", overlay=false)
firstBar = bar_index == 0 ? 1 : 0
every10th = (bar_index % 10) == 0 ? 1 : 0
plot(firstBar, "First Bar")
plot(every10th, "Every 10th")
`

	output := runPineScript(t, "bar-index-conditional", pineScript)

	firstBar := extractPlotValues(t, output, "First Bar")
	every10th := extractPlotValues(t, output, "Every 10th")

	/* First bar flag should be 1 only at bar 0 */
	if firstBar[0] != 1 {
		t.Error("First bar flag should be 1 at bar 0")
	}
	if len(firstBar) > 1 && firstBar[1] != 0 {
		t.Error("First bar flag should be 0 after bar 0")
	}

	/* Every 10th bar flag should be 1 at 0, 10, 20... */
	if len(every10th) > 20 {
		if every10th[0] != 1 || every10th[10] != 1 || every10th[20] != 1 {
			t.Error("Every 10th bar flag incorrect")
		}
	}

	t.Log("✅ bar_index conditional logic validated")
}

func TestBarIndexComparisons(t *testing.T) {
	pineScript := `//@version=5
indicator("bar_index Comparisons", overlay=false)
gtTen = bar_index > 10 ? 1 : 0
eqTwenty = bar_index == 20 ? 1 : 0
plot(gtTen, "Greater Than 10")
plot(eqTwenty, "Equals 20")
`

	output := runPineScript(t, "bar-index-comparisons", pineScript)

	gtTen := extractPlotValues(t, output, "Greater Than 10")
	eqTwenty := extractPlotValues(t, output, "Equals 20")

	/* > 10 should be false until bar 11 */
	if len(gtTen) > 11 {
		if gtTen[10] != 0 || gtTen[11] != 1 {
			t.Error("Greater than 10 comparison incorrect")
		}
	}

	/* == 20 should be true only at bar 20 */
	if len(eqTwenty) > 21 {
		if eqTwenty[19] != 0 || eqTwenty[20] != 1 || eqTwenty[21] != 0 {
			t.Error("Equals 20 comparison incorrect")
		}
	}

	t.Log("✅ bar_index comparisons validated")
}

func TestBarIndexHistorical(t *testing.T) {
	t.Skip("Requires bar_index historical access codegen - see e2e/fixtures/strategies/test-bar-index-historical.pine.skip")

	pineScript := `//@version=5
indicator("bar_index Historical", overlay=false)
prevBar = bar_index[1]
barDiff = bar_index - nz(bar_index[1])
plot(prevBar, "Previous Bar")
plot(barDiff, "Bar Diff")
`

	output := runPineScript(t, "bar-index-historical", pineScript)

	prevBar := extractPlotValues(t, output, "Previous Bar")
	barDiff := extractPlotValues(t, output, "Bar Diff")

	/* bar_index[1] at bar N should equal N-1 */
	if len(prevBar) > 5 {
		/* At bar 5, bar_index[1] should be 4 */
		if prevBar[5] != 4 {
			t.Errorf("bar_index[1] at bar 5 = %f, want 4", prevBar[5])
		}
	}

	/* bar_index - bar_index[1] should always be 1 (after first bar) */
	if len(barDiff) > 5 {
		if barDiff[5] != 1 {
			t.Errorf("bar_index diff at bar 5 = %f, want 1", barDiff[5])
		}
	}

	t.Log("✅ bar_index historical access validated")
}

/* Helper functions */

func runPineScript(t *testing.T, testName, pineScript string) ChartOutput {
	t.Helper()

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "test.pine")
	outputBinary := filepath.Join(tmpDir, "test_binary")
	outputJSON := filepath.Join(tmpDir, "output.json")
	dataFile := filepath.Join("testdata", "simple-bars.json")

	err := os.WriteFile(pineFile, []byte(pineScript), 0644)
	if err != nil {
		t.Fatalf("Write Pine file: %v", err)
	}

	originalDir, _ := os.Getwd()
	os.Chdir("../..")
	defer os.Chdir(originalDir)

	buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
		"-input", pineFile, "-output", outputBinary)
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Build failed: %v\nOutput: %s", err, buildOut)
	}

	tempGoFile := parseGeneratedFilePath(t, buildOut)

	compileCmd := exec.Command("go", "build", "-o", outputBinary, tempGoFile)
	compileOut, err := compileCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Compile failed: %v\nOutput: %s", err, compileOut)
	}

	execCmd := exec.Command(outputBinary, "-symbol", "TEST", "-data", dataFile, "-output", outputJSON)
	execOut, err := execCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Execution failed: %v\nOutput: %s", err, execOut)
	}

	jsonBytes, err := os.ReadFile(outputJSON)
	if err != nil {
		t.Fatalf("Read output: %v", err)
	}

	var output ChartOutput
	if err := json.Unmarshal(jsonBytes, &output); err != nil {
		t.Fatalf("Parse JSON: %v", err)
	}

	return output
}

func extractPlotValues(t *testing.T, output ChartOutput, plotTitle string) []float64 {
	t.Helper()

	plot, ok := output.Indicators[plotTitle]
	if !ok {
		t.Fatalf("Plot %q not found", plotTitle)
	}

	values := make([]float64, len(plot.Data))
	for i, d := range plot.Data {
		values[i] = d.Value
	}

	return values
}

func parseGeneratedFilePath(t *testing.T, output []byte) string {
	t.Helper()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Generated:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}

	t.Fatal("Could not find generated Go file path")
	return ""
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
