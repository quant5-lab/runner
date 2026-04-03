package codegen

import (
	"fmt"
	"strings"
)

type ExpressionNormalizer struct{}

func NewExpressionNormalizer() *ExpressionNormalizer {
	return &ExpressionNormalizer{}
}

func (n *ExpressionNormalizer) NormalizeForSeriesStorage(exprCode string) string {
	if isSimpleIntegerLiteral(exprCode) {
		return fmt.Sprintf("float64(%s)", exprCode)
	}
	return exprCode
}

func isSimpleIntegerLiteral(expr string) bool {
	if strings.Contains(expr, ".") || strings.Contains(expr, "(") {
		return false
	}
	if len(expr) == 0 {
		return false
	}
	for i, ch := range expr {
		if i == 0 && ch == '-' {
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
