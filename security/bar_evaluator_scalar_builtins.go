package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

func evaluateNz(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) == 0 {
		return 0.0, newInsufficientArgumentsError("nz", 1, 0)
	}
	val, err := e.EvaluateAtBar(call.Arguments[0], secCtx, barIdx)
	if err != nil {
		return 0.0, err
	}
	if !math.IsNaN(val) {
		return val, nil
	}
	if len(call.Arguments) >= 2 {
		return e.EvaluateAtBar(call.Arguments[1], secCtx, barIdx)
	}
	return 0.0, nil
}

// The standalone identifier `na` is handled separately in evaluateIdentifierAtBar.
func evaluateNaFunc(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "na")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return 1.0, nil
	}
	if math.IsNaN(val) {
		return 1.0, nil
	}
	return 0.0, nil
}

// Truncates toward zero (not toward -∞), matching PineScript int() cast semantics.
func evaluateInt(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "int")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Trunc(val), nil
}

func evaluateFloat(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "float")
	if err != nil {
		return 0.0, err
	}
	return e.EvaluateAtBar(expr, secCtx, barIdx)
}

func init() {
	registerCallHandlerAliases(evaluateNz, "nz", "ta.nz")
	registerCallHandlerAliases(evaluateNaFunc, "na", "ta.na")
	// "int_" and "ta.int_" are the Go-sanitized forms produced by
	// preprocessor.IdentifierSanitizer (int is a Go builtin type).
	registerCallHandlerAliases(evaluateInt, "int", "ta.int", "int_", "ta.int_")
	registerCallHandlerAliases(evaluateFloat, "float", "ta.float")
}
