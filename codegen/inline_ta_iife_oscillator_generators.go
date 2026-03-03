package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* IIFE generators for CCI, BBW, and COG — used inside arrow function bodies. */

type CCIIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type BBWIIFEGenerator struct {
	namingStrategy series_naming.Strategy
	multLiteral    float64
	multExpression string
	useLiteralMult bool
}

type COGIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *CCIIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("sma := 0.0; for j := 0; j < %s; j++ { sma += %s }; sma /= %s; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"), period.AsFloat64Cast())
	body += fmt.Sprintf("dev := 0.0; for j := 0; j < %s; j++ { v := %s; if v > sma { dev += v - sma } else { dev += sma - v } }; dev /= %s; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"), period.AsFloat64Cast())
	body += fmt.Sprintf("if dev == 0.0 { return 0.0 }; return (%s - sma) / (0.015 * dev)",
		accessor.GenerateLoopValueAccess("0"))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *BBWIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("sma := 0.0; for j := 0; j < %s; j++ { sma += %s }; sma /= %s; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"), period.AsFloat64Cast())
	body += fmt.Sprintf("sd := 0.0; for j := 0; j < %s; j++ { d := %s - sma; sd += d * d }; sd = math.Sqrt(sd / %s); ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"), period.AsFloat64Cast())

	multStr := fmt.Sprintf("%g", g.multLiteral)
	if !g.useLiteralMult {
		multStr = g.multExpression
	}

	body += fmt.Sprintf("if sma == 0.0 { return 0.0 }; return 2.0 * %s * sd / sma", multStr)

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *COGIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("num, den := 0.0, 0.0; for j := 0; j < %s; j++ { v := %s; num += v * float64(j+1); den += v }; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "if den == 0.0 { return 0.0 }; return -num / den"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}
