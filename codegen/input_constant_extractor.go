package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
InputConstantExtractor inlines compile-time constant values from input.* functions.

Pine type system: input.float/int/bool/string/price return "input" qualified constants (not series).
Exception: input.source returns series float, must fall through to series handling.
*/
type InputConstantExtractor struct {
	argParser *ArgumentParser
}

func NewInputConstantExtractor() *InputConstantExtractor {
	return &InputConstantExtractor{
		argParser: NewArgumentParser(),
	}
}

/* ExtractInputConstant returns constant value string or empty for non-input functions. */
func (ice *InputConstantExtractor) ExtractInputConstant(call *ast.CallExpression, funcName string) string {
	if call == nil {
		return ""
	}

	switch funcName {
	case "input.float":
		return ice.extractInputFloatValue(call)
	case "input.int":
		return ice.extractInputIntValue(call)
	case "input.bool":
		return ice.extractInputBoolValue(call)
	case "input.string":
		return ice.extractInputStringValue(call)
	case "input.price":
		return ice.extractInputFloatValue(call)
	default:
		return ""
	}
}

func (ice *InputConstantExtractor) extractInputFloatValue(call *ast.CallExpression) string {
	if len(call.Arguments) == 0 || call.Arguments == nil {
		return "0.0"
	}

	result := ice.argParser.ParseFloat(call.Arguments[0])
	if result.IsValid {
		return fmt.Sprintf("%v", result.MustBeFloat())
	}

	if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
		val := ice.extractFloatFromObject(obj, "defval", 0.0)
		if val == 0.0 {
			return "0.0"
		}
		return fmt.Sprintf("%v", val)
	}

	return "0.0"
}

func (ice *InputConstantExtractor) extractInputIntValue(call *ast.CallExpression) string {
	if len(call.Arguments) == 0 {
		return "0"
	}

	result := ice.argParser.ParseInt(call.Arguments[0])
	if result.IsValid {
		return fmt.Sprintf("%d", result.MustBeInt())
	}

	if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
		val := int(ice.extractFloatFromObject(obj, "defval", 0.0))
		return fmt.Sprintf("%d", val)
	}

	return "0"
}

func (ice *InputConstantExtractor) extractInputBoolValue(call *ast.CallExpression) string {
	if len(call.Arguments) == 0 {
		return "false"
	}

	result := ice.argParser.ParseBool(call.Arguments[0])
	if result.IsValid {
		return fmt.Sprintf("%t", result.MustBeBool())
	}

	if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
		val := ice.extractBoolFromObject(obj, "defval", false)
		return fmt.Sprintf("%t", val)
	}

	return "false"
}

func (ice *InputConstantExtractor) extractInputStringValue(call *ast.CallExpression) string {
	if len(call.Arguments) == 0 {
		return "\"\""
	}

	result := ice.argParser.ParseString(call.Arguments[0])
	if result.IsValid {
		return fmt.Sprintf("\"%s\"", result.MustBeString())
	}

	if obj, ok := call.Arguments[0].(*ast.ObjectExpression); ok {
		val := ice.extractStringFromObject(obj, "defval", "")
		return fmt.Sprintf("\"%s\"", val)
	}

	return "\"\""
}

func (ice *InputConstantExtractor) extractFloatFromObject(obj *ast.ObjectExpression, key string, defaultValue float64) float64 {
	for _, prop := range obj.Properties {
		if keyIdent, ok := prop.Key.(*ast.Identifier); ok && keyIdent.Name == key {
			result := ice.argParser.ParseFloat(prop.Value)
			if result.IsValid {
				return result.MustBeFloat()
			}
		}
	}
	return defaultValue
}

func (ice *InputConstantExtractor) extractBoolFromObject(obj *ast.ObjectExpression, key string, defaultValue bool) bool {
	for _, prop := range obj.Properties {
		if keyIdent, ok := prop.Key.(*ast.Identifier); ok && keyIdent.Name == key {
			result := ice.argParser.ParseBool(prop.Value)
			if result.IsValid {
				return result.MustBeBool()
			}
		}
	}
	return defaultValue
}

func (ice *InputConstantExtractor) extractStringFromObject(obj *ast.ObjectExpression, key string, defaultValue string) string {
	for _, prop := range obj.Properties {
		if keyIdent, ok := prop.Key.(*ast.Identifier); ok && keyIdent.Name == key {
			result := ice.argParser.ParseString(prop.Value)
			if result.IsValid {
				return result.MustBeString()
			}
		}
	}
	return defaultValue
}
