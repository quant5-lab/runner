package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// strategyExitArgParser resolves strategy.exit() arguments regardless of call shape.
//
// Pine allows three call shapes:
//
//	(A) All named:   strategy.exit(id="X", from_entry="Y", stop=s, limit=l)
//	(B) id only:     strategy.exit("X", stop=s, limit=l)          ← from_entry omitted
//	(C) Full:        strategy.exit("X", "Y", stop=s, limit=l)
//
// The parser packs every named argument into one ObjectExpression appended after all
// positional arguments, so the three shapes produce:
//
//	(A) Arguments = [ObjExpr{id, from_entry, stop, limit}]
//	(B) Arguments = [Literal("X"), ObjExpr{stop, limit}]
//	(C) Arguments = [Literal("X"), Literal("Y"), ObjExpr{stop, limit}]
//
// strategyExitArgParser splits args into positional and namedObj, then provides
// uniform string/expression accessors across all shapes.
type strategyExitArgParser struct {
	positional []ast.Expression
	namedObj   *ast.ObjectExpression
	g          *generator
}

func newStrategyExitArgParser(args []ast.Expression, g *generator) *strategyExitArgParser {
	p := &strategyExitArgParser{g: g}
	for _, arg := range args {
		if obj, ok := arg.(*ast.ObjectExpression); ok {
			p.namedObj = obj
		} else {
			p.positional = append(p.positional, arg)
		}
	}
	return p
}

// stringArg returns a Go string constant from positional index or named key, or def.
func (p *strategyExitArgParser) stringArg(positionalIdx int, namedKey, def string) string {
	if p.namedObj != nil {
		if lit, ok := p.namedLiteral(namedKey); ok {
			if s, ok := lit.(string); ok {
				return s
			}
		}
	}
	if positionalIdx < len(p.positional) {
		return p.g.extractStringLiteral(p.positional[positionalIdx])
	}
	return def
}

// exprArg returns Go expression code from positional index or named key, or def.
func (p *strategyExitArgParser) exprArg(positionalIdx int, namedKey, def string) string {
	if p.namedObj != nil {
		for _, prop := range p.namedObj.Properties {
			ident, ok := prop.Key.(*ast.Identifier)
			if !ok || ident.Name != namedKey {
				continue
			}
			code := p.g.extractSeriesExpression(prop.Value)
			return strings.TrimRight(code, "\n")
		}
	}
	if positionalIdx < len(p.positional) {
		code := p.g.extractSeriesExpression(p.positional[positionalIdx])
		return strings.TrimRight(code, "\n")
	}
	return def
}

// commentArg returns a Go string expression suitable for a comment parameter.
func (p *strategyExitArgParser) commentArg() string {
	extractor := &ArgumentExtractor{generator: p.g}
	unified := append(p.positional, p.namedExprOrNil()...)
	return extractor.ExtractCommentArgument(unified, "comment", -1, `""`)
}

// whenCondition returns the when= condition and whether one was found.
func (p *strategyExitArgParser) whenCondition() (string, bool) {
	extractor := &ArgumentExtractor{generator: p.g}
	unified := append(p.positional, p.namedExprOrNil()...)
	return extractor.ExtractWhenCondition(unified)
}

func (p *strategyExitArgParser) namedLiteral(key string) (interface{}, bool) {
	if p.namedObj == nil {
		return nil, false
	}
	for _, prop := range p.namedObj.Properties {
		ident, ok := prop.Key.(*ast.Identifier)
		if !ok || ident.Name != key {
			continue
		}
		lit, ok := prop.Value.(*ast.Literal)
		if !ok {
			return nil, false
		}
		return lit.Value, true
	}
	return nil, false
}

func (p *strategyExitArgParser) namedExprOrNil() []ast.Expression {
	if p.namedObj == nil {
		return nil
	}
	return []ast.Expression{p.namedObj}
}
