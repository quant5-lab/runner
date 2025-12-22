package codegen

import "fmt"

/*
ArrowIdentifierResolver resolves identifiers to their correct access patterns in arrow function context.

Responsibility (SRP):
  - Single purpose: determine if identifier needs Series.GetCurrent() wrapper
  - Delegates classification to ArrowSeriesAccessResolver
  - No knowledge of code generation or expression evaluation

Design Pattern: Strategy Pattern
  - Uses ArrowSeriesAccessResolver as strategy for classification
  - Provides clean interface for identifier resolution logic
*/
type ArrowIdentifierResolver struct {
	accessResolver *ArrowSeriesAccessResolver
}

func NewArrowIdentifierResolver(resolver *ArrowSeriesAccessResolver) *ArrowIdentifierResolver {
	return &ArrowIdentifierResolver{
		accessResolver: resolver,
	}
}

/*
ResolveIdentifier determines the correct Go code for accessing an identifier.
*/
func (r *ArrowIdentifierResolver) ResolveIdentifier(identifierName string) string {
	if access, resolved := r.accessResolver.ResolveAccess(identifierName); resolved {
		return access
	}
	return identifierName
}

/*
IsLocalVariable checks if identifier is a local variable requiring Series access.
*/
func (r *ArrowIdentifierResolver) IsLocalVariable(identifierName string) bool {
	return r.accessResolver.IsLocalVariable(identifierName)
}

/*
IsParameter checks if identifier is a function parameter (scalar).
*/
func (r *ArrowIdentifierResolver) IsParameter(identifierName string) bool {
	return r.accessResolver.IsParameter(identifierName)
}

/*
ResolveBinaryExpression resolves all identifiers in a binary expression.
*/
func (r *ArrowIdentifierResolver) ResolveBinaryExpression(leftCode, operator, rightCode string) string {
	return fmt.Sprintf("(%s %s %s)", leftCode, operator, rightCode)
}
