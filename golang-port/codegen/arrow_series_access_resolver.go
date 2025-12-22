package codegen

/*
ArrowSeriesAccessResolver determines how to access identifiers in arrow function context.

Resolution rules:
 1. Function parameters (analyzed from signature) → direct access (len, mult)
 2. Local variables (declared in function body) → Series access (upSeries.GetCurrent())
 3. Builtin identifiers → delegated to builtin handler

This implements the universal ForwardSeriesBuffer paradigm for arrow functions.
*/
type ArrowSeriesAccessResolver struct {
	localVariables map[string]bool // Variables declared in arrow function
	parameters     map[string]bool // Function parameters (scalars)
}

func NewArrowSeriesAccessResolver() *ArrowSeriesAccessResolver {
	return &ArrowSeriesAccessResolver{
		localVariables: make(map[string]bool),
		parameters:     make(map[string]bool),
	}
}

/*
RegisterLocalVariable marks a variable as local (needs Series access).
*/
func (r *ArrowSeriesAccessResolver) RegisterLocalVariable(varName string) {
	r.localVariables[varName] = true
}

/*
RegisterParameter marks an identifier as a function parameter (scalar access).
*/
func (r *ArrowSeriesAccessResolver) RegisterParameter(paramName string) {
	r.parameters[paramName] = true
}

/*
ResolveAccess determines the correct access pattern for an identifier.

Returns:
  - "upSeries.GetCurrent()" for local variables
  - "len" for parameters
  - "", false if identifier is not registered (delegate to builtin handler)
*/
func (r *ArrowSeriesAccessResolver) ResolveAccess(identifierName string) (string, bool) {
	if r.parameters[identifierName] {
		// Function parameter - direct scalar access
		return identifierName, true
	}

	if r.localVariables[identifierName] {
		// Local variable - Series access
		return identifierName + "Series.GetCurrent()", true
	}

	// Not found - delegate to caller (probably builtin)
	return "", false
}

/*
IsLocalVariable checks if identifier is a local variable.
*/
func (r *ArrowSeriesAccessResolver) IsLocalVariable(identifierName string) bool {
	return r.localVariables[identifierName]
}

/*
IsParameter checks if identifier is a function parameter.
*/
func (r *ArrowSeriesAccessResolver) IsParameter(identifierName string) bool {
	return r.parameters[identifierName]
}
