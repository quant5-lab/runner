package security

import (
	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// barCallHandler evaluates a named PineScript function call at a specific security bar.
//
// Receiver-first convention matches Go method expression signatures directly, so any
// (*StreamingBarEvaluator).methodName can be registered without a wrapping closure.
type barCallHandler func(*StreamingBarEvaluator, *ast.CallExpression, *context.Context, int) (float64, error)

var (
	callHandlerRegistry      = map[string]barCallHandler{}
	memberConstantNamespaces = map[string]map[string]float64{}
)

// registerCallHandler maps a PineScript function name to its bar-level evaluator.
func registerCallHandler(funcName string, h barCallHandler) {
	callHandlerRegistry[funcName] = h
}

// registerCallHandlerAliases maps multiple PineScript names to the same evaluator.
func registerCallHandlerAliases(h barCallHandler, funcNames ...string) {
	for _, name := range funcNames {
		callHandlerRegistry[name] = h
	}
}

// registerMemberConstants registers all named constant properties for a PineScript namespace.
func registerMemberConstants(namespace string, constants map[string]float64) {
	memberConstantNamespaces[namespace] = constants
}

func lookupMemberConstant(namespace, property string) (float64, bool) {
	if ns, ok := memberConstantNamespaces[namespace]; ok {
		v, ok := ns[property]
		return v, ok
	}
	return 0, false
}

func (e *StreamingBarEvaluator) evaluateTACallAtBar(call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	funcName := extractCallFunctionName(call.Callee)
	if h, ok := callHandlerRegistry[funcName]; ok {
		return h(e, call, secCtx, barIdx)
	}
	return 0.0, newUnsupportedFunctionError(funcName)
}
