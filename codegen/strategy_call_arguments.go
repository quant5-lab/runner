package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/*
strategyCallArguments gives uniform access to a Pine strategy command's
arguments whether they are written positionally, by name, or mixed.

The converter appends one ObjectExpression holding every named argument after
the positional arguments, so a parameter can be read by its name or by its Pine
positional slot without the reader knowing the call shape.
*/
type strategyCallArguments struct {
	positional []ast.Expression
	named      *ast.ObjectExpression
}

// newStrategyCallArguments splits a Pine call into positional arguments and the
// named-argument container. A converter appends one container; a call carrying
// several still resolves every named parameter, because the containers are read
// as one. A single container is kept as-is so the resolver returns the original
// node.
func newStrategyCallArguments(args []ast.Expression) strategyCallArguments {
	call := strategyCallArguments{}
	containers := make([]*ast.ObjectExpression, 0, 1)
	for _, arg := range args {
		if object, ok := arg.(*ast.ObjectExpression); ok {
			containers = append(containers, object)
			continue
		}
		call.positional = append(call.positional, arg)
	}
	switch len(containers) {
	case 0:
	case 1:
		call.named = containers[0]
	default:
		var properties []ast.Property
		for _, container := range containers {
			properties = append(properties, container.Properties...)
		}
		call.named = &ast.ObjectExpression{NodeType: ast.TypeObjectExpression, Properties: properties}
	}
	return call
}

func (c strategyCallArguments) namedValue(name string) (ast.Expression, bool) {
	if c.named == nil {
		return nil, false
	}
	for _, prop := range c.named.Properties {
		if propertyName(prop.Key) == name {
			return prop.Value, true
		}
	}
	return nil, false
}

func (c strategyCallArguments) positionalValue(slot int) (ast.Expression, bool) {
	if slot < 0 || slot >= len(c.positional) {
		return nil, false
	}
	return c.positional[slot], true
}

// value returns the argument for a parameter, preferring the named form over
// the positional slot. A negative slot marks a parameter with no Pine position.
func (c strategyCallArguments) value(name string, slot int) (ast.Expression, bool) {
	if expr, ok := c.namedValue(name); ok {
		return expr, true
	}
	return c.positionalValue(slot)
}

// valueOrContainer returns the positional argument at slot when present, and
// otherwise the named-argument container so a resolver can read a named key.
func (c strategyCallArguments) valueOrContainer(slot int) (ast.Expression, bool) {
	if expr, ok := c.positionalValue(slot); ok {
		return expr, true
	}
	if c.named != nil {
		return c.named, true
	}
	return nil, false
}

// strategyStringArgument reads a string parameter by name, then by positional
// slot, returning the fallback when the parameter is absent.
func (g *generator) strategyStringArgument(call strategyCallArguments, name string, slot int, fallback string) string {
	if expr, ok := call.namedValue(name); ok {
		if value, isString := literalString(expr); isString {
			return value
		}
	}
	if expr, ok := call.positionalValue(slot); ok {
		return g.extractStringLiteral(expr)
	}
	return fallback
}

// strategyExpressionArgument reads a named-only parameter as a Go expression,
// returning the fallback when the parameter is absent.
func (g *generator) strategyExpressionArgument(call strategyCallArguments, name, fallback string) string {
	expr, ok := call.namedValue(name)
	if !ok {
		return fallback
	}
	return strings.TrimRight(g.extractSeriesExpression(expr), "\n")
}

func literalString(expr ast.Expression) (string, bool) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return "", false
	}
	value, ok := lit.Value.(string)
	return value, ok
}
