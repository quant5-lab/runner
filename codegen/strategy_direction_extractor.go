package codegen

import "github.com/quant5-lab/runner/ast"

type DirectionExtractor interface {
	Extract(expr ast.Expression) (string, bool)
}

type MemberExpressionDirectionExtractor struct{}

func (e *MemberExpressionDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	mem, ok := expr.(*ast.MemberExpression)
	if !ok {
		return "", false
	}
	prop, ok := mem.Property.(*ast.Identifier)
	if !ok {
		return "", false
	}
	switch prop.Name {
	case "long":
		return "strategy.Long", true
	case "short":
		return "strategy.Short", true
	}
	return "", false
}

type BooleanLiteralDirectionExtractor struct{}

func (e *BooleanLiteralDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return "", false
	}
	boolVal, ok := lit.Value.(bool)
	if !ok {
		return "", false
	}
	if boolVal {
		return "strategy.Long", true
	}
	return "strategy.Short", true
}

type IdentifierDirectionExtractor struct{}

func (e *IdentifierDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	id, ok := expr.(*ast.Identifier)
	if !ok {
		return "", false
	}
	switch id.Name {
	case "true":
		return "strategy.Long", true
	case "false":
		return "strategy.Short", true
	}
	return "", false
}

type ChainDirectionExtractor struct {
	extractors []DirectionExtractor
}

func NewChainDirectionExtractor(extractors ...DirectionExtractor) *ChainDirectionExtractor {
	return &ChainDirectionExtractor{extractors: extractors}
}

func (e *ChainDirectionExtractor) Extract(expr ast.Expression) string {
	for _, extractor := range e.extractors {
		if direction, ok := extractor.Extract(expr); ok {
			return direction
		}
	}
	return "strategy.Long"
}

func NewDefaultDirectionExtractor() *ChainDirectionExtractor {
	return NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
	)
}
