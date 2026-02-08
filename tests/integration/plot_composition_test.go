package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

/* TestPlotComposition validates inline TA composition patterns in plot() expressions */
func TestPlotComposition_NzWithTA(t *testing.T) {
	pineScript := `//@version=5
indicator("Plot nz(ta.sma())")
plot(nz(ta.sma(close, 14), 0))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "plot-nz-ta", pineScript)

	if !strings.Contains(code, "series.NewSeries") {
		t.Error("Generated code should contain Series initialization")
	}
	if !strings.Contains(code, "value.Nz") {
		t.Error("Generated code should contain value.Nz call")
	}
	if !strings.Contains(code, ".Get(0)") && !strings.Contains(code, ".GetCurrent()") {
		t.Error("Generated code should access hoisted Series variable")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Generated code should compile: %v", err)
	}
}

/* TestPlotComposition_FixnanWithTA tests inline fixnan with TA in plot() */
func TestPlotComposition_FixnanWithTA(t *testing.T) {
	pineScript := `//@version=5
indicator("Plot fixnan(ta.sma())")
plot(fixnan(ta.sma(close, 14)))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "plot-fixnan-ta", pineScript)

	if !strings.Contains(code, "series.NewSeries") {
		t.Error("Generated code should contain Series initialization")
	}
	if !strings.Contains(code, "fixnanState_") {
		t.Error("Generated code should contain fixnan state variable")
	}
	if !strings.Contains(code, ".Get(0)") || !strings.Contains(code, ".GetCurrent()") {
		t.Error("Generated code should access hoisted Series variables")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Generated code should compile: %v", err)
	}
}

func TestPlotComposition_BinaryTAExpression(t *testing.T) {
	pineScript := `//@version=5
indicator("Plot ta.sma() + ta.ema()")
plot(ta.sma(close, 14) + ta.ema(close, 14))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "plot-binary-ta", pineScript)

	if strings.Count(code, "series.NewSeries") < 2 {
		t.Error("Generated code should contain at least 2 Series initializations for sma and ema")
	}
	if !strings.Contains(code, "+") {
		t.Error("Generated code should contain binary addition")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Generated code should compile: %v", err)
	}
}

func TestPlotComposition_TernaryWithTA(t *testing.T) {
	pineScript := `//@version=5
indicator("Plot ternary with TA branches")
plot(close > open ? ta.sma(close, 20) : ta.ema(close, 20))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "plot-ternary-ta", pineScript)

	if strings.Count(code, "series.NewSeries") < 2 {
		t.Error("Generated code should contain at least 2 Series for both ternary branches")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Generated code should compile: %v", err)
	}
}

/* TestPlotComposition_NestedValueFunctions tests nz(fixnan(ta.sma())) */
func TestPlotComposition_NestedValueFunctions(t *testing.T) {
	pineScript := `//@version=5
indicator("Plot nz(fixnan(ta.sma()))")
plot(nz(fixnan(ta.sma(close, 14)), 0))
`
	exec := util.NewPineExecutor(t)
	code, _ := exec.GenerateCode(t, "plot-nested-value", pineScript)

	if !strings.Contains(code, "series.NewSeries") {
		t.Error("Generated code should contain Series initialization")
	}
	if !strings.Contains(code, "fixnanState_") {
		t.Error("Generated code should contain fixnan state variable")
	}
	if !strings.Contains(code, "value.Nz") {
		t.Error("Generated code should contain value.Nz for outer nz() wrapper")
	}

	err := exec.CompileCode(t, code)
	if err != nil {
		t.Fatalf("Generated code should compile: %v", err)
	}
}

/* TestPlotComposition_Execution validates runtime behavior with actual data */
func TestPlotComposition_Execution_NzWithTA(t *testing.T) {
	pineScript := `//@version=5
indicator("Execution: nz(ta.sma())")
plot(nz(ta.sma(close, 5), 0), title="SMA or Zero")
`
	baseTime := int64(1700000000)
	prices := []float64{100, 102, 98, 105, 103, 101, 107, 110, 108, 112}

	testData := []map[string]interface{}{}
	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "exec-nz-ta", pineScript, testData)

	if len(result.Plots) == 0 {
		t.Fatal("Expected at least one plot")
	}

	foundPlot := false
	for _, plot := range result.Plots {
		if plot.Title == "SMA or Zero" {
			foundPlot = true
			if len(plot.Data) != len(prices) {
				t.Errorf("Expected %d plot points, got %d", len(prices), len(plot.Data))
			}
		}
	}

	if !foundPlot {
		t.Error("Expected plot with title 'SMA or Zero'")
	}
}

func TestPlotComposition_Execution_BinaryTA(t *testing.T) {
	pineScript := `//@version=5
indicator("Execution: sma + ema")
plot(ta.sma(close, 3) + ta.ema(close, 3), title="SMA+EMA")
`
	baseTime := int64(1700000000)
	prices := []float64{100, 102, 98, 105, 103, 101, 107, 110}

	testData := []map[string]interface{}{}
	for i, price := range prices {
		testData = append(testData, map[string]interface{}{
			"time":   baseTime + int64(i*3600),
			"open":   price - 1.0,
			"high":   price + 2.0,
			"low":    price - 2.0,
			"close":  price,
			"volume": 1000.0,
		})
	}

	exec := util.NewPineExecutor(t)
	result := exec.ExecuteScriptWithCustomData(t, "exec-binary-ta", pineScript, testData)

	if len(result.Plots) == 0 {
		t.Fatal("Expected at least one plot")
	}

	foundPlot := false
	for _, plot := range result.Plots {
		if plot.Title == "SMA+EMA" {
			foundPlot = true
			hasNonZero := false
			for _, pt := range plot.Data {
				if pt.Value != 0 {
					hasNonZero = true
					break
				}
			}
			if !hasNonZero {
				t.Error("Expected non-zero values after warmup period")
			}
		}
	}

	if !foundPlot {
		t.Error("Expected plot with title 'SMA+EMA'")
	}
}
