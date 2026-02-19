package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

/* IIFE generators for momentum and monotone functions — used inside arrow function bodies. */

type RisingIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type FallingIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type HighestbarsIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type LowestbarsIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type MomIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RocIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type CmoIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type WprIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *RisingIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("for j := 0; j < %s; j++ { curr := %s; prev := %s; if curr <= prev { return 0.0 } }; return 1.0",
		period.AsIntCast(),
		accessor.GenerateLoopValueAccess("j"),
		accessor.GenerateLoopValueAccess("j+1"))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *FallingIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("for j := 0; j < %s; j++ { curr := %s; prev := %s; if curr >= prev { return 0.0 } }; return 1.0",
		period.AsIntCast(),
		accessor.GenerateLoopValueAccess("j"),
		accessor.GenerateLoopValueAccess("j+1"))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *HighestbarsIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("extIdx := 0; extVal := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("for j := 1; j < %s; j++ { v := %s; if !math.IsNaN(v) && v > extVal { extVal = v; extIdx = j } }; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return float64(-extIdx)"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *LowestbarsIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("extIdx := 0; extVal := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("for j := 1; j < %s; j++ { v := %s; if !math.IsNaN(v) && v < extVal { extVal = v; extIdx = j } }; ",
		period.AsIntCast(), accessor.GenerateLoopValueAccess("j"))
	body += "return float64(-extIdx)"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *MomIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("return %s - %s",
		accessor.GenerateLoopValueAccess("0"),
		accessor.GenerateLoopValueAccess(period.AsIntCast()))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *RocIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("past := %s; if math.IsNaN(past) || past == 0.0 { return math.NaN() }; ",
		accessor.GenerateLoopValueAccess(period.AsIntCast()))
	body += fmt.Sprintf("return (%s - past) / past * 100.0", accessor.GenerateLoopValueAccess("0"))

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *CmoIIFEGenerator) Generate(accessor AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("up, down := 0.0, 0.0; for j := 0; j < %s; j++ { curr := %s; prev := %s; if !math.IsNaN(curr) && !math.IsNaN(prev) { if curr > prev { up += curr - prev } else { down += prev - curr } } }; ",
		period.AsIntCast(),
		accessor.GenerateLoopValueAccess("j"),
		accessor.GenerateLoopValueAccess("j+1"))
	body += "total := up + down; if total == 0.0 { return 0.0 }; return (up - down) / total * 100.0"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, accessor.GetBaseOffset()).
		WithBody(body).
		Build()
}

func (g *WprIIFEGenerator) Generate(_ AccessGenerator, period PeriodExpression, _ string) string {
	body := fmt.Sprintf("hh := ctx.Data[ctx.BarIndex].High; ll := ctx.Data[ctx.BarIndex].Low; ")
	body += fmt.Sprintf("for j := 1; j < %s; j++ { b := ctx.Data[ctx.BarIndex-j]; if b.High > hh { hh = b.High }; if b.Low < ll { ll = b.Low } }; ", period.AsIntCast())
	body += "d := hh - ll; if d == 0.0 { return 0.0 }; return (ctx.Data[ctx.BarIndex].Close - hh) / d * 100.0"

	return NewIIFECodeBuilder().
		WithWarmupCheckPeriodExpression(period, 0).
		WithBody(body).
		Build()
}
