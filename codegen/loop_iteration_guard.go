package codegen

import "fmt"

const MaxWhileIterations = 100000

const whileGuardVar = "__whileGuard"

type LoopIterationGuard struct {
	maxIterations int
}

func NewLoopIterationGuard() *LoopIterationGuard {
	return &LoopIterationGuard{maxIterations: MaxWhileIterations}
}

func (g *LoopIterationGuard) InitCode(indent string) string {
	return fmt.Sprintf("%s%s := 0\n", indent, whileGuardVar)
}

func (g *LoopIterationGuard) CheckCode(indent string) string {
	return fmt.Sprintf("%sif %s++; %s > %d {\n%s\tbreak\n%s}\n",
		indent, whileGuardVar, whileGuardVar, g.maxIterations, indent, indent)
}
