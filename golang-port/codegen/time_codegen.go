package codegen

import (
	"fmt"
)

type TimeCodeGenerator struct {
	indentation string
}

func NewTimeCodeGenerator(indentation string) *TimeCodeGenerator {
	return &TimeCodeGenerator{indentation: indentation}
}

func (g *TimeCodeGenerator) GenerateNoArguments(varName string) string {
	return g.indentation + fmt.Sprintf("%sSeries.Set(float64(ctx.Data[ctx.BarIndex].Time))\n", varName)
}

func (g *TimeCodeGenerator) GenerateSingleArgument(varName string) string {
	return g.indentation + fmt.Sprintf("%sSeries.Set(float64(ctx.Data[ctx.BarIndex].Time))\n", varName)
}

func (g *TimeCodeGenerator) GenerateWithSession(varName string, session SessionArgument) string {
	if !session.IsValid() {
		return g.generateInvalidSession(varName)
	}

	if session.IsLiteral() {
		return g.generateLiteralSession(varName, session.Value)
	}

	return g.generateVariableSession(varName, session.Value)
}

func (g *TimeCodeGenerator) generateInvalidSession(varName string) string {
	return g.indentation + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
}

func (g *TimeCodeGenerator) generateLiteralSession(varName, sessionValue string) string {
	code := g.indentation + fmt.Sprintf("/* time(timeframe.period, %q) */\n", sessionValue)
	code += g.indentation + fmt.Sprintf("%s_result := session.TimeFunc(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, %q, ctx.Timezone)\n", varName, sessionValue)
	code += g.indentation + fmt.Sprintf("%sSeries.Set(%s_result)\n", varName, varName)
	return code
}

func (g *TimeCodeGenerator) generateVariableSession(varName, sessionValue string) string {
	code := g.indentation + fmt.Sprintf("/* time(timeframe.period, %s) */\n", sessionValue)
	code += g.indentation + fmt.Sprintf("%s_result := session.TimeFunc(ctx.Data[ctx.BarIndex].Time*1000, ctx.Timeframe, %s, ctx.Timezone)\n", varName, sessionValue)
	code += g.indentation + fmt.Sprintf("%sSeries.Set(%s_result)\n", varName, varName)
	return code
}
