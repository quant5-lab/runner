package preprocessor

import (
	"strings"

	"github.com/quant5-lab/runner/parser"
)

/* Transforms Ident → MemberAccess for namespace-qualified functions (math.max, ta.sma) */
type CalleeRewriter struct{}

func NewCalleeRewriter() *CalleeRewriter {
	return &CalleeRewriter{}
}

/* Rewrites: CallCallee{Ident:"max"} + "math.max" → CallCallee{MemberAccess{Object:"math", Property:"max"}} */
func (r *CalleeRewriter) Rewrite(callee *parser.CallCallee, qualifiedName string) bool {
	if callee == nil || callee.Ident == nil {
		return false
	}

	if !strings.Contains(qualifiedName, ".") {
		return false
	}

	parts := strings.SplitN(qualifiedName, ".", 2)
	if len(parts) != 2 {
		return false
	}

	callee.Ident = nil
	callee.MemberAccess = &parser.MemberAccess{
		Object:   parts[0],
		Property: parts[1],
	}

	return true
}

// RewriteIfMapped checks mapping and conditionally rewrites callee.
//
// Combines: (1) mapping lookup + (2) conditional rewrite
// Use case: namespace transformers (TA, Math, Request) apply mappings during traversal
//
// Example:
//
//	mappings := map[string]string{"max": "math.max", "min": "math.min"}
//	rewriter.RewriteIfMapped(call.Callee, "max", mappings)  // Transforms to math.max
func (r *CalleeRewriter) RewriteIfMapped(callee *parser.CallCallee, funcName string, mappings map[string]string) bool {
	if mappings == nil || callee == nil || callee.Ident == nil {
		return false
	}

	qualifiedName, exists := mappings[funcName]
	if !exists {
		return false
	}

	return r.Rewrite(callee, qualifiedName)
}
