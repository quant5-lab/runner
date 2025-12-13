package codegen

type RuntimeOnlyFunctionFilter struct {
	runtimeOnlyFunctions map[string]bool
}

func NewRuntimeOnlyFunctionFilter() *RuntimeOnlyFunctionFilter {
	return &RuntimeOnlyFunctionFilter{
		runtimeOnlyFunctions: map[string]bool{
			"ta.pivothigh": true,
			"pivothigh":    true,
			"ta.pivotlow":  true,
			"pivotlow":     true,
			"fixnan":       true,
		},
	}
}

func (f *RuntimeOnlyFunctionFilter) IsRuntimeOnly(funcName string) bool {
	return f.runtimeOnlyFunctions[funcName]
}
