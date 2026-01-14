package codegen

import (
	"fmt"
)

// ConstantRegistry manages Pine input constants (input.float, input.int, input.bool, input.string).
// Single source of truth for constant values during code generation.
type ConstantRegistry struct {
	constants map[string]interface{}
}

func NewConstantRegistry() *ConstantRegistry {
	return &ConstantRegistry{
		constants: make(map[string]interface{}),
	}
}

func (cr *ConstantRegistry) Register(name string, value interface{}) {
	cr.constants[name] = value
}

func (cr *ConstantRegistry) Get(name string) (interface{}, bool) {
	val, exists := cr.constants[name]
	return val, exists
}

func (cr *ConstantRegistry) IsConstant(name string) bool {
	_, exists := cr.constants[name]
	return exists
}

func (cr *ConstantRegistry) IsBoolConstant(name string) bool {
	if val, exists := cr.constants[name]; exists {
		_, isBool := val.(bool)
		return isBool
	}
	return false
}

// ExtractFromGeneratedCode parses const declaration: "const name = value\n"
func (cr *ConstantRegistry) ExtractFromGeneratedCode(code string) interface{} {
	var varName string
	var floatVal float64
	var intVal int
	var boolVal bool

	if _, err := fmt.Sscanf(code, "const %s = %f", &varName, &floatVal); err == nil {
		return floatVal
	}
	if _, err := fmt.Sscanf(code, "const %s = %d", &varName, &intVal); err == nil {
		return intVal
	}
	if _, err := fmt.Sscanf(code, "const %s = %t", &varName, &boolVal); err == nil {
		return boolVal
	}
	return nil
}

func (cr *ConstantRegistry) GetAll() map[string]interface{} {
	return cr.constants
}

func (cr *ConstantRegistry) Count() int {
	return len(cr.constants)
}
