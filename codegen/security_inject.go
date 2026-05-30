package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

/* SecurityInjection holds prefetch code to inject before bar loop */
type SecurityInjection struct {
	PrefetchCode string // Code to execute before bar loop
	ImportPaths  []string
}

type resolvedSecurityCall struct {
	call           security.SecurityCall
	resolvedSym    string
	resolvedTf     string
	isSymRuntime   bool
	isTfRuntime    bool
	modifierPrefix string // e.g., "HEIKINASHI" if symbol was "HEIKINASHI:syminfo.tickerid"
}

func buildVariableMap(program *ast.Program) map[string]string {
	vars := make(map[string]string)
	if program == nil {
		return vars
	}

	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, decl := range varDecl.Declarations {
				if decl.Init == nil {
					continue
				}
				id, ok := decl.ID.(*ast.Identifier)
				if !ok {
					continue
				}
				if lit, ok := decl.Init.(*ast.Literal); ok {
					if s, ok := lit.Value.(string); ok {
						vars[id.Name] = strings.Trim(s, "\"'")
					}
					continue
				}
				if call, ok := decl.Init.(*ast.CallExpression); ok {
					if isInputCallExpr(call) {
						if defval := extractInputDefvalLiteral(call); defval != "" {
							vars[id.Name] = strings.Trim(defval, "\"'")
						}
					}
				}
			}
		}
	}
	return vars
}

func resolveSecurityArgument(expr ast.Expression, rawValue string, vars map[string]string) (resolved string, isRuntime bool) {
	if isRuntimeSymbol(rawValue) || isRuntimeTimeframe(rawValue) {
		return rawValue, true
	}

	if lit, ok := expr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			return strings.Trim(s, "\"'"), false
		}
	}

	if id, ok := expr.(*ast.Identifier); ok {
		if val, found := vars[id.Name]; found {
			return val, false
		}
		return rawValue, true
	}

	if _, ok := expr.(*ast.MemberExpression); ok {
		return rawValue, true
	}

	return rawValue, false
}

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

	/* Build variable map for user variable resolution */
	vars := buildVariableMap(program)

	/* Resolve symbol/timeframe for each call */
	resolved := make([]resolvedSecurityCall, len(calls))
	for i, call := range calls {
		sym, symRuntime := resolveSecurityArgument(call.SymbolExpr, call.Symbol, vars)
		tf, tfRuntime := resolveSecurityArgument(call.TimeframeExpr, call.Timeframe, vars)

		/* Extract modifier prefix from symbol (e.g., HEIKINASHI:syminfo.tickerid → prefix=HEIKINASHI, sym=syminfo.tickerid) */
		modPrefix, baseSym, hasModifier := extractModifierPrefix(sym)
		if hasModifier {
			sym = baseSym
		}

		resolved[i] = resolvedSecurityCall{
			call:           call,
			resolvedSym:    sym,
			resolvedTf:     normalizeTimeframe(tf),
			isSymRuntime:   symRuntime,
			isTfRuntime:    tfRuntime,
			modifierPrefix: modPrefix,
		}
	}

	var codeBuilder strings.Builder

	codeBuilder.WriteString("\tfetcher := datafetcher.NewFileFetcher(dataDir, 0)\n\n")

	dedupMap := make(map[string][]resolvedSecurityCall)
	for _, r := range resolved {
		sym := r.resolvedSym
		if r.isSymRuntime {
			sym = runtimePlaceholder()
		}

		tf := r.resolvedTf
		if r.isTfRuntime {
			tf = runtimePlaceholder()
		}

		key := fmt.Sprintf("%s:%s", sym, tf)
		dedupMap[key] = append(dedupMap[key], r)
	}

	codeBuilder.WriteString("\tbaseTimeframeSeconds := context.TimeframeToSeconds(ctx.Timeframe)\n")
	codeBuilder.WriteString("\tvar secTimeframeSeconds int64\n")
	codeBuilder.WriteString("\tbaseDateRange := request.NewDateRangeFromBars(ctx.Data, ctx.Timezone)\n")

	for key, callsForKey := range dedupMap {
		firstCall := callsForKey[0]

		parts := strings.Split(key, ":")
		tf := parts[len(parts)-1]
		sym := strings.Join(parts[:len(parts)-1], ":")

		isSymbolPlaceholder := sym == runtimePlaceholder()
		isTimeframePlaceholder := tf == runtimePlaceholder()

		symbolCode := "ctx.Symbol"
		if !isSymbolPlaceholder {
			symbolCode = fmt.Sprintf("%q", firstCall.resolvedSym)
		}

		timeframeCode := "ctx.Timeframe"
		timeframe := firstCall.resolvedTf
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

		maxPeriod := 0
		for _, r := range callsForKey {
			period := security.ExtractMaxPeriod(r.call.Expression)
			if period > maxPeriod {
				maxPeriod = period
			}
		}

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
		codeBuilder.WriteString(fmt.Sprintf("\t%s_marketData, %s_err := fetcher.FetchWithMetadata(%s, %s, 0)\n",
			varName, varName, symbolCode, timeframeCode))
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_err != nil {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\tfmt.Fprintf(os.Stderr, \"Failed to fetch %%s:%%s: %%v\\n\", %s, %s, %s_err)\n", symbolCode, timeframeCode, varName))
		codeBuilder.WriteString("\t\tos.Exit(1)\n")
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t%s_metadata := %s_marketData.SourceMetadata\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_metadata.ReferenceSession == \"\" {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_metadata.ReferenceSession = ctx.ReferenceSession\n", varName))
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_metadata.Timezone == \"\" {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_metadata.Timezone = ctx.Timezone\n", varName))
		codeBuilder.WriteString("\t}\n")
		codeBuilder.WriteString(fmt.Sprintf("\t%s_data, %s_profile := market.NormalizeBarsWithMetadata(%s, %s, %s_metadata, %s_marketData.Bars)\n",
			varName, varName, symbolCode, timeframeCode, varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tif %s_limit > 0 && %s_limit < len(%s_data) {\n",
			varName, varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_data = %s_data[len(%s_data)-%s_limit:]\n",
			varName, varName, varName, varName))
		codeBuilder.WriteString("\t}\n")

		hasModifier := firstCall.modifierPrefix != ""

		// Transform data if modifier present; capture TransformResult for bar-mapping.
		if hasModifier {
			ctorCode := transformerConstructorCode(firstCall.call.SymbolExpr, firstCall.modifierPrefix)
			codeBuilder.WriteString(fmt.Sprintf("\t%s_transformResult := (%s).Transform(%s_data)\n",
				varName, ctorCode, varName))
			codeBuilder.WriteString(fmt.Sprintf("\t%s_data = %s_transformResult.Bars\n", varName, varName))
		}

		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx := context.New(%s, %s, len(%s_data))\n",
			varName, symbolCode, timeframeCode, varName))
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx.Timezone = %s_profile.Timezone\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx.ReferenceSession = string(%s_profile.ReferenceSession)\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tfor _, bar := range %s_data {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_ctx.AddBar(bar)\n", varName))
		codeBuilder.WriteString("\t}\n")

		resolvedKey := fmt.Sprintf("%s:%s", firstCall.resolvedSym, firstCall.resolvedTf)

		mapperCode := buildMapperCode(varName, hasModifier && isVariableBarCountModifier(firstCall.modifierPrefix))

		if isSymbolPlaceholder || isTimeframePlaceholder {
			var runtimeKeyArgs []string
			symbolArg := "ctx.Symbol"
			if hasModifier {
				symbolArg = generateModifierCall(firstCall.modifierPrefix, "ctx.Symbol", firstCall.call.SymbolExpr)
			}
			if isSymbolPlaceholder {
				runtimeKeyArgs = append(runtimeKeyArgs, symbolArg)
			}
			if isTimeframePlaceholder {
				runtimeKeyArgs = append(runtimeKeyArgs, "ctx.Timeframe")
			}
			keyExpr := fmt.Sprintf("fmt.Sprintf(%q, %s)", runtimeKey, strings.Join(runtimeKeyArgs, ", "))

			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%s] = %s_ctx\n", keyExpr, varName))
			codeBuilder.WriteString(mapperCode)
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityBarMappers[%s] = %s_mapper\n\n", keyExpr, varName))
		} else {
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%q] = %s_ctx\n", resolvedKey, varName))
			codeBuilder.WriteString(mapperCode)
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityBarMappers[%q] = %s_mapper\n\n", resolvedKey, varName))
		}
	}

	codeBuilder.WriteString("\t_ = fetcher\n\n")

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

