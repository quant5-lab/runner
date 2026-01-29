package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

/* SecurityInjection holds prefetch code to inject before bar loop */
type SecurityInjection struct {
	PrefetchCode string   // Code to execute before bar loop
	ImportPaths  []string // Additional imports needed
}

/* AnalyzeAndGeneratePrefetch analyzes AST for security() calls and generates prefetch code */
func AnalyzeAndGeneratePrefetch(program *ast.Program) (*SecurityInjection, error) {
	calls := security.AnalyzeAST(program)

	if len(calls) == 0 {
		return &SecurityInjection{
			PrefetchCode: "",
			ImportPaths:  []string{},
		}, nil
	}

	limits := NewCodeGenerationLimits()
	validator := NewSecurityCallValidator(limits)
	if err := validator.ValidateCallCount(len(calls)); err != nil {
		return nil, err
	}

	var codeBuilder strings.Builder

	codeBuilder.WriteString("\n\t/* === request.security() Prefetch === */\n")
	codeBuilder.WriteString("\tfetcher := datafetcher.NewFileFetcher(dataDir, 0)\n\n")

	/* Generate prefetch request map (deduplicated symbol:timeframe pairs) */
	codeBuilder.WriteString("\t/* Fetch and cache multi-timeframe data */\n")

	/* Build deduplicated map of symbol:timeframe → expressions */
	dedupMap := make(map[string][]security.SecurityCall)
	for _, call := range calls {
		sym := call.Symbol
		if isRuntimeSymbol(sym) {
			sym = runtimePlaceholder()
		}

		tf := normalizeTimeframe(call.Timeframe)
		if isRuntimeTimeframe(tf) {
			tf = runtimePlaceholder()
		}

		key := fmt.Sprintf("%s:%s", sym, tf)
		dedupMap[key] = append(dedupMap[key], call)
	}

	codeBuilder.WriteString("\n\t/* Calculate base timeframe in seconds for warmup comparison */\n")
	codeBuilder.WriteString("\tbaseTimeframeSeconds := context.TimeframeToSeconds(ctx.Timeframe)\n")
	codeBuilder.WriteString("\tvar secTimeframeSeconds int64\n")
	codeBuilder.WriteString("\tbaseDateRange := request.NewDateRangeFromBars(ctx.Data, ctx.Timezone)\n")

	/* Generate fetch and store code for each unique symbol:timeframe */
	for key, callsForKey := range dedupMap {
		firstCall := callsForKey[0]

		parts := strings.Split(key, ":")
		tf := parts[len(parts)-1]
		sym := strings.Join(parts[:len(parts)-1], ":")

		isSymbolPlaceholder := sym == runtimePlaceholder()
		isTimeframePlaceholder := tf == runtimePlaceholder()

		symbolCode := "ctx.Symbol"
		if !isSymbolPlaceholder {
			symbolCode = fmt.Sprintf("%q", firstCall.Symbol)
		}

		timeframeCode := "ctx.Timeframe"
		timeframe := normalizeTimeframe(tf)
		if !isTimeframePlaceholder {
			timeframeCode = fmt.Sprintf("%q", timeframe)
		}

		varName := generateContextVarName(key, isSymbolPlaceholder, isTimeframePlaceholder)

		runtimeKey := key
		if isSymbolPlaceholder && isTimeframePlaceholder {
			runtimeKey = fmt.Sprintf("%%s:%%s")
		} else if isSymbolPlaceholder {
			runtimeKey = fmt.Sprintf("%%s:%s", tf)
		} else if isTimeframePlaceholder {
			runtimeKey = fmt.Sprintf("%s:%%s", sym)
		}

		if isTimeframePlaceholder {
			codeBuilder.WriteString("\tsecTimeframeSeconds = context.TimeframeToSeconds(ctx.Timeframe)\n")
		} else {
			codeBuilder.WriteString(fmt.Sprintf("\tsecTimeframeSeconds = context.TimeframeToSeconds(%q)\n", timeframe))
		}
		codeBuilder.WriteString("\tif secTimeframeSeconds == 0 {\n")
		codeBuilder.WriteString("\t\tsecTimeframeSeconds = baseTimeframeSeconds\n")
		codeBuilder.WriteString("\t}\n")
		/* Calculate dynamic warmup based on indicator periods in expressions */
		maxPeriod := 0
		for _, call := range callsForKey {
			period := security.ExtractMaxPeriod(call.Expression)
			if period > maxPeriod {
				maxPeriod = period
			}
		}
		/* Minimum warmup if no periods found or very small periods */
		warmupBars := maxPeriod
		if warmupBars < 50 {
			warmupBars = 50
		}

		codeBuilder.WriteString(fmt.Sprintf("\t%s_limit := len(ctx.Data)\n", varName))
		codeBuilder.WriteString("\tif secTimeframeSeconds != baseTimeframeSeconds && len(ctx.Data) > 0 {\n")
		codeBuilder.WriteString("\t\tfirstBarTime := ctx.Data[0].Time\n")
		codeBuilder.WriteString("\t\tlastBarTime := ctx.Data[len(ctx.Data)-1].Time\n")
		codeBuilder.WriteString("\t\ttimeSpanSeconds := lastBarTime - firstBarTime\n")
		codeBuilder.WriteString(fmt.Sprintf("\t\tbaseSecurityBars := int(timeSpanSeconds/secTimeframeSeconds) + 1\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\tdetectedWarmup := %d\n", warmupBars))
		codeBuilder.WriteString(fmt.Sprintf("\t\tfixedMinimumWarmup := 500\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\trequiredWarmup := fixedMinimumWarmup\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\tif detectedWarmup > requiredWarmup {\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\t\trequiredWarmup = detectedWarmup\n"))
		codeBuilder.WriteString("\t\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_limit = baseSecurityBars + requiredWarmup\n", varName))
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t%s_data, %s_err := fetcher.Fetch(%s, %s, %s_limit)\n",
			varName, varName, symbolCode, timeframeCode, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_err != nil {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\tfmt.Fprintf(os.Stderr, \"Failed to fetch %%s:%%s: %%%%v\\n\", %s, %s, %s_err)\n", symbolCode, timeframeCode, varName))
		codeBuilder.WriteString("\t\tos.Exit(1)\n")
		codeBuilder.WriteString("\t}\n")

		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx := context.New(%s, %s, len(%s_data))\n",
			varName, symbolCode, timeframeCode, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tfor _, bar := range %s_data {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_ctx.AddBar(bar)\n", varName))
		codeBuilder.WriteString("\t}\n")

		if isSymbolPlaceholder || isTimeframePlaceholder {
			var runtimeKeyArgs []string
			if isSymbolPlaceholder {
				runtimeKeyArgs = append(runtimeKeyArgs, "ctx.Symbol")
			}
			if isTimeframePlaceholder {
				runtimeKeyArgs = append(runtimeKeyArgs, "ctx.Timeframe")
			}
			keyExpr := fmt.Sprintf("fmt.Sprintf(%q, %s)", runtimeKey, strings.Join(runtimeKeyArgs, ", "))

			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%s] = %s_ctx\n", keyExpr, varName))
			codeBuilder.WriteString(fmt.Sprintf("\t%s_mapper := request.NewSecurityBarMapper()\n", varName))
			codeBuilder.WriteString("\tif secTimeframeSeconds < baseTimeframeSeconds {\n")
			codeBuilder.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingForUpscaling(%s_ctx.Data, ctx.Data, ctx.Timezone)\n", varName, varName))
			codeBuilder.WriteString("\t} else {\n")
			codeBuilder.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingWithDateFilter(%s_ctx.Data, ctx.Data, baseDateRange, ctx.Timezone)\n", varName, varName))
			codeBuilder.WriteString("\t}\n")
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityBarMappers[%s] = %s_mapper\n\n", keyExpr, varName))
		} else {
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%q] = %s_ctx\n", key, varName))
			codeBuilder.WriteString(fmt.Sprintf("\t%s_mapper := request.NewSecurityBarMapper()\n", varName))
			codeBuilder.WriteString("\tif secTimeframeSeconds < baseTimeframeSeconds {\n")
			codeBuilder.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingForUpscaling(%s_ctx.Data, ctx.Data, ctx.Timezone)\n", varName, varName))
			codeBuilder.WriteString("\t} else {\n")
			codeBuilder.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingWithDateFilter(%s_ctx.Data, ctx.Data, baseDateRange, ctx.Timezone)\n", varName, varName))
			codeBuilder.WriteString("\t}\n")
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityBarMappers[%q] = %s_mapper\n\n", key, varName))
		}
	}

	codeBuilder.WriteString("\t_ = fetcher\n")
	codeBuilder.WriteString("\t/* === End Prefetch === */\n\n")

	/* Required imports */
	imports := []string{
		"github.com/quant5-lab/runner/datafetcher",
		"github.com/quant5-lab/runner/security",
		"github.com/quant5-lab/runner/ast",
	}

	return &SecurityInjection{
		PrefetchCode: codeBuilder.String(),
		ImportPaths:  imports,
	}, nil
}

