package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// mathNamespaceConstants covers all compile-time constants in the math.* namespace.
var mathNamespaceConstants = map[string]float64{
	"pi":   math.Pi,
	"e":    math.E,
	"phi":  1.6180339887498948482,
	"rphi": 0.6180339887498948482,
	"huge": math.MaxFloat64,
	"tiny": math.SmallestNonzeroFloat64,
}

func evaluateMathLog(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.log")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	if val <= 0 {
		return math.NaN(), nil
	}
	return math.Log(val), nil
}

func evaluateMathLog10(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.log10")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	if val <= 0 {
		return math.NaN(), nil
	}
	return math.Log10(val), nil
}

func evaluateMathExp(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.exp")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Exp(val), nil
}

func evaluateMathSqrt(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.sqrt")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Sqrt(val), nil
}

func evaluateMathPow(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if len(call.Arguments) < 2 {
		return 0.0, newInsufficientArgumentsError("math.pow", 2, len(call.Arguments))
	}
	base, err := e.EvaluateAtBar(call.Arguments[0], secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	exp, err := e.EvaluateAtBar(call.Arguments[1], secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Pow(base, exp), nil
}

func evaluateMathRound(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.round")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Round(val), nil
}

// evaluateMathRoundToMintick returns the value unchanged: tick size is not available
// in the security evaluation context, so rounding to mintick is a no-op here.
func evaluateMathRoundToMintick(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.round_to_mintick")
	if err != nil {
		return 0.0, err
	}
	return e.EvaluateAtBar(expr, secCtx, barIdx)
}

func evaluateMathFloor(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.floor")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Floor(val), nil
}

func evaluateMathCeil(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.ceil")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Ceil(val), nil
}

func evaluateMathSign(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.sign")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	switch {
	case val > 0:
		return 1.0, nil
	case val < 0:
		return -1.0, nil
	default:
		return 0.0, nil
	}
}

func evaluateMathSin(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.sin")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Sin(val), nil
}

func evaluateMathCos(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.cos")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Cos(val), nil
}

func evaluateMathTan(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.tan")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Tan(val), nil
}

func evaluateMathAsin(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.asin")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Asin(val), nil
}

func evaluateMathAcos(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.acos")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Acos(val), nil
}

func evaluateMathAtan(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.atan")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return math.Atan(val), nil
}

func evaluateMathToRadians(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.toradians")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return val * math.Pi / 180.0, nil
}

func evaluateMathToDegrees(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	expr, err := extractSingleExpressionArgument(call, "math.todegrees")
	if err != nil {
		return 0.0, err
	}
	val, err := e.EvaluateAtBar(expr, secCtx, barIdx)
	if err != nil {
		return math.NaN(), err
	}
	return val * 180.0 / math.Pi, nil
}

func init() {
	registerMemberConstants("math", mathNamespaceConstants)

	registerCallHandler("math.log", evaluateMathLog)
	registerCallHandler("math.log10", evaluateMathLog10)
	registerCallHandler("math.exp", evaluateMathExp)
	registerCallHandler("math.sqrt", evaluateMathSqrt)
	registerCallHandler("math.pow", evaluateMathPow)
	registerCallHandler("math.round", evaluateMathRound)
	registerCallHandler("math.round_to_mintick", evaluateMathRoundToMintick)
	registerCallHandler("math.floor", evaluateMathFloor)
	registerCallHandler("math.ceil", evaluateMathCeil)
	registerCallHandler("math.sign", evaluateMathSign)
	registerCallHandler("math.sin", evaluateMathSin)
	registerCallHandler("math.cos", evaluateMathCos)
	registerCallHandler("math.tan", evaluateMathTan)
	registerCallHandler("math.asin", evaluateMathAsin)
	registerCallHandler("math.acos", evaluateMathAcos)
	registerCallHandler("math.atan", evaluateMathAtan)
	registerCallHandler("math.toradians", evaluateMathToRadians)
	registerCallHandler("math.todegrees", evaluateMathToDegrees)
}
