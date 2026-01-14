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

func (g *WMAGenerator) Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string {
	context := codegen.NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("wma", period.AsSeriesNamePart(), sourceHash)

	/* Extract int value for TAIndicatorBuilder */
	periodInt := 0
	if constPeriod, ok := period.(*codegen.ConstantPeriod); ok {
		periodInt = constPeriod.Value()
	}

	builder := codegen.NewTAIndicatorBuilder(
		"ta.wma",
		varName,
		periodInt,
		accessor,
		false,
	)
	builder.WithAccumulator(codegen.NewWeightedSumAccumulator(periodInt))
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}
