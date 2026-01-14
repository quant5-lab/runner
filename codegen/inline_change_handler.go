package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ChangeInlineHandler generates inline expressions for ta.change (difference from N bars ago) */
type ChangeInlineHandler struct{}

func NewChangeInlineHandler() *ChangeInlineHandler {
	return &ChangeInlineHandler{}
}

func (h *ChangeInlineHandler) CanHandle(funcName string) bool {
	return funcName == "ta.change" || funcName == "change"
}

func (h *ChangeInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	if len(expr.Arguments) < 1 {
		return "", fmt.Errorf("ta.change requires at least 1 argument")
	}

	/* Extract offset (default 1 if not specified) */
	offset := 1
	if len(expr.Arguments) >= 2 {
		if lit, ok := expr.Arguments[1].(*ast.Literal); ok {
			switch v := lit.Value.(type) {
			case float64:
				offset = int(v)
			case int:
				offset = v
			}
		}
	}

	sourceExpr := g.extractSeriesExpression(expr.Arguments[0])
	currentVal := sourceExpr
	prevVal := g.convertSeriesAccessToIntOffset(sourceExpr, offset)

	/* Generate IIFE that returns current - previous, or NaN if not enough bars */
	return fmt.Sprintf("(func() float64 { if ctx.BarIndex < %d { return math.NaN() }; return %s - %s }())",
		offset, currentVal, prevVal), nil
}
