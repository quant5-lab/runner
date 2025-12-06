package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ta"
)

/* EvaluateExpression calculates expression values in security context */
func EvaluateExpression(expr ast.Expression, secCtx *context.Context) ([]float64, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		return evaluateIdentifier(e, secCtx)
	case *ast.CallExpression:
		return evaluateCallExpression(e, secCtx)
	case *ast.MemberExpression:
		return evaluateMemberExpression(e, secCtx)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

/* evaluateIdentifier handles close, open, high, low, volume */
func evaluateIdentifier(id *ast.Identifier, secCtx *context.Context) ([]float64, error) {
	values := make([]float64, len(secCtx.Data))

	switch id.Name {
	case "close":
		for i, bar := range secCtx.Data {
			values[i] = bar.Close
		}
	case "open":
		for i, bar := range secCtx.Data {
			values[i] = bar.Open
		}
	case "high":
		for i, bar := range secCtx.Data {
			values[i] = bar.High
		}
	case "low":
		for i, bar := range secCtx.Data {
			values[i] = bar.Low
		}
	case "volume":
		for i, bar := range secCtx.Data {
			values[i] = bar.Volume
		}
	default:
		return nil, fmt.Errorf("unknown identifier: %s", id.Name)
	}

	return values, nil
}

/* evaluateMemberExpression handles ta.sma, ta.ema, etc */
func evaluateMemberExpression(mem *ast.MemberExpression, secCtx *context.Context) ([]float64, error) {
	/* For now, only used in CallExpression context */
	return nil, fmt.Errorf("member expression not directly evaluable: %v", mem)
}

/* evaluateCallExpression handles ta.sma(close, 20), ta.ema(...), etc */
func evaluateCallExpression(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	/* Extract function name */
	funcName := extractCallFunctionName(call.Callee)

	switch funcName {
	case "ta.sma":
		return evaluateTASma(call, secCtx)
	case "ta.ema":
		return evaluateTAEma(call, secCtx)
	case "ta.rma":
		return evaluateTARma(call, secCtx)
	case "ta.rsi":
		return evaluateTARsi(call, secCtx)
	default:
		return nil, fmt.Errorf("unsupported function: %s", funcName)
	}
}

/* extractCallFunctionName gets function name from callee */
func extractCallFunctionName(callee ast.Expression) string {
	if mem, ok := callee.(*ast.MemberExpression); ok {
		obj := ""
		if id, ok := mem.Object.(*ast.Identifier); ok {
			obj = id.Name
		}
		prop := ""
		if id, ok := mem.Property.(*ast.Identifier); ok {
			prop = id.Name
		}
		return obj + "." + prop
	}

	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name
	}

	return ""
}

/* evaluateTASma evaluates ta.sma(source, period) */
func evaluateTASma(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("ta.sma requires 2 arguments")
	}

	/* Evaluate source (close, high, etc) */
	sourceValues, err := EvaluateExpression(call.Arguments[0], secCtx)
	if err != nil {
		return nil, fmt.Errorf("ta.sma source: %w", err)
	}

	/* Extract period (literal number) */
	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("ta.sma period: %w", err)
	}

	/* Calculate SMA - ta package expects []float64 */
	smaValues := ta.Sma(sourceValues, int(period))
	return smaValues, nil
}

/* evaluateTAEma evaluates ta.ema(source, period) */
func evaluateTAEma(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("ta.ema requires 2 arguments")
	}

	sourceValues, err := EvaluateExpression(call.Arguments[0], secCtx)
	if err != nil {
		return nil, fmt.Errorf("ta.ema source: %w", err)
	}

	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("ta.ema period: %w", err)
	}

	/* Calculate EMA - ta package expects []float64 */
	emaValues := ta.Ema(sourceValues, int(period))
	return emaValues, nil
}

/* evaluateTARma evaluates ta.rma(source, period) */
func evaluateTARma(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("ta.rma requires 2 arguments")
	}

	sourceValues, err := EvaluateExpression(call.Arguments[0], secCtx)
	if err != nil {
		return nil, fmt.Errorf("ta.rma source: %w", err)
	}

	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("ta.rma period: %w", err)
	}

	/* Calculate RMA - ta package expects []float64 */
	rmaValues := ta.Rma(sourceValues, int(period))
	return rmaValues, nil
}

/* evaluateTARsi evaluates ta.rsi(source, period) */
func evaluateTARsi(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("ta.rsi requires 2 arguments")
	}

	sourceValues, err := EvaluateExpression(call.Arguments[0], secCtx)
	if err != nil {
		return nil, fmt.Errorf("ta.rsi source: %w", err)
	}

	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("ta.rsi period: %w", err)
	}

	/* Calculate RSI - ta package expects []float64 */
	rsiValues := ta.Rsi(sourceValues, int(period))
	return rsiValues, nil
}

/* extractNumberLiteral extracts number from literal expression */
func extractNumberLiteral(expr ast.Expression) (float64, error) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return 0, fmt.Errorf("expected literal, got %T", expr)
	}

	switch v := lit.Value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("expected number literal, got %T", v)
	}
}

/* Helper: check if all values are NaN (warmup period) */
func allNaN(values []float64) bool {
	for _, v := range values {
		if !math.IsNaN(v) {
			return false
		}
	}
	return true
}
