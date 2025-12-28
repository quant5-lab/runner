package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type RMAGenerator struct {
	namingStrategy SeriesNamer
}

func NewRMAGenerator(namer SeriesNamer) *RMAGenerator {
	return &RMAGenerator{namingStrategy: namer}
}

func (g *RMAGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("rma", period, sourceHash)

	builder := codegen.NewStatefulIndicatorBuilder(
		"ta.rma",
		varName,
		period,
		accessor,
		false,
		context,
	)

	statefulCode := builder.BuildRMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

type SeriesNamer interface {
	GenerateName(indicatorType string, period int, sourceHash string) string
}
