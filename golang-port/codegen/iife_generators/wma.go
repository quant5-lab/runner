package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type WMAGenerator struct {
	namingStrategy SeriesNamer
}

func NewWMAGenerator(namer SeriesNamer) *WMAGenerator {
	return &WMAGenerator{namingStrategy: namer}
}

func (g *WMAGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("wma", period, sourceHash)

	builder := codegen.NewTAIndicatorBuilder(
		"ta.wma",
		varName,
		period,
		accessor,
		false,
	)
	builder.WithAccumulator(codegen.NewWeightedSumAccumulator(period))
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
