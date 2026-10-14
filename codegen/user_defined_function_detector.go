package codegen

import "github.com/quant5-lab/runner/ast"

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

func (d *UserDefinedFunctionDetector) IsUserDefinedCall(expr ast.Expression) bool {
	call, ok := expr.(*ast.CallExpression)
	if !ok || call == nil {
		return false
	}
	return d.IsUserDefinedFunction(extractCallFunctionName(call))
}