/* GenerateSecurityLookup generates runtime cache lookup code for security() calls */
func GenerateSecurityLookup(call *security.SecurityCall, varName string) string {
	/* Generate cache lookup:
	 * entry, found := securityCache.Get(symbol, timeframe)
	 * if !found { return NaN }
	 * values, err := securityCache.GetExpression(symbol, timeframe, exprName)
	 * if err != nil { return NaN }
	 * value := values[ctx.BarIndex] // Index matching logic
	 */

	var code strings.Builder

	code.WriteString(fmt.Sprintf("\t/* security(%q, %q, ...) lookup */\n", call.Symbol, call.Timeframe))
	code.WriteString(fmt.Sprintf("\t%s_values, err := securityCache.GetExpression(%q, %q, %q)\n",
		varName, call.Symbol, call.Timeframe, call.ExprName))
	code.WriteString(fmt.Sprintf("\tif err != nil {\n"))
	code.WriteString(fmt.Sprintf("\t\t%s = math.NaN()\n", varName))
	code.WriteString(fmt.Sprintf("\t} else {\n"))
	code.WriteString(fmt.Sprintf("\t\tif ctx.BarIndex < len(%s_values) {\n", varName))
	code.WriteString(fmt.Sprintf("\t\t\t%s = %s_values[ctx.BarIndex]\n", varName, varName))
	code.WriteString(fmt.Sprintf("\t\t} else {\n"))
	code.WriteString(fmt.Sprintf("\t\t\t%s = math.NaN()\n", varName))
	code.WriteString(fmt.Sprintf("\t\t}\n"))
	code.WriteString(fmt.Sprintf("\t}\n"))

	return code.String()
}

