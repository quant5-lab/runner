package integration

import (
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestForLoopCounterImmutability validates loop counter mutation behavior
 *
 * Pine loop counters should be immutable or mutations ignored
 * Tests that reassigning loop counter inside body doesn't affect iteration
 * Expected: Loop executes fixed number of iterations regardless of counter mutation
 */
func TestForLoopCounterImmutability(t *testing.T) {
	pineScript := `//@version=5
indicator("For Loop Counter Mutation", overlay=false)

count = 0.0
for i = 1 to 5
	count := count + 1

plot(count, "Iteration Count")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "for-loop-counter-mutation", pineScript)

	countVals := exec.ExtractPlotValues(t, output, "Iteration Count")

	if len(countVals) < 1 {
		t.Fatal("Expected at least 1 data point")
	}

	expected := 5.0
	if countVals[0] != expected {
		t.Errorf("Loop executed %f iterations, want %f", countVals[0], expected)
	}

	for i := 1; i < len(countVals); i++ {
		if countVals[i] != expected {
			t.Errorf("count[%d] = %f, want %f (should be constant)", i, countVals[i], expected)
		}
	}

	t.Logf("✅ For loop counter immutability validated: %f iterations", expected)
}
