package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type LowestGenerator struct {
	namingStrategy SeriesNamer
}

func NewLowestGenerator(namer SeriesNamer) *LowestGenerator {
	return &LowestGenerator{namingStrategy: namer}
}

func (g *LowestGenerator) Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string {
	/* Extract int value for code generation */
	periodInt := 0
	if constPeriod, ok := period.(*codegen.ConstantPeriod); ok {
		periodInt = constPeriod.Value()
	}

	body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", periodInt-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val < lowest { lowest = val } }; "
	body += "return lowest"

	return codegen.NewIIFECodeBuilder().
		WithWarmupCheck(periodInt).
		WithBody(body).
		Build()
}
