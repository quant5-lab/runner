package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type ChangeGenerator struct {
	namingStrategy SeriesNamer
}

func NewChangeGenerator(namer SeriesNamer) *ChangeGenerator {
	return &ChangeGenerator{namingStrategy: namer}
}

func (g *ChangeGenerator) Generate(accessor AccessGenerator, offset codegen.PeriodExpression, sourceHash string) string {
	warmupPeriod := 1
	offsetExpr := "1"

	if offset.IsConstant() {
		warmupPeriod = offset.AsInt()
		if warmupPeriod <= 0 {
			warmupPeriod = 1
		}
		offsetExpr = offset.AsGoExpr()
	} else {
		warmupPeriod = -1
		offsetExpr = offset.AsIntCast()
	}

	body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(offsetExpr))
	body += "return current - previous"

	builder := codegen.NewIIFECodeBuilder().WithBody(body)

	if warmupPeriod > 0 {
		builder = builder.WithWarmupCheck(warmupPeriod + 1)
	}

	return builder.Build()
}
