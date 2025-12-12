package codegen

import "github.com/quant5-lab/runner/ast"

/*
PropertyParser extracts typed values from ObjectExpression properties.

Reusability: Delegates parsing to unified ArgumentParser framework.
Design: Provides high-level API for object property extraction while
leveraging ArgumentParser for type-safe value parsing.
*/
type PropertyParser struct {
	argParser *ArgumentParser // Unified parsing infrastructure
}

func NewPropertyParser() *PropertyParser {
	return &PropertyParser{
		argParser: NewArgumentParser(),
	}
}

func (p *PropertyParser) ParseString(obj *ast.ObjectExpression, key string) (string, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return "", false
	}

	result := p.argParser.ParseString(value)
	if !result.IsValid {
		return "", false
	}
	return result.MustBeString(), true
}

func (p *PropertyParser) ParseInt(obj *ast.ObjectExpression, key string) (int, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return 0, false
	}

	result := p.argParser.ParseInt(value)
	if !result.IsValid {
		return 0, false
	}
	return result.MustBeInt(), true
}

func (p *PropertyParser) ParseFloat(obj *ast.ObjectExpression, key string) (float64, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return 0, false
	}

	result := p.argParser.ParseFloat(value)
	if !result.IsValid {
		return 0, false
	}
	return result.MustBeFloat(), true
}

func (p *PropertyParser) ParseBool(obj *ast.ObjectExpression, key string) (bool, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return false, false
	}

	result := p.argParser.ParseBool(value)
	if !result.IsValid {
		return false, false
	}
	return result.MustBeBool(), true
}

func (p *PropertyParser) ParseIdentifier(obj *ast.ObjectExpression, key string) (string, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return "", false
	}

	result := p.argParser.ParseIdentifier(value)
	if !result.IsValid {
		return "", false
	}
	return result.Identifier, true
}

func (p *PropertyParser) findProperty(obj *ast.ObjectExpression, key string) ast.Expression {
	for _, prop := range obj.Properties {
		keyID, ok := prop.Key.(*ast.Identifier)
		if !ok {
			continue
		}
		if keyID.Name == key {
			return prop.Value
		}
	}
	return nil
}
