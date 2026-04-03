package codegen

import (
	"strings"
	"testing"
)

/* TestVarPersistence_BarZeroGuard validates that var/varip declarations produce a bar-0 initialization guard across all value types and init expressions. */
func TestVarPersistence_BarZeroGuard(t *testing.T) {
	tests := []struct {
		name string
		pine string
	}{
		{
			name: "float zero init",
			pine: `//@version=5
strategy("Test", overlay=true)
var cumulative = 0.0
cumulative := cumulative + close
`,
		},
		{
			name: "float non-zero init",
			pine: `//@version=5
strategy("Test", overlay=true)
var myPrice = 100.5
`,
		},
		{
			name: "integer init",
			pine: `//@version=5
strategy("Test", overlay=true)
var counter = 0
counter := counter + 1
`,
		},
		{
			name: "boolean init",
			pine: `//@version=5
strategy("Test", overlay=true)
var triggered = false
if close > open
    triggered := true
`,
		},
		{
			name: "string init",
			pine: `//@version=5
strategy("Test", overlay=true)
var label = "initial"
`,
		},
		{
			name: "identifier init",
			pine: `//@version=5
strategy("Test", overlay=true)
var entryPrice = close
`,
		},
		{
			name: "na init",
			pine: `//@version=5
strategy("Test", overlay=true)
var entryPrice = na
if close > open
    entryPrice := close
`,
		},
		{
			name: "function call init",
			pine: `//@version=5
strategy("Test", overlay=true)
var highest = ta.highest(high, 10)
`,
		},
		{
			name: "typed float",
			pine: `//@version=5
strategy("Test", overlay=true)
var float myVal = 1.5
`,
		},
		{
			name: "typed int",
			pine: `//@version=5
strategy("Test", overlay=true)
var int count = 0
count := count + 1
`,
		},
		{
			name: "typed bool",
			pine: `//@version=5
strategy("Test", overlay=true)
var bool flag = false
`,
		},
		{
			name: "varip modifier",
			pine: `//@version=5
strategy("Test", overlay=true)
varip cumulative = 0.0
cumulative := cumulative + close
`,
		},
		{
			name: "string constant value",
			pine: `//@version=5
strategy("Test", overlay=true)
var sym = "BTCUSD"
plot(close)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
			NewCodeVerifier(code, t).MustContain("if i == 0 {")
		})
	}
}

/* TestVarPersistence_CarryForward validates that series-type var/varip declarations generate Get(1) carry-forward in the else branch. */
func TestVarPersistence_CarryForward(t *testing.T) {
	tests := []struct {
		name    string
		pine    string
		varName string
	}{
		{
			name:    "float accumulator",
			varName: "cumulative",
			pine: `//@version=5
strategy("Test", overlay=true)
var cumulative = 0.0
cumulative := cumulative + close
`,
		},
		{
			name:    "non-zero float",
			varName: "myPrice",
			pine: `//@version=5
strategy("Test", overlay=true)
var myPrice = 100.5
`,
		},
		{
			name:    "integer with reassignment",
			varName: "counter",
			pine: `//@version=5
strategy("Test", overlay=true)
var counter = 0
counter := counter + 1
`,
		},
		{
			name:    "boolean flag",
			varName: "triggered",
			pine: `//@version=5
strategy("Test", overlay=true)
var triggered = false
if close > open
    triggered := true
`,
		},
		{
			name:    "identifier init",
			varName: "entryPrice",
			pine: `//@version=5
strategy("Test", overlay=true)
var entryPrice = close
`,
		},
		{
			name:    "conditional update",
			varName: "maxClose",
			pine: `//@version=5
strategy("Test", overlay=true)
var maxClose = 0.0
if close > maxClose
    maxClose := close
plot(maxClose)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
			carryForward := tt.varName + "Series.Set(" + tt.varName + "Series.Get(1))"
			badCarryForward := tt.varName + "Series.Set(" + tt.varName + "Series.Get(0))"
			NewCodeVerifier(code, t).
				MustContain("} else {", carryForward).
				MustNotContain(badCarryForward)
		})
	}
}

