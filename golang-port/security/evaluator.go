package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ta"
)

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

func evaluateMemberExpression(mem *ast.MemberExpression, secCtx *context.Context) ([]float64, error) {
	return nil, fmt.Errorf("member expression not directly evaluable: %v", mem)
}

func evaluateCallExpression(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
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

func evaluateTASma(call *ast.CallExpression, secCtx *context.Context) ([]float64, error) {
	if len(call.Arguments) < 2 {
		return nil, fmt.Errorf("ta.sma requires 2 arguments")
	}

	sourceValues, err := EvaluateExpression(call.Arguments[0], secCtx)
	if err != nil {
		return nil, fmt.Errorf("ta.sma source: %w", err)
	}

	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("ta.sma period: %w", err)
	}

	smaValues := ta.Sma(sourceValues, int(period))
	return smaValues, nil
}

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

	emaValues := ta.Ema(sourceValues, int(period))
	return emaValues, nil
}

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

	rmaValues := ta.Rma(sourceValues, int(period))
	return rmaValues, nil
}

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

	rsiValues := ta.Rsi(sourceValues, int(period))
	return rsiValues, nil
}

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

func allNaN(values []float64) bool {
	for _, v := range values {
		if !math.IsNaN(v) {
			return false
		}
	}
	return true
}
