package validation

import (
	"math"

	"github.com/quant5-lab/runner/ast"
)

type IdentifierLookup struct {
	constants ConstantStore
}

func NewIdentifierLookup(constants ConstantStore) *IdentifierLookup {
	return &IdentifierLookup{
		constants: constants,
	}
}

func (l *IdentifierLookup) Resolve(identifier *ast.Identifier) float64 {
	if identifier == nil {
		return math.NaN()
	}

	if value, exists := l.constants.Get(identifier.Name); exists {
		return value
	}

	return math.NaN()
}

func (l *IdentifierLookup) ResolveWrappedVariable(member *ast.MemberExpression) float64 {
	if !l.isParserWrappedVariable(member) {
		return math.NaN()
	}

	identifier, ok := member.Object.(*ast.Identifier)
	if !ok {
		return math.NaN()
	}

	return l.Resolve(identifier)
}

func (l *IdentifierLookup) isParserWrappedVariable(member *ast.MemberExpression) bool {
	if !member.Computed {
		return false
	}

	literal, ok := member.Property.(*ast.Literal)
	if !ok {
		return false
	}

	index, ok := literal.Value.(int)
	return ok && index == 0
}
