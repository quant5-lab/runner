package integration

import (
	"math"
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* Bars with known timestamps for deterministic time validation */
func knownTimeBars() []map[string]interface{} {
	base := int64(1700000000)
	bars := make([]map[string]interface{}, 20)
	for i := range bars {
		bars[i] = map[string]interface{}{
			"time":   base + int64(i*3600),
			"open":   100.0 + float64(i),
			"high":   105.0 + float64(i),
			"low":    95.0 + float64(i),
			"close":  102.0 + float64(i),
			"volume": 1000.0,
		}
	}
	return bars
}

func TestBuiltinTime_CurrentBarAccess(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Time Current Bar", overlay=false)
plot(time, "time_val")
`
	bars := knownTimeBars()
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "time-current-bar", script, bars)
	vals := exec.ExtractPlotValues(t, output, "time_val")

	if len(vals) < len(bars) {
		t.Fatalf("Expected %d bars, got %d", len(bars), len(vals))
	}

	for i, bar := range bars {
		expectedMs := bar["time"].(int64) * 1000
		if vals[i] != float64(expectedMs) {
			t.Errorf("bar[%d]: got %f, want %f (seconds→ms conversion)", i, vals[i], float64(expectedMs))
		}
	}
}

func TestBuiltinTime_HistoricalSubscript(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Time Subscript", overlay=false)
plot(time, "t0")
plot(time[1], "t1")
plot(time[2], "t2")
`
	bars := knownTimeBars()
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "time-subscript", script, bars)

	t0 := exec.ExtractPlotValues(t, output, "t0")
	t1 := exec.ExtractPlotValues(t, output, "t1")
	t2 := exec.ExtractPlotValues(t, output, "t2")

	for i := 2; i < len(bars); i++ {
		if t0[i] <= t1[i] {
			t.Errorf("bar[%d]: time[0]=%f should be > time[1]=%f", i, t0[i], t1[i])
		}
		if t1[i] <= t2[i] {
			t.Errorf("bar[%d]: time[1]=%f should be > time[2]=%f", i, t1[i], t2[i])
		}
		expectedDeltaMs := float64(3600 * 1000)
		if math.Abs(t0[i]-t1[i]-expectedDeltaMs) > 0.001 {
			t.Errorf("bar[%d]: time[0]-time[1]=%f, want %f", i, t0[i]-t1[i], expectedDeltaMs)
		}
	}
}

func TestBuiltinTime_SecurityArrowCodegen(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		script   string
		contains string
	}{
		{
			name: "time in security arrow body",
			script: `//@version=5
indicator("Time Security Arrow", overlay=false)

getDailyTime() =>
    request.security(syminfo.tickerid, "1D", time)

result = getDailyTime()
plot(result, "daily_time")
`,
			contains: "timeSeries",
		},
		{
			name: "time subscript in security arrow body",
			script: `//@version=5
indicator("Time Security Arrow Sub", overlay=false)

getDailyTimeDelta() =>
    request.security(syminfo.tickerid, "1D", time - time[1])

result = getDailyTimeDelta()
plot(result, "delta")
`,
			contains: "timeSeries",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, _ := exec.GenerateCode(t, "time-sec-arrow", tc.script)

			if !strings.Contains(code, tc.contains) {
				t.Errorf("generated code missing %q", tc.contains)
			}
		})
	}
}

