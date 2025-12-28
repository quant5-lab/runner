package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type STDEVGenerator struct {
	namingStrategy SeriesNamer
}

func NewSTDEVGenerator(namer SeriesNamer) *STDEVGenerator {
	return &STDEVGenerator{namingStrategy: namer}
}

func (g *STDEVGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("stdev", period, sourceHash)

	builder := codegen.NewTAIndicatorBuilder(
		"ta.stdev",
		varName,
		period,
		accessor,
		false,
	)
	statefulCode := builder.BuildSTDEV()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
