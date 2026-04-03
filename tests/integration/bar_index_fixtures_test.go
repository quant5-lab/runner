//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

const barIndexFixturesDir = "../../e2e/fixtures/strategies"

func TestBarIndexFixtures(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(barIndexFixturesDir)
	if err != nil {
		t.Fatalf("fixtures directory: %v", err)
	}

	exec := util.NewPineExecutor(t)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pine" {
			continue
		}
		name := entry.Name()
		if len(name) < 15 || name[:15] != "test-bar-index-" {
			continue
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, name))
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			output := exec.ExecuteScript(t, name[:len(name)-5], string(content))
			if output == nil {
				t.Error("execution produced no output")
			}
		})
	}
}

func TestBarIndexFixture_Basic(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, "test-bar-index-basic.pine"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-bar-index-basic", string(content))

	doubled := exec.ExtractPlotValues(t, output, "Doubled")
	incremented := exec.ExtractPlotValues(t, output, "Plus 10")
	asFloat := exec.ExtractPlotValues(t, output, "As Float")

	if len(doubled) < 10 {
		t.Fatal("expected at least 10 bars")
	}

	for i := 0; i < minInt(len(doubled), 50); i++ {
		if doubled[i] != float64(i*2) {
			t.Errorf("bar %d: doubled = %f, want %f", i, doubled[i], float64(i*2))
		}
	}

	for i := 0; i < minInt(len(incremented), 50); i++ {
		if incremented[i] != float64(i+10) {
			t.Errorf("bar %d: incremented = %f, want %f", i, incremented[i], float64(i+10))
		}
	}

	for i := 0; i < minInt(len(asFloat), 50); i++ {
		if asFloat[i] != float64(i) {
			t.Errorf("bar %d: asFloat = %f, want %f", i, asFloat[i], float64(i))
		}
	}
}

func TestBarIndexFixture_Comparisons(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, "test-bar-index-comparisons.pine"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-bar-index-comparisons", string(content))

	ltTwenty := exec.ExtractPlotValues(t, output, "Less Than 20")
	lteFive := exec.ExtractPlotValues(t, output, "Less or Equal 5")
	eqTwenty := exec.ExtractPlotValues(t, output, "Equals 20")
	neqZero := exec.ExtractPlotValues(t, output, "Not Equal 0")
	inRange := exec.ExtractPlotValues(t, output, "In Range")
	outOfRange := exec.ExtractPlotValues(t, output, "Out of Range")

	n := minInt(len(inRange), 60)
	if n < 25 {
		t.Fatal("expected at least 25 bars")
	}

	for i := 0; i < n; i++ {
		if ltTwenty[i] != boolToFloat(i < 20) {
			t.Errorf("bar %d: ltTwenty = %f, want %f", i, ltTwenty[i], boolToFloat(i < 20))
		}
		if lteFive[i] != boolToFloat(i <= 5) {
			t.Errorf("bar %d: lteFive = %f, want %f", i, lteFive[i], boolToFloat(i <= 5))
		}
		if eqTwenty[i] != boolToFloat(i == 20) {
			t.Errorf("bar %d: eqTwenty = %f, want %f", i, eqTwenty[i], boolToFloat(i == 20))
		}
		if neqZero[i] != boolToFloat(i != 0) {
			t.Errorf("bar %d: neqZero = %f, want %f", i, neqZero[i], boolToFloat(i != 0))
		}
		if inRange[i] != boolToFloat(i >= 10 && i <= 30) {
			t.Errorf("bar %d: inRange = %f, want %f", i, inRange[i], boolToFloat(i >= 10 && i <= 30))
		}
		if outOfRange[i] != boolToFloat(i < 10 || i > 50) {
			t.Errorf("bar %d: outOfRange = %f, want %f", i, outOfRange[i], boolToFloat(i < 10 || i > 50))
		}
	}
}

