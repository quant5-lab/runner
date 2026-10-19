package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// ArrowTupleSecurityGenerator emits tuple security() destructuring inside arrow
// function bodies. All context lookups go through arrowCtx rather than the
// top-level securityContexts/securityBarMappers maps used in the main bar loop.
type ArrowTupleSecurityGenerator struct {
	gen          *generator
	localStorage *ArrowLocalVariableStorage
	arrowScope   map[string]string
}

func NewArrowTupleSecurityGenerator(
	gen *generator,
	localStorage *ArrowLocalVariableStorage,
	arrowScope map[string]string,
) *ArrowTupleSecurityGenerator {
	return &ArrowTupleSecurityGenerator{
		gen:          gen,
		localStorage: localStorage,
		arrowScope:   arrowScope,
	}
}

// Generate produces the tuple destructuring block for security() inside an arrow body,
// then extracts each result as a Go local so downstream arrow statements can reference
// the names without an explicit series lookup.
func (a *ArrowTupleSecurityGenerator) Generate(varNames []string, call *ast.CallExpression) (string, error) {
	args, err := parseTupleSecurityArgumentsWithScope(a.gen, varNames, call, a.arrowScope)
	if err != nil {
		return "", err
	}
	a.gen.hasSecurityCalls = true
	code, err := a.emitBlock(args)
	if err != nil {
		return "", err
	}
	for _, name := range varNames {
		code += a.gen.ind() + fmt.Sprintf("%s := %sSeries.GetCurrent()\n", name, name)
	}
	return code, nil
}

