package codegen

import (
	"strings"
	"testing"
)

func TestUnusedSeriesSuppression(t *testing.T) {
	pine := `
//@version=5
indicator("Test")
calc(len) =>
    sumGain = 0.0
    count = 0
    for i = 0 to len - 1
        if close[i] > open[i]
            gain = ((close[i] - open[i]) / open[i]) * 100
            sumGain := sumGain + gain
            count := count + 1
    sumGain / count
plot(calc(10))
`

	goCode, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Should have suppression for 'gain' (not loop-modified)
	if !strings.Contains(goCode, "_ = gainSeries") {
		t.Errorf("Missing suppression for gainSeries\nGenerated code:\n%s", goCode)
	}

	// Should NOT have suppression for 'sumGain' (loop-modified)
	if strings.Contains(goCode, "_ = sumGainSeries") {
		t.Errorf("Should not suppress sumGainSeries (loop-modified)\nGenerated code:\n%s", goCode)
	}

	// Should NOT have suppression for 'count' (loop-modified)
	if strings.Contains(goCode, "_ = countSeries") {
		t.Errorf("Should not suppress countSeries (loop-modified)\nGenerated code:\n%s", goCode)
	}
}
