package codegen

/* UserDefinedFunctionDetector identifies user-defined arrow functions in variables registry */
type UserDefinedFunctionDetector struct {
	variablesRegistry map[string]string
}

func NewUserDefinedFunctionDetector(variablesRegistry map[string]string) *UserDefinedFunctionDetector {
	return &UserDefinedFunctionDetector{
		variablesRegistry: variablesRegistry,
	}
}

func (d *UserDefinedFunctionDetector) IsUserDefinedFunction(funcName string) bool {
	varType, exists := d.variablesRegistry[funcName]
	return exists && varType == "function"
}
