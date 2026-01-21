package codegen

/* ArrowSeriesAccessResolver determines identifier access (parameters, local vars, builtins) in arrow functions */
type ArrowSeriesAccessResolver struct {
	localVariables map[string]bool // Variables declared in arrow function
	parameters     map[string]bool // Function parameters (scalars)
	loopModified   map[string]bool // Variables modified inside for-loops (use Series.GetCurrent())
}

func NewArrowSeriesAccessResolver() *ArrowSeriesAccessResolver {
	return &ArrowSeriesAccessResolver{
		localVariables: make(map[string]bool),
		parameters:     make(map[string]bool),
		loopModified:   make(map[string]bool),
	}
}

/* RegisterLocalVariable marks a variable as local (scalar access for current bar) */
func (r *ArrowSeriesAccessResolver) RegisterLocalVariable(varName string) {
	r.localVariables[varName] = true
}

/* RegisterParameter marks an identifier as a function parameter (scalar access) */
func (r *ArrowSeriesAccessResolver) RegisterParameter(paramName string) {
	r.parameters[paramName] = true
}

/* RegisterLoopModified marks a variable as modified in for-loop (Series access required) */
func (r *ArrowSeriesAccessResolver) RegisterLoopModified(varName string) {
	r.loopModified[varName] = true
}

/* ResolveAccess returns scalar access for parameters/local vars, delegates builtins to caller */
func (r *ArrowSeriesAccessResolver) ResolveAccess(identifierName string) (string, bool) {
	if r.parameters[identifierName] {
		// Function parameter - direct scalar access
		return identifierName, true
	}

	if r.loopModified[identifierName] {
		// Loop-modified variable - Series access required
		return identifierName + "Series.GetCurrent()", true
	}

	if r.localVariables[identifierName] {
		// Local variable - scalar access (current bar)
		return identifierName, true
	}

	// Not found - delegate to caller (probably builtin)
	return "", false
}

/* IsLocalVariable checks if identifier is a local variable */
func (r *ArrowSeriesAccessResolver) IsLocalVariable(identifierName string) bool {
	return r.localVariables[identifierName]
}

/* IsParameter checks if identifier is a function parameter */
func (r *ArrowSeriesAccessResolver) IsParameter(identifierName string) bool {
	return r.parameters[identifierName]
}
