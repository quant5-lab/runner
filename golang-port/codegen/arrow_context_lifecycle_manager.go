package codegen

import "fmt"

/*
ArrowContextLifecycleManager tracks ArrowContext instances across call sites.

Responsibilities:
- Generates unique context variable names per function call
- Tracks instance counts to avoid variable redeclaration
- Enforces ForwardSeriesBuffer paradigm (pre-allocation)

Design:
- SRP: Single responsibility - manage ArrowContext naming/lifecycle
- KISS: Simple counter-based unique name generation
- DRY: Centralizes all ArrowContext instance tracking logic
*/
type ArrowContextLifecycleManager struct {
	instanceCounts map[string]int
}

func NewArrowContextLifecycleManager() *ArrowContextLifecycleManager {
	return &ArrowContextLifecycleManager{
		instanceCounts: make(map[string]int),
	}
}

/* Generates unique ArrowContext variable name per function call to avoid redeclaration */
func (m *ArrowContextLifecycleManager) AllocateContextVariable(funcName string) string {
	m.instanceCounts[funcName]++
	instanceNum := m.instanceCounts[funcName]
	return fmt.Sprintf("arrowCtx_%s_%d", funcName, instanceNum)
}

func (m *ArrowContextLifecycleManager) GetInstanceCount(funcName string) int {
	return m.instanceCounts[funcName]
}

func (m *ArrowContextLifecycleManager) Reset() {
	m.instanceCounts = make(map[string]int)
}
