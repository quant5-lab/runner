package codegen

import "github.com/quant5-lab/runner/ast"

var v4InputTypeMapping = map[string]string{
	"session":   "input.session",
	"source":    "input.source",
	"integer":   "input.int",
	"float":     "input.float",
	"bool":      "input.bool",
	"string":    "input.string",
	"symbol":    "input.symbol",
	"time":      "input.time",
	"timeframe": "input.timeframe",
	"price":     "input.price",
	"color":     "input.color",
	"text_area": "input.text_area",
}

var sourceIdentifierNames = map[string]bool{
	"hl2": true, "hlc3": true, "ohlc4": true, "hlcc4": true,
	"open": true, "high": true, "low": true, "close": true, "volume": true,
}

func resolveInputFuncName(call *ast.CallExpression) string {
	if resolved := resolveFromExplicitTypeParam(call); resolved != "" {
		return resolved
	}
	return resolveFromDefvalType(call)
}

func resolveFromExplicitTypeParam(call *ast.CallExpression) string {
	typeExpr := findNamedArgValue(call, "type")
	if typeExpr == nil {
		return ""
	}

	memExpr, ok := typeExpr.(*ast.MemberExpression)
	if !ok {
		return ""
	}

	objId, ok := memExpr.Object.(*ast.Identifier)
	if !ok || objId.Name != "input" {
		return ""
	}

	propId, ok := memExpr.Property.(*ast.Identifier)
	if !ok {
		return ""
	}

	return v4InputTypeMapping[propId.Name]
}

func resolveFromDefvalType(call *ast.CallExpression) string {
	defvalExpr := extractDefvalExpression(call)
	if defvalExpr == nil {
		return ""
	}

	if ident, ok := defvalExpr.(*ast.Identifier); ok {
		if sourceIdentifierNames[ident.Name] {
			return "input.source"
		}
		return ""
	}

	lit, ok := defvalExpr.(*ast.Literal)
	if !ok {
		return ""
	}

	return inputFuncNameFromLiteral(lit)
}

func extractDefvalExpression(call *ast.CallExpression) ast.Expression {
	if len(call.Arguments) == 0 {
		return nil
	}

	firstArg := call.Arguments[0]

	obj, ok := firstArg.(*ast.ObjectExpression)
	if !ok {
		return firstArg
	}

	return findPropertyValue(obj, "defval")
}

func inputFuncNameFromLiteral(lit *ast.Literal) string {
	switch v := lit.Value.(type) {
	case float64:
		if v == float64(int(v)) {
			return "input.int"
		}
		return "input.float"
	case int:
		return "input.int"
	case bool:
		return "input.bool"
	case string:
		return "input.string"
	default:
		return ""
	}
}

func findNamedArgValue(call *ast.CallExpression, name string) ast.Expression {
	for _, arg := range call.Arguments {
		obj, ok := arg.(*ast.ObjectExpression)
		if !ok {
			continue
		}
		if value := findPropertyValue(obj, name); value != nil {
			return value
		}
	}
	return nil
}

func findPropertyValue(obj *ast.ObjectExpression, name string) ast.Expression {
	for _, prop := range obj.Properties {
		keyId, ok := prop.Key.(*ast.Identifier)
		if ok && keyId.Name == name {
			return prop.Value
		}
	}
	return nil
}
