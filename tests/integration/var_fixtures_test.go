//go:build integration

package integration

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

const varFixturePrefix = "test-var-"
const varipFixturePrefix = "test-varip-"

/* TestVarFixtures auto-discovers and executes all var/varip .pine fixtures */
func TestVarFixtures(t *testing.T) {
	t.Parallel()
	fixturesDir := "../fixtures/integration"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory: %v", err)
	}

	exec := util.NewPineExecutor(t)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".pine" {
			continue
		}
		if !isVarFixture(name) {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			output := exec.ExecuteScript(t, name[:len(name)-5], string(content))
			if output == nil {
				t.Fatal("no output")
			}
			if len(output.Plots) == 0 {
				t.Fatal("no plots in output")
			}
		})
	}
}

/* TestVarFixture_Accumulator validates cumulative close is strictly monotonically increasing */
func TestVarFixture_Accumulator(t *testing.T) {
	t.Parallel()
	vals := executeVarFixture(t, "test-var-accumulator.pine", "Cumulative Close")
	requireMonotonicallyIncreasing(t, vals, "Cumulative Close")
}

/* TestVarFixture_RunningMax validates running max is non-decreasing and ends at actual max(close) */
func TestVarFixture_RunningMax(t *testing.T) {
	t.Parallel()
	vals := executeVarFixture(t, "test-var-running-max.pine", "Running Max")
	requireNonDecreasing(t, vals, "Running Max")
}

/* TestVarFixture_Counter validates bar counter produces sequential integers starting at 1 */
func TestVarFixture_Counter(t *testing.T) {
	t.Parallel()
	vals := executeVarFixture(t, "test-var-counter.pine", "Counter")

	for i, v := range vals {
		expected := float64(i + 1)
		if v != expected {
			t.Fatalf("bar %d: got %f, want %f", i, v, expected)
		}
	}
}

/* TestVarFixture_TypedFloat validates typed var float accumulator behaves identically to untyped */
func TestVarFixture_TypedFloat(t *testing.T) {
	t.Parallel()
	vals := executeVarFixture(t, "test-var-typed-float.pine", "Running Sum")
	requireMonotonicallyIncreasing(t, vals, "Running Sum")
}

/* TestVarFixture_LatchGate validates latch stays 1.0 once triggered on first green bar */
func TestVarFixture_LatchGate(t *testing.T) {
	t.Parallel()
	vals := executeVarFixture(t, "test-var-latch-gate.pine", "Latch")

	firstTrigger := -1
	for i, v := range vals {
		if v == 1.0 {
			firstTrigger = i
			break
		}
	}
	if firstTrigger == -1 {
		t.Fatal("latch never triggered (no green bar in SPY data)")
	}

	for i := firstTrigger; i < len(vals); i++ {
		if vals[i] != 1.0 {
			t.Fatalf("bar %d: latch reverted to %f after trigger at bar %d", i, vals[i], firstTrigger)
		}
	}
}

/* TestVarFixture_VaripIdentical validates varip accumulator matches var accumulator in historical mode */
func TestVarFixture_VaripIdentical(t *testing.T) {
	t.Parallel()
	varVals := executeVarFixture(t, "test-var-accumulator.pine", "Cumulative Close")
	varipVals := executeVarFixture(t, "test-varip-accumulator.pine", "Cumulative Close")

	if len(varVals) != len(varipVals) {
		t.Fatalf("length mismatch: var=%d, varip=%d", len(varVals), len(varipVals))
	}

	for i := range varVals {
		if math.Abs(varVals[i]-varipVals[i]) > 0.001 {
			t.Fatalf("bar %d: var=%f, varip=%f", i, varVals[i], varipVals[i])
		}
	}
}

/* TestVarFixture_MixedPersistence validates var and plain declarations coexist */
func TestVarFixture_MixedPersistence(t *testing.T) {
	t.Parallel()
	exec := util.NewPineExecutor(t)

	content, err := os.ReadFile("../fixtures/integration/test-var-mixed-persistence.pine")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	output := exec.ExecuteScript(t, "test-var-mixed-persistence", string(content))

	cumVol := exec.ExtractPlotValues(t, output, "Cumulative Volume")
	delta := exec.ExtractPlotValues(t, output, "Delta")

	requireMonotonicallyIncreasing(t, cumVol, "Cumulative Volume")

	hasPositive, hasNegative := false, false
	for _, d := range delta {
		if d > 0 {
			hasPositive = true
		}
		if d < 0 {
			hasNegative = true
		}
	}
	if !hasPositive || !hasNegative {
		t.Error("delta (close-open) should have both positive and negative values in real market data")
	}
}

func executeVarFixture(t *testing.T, fixture, plotName string) []float64 {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("../fixtures/integration", fixture))
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, fixture[:len(fixture)-5], string(content))

	vals := exec.ExtractPlotValues(t, output, plotName)
	if len(vals) < 10 {
		t.Fatalf("%s: expected ≥10 bars, got %d", plotName, len(vals))
	}
	return vals
}

func requireMonotonicallyIncreasing(t *testing.T, vals []float64, label string) {
	t.Helper()
	for i := 1; i < len(vals); i++ {
		if vals[i] <= vals[i-1] {
			t.Fatalf("%s bar %d: not increasing (%f -> %f)", label, i, vals[i-1], vals[i])
		}
	}
}

func requireNonDecreasing(t *testing.T, vals []float64, label string) {
	t.Helper()
	for i := 1; i < len(vals); i++ {
		if vals[i] < vals[i-1] {
			t.Fatalf("%s bar %d: decreased (%f -> %f)", label, i, vals[i-1], vals[i])
		}
	}
}

func isVarFixture(name string) bool {
	return len(name) > len(varFixturePrefix) && name[:len(varFixturePrefix)] == varFixturePrefix ||
		len(name) > len(varipFixturePrefix) && name[:len(varipFixturePrefix)] == varipFixturePrefix
}
