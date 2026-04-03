package codegen

import (
	"fmt"
	"strings"
)

type VarPersistenceEmitter struct {
	iterationVar string
}

func NewVarPersistenceEmitter(guard RuntimeSafetyGuard) *VarPersistenceEmitter {
	return &VarPersistenceEmitter{
		iterationVar: guard.GenerateIterationVariableReference(),
	}
}

func (e *VarPersistenceEmitter) EmitSeriesGuard(indent, varName, initCode string) string {
	var b strings.Builder
	b.WriteString(indent + fmt.Sprintf("if %s == 0 {\n", e.iterationVar))
	b.WriteString(initCode)
	b.WriteString(indent + "} else {\n")
	b.WriteString(indent + "\t" + seriesCarryForward(varName))
	b.WriteString(indent + "}\n")
	return b.String()
}

func (e *VarPersistenceEmitter) EmitStringGuard(indent, initCode string) string {
	var b strings.Builder
	b.WriteString(indent + fmt.Sprintf("if %s == 0 {\n", e.iterationVar))
	b.WriteString(initCode)
	b.WriteString(indent + "}\n")
	return b.String()
}

func (e *VarPersistenceEmitter) EmitTupleGuard(indent string, elementNames []string, initCode string) string {
	var b strings.Builder
	b.WriteString(indent + fmt.Sprintf("if %s == 0 {\n", e.iterationVar))
	b.WriteString(initCode)
	b.WriteString(indent + "} else {\n")
	for _, name := range elementNames {
		b.WriteString(indent + "\t" + seriesCarryForward(name))
	}
	b.WriteString(indent + "}\n")
	return b.String()
}

func seriesCarryForward(varName string) string {
	return fmt.Sprintf("%sSeries.Set(%sSeries.Get(1))\n", varName, varName)
}
