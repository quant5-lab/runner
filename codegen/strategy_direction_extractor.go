package codegen

import "github.com/quant5-lab/runner/ast"

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

// resolveStrategyDirectionMember matches the nested member expression
// `strategy.direction.{long,short,all}` and returns the corresponding Go
// runtime constant. Returns ("", false) for any other expression shape.
//
// This is the canonical resolver used by both the string-expression generator
// (for `strat_dir_value = strategy.direction.long` style assignments) and the
// direction-argument extractor chain (for direct calls like
// `strategy.risk.allow_entry_in(strategy.direction.all)`).
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

func (e *ChainDirectionExtractor) Extract(expr ast.Expression) string {
	for _, extractor := range e.extractors {
		if direction, ok := extractor.Extract(expr); ok {
			return direction
		}
	}
	return "strategy.Long"
}

// Deprecated: use NewContextAwareDirectionExtractor for production codegen; kept for unit tests of individual extractor components.
func NewDefaultDirectionExtractor() *ChainDirectionExtractor {
	return NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
	)
}

// VariableDirectionExtractor resolves an Identifier to a Go variable name when
// that identifier exists in the generator's variable registry. This enables
// `strategy.entry(entry_type, ...)` where entry_type is a runtime string variable.
type VariableDirectionExtractor struct {
	variables map[string]string
}

func (e *VariableDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	id, ok := expr.(*ast.Identifier)
	if !ok {
		return "", false
	}
	if varType, exists := e.variables[id.Name]; exists && varType == "string" {
		return id.Name, true
	}
	return "", false
}

// ConditionalDirectionExtractor handles ternary direction expressions by emitting
// a Go IIFE: `func() string { if <cond> { return strategy.Long } else { return strategy.Short } }()`.
type ConditionalDirectionExtractor struct {
	chain *ChainDirectionExtractor
}

func (e *ConditionalDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	cond, ok := expr.(*ast.ConditionalExpression)
	if !ok {
		return "", false
	}
	consequent, consOk := e.chain.extractors[0].Extract(cond.Consequent)
	if !consOk {
		// Try full chain for nested ternary
		for _, ex := range e.chain.extractors {
			if v, ok2 := ex.Extract(cond.Consequent); ok2 {
				consequent = v
				consOk = true
				break
			}
		}
	}
	alternate, altOk := e.chain.extractors[0].Extract(cond.Alternate)
	if !altOk {
		for _, ex := range e.chain.extractors {
			if v, ok2 := ex.Extract(cond.Alternate); ok2 {
				alternate = v
				altOk = true
				break
			}
		}
	}
	if !consOk || !altOk {
		return "", false
	}
	// We can't call generateConditionExpression here without the generator;
	// delegate to the context-aware extractor which has the generator.
	_, _ = consequent, alternate
	return "", false
}

// contextualConditionalDirectionExtractor has access to the generator for full expression codegen.
type contextualConditionalDirectionExtractor struct {
	gen   *generator
	chain *ChainDirectionExtractor
}

func (e *contextualConditionalDirectionExtractor) Extract(expr ast.Expression) (string, bool) {
	cond, ok := expr.(*ast.ConditionalExpression)
	if !ok {
		return "", false
	}
	var consequent, alternate string
	for _, ex := range e.chain.extractors {
		if v, ok2 := ex.Extract(cond.Consequent); ok2 {
			consequent = v
			break
		}
	}
	for _, ex := range e.chain.extractors {
		if v, ok2 := ex.Extract(cond.Alternate); ok2 {
			alternate = v
			break
		}
	}
	if consequent == "" || alternate == "" {
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

// NewContextAwareDirectionExtractor returns a chain that handles member expressions,
// boolean literals, variable identifiers, and conditional expressions.
func NewContextAwareDirectionExtractor(gen *generator) *ChainDirectionExtractor {
	chain := &ChainDirectionExtractor{}
	chain.extractors = []DirectionExtractor{
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&VariableDirectionExtractor{variables: gen.variables},
		&IdentifierDirectionExtractor{},
		&contextualConditionalDirectionExtractor{gen: gen, chain: chain},
	}
	return chain
}
