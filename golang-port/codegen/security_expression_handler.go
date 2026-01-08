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
	symbolTable          SymbolTable
	gen                  *generator // Access to generator for input constants
}

type SecurityExpressionConfig struct {
	IndentFunc           func() string
	IncrementIndent      func()
	DecrementIndent      func()
	SerializeExpr        func(ast.Expression) (string, error)
	MarkSecurityExprEval func()
	SymbolTable          SymbolTable
	Generator            *generator
}

func NewSecurityExpressionHandler(config SecurityExpressionConfig) *SecurityExpressionHandler {
	return &SecurityExpressionHandler{
		indentFunc:           config.IndentFunc,
		incrementIndent:      config.IncrementIndent,
		decrementIndent:      config.DecrementIndent,
		serializeExpr:        config.SerializeExpr,
		markSecurityExprEval: config.MarkSecurityExprEval,
		symbolTable:          config.SymbolTable,
		gen:                  config.Generator,
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

	// Generate evaluator initialization with variable registry and bar mapper support
	h.markSecurityExprEval()
	code += h.indentFunc() + "if secBarEvaluator == nil {\n"
	h.incrementIndent()

	// EVIDENCE GATHERING: Log security context initialization
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] Creating evaluator for security() expression\")\n"

	code += h.indentFunc() + "baseEvaluator := security.NewStreamingBarEvaluator()\n"
	code += h.indentFunc() + "varRegistry := security.NewVariableRegistry()\n"
	code += h.indentFunc() + "baseEvaluator.SetVariableRegistry(varRegistry)\n"
	code += h.indentFunc() + "barMapper := security.NewBarIndexMapper()\n"

	// EVIDENCE GATHERING: Log bar mapping setup
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] Setting up bar mapper\")\n"

	// Convert request.BarRange to security.BarRange to populate mapper
	code += h.indentFunc() + "requestRanges := securityBarMapper.GetRanges()\n"
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] Bar mapper has %d ranges\", len(requestRanges))\n"
	code += h.indentFunc() + "for _, rr := range requestRanges {\n"
	h.incrementIndent()
	code += h.indentFunc() + "if rr.StartHourlyIndex >= 0 {\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] Mapping: DailyBarIndex=%d → StartHourlyIndex=%d\", rr.DailyBarIndex, rr.StartHourlyIndex)\n"
	code += h.indentFunc() + "barMapper.SetMapping(rr.DailyBarIndex, rr.StartHourlyIndex)\n"
	h.decrementIndent()
	code += h.indentFunc() + "} else {\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] ⚠️  SKIPPED: DailyBarIndex=%d has negative StartHourlyIndex=%d\", rr.DailyBarIndex, rr.StartHourlyIndex)\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"
	code += h.indentFunc() + "baseEvaluator.SetBarIndexMapper(barMapper)\n"

	// CRITICAL ASSESSMENT: Is bar mapper complete?
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] ✅ Bar mapper configured with %d mappings\", len(requestRanges))\n"

	// Set up main context variable lookup fallback (PineScript lexical scoping)
	code += h.indentFunc() + "baseEvaluator.SetVarLookup(func(varName string, secBarIdx int) (*series.Series, int, bool) {\n"
	h.incrementIndent()

	// EVIDENCE GATHERING: Log variable lookup attempts
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] Request: varName=%q secBarIdx=%d\", varName, secBarIdx)\n"

	code += h.indentFunc() + "var varSeries *series.Series\n"
	code += h.indentFunc() + "switch varName {\n"

	// Generate case for each series variable in the symbol table
	// Filter out TA function names that don't have series declarations
	taFunctions := map[string]bool{
		"minus": true, "plus": true, "sum": true, "truerange": true,
		"abs": true, "max": true, "min": true, "sign": true,
	}

	for _, symbol := range h.symbolTable.AllSymbols() {
		if symbol.Type == VariableTypeSeries {
			varName := symbol.Name
			// Skip TA function names
			if taFunctions[varName] {
				continue
			}
			code += h.indentFunc() + fmt.Sprintf("case %q:\n", varName)
			h.incrementIndent()
			code += h.indentFunc() + fmt.Sprintf("varSeries = %sSeries\n", varName)
			h.decrementIndent()
		}
	}

	code += h.indentFunc() + "default:\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] ❌ UNKNOWN VARIABLE: %q not in symbol table\", varName)\n"
	code += h.indentFunc() + "return nil, -1, false\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"
	code += h.indentFunc() + "if varSeries == nil {\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] ❌ NIL SERIES: %q found in switch but Series is nil\", varName)\n"
	code += h.indentFunc() + "return nil, -1, false\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	// EVIDENCE GATHERING: Bar mapping assessment
	code += h.indentFunc() + "mainIdx := barMapper.GetMainBarIndexForSecurityBar(secBarIdx)\n"
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] ✅ Resolved: varName=%q secBarIdx=%d → mainIdx=%d seriesCursor=%d seriesCapacity=%d\", varName, secBarIdx, mainIdx, varSeries.Position(), varSeries.Capacity())\n"

	// CRITICAL ASSESSMENT: Check if this is a bandaid
	code += h.indentFunc() + "if mainIdx < 0 {\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] ⚠️  NEGATIVE MAIN INDEX: secBarIdx=%d mapped to mainIdx=%d (warmup period?)\", secBarIdx, mainIdx)\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"
	code += h.indentFunc() + "if mainIdx >= varSeries.Capacity() {\n"
	h.incrementIndent()
	code += h.indentFunc() + "log.Printf(\"[VARLOOKUP] ❌ INDEX OUT OF BOUNDS: mainIdx=%d >= capacity=%d\", mainIdx, varSeries.Capacity())\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	code += h.indentFunc() + "return varSeries, mainIdx, true\n"
	h.decrementIndent()
	code += h.indentFunc() + "})\n"

	// Pass input constants to evaluator for identifier resolution
	code += h.indentFunc() + "inputConstantsMap := " + h.generateInputConstantsMap() + "\n"
	code += h.indentFunc() + "baseEvaluator.SetInputConstantsMap(inputConstantsMap)\n"

	code += h.indentFunc() + "secBarEvaluator = security.NewSeriesCachingEvaluator(baseEvaluator)\n"
	code += h.indentFunc() + "log.Printf(\"[SECURITY-INIT] ✅ Evaluator created and cached\")\n"
	h.decrementIndent()
	code += h.indentFunc() + "}\n"

	// No need to register variables - evaluator will access main context directly via fallback

	// Serialize expression for runtime evaluation (WITH offset if present)
	exprJSON, err := h.serializeExpr(exprArg)
	if err != nil {
		return "", fmt.Errorf("failed to serialize security expression: %w", err)
	}

	// EVIDENCE GATHERING: Log expression evaluation
	code += h.indentFunc() + fmt.Sprintf("log.Printf(\"[SECURITY-EVAL] Evaluating expression at secBarIdx=%%d\", %s)\n", secBarIdxVar)

	// Generate EvaluateAtBar call - runtime will extract/apply offset
	code += h.indentFunc() + fmt.Sprintf("secValue, err := secBarEvaluator.EvaluateAtBar(%s, secCtx, %s)\n", exprJSON, secBarIdxVar)
	code += h.indentFunc() + "if err != nil {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("log.Printf(\"[SECURITY-EVAL] ❌ ERROR: %%v\", err)\n")
	code += h.indentFunc() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	h.decrementIndent()
	code += h.indentFunc() + "} else {\n"
	h.incrementIndent()
	code += h.indentFunc() + fmt.Sprintf("log.Printf(\"[SECURITY-EVAL] ✅ Result: secValue=%%f for %s\", secValue)\n", varName)
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

