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
	return IsSecurityFunction(funcName)
}

func (h *SecurityInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	if len(expr.Arguments) < 3 {
		return "(func() float64 { return math.NaN() }())", nil
	}

	argExtractor := NewSecurityArgumentExtractor(g)

	symbolResult, err := argExtractor.ExtractSymbol(expr.Arguments[0])
	if err != nil {
		return "(func() float64 { return math.NaN() }())", nil
	}

	timeframeResult, err := argExtractor.ExtractTimeframe(expr.Arguments[1])
	if err != nil {
		return "(func() float64 { return math.NaN() }())", nil
	}

	expressionArg := expr.Arguments[2]
	lookahead := resolveSecurityLookahead(expr, g.pineVersion)

	g.hasSecurityCalls = true

	return h.generateIIFE(symbolResult, timeframeResult, expressionArg, lookahead, g)
}

func (h *SecurityInlineHandler) generateIIFE(symbolResult, timeframeResult *ExtractionResult, exprArg ast.Expression, lookahead bool, g *generator) (string, error) {
	keyBuilder := NewSecurityCacheKeyBuilder()
	keyComponents := keyBuilder.Build(symbolResult, timeframeResult)

	var iife strings.Builder
	iife.WriteString("(func() float64 {\n")

	if keyComponents.FormatArgs == "" {
		iife.WriteString(fmt.Sprintf("\t\tsecKey := %q\n", keyComponents.KeyPattern))
	} else {
		iife.WriteString(fmt.Sprintf("\t\tsecKey := fmt.Sprintf(%q, %s)\n", keyComponents.KeyPattern, keyComponents.FormatArgs))
	}
	iife.WriteString("\t\tsecCtx, secFound := securityContexts[secKey]\n")
	iife.WriteString("\t\tif !secFound { return math.NaN() }\n\n")

	iife.WriteString("\t\tsecurityBarMapper, mapperFound := securityBarMappers[secKey]\n")
	iife.WriteString("\t\tif !mapperFound { return math.NaN() }\n\n")

	iife.WriteString(fmt.Sprintf("\t\tsecLookahead := %v\n", lookahead))
	iife.WriteString(fmt.Sprintf("\t\tif %s == ctx.Timeframe { secLookahead = true }\n", timeframeResult.Code))
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

func (h *SecurityInlineHandler) generateExpressionEvaluation(exprArg ast.Expression, g *generator) (string, error) {
	switch expr := exprArg.(type) {
	case *ast.Identifier:
		if fieldExpr, ok := SecurityBarFieldExpression(expr.Name, "secCtx.Data[secBarIdx]"); ok {
			return "\t\treturn " + fieldExpr + "\n", nil
		}
		return "\t\treturn math.NaN()\n", nil
	case *ast.CallExpression, *ast.BinaryExpression, *ast.ConditionalExpression:
		return h.generateStreamingEvaluation(exprArg, g)
	default:
		return "\t\treturn math.NaN()\n", nil
	}
}

func (h *SecurityInlineHandler) generateStreamingEvaluation(exprArg ast.Expression, g *generator) (string, error) {
	g.hasSecurityExprEvals = true

	exprJSON, err := g.serializeExpressionForRuntime(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	var code strings.Builder

	initializer := NewSecurityEvaluatorInitializer(g.symbolTable, g)
	indentLevel := 0
	indentFunc := func() string {
		return strings.Repeat("\t", 2+indentLevel)
	}
	incrementIndent := func() { indentLevel++ }
	decrementIndent := func() { indentLevel-- }

	initCode := initializer.EmitInitialization(indentFunc, incrementIndent, decrementIndent)
	code.WriteString(initCode)

	code.WriteString(fmt.Sprintf("\t\tsecValue, err := secBarEvaluator.EvaluateAtBar(%s, secCtx, secBarIdx)\n", exprJSON))
	code.WriteString("\t\tif err != nil { return math.NaN() }\n")
	code.WriteString("\t\treturn secValue\n")

	return code.String(), nil
}
