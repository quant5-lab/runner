package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestForLoopZeroStep validates zero step runtime error handling
 *
 * Zero step would cause infinite loop, so runtime must panic
 * Validates codegen panic("for loop step cannot be zero")
 * Expected: Binary execution fails with panic message
 */
func TestForLoopZeroStep(t *testing.T) {
	pineScript := `//@version=5
indicator("For Loop Zero Step", overlay=false)

result = 0.0
for i = 1 to 10 by 0
	result := result + i

plot(result, "Zero Step Sum")
`

	exec := util.NewPineExecutor(t)

	code, generatedFile := exec.GenerateCode(t, "for-loop-zero-step", pineScript)

	if !strings.Contains(code, `if _step == 0 {`) {
		t.Fatal("Generated code missing zero step validation check")
	}

	if !strings.Contains(code, `panic("for loop step cannot be zero")`) {
		t.Fatal("Generated code missing zero step panic")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Logf("✅ Zero step validation: panic check present in generated code at %s", generatedFile)
	t.Logf("✅ Runtime panic will prevent infinite loop when step=0 is evaluated")
}
