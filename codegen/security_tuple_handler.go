package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type tupleSecurityArguments struct {
	varNames        []string
	elements        []ast.Expression
	udfCall         *ast.CallExpression
	symbolResult    *ExtractionResult
	timeframeResult *ExtractionResult
	cacheKey        CacheKeyComponents
	lookahead       bool
}

func (g *generator) generateTupleSecurityDeclaration(varNames []string, call *ast.CallExpression) (string, error) {
	args, err := parseTupleSecurityArguments(g, varNames, call)
	if err != nil {
		return "", err
	}
	g.hasSecurityCalls = true
	return g.emitTupleSecurityBlock(args)
}

func parseTupleSecurityArguments(g *generator, varNames []string, call *ast.CallExpression) (*tupleSecurityArguments, error) {
	return assembleTupleSecurityArguments(g, varNames, call, NewSecurityArgumentExtractor(g))
}

func parseTupleSecurityArgumentsWithScope(g *generator, varNames []string, call *ast.CallExpression, arrowScope map[string]string) (*tupleSecurityArguments, error) {
	return assembleTupleSecurityArguments(g, varNames, call, NewSecurityArgumentExtractor(g).WithArrowScope(arrowScope))
}

func assembleTupleSecurityArguments(g *generator, varNames []string, call *ast.CallExpression, extractor *SecurityArgumentExtractor) (*tupleSecurityArguments, error) {
	if len(call.Arguments) < 3 {
		return nil, fmt.Errorf("security() requires at least 3 arguments, got %d", len(call.Arguments))
	}

	elements, udfCall, err := resolveTupleExpressionArg(
		NewUserDefinedFunctionDetector(g.variables),
		call.Arguments[2],
		len(varNames),
	)
	if err != nil {
		return nil, err
	}

	symbolResult, err := extractor.ExtractSymbol(call.Arguments[0])
	if err != nil {
		return nil, fmt.Errorf("security symbol: %w", err)
	}

	timeframeResult, err := extractor.ExtractTimeframe(call.Arguments[1])
	if err != nil {
		return nil, fmt.Errorf("security timeframe: %w", err)
	}

	return &tupleSecurityArguments{
		varNames:        varNames,
		elements:        elements,
		udfCall:         udfCall,
		symbolResult:    symbolResult,
		timeframeResult: timeframeResult,
		cacheKey:        NewSecurityCacheKeyBuilder().Build(symbolResult, timeframeResult),
		lookahead:       resolveSecurityLookahead(call, g.pineVersion),
	}, nil
}

// Non-UDF CallExpression arguments (e.g. bare TA calls) are rejected; only array
// literals and registered UDF calls are valid in the third security() position.
func resolveTupleExpressionArg(
	detector *UserDefinedFunctionDetector,
	arg ast.Expression,
	expectedCount int,
) (elements []ast.Expression, udfCall *ast.CallExpression, err error) {
	if callArg, ok := arg.(*ast.CallExpression); ok &&
		detector.IsUserDefinedFunction(extractCallFunctionName(callArg)) {
		return nil, callArg, nil
	}
	elems, extractErr := extractTupleExpressionElements(arg)
	if extractErr != nil {
		return nil, nil, fmt.Errorf("security() tuple: %w", extractErr)
	}
	if len(elems) != expectedCount {
		return nil, nil, fmt.Errorf("security() tuple: cardinality mismatch: %d variables vs %d expressions", expectedCount, len(elems))
	}
	return elems, nil, nil
}

