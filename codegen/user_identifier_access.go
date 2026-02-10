package codegen

import "fmt"

/* resolveUserIdentifierAccess resolves user identifier to series access code, respecting active context */
func (g *generator) resolveUserIdentifierAccess(identifierName string) string {
	if g.arrowAccessResolver != nil {
		if access, resolved := g.arrowAccessResolver.ResolveAccess(identifierName); resolved {
			return access
		}
	}

	return fmt.Sprintf("%sSeries.GetCurrent()", identifierName)
}
