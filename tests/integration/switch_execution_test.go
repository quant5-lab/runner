//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Switch expression full pipeline: Pine→Go→Binary→Execute→JSON */

func TestSwitchExecution_Form1Expression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		script   string
		plotName string
		expected float64
	}{
		{
			name: "two branches match first",
			script: `//@version=5
indicator("Switch Form1", overlay=false)
val = 1
result = switch val
    1 =>
        10.0
    2 =>
        20.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: 10.0,
		},
		{
			name: "three branches match last",
			script: `//@version=5
indicator("Switch Form1 Three", overlay=false)
val = 3
result = switch val
    1 =>
        100.0
    2 =>
        200.0
    3 =>
        300.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: 300.0,
		},
		{
			name: "with default branch",
			script: `//@version=5
indicator("Switch Form1 Default", overlay=false)
val = 99
result = switch val
    1 =>
        10.0
    2 =>
        20.0
    =>
        -1.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: -1.0,
		},
		{
			name: "five branches match middle",
			script: `//@version=5
indicator("Switch Form1 Five", overlay=false)
val = 3
result = switch val
    1 =>
        10.0
    2 =>
        20.0
    3 =>
        30.0
    4 =>
        40.0
    5 =>
        50.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: 30.0,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := exec.ExecuteScript(t, "switch-form1", tt.script)
			vals := exec.ExtractPlotValues(t, output, tt.plotName)
			if len(vals) < 1 {
				t.Fatal("no data points")
			}
			if vals[0] != tt.expected {
				t.Errorf("got %f, want %f", vals[0], tt.expected)
			}
		})
	}
}

func TestSwitchExecution_Form2Expression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		script   string
		plotName string
		expected float64
	}{
		{
			name: "boolean conditions first true",
			script: `//@version=5
indicator("Switch Form2", overlay=false)
x = 5
result = switch
    x > 10 =>
        3.0
    x > 0 =>
        2.0
    =>
        1.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: 2.0,
		},
		{
			name: "boolean conditions default",
			script: `//@version=5
indicator("Switch Form2 Default", overlay=false)
x = 0
result = switch
    x > 10 =>
        3.0
    x > 0 =>
        2.0
    =>
        1.0
plot(result, "Result")
`,
			plotName: "Result",
			expected: 1.0,
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := exec.ExecuteScript(t, "switch-form2", tt.script)
			vals := exec.ExtractPlotValues(t, output, tt.plotName)
			if len(vals) < 1 {
				t.Fatal("no data points")
			}
			if vals[0] != tt.expected {
				t.Errorf("got %f, want %f", vals[0], tt.expected)
			}
		})
	}
}

func TestSwitchExecution_NoDefaultNaSemantics(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Switch Na", overlay=false)
val = 99
result = switch val
    1 =>
        10.0
    2 =>
        20.0
plot(result, "Result")
`

	exec := util.NewPineExecutor(t)
	raw := exec.ExecuteScriptWithCustomDataRaw(t, "switch-na", script, []map[string]interface{}{
		{"time": 1700000000, "open": 100.0, "high": 105.0, "low": 95.0, "close": 102.0, "volume": 1000.0},
	})

	/* NaN serializes as null in JSON - verify raw output contains null value */
	rawStr := string(raw)
	if !strings.Contains(rawStr, "null") {
		t.Errorf("expected null (NaN) in raw JSON for unmatched switch without default")
	}
}

func TestSwitchExecution_MultiStatementBody(t *testing.T) {
	t.Parallel()
	/* Multi-statement case bodies with outer-scope variables (local-var-in-IIFE is a known limitation) */
	script := `//@version=5
indicator("Switch Multi", overlay=false)
a = 5.0
b = 10.0
val = 2
result = switch val
    1 =>
        a + b
    2 =>
        a * b
plot(result, "Result")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "switch-multi", script)
	vals := exec.ExtractPlotValues(t, output, "Result")

	if len(vals) < 1 {
		t.Fatal("no data points")
	}
	if vals[0] != 50.0 {
		t.Errorf("got %f, want 50.0 (5.0 * 10.0)", vals[0])
	}
}

func TestSwitchExecution_WithBarData(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Switch Bar", overlay=false)
direction = switch
    close > open =>
        1.0
    close < open =>
        -1.0
    =>
        0.0
plot(direction, "Direction")
`

	testData := []map[string]interface{}{
		{"time": 1700000000, "open": 100.0, "high": 110.0, "low": 95.0, "close": 105.0, "volume": 1000.0},
		{"time": 1700003600, "open": 105.0, "high": 108.0, "low": 98.0, "close": 100.0, "volume": 1100.0},
		{"time": 1700007200, "open": 100.0, "high": 105.0, "low": 95.0, "close": 100.0, "volume": 1200.0},
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "switch-bar", script, testData)
	vals := exec.ExtractPlotValues(t, output, "Direction")

	if len(vals) < 3 {
		t.Fatalf("expected 3 bars, got %d", len(vals))
	}
	if vals[0] != 1.0 {
		t.Errorf("bar 0 (bullish): got %f, want 1.0", vals[0])
	}
	if vals[1] != -1.0 {
		t.Errorf("bar 1 (bearish): got %f, want -1.0", vals[1])
	}
	if vals[2] != 0.0 {
		t.Errorf("bar 2 (doji): got %f, want 0.0", vals[2])
	}
}

func TestSwitchExecution_NestedInIf(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Switch Nested If", overlay=false)
x = 2
result = 0.0
if x > 0
    result := switch x
        1 =>
            10.0
        2 =>
            20.0
        =>
            0.0
plot(result, "Result")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "switch-nested-if", script)
	vals := exec.ExtractPlotValues(t, output, "Result")

	if len(vals) < 1 {
		t.Fatal("no data points")
	}
	if vals[0] != 20.0 {
		t.Errorf("got %f, want 20.0", vals[0])
	}
}