/* InjectSecurityCode updates StrategyCode with security prefetch and lookups */
func InjectSecurityCode(code *StrategyCode, program *ast.Program) (*StrategyCode, error) {
	/* Analyze and generate prefetch code */
	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze security calls: %w", err)
	}

	if injection.PrefetchCode == "" {
		/* No security() calls - return unchanged */
		return code, nil
	}

	/* Inject prefetch code before strategy execution */
	/* Expected structure:
	 * func executeStrategy(ctx *context.Context) (*output.Collector, *strategy.Strategy) {
	 *     collector := output.NewCollector()
	 *     strat := strategy.NewStrategy()
	 *
	 *     <<< INJECT PREFETCH HERE >>>
	 *
	 *     for i := 0; i < len(ctx.Data); i++ {
	 *         ...
	 *     }
	 * }
	 */

	/* Find insertion point: after strat initialization, before for loop */
	functionBody := code.FunctionBody

	/* Simple injection: prepend before existing body */
	updatedBody := injection.PrefetchCode + functionBody

	mergedImports := mergeImports(code.AdditionalImports, injection.ImportPaths)

	return &StrategyCode{
		UserDefinedFunctions: code.UserDefinedFunctions,
		FunctionBody:         updatedBody,
		StrategyName:         code.StrategyName,
		AdditionalImports:    mergedImports,
	}, nil
}

/* mergeImports combines two import lists without duplicates */
func mergeImports(existing, additional []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(additional))

	for _, imp := range existing {
		if !seen[imp] {
			seen[imp] = true
			result = append(result, imp)
		}
	}

	for _, imp := range additional {
		if !seen[imp] {
			seen[imp] = true
			result = append(result, imp)
		}
	}

	return result
}

/* normalizeTimeframe converts short forms to canonical format */
func normalizeTimeframe(tf string) string {
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

/* generateContextVarName creates unique variable name for each symbol:timeframe */
func generateContextVarName(key string, isSymbolPlaceholder, isTimeframePlaceholder bool) string {
	parts := strings.Split(key, ":")
	if len(parts) < 2 {
		return sanitizeVarName(key)
	}
	sym := strings.Join(parts[:len(parts)-1], ":")
	tf := parts[len(parts)-1]

	if isSymbolPlaceholder && isTimeframePlaceholder {
		return "sec_runtime"
	} else if isSymbolPlaceholder {
		return sanitizeVarName(fmt.Sprintf("sec_runtime_%s", tf))
	} else if isTimeframePlaceholder {
		return sanitizeVarName(fmt.Sprintf("sec_%s_runtime", sym))
	}
	return sanitizeVarName(key)
}

/* sanitizeVarName converts "SYMBOL:TIMEFRAME" to valid Go variable name */
func sanitizeVarName(s string) string {
	// Replace colons and special chars with underscores
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return strings.ToLower(s)
}
