package codegen

import (
	"fmt"
	"strings"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/security"
)

/* SecurityInjection holds prefetch code to inject before bar loop */
type SecurityInjection struct {
	PrefetchCode string   // Code to execute before bar loop
	ImportPaths  []string // Additional imports needed
}

/* AnalyzeAndGeneratePrefetch analyzes AST for security() calls and generates prefetch code */
func AnalyzeAndGeneratePrefetch(program *ast.Program) (*SecurityInjection, error) {
	/* Analyze AST for security() calls */
	calls := security.AnalyzeAST(program)
	
	if len(calls) == 0 {
		/* No security() calls - return empty injection */
		return &SecurityInjection{
			PrefetchCode: "",
			ImportPaths:  []string{},
		}, nil
	}

	/* Generate prefetch initialization code */
	var codeBuilder strings.Builder
	
	codeBuilder.WriteString("\n\t/* === request.security() Prefetch === */\n")
	codeBuilder.WriteString("\tfetcher := datafetcher.NewFileFetcher(dataDir, 0)\n\n")
	
	/* Generate prefetch request map (deduplicated symbol:timeframe pairs) */
	codeBuilder.WriteString("\t/* Fetch and cache multi-timeframe data */\n")
	
	/* Build deduplicated map of symbol:timeframe → expressions */
	dedupMap := make(map[string][]security.SecurityCall)
	for _, call := range calls {
		/* Normalize symbol for grouping: empty/"tickerid"/"syminfo.tickerid" → "%s" placeholder */
		sym := call.Symbol
		if sym == "" || sym == "tickerid" || sym == "syminfo.tickerid" {
			sym = "%s" // Runtime placeholder
		}
		
		/* Normalize timeframe */
		tf := call.Timeframe
		if tf == "D" {
			tf = "1D"
		} else if tf == "W" {
			tf = "1W"
		} else if tf == "M" {
			tf = "1M"
		}
		
		key := fmt.Sprintf("%s:%s", sym, tf)
		dedupMap[key] = append(dedupMap[key], call)
	}
	
	/* Don't create new map - use parameter passed to function */
	
	codeBuilder.WriteString("\n\t/* Calculate base timeframe in seconds for warmup comparison */\n")
	codeBuilder.WriteString("\tbaseTimeframeSeconds := context.TimeframeToSeconds(ctx.Timeframe)\n")
	codeBuilder.WriteString("\tvar secTimeframeSeconds int64 /* Reused for multiple security() calls */\n")
	
	/* Generate fetch and store code for each unique symbol:timeframe */
	for key, callsForKey := range dedupMap {
		firstCall := callsForKey[0]
		
		/* Extract normalized symbol and timeframe from key */
		sym := "%s"
		tf := ""
		parts := strings.Split(key, ":")
		if len(parts) == 2 {
			sym = parts[0]
			tf = parts[1]
		}
		
		/* Resolve symbol: use ctx.Symbol for runtime placeholders */
		symbolCode := firstCall.Symbol
		if symbolCode == "" || symbolCode == "tickerid" || symbolCode == "syminfo.tickerid" || sym == "%s" {
			symbolCode = "ctx.Symbol"
		} else {
			symbolCode = fmt.Sprintf("%q", symbolCode)
		}
		
		/* Normalize timeframe for fetcher */
		timeframe := tf
		if timeframe == "" {
			timeframe = firstCall.Timeframe
		}
		if timeframe == "D" {
			timeframe = "1D"
		} else if timeframe == "W" {
			timeframe = "1W"
		} else if timeframe == "M" {
			timeframe = "1M"
		}
		
		varName := sanitizeVarName(fmt.Sprintf("ctx_%s", tf))
		/* Generate runtime key - if symbol is placeholder, use fmt.Sprintf at runtime */
		runtimeKey := key
		if sym == "%s" {
			runtimeKey = fmt.Sprintf("%%s:%s", tf) // Will be formatted at runtime
		}
		
		codeBuilder.WriteString(fmt.Sprintf("\t/* Fetch %s data */\n", key))
		codeBuilder.WriteString(fmt.Sprintf("\tsecTimeframeSeconds = context.TimeframeToSeconds(%q)\n", timeframe))
		codeBuilder.WriteString("\t/* Empty timeframe means use base timeframe (same timeframe) */\n")
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
		/* Default minimum warmup if no periods found or very small periods */
		warmupBars := maxPeriod
		if warmupBars < 50 {
			warmupBars = 50 /* Minimum warmup for basic indicators */
		}
		
		codeBuilder.WriteString(fmt.Sprintf("\t/* Dynamic warmup based on indicators: %d bars */\n", warmupBars))
		codeBuilder.WriteString(fmt.Sprintf("\t%s_limit := len(ctx.Data)\n", varName))
		codeBuilder.WriteString("\tif secTimeframeSeconds > baseTimeframeSeconds {\n")
		codeBuilder.WriteString(fmt.Sprintf("\t\t/* Convert base timeframe bars to security timeframe bars + warmup */\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\ttimeframeRatio := float64(secTimeframeSeconds) / float64(baseTimeframeSeconds)\n"))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_limit = int(float64(len(ctx.Data)) * timeframeRatio) + %d\n", varName, warmupBars))
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t%s_data, %s_err := fetcher.Fetch(%s, %q, %s_limit)\n",
			varName, varName, symbolCode, timeframe, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_err != nil {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\tfmt.Fprintf(os.Stderr, \"Failed to fetch %s: %%v\\n\", %s_err)\n", key, varName))
		codeBuilder.WriteString("\t\tos.Exit(1)\n")
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx := context.New(%s, %q, len(%s_data))\n",
			varName, symbolCode, timeframe, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tfor _, bar := range %s_data {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_ctx.AddBar(bar)\n", varName))
		codeBuilder.WriteString("\t}\n")
		
		/* Store in map with runtime key resolution */
		if sym == "%s" {
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[fmt.Sprintf(%q, ctx.Symbol)] = %s_ctx\n\n", runtimeKey, varName))
		} else {
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%q] = %s_ctx\n\n", key, varName))
		}
	}
	
	codeBuilder.WriteString("\t_ = fetcher // Suppress unused warning\n")
	codeBuilder.WriteString("\t/* === End Prefetch === */\n\n")

	/* Required imports */
	imports := []string{
		"github.com/borisquantlab/pinescript-go/datafetcher",
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

	return &StrategyCode{
		FunctionBody:       updatedBody,
		StrategyName:       code.StrategyName,
		NeedsSeriesPreCalc: true, // Security requires imports
	}, nil
}

/* sanitizeVarName converts "SYMBOL:TIMEFRAME" to valid Go variable name */
func sanitizeVarName(s string) string {
	// Replace colons and special chars with underscores
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return strings.ToLower(s)
}
