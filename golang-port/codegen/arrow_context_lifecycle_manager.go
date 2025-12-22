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

/*
AllocateContextVariable generates unique ArrowContext variable name for function call.

Returns: "arrowCtx_<funcName>_<instanceNum>"
Example: "arrowCtx_adx_1", "arrowCtx_adx_2"

Ensures no variable redeclaration within same scope.
*/
func (m *ArrowContextLifecycleManager) AllocateContextVariable(funcName string) string {
	m.instanceCounts[funcName]++
	instanceNum := m.instanceCounts[funcName]
	return fmt.Sprintf("arrowCtx_%s_%d", funcName, instanceNum)
}

/*
GetInstanceCount returns number of allocated contexts for function.
Used for testing and validation.
*/
func (m *ArrowContextLifecycleManager) GetInstanceCount(funcName string) int {
	return m.instanceCounts[funcName]
}

/*
Reset clears all instance counts.
Used between strategy compilations.
*/
func (m *ArrowContextLifecycleManager) Reset() {
	m.instanceCounts = make(map[string]int)
}
