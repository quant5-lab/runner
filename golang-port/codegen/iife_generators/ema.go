package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type EMAGenerator struct {
	namingStrategy SeriesNamer
}

func NewEMAGenerator(namer SeriesNamer) *EMAGenerator {
	return &EMAGenerator{namingStrategy: namer}
}

func (g *EMAGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("ema", period, sourceHash)

	builder := codegen.NewStatefulIndicatorBuilder(
		"ta.ema",
		varName,
		period,
		accessor,
		false,
		context,
	)

	statefulCode := builder.BuildEMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
