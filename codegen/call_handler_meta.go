package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

// MetaFunctionHandler handles Pine Script meta functions.
//
// Handles: indicator(), strategy()
// Behavior: Extracts metadata and config values, produces no runtime code
type MetaFunctionHandler struct {
	configExtractor *StrategyConfigExtractor
}

// NewMetaFunctionHandler creates a handler.
func NewMetaFunctionHandler() *MetaFunctionHandler {
	return &MetaFunctionHandler{
		configExtractor: NewStrategyConfigExtractor(),
	}
}

func (h *MetaFunctionHandler) CanHandle(funcName string) bool {
	return funcName == "indicator" || funcName == "strategy"
}

func (h *MetaFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	/* Extract function name for validation */
	funcName := ""
	if id, ok := call.Callee.(*ast.Identifier); ok {
		funcName = id.Name
	} else if member, ok := call.Callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				funcName = obj.Name + "." + prop.Name
			}
		}
	}

	if !h.CanHandle(funcName) {
		return "", nil
	}

	extractedConfig := h.configExtractor.ExtractFromCall(call)
	g.strategyConfig.MergeFrom(extractedConfig)
	return "", nil
}
