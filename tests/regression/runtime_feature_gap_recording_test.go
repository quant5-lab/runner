package regression

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRuntimeFeatureGap_RecordEmitsAtUnknownSites runs the blocker fixture
// test-runtime-feature-gap.pine through pine-gen + go build and asserts:
//  1. the binary compiles (no bare math.NaN orphans in the generated source
//     at the migrated codegen sites — we grep for featuregap.Record presence);
//  2. the binary executes without crashing (crashless contract preserved);
//  3. stderr surfaces FEATURE-GAP messages for at least one of the unknown calls
//     (i.e. featuregap.Record actually fired at runtime, not just compiled away).
func TestRuntimeFeatureGap_RecordEmitsAtUnknownSites(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := projectRootFromCwd()
	fixturePath := filepath.Join(projectRoot, "tests", "fixtures", "blockers", "test-runtime-feature-gap.pine")
	pineBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, pineBytes, 0644); err != nil {
		t.Fatal(err)
	}

	genStdout, _ := runPineGen(t, strategyPath, testDir, projectRoot)

	generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
	generatedSrc, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("read generated source: %v", err)
	}

	// Assert featuregap.Record appears in the generated source for at least
	// one of the unknown function sites.
	if !strings.Contains(string(generatedSrc), "featuregap.Record") {
		t.Errorf("generated source does NOT contain featuregap.Record — orphan silent-NaN site suspected\nsource excerpt: %s",
			truncate(string(generatedSrc), 4000))
	}

	exePath := buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)

	dataPath := filepath.Join(testDir, "data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated binary crashed on feature-gap fixture (crashless contract violated): %v\noutput: %s",
			err, runOutput)
	}

	// Runtime FEATURE-GAP log line emitted via featuregap.Record.
	if !strings.Contains(string(runOutput), "FEATURE-GAP runtime:") {
		t.Errorf("runtime output missing FEATURE-GAP record — featuregap.Record did not fire at runtime\noutput: %s",
			runOutput)
	}

	if _, err := os.Stat(resultPath); os.IsNotExist(err) {
		t.Error("result.json not written — binary did not complete successfully")
	}
}

// TestCalendarFuncs_MsCanonicalization runs the calendar fixture and asserts
// year(time)/month(time)/etc. return the correct values (years near 2021/2022,
// NOT year 55000+ that would happen if calendar pkg still treated ms as seconds).
func TestCalendarFuncs_MsCanonicalization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot := projectRootFromCwd()
	fixturePath := filepath.Join(projectRoot, "tests", "fixtures", "blockers", "test-calendar-funcs.pine")
	pineBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	testDir := t.TempDir()
	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, pineBytes, 0644); err != nil {
		t.Fatal(err)
	}

	genStdout, _ := runPineGen(t, strategyPath, testDir, projectRoot)
	generatedFile := resolveGeneratedFilePath([]byte(genStdout), filepath.Join(testDir, "strategy.go"))
	exePath := buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)

	// Test data starts at startTime=1640000000 sec (2021-12-20 UTC) with daily steps.
	dataPath := filepath.Join(testDir, "data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(30, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	if out, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("generated calendar binary crashed: %v\noutput: %s", err, out)
	}

	resultBytes, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	// All three call forms of Pine's time() must yield ms-unit timestamps.
	// Fixture epoch: 1640000000 sec = 2021-12-20 UTC; 30 daily bars stays inside 2022.
	// If any path stores seconds instead of ms, year() returns ~55000 (1970 + 55000/365 ≈ year 2120+).
	calendarPlots := []struct {
		plotName string
		desc     string
	}{
		{"year", "year(time) — built-in time variable"},
		{"year_via_t", "year(t) where t = time() — no-arg variable-init path"},
		{"year_via_t1d", "year(t) where t = time(\"D\") — single-arg variable-init path"},
	}
	for _, cp := range calendarPlots {
		assertYearPlotInRange(t, result, resultBytes, cp.plotName, cp.desc)
	}
}

func assertYearPlotInRange(t *testing.T, result map[string]any, raw []byte, plotName, desc string) {
	t.Helper()
	values := extractPlotValues(t, result, plotName)
	if len(values) == 0 {
		t.Errorf("plot %q (%s) missing or empty in result\nraw: %s",
			plotName, desc, truncate(string(raw), 2000))
		return
	}
	for _, y := range values {
		if y < 2020 || y > 2030 {
			t.Errorf("%s (%s): year = %v, want value in [2020, 2030]", plotName, desc, y)
			break
		}
	}
}

// extractPlotValues walks a chartdata result structure looking for the named plot
// and returns its float values (skipping NaN/null). Tolerant of multiple known
// schema variants (panes[].plots[], plots[], data[].plots[]).
func extractPlotValues(t *testing.T, result map[string]any, name string) []float64 {
	t.Helper()
	var out []float64

	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if title, _ := x["title"].(string); title == name {
				if data, ok := x["data"].([]any); ok {
					for _, d := range data {
						switch dv := d.(type) {
						case float64:
							out = append(out, dv)
						case map[string]any:
							if val, ok := dv["value"].(float64); ok {
								out = append(out, val)
							}
						}
					}
				}
			}
			for _, child := range x {
				walk(child)
			}
		case []any:
			for _, item := range x {
				walk(item)
			}
		}
	}
	walk(result)
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(truncated)"
}