func (g *generator) emitTupleSecurityBlock(args *tupleSecurityArguments) (string, error) {
	code := g.ind() + fmt.Sprintf("/* security(%s, %s, [%s]) */\n",
		args.symbolResult.Code, args.timeframeResult.Code, strings.Join(args.varNames, ", "))
	code += g.ind() + "{\n"
	g.indent++

	code += g.emitSecurityCacheKeyLookup(args.cacheKey)
	code += g.emitSecurityContextGuard(args.varNames)
	code += g.emitBarMapperGuard(args.varNames)
	code += g.emitLookaheadAndParentLinkage(args)

	code += g.ind() + "secBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)\n"
	code += g.ind() + "if secBarIdx < 0 {\n"
	g.indent++
	code += g.emitNaNFallbackForAllVars(args.varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++

	evalCode, err := g.emitScopedElementEvaluations(args)
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

func (g *generator) emitSecurityCacheKeyLookup(key CacheKeyComponents) string {
	if key.FormatArgs == "" {
		return g.ind() + fmt.Sprintf("secKey := %q\n", key.KeyPattern)
	}
	return g.ind() + fmt.Sprintf("secKey := fmt.Sprintf(%q, %s)\n", key.KeyPattern, key.FormatArgs)
}

func (g *generator) emitSecurityContextGuard(varNames []string) string {
	code := g.ind() + "secCtx, secFound := securityContexts[secKey]\n"
	code += g.ind() + "if !secFound {\n"
	g.indent++
	code += g.emitNaNFallbackForAllVars(varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	return code
}

func (g *generator) emitBarMapperGuard(varNames []string) string {
	code := g.ind() + "securityBarMapper, mapperFound := securityBarMappers[secKey]\n"
	code += g.ind() + "if !mapperFound {\n"
	g.indent++
	code += g.emitNaNFallbackForAllVars(varNames)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	return code
}

func (g *generator) emitLookaheadAndParentLinkage(args *tupleSecurityArguments) string {
	code := g.ind() + fmt.Sprintf("secLookahead := %v\n", args.lookahead)
	code += g.ind() + fmt.Sprintf("if %s == ctx.Timeframe {\n", args.timeframeResult.Code)
	g.indent++
	code += g.ind() + "secLookahead = true\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + "\n"

	code += g.ind() + "if secCtx.GetParent() == nil {\n"
	g.indent++
	code += g.ind() + "barAligner := request.NewSecurityBarMapperAligner(securityBarMapper, secLookahead)\n"
	code += g.ind() + "secCtx.SetParent(ctx, barAligner)\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + "\n"

	return code
}

// Mirrors ArrowTupleSecurityGenerator.emitElementEvaluations — both paths must stay in sync.
func (g *generator) emitScopedElementEvaluations(args *tupleSecurityArguments) (string, error) {
	if args.udfCall != nil {
		return NewSecurityUDFCallGenerator(g).EmitTupleCall(args.varNames, args.udfCall)
	}

	handler := NewSecurityExpressionHandler(SecurityExpressionConfig{
		IndentFunc:           g.ind,
		IncrementIndent:      func() { g.indent++ },
		DecrementIndent:      func() { g.indent-- },
		SerializeExpr:        g.serializeExpressionForRuntime,
		MarkSecurityExprEval: func() { g.hasSecurityExprEvals = true },
		SymbolTable:          g.symbolTable,
		Generator:            g,
	})

	code := ""
	for i, varName := range args.varNames {
		code += g.ind() + "{\n"
		g.indent++

		evalCode, err := handler.GenerateEvaluationCode(varName, args.elements[i], "secBarIdx")
		if err != nil {
			return "", fmt.Errorf("element %d (%s): %w", i, varName, err)
		}
		code += evalCode

		g.indent--
		code += g.ind() + "}\n"
	}
	return code, nil
}

func (g *generator) emitNaNFallbackForAllVars(varNames []string) string {
	code := ""
	for _, varName := range varNames {
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	}
	return code
}

func extractTupleExpressionElements(expr ast.Expression) ([]ast.Expression, error) {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return nil, fmt.Errorf("expected array literal, got %T", expr)
	}

	elements, ok := lit.Value.([]ast.Expression)
	if !ok {
		return nil, fmt.Errorf("expected []ast.Expression in literal, got %T", lit.Value)
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("empty expression array")
	}

	return elements, nil
}
