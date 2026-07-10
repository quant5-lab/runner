package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
Input values are compile-time constants (don't change per bar).
Exception: input.source returns a runtime series reference.
*/
type InputHandler struct {
	inputConstants map[string]string
	argParser      *ArgumentParser
	colorResolver  *ColorConstantResolver
}

func NewInputHandler() *InputHandler {
	return &InputHandler{
		inputConstants: make(map[string]string),
		argParser:      NewArgumentParser(),
		colorResolver:  NewColorConstantResolver(),
	}
}

func (ih *InputHandler) DetectInputFunction(call *ast.CallExpression) bool {
	return IsInputFuncName(extractFunctionNameFromCall(call))
}

func (ih *InputHandler) GenerateInputFloat(call *ast.CallExpression, varName string) (string, error) {
	defval := 0.0

	if len(call.Arguments) > 0 {
		result := ih.argParser.ParseFloat(call.Arguments[0])
		if result.IsValid {
			defval = result.MustBeFloat()
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			defval = ih.extractFloatFromObject(obj, "defval", 0.0)
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %.2f\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

func (ih *InputHandler) GenerateInputInt(call *ast.CallExpression, varName string) (string, error) {
	defval := 0

	if len(call.Arguments) > 0 {
		result := ih.argParser.ParseInt(call.Arguments[0])
		if result.IsValid {
			defval = result.MustBeInt()
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			if ms, ok := extractTimestampFromDefval(obj); ok {
				sanitizedName := SanitizeGoIdentifier(varName)
				code := fmt.Sprintf("const %s = %d\n", sanitizedName, ms)
				ih.inputConstants[varName] = code
				return code, nil
			}
			defval = int(ih.extractFloatFromObject(obj, "defval", 0.0))
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %d\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

func (ih *InputHandler) GenerateInputBool(call *ast.CallExpression, varName string) (string, error) {
	defval := false

	if len(call.Arguments) > 0 {
		result := ih.argParser.ParseBool(call.Arguments[0])
		if result.IsValid {
			defval = result.MustBeBool()
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			defval = ih.extractBoolFromObject(obj, "defval", false)
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %t\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

func (ih *InputHandler) GenerateInputString(call *ast.CallExpression, varName string) (string, error) {
	defval := ""

	if len(call.Arguments) > 0 {
		result := ih.argParser.ParseString(call.Arguments[0])
		if result.IsValid {
			defval = result.MustBeString()
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			defval = ih.extractStringFromObject(obj, "defval", "")
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %q\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

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

/* Session format: "HHMM-HHMM" (e.g., "0950-1345") */
func (ih *InputHandler) GenerateInputSession(call *ast.CallExpression, varName string) (string, error) {
	defval := "0000-2359"

	if len(call.Arguments) > 0 {
		result := ih.argParser.ParseString(call.Arguments[0])
		if result.IsValid {
			defval = result.MustBeString()
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			defval = ih.extractStringFromObject(obj, "defval", "0000-2359")
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %q\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

func (ih *InputHandler) GenerateInputSource(call *ast.CallExpression, varName string) (string, error) {
	source := "close"
	if len(call.Arguments) > 0 {
		if id, ok := call.Arguments[0].(*ast.Identifier); ok {
			source = id.Name
		}
	}
	sanitizedName := SanitizeGoIdentifier(varName)
	return fmt.Sprintf("// %s = input.source(defval=%s) - using source directly\n", sanitizedName, source), nil
}

func (ih *InputHandler) GenerateInputColor(call *ast.CallExpression, varName string) (string, error) {
	defval := ""

	if len(call.Arguments) > 0 {
		if resolved, ok := ih.colorResolver.ResolveExpression(call.Arguments[0]); ok {
			defval = resolved
		} else if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
			defval = ih.extractColorFromObject(obj, "defval")
		}
	}

	sanitizedName := SanitizeGoIdentifier(varName)
	code := fmt.Sprintf("const %s = %q\n", sanitizedName, defval)
	ih.inputConstants[varName] = code
	return code, nil
}

func (ih *InputHandler) extractColorFromObject(obj *ast.ObjectExpression, key string) string {
	parser := NewPropertyParser()
	if expr, ok := parser.ParseExpression(obj, key); ok {
		if resolved, ok := ih.colorResolver.ResolveExpression(expr); ok {
			return resolved
		}
	}
	return ""
}

/* Converts input constants to float64 map for security evaluator */
func (ih *InputHandler) GetInputConstantsMap() map[string]float64 {
	result := make(map[string]float64)
	for varName, code := range ih.inputConstants {
		var floatVal float64
		var intVal int
		var boolVal bool
		if _, err := fmt.Sscanf(code, "const "+varName+" = %f", &floatVal); err == nil {
			result[varName] = floatVal
		} else if _, err := fmt.Sscanf(code, "const "+varName+" = %d", &intVal); err == nil {
			result[varName] = float64(intVal)
		} else if _, err := fmt.Sscanf(code, "const "+varName+" = %t", &boolVal); err == nil {
			if boolVal {
				result[varName] = 1.0
			} else {
				result[varName] = 0.0
			}
		}
	}
	return result
}

func (ih *InputHandler) IsInputConstant(varName string) bool {
	_, exists := ih.inputConstants[varName]
	return exists
}

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
