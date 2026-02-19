package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* IIFE generators for window-based functions that have top-level handlers but no arrow-context support.
 * Each implements InlineTAIIFEGenerator for use inside arrow function bodies. */

type MaxIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type MinIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RangeIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type VarianceIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type DevIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type MedianIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type ModeIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *MaxIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("maxVal := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("for j := 1; j < %s; j++ { v := %s; if !math.IsNaN(v) && v > maxVal { maxVal = v } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return maxVal"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *MinIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("minVal := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("for j := 1; j < %s; j++ { v := %s; if !math.IsNaN(v) && v < minVal { minVal = v } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return minVal"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *RangeIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("first := %s; maxVal := first; minVal := first; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("for j := 1; j < %s; j++ { v := %s; if !math.IsNaN(v) { if v > maxVal { maxVal = v }; if v < minVal { minVal = v } } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return maxVal - minVal"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *VarianceIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %s; j++ { sum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("mean := sum / %s; ", period.AsFloat64Cast())
	body += fmt.Sprintf("varianceAcc := 0.0; for j := 0; j < %s; j++ { diff := %s - mean; varianceAcc += diff * diff }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return varianceAcc / %s", period.AsFloat64Cast())

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *DevIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("devSum := 0.0; for j := 0; j < %s; j++ { devSum += %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("devMean := devSum / %s; ", period.AsFloat64Cast())
	body += fmt.Sprintf("devAcc := 0.0; for j := 0; j < %s; j++ { devAcc += math.Abs(%s - devMean) }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return devAcc / %s", period.AsFloat64Cast())

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *MedianIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("medVals := make([]float64, 0, %s); ", period.AsIntCast())
	body += fmt.Sprintf("for j := 0; j < %s; j++ { v := %s; if !math.IsNaN(v) { medVals = append(medVals, v) } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "if len(medVals) == 0 { return math.NaN() }; "
	body += "sort.Float64s(medVals); n := len(medVals); if n%2 == 0 { return (medVals[n/2-1] + medVals[n/2]) / 2.0 } else { return medVals[n/2] }"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *ModeIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("modeFreq := make(map[float64]int); for j := 0; j < %s; j++ { v := %s; if !math.IsNaN(v) { modeFreq[v]++ } }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "maxF := 0; modeVal := 0.0; for v, f := range modeFreq { if f > maxF || (f == maxF && v > modeVal) { maxF = f; modeVal = v } }; "
	body += "return modeVal"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}
