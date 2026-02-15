package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/*
	ArrowSecurityCallGenerator produces IIFE code for security() calls inside arrow functions.

Accesses security data through ArrowContext bridge instead of direct scope variables.
*/
type ArrowSecurityCallGenerator struct {
	gen *generator
}

func NewArrowSecurityCallGenerator(gen *generator) *ArrowSecurityCallGenerator {
	return &ArrowSecurityCallGenerator{gen: gen}
}

func (g *ArrowSecurityCallGenerator) CanHandle(call *ast.CallExpression) bool {
	return isSecurityCallExpression(call)
}

func (g *ArrowSecurityCallGenerator) Generate(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "(func() float64 { return math.NaN() }())", nil
	}

	argExtractor := NewSecurityArgumentExtractor(g.gen)

	symbolResult, err := argExtractor.ExtractSymbol(call.Arguments[0])
	if err != nil {
		return "(func() float64 { return math.NaN() }())", nil
	}

	timeframeResult, err := argExtractor.ExtractTimeframe(call.Arguments[1])
	if err != nil {
		return "(func() float64 { return math.NaN() }())", nil
	}

	g.gen.hasSecurityCalls = true

	lookahead := resolveSecurityLookahead(call, g.gen.pineVersion)
	exprArg := call.Arguments[2]

	return g.generateIIFE(symbolResult, timeframeResult, exprArg, lookahead)
}

func (g *ArrowSecurityCallGenerator) generateIIFE(symbolResult, timeframeResult *ExtractionResult, exprArg ast.Expression, lookahead bool) (string, error) {
	keyBuilder := NewSecurityCacheKeyBuilder()
	keyComponents := keyBuilder.Build(symbolResult, timeframeResult)

	var b strings.Builder
	b.WriteString("(func() float64 {\n")

	if keyComponents.FormatArgs == "" {
		b.WriteString(fmt.Sprintf("\t\tsecKey := %q\n", keyComponents.KeyPattern))
	} else {
		b.WriteString(fmt.Sprintf("\t\tsecKey := fmt.Sprintf(%q, %s)\n", keyComponents.KeyPattern, keyComponents.FormatArgs))
	}

	b.WriteString("\t\tsecCtx, secFound := arrowCtx.SecurityContexts[secKey]\n")
	b.WriteString("\t\tif !secFound { return math.NaN() }\n\n")

	b.WriteString("\t\tsecBarMapper, mapperFound := arrowCtx.SecurityBarMappers[secKey]\n")
	b.WriteString("\t\tif !mapperFound { return math.NaN() }\n\n")

	b.WriteString(fmt.Sprintf("\t\tsecLookahead := %v\n", lookahead))
	b.WriteString(fmt.Sprintf("\t\tif %s == ctx.Timeframe { secLookahead = true }\n", timeframeResult.Code))
	b.WriteString("\t\tsecBarIdx := secBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)\n")
	b.WriteString("\t\tif secBarIdx < 0 { return math.NaN() }\n\n")

	evalCode, err := g.generateEvaluation(exprArg)
	if err != nil {
		return "", err
	}
	b.WriteString(evalCode)

	b.WriteString("\t}())")
	return b.String(), nil
}

func (g *ArrowSecurityCallGenerator) generateEvaluation(exprArg ast.Expression) (string, error) {
	if id, ok := exprArg.(*ast.Identifier); ok {
		if expr, ok := SecurityBarFieldExpression(id.Name, "secCtx.Data[secBarIdx]"); ok {
			return "\t\treturn " + expr + "\n", nil
		}
		return "\t\treturn math.NaN()\n", nil
	}
	return g.generateStreamingEvaluation(exprArg)
}

func (g *ArrowSecurityCallGenerator) generateStreamingEvaluation(exprArg ast.Expression) (string, error) {
	g.gen.hasSecurityExprEvals = true

	exprJSON, err := g.gen.serializeExpressionForRuntime(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	var b strings.Builder
	b.WriteString("\t\tsecEval := security.NewSeriesCachingEvaluator(security.NewStreamingBarEvaluator())\n")
	b.WriteString(fmt.Sprintf("\t\tsecValue, evalErr := secEval.EvaluateAtBar(%s, secCtx, secBarIdx)\n", exprJSON))
	b.WriteString("\t\tif evalErr != nil { return math.NaN() }\n")
	b.WriteString("\t\treturn secValue\n")
	return b.String(), nil
}
