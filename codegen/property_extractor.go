package codegen

import "github.com/quant5-lab/runner/ast"

type PropertyExtractor interface {
	Extract(value ast.Expression) (interface{}, bool)
}

type StringExtractor struct{}

func (e StringExtractor) Extract(value ast.Expression) (interface{}, bool) {
	lit, ok := value.(*ast.Literal)
	if !ok {
		return nil, false
	}

	str, ok := lit.Value.(string)
	return str, ok
}

type IntExtractor struct{}

func (e IntExtractor) Extract(value ast.Expression) (interface{}, bool) {
	lit, ok := value.(*ast.Literal)
	if !ok {
		return nil, false
	}

	switch v := lit.Value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return nil, false
	}
}

type FloatExtractor struct{}

func (e FloatExtractor) Extract(value ast.Expression) (interface{}, bool) {
	lit, ok := value.(*ast.Literal)
	if !ok {
		return nil, false
	}

	switch v := lit.Value.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	default:
		return nil, false
	}
}

type BoolExtractor struct{}

func (e BoolExtractor) Extract(value ast.Expression) (interface{}, bool) {
	lit, ok := value.(*ast.Literal)
	if !ok {
		return nil, false
	}

	b, ok := lit.Value.(bool)
	return b, ok
}

type IdentifierExtractor struct{}

func (e IdentifierExtractor) Extract(value ast.Expression) (interface{}, bool) {
	id, ok := value.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	return id.Name, true
}
