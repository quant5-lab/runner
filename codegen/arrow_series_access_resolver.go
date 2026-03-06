package codegen

/* ArrowSeriesAccessResolver determines identifier access (parameters, local vars, builtins) in arrow functions */
type ArrowSeriesAccessResolver struct {
	localVariables   map[string]bool // Variables declared in arrow function
	parameters       map[string]bool // Scalar function parameters (float64)
	seriesParameters map[string]bool // Series-typed function parameters (*series.Series)
	loopModified     map[string]bool // Variables modified inside for-loops (use Series.GetCurrent())
}

func NewArrowSeriesAccessResolver() *ArrowSeriesAccessResolver {
	return &ArrowSeriesAccessResolver{
		localVariables:   make(map[string]bool),
		parameters:       make(map[string]bool),
		seriesParameters: make(map[string]bool),
		loopModified:     make(map[string]bool),
	}
}

/* RegisterLocalVariable marks a variable as local (scalar access for current bar) */
func (r *ArrowSeriesAccessResolver) RegisterLocalVariable(varName string) {
	r.localVariables[varName] = true
}

/* RegisterParameter marks an identifier as a scalar function parameter (float64) */
func (r *ArrowSeriesAccessResolver) RegisterParameter(paramName string) {
	r.parameters[paramName] = true
}

/* RegisterSeriesParameter marks an identifier as a series-typed function parameter (*series.Series) */
func (r *ArrowSeriesAccessResolver) RegisterSeriesParameter(paramName string) {
	r.seriesParameters[paramName] = true
}

/* RegisterLoopModified marks a variable as modified in for-loop (Series access required) */
func (r *ArrowSeriesAccessResolver) RegisterLoopModified(varName string) {
	r.loopModified[varName] = true
}

/* ResolveAccess returns the correct Go access expression for the identifier in arrow context.
 * Resolution priority: series params → scalar params → loop-modified → local vars → not found (delegate to caller) */
func (r *ArrowSeriesAccessResolver) ResolveAccess(identifierName string) (string, bool) {
	if r.seriesParameters[identifierName] {
		return identifierName + "Series.GetCurrent()", true
	}
	if r.parameters[identifierName] {
		return identifierName, true
	}
	if r.loopModified[identifierName] {
		return identifierName + "Series.GetCurrent()", true
	}
	if r.localVariables[identifierName] {
		return identifierName, true
	}
	return "", false
}

/* IsSeriesParameter checks if identifier is a series-typed function parameter */
func (r *ArrowSeriesAccessResolver) IsSeriesParameter(identifierName string) bool {
	return r.seriesParameters[identifierName]
}

/* IsLocalVariable checks if identifier is a local variable */
func (r *ArrowSeriesAccessResolver) IsLocalVariable(identifierName string) bool {
	return r.localVariables[identifierName]
}

/* IsParameter checks if identifier is any kind of function parameter (scalar or series) */
func (r *ArrowSeriesAccessResolver) IsParameter(identifierName string) bool {
	return r.parameters[identifierName] || r.seriesParameters[identifierName]
}
