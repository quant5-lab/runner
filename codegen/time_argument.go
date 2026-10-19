package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

type ArgumentType int

const (
	ArgumentTypeUnknown ArgumentType = iota
	ArgumentTypeLiteral
	ArgumentTypeIdentifier
	ArgumentTypeWrappedIdentifier
)

type SessionArgument struct {
	Type  ArgumentType
	Value string
}

func (a SessionArgument) IsVariable() bool {
	return a.Type == ArgumentTypeIdentifier || a.Type == ArgumentTypeWrappedIdentifier
}

func (a SessionArgument) IsLiteral() bool {
	return a.Type == ArgumentTypeLiteral
}

func (a SessionArgument) IsValid() bool {
	return a.Type != ArgumentTypeUnknown && a.Value != ""
}

// SessionArgumentParser parses session arguments using the unified ArgumentParser
// This demonstrates reusability: delegating to shared parsing infrastructure
type SessionArgumentParser struct {
	argParser *ArgumentParser
}

func NewSessionArgumentParser() *SessionArgumentParser {
	return &SessionArgumentParser{
		argParser: NewArgumentParser(),
	}
}

// Parse uses the unified ArgumentParser.ParseSession method
// This eliminates duplicate parsing logic and improves maintainability
func (p *SessionArgumentParser) Parse(expr ast.Expression) SessionArgument {
	if expr == nil {
		return SessionArgument{Type: ArgumentTypeUnknown}
	}

	// Leverage unified parsing framework
	result := p.argParser.ParseSession(expr)

	if !result.IsValid {
		return SessionArgument{Type: ArgumentTypeUnknown}
	}

	// Map ParsedArgument to SessionArgument
	if result.IsLiteral {
		return SessionArgument{
			Type:  ArgumentTypeLiteral,
			Value: result.MustBeString(),
		}
	}

	// It's an identifier (possibly wrapped)
	// Check if it was originally wrapped by inspecting the source
	if _, ok := result.SourceExpr.(*ast.MemberExpression); ok {
		return SessionArgument{
			Type:  ArgumentTypeWrappedIdentifier,
			Value: result.Identifier,
		}
	}

	return SessionArgument{
		Type:  ArgumentTypeIdentifier,
		Value: result.Identifier,
	}
}

// Legacy methods kept for backward compatibility with existing tests
// These now delegate to the unified ArgumentParser
