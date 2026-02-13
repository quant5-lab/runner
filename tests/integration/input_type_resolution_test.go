//go:build integration

package integration

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Named-arg-only input(defval=bool) resolves to input.bool, not float64 */
func TestInputTypeResolution_NamedBoolDefval(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Named Bool Input", overlay=true)

enableLong = input(defval=true, title="Enable Long")
sma20 = ta.sma(close, 20)

if enableLong and close > sma20
    strategy.entry("Long", strategy.long)
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-named-bool", pineScript)

	if strings.Contains(code, "var enableLong float64") {
		t.Fatal("Named bool defval incorrectly inferred as float64")
	}
	if !strings.Contains(code, "const enableLong = true") {
		t.Error("Expected const enableLong = true")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* Named-arg-only input(defval=string) resolves to input.string */
func TestInputTypeResolution_NamedStringDefval(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Named String Input")

maType = input(defval="EMA", title="MA Type")
plot(ta.sma(close, 14))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-named-string", pineScript)

	if strings.Contains(code, "var maType float64") {
		t.Fatal("Named string defval incorrectly inferred as float64")
	}
	if !strings.Contains(code, `const maType = "EMA"`) {
		t.Error(`Expected const maType = "EMA"`)
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* Named-arg-only input(defval=int) resolves to input.int */
func TestInputTypeResolution_NamedIntDefval(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Named Int Input")

length = input(defval=14, title="Length")
plot(ta.sma(close, length))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-named-int", pineScript)

	if strings.Contains(code, "var length float64") {
		t.Fatal("Named int defval incorrectly inferred as float64")
	}
	if !strings.Contains(code, "const length = 14") {
		t.Error("Expected const length = 14")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* v4 type=input.source resolves via explicit type parameter */
func TestInputTypeResolution_V4ExplicitSource(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=4
study("V4 Source Input")

src = input(title="Source", type=input.source, defval=close)
plot(ta.sma(src, 14))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-v4-source", pineScript)

	if strings.Contains(code, "var src float64") {
		t.Fatal("v4 input.source incorrectly inferred as plain float64 variable")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* v4 type=input.integer resolves to input.int */
func TestInputTypeResolution_V4ExplicitInteger(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=4
study("V4 Integer Input")

length = input(title="Period", type=input.integer, defval=20)
plot(ta.sma(close, length))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-v4-integer", pineScript)

	if !strings.Contains(code, "const length = 20") {
		t.Error("Expected const length = 20 from v4 input.integer")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* Multiple named-arg bool inputs in same strategy resolve independently */
func TestInputTypeResolution_MultipleBoolInputs(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Multi Bool Inputs", overlay=true)

enableLong = input(defval=true, title="Enable Long")
enableShort = input(defval=false, title="Enable Short")
sma20 = ta.sma(close, 20)

if enableLong and close > sma20
    strategy.entry("Long", strategy.long)

if enableShort and close < sma20
    strategy.entry("Short", strategy.short)
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "input-multi-bool", pineScript)

	if !strings.Contains(code, "const enableLong = true") {
		t.Error("Expected const enableLong = true")
	}
	if !strings.Contains(code, "const enableShort = false") {
		t.Error("Expected const enableShort = false")
	}

	if err := exec.CompileCode(t, code); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}
}

/* Bool input gates strategy entry — end-to-end execution */
func TestInputTypeResolution_BoolGatedExecution(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Bool Gated Strategy", overlay=true)

enableLong = input(defval=true, title="Enable Long")
enableShort = input(defval=false, title="Enable Short")

if enableLong and close > open
    strategy.entry("Long", strategy.long)

if enableShort and close < open
    strategy.entry("Short", strategy.short)
`
	testData := []map[string]interface{}{
		{"time": 1704067200, "open": 100.0, "high": 102.0, "low": 98.0, "close": 99.0, "volume": 1000.0},
		{"time": 1704070800, "open": 99.0, "high": 103.0, "low": 98.0, "close": 101.0, "volume": 1000.0},
		{"time": 1704074400, "open": 101.0, "high": 104.0, "low": 100.0, "close": 103.0, "volume": 1000.0},
		{"time": 1704078000, "open": 103.0, "high": 105.0, "low": 101.0, "close": 102.0, "volume": 1000.0},
		{"time": 1704081600, "open": 102.0, "high": 106.0, "low": 101.0, "close": 105.0, "volume": 1000.0},
		{"time": 1704085200, "open": 105.0, "high": 107.0, "low": 103.0, "close": 104.0, "volume": 1000.0},
		{"time": 1704088800, "open": 104.0, "high": 108.0, "low": 103.0, "close": 107.0, "volume": 1000.0},
		{"time": 1704092400, "open": 107.0, "high": 109.0, "low": 105.0, "close": 106.0, "volume": 1000.0},
		{"time": 1704096000, "open": 106.0, "high": 110.0, "low": 105.0, "close": 109.0, "volume": 1000.0},
		{"time": 1704099600, "open": 109.0, "high": 111.0, "low": 107.0, "close": 108.0, "volume": 1000.0},
	}

	exec := util.NewPineExecutor(t)
	rawOutput := exec.ExecuteScriptWithCustomDataRaw(t, "bool-gated", pineScript, testData)

	var result struct {
		Strategy struct {
			OpenTrades []struct {
				Direction string `json:"direction"`
			} `json:"openTrades"`
			ClosedTrades []struct {
				Direction string `json:"direction"`
			} `json:"closedTrades"`
		} `json:"strategy"`
	}

	if err := json.Unmarshal(rawOutput, &result); err != nil {
		t.Fatalf("Parse result: %v", err)
	}

	allTrades := append(result.Strategy.OpenTrades, result.Strategy.ClosedTrades...)
	for i, trade := range allTrades {
		if trade.Direction == "short" {
			t.Errorf("Trade %d is short — enableShort=false should prevent short entries", i)
		}
	}

	hasLong := false
	for _, trade := range allTrades {
		if trade.Direction == "long" {
			hasLong = true
			break
		}
	}
	if !hasLong {
		t.Fatal("Expected at least one long entry — enableLong=true should allow entries")
	}
}

/* All const-producing input types: explicit v4 type param and v5 direct call → const → compile */
func TestInputTypeResolution_ExplicitTypeToConst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		pineScript    string
		expectedConst string
	}{
		/* v4 explicit type param → const */
		{
			name: "v4_float",
			pineScript: `//@version=4
study("V4 Float")
mult = input(1.5, title="Mult", type=input.float)
plot(close)
`,
			expectedConst: "const mult = 1.50",
		},
		{
			name: "v4_int",
			pineScript: `//@version=4
study("V4 Int")
period = input(14, title="Length", type=input.integer)
plot(ta.sma(close, period))
`,
			expectedConst: "const period = 14",
		},
		{
			name: "v4_bool",
			pineScript: `//@version=4
study("V4 Bool")
show = input(defval=true, title="Show", type=input.bool)
plot(close)
`,
			expectedConst: "const show = true",
		},
		{
			name: "v4_string",
			pineScript: `//@version=4
study("V4 String")
maType = input(defval="EMA", title="MA", type=input.string)
plot(close)
`,
			expectedConst: `const maType = "EMA"`,
		},
		{
			name: "v4_session",
			pineScript: `//@version=4
study("V4 Session")
sess = input(defval="0930-1600", title="Session", type=input.session)
plot(close)
`,
			expectedConst: `const sess = "0930-1600"`,
		},
		{
			name: "v4_symbol",
			pineScript: `//@version=4
study("V4 Symbol")
sym = input("EURUSD", title="Symbol", type=input.symbol)
plot(close)
`,
			expectedConst: `const sym = "EURUSD"`,
		},
		{
			name: "v4_timeframe",
			pineScript: `//@version=4
study("V4 Timeframe")
tf = input("D", title="Timeframe", type=input.timeframe)
plot(close)
`,
			expectedConst: `const tf = "D"`,
		},
		{
			name: "v4_price",
			pineScript: `//@version=4
study("V4 Price")
targetPrice = input(99.5, title="Target", type=input.price)
plot(close)
`,
			expectedConst: "const targetPrice = 99.50",
		},
		{
			name: "v4_time",
			pineScript: `//@version=4
study("V4 Time")
startTime = input(defval=0, title="Start", type=input.time)
plot(close)
`,
			expectedConst: "const startTime = 0",
		},
		/* v5 direct call → const */
		{
			name: "v5_text_area",
			pineScript: `//@version=5
indicator("V5 Text Area")
notes = input.text_area(defval="my notes", title="Notes")
plot(close)
`,
			expectedConst: `const notes = "my notes"`,
		},
		{
			name: "v5_color",
			pineScript: `//@version=5
indicator("V5 Color")
lineColor = input.color(title="Line Color")
plot(close)
`,
			expectedConst: `const lineColor = ""`,
		},
		{
			name: "v5_color_with_defval",
			pineScript: `//@version=5
indicator("V5 Color Defval")
lineColor = input.color(defval=color.red, title="Line Color")
plot(close)
`,
			expectedConst: `const lineColor = "#FF5252"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := util.NewPineExecutor(t)
			code, _ := exec.GenerateCode(t, tt.name, tt.pineScript)

			if !strings.Contains(code, tt.expectedConst) {
				t.Fatalf("Expected %q in generated code:\n%s", tt.expectedConst, code)
			}

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
		})
	}
}
