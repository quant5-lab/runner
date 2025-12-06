package codegen

import (
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/*
ArgumentParser provides a unified, reusable framework for parsing AST expressions
into typed argument values across all function handlers.

Design Philosophy:
- Single Responsibility: Only parsing, no code generation
- Open/Closed: Easily extended with new argument types
- DRY: Eliminates duplicate parsing logic across handlers
- Type Safety: Returns strongly-typed argument values

Usage:

	parser := NewArgumentParser()
	arg := parser.ParseString(expr)        // Parse string literal
	arg := parser.ParseInt(expr)           // Parse int literal
	arg := parser.ParseFloat(expr)         // Parse float literal
	arg := parser.ParseBool(expr)          // Parse bool literal
	arg := parser.ParseIdentifier(expr)    // Parse identifier
	arg := parser.ParseSession(expr)       // Parse session string (literal or identifier)
*/
type ArgumentParser struct{}

func NewArgumentParser() *ArgumentParser {
	return &ArgumentParser{}
}

/*
ParsedArgument represents a successfully parsed argument with its type and value.
*/
type ParsedArgument struct {
	IsValid    bool
	IsLiteral  bool           // true if literal value, false if identifier/expression
	Value      interface{}    // The parsed value (string, int, float64, bool)
	Identifier string         // Identifier name if IsLiteral=false
	SourceExpr ast.Expression // Original expression for debugging
}

// ============================================================================
// String Parsing
// ============================================================================

/*
ParseString extracts a string literal from an AST expression.
Handles both single and double quotes, and trims them.

Returns:

	ParsedArgument.IsValid = true if string literal found
	ParsedArgument.Value = trimmed string value
*/
func (p *ArgumentParser) ParseString(expr ast.Expression) ParsedArgument {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	str, ok := lit.Value.(string)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	// Trim quotes
	trimmed := strings.Trim(str, "'\"")

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  true,
		Value:      trimmed,
		SourceExpr: expr,
	}
}

/*
ParseStringOrIdentifier extracts a string literal OR identifier name.
Useful for arguments that accept both: "literal" or variable_name

Returns:

	ParsedArgument.IsLiteral = true if string literal, false if identifier
	ParsedArgument.Value = string value (if literal)
	ParsedArgument.Identifier = identifier name (if identifier)
*/
func (p *ArgumentParser) ParseStringOrIdentifier(expr ast.Expression) ParsedArgument {
	// Try string literal first
	if result := p.ParseString(expr); result.IsValid {
		return result
	}

	// Try identifier
	if result := p.ParseIdentifier(expr); result.IsValid {
		return result
	}

	return ParsedArgument{IsValid: false, SourceExpr: expr}
}

// ============================================================================
// Numeric Parsing
// ============================================================================

/*
ParseInt extracts an integer literal from an AST expression.
Handles both int and float64 AST literal types.

Returns:

	ParsedArgument.IsValid = true if numeric literal found
	ParsedArgument.Value = int value
*/
func (p *ArgumentParser) ParseInt(expr ast.Expression) ParsedArgument {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	var intValue int
	switch v := lit.Value.(type) {
	case int:
		intValue = v
	case float64:
		intValue = int(v)
	default:
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  true,
		Value:      intValue,
		SourceExpr: expr,
	}
}

/*
ParseFloat extracts a float literal from an AST expression.
Handles both float64 and int AST literal types.

Returns:

	ParsedArgument.IsValid = true if numeric literal found
	ParsedArgument.Value = float64 value
*/
func (p *ArgumentParser) ParseFloat(expr ast.Expression) ParsedArgument {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	var floatValue float64
	switch v := lit.Value.(type) {
	case float64:
		floatValue = v
	case int:
		floatValue = float64(v)
	default:
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  true,
		Value:      floatValue,
		SourceExpr: expr,
	}
}

// ============================================================================
// Boolean Parsing
// ============================================================================

/*
ParseBool extracts a boolean literal from an AST expression.

Returns:

	ParsedArgument.IsValid = true if bool literal found
	ParsedArgument.Value = bool value
*/
func (p *ArgumentParser) ParseBool(expr ast.Expression) ParsedArgument {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	boolValue, ok := lit.Value.(bool)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  true,
		Value:      boolValue,
		SourceExpr: expr,
	}
}

// ============================================================================
// Identifier Parsing
// ============================================================================

