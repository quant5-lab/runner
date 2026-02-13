//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* for-in single element generates `for _, elem := range collection` */
func TestForInSingleElementCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("ForIn Single", overlay=false)

arr = array.from(1.0, 2.0, 3.0)
sum = 0.0
for val in arr
	sum := sum + val

plot(sum, "ForIn Sum")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "forin-single", pineScript)

	if !strings.Contains(code, "for _, val := range") {
		t.Fatal("Single element for-in should generate `for _, val := range`")
	}
}

/* for-in tuple destructuring generates `for idx, elem := range collection` */
func TestForInTupleDestructuringCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("ForIn Tuple", overlay=false)

arr = array.from(10.0, 20.0, 30.0)
total = 0.0
for [idx, val] in arr
	total := total + val + idx

plot(total, "ForIn Tuple Sum")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "forin-tuple", pineScript)

	if !strings.Contains(code, "for idx, val := range") {
		t.Fatal("Tuple for-in should generate `for idx, val := range`")
	}
}

/* for-in with break generates break inside loop body */
func TestForInWithBreakCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("ForIn Break", overlay=false)

arr = array.from(1.0, 2.0, 3.0, 4.0, 5.0)
sum = 0.0
for val in arr
	if val > 3
		break
	sum := sum + val

plot(sum, "ForIn Break Sum")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "forin-break", pineScript)

	if !strings.Contains(code, "for _, val := range") {
		t.Fatal("for-in should generate range loop")
	}
	if !strings.Contains(code, "break") {
		t.Fatal("Generated code should contain break statement")
	}
}

/* for-in with continue generates continue inside loop body */
func TestForInWithContinueCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("ForIn Continue", overlay=false)

arr = array.from(1.0, 2.0, 3.0)
sum = 0.0
for val in arr
	if val == 2
		continue
	sum := sum + val

plot(sum, "ForIn Continue Sum")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "forin-continue", pineScript)

	if !strings.Contains(code, "for _, val := range") {
		t.Fatal("for-in should generate range loop")
	}
	if !strings.Contains(code, "continue") {
		t.Fatal("Generated code should contain continue statement")
	}
}

/* for-in tuple with index counter resolves index to float64() */
func TestForInTupleIndexResolutionCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("ForIn Index Resolution", overlay=false)

arr = array.from(10.0, 20.0, 30.0)
result = 0.0
for [i, val] in arr
	result := result + val + i

plot(result, "Index Resolution")
`

	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "forin-index-resolution", pineScript)

	if !strings.Contains(code, "for i, val := range") {
		t.Fatal("Tuple for-in with index should generate `for i, val := range`")
	}

	if !strings.Contains(code, "float64(i)") {
		t.Fatal("Index variable should resolve to float64(i) in expressions")
	}
}
