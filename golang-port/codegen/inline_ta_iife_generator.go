package codegen

import "fmt"

type InlineTAIIFEGenerator interface {
	Generate(accessor AccessGenerator, period int) string
}

type SMAIIFEGenerator struct{}

func (g *SMAIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := "sum := 0.0; "
	body += fmt.Sprintf("for j := 0; j < %d; j++ { ", period)
	body += fmt.Sprintf("sum += %s }; ", accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return sum / %d.0", period)

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type EMAIIFEGenerator struct{}

func (g *EMAIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	context := NewArrowFunctionIndicatorContext()
	varName := fmt.Sprintf("_ema_%d", period)

	builder := NewStatefulIndicatorBuilder(
		"ta.ema",
		varName,
		period,
		accessor,
		false,
		context,
	)

	statefulCode := builder.BuildEMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0) + ".GetCurrent()"

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

type RMAIIFEGenerator struct{}

func (g *RMAIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	context := NewArrowFunctionIndicatorContext()
	varName := fmt.Sprintf("_rma_%d", period)

	builder := NewStatefulIndicatorBuilder(
		"ta.rma",
		varName,
		period,
		accessor,
		false,
		context,
	)

	statefulCode := builder.BuildRMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0) + ".GetCurrent()"

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

type WMAIIFEGenerator struct{}

func (g *WMAIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := "weightedSum := 0.0; "
	body += fmt.Sprintf("weightSum := %d.0; ", (period*(period+1))/2)
	body += fmt.Sprintf("for j := 0; j < %d; j++ { ", period)
	body += fmt.Sprintf("weight := float64(%d - j); ", period)
	body += fmt.Sprintf("weightedSum += %s * weight }; ", accessor.GenerateLoopValueAccess("j"))
	body += "return weightedSum / weightSum"

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type STDEVIIFEGenerator struct{}

func (g *STDEVIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := "sum := 0.0; "
	body += fmt.Sprintf("for j := 0; j < %d; j++ { ", period)
	body += fmt.Sprintf("sum += %s }; ", accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("mean := sum / %d.0; ", period)
	body += "variance := 0.0; "
	body += fmt.Sprintf("for j := 0; j < %d; j++ { ", period)
	body += fmt.Sprintf("diff := %s - mean; ", accessor.GenerateLoopValueAccess("j"))
	body += "variance += diff * diff }; "
	body += fmt.Sprintf("return math.Sqrt(variance / %d.0)", period)

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type HighestIIFEGenerator struct{}

func (g *HighestIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val > highest { highest = val } }; "
	body += "return highest"

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type LowestIIFEGenerator struct{}

func (g *LowestIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val < lowest { lowest = val } }; "
	body += "return lowest"

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type ChangeIIFEGenerator struct{}

func (g *ChangeIIFEGenerator) Generate(accessor AccessGenerator, offset int) string {
	if offset <= 0 {
		offset = 1
	}

	body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset)))
	body += "return current - previous"

	return NewIIFECodeBuilder().
		WithWarmupCheck(offset + 1).
		WithBody(body).
		Build()
}
