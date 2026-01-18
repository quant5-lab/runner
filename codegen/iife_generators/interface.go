package iife_generators

import "github.com/quant5-lab/runner/codegen"

type Generator interface {
	Generate(accessor AccessGenerator, period codegen.PeriodExpression, sourceHash string) string
}

type AccessGenerator interface {
	GenerateLoopValueAccess(loopVar string) string
	GenerateInitialValueAccess(period int) string
	GenerateCurrentValueAccess() string
}
