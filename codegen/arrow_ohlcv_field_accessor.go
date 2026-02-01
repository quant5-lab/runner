package codegen

import "fmt"

/*
ArrowOHLCVFieldAccessGenerator generates ctx.Data[ctx.BarIndex-*].Field access code for arrow functions.
Unlike OHLCVFieldAccessGenerator (which uses highSeries.Get() for main scope),
this uses ctx.Data[*].* pattern because arrow functions don't have access to main scope Series variables.
*/
type ArrowOHLCVFieldAccessGenerator struct {
	fieldName  string
	baseOffset int
}

func NewArrowOHLCVFieldAccessGenerator(fieldName string) *ArrowOHLCVFieldAccessGenerator {
	return &ArrowOHLCVFieldAccessGenerator{
		fieldName:  fieldName,
		baseOffset: 0,
	}
}

func NewArrowOHLCVFieldAccessGeneratorWithOffset(fieldName string, baseOffset int) *ArrowOHLCVFieldAccessGenerator {
	return &ArrowOHLCVFieldAccessGenerator{
		fieldName:  fieldName,
		baseOffset: baseOffset,
	}
}

func (g *ArrowOHLCVFieldAccessGenerator) GenerateInitialValueAccess(period int) string {
	totalOffset := period - 1 + g.baseOffset
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", totalOffset, g.fieldName)
}

func (g *ArrowOHLCVFieldAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("ctx.Data[ctx.BarIndex-%s].%s", loopVar, g.fieldName)
	}
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-(%s+%d)].%s", loopVar, g.baseOffset, g.fieldName)
}

func (g *ArrowOHLCVFieldAccessGenerator) GenerateCurrentValueAccess() string {
	if g.baseOffset == 0 {
		return fmt.Sprintf("ctx.Data[ctx.BarIndex].%s", g.fieldName)
	}
	return fmt.Sprintf("ctx.Data[ctx.BarIndex-%d].%s", g.baseOffset, g.fieldName)
}

func (g *ArrowOHLCVFieldAccessGenerator) GetBaseOffset() int {
	return g.baseOffset
}
