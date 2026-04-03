package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/* TupleIndicatorHandler orchestrates tuple indicator code generation.
 *
 * Two dispatch paths — checked in order:
 *   1. customHandlers  — stateful ForwardSeriesBuffer indicators (ta.kc, ta.supertrend)
 *   2. registry        — windowed-array runtime delegation (ta.macd, ta.bb, ta.stoch, ta.dmi)
 */
type TupleIndicatorHandler struct {
	customHandlers []CustomTupleFunctionHandler
	registry       *TupleIndicatorRegistry
	generator      *TupleIndicatorCodeGenerator
}

func NewTupleIndicatorHandler() *TupleIndicatorHandler {
	h := &TupleIndicatorHandler{
		registry:  sharedTupleIndicatorRegistry,
		generator: NewTupleIndicatorCodeGenerator(),
	}
	h.customHandlers = []CustomTupleFunctionHandler{
		&KcHandler{},
		&SupertrendHandler{},
	}
	return h
}

func (h *TupleIndicatorHandler) CanHandle(funcName string) bool {
	for _, ch := range h.customHandlers {
		if ch.CanHandle(funcName) {
			return true
		}
	}
	return h.registry.IsRegistered(funcName)
}

func (h *TupleIndicatorHandler) GenerateTupleCode(
	g *generator,
	varNames []string,
	call *ast.CallExpression,
) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	for _, ch := range h.customHandlers {
		if ch.CanHandle(funcName) {
			return ch.GenerateTupleCode(g, varNames, call)
		}
	}

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

// Used by the generator to collect tupleTAFunctions for internal series lifecycle.
func (h *TupleIndicatorHandler) IsCustomHandler(funcName string) bool {
	for _, ch := range h.customHandlers {
		if ch.CanHandle(funcName) {
			return true
		}
	}
	return false
}

// firstOutputVar is the naming prefix derived from the first output variable name.
func (h *TupleIndicatorHandler) InternalSeriesNamesFor(funcName, firstOutputVar string, call *ast.CallExpression) []string {
	for _, ch := range h.customHandlers {
		if ch.CanHandle(funcName) {
			return ch.InternalSeriesNames(firstOutputVar, call)
		}
	}
	return nil
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
