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
	return h.indentation + fmt.Sprintf("%s := context.NewArrowContext(ctx)\n", site.ContextVar)
}
