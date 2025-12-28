package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type SMAGenerator struct {
	namingStrategy SeriesNamer
}

func NewSMAGenerator(namer SeriesNamer) *SMAGenerator {
	return &SMAGenerator{namingStrategy: namer}
}

func (g *SMAGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("sma", period, sourceHash)

	builder := codegen.NewTAIndicatorBuilder(
		"ta.sma",
		varName,
		period,
		accessor,
		false,
	)
	builder.WithAccumulator(codegen.NewSumAccumulator())
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
