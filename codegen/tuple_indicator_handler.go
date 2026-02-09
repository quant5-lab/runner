package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/* TupleIndicatorHandler orchestrates tuple indicator code generation */
type TupleIndicatorHandler struct {
	registry  *TupleIndicatorRegistry
	generator *TupleIndicatorCodeGenerator
}

func NewTupleIndicatorHandler() *TupleIndicatorHandler {
	return &TupleIndicatorHandler{
		registry:  sharedTupleIndicatorRegistry,
		generator: NewTupleIndicatorCodeGenerator(),
	}
}

func (h *TupleIndicatorHandler) CanHandle(funcName string) bool {
	return h.registry.IsRegistered(funcName)
}

func (h *TupleIndicatorHandler) GenerateTupleCode(
	g *generator,
	varNames []string,
	call *ast.CallExpression,
) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	spec := h.registry.Lookup(funcName)
	if spec == nil {
		return "", errIndicatorNotRegistered(funcName)
	}

	ctx := h.buildCodeGenContext(g)

	return h.generator.Generate(
		spec,
		varNames,
		call,
		ctx,
		g.extractSeriesExpression,
		g.constants,
	)
}

func (h *TupleIndicatorHandler) buildCodeGenContext(g *generator) CodeGenContext {
	return CodeGenContext{
		Indenter:    g.ind,
		IndentLevel: &g.indent,
		IncreaseIndent: func() {
			g.indent++
		},
		DecreaseIndent: func() {
			g.indent--
		},
	}
}