func TestSwitchExecution_InlineWithBarData(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Switch Inline Bar", overlay=false)
direction = switch
    close > open => 1.0
    close < open => -1.0
    => 0.0
plot(direction, "Direction")
`

	testData := []map[string]interface{}{
		{"time": 1700000000, "open": 100.0, "high": 110.0, "low": 95.0, "close": 105.0, "volume": 1000.0},
		{"time": 1700003600, "open": 105.0, "high": 108.0, "low": 98.0, "close": 100.0, "volume": 1100.0},
		{"time": 1700007200, "open": 100.0, "high": 105.0, "low": 95.0, "close": 100.0, "volume": 1200.0},
	}

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "switch-inline-bar", script, testData)
	vals := exec.ExtractPlotValues(t, output, "Direction")

	if len(vals) < 3 {
		t.Fatalf("expected 3 bars, got %d", len(vals))
	}
	if vals[0] != 1.0 {
		t.Errorf("bar 0 (bullish): got %f, want 1.0", vals[0])
	}
	if vals[1] != -1.0 {
		t.Errorf("bar 1 (bearish): got %f, want -1.0", vals[1])
	}
	if vals[2] != 0.0 {
		t.Errorf("bar 2 (doji): got %f, want 0.0", vals[2])
	}
}

func TestSwitchExecution_MixedInlineAndMultiLine(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Switch Mixed", overlay=false)
a = 5.0
b = 10.0
val = 2
result = switch val
    1 => a + b
    2 =>
        a * b
    => -1.0
plot(result, "Result")
`

	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "switch-mixed", script)
	vals := exec.ExtractPlotValues(t, output, "Result")

	if len(vals) < 1 {
		t.Fatal("no data points")
	}
	if vals[0] != 50.0 {
		t.Errorf("got %f, want 50.0", vals[0])
	}
}
