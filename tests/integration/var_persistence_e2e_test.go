//go:build integration

package integration

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestVarPersistence_Accumulator validates that var produces a monotonically increasing cumulative sum. */
func TestVarPersistence_Accumulator(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Var Accumulator", overlay=false)
var cumVol = 0.0
cumVol := cumVol + volume
plot(cumVol, "Cumulative Volume")
`

	bars := varTestBars(20, 100.0)
	vals := executeVarScript(t, "var-accumulator", pineScript, bars, "Cumulative Volume")

	prev := 0.0
	for i, v := range vals {
		if i > 0 && v < prev {
			t.Errorf("bar %d: cumulative decreased (%f -> %f)", i, prev, v)
		}
		prev = v
	}

	expectedTotal := 0.0
	for _, bar := range bars {
		expectedTotal += bar["volume"].(float64)
	}
	if math.Abs(vals[len(vals)-1]-expectedTotal) > 0.01 {
		t.Errorf("final cumVol: got %f, want %f", vals[len(vals)-1], expectedTotal)
	}
}

/* TestVarPersistence_RunningMax validates that var conditional update produces a non-decreasing series. */
func TestVarPersistence_RunningMax(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Var Running Max", overlay=false)
var maxClose = 0.0
if close > maxClose
    maxClose := close
plot(maxClose, "Running Max")
`

	bars := varTestBars(30, 50.0)
	vals := executeVarScript(t, "var-running-max", pineScript, bars, "Running Max")

	prev := 0.0
	for i, v := range vals {
		if v < prev {
			t.Errorf("bar %d: running max decreased (%f -> %f)", i, prev, v)
		}
		prev = v
	}

	highestClose := 0.0
	for _, bar := range bars {
		c := bar["close"].(float64)
		if c > highestClose {
			highestClose = c
		}
	}
	if math.Abs(vals[len(vals)-1]-highestClose) > 0.01 {
		t.Errorf("final max: got %f, want %f", vals[len(vals)-1], highestClose)
	}
}

/* TestVarPersistence_NonZeroInit validates that var with non-zero init starts at the correct offset. */
func TestVarPersistence_NonZeroInit(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Var NonZero", overlay=false)
var counter = 100.0
counter := counter + 1
plot(counter, "Counter")
`

	bars := varTestBars(10, 50.0)
	vals := executeVarScript(t, "var-nonzero", pineScript, bars, "Counter")

	/* Bar N: counter = 100 + N + 1 */
	for i, v := range vals {
		expected := 100.0 + float64(i) + 1.0
		if math.Abs(v-expected) > 0.01 {
			t.Errorf("bar %d: got %f, want %f", i, v, expected)
		}
	}
}

/* TestVarPersistence_VaripIdentical validates var and varip produce identical results in historical mode. */
func TestVarPersistence_VaripIdentical(t *testing.T) {
	t.Parallel()
	pineVar := `//@version=5
strategy("Var Test", overlay=false)
var cumVol = 0.0
cumVol := cumVol + volume
plot(cumVol, "Cumulative Volume")
`
	pineVarip := `//@version=5
strategy("Varip Test", overlay=false)
varip cumVol = 0.0
cumVol := cumVol + volume
plot(cumVol, "Cumulative Volume")
`

	bars := varTestBars(15, 100.0)
	valsVar := executeVarScript(t, "var-test", pineVar, bars, "Cumulative Volume")
	valsVarip := executeVarScript(t, "varip-test", pineVarip, bars, "Cumulative Volume")

	if len(valsVar) != len(valsVarip) {
		t.Fatalf("length mismatch: var=%d, varip=%d", len(valsVar), len(valsVarip))
	}

	for i := range valsVar {
		if math.Abs(valsVar[i]-valsVarip[i]) > 0.001 {
			t.Errorf("bar %d: var=%f, varip=%f", i, valsVar[i], valsVarip[i])
		}
	}
}

/* TestVarPersistence_TypedDeclaration validates typed var declarations (var float x = ...) execute correctly. */
func TestVarPersistence_TypedDeclaration(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Var Typed", overlay=false)
var float runSum = 0.0
runSum := runSum + close
plot(runSum, "Running Sum")
`

	bars := varTestBars(10, 100.0)
	vals := executeVarScript(t, "var-typed", pineScript, bars, "Running Sum")

	prev := 0.0
	for i, v := range vals {
		if i > 0 && v <= prev {
			t.Errorf("bar %d: running sum did not increase (%f -> %f)", i, prev, v)
		}
		prev = v
	}
}

/* TestVarPersistence_Compilation validates that various var/varip patterns compile without Go errors. */
func TestVarPersistence_Compilation(t *testing.T) {
	t.Parallel()
	scripts := []struct {
		name   string
		script string
	}{
		{
			name: "basic float",
			script: `//@version=5
strategy("Test", overlay=false)
var x = 0.0
x := x + close
plot(x, "X")
`,
		},
		{
			name: "typed float",
			script: `//@version=5
strategy("Test", overlay=false)
var float y = 1.5
y := y + close
plot(y, "Y")
`,
		},
		{
			name: "varip float",
			script: `//@version=5
strategy("Test", overlay=false)
varip z = 0.0
z := z + volume
plot(z, "Z")
`,
		},
		{
			name: "conditional update",
			script: `//@version=5
strategy("Test", overlay=false)
var maxHigh = 0.0
if high > maxHigh
    maxHigh := high
plot(maxHigh, "MaxHigh")
`,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range scripts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			code, _ := exec.GenerateCode(t, tt.name, tt.script)
			if code == "" {
				t.Fatal("Generated code is empty")
			}
		})
	}
}

/* executeVarScript is a test helper that runs a Pine script and extracts plot values. */
func executeVarScript(t *testing.T, name, script string, bars []map[string]interface{}, plotName string) []float64 {
	t.Helper()

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, name, script, bars)

	var parsed struct {
		Indicators map[string]struct {
			Data []struct {
				Time  int64   `json:"time"`
				Value float64 `json:"value"`
			} `json:"data"`
		} `json:"indicators"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("Parse result JSON: %v", err)
	}

	indicator, ok := parsed.Indicators[plotName]
	if !ok {
		keys := make([]string, 0, len(parsed.Indicators))
		for k := range parsed.Indicators {
			keys = append(keys, k)
		}
		t.Fatalf("Missing indicator %q, available: %v", plotName, keys)
	}

	if len(indicator.Data) == 0 {
		t.Fatal("No indicator data points")
	}

	vals := make([]float64, len(indicator.Data))
	for i, pt := range indicator.Data {
		vals[i] = pt.Value
	}
	return vals
}

func varTestBars(count int, baseClose float64) []map[string]interface{} {
	bars := make([]map[string]interface{}, count)
	baseTime := int64(1700000000)
	for i := 0; i < count; i++ {
		close := baseClose + float64(i)*0.5
		bars[i] = map[string]interface{}{
			"time":   baseTime + int64(i)*3600,
			"open":   close - 0.5,
			"high":   close + 1.0,
			"low":    close - 1.0,
			"close":  close,
			"volume": 1000.0 + float64(i)*100,
		}
	}
	return bars
}
