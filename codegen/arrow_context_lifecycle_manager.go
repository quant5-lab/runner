package codegen

import "fmt"

/*
ArrowContextLifecycleManager tracks ArrowContext instances and hoisting state.

Design (SRP): Single responsibility - naming and lifecycle tracking only
*/
type ArrowContextLifecycleManager struct {
	instanceCounts  map[string]int
	hoistedContexts map[string]bool
}

func NewArrowContextLifecycleManager() *ArrowContextLifecycleManager {
	return &ArrowContextLifecycleManager{
		instanceCounts:  make(map[string]int),
		hoistedContexts: make(map[string]bool),
	}
}

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
	m.hoistedContexts = make(map[string]bool)
}

func (m *ArrowContextLifecycleManager) MarkAsHoisted(contextVar string) {
	m.hoistedContexts[contextVar] = true
}

func (m *ArrowContextLifecycleManager) IsHoisted(contextVar string) bool {
	return m.hoistedContexts[contextVar]
}
