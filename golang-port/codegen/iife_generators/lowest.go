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

func (g *LowestGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j >= 0; j-- { ", period-1)
	body += fmt.Sprintf("val := %s; ", accessor.GenerateLoopValueAccess("j"))
	body += "if val < lowest { lowest = val } }; "
	body += "return lowest"

	return codegen.NewIIFECodeBuilder().
		WithWarmupCheck(period).
		WithBody(body).
		Build()
}
