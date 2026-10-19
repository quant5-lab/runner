package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func isComparisonOperator(op string) bool {
	switch op {
	case ">", "<", ">=", "<=", "==", "!=":
		return true
	}
	return false
}

func liftComparisonToFloat64(expr ast.Expression, code string) string {
	bin, ok := expr.(*ast.BinaryExpression)
	if !ok || !isComparisonOperator(bin.Operator) {
		return code
	}
	return fmt.Sprintf("func() float64 { if %s { return 1.0 } else { return 0.0 } }()", code)
}
