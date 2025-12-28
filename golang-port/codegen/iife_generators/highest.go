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

func (g *HighestGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val > highest { highest = val } }; "
	body += "return highest"

	return codegen.NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}