/*
ParseIdentifier extracts an identifier name from an AST expression.

Returns:

	ParsedArgument.IsValid = true if identifier found
	ParsedArgument.IsLiteral = false (it's a variable reference)
	ParsedArgument.Identifier = identifier name
*/
func (p *ArgumentParser) ParseIdentifier(expr ast.Expression) ParsedArgument {
	ident, ok := expr.(*ast.Identifier)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  false,
		Identifier: ident.Name,
		SourceExpr: expr,
	}
}

// ============================================================================
// Complex Parsing (Wrapped Identifiers)
// ============================================================================

/*
ParseWrappedIdentifier extracts an identifier from a parser-wrapped expression.
Pine parser sometimes wraps variables as: my_var → MemberExpression(my_var, Literal(0), computed=true)

This handles the pattern: identifier[0] where the [0] is a parser artifact.

Returns:

	ParsedArgument.IsValid = true if wrapped identifier found
	ParsedArgument.IsLiteral = false (it's a variable reference)
	ParsedArgument.Identifier = unwrapped identifier name
*/
func (p *ArgumentParser) ParseWrappedIdentifier(expr ast.Expression) ParsedArgument {
	mem, ok := expr.(*ast.MemberExpression)
	if !ok || !mem.Computed {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	obj, ok := mem.Object.(*ast.Identifier)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	lit, ok := mem.Property.(*ast.Literal)
	if !ok {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	idx, ok := lit.Value.(int)
	if !ok || idx != 0 {
		return ParsedArgument{IsValid: false, SourceExpr: expr}
	}

	return ParsedArgument{
		IsValid:    true,
		IsLiteral:  false,
		Identifier: obj.Name,
		SourceExpr: expr,
	}
}

/*
ParseIdentifierOrWrapped tries to parse as identifier first, then wrapped identifier.
This is the most common pattern for variable references.

Returns:

	ParsedArgument.IsValid = true if identifier or wrapped identifier found
	ParsedArgument.IsLiteral = false
	ParsedArgument.Identifier = identifier name (unwrapped if necessary)
*/
func (p *ArgumentParser) ParseIdentifierOrWrapped(expr ast.Expression) ParsedArgument {
	// Try simple identifier first
	if result := p.ParseIdentifier(expr); result.IsValid {
		return result
	}

	// Try wrapped identifier
	if result := p.ParseWrappedIdentifier(expr); result.IsValid {
		return result
	}

	return ParsedArgument{IsValid: false, SourceExpr: expr}
}

// ============================================================================
// Session-Specific Parsing (Reusable Pattern)
// ============================================================================

/*
ParseSession extracts a session string from various forms:
  - String literal: "0950-1645"
  - Identifier: entry_time_input
  - Wrapped identifier: my_session[0]

This combines multiple parsing strategies for maximum flexibility.

Returns:

	ParsedArgument.IsValid = true if any valid form found
	ParsedArgument.IsLiteral = true if string literal, false if variable
	ParsedArgument.Value = string value (if literal)
	ParsedArgument.Identifier = identifier name (if variable)
*/
func (p *ArgumentParser) ParseSession(expr ast.Expression) ParsedArgument {
	// Try string literal first
	if result := p.ParseString(expr); result.IsValid {
		return result
	}

	// Try identifier or wrapped identifier
	if result := p.ParseIdentifierOrWrapped(expr); result.IsValid {
		return result
	}

	return ParsedArgument{IsValid: false, SourceExpr: expr}
}

// ============================================================================
// Helper Methods
// ============================================================================

/*
MustBeString returns the string value or empty string if invalid.
Useful for quick extraction when you know the type.
*/
func (arg ParsedArgument) MustBeString() string {
	if !arg.IsValid {
		return ""
	}
	if arg.IsLiteral {
		if str, ok := arg.Value.(string); ok {
			return str
		}
	}
	return arg.Identifier
}

/*
MustBeInt returns the int value or 0 if invalid.
*/
func (arg ParsedArgument) MustBeInt() int {
	if !arg.IsValid || !arg.IsLiteral {
		return 0
	}
	if val, ok := arg.Value.(int); ok {
		return val
	}
	return 0
}

/*
MustBeFloat returns the float64 value or 0.0 if invalid.
*/
func (arg ParsedArgument) MustBeFloat() float64 {
	if !arg.IsValid || !arg.IsLiteral {
		return 0.0
	}
	if val, ok := arg.Value.(float64); ok {
		return val
	}
	return 0.0
}

/*
MustBeBool returns the bool value or false if invalid.
*/
func (arg ParsedArgument) MustBeBool() bool {
	if !arg.IsValid || !arg.IsLiteral {
		return false
	}
	if val, ok := arg.Value.(bool); ok {
		return val
	}
	return false
}
