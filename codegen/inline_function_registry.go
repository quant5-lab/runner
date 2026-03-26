package codegen

// InlineFunctionRegistry identifies functions that generate inline code
// rather than creating temp variables with Series storage.
//
// Inline functions compute values on-demand within the bar loop,
// while Series functions pre-compute and store historical values.
type InlineFunctionRegistry struct {
	inlineFunctions map[string]bool
}

// NewInlineFunctionRegistry creates registry with known inline-only functions
func NewInlineFunctionRegistry() *InlineFunctionRegistry {
	return &InlineFunctionRegistry{
		inlineFunctions: map[string]bool{},
	}
}

// IsInlineOnly returns true if function generates inline code only
func (r *InlineFunctionRegistry) IsInlineOnly(funcName string) bool {
	return r.inlineFunctions[funcName]
}

// Register adds function to inline-only registry
func (r *InlineFunctionRegistry) Register(funcName string) {
	r.inlineFunctions[funcName] = true
}
