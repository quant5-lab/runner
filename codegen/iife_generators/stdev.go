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

func (g *STDEVGenerator) Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("stdev", period.AsSeriesNamePart(), sourceHash)

	/* Extract int value for TAIndicatorBuilder */
	periodInt := 0
	if constPeriod, ok := period.(*codegen.ConstantPeriod); ok {
		periodInt = constPeriod.Value()
	}

	builder := codegen.NewTAIndicatorBuilder(
		"ta.stdev",
		varName,
		periodInt,
		accessor,
		false,
	)
	statefulCode := builder.BuildSTDEV()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
