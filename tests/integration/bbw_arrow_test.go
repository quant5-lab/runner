//go:build integration

package integration

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestBBWArrowMultLiteral(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-bbw-arrow-mult-literal.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bbw-arrow-mult-literal", script)

	width := exec.ExtractPlotValues(t, output, "BBW 2.5x")

	if len(width) < 20 {
		t.Fatal("Expected at least 20 bars for warmup")
	}

	for i := 0; i < 19; i++ {
		if !math.IsNaN(width[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, width[i])
		}
	}

	if math.IsNaN(width[19]) {
		t.Error("bar[19] should not be NaN after warmup")
	}

	for i := 20; i < minIntBBW(len(width), 50); i++ {
		if math.IsNaN(width[i]) {
			t.Errorf("bar[%d] is NaN after warmup", i)
		}
	}

	t.Logf("✅ BBW arrow mult literal validated: %d bars", len(width))
}

func TestBBWArrowMultParameter(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-bbw-arrow-mult-parameter.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bbw-arrow-mult-parameter", script)

	narrow := exec.ExtractPlotValues(t, output, "Narrow 1.0x")
	wide := exec.ExtractPlotValues(t, output, "Wide 3.0x")

	if len(narrow) != len(wide) {
		t.Fatal("Narrow and wide must have same length")
	}

	if len(narrow) < 20 {
		t.Fatal("Expected at least 20 bars for warmup")
	}

	for i := 20; i < minIntBBW(len(narrow), 50); i++ {
		if math.IsNaN(narrow[i]) || math.IsNaN(wide[i]) {
			t.Errorf("bar[%d] has NaN after warmup", i)
			continue
		}

		if narrow[i] >= wide[i] {
			t.Errorf("bar[%d]: narrow=%f should be < wide=%f (mult 1.0 < 3.0)", i, narrow[i], wide[i])
		}

		ratio := wide[i] / narrow[i]
		if ratio < 2.5 || ratio > 3.5 {
			t.Errorf("bar[%d]: ratio wide/narrow = %f, expected ~3.0", i, ratio)
		}
	}

	t.Logf("✅ BBW arrow mult parameter validated: narrow < wide, ratio ~3.0")
}

func TestBBWArrowBareAlias(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-bbw-arrow-bare-alias.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bbw-arrow-bare-alias", script)

	width := exec.ExtractPlotValues(t, output, "BBW Default Source")

	if len(width) < 14 {
		t.Fatal("Expected at least 14 bars for warmup")
	}

	for i := 0; i < 13; i++ {
		if !math.IsNaN(width[i]) {
			t.Errorf("bar[%d] = %f, want NaN during warmup", i, width[i])
		}
	}

	if math.IsNaN(width[13]) {
		t.Error("bar[13] should not be NaN after warmup")
	}

	t.Logf("✅ BBW bare alias validated: %d bars", len(width))
}

func TestBBWArrowTupleReturn(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-bbw-arrow-tuple-return.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bbw-arrow-tuple-return", script)

	narrow := exec.ExtractPlotValues(t, output, "Narrow 1.5x")
	standard := exec.ExtractPlotValues(t, output, "Standard 2.0x")
	wide := exec.ExtractPlotValues(t, output, "Wide 2.5x")

	if len(narrow) != len(standard) || len(standard) != len(wide) {
		t.Fatal("All tuple elements must have same length")
	}

	if len(narrow) < 20 {
		t.Fatal("Expected at least 20 bars")
	}

	for i := 20; i < minIntBBW(len(narrow), 50); i++ {
		if math.IsNaN(narrow[i]) || math.IsNaN(standard[i]) || math.IsNaN(wide[i]) {
			continue
		}

		if !(narrow[i] < standard[i] && standard[i] < wide[i]) {
			t.Errorf("bar[%d]: narrow=%f, standard=%f, wide=%f - expected narrow < standard < wide",
				i, narrow[i], standard[i], wide[i])
		}
	}

	t.Logf("✅ BBW tuple return validated: narrow < standard < wide")
}

func TestBBWArrowMixed(t *testing.T) {
	t.Parallel()

	script := loadFixture(t, "integration/test-bbw-arrow-mixed.pine")
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "bbw-arrow-mixed", script)

	direct := exec.ExtractPlotValues(t, output, "Direct")
	arrow := exec.ExtractPlotValues(t, output, "Arrow")
	diff := exec.ExtractPlotValues(t, output, "Difference")

	if len(direct) != len(arrow) || len(arrow) != len(diff) {
		t.Fatal("All plots must have same length")
	}

	if len(direct) < 20 {
		t.Fatal("Expected at least 20 bars")
	}

	for i := 20; i < minIntBBW(len(direct), 50); i++ {
		if math.IsNaN(direct[i]) || math.IsNaN(arrow[i]) {
			continue
		}

		if math.Abs(diff[i]) > 1e-10 {
			t.Errorf("bar[%d]: direct=%f, arrow=%f, diff=%f - expected identical values",
				i, direct[i], arrow[i], diff[i])
		}
	}

	t.Logf("✅ BBW mixed validated: direct == arrow")
}

func loadFixture(t *testing.T, relativePath string) string {
	t.Helper()

	fixturePath := filepath.Join("..", "fixtures", relativePath)
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", relativePath, err)
	}

	return string(content)
}

func minIntBBW(a, b int) int {
	if a < b {
		return a
	}
	return b
}