func (a *ArrowTupleSecurityGenerator) emitBlock(args *tupleSecurityArguments) (string, error) {
	g := a.gen

	code := g.ind() + fmt.Sprintf("/* security(%s, %s, [%s]) */\n",
		args.symbolResult.Code, args.timeframeResult.Code, strings.Join(args.varNames, ", "))
	code += g.ind() + "{\n"
	g.indent++

	code += a.emitCacheKeyLookup(args.cacheKey)
	code += a.emitContextGuard(args.varNames)
	code += a.emitBarMapperGuard(args.varNames)
	code += a.emitLookaheadAndParentLinkage(args)

	code += g.ind() + "secBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)\n"
	code += g.ind() + "if secBarIdx < 0 {\n"
	g.indent++
	code += a.emitNaNFallback(args.varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++

	evalCode, err := a.emitElementEvaluations(args)
	if err != nil {
		return "", err
	}
	code += evalCode

	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (a *ArrowTupleSecurityGenerator) emitCacheKeyLookup(key CacheKeyComponents) string {
	if key.FormatArgs == "" {
		return a.gen.ind() + fmt.Sprintf("secKey := %q\n", key.KeyPattern)
	}
	return a.gen.ind() + fmt.Sprintf("secKey := fmt.Sprintf(%q, %s)\n", key.KeyPattern, key.FormatArgs)
}

func (a *ArrowTupleSecurityGenerator) emitContextGuard(varNames []string) string {
	g := a.gen
	code := g.ind() + "secCtx, secFound := arrowCtx.SecurityContexts[secKey]\n"
	code += g.ind() + "if !secFound {\n"
	g.indent++
	code += a.emitNaNFallback(varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	return code
}

func (a *ArrowTupleSecurityGenerator) emitBarMapperGuard(varNames []string) string {
	g := a.gen
	code := g.ind() + "_, mapperFound := arrowCtx.SecurityBarMappers[secKey]\n"
	code += g.ind() + "if !mapperFound {\n"
	g.indent++
	code += a.emitNaNFallback(varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + "securityBarMapper := arrowCtx.ConcreteBarMappers[secKey].(*request.SecurityBarMapper)\n"
	return code
}

func (a *ArrowTupleSecurityGenerator) emitLookaheadAndParentLinkage(args *tupleSecurityArguments) string {
	g := a.gen
	code := g.ind() + fmt.Sprintf("secLookahead := %v\n", args.lookahead)
	code += g.ind() + fmt.Sprintf("if %s == ctx.Timeframe {\n", args.timeframeResult.Code)
	g.indent++
	code += g.ind() + "secLookahead = true\n"
	g.indent--
	code += g.ind() + "}\n\n"

	code += g.ind() + "if secCtx.GetParent() == nil {\n"
	g.indent++
	code += g.ind() + "barAligner := request.NewSecurityBarMapperAligner(securityBarMapper, secLookahead)\n"
	code += g.ind() + "secCtx.SetParent(ctx, barAligner)\n"
	g.indent--
	code += g.ind() + "}\n\n"

	return code
}

func (a *ArrowTupleSecurityGenerator) emitNaNFallback(varNames []string) string {
	code := ""
	for _, name := range varNames {
		code += a.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", name)
	}
	return code
}

func (a *ArrowTupleSecurityGenerator) emitElementEvaluations(args *tupleSecurityArguments) (string, error) {
	if args.udfCall != nil {
		return NewSecurityUDFCallGenerator(a.gen).EmitTupleCall(args.varNames, args.udfCall)
	}

	code := ""
	for i, varName := range args.varNames {
		evalCode, err := a.emitArrowElementEvaluation(varName, args.elements[i])
		if err != nil {
			return "", fmt.Errorf("element %d (%s): %w", i, varName, err)
		}
		code += evalCode
	}
	return code, nil
}

// emitArrowElementEvaluation generates per-element evaluation code valid inside an
// arrow function body. OHLCV identifiers resolve to direct data field access; all
// other expressions go through the arrow context security evaluator map.
func (a *ArrowTupleSecurityGenerator) emitArrowElementEvaluation(varName string, exprArg ast.Expression) (string, error) {
	g := a.gen
	if ident, ok := exprArg.(*ast.Identifier); ok {
		if fieldExpr, ok := SecurityBarFieldExpression(ident.Name, "secCtx.Data[secBarIdx]"); ok {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, fieldExpr), nil
		}
	}
	return a.emitArrowStreamingEvaluation(varName, exprArg)
}

// emitArrowStreamingEvaluation uses the arrowCtx security evaluator map rather than the
// top-level secBarEvaluator that only exists in the main bar loop.
func (a *ArrowTupleSecurityGenerator) emitArrowStreamingEvaluation(varName string, exprArg ast.Expression) (string, error) {
	g := a.gen
	g.hasSecurityExprEvals = true

	exprJSON, err := g.serializeExpressionForRuntime(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression for %s: %w", varName, err)
	}

	evalMapVar := "arrowSecEvalMap_" + varName

	var b strings.Builder
	b.WriteString(g.ind() + "{\n")
	g.indent++

	b.WriteString(g.ind() + evalMapVar + " := arrowCtx.GetOrCreateSecurityEvaluators()\n")
	b.WriteString(g.ind() + "if " + evalMapVar + "[secKey] == nil {\n")
	g.indent++

	initializer := NewArrowSecurityEvaluatorInitializer(g.symbolTable, g)
	b.WriteString(initializer.EmitInitializationBody(g.ind, func() { g.indent++ }, func() { g.indent-- }))
	b.WriteString(g.ind() + evalMapVar + "[secKey] = security.NewSeriesCachingEvaluator(baseEvaluator)\n")

	g.indent--
	b.WriteString(g.ind() + "}\n")

	b.WriteString(g.ind() + fmt.Sprintf(
		"%sVal, %sErr := %s[secKey].(security.BarEvaluator).EvaluateAtBar(%s, secCtx, secBarIdx)\n",
		varName, varName, evalMapVar, exprJSON))
	b.WriteString(g.ind() + fmt.Sprintf(
		"if %sErr != nil { %sSeries.Set(math.NaN()) } else { %sSeries.Set(%sVal) }\n",
		varName, varName, varName, varName))

	g.indent--
	b.WriteString(g.ind() + "}\n")

	return b.String(), nil
}
