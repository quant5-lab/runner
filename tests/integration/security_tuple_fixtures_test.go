package integration

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

const securityTupleFixturePrefix = "test-security-tuple-"

/* Auto-discover and execute all security tuple .pine fixtures */
func TestSecurityTupleFixtures(t *testing.T) {
	t.Parallel()
	fixturesDir := "../fixtures/integration"

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory not found: %v", err)
	}

	exec := util.NewPineExecutor(t)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}
		name := entry.Name()
		if len(name) < len(securityTupleFixturePrefix) || name[:len(securityTupleFixturePrefix)] != securityTupleFixturePrefix {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatalf("read fixture %s: %v", name, err)
			}

			output := exec.ExecuteScript(t, name[:len(name)-5], string(content))
			if output == nil {
				t.Fatal("execution produced no output")
			}
			if len(output.Plots) == 0 {
				t.Fatal("no plots in output")
			}
		})
	}
}

/* OHLCV tuple values must be finite market data */
func TestSecurityTupleFixture_OHLCV(t *testing.T) {
	t.Parallel()
	content := readFixture(t, "test-security-tuple-ohlcv.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "security-tuple-ohlcv", string(content))

	closeVals := exec.ExtractPlotValues(t, output, "Close")
	volumeVals := exec.ExtractPlotValues(t, output, "Volume")

	requireMinBars(t, closeVals, 20, "Close")
	requireMinBars(t, volumeVals, 20, "Volume")

	requireAllFinite(t, closeVals, "Close")
	requireAllFinite(t, volumeVals, "Volume")

	requirePositive(t, closeVals, "Close")
	requirePositive(t, volumeVals, "Volume")
}

/* TA tuple values must be finite after indicator warmup */
func TestSecurityTupleFixture_TA(t *testing.T) {
	t.Parallel()
	content := readFixture(t, "test-security-tuple-ta.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "security-tuple-ta", string(content))

	emaVals := exec.ExtractPlotValues(t, output, "EMA 10")
	smaVals := exec.ExtractPlotValues(t, output, "SMA 20")

	requireMinBars(t, emaVals, 20, "EMA 10")
	requireMinBars(t, smaVals, 20, "SMA 20")

	requireFiniteAfterWarmup(t, emaVals, 10, "EMA 10")
	requireFiniteAfterWarmup(t, smaVals, 20, "SMA 20")
}

/* Mixed tuple: OHLCV elements finite everywhere, TA finite after warmup */
func TestSecurityTupleFixture_Mixed(t *testing.T) {
	t.Parallel()
	content := readFixture(t, "test-security-tuple-mixed.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "security-tuple-mixed", string(content))

	closeVals := exec.ExtractPlotValues(t, output, "Close")
	emaVals := exec.ExtractPlotValues(t, output, "EMA 10")
	volumeVals := exec.ExtractPlotValues(t, output, "Volume")

	requireMinBars(t, closeVals, 20, "Close")
	requireAllFinite(t, closeVals, "Close")
	requirePositive(t, closeVals, "Close")

	requireMinBars(t, emaVals, 20, "EMA 10")
	requireFiniteAfterWarmup(t, emaVals, 10, "EMA 10")

	requireMinBars(t, volumeVals, 20, "Volume")
	requireAllFinite(t, volumeVals, "Volume")
	requirePositive(t, volumeVals, "Volume")
}

/* Multiple independent tuple calls must all produce data */
func TestSecurityTupleFixture_Multi(t *testing.T) {
	t.Parallel()
	content := readFixture(t, "test-security-tuple-multi.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "security-tuple-multi", string(content))

	closeVals := exec.ExtractPlotValues(t, output, "Close")
	volumeVals := exec.ExtractPlotValues(t, output, "Volume")
	emaVals := exec.ExtractPlotValues(t, output, "EMA 10")
	smaVals := exec.ExtractPlotValues(t, output, "SMA 20")

	requireMinBars(t, closeVals, 20, "Close")
	requireMinBars(t, volumeVals, 20, "Volume")
	requireMinBars(t, emaVals, 20, "EMA 10")
	requireMinBars(t, smaVals, 20, "SMA 20")

	requireAllFinite(t, closeVals, "Close")
	requireAllFinite(t, volumeVals, "Volume")
	requireFiniteAfterWarmup(t, emaVals, 10, "EMA 10")
	requireFiniteAfterWarmup(t, smaVals, 20, "SMA 20")
}

/* Tuple and single-var security calls coexisting in same script */
func TestSecurityTupleFixture_Coexist(t *testing.T) {
	t.Parallel()
	content := readFixture(t, "test-security-tuple-coexist.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "security-tuple-coexist", string(content))

	singleClose := exec.ExtractPlotValues(t, output, "Single Close")
	tupleEMA := exec.ExtractPlotValues(t, output, "Tuple EMA 10")
	tupleSMA := exec.ExtractPlotValues(t, output, "Tuple SMA 20")
	singleVol := exec.ExtractPlotValues(t, output, "Single Volume")

	requireMinBars(t, singleClose, 20, "Single Close")
	requireMinBars(t, tupleEMA, 20, "Tuple EMA 10")
	requireMinBars(t, tupleSMA, 20, "Tuple SMA 20")
	requireMinBars(t, singleVol, 20, "Single Volume")

	requireAllFinite(t, singleClose, "Single Close")
	requireAllFinite(t, singleVol, "Single Volume")
	requireFiniteAfterWarmup(t, tupleEMA, 10, "Tuple EMA 10")
	requireFiniteAfterWarmup(t, tupleSMA, 20, "Tuple SMA 20")
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("../fixtures/integration", name))
	if err != nil {
		t.Fatalf("%s not found: %v", name, err)
	}
	return content
}

func requireMinBars(t *testing.T, values []float64, minBars int, label string) {
	t.Helper()
	if len(values) < minBars {
		t.Fatalf("%s: expected >= %d bars, got %d", label, minBars, len(values))
	}
}

func requireAllFinite(t *testing.T, values []float64, label string) {
	t.Helper()
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("%s[%d] = %v, expected finite", label, i, v)
		}
	}
}

func requirePositive(t *testing.T, values []float64, label string) {
	t.Helper()
	for i, v := range values {
		if v <= 0 {
			t.Errorf("%s[%d] = %v, expected positive", label, i, v)
		}
	}
}

func requireFiniteAfterWarmup(t *testing.T, values []float64, warmupBars int, label string) {
	t.Helper()
	if len(values) <= warmupBars {
		t.Fatalf("%s: insufficient bars (%d) for warmup period %d", label, len(values), warmupBars)
	}
	for i := warmupBars; i < len(values); i++ {
		if math.IsNaN(values[i]) || math.IsInf(values[i], 0) {
			t.Errorf("%s[%d] = %v, expected finite after warmup", label, i, values[i])
		}
	}
}
