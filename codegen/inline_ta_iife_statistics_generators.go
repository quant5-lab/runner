package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

// PercentrankIIFEGenerator generates inline IIFE code for ta.percentrank in arrow function context.
type PercentrankIIFEGenerator struct{ namingStrategy series_naming.Strategy }

// PercentileNearestRankIIFEGenerator generates inline IIFE code for ta.percentile_nearest_rank.
type PercentileNearestRankIIFEGenerator struct {
	namingStrategy series_naming.Strategy
	pct            float64
}

// PercentileLinearInterpolationIIFEGenerator generates inline IIFE code for ta.percentile_linear_interpolation.
type PercentileLinearInterpolationIIFEGenerator struct {
	namingStrategy series_naming.Strategy
	pct            float64
}

// CorrelationIIFEGenerator generates inline IIFE code for ta.correlation.
type CorrelationIIFEGenerator struct {
	namingStrategy series_naming.Strategy
	src2Accessor   AccessGenerator
}

func (g *PercentrankIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("cur := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("count := 0; for j := 0; j < %s; j++ { if %s < cur { count++ } }; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return float64(count) / %s * 100.0", period.AsFloat64Cast())

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *PercentileNearestRankIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	pct := g.pct

	body := fmt.Sprintf("w := make([]float64, %s); ", period.AsIntCast())
	body += fmt.Sprintf("for j := 0; j < %s; j++ { w[j] = %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "sort.Float64s(w); "
	body += fmt.Sprintf("idx := int(math.Ceil(%g/100.0*%s)) - 1; ", pct, period.AsFloat64Cast())
	body += fmt.Sprintf("if idx < 0 { idx = 0 }; if idx >= %s { idx = %s - 1 }; ", period.AsIntCast(), period.AsIntCast())
	body += "return w[idx]"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *PercentileLinearInterpolationIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	pct := g.pct

	body := fmt.Sprintf("w := make([]float64, %s); ", period.AsIntCast())
	body += fmt.Sprintf("for j := 0; j < %s; j++ { w[j] = %s }; ", period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "sort.Float64s(w); "
	body += fmt.Sprintf("rank := %g / 100.0 * float64(%s-1); ", pct, period.AsIntCast())
	body += "lower := int(math.Floor(rank)); upper := lower + 1; frac := rank - float64(lower); "
	body += fmt.Sprintf("if upper >= %s { return w[%s-1] }; ", period.AsIntCast(), period.AsIntCast())
	body += "return w[lower] + frac*(w[upper]-w[lower])"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *CorrelationIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	if g.src2Accessor == nil {
		return "0.0"
	}
	src2 := g.src2Accessor

	body := fmt.Sprintf("sum1, sum2 := 0.0, 0.0; for j := 0; j < %s; j++ { sum1 += %s; sum2 += %s }; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"), src2.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("m1 := sum1 / %s; m2 := sum2 / %s; ", period.AsFloat64Cast(), period.AsFloat64Cast())
	body += fmt.Sprintf("cov, v1, v2 := 0.0, 0.0, 0.0; for j := 0; j < %s; j++ { ", period.AsIntCast())
	body += fmt.Sprintf("d1 := %s - m1; d2 := %s - m2; cov += d1*d2; v1 += d1*d1; v2 += d2*d2 }; ",
		accessor.GenerateLoopValueAccess("j"), src2.GenerateLoopValueAccess("j"))
	body += "if v1 == 0 || v2 == 0 { return 0.0 }; return cov / math.Sqrt(v1*v2)"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}
