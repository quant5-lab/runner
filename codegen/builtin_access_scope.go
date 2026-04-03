package codegen

type AccessScope int

const (
	BarLoopScope AccessScope = iota
	SecurityScope
	ArrowScope
)

func ScopeFromSecurityFlag(inSecurityContext bool) AccessScope {
	if inSecurityContext {
		return SecurityScope
	}
	return BarLoopScope
}
