package codegen

type RuntimeOnlyFunctionFilter struct {
	runtimeOnlyFunctions map[string]bool
}

func NewRuntimeOnlyFunctionFilter() *RuntimeOnlyFunctionFilter {
	return &RuntimeOnlyFunctionFilter{
		runtimeOnlyFunctions: map[string]bool{
			"fixnan": true,
		},
	}
}

func (f *RuntimeOnlyFunctionFilter) IsRuntimeOnly(funcName string) bool {
	return f.runtimeOnlyFunctions[funcName]
}
