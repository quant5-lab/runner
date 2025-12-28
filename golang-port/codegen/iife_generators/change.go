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

func (g *ChangeGenerator) Generate(accessor AccessGenerator, offset int, sourceHash string) string {
	if offset <= 0 {
		offset = 1
	}

	body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset)))
	body += "return current - previous"

	return codegen.NewIIFECodeBuilder().
		WithWarmupCheck(offset + 1).
		WithBody(body).
		Build()
}
