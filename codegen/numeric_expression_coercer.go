package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
	NumericExpressionCoercer converts boolean-producing expressions to float64 (1.0/0.0).

Centralizes the bool→float64 IIFE pattern used across AEG, arrow, and variable declarations.
*/
type NumericExpressionCoercer struct {
	boolDetector *BooleanConverter
}

func NewNumericExpressionCoercer(boolDetector *BooleanConverter) *NumericExpressionCoercer {
	return &NumericExpressionCoercer{boolDetector: boolDetector}
}

/*
	CoerceToFloat64 ensures generatedCode evaluates to float64.

Bool literals → "1.0"/"0.0". Bool expressions → IIFE wrap. Everything else → pass-through.
*/
func (c *NumericExpressionCoercer) CoerceToFloat64(expr ast.Expression, generatedCode string) string {
	if result, ok := c.tryCoerceBoolLiteral(expr); ok {
		return result
	}

	if c.boolDetector.IsAlreadyBoolean(expr) {
		return c.wrapBoolExprAsFloat64(generatedCode)
	}

	return generatedCode
}

func (c *NumericExpressionCoercer) tryCoerceBoolLiteral(expr ast.Expression) (string, bool) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return "", false
	}

	boolVal, ok := lit.Value.(bool)
	if !ok {
		return "", false
	}

	if boolVal {
		return "1.0", true
	}
	return "0.0", true
}

func (c *NumericExpressionCoercer) wrapBoolExprAsFloat64(code string) string {
	return fmt.Sprintf("func() float64 { if %s { return 1.0 } else { return 0.0 } }()", code)
}
