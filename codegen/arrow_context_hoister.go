package codegen

import "fmt"

/*
ArrowContextHoister generates pre-loop ArrowContext declarations.

Design (SRP): Single purpose - code generation, delegate scanning and lifecycle
*/
type ArrowContextHoister struct {
	indentation string
}

func NewArrowContextHoister(indent string) *ArrowContextHoister {
	return &ArrowContextHoister{
		indentation: indent,
	}
}

func (h *ArrowContextHoister) GeneratePreLoopDeclarations(callSites []ArrowCallSite) string {
	if len(callSites) == 0 {
		return ""
	}

	code := ""

	for _, site := range callSites {
		code += h.generateSingleDeclaration(site)
	}

	return code
}

func (h *ArrowContextHoister) generateSingleDeclaration(site ArrowCallSite) string {
	code := h.indentation + fmt.Sprintf("%s := context.NewArrowContext(ctx)\n", site.ContextVar)
	if site.NeedsSecurity {
		code += h.generateSecurityBridge(site.ContextVar)
	}
	return code
}

func (h *ArrowContextHoister) generateSecurityBridge(contextVar string) string {
	code := h.indentation + fmt.Sprintf("%s.SecurityContexts = securityContexts\n", contextVar)
	code += h.indentation + fmt.Sprintf("for secKey, mapper := range securityBarMappers {\n")
	code += h.indentation + fmt.Sprintf("\t%s.SetBarMapper(secKey, mapper)\n", contextVar)
	code += h.indentation + "}\n"
	return code
}
