package codegen

import "fmt"

// AccessGenerator provides methods to generate code for accessing series values
type AccessGenerator interface {
	GenerateLoopValueAccess(loopVar string) string
	GenerateInitialValueAccess(period int) string
}

// LoopGenerator creates loop structures for iterating over periods
type LoopGenerator struct {
	period   int
	loopVar  string
	accessor AccessGenerator
	needsNaN bool
}

func NewLoopGenerator(period int, accessor AccessGenerator, needsNaN bool) *LoopGenerator {
	return &LoopGenerator{
		period:   period,
		loopVar:  "j",
		accessor: accessor,
		needsNaN: needsNaN,
	}
}

func (l *LoopGenerator) GenerateForwardLoop(indenter *CodeIndenter) string {
	return indenter.Line(fmt.Sprintf("for %s := 0; %s < %d; %s++ {",
		l.loopVar, l.loopVar, l.period, l.loopVar))
}

func (l *LoopGenerator) GenerateBackwardLoop(indenter *CodeIndenter) string {
	return indenter.Line(fmt.Sprintf("for %s := %d-2; %s >= 0; %s-- {",
		l.loopVar, l.period, l.loopVar, l.loopVar))
}

func (l *LoopGenerator) GenerateValueAccess() string {
	return l.accessor.GenerateLoopValueAccess(l.loopVar)
}

func (l *LoopGenerator) RequiresNaNCheck() bool {
	return l.needsNaN
}
