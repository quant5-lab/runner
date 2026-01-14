package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* DevInlineHandler generates inline expressions for ta.dev (mean absolute deviation) */
type DevInlineHandler struct{}

func NewDevInlineHandler() *DevInlineHandler {
	return &DevInlineHandler{}
}

func (h *DevInlineHandler) CanHandle(funcName string) bool {
	return funcName == "ta.dev" || funcName == "dev"
}

func (h *DevInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	if len(expr.Arguments) < 2 {
		return "", fmt.Errorf("dev requires 2 arguments (source, length)")
	}

	sourceExpr := g.extractSeriesExpression(expr.Arguments[0])
	lengthExpr := g.extractSeriesExpression(expr.Arguments[1])

	// Convert sourceExpr from GetCurrent() to Get(j) for loop context
	sourceAccessInLoop := g.convertSeriesAccessToOffset(sourceExpr, "j")

	/* Generate two-pass algorithm: 1) calculate mean, 2) calculate mean absolute deviation
	 * Returns NaN if not enough bars (ctx.BarIndex < length-1)
	 * ForwardSeriesBuffer: Uses Series.Get(j) for historical access within loop
	 */
	return fmt.Sprintf("(func() float64 { length := int(%s); if ctx.BarIndex < length-1 { return math.NaN() }; sum := 0.0; for j := 0; j < length; j++ { sum += %s }; mean := sum / float64(length); devSum := 0.0; for j := 0; j < length; j++ { devSum += math.Abs(%s - mean) }; return devSum / float64(length) }())",
		lengthExpr, sourceAccessInLoop, sourceAccessInLoop), nil
}
