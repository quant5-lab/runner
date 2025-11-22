package codegen

import "github.com/borisquantlab/pinescript-go/ast"

type PropertyParser struct {
	extractors map[string]PropertyExtractor
}

func NewPropertyParser() *PropertyParser {
	return &PropertyParser{
		extractors: map[string]PropertyExtractor{
			"string":     StringExtractor{},
			"int":        IntExtractor{},
			"float":      FloatExtractor{},
			"bool":       BoolExtractor{},
			"identifier": IdentifierExtractor{},
		},
	}
}

func (p *PropertyParser) ParseString(obj *ast.ObjectExpression, key string) (string, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return "", false
	}

	result, ok := p.extractors["string"].Extract(value)
	if !ok {
		return "", false
	}
	return result.(string), true
}

func (p *PropertyParser) ParseInt(obj *ast.ObjectExpression, key string) (int, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return 0, false
	}

	result, ok := p.extractors["int"].Extract(value)
	if !ok {
		return 0, false
	}
	return result.(int), true
}

func (p *PropertyParser) ParseFloat(obj *ast.ObjectExpression, key string) (float64, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return 0, false
	}

	result, ok := p.extractors["float"].Extract(value)
	if !ok {
		return 0, false
	}
	return result.(float64), true
}

func (p *PropertyParser) ParseBool(obj *ast.ObjectExpression, key string) (bool, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return false, false
	}

	result, ok := p.extractors["bool"].Extract(value)
	if !ok {
		return false, false
	}
	return result.(bool), true
}

func (p *PropertyParser) ParseIdentifier(obj *ast.ObjectExpression, key string) (string, bool) {
	value := p.findProperty(obj, key)
	if value == nil {
		return "", false
	}

	result, ok := p.extractors["identifier"].Extract(value)
	if !ok {
		return "", false
	}
	return result.(string), true
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
