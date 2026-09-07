package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// evaluateTimeAtBar uses secCtx.PeriodAnchor for period alignment so that
// ta.change(time(tf)) boundaries follow the secondary exchange's session open
// rather than UTC midnight.
func evaluateTimeAtBar(e *StreamingBarEvaluator, call *ast.CallExpression, secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return math.NaN(), nil
	}
	barTimeSec := secCtx.Data[barIdx].Time

	if len(call.Arguments) == 0 {
		return float64(barTimeSec * 1000), nil
	}

	tf, err := resolveTimeframeArg(call.Arguments[0], secCtx)
	if err != nil {
		return math.NaN(), err
	}

	aligned := context.BarOpenTimeAtTimeframeWithAnchor(barTimeSec, tf, secCtx.PeriodAnchor)
	return float64(aligned * 1000), nil
}

func resolveTimeframeArg(expr ast.Expression, secCtx *context.Context) (string, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		switch v := e.Value.(type) {
		case string:
			return v, nil
		case float64:
			return fmt.Sprintf("%d", int(v)), nil
		}
	case *ast.MemberExpression:
		if isTimeframePeriodExpr(e) {
			return secCtx.Timeframe, nil
		}
	}
	return "", newUnsupportedExpressionError(expr)
}

func isTimeframePeriodExpr(expr *ast.MemberExpression) bool {
	obj, okObj := expr.Object.(*ast.Identifier)
	prop, okProp := expr.Property.(*ast.Identifier)
	return okObj && okProp && obj.Name == "timeframe" && prop.Name == "period"
}

func init() {
	registerCallHandlerAliases(evaluateTimeAtBar, "time", "ta.time")
}