func (h *SecurityExpressionHandler) collectVariableReferences(expr ast.Expression) []string {
	vars := make(map[string]bool)
	h.walkExpression(expr, func(node ast.Expression) {
		if ident, ok := node.(*ast.Identifier); ok {
			switch ident.Name {
			case "close", "open", "high", "low", "volume":
				// OHLCV fields - handled by evaluator
			default:
				// Only register variables that start with known prefixes indicating they're computed
				// This excludes inputs like leftBars, bb_1d_bblenght which are constants
				if hasComputedVariablePrefix(ident.Name) {
					vars[ident.Name] = true
				}
			}
		}
	})

	result := make([]string, 0, len(vars))
	for varName := range vars {
		result = append(result, varName)
	}
	return result
}

func hasComputedVariablePrefix(name string) bool {
	// Computed variables typically have patterns like:
	// bb_1d_newisOverBBTop, bb_1d_newisUnderBBBottom, etc.
	// Look for "newis" or "is" followed by uppercase (indicates boolean state variable)
	if len(name) < 4 {
		return false
	}

	// Check for common computed variable patterns
	patterns := []string{"newis", "is_", "_is"}
	for _, pattern := range patterns {
		for i := 0; i <= len(name)-len(pattern); i++ {
			if name[i:i+len(pattern)] == pattern {
				return true
			}
		}
	}

	return false
}

func (h *SecurityExpressionHandler) walkExpression(expr ast.Expression, visitor func(ast.Expression)) {
	if expr == nil {
		return
	}

	visitor(expr)

	switch e := expr.(type) {
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			h.walkExpression(arg, visitor)
		}
	case *ast.BinaryExpression:
		h.walkExpression(e.Left, visitor)
		h.walkExpression(e.Right, visitor)
	case *ast.ConditionalExpression:
		h.walkExpression(e.Test, visitor)
		h.walkExpression(e.Consequent, visitor)
		h.walkExpression(e.Alternate, visitor)
	case *ast.MemberExpression:
		h.walkExpression(e.Object, visitor)
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

func (h *SecurityExpressionHandler) generateInputConstantsMap() string {
	if h.gen.inputHandler == nil {
		return "map[string]float64(nil)"
	}

	constantsMap := h.gen.inputHandler.GetInputConstantsMap()
	if len(constantsMap) == 0 {
		return "map[string]float64(nil)"
	}

	result := "map[string]float64{"
	first := true
	for varName, value := range constantsMap {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%q: %f", varName, value)
		first = false
	}
	result += "}"
	return result
}
