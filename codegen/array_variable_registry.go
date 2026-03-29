package codegen

type ArrayVariableRegistry struct {
	variables map[string]ArrayElementType
}

func NewArrayVariableRegistry() *ArrayVariableRegistry {
	return &ArrayVariableRegistry{
		variables: make(map[string]ArrayElementType),
	}
}

func (r *ArrayVariableRegistry) Register(varName string, elemType ArrayElementType) {
	r.variables[varName] = elemType
}

func (r *ArrayVariableRegistry) Lookup(varName string) (ArrayElementType, bool) {
	elemType, exists := r.variables[varName]
	return elemType, exists
}

func (r *ArrayVariableRegistry) IsArrayVariable(varName string) bool {
	_, exists := r.variables[varName]
	return exists
}