func TestBuiltinTime_InTAFunction(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Time in TA", overlay=false)
timeSma = ta.sma(time, 5)
plot(timeSma, "time_sma")
`
	bars := knownTimeBars()
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "time-ta", script, bars)
	vals := exec.ExtractPlotValues(t, output, "time_sma")

	for i := 4; i < len(bars); i++ {
		if math.IsNaN(vals[i]) || vals[i] == 0 {
			t.Errorf("bar[%d]: time SMA should be non-zero after warmup, got %f", i, vals[i])
		}
	}
}

func TestBuiltinNamespace_Codegen(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		expr     string
		contains []string
	}{
		{
			name:     "barstate.isfirst",
			expr:     "barstate.isfirst ? 1.0 : 0.0",
			contains: []string{"ctx.BarIndex == 0"},
		},
		{
			name:     "barstate.islast",
			expr:     "barstate.islast ? 1.0 : 0.0",
			contains: []string{"ctx.BarIndex == len(ctx.Data)-1"},
		},
		{
			name:     "timeframe.isintraday",
			expr:     "timeframe.isintraday ? 1.0 : 0.0",
			contains: []string{"ctx.IsIntraday"},
		},
		{
			name:     "timeframe.isdaily",
			expr:     "timeframe.isdaily ? 1.0 : 0.0",
			contains: []string{"ctx.IsDaily"},
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			script := `//@version=5
indicator("NS ` + tc.name + `", overlay=false)
plot(` + tc.expr + `, "result")
`
			code, _ := exec.GenerateCode(t, "ns-"+strings.ReplaceAll(tc.name, ".", "-"), script)

			for _, expected := range tc.contains {
				if !strings.Contains(code, expected) {
					t.Errorf("generated code missing %q for %s", expected, tc.name)
				}
			}

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("compilation failed for %s: %v", tc.name, err)
			}
		})
	}
}

func TestBuiltinNamespace_StringCodegen(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		script   string
		contains string
	}{
		{
			name: "syminfo.tickerid in security",
			script: `//@version=5
indicator("NS syminfo.tickerid", overlay=true)
result = request.security(syminfo.tickerid, "1D", close)
plot(result, "result")
`,
			contains: "ctx.Symbol",
		},
		{
			name: "timeframe.period in security",
			script: `//@version=5
indicator("NS timeframe.period", overlay=true)
result = request.security(syminfo.tickerid, timeframe.period, close)
plot(result, "result")
`,
			contains: "ctx.Timeframe",
		},
		{
			name: "syminfo.timezone codegen",
			script: `//@version=5
indicator("NS syminfo.timezone", overlay=true)
result = request.security(syminfo.tickerid, "1D", close)
plot(result, "result")
`,
			contains: "ctx.Symbol",
		},
	}

	exec := util.NewPineExecutor(t)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			code, _ := exec.GenerateCode(t, "ns-str-"+strings.ReplaceAll(tc.name, " ", "-"), tc.script)

			if !strings.Contains(code, tc.contains) {
				t.Errorf("generated code missing %q for %s", tc.contains, tc.name)
			}

			if err := exec.CompileCode(t, code); err != nil {
				t.Fatalf("compilation failed for %s: %v", tc.name, err)
			}
		})
	}
}

func TestBuiltinNamespace_BarstateExecution(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Barstate Execution", overlay=false)
plot(barstate.isfirst ? 1.0 : 0.0, "isfirst")
plot(barstate.islast ? 1.0 : 0.0, "islast")
`
	bars := knownTimeBars()
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "barstate-exec", script, bars)

	isfirst := exec.ExtractPlotValues(t, output, "isfirst")
	islast := exec.ExtractPlotValues(t, output, "islast")

	if isfirst[0] != 1.0 {
		t.Errorf("barstate.isfirst on bar 0: got %f, want 1.0", isfirst[0])
	}
	for i := 1; i < len(bars); i++ {
		if isfirst[i] != 0.0 {
			t.Errorf("barstate.isfirst on bar %d: got %f, want 0.0", i, isfirst[i])
		}
	}

	if islast[len(bars)-1] != 1.0 {
		t.Errorf("barstate.islast on last bar: got %f, want 1.0", islast[len(bars)-1])
	}
	for i := 0; i < len(bars)-1; i++ {
		if islast[i] != 0.0 {
			t.Errorf("barstate.islast on bar %d: got %f, want 0.0", i, islast[i])
		}
	}
}

func TestBuiltinNamespace_TimeframeExecution(t *testing.T) {
	t.Parallel()
	script := `//@version=5
indicator("Timeframe Execution", overlay=false)
plot(timeframe.isintraday ? 1.0 : 0.0, "isintraday")
plot(timeframe.isdaily ? 1.0 : 0.0, "isdaily")
`
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScript(t, "timeframe-exec", script)

	/* Default executor runs with -timeframe 1h → IsIntraday=true, IsDaily=false */
	isintraday := exec.ExtractPlotValues(t, output, "isintraday")
	isdaily := exec.ExtractPlotValues(t, output, "isdaily")

	if len(isintraday) < 1 {
		t.Fatal("no data points")
	}
	if isintraday[0] != 1.0 {
		t.Errorf("timeframe.isintraday on default 1h: got %f, want 1.0", isintraday[0])
	}
	if isdaily[0] != 0.0 {
		t.Errorf("timeframe.isdaily on default 1h: got %f, want 0.0", isdaily[0])
	}
}

func TestBuiltinTime_WithCondition(t *testing.T) {
	t.Parallel()
	script := `//@version=5
strategy("Time Condition", overlay=true)
threshold = time > 1700036000000
if threshold
    strategy.entry("Long", strategy.long)
if not threshold
    strategy.close("Long")
plot(time, "time_val")
`
	bars := knownTimeBars()
	exec := util.NewPineExecutor(t)
	output := exec.ExecuteScriptWithCustomData(t, "time-condition", script, bars)
	vals := exec.ExtractPlotValues(t, output, "time_val")

	if len(vals) < len(bars) {
		t.Fatalf("Expected %d bars, got %d", len(bars), len(vals))
	}

	/* threshold = 1700036000000ms = 1700036000s → bar index 10 (base + 10*3600) */
	thresholdSec := int64(1700036000)
	for i, bar := range bars {
		barTimeSec := bar["time"].(int64)
		expectedAbove := barTimeSec > thresholdSec
		plotMs := vals[i]
		actualAbove := plotMs > float64(thresholdSec*1000)
		if expectedAbove != actualAbove {
			t.Errorf("bar[%d]: time=%d, expectedAbove=%v, actualAbove=%v", i, barTimeSec, expectedAbove, actualAbove)
		}
	}
}
