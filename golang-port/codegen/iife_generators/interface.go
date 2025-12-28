package iife_generators

type Generator interface {
	Generate(accessor AccessGenerator, period int, sourceHash string) string
}

type AccessGenerator interface {
	GenerateLoopValueAccess(loopVar string) string
	GenerateInitialValueAccess(period int) string
}