func TestBarIndexFixture_Conditional(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, "test-bar-index-conditional.pine"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-bar-index-conditional", string(content))

	afterWarmup := exec.ExtractPlotValues(t, output, "After Warmup")
	every10Bars := exec.ExtractPlotValues(t, output, "Every 10 Bars")
	every25Bars := exec.ExtractPlotValues(t, output, "Every 25 Bars")
	inRange := exec.ExtractPlotValues(t, output, "In Range 10-20")

	n := minInt(len(afterWarmup), 60)
	if n < 30 {
		t.Fatal("expected at least 30 bars")
	}

	for i := 0; i < n; i++ {
		if afterWarmup[i] != boolToFloat(i > 20) {
			t.Errorf("bar %d: afterWarmup = %f, want %f", i, afterWarmup[i], boolToFloat(i > 20))
		}
		if every10Bars[i] != boolToFloat(i%10 == 0) {
			t.Errorf("bar %d: every10Bars = %f, want %f", i, every10Bars[i], boolToFloat(i%10 == 0))
		}
		if every25Bars[i] != boolToFloat(i%25 == 0) {
			t.Errorf("bar %d: every25Bars = %f, want %f", i, every25Bars[i], boolToFloat(i%25 == 0))
		}
		if inRange[i] != boolToFloat(i >= 10 && i <= 20) {
			t.Errorf("bar %d: inRange = %f, want %f", i, inRange[i], boolToFloat(i >= 10 && i <= 20))
		}
	}
}

func TestBarIndexFixture_Historical(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, "test-bar-index-historical.pine"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-bar-index-historical-fixture", string(content))

	currentBarIndex := exec.ExtractPlotValues(t, output, "Current Bar Index [0]")
	incrementedBy1 := exec.ExtractPlotValues(t, output, "Incremented By 1")
	history3 := exec.ExtractPlotValues(t, output, "History [3]")

	if len(currentBarIndex) < 10 {
		t.Fatal("expected at least 10 bars")
	}

	for i := 0; i < minInt(len(currentBarIndex), 30); i++ {
		if currentBarIndex[i] != float64(i) {
			t.Errorf("bar %d: bar_index[0] = %f, want %f", i, currentBarIndex[i], float64(i))
		}
	}

	for i := 1; i < minInt(len(incrementedBy1), 30); i++ {
		if incrementedBy1[i] != 1 {
			t.Errorf("bar %d: incrementedBy1 = %f, want 1", i, incrementedBy1[i])
		}
	}

	for i := 3; i < minInt(len(history3), 30); i++ {
		if history3[i] != float64(i-3) {
			t.Errorf("bar %d: bar_index[3] = %f, want %f", i, history3[i], float64(i-3))
		}
	}
}

func TestBarIndexFixture_Modulo(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join(barIndexFixturesDir, "test-bar-index-modulo.pine"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-bar-index-modulo-fixture", string(content))

	mod10 := exec.ExtractPlotValues(t, output, "Mod 10")
	mod20 := exec.ExtractPlotValues(t, output, "Mod 20")
	every20th := exec.ExtractPlotValues(t, output, "Every 20th Bar")

	if len(mod10) < 30 {
		t.Fatal("expected at least 30 bars")
	}

	for i := 0; i < minInt(len(mod10), 60); i++ {
		if mod10[i] != float64(i%10) {
			t.Errorf("bar %d: mod10 = %f, want %f", i, mod10[i], float64(i%10))
		}
	}

	for i := 0; i < minInt(len(mod20), 60); i++ {
		if mod20[i] != float64(i%20) {
			t.Errorf("bar %d: mod20 = %f, want %f", i, mod20[i], float64(i%20))
		}
	}

	for i := 0; i < minInt(len(every20th), 60); i++ {
		if every20th[i] != boolToFloat(i%20 == 0) {
			t.Errorf("bar %d: every20th = %f, want %f", i, every20th[i], boolToFloat(i%20 == 0))
		}
	}
}
