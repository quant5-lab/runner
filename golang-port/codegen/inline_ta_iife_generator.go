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
	body := fmt.Sprintf("alpha := 2.0 / float64(%d+1); ", period)
	body += fmt.Sprintf("ema := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-2)
	body += fmt.Sprintf("ema = alpha*%s + (1-alpha)*ema }; ", accessor.GenerateLoopValueAccess("j"))
	body += "return ema"

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}

type RMAIIFEGenerator struct{}

func (g *RMAIIFEGenerator) Generate(accessor AccessGenerator, period int) string {
	body := fmt.Sprintf("alpha := 1.0 / %d.0; ", period)
	body += "sum := 0.0; "
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-1)
	body += fmt.Sprintf("sum += %s }; ", accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("sma := sum / %d.0; ", period)
	body += fmt.Sprintf("rma := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-2)
	body += fmt.Sprintf("rma = alpha*%s + (1-alpha)*rma }; ", accessor.GenerateLoopValueAccess("j"))
	body += "return rma"

	return NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
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
