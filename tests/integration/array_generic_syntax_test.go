//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/tests/util"
)

func TestArrayGenericSyntax_RewritesToLegacyForm(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Generic Array Syntax", overlay=true)

prices = array.new<float>(0)

if bar_index < 10
    array.push(prices, close)

avgPrice = array.size(prices) > 0 ? array.avg(prices) : close

plot(avgPrice, "Average")

if bar_index == 5 and close > avgPrice
    strategy.entry("Long", strategy.long)
if bar_index == 8
    strategy.close("Long")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-generic-array", pineScript)

	requiredCode := []string{
		"pricesArraySeries",
		"arrayops.NewMutator().Push",
		"arrayops.NewStatistics().Avg",
	}

	for _, code := range requiredCode {
		if !strings.Contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-generic-array", pineScript)
	if output == nil {
		t.Fatal("Script execution returned nil output")
	}

	// Verify script compiles and runs without errors
	if len(output.Plots) == 0 {
		t.Error("Expected at least one plot output")
	}
}

func TestArrayGenericSyntax_MixedWithLegacy(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Mixed Syntax", overlay=true)

legacy = array.new_float(0)
generic = array.new<float>(0)

if bar_index < 5
    array.push(legacy, close)
    array.push(generic, close + 1.0)

legacyAvg = array.size(legacy) > 0 ? array.avg(legacy) : close
genericAvg = array.size(generic) > 0 ? array.avg(generic) : close

plot(legacyAvg, "Legacy Avg")
plot(genericAvg, "Generic Avg")

if bar_index == 3 and genericAvg > legacyAvg
    strategy.entry("Long", strategy.long)
if bar_index == 6
    strategy.close("Long")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-mixed-syntax", pineScript)

	requiredCode := []string{
		"legacyArraySeries",
		"genericArraySeries",
		"arrayops.NewStatistics().Avg",
	}

	for _, code := range requiredCode {
		if !strings.Contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-mixed-syntax", pineScript)
	if output == nil {
		t.Fatal("Script execution returned nil output")
	}

	if len(output.Plots) == 0 {
		t.Error("Expected at least one plot output")
	}
}

func TestArrayGenericSyntax_WithComments(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Commented Generics", overlay=true)

// Generic syntax: array.new<float>
prices = array.new<float>(0)  // Inline comment

if bar_index < 8
    array.push(prices, close)

avgPrice = array.size(prices) > 0 ? array.avg(prices) : close

plot(avgPrice, "Average")

if bar_index == 4 and close > avgPrice
    strategy.entry("Long", strategy.long)
if bar_index == 7
    strategy.close("Long")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-commented-generics", pineScript)

	requiredCode := []string{
		"pricesArraySeries",
		"arrayops.NewStatistics().Avg",
	}

	for _, code := range requiredCode {
		if !strings.Contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-commented-generics", pineScript)
	if output == nil {
		t.Fatal("Script execution returned nil output")
	}

	if len(output.Plots) == 0 {
		t.Error("Expected at least one plot output")
	}
}

func TestArrayGenericSyntax_NestedInExpressions(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Nested Generic Syntax", overlay=true)

prices = array.from(close, open, high, low)
avgPrice = array.avg(prices)

plot(avgPrice, "Average")

if bar_index == 10 and close > avgPrice
    strategy.entry("Long", strategy.long)
if bar_index == 15
    strategy.close("Long")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-nested-generics", pineScript)

	requiredCode := []string{
		"pricesArraySeries",
		"arrayops.NewStatistics().Avg",
	}

	for _, code := range requiredCode {
		if !strings.Contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-nested-generics", pineScript)
	if output == nil {
		t.Fatal("Script execution returned nil output")
	}

	if len(output.Plots) == 0 {
		t.Error("Expected at least one plot output")
	}
}

func TestArrayGenericSyntax_FloatType(t *testing.T) {
	t.Parallel()

	pineScript := `//@version=5
strategy("Float Type", overlay=true)

floatArr = array.new<float>(0)

if bar_index < 6
    array.push(floatArr, close)

avgPrice = array.size(floatArr) > 0 ? array.avg(floatArr) : close

plot(avgPrice, "Average")

if bar_index == 4 and close > avgPrice
    strategy.entry("Long", strategy.long)
if bar_index == 8
    strategy.close("Long")
`

	exec := util.NewPineExecutor(t)
	goCode, _ := exec.GenerateCode(t, "test-float-type", pineScript)

	requiredCode := []string{
		"floatArrArraySeries",
	}

	for _, code := range requiredCode {
		if !strings.Contains(goCode, code) {
			t.Errorf("Generated code missing: %s", code)
		}
	}

	output := exec.ExecuteScript(t, "test-float-type", pineScript)
	if output == nil {
		t.Fatal("Script execution returned nil output")
	}

	if len(output.Plots) == 0 {
		t.Error("Expected at least one plot output")
	}
}
