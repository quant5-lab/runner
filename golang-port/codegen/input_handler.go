package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
InputHandler manages Pine Script input.* function code generation.

Design: Input values are compile-time constants (don't change per bar).
Exception: input.source returns a runtime series reference.
Rationale: Aligns with Pine Script's input semantics.
*/
type InputHandler struct {
	inputConstants map[string]string // varName -> constant value
}

func NewInputHandler() *InputHandler {
	return &InputHandler{
		inputConstants: make(map[string]string),
	}
}

/*
DetectInputFunction checks if a call expression is an input.* function.
*/
func (ih *InputHandler) DetectInputFunction(call *ast.CallExpression) bool {
	funcName := extractFunctionNameFromCall(call)
	return funcName == "input.float" || funcName == "input.int" ||
		funcName == "input.bool" || funcName == "input.string" ||
		funcName == "input.source"
}

/*
GenerateInputFloat generates code for input.float(defval, title, ...).
Extracts defval from positional OR named parameter.
Returns const declaration.
*/
func (ih *InputHandler) GenerateInputFloat(call *ast.CallExpression, varName string) (string, error) {
	defval := 0.0

	// Try positional argument first
	if len(call.Arguments) > 0 {
		if lit, ok := call.Arguments[0].(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				defval = val
			}
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			// Named parameters in first argument
			defval = ih.extractFloatFromObject(obj, "defval", 0.0)
		}
	}

	code := fmt.Sprintf("const %s = %.2f\n", varName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

/*
GenerateInputInt generates code for input.int(defval, title, ...).
Extracts defval from positional OR named parameter.
Returns const declaration.
*/
func (ih *InputHandler) GenerateInputInt(call *ast.CallExpression, varName string) (string, error) {
	defval := 0

	// Try positional argument first
	if len(call.Arguments) > 0 {
		if lit, ok := call.Arguments[0].(*ast.Literal); ok {
			if val, ok := lit.Value.(float64); ok {
				defval = int(val)
			}
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			// Named parameters in first argument
			defval = int(ih.extractFloatFromObject(obj, "defval", 0.0))
		}
	}

	code := fmt.Sprintf("const %s = %d\n", varName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

/*
GenerateInputBool generates code for input.bool(defval, title, ...).
Extracts defval from positional OR named parameter.
Returns const declaration.
*/
func (ih *InputHandler) GenerateInputBool(call *ast.CallExpression, varName string) (string, error) {
	defval := false

	// Try positional argument first
	if len(call.Arguments) > 0 {
		if lit, ok := call.Arguments[0].(*ast.Literal); ok {
			if val, ok := lit.Value.(bool); ok {
				defval = val
			}
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			// Named parameters in first argument
			defval = ih.extractBoolFromObject(obj, "defval", false)
		}
	}

	code := fmt.Sprintf("const %s = %t\n", varName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

/*
GenerateInputString generates code for input.string(defval, title, ...).
Extracts defval from positional OR named parameter.
Returns const declaration.
*/
func (ih *InputHandler) GenerateInputString(call *ast.CallExpression, varName string) (string, error) {
	defval := ""

	// Try positional argument first
	if len(call.Arguments) > 0 {
		if lit, ok := call.Arguments[0].(*ast.Literal); ok {
			if val, ok := lit.Value.(string); ok {
				defval = val
			}
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			// Named parameters in first argument
			defval = ih.extractStringFromObject(obj, "defval", "")
		}
	}

	code := fmt.Sprintf("const %s = %q\n", varName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

/* Helper: extract float from ObjectExpression property */
func (ih *InputHandler) extractFloatFromObject(obj *ast.ObjectExpression, key string, defaultVal float64) float64 {
	parser := NewPropertyParser()
	if val, ok := parser.ParseFloat(obj, key); ok {
		return val
	}
	return defaultVal
}

func (ih *InputHandler) extractBoolFromObject(obj *ast.ObjectExpression, key string, defaultVal bool) bool {
	parser := NewPropertyParser()
	if val, ok := parser.ParseBool(obj, key); ok {
		return val
	}
	return defaultVal
}

func (ih *InputHandler) extractStringFromObject(obj *ast.ObjectExpression, key string, defaultVal string) string {
	parser := NewPropertyParser()
	if val, ok := parser.ParseString(obj, key); ok {
		return val
	}
	return defaultVal
}

func (ih *InputHandler) GenerateInputSource(call *ast.CallExpression, varName string) (string, error) {
	source := "close"
	if len(call.Arguments) > 0 {
		if id, ok := call.Arguments[0].(*ast.Identifier); ok {
			source = id.Name
		}
	}
	return fmt.Sprintf("// %s = input.source(defval=%s) - using source directly\n", varName, source), nil
}

/* Helper function to extract function name from CallExpression */
func extractFunctionNameFromCall(call *ast.CallExpression) string {
	if member, ok := call.Callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				return obj.Name + "." + prop.Name
			}
		}
	}
	if id, ok := call.Callee.(*ast.Identifier); ok {
		return id.Name
	}
	return ""
}