/* TestVarPersistence_StringNativeStorage validates that var strings get bar-0 guard but no series carry-forward since Go strings persist natively. */
func TestVarPersistence_StringNativeStorage(t *testing.T) {
	pine := `//@version=5
strategy("Test", overlay=true)
var label = "initial"
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
	NewCodeVerifier(code, t).
		MustContain("if i == 0 {", "initial").
		MustNotContain("labelSeries.Set(labelSeries.Get(1))")
}

/* TestVarPersistence_PlainDeclarationsUnaffected validates that non-var declarations have no bar-0 guard or carry-forward. */
func TestVarPersistence_PlainDeclarationsUnaffected(t *testing.T) {
	tests := []struct {
		name    string
		pine    string
		varName string
	}{
		{
			name:    "expression assignment",
			varName: "x",
			pine: `//@version=5
strategy("Test", overlay=true)
x = close + open
`,
		},
		{
			name:    "typed assignment",
			varName: "y",
			pine: `//@version=5
strategy("Test", overlay=true)
float y = close * 2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
			NewCodeVerifier(code, t).
				MustContain(tt.varName + "Series.Set(").
				MustNotContain(tt.varName + "Series.Set(" + tt.varName + "Series.Get(1))")
		})
	}
}

/* TestVarPersistence_MultipleDeclarations validates each var declaration gets an independent bar-0 guard. */
func TestVarPersistence_MultipleDeclarations(t *testing.T) {
	pine := `//@version=5
strategy("Test", overlay=true)
var cumHigh = 0.0
var cumLow = 0.0
cumHigh := math.max(cumHigh, high)
cumLow := math.min(cumLow, low)
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
	NewCodeVerifier(code, t).
		MustContain(
			"cumHighSeries.Set(cumHighSeries.Get(1))",
			"cumLowSeries.Set(cumLowSeries.Get(1))",
		).
		CountOccurrences("if i == 0 {", 2)
}

/* TestVarPersistence_MixedVarAndPlain validates var and plain declarations coexist without interference. */
func TestVarPersistence_MixedVarAndPlain(t *testing.T) {
	pine := `//@version=5
strategy("Test", overlay=true)
var cumulative = 0.0
delta = close - open
cumulative := cumulative + delta
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
	NewCodeVerifier(code, t).
		MustContain(
			"cumulativeSeries.Set(cumulativeSeries.Get(1))",
			"deltaSeries.Set(",
		).
		MustNotContain("deltaSeries.Set(deltaSeries.Get(1))")
}

/* TestVarPersistence_VaripIdentical validates var and varip produce identical output in historical-only mode. */
func TestVarPersistence_VaripIdentical(t *testing.T) {
	pineVar := `//@version=5
strategy("Test", overlay=true)
var cumulative = 0.0
cumulative := cumulative + close
`
	pineVarip := `//@version=5
strategy("Test", overlay=true)
varip cumulative = 0.0
cumulative := cumulative + close
`
	codeVar, err := compilePineScript(pineVar)
	if err != nil {
		t.Fatalf("var compilation failed: %v", err)
	}
	codeVarip, err := compilePineScript(pineVarip)
	if err != nil {
		t.Fatalf("varip compilation failed: %v", err)
	}

	if codeVar != codeVarip {
		t.Error("var and varip should produce identical code in historical-only mode")
	}
}

/* TestVarPersistence_ExecutionContexts validates persistence works in all execution contexts. */
func TestVarPersistence_ExecutionContexts(t *testing.T) {
	tests := []struct {
		name    string
		pine    string
		varName string
	}{
		{
			name:    "top-level scope",
			varName: "x",
			pine: `//@version=5
strategy("Test", overlay=true)
var x = 0.0
x := x + close
plot(x)
`,
		},
		{
			name:    "for loop body",
			varName: "loopAccum",
			pine: `//@version=5
strategy("Test", overlay=true)
for j = 0 to 5
    var loopAccum = 0.0
    loopAccum := loopAccum + close
plot(close)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
			carryForward := tt.varName + "Series.Set(" + tt.varName + "Series.Get(1))"
			NewCodeVerifier(code, t).MustContain("if i == 0 {", carryForward)
		})
	}
}

/* TestVarPersistence_StructuralIntegrity validates balanced braces in generated code. */
func TestVarPersistence_StructuralIntegrity(t *testing.T) {
	pine := `//@version=5
strategy("Test", overlay=true)
var cumulative = 0.0
cumulative := cumulative + close
plot(cumulative)
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	opens := strings.Count(code, "{")
	closes := strings.Count(code, "}")
	if opens != closes {
		t.Errorf("Unbalanced braces: %d opens, %d closes", opens, closes)
	}
}

/* TestVarPersistence_ZeroInitNotSkipped validates that zero-literal var declarations with reassignment are not dropped by the reassignedVars optimizer. */
func TestVarPersistence_ZeroInitNotSkipped(t *testing.T) {
	pine := `//@version=5
strategy("Test", overlay=true)
var x = 0.0
x := x + 1
plot(x)
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
	NewCodeVerifier(code, t).
		MustContain("if i == 0 {", "xSeries.Set(xSeries.Get(1))")
}
