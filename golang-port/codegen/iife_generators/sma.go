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

func (g *SMAGenerator) Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("sma", period.AsSeriesNamePart(), sourceHash)

	/* Extract int value for TAIndicatorBuilder */
	periodInt := 0
	if constPeriod, ok := period.(*codegen.ConstantPeriod); ok {
		periodInt = constPeriod.Value()
	}

	builder := codegen.NewTAIndicatorBuilder(
		"ta.sma",
		varName,
		periodInt,
		accessor,
		false,
	)
	builder.WithAccumulator(codegen.NewSumAccumulator())
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
