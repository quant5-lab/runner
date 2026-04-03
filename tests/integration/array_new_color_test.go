//go:build integration

package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestArrayNewColorEmpty(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
indicator("Color Array Empty", overlay=false)
arr = array.new_color()
sz = array.size(arr)
plot(sz, "Empty Size")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-array-new-color-empty", pineScript)

	emptySize := exec.ExtractPlotValues(t, output, "Empty Size")
	if len(emptySize) == 0 {
		t.Fatal("Expected plot values for Empty Size")
	}
	if emptySize[0] != 0 {
		t.Errorf("Empty color array size = %f, want 0", emptySize[0])
	}
}

func TestArrayNewColorSized(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
indicator("Color Array Sized", overlay=false)
arr = array.new_color(5)
sz = array.size(arr)
plot(sz, "Sized Size")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "test-array-new-color-sized", pineScript)

	sizedSize := exec.ExtractPlotValues(t, output, "Sized Size")
	if len(sizedSize) == 0 {
		t.Fatal("Expected plot values for Sized Size")
	}
	if sizedSize[0] != 5 {
		t.Errorf("Sized color array size = %f, want 5", sizedSize[0])
	}
}

func TestArrayNewColorCodegen(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
indicator("Color Array Codegen")
arr1 = array.new_color()
arr2 = array.new_color(5)
plot(1)
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "color-array-codegen", pineScript)

	requiredCode := []string{
		"[]float64{}",
		"make([]float64, int(5))",
	}

	for _, exp := range requiredCode {
		if !contains(goCode, exp) {
			t.Errorf("Missing codegen: %s", exp)
		}
	}
}