// GenerateSecurityLookup generates runtime cache lookup code for security() calls
func GenerateSecurityLookup(call *security.SecurityCall, varName string) string {
	var code strings.Builder

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

// InjectSecurityCode updates StrategyCode with security prefetch and lookups
func InjectSecurityCode(code *StrategyCode, program *ast.Program) (*StrategyCode, error) {
	injection, err := AnalyzeAndGeneratePrefetch(program)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze security calls: %w", err)
	}

	if injection.PrefetchCode == "" {
		return code, nil
	}

	functionBody := code.FunctionBody
	updatedBody := injection.PrefetchCode + functionBody
	mergedImports := mergeImports(code.AdditionalImports, injection.ImportPaths)

	return &StrategyCode{
		UserDefinedFunctions: code.UserDefinedFunctions,
		FunctionBody:         updatedBody,
		StrategyName:         code.StrategyName,
		AdditionalImports:    mergedImports,
	}, nil
}

// buildMapperCode emits the mapper initialisation block for one security symbol.
//
// For variable-bar-count modifiers (Renko, Kagi, …) the same-TF branch uses
// BuildMappingFromTransform; all other same-TF cases use BuildIdentityMapping.
func buildMapperCode(varName string, variableBarCount bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s_mapper := request.NewSecurityBarMapper()\n", varName))
	b.WriteString("\tif secTimeframeSeconds < baseTimeframeSeconds {\n")
	b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingForUpscaling(%s_ctx.Data, ctx.Data, ctx.Timezone)\n", varName, varName))
	b.WriteString("\t} else if secTimeframeSeconds == baseTimeframeSeconds {\n")
	if variableBarCount {
		b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingFromTransform(%s_transformResult.MainToSynthetic)\n", varName, varName))
	} else {
		b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildIdentityMapping(len(ctx.Data))\n", varName))
	}
	b.WriteString("\t} else {\n")
	b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingWithDateFilter(%s_ctx.Data, ctx.Data, baseDateRange, ctx.Timezone)\n", varName, varName))
	b.WriteString("\t}\n")
	return b.String()
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

// generateContextVarName creates unique variable name for each symbol:timeframe
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

// sanitizeVarName converts "SYMBOL:TIMEFRAME" to valid Go variable name
func sanitizeVarName(s string) string {
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return strings.ToLower(s)
}
