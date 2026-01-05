package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// SecurityExpressionHandler generates code for security() expression evaluation
// Handles historical offset extraction and bar index adjustment
type SecurityExpressionHandler struct {
	indentFunc           func() string
	incrementIndent      func()
	decrementIndent      func()
	serializeExpr        func(ast.Expression) (string, error)
	markSecurityExprEval func()
}

type SecurityExpressionConfig struct {
	IndentFunc           func() string
	IncrementIndent      func()
	DecrementIndent      func()
	SerializeExpr        func(ast.Expression) (string, error)
	MarkSecurityExprEval func()
}

func NewSecurityExpressionHandler(config SecurityExpressionConfig) *SecurityExpressionHandler {
	return &SecurityExpressionHandler{
		indentFunc:           config.IndentFunc,
		incrementIndent:      config.IncrementIndent,
		decrementIndent:      config.DecrementIndent,
		serializeExpr:        config.SerializeExpr,
		markSecurityExprEval: config.MarkSecurityExprEval,
	}
}

// GenerateEvaluationCode produces code to evaluate expression in security context
// Handles patterns: close, pivothigh(), fixnan(pivothigh()[1])
// Historical offset extraction delegated to runtime StreamingRequest
func (h *SecurityExpressionHandler) GenerateEvaluationCode(
	varName string,
	exprArg ast.Expression,
	secBarIdxVar string,
) (string, error) {
	// Check for simple OHLCV field access
	if ident, ok := exprArg.(*ast.Identifier); ok {
		return h.generateOHLCVAccess(varName, ident, secBarIdxVar), nil
	}

	// Complex expression - delegate offset extraction to runtime
	code := ""

	// Generate evaluator initialization
	h.markSecurityExprEval()
	code += h.indentFunc() + "if secBarEvaluator == nil {\n"
	h.incrementIndent()
	code += h.indentFunc() + "secBarEvaluator = security.NewSeriesCachingEvaluator(security.NewStreamingBarEvaluator())\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	// Serialize expression for runtime evaluation (WITH offset if present)
	exprJSON, err := h.serializeExpr(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	// Generate EvaluateAtBar call - runtime will extract/apply offset
	code += h.indentFunc() + fmt.Sprintf("secValue, err := secBarEvaluator.EvaluateAtBar(%s, secCtx, %s)\n", exprJSON, secBarIdxVar)
	code += h.indentFunc() + "if err != nil {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	h.decrementIndent()
	code += h.indentFunc() + "} else {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	return code, nil
}

func (h *SecurityExpressionHandler) generateOHLCVAccess(varName string, ident *ast.Identifier, barIdxVar string) string {
	fieldName := ident.Name
	switch fieldName {
	case "close":
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(secCtx.Data[%s].Close)\n", varName, barIdxVar)
	case "open":
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(secCtx.Data[%s].Open)\n", varName, barIdxVar)
	case "high":
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(secCtx.Data[%s].High)\n", varName, barIdxVar)
	case "low":
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(secCtx.Data[%s].Low)\n", varName, barIdxVar)
	case "volume":
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(secCtx.Data[%s].Volume)\n", varName, barIdxVar)
	default:
		return h.indentFunc() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	}
}

func (h *SecurityExpressionHandler) extractHistoricalOffset(expr ast.Expression) (ast.Expression, int) {
	// Direct subscript: close[1]
	if memberExpr, ok := expr.(*ast.MemberExpression); ok {
		if offsetLit, ok := memberExpr.Property.(*ast.Literal); ok {
			if offsetVal, ok := offsetLit.Value.(float64); ok {
				return memberExpr.Object, int(offsetVal)
			}
		}
	}

	// Nested subscript: fixnan(pivothigh()[1])
	if callExpr, ok := expr.(*ast.CallExpression); ok {
		for i, arg := range callExpr.Arguments {
			if memberExpr, ok := arg.(*ast.MemberExpression); ok {
				if offsetLit, ok := memberExpr.Property.(*ast.Literal); ok {
					if offsetVal, ok := offsetLit.Value.(float64); ok {
						// Rebuild call with inner expression (without subscript)
						newArgs := make([]ast.Expression, len(callExpr.Arguments))
						copy(newArgs, callExpr.Arguments)
						newArgs[i] = memberExpr.Object

						newCall := &ast.CallExpression{
							Callee:    callExpr.Callee,
							Arguments: newArgs,
						}
						return newCall, int(offsetVal)
					}
				}
			}
		}
	}

	return expr, 0
}
