package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestValuewhen_BasicCodegen(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Valuewhen Basic", overlay=true)

bullish = close > open
lastBullishClose = ta.valuewhen(bullish, close, 0)
prevBullishClose = ta.valuewhen(bullish, close, 1)

plot(lastBullishClose, "Last Bullish", color=color.green)
plot(prevBullishClose, "Prev Bullish", color=color.blue)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "valuewhen-basic", pineScript)

	if !strings.Contains(generatedCode, "Inline valuewhen") {
		t.Error("Expected inline valuewhen generation")
	}

	if !strings.Contains(generatedCode, "occurrenceCount") {
		t.Error("Expected occurrenceCount variable in generated code")
	}

	if !strings.Contains(generatedCode, "lookbackOffset") {
		t.Error("Expected lookbackOffset loop variable")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Valuewhen basic codegen test passed")
}

func TestValuewhen_WithSeriesSources(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Valuewhen Series", overlay=true)

sma20 = ta.sma(close, 20)
crossUp = ta.crossover(close, sma20)
crossLevel = ta.valuewhen(crossUp, close, 0)

plot(crossLevel, "Cross Level", color=color.orange)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "valuewhen-series", pineScript)

	if !strings.Contains(generatedCode, "valuewhen") {
		t.Error("Expected valuewhen in generated code")
	}

	if !strings.Contains(generatedCode, "crossUpSeries.Get") {
		t.Error("Expected Series.Get() for condition access")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Valuewhen with series sources test passed")
}

func TestValuewhen_MultipleOccurrences(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Valuewhen Multiple", overlay=true)

signal = close > ta.sma(close, 10)
val0 = ta.valuewhen(signal, high, 0)
val1 = ta.valuewhen(signal, high, 1)
val2 = ta.valuewhen(signal, high, 2)

plot(val0, "Occurrence 0", color=color.red)
plot(val1, "Occurrence 1", color=color.orange)
plot(val2, "Occurrence 2", color=color.yellow)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "valuewhen-multiple", pineScript)

	occurrenceCount := strings.Count(generatedCode, "Inline valuewhen")
	if occurrenceCount != 3 {
		t.Errorf("Expected 3 valuewhen calls, got %d", occurrenceCount)
	}

	if !strings.Contains(generatedCode, "occurrenceCount == 0") {
		t.Error("Expected occurrence 0 check")
	}
	if !strings.Contains(generatedCode, "occurrenceCount == 1") {
		t.Error("Expected occurrence 1 check")
	}
	if !strings.Contains(generatedCode, "occurrenceCount == 2") {
		t.Error("Expected occurrence 2 check")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Valuewhen multiple occurrences test passed")
}

func TestValuewhen_InStrategyContext(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
strategy("Valuewhen Strategy", overlay=true)

buySignal = ta.crossover(close, ta.sma(close, 20))
buyPrice = ta.valuewhen(buySignal, close, 0)

if buySignal
    strategy.entry("Long", strategy.long)

plot(buyPrice, "Buy Price", color=color.green)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "valuewhen-strategy", pineScript)

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Valuewhen in strategy context test passed")
}

func TestValuewhen_ComplexConditions(t *testing.T) {
	t.Parallel()
	pineScript := `//@version=5
indicator("Valuewhen Complex", overlay=true)

sma20 = ta.sma(close, 20)
above = close > sma20
crossUp = ta.crossover(close, sma20)
trigger = above and crossUp

lastTriggerPrice = ta.valuewhen(trigger, low, 0)
plot(lastTriggerPrice, "Trigger Price", color=color.purple)
`

	exec := util.NewPineExecutor(t)
	generatedCode, _ := exec.GenerateCode(t, "valuewhen-complex", pineScript)

	if !strings.Contains(generatedCode, "triggerSeries.Get(lookbackOffset)") {
		t.Error("Expected condition Series.Get() access with lookbackOffset")
	}

	if err := exec.CompileCode(t, generatedCode); err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Log("✓ Valuewhen complex conditions test passed")
}

func TestValuewhen_RegressionStability(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "bar field sources",
			script: `//@version=5
indicator("Bar Fields", overlay=true)
signal = close > open
h = ta.valuewhen(signal, high, 0)
l = ta.valuewhen(signal, low, 0)
plot(h, "High")
plot(l, "Low")
`,
		},
		{
			name: "series expression source",
			script: `//@version=5
indicator("Series Expression", overlay=true)
sma = ta.sma(close, 20)
cross = ta.crossover(close, sma)
level = ta.valuewhen(cross, sma, 0)
plot(level, "Level")
`,
		},
		{
			name: "chained valuewhen",
			script: `//@version=5
indicator("Chained", overlay=true)
sig = close > ta.sma(close, 20)
v0 = ta.valuewhen(sig, close, 0)
v1 = ta.valuewhen(sig, v0, 0)
plot(v1, "Chained")
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			exec := util.NewPineExecutor(t)
			generatedCode, _ := exec.GenerateCode(t, "valuewhen-regression", tt.script)

			if err := exec.CompileCode(t, generatedCode); err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
		})
	}

	t.Log("✓ Valuewhen regression stability tests passed")
}
