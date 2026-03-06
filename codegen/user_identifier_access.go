package codegen

import "fmt"

/* resolveUserIdentifierAccess resolves user identifier to series access code, respecting active context */
func (g *generator) resolveUserIdentifierAccess(identifierName string) string {
	// For-loop counters inside arrow functions are int variables; float64-cast for arithmetic compatibility
	if g.arrowAccessResolver != nil && g.loopContextStack != nil && g.loopContextStack.IsLoopCounter(identifierName) {
		return fmt.Sprintf("float64(%s)", identifierName)
	}

	if g.arrowAccessResolver != nil {
		if access, resolved := g.arrowAccessResolver.ResolveAccess(identifierName); resolved {
			return access
		}
	}

	return fmt.Sprintf("%sSeries.GetCurrent()", identifierName)
}
