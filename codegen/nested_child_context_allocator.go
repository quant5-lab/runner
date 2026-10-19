package codegen

import "fmt"

// NestedChildContextAllocator assigns stable "funcName_N" keys to nested UDF call sites
// within one arrow body compilation pass. Reset() must be called at each new arrow body.
type NestedChildContextAllocator struct {
	callCounts map[string]int
}

func NewNestedChildContextAllocator() *NestedChildContextAllocator {
	return &NestedChildContextAllocator{callCounts: make(map[string]int)}
}

func (a *NestedChildContextAllocator) Reset() {
	a.callCounts = make(map[string]int)
}

// AllocateChildContextExpr returns a GetOrCreateChildContext call with a deterministic key.
func (a *NestedChildContextAllocator) AllocateChildContextExpr(funcName string) string {
	a.callCounts[funcName]++
	key := fmt.Sprintf("%s_%d", funcName, a.callCounts[funcName])
	return fmt.Sprintf("arrowCtx.GetOrCreateChildContext(%q)", key)
}
