package iife_generators

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen"
)

type HighestGenerator struct {
	namingStrategy SeriesNamer
}

func NewHighestGenerator(namer SeriesNamer) *HighestGenerator {
	return &HighestGenerator{namingStrategy: namer}
}

func (g *HighestGenerator) Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string {
	/* Extract int value for code generation */
	periodInt := 0
	if constPeriod, ok := period.(*codegen.ConstantPeriod); ok {
		periodInt = constPeriod.Value()
	}

	body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(periodInt))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", periodInt-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val > highest { highest = val } }; "
	body += "return highest"

	return codegen.NewIIFECodeBuilder().
		WithWarmupCheck(periodInt).
		WithBody(body).
		Build()
}
