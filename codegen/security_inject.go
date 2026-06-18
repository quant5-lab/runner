package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/security"
)

// SecurityInjection holds prefetch code to inject before the bar loop.
type SecurityInjection struct {
	PrefetchCode string
	ImportPaths  []string
}

type resolvedSecurityCall struct {
	call           security.SecurityCall
	resolvedSym    string
	resolvedTf     string
	isSymRuntime   bool
	isTfRuntime    bool
	modifierPrefix string // e.g. "HEIKINASHI" when symbol was "HEIKINASHI:syminfo.tickerid"
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

// AnalyzeAndGeneratePrefetch inspects program for security() calls, eliminates
// calls whose outputs are consumed exclusively by chart-drawing namespaces
// (label/line/box/table — no effect on strategy output), and emits Go prefetch
// code for each remaining call.
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

	vars := buildVariableMap(program)

	resolved := make([]resolvedSecurityCall, len(calls))
	for i, call := range calls {
		sym, symRuntime := resolveSecurityArgument(call.SymbolExpr, call.Symbol, vars)
		tf, tfRuntime := resolveSecurityArgument(call.TimeframeExpr, call.Timeframe, vars)

		modPrefix, baseSym, hasModifier := extractModifierPrefix(sym)
		if hasModifier {
			sym = baseSym
		}

		resolved[i] = resolvedSecurityCall{
			call:           call,
			resolvedSym:    sym,
			resolvedTf:     canonicalizeTimeframe(tf),
			isSymRuntime:   symRuntime,
			isTfRuntime:    tfRuntime,
			modifierPrefix: modPrefix,
		}
	}

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

	// Remove feeds whose outputs only reach chart-drawing namespaces — they have
	// no effect on strategy execution or golden output and must not cause a hard
	// exit when their fixture is absent.
	lhsIndex := buildSecurityLHSIndex(program)
	chartOnlyUDFs := NewChartOnlyUDFDetector().Detect(program)
	chartOnlyKeys := SecurityChartOnlyClassifier{}.ChartOnlySecurityKeys(resolved, lhsIndex, chartOnlyUDFs, program)
	for key := range chartOnlyKeys {
		delete(dedupMap, key)
	}

	if len(dedupMap) == 0 {
		return &SecurityInjection{
			PrefetchCode: "",
			ImportPaths:  []string{},
		}, nil
	}

	var codeBuilder strings.Builder

	codeBuilder.WriteString("\tfetcher := datafetcher.NewFileFetcher(dataDir, 0)\n\n")
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

		codeBuilder.WriteString(fetchDataBlock(varName, symbolCode, timeframeCode, warmupBars))

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
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx.Timezone = %s_tz\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx.ReferenceSession = %s_refsession\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\t%s_ctx.PeriodAnchor = %s_anchor\n", varName, varName))
		codeBuilder.WriteString(fmt.Sprintf("\tfor _, bar := range %s_data {\n", varName))
		codeBuilder.WriteString(fmt.Sprintf("\t\t%s_ctx.AddBar(bar)\n", varName))
		codeBuilder.WriteString("\t}\n")

		resolvedKey := fmt.Sprintf("%s:%s", firstCall.resolvedSym, firstCall.resolvedTf)

		mapper := mapperInitBlock(varName, hasModifier && isVariableBarCountModifier(firstCall.modifierPrefix))

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
			codeBuilder.WriteString(mapper)
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityBarMappers[%s] = %s_mapper\n\n", keyExpr, varName))
		} else {
			codeBuilder.WriteString(fmt.Sprintf("\tsecurityContexts[%q] = %s_ctx\n", resolvedKey, varName))
			codeBuilder.WriteString(mapper)
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

// GenerateSecurityLookup generates runtime cache lookup code for a security() call.
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

// InjectSecurityCode prepends security prefetch code into the strategy function body.
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

// mergeImports combines two import lists without duplicates.
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

// generateContextVarName creates a unique Go variable name for a symbol:timeframe pair.
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

// sanitizeVarName converts a "SYMBOL:TIMEFRAME" key to a valid Go identifier.
func sanitizeVarName(s string) string {
	s = strings.ReplaceAll(s, ":", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
	return strings.ToLower(s)
}
