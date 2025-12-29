package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* Generates IIFE for security() in conditionals/ternaries */
type SecurityInlineHandler struct{}

func NewSecurityInlineHandler() *SecurityInlineHandler {
	return &SecurityInlineHandler{}
}

func (h *SecurityInlineHandler) CanHandle(funcName string) bool {
	return funcName == "request.security" || funcName == "security"
}

func (h *SecurityInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	if len(expr.Arguments) < 3 {
		return "(func() float64 { return math.NaN() }())", nil
	}

	symbolExpr := expr.Arguments[0]
	timeframeExpr := expr.Arguments[1]
	expressionArg := expr.Arguments[2]

	symbolCode := h.extractSymbol(symbolExpr)
	timeframe := h.extractTimeframe(timeframeExpr)
	lookahead := h.extractLookahead(expr.Arguments)

	if symbolCode == "" || timeframe == "" {
		return "(func() float64 { return math.NaN() }())", nil
	}

	g.hasSecurityCalls = true

	return h.generateIIFE(symbolCode, timeframe, expressionArg, lookahead, g)
}

func (h *SecurityInlineHandler) extractSymbol(symbolExpr ast.Expression) string {
	switch expr := symbolExpr.(type) {
	case *ast.Identifier:
		if expr.Name == "tickerid" {
			return "ctx.Symbol"
		}
		return fmt.Sprintf("%q", expr.Name)
	case *ast.MemberExpression:
		return "ctx.Symbol"
	case *ast.Literal:
		if s, ok := expr.Value.(string); ok {
			return fmt.Sprintf("%q", s)
		}
	}
	return ""
}

func (h *SecurityInlineHandler) extractTimeframe(timeframeExpr ast.Expression) string {
	lit, ok := timeframeExpr.(*ast.Literal)
	if !ok {
		return ""
	}

	s, ok := lit.Value.(string)
	if !ok {
		return ""
	}

	tf := strings.Trim(s, "'\"")
	return h.normalizeTimeframe(tf)
}

func (h *SecurityInlineHandler) normalizeTimeframe(tf string) string {
	switch tf {
	case "D":
		return "1D"
	case "W":
		return "1W"
	case "M":
		return "1M"
	default:
		return tf
	}
}

func (h *SecurityInlineHandler) extractLookahead(args []ast.Expression) bool {
	if len(args) < 4 {
		return false
	}

	resolver := NewConstantResolver()
	fourthArg := args[3]

	if objExpr, ok := fourthArg.(*ast.ObjectExpression); ok {
		for _, prop := range objExpr.Properties {
			if keyIdent, ok := prop.Key.(*ast.Identifier); ok && keyIdent.Name == "lookahead" {
				if resolved, ok := resolver.ResolveToBool(prop.Value); ok {
					return resolved
				}
				break
			}
		}
	} else {
		if resolved, ok := resolver.ResolveToBool(fourthArg); ok {
			return resolved
		}
	}

	return false
}

func (h *SecurityInlineHandler) generateIIFE(symbolCode, timeframe string, exprArg ast.Expression, lookahead bool, g *generator) (string, error) {
	cacheKeyPattern := h.buildCacheKeyPattern(symbolCode, timeframe)

	var iife strings.Builder
	iife.WriteString("(func() float64 {\n")

	iife.WriteString(fmt.Sprintf("\t\tsecKey := fmt.Sprintf(%q, %s)\n", cacheKeyPattern, symbolCode))
	iife.WriteString("\t\tsecCtx, secFound := securityContexts[secKey]\n")
	iife.WriteString("\t\tif !secFound { return math.NaN() }\n\n")

	iife.WriteString("\t\tsecurityBarMapper, mapperFound := securityBarMappers[secKey]\n")
	iife.WriteString("\t\tif !mapperFound { return math.NaN() }\n\n")

	iife.WriteString(fmt.Sprintf("\t\tsecLookahead := %v\n", lookahead))
	iife.WriteString(fmt.Sprintf("\t\tif %q == ctx.Timeframe { secLookahead = true }\n", timeframe))
	iife.WriteString("\t\tsecBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)\n")
	iife.WriteString("\t\tif secBarIdx < 0 { return math.NaN() }\n\n")

	evaluationCode, err := h.generateExpressionEvaluation(exprArg, g)
	if err != nil {
		return "", err
	}
	iife.WriteString(evaluationCode)

	iife.WriteString("\t}())")

	return iife.String(), nil
}

func (h *SecurityInlineHandler) buildCacheKeyPattern(symbolCode, timeframe string) string {
	if symbolCode == "ctx.Symbol" {
		return fmt.Sprintf("%%s:%s", timeframe)
	}
	return fmt.Sprintf("%s:%s", strings.Trim(symbolCode, `"`), timeframe)
}

func (h *SecurityInlineHandler) generateExpressionEvaluation(exprArg ast.Expression, g *generator) (string, error) {
	switch expr := exprArg.(type) {
	case *ast.Identifier:
		return h.generateOHLCVAccess(expr.Name), nil
	case *ast.CallExpression, *ast.BinaryExpression, *ast.ConditionalExpression:
		return h.generateStreamingEvaluation(exprArg, g)
	default:
		return "\t\treturn math.NaN()\n", nil
	}
}

func (h *SecurityInlineHandler) generateOHLCVAccess(fieldName string) string {
	switch fieldName {
	case "close":
		return "\t\treturn secCtx.Data[secBarIdx].Close\n"
	case "open":
		return "\t\treturn secCtx.Data[secBarIdx].Open\n"
	case "high":
		return "\t\treturn secCtx.Data[secBarIdx].High\n"
	case "low":
		return "\t\treturn secCtx.Data[secBarIdx].Low\n"
	case "volume":
		return "\t\treturn secCtx.Data[secBarIdx].Volume\n"
	default:
		return "\t\treturn math.NaN()\n"
	}
}

func (h *SecurityInlineHandler) generateStreamingEvaluation(exprArg ast.Expression, g *generator) (string, error) {
	g.hasSecurityExprEvals = true

	exprJSON, err := g.serializeExpressionForRuntime(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	var code strings.Builder
	code.WriteString("\t\tif secBarEvaluator == nil {\n")
	code.WriteString("\t\t\tsecBarEvaluator = security.NewStreamingBarEvaluator()\n")
	code.WriteString("\t\t}\n")
	code.WriteString(fmt.Sprintf("\t\tsecValue, err := secBarEvaluator.EvaluateAtBar(%s, secCtx, secBarIdx)\n", exprJSON))
	code.WriteString("\t\tif err != nil { return math.NaN() }\n")
	code.WriteString("\t\treturn secValue\n")

	return code.String(), nil
}
