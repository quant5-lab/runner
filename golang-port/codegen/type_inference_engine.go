package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

// TypeInferenceEngine determines variable types from AST expressions.
// Type system: "float64" (default), "bool", "string"
type TypeInferenceEngine struct {
	variables map[string]string
	constants map[string]interface{}
}

func NewTypeInferenceEngine() *TypeInferenceEngine {
	return &TypeInferenceEngine{
		variables: make(map[string]string),
		constants: make(map[string]interface{}),
	}
}

func (te *TypeInferenceEngine) RegisterVariable(name string, varType string) {
	te.variables[name] = varType
}

func (te *TypeInferenceEngine) RegisterConstant(name string, value interface{}) {
	te.constants[name] = value
}

func (te *TypeInferenceEngine) InferType(expr ast.Expression) string {
	if expr == nil {
		return "float64"
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		return te.inferMemberExpressionType(e)
	case *ast.BinaryExpression:
		return te.inferBinaryExpressionType(e)
	case *ast.LogicalExpression:
		return "bool"
	case *ast.UnaryExpression:
		return te.inferUnaryExpressionType(e)
	case *ast.CallExpression:
		return te.inferCallExpressionType(e)
	case *ast.ConditionalExpression:
		return te.InferType(e.Consequent)
	default:
		return "float64"
	}
}

func (te *TypeInferenceEngine) inferMemberExpressionType(e *ast.MemberExpression) string {
	if obj, ok := e.Object.(*ast.Identifier); ok {
		if obj.Name == "syminfo" {
			if prop, ok := e.Property.(*ast.Identifier); ok {
				if prop.Name == "tickerid" {
					return "string"
				}
			}
		}
	}
	return "float64"
}

func (te *TypeInferenceEngine) inferBinaryExpressionType(e *ast.BinaryExpression) string {
	if te.isComparisonOperator(e.Operator) {
		return "bool"
	}
	return "float64"
}

func (te *TypeInferenceEngine) isComparisonOperator(op string) bool {
	return op == ">" || op == "<" || op == ">=" || op == "<=" || op == "==" || op == "!="
}

func (te *TypeInferenceEngine) inferUnaryExpressionType(e *ast.UnaryExpression) string {
	if e.Operator == "not" || e.Operator == "!" {
		return "bool"
	}
	return te.InferType(e.Argument)
}

func (te *TypeInferenceEngine) inferCallExpressionType(e *ast.CallExpression) string {
	funcName := extractFunctionName(e.Callee)

	if funcName == "ta.crossover" || funcName == "ta.crossunder" {
		return "bool"
	}
	if funcName == "input.bool" {
		return "bool"
	}

	return "float64"
}

func (te *TypeInferenceEngine) IsBoolVariable(expr ast.Expression) bool {
	if ident, ok := expr.(*ast.Identifier); ok {
		return te.IsBoolVariableByName(ident.Name)
	}
	return false
}

func (te *TypeInferenceEngine) IsBoolVariableByName(name string) bool {
	varType, exists := te.variables[name]
	return exists && varType == "bool"
}

func (te *TypeInferenceEngine) IsBoolConstant(name string) bool {
	if val, exists := te.constants[name]; exists {
		_, isBool := val.(bool)
		return isBool
	}
	return false
}

func (te *TypeInferenceEngine) GetVariableType(name string) (string, bool) {
	varType, exists := te.variables[name]
	return varType, exists
}

func extractFunctionName(callee ast.Expression) string {
	switch c := callee.(type) {
	case *ast.Identifier:
		return c.Name
	case *ast.MemberExpression:
		if obj, ok := c.Object.(*ast.Identifier); ok {
			if prop, ok := c.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}
	return ""
}
