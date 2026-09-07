package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type DirectionExtractor interface {
	Extract(expr ast.Expression) (string, bool)
}

type MemberExpressionDirectionExtractor struct{}

func (e *MemberExpressionDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	if dir, ok := resolveStrategyDirectionMember(expr); ok {
		return dir, true
	}
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

func resolveStrategyDirectionMember(expr ast.Expression) (string, bool) {
	outer, ok := expr.(*ast.MemberExpression)
	if !ok {
		return "", false
	}
	prop, ok := outer.Property.(*ast.Identifier)
	if !ok {
		return "", false
	}
	inner, ok := outer.Object.(*ast.MemberExpression)
	if !ok {
		return "", false
	}
	innerObj, ok := inner.Object.(*ast.Identifier)
	if !ok || innerObj.Name != "strategy" {
		return "", false
	}
	innerProp, ok := inner.Property.(*ast.Identifier)
	if !ok || innerProp.Name != "direction" {
		return "", false
	}
	switch prop.Name {
	case "long":
		return "strategy.DirectionLong", true
	case "short":
		return "strategy.DirectionShort", true
	case "all":
		return "strategy.DirectionAll", true
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

func (e *ChainDirectionExtractor) Resolve(expr ast.Expression) (string, error) {
	for _, extractor := range e.extractors {
		if direction, ok := extractor.Extract(expr); ok {
			return direction, nil
		}
	}
	return "", fmt.Errorf(
		"direction expression %T cannot be resolved to strategy.Long or strategy.Short — "+
			"use strategy.long, strategy.short, true, false, or a string variable holding a direction value",
		expr,
	)
}

type VariableDirectionExtractor struct {
	variables map[string]string
}

func (e *VariableDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	id, ok := expr.(*ast.Identifier)
	if !ok || id.Name == "" {
		return "", false
	}
	if varType, exists := e.variables[id.Name]; exists && varType == "string" {
		return id.Name, true
	}
	return "", false
}

// namedDirectionArgExtractor handles the Pine named-argument form
//
//	strategy.entry(id, long=<direction>, ...)
//
// The parser bundles all named arguments into a single *ast.ObjectExpression
// appended to call.Arguments. The "long" property value is resolved through
// chain.Resolve, so every direction form supported for positional arguments is
// automatically available inside a named argument.
type namedDirectionArgExtractor struct {
	chain *ChainDirectionExtractor
}

func (e *namedDirectionArgExtractor) Extract(expr ast.Expression) (string, bool) {
	obj, ok := expr.(*ast.ObjectExpression)
	if !ok {
		return "", false
	}
	for _, prop := range obj.Properties {
		key, ok := prop.Key.(*ast.Identifier)
		if !ok || key.Name != "long" {
			continue
		}
		dir, err := e.chain.Resolve(prop.Value)
		if err != nil {
			return "", false
		}
		return dir, true
	}
	return "", false
}

// contextualConditionalDirectionExtractor branch resolution delegates to chain.Resolve
// so any extractor added to the chain automatically applies to ternary branches.
type contextualConditionalDirectionExtractor struct {
	gen   *generator
	chain *ChainDirectionExtractor
}

func (e *contextualConditionalDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	cond, ok := expr.(*ast.ConditionalExpression)
	if !ok {
		return "", false
	}
	consequent, err := e.chain.Resolve(cond.Consequent)
	if err != nil {
		return "", false
	}
	alternate, err := e.chain.Resolve(cond.Alternate)
	if err != nil {
		return "", false
	}
	condCode, err := e.gen.generateConditionExpression(cond.Test)
	if err != nil {
		return "", false
	}
	condCode = e.gen.addBoolConversionIfNeeded(cond.Test, condCode)
	iife := "func() string { if " + condCode + " { return " + consequent + " } else { return " + alternate + " } }()"
	return iife, true
}

func NewContextAwareDirectionExtractor(gen *generator) *ChainDirectionExtractor {
	chain := &ChainDirectionExtractor{}
	chain.extractors = []DirectionExtractor{
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&VariableDirectionExtractor{variables: gen.variables},
		&IdentifierDirectionExtractor{},
		&namedDirectionArgExtractor{chain: chain},
		&contextualConditionalDirectionExtractor{gen: gen, chain: chain},
	}
	return chain
}
