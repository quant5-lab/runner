package codegen

/*
VariableRegistryGuard protects arrow function type registrations from being
overwritten during multi-phase code generation.

Problem: Arrow functions are registered as "function" type in Phase 2, but
Phase 3 statement generation may re-infer types from call expressions and
overwrite the registration with result types (e.g., "float64").

Solution: Guard checks if a variable is already registered as "function" and
prevents type changes that would break user-defined function detection.
*/
type VariableRegistryGuard struct {
	registry map[string]string
}

func NewVariableRegistryGuard(registry map[string]string) *VariableRegistryGuard {
	return &VariableRegistryGuard{
		registry: registry,
	}
}

func (g *VariableRegistryGuard) ShouldPreserveExistingType(varName string, newType string) bool {
	existingType, exists := g.registry[varName]
	return exists && existingType == "function" && newType != "function"
}

func (g *VariableRegistryGuard) SafeRegister(varName string, varType string) bool {
	if g.ShouldPreserveExistingType(varName, varType) {
		return false
	}
	g.registry[varName] = varType
	return true
}
