package codegen

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
	"github.com/quant5-lab/runner/security"
)

/* StrategyCode holds generated Go code for strategy execution */
type StrategyCode struct {
	UserDefinedFunctions string   // Arrow functions defined before executeStrategy
	FunctionBody         string   // executeStrategy() function body
	StrategyName         string   // Pine Script strategy name
	AdditionalImports    []string // Additional imports needed for security() streaming evaluation
}

/* GenerateStrategyCodeFromAST converts parsed Pine ESTree to Go runtime code */
func GenerateStrategyCodeFromAST(program *ast.Program) (*StrategyCode, error) {
	constantRegistry := NewConstantRegistry()
	typeSystem := NewTypeInferenceEngine()
	boolConverter := NewBooleanConverter(typeSystem)

	variablesRegistry := make(map[string]string)
	registryGuard := NewVariableRegistryGuard(variablesRegistry)

	gen := &generator{
		imports:            make(map[string]bool),
		variables:          variablesRegistry,
		varInits:           make(map[string]ast.Expression),
		constants:          make(map[string]interface{}),
		reassignedVars:     make(map[string]bool),
		strategyConfig:     NewStrategyConfig(),
		limits:             NewCodeGenerationLimits(),
		safetyGuard:        NewRuntimeSafetyGuard(),
		persistenceEmitter: NewVarPersistenceEmitter(NewRuntimeSafetyGuard()),
		loopContextStack:   NewLoopContextStack(),
		constantRegistry:   constantRegistry,
		typeSystem:         typeSystem,
		boolConverter:      boolConverter,
		registryGuard:      registryGuard,
		pineVersion:        program.PineVersion,
	}

	gen.inputHandler = NewInputHandler()
	gen.inputConstExtractor = NewInputConstantExtractor()
	gen.mathHandler = NewMathHandler()
	gen.calendarHandler = NewCalendarHandler()
	gen.timeframeFuncHandler = NewTimeframeFuncCallHandler()
	gen.colorHandler = NewColorHandler()
	gen.valueHandler = NewValueHandler()
	gen.subscriptResolver = NewSubscriptResolver()
	gen.builtinHandler = NewBuiltinIdentifierHandler()
	gen.taRegistry = NewTAFunctionRegistry()
	gen.compositeIndicatorRegistry = NewCompositeIndicatorRegistry()
	gen.compositeIndicatorRegistry.Register("ta.rsi", &RSIHandler{})
	gen.compositeIndicatorRegistry.Register("rsi", &RSIHandler{})
	gen.compositeIndicatorRegistry.Register("ta.mfi", &MFIHandler{})
	gen.compositeIndicatorRegistry.Register("mfi", &MFIHandler{})
	gen.compositeIndicatorRegistry.Register("ta.tsi", &TsiHandler{})
	gen.compositeIndicatorRegistry.Register("tsi", &TsiHandler{})
	gen.compositeIndicatorRegistry.Register("ta.hma", &HmaHandler{})
	gen.compositeIndicatorRegistry.Register("hma", &HmaHandler{})
	gen.compositeIndicatorRegistry.Register("ta.kcw", &KcwHandler{})
	gen.compositeIndicatorRegistry.Register("kcw", &KcwHandler{})
	gen.compositeIndicatorRegistry.Register("ta.sar", &SarHandler{})
	gen.compositeIndicatorRegistry.Register("sar", &SarHandler{})
	gen.compositeIndicatorRegistry.Register("ta.pivot_point_levels", &PivotPointLevelsHandler{})
	gen.compositeIndicatorRegistry.Register("pivot_point_levels", &PivotPointLevelsHandler{})
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)
	gen.tempVarMgr = NewTempVariableManager(gen)
	gen.constEvaluator = validation.NewWarmupAnalyzer()
	gen.plotExprHandler = NewPlotExpressionHandler(gen)
	gen.barFieldRegistry = NewBarFieldSeriesRegistry()
	gen.inlineRegistry = NewInlineFunctionRegistry()
	gen.runtimeOnlyFilter = NewRuntimeOnlyFunctionFilter()
	gen.inlineConditionRegistry = NewInlineConditionHandlerRegistry(gen.tempVarMgr)
	gen.plotCollector = NewPlotCollector()
	gen.callRouter = NewCallExpressionRouter()
	gen.funcSigRegistry = NewFunctionSignatureRegistry()
	gen.signatureRegistrar = NewSignatureRegistrar(gen.funcSigRegistry)
	gen.arrowContextLifecycle = NewArrowContextLifecycleManager()
	gen.returnValueStorage = NewReturnValueSeriesStorageHandler("\t")
	gen.symbolTable = NewSymbolTable()
	gen.literalFormatter = NewLiteralFormatter()
	gen.tupleIndicatorHandler = NewTupleIndicatorHandler()
	gen.directionExtractor = NewDefaultDirectionExtractor()

	gen.conditionalArgAnalyzer = NewConditionalArgumentAnalyzer(&ExpressionHasher{})
	gen.conditionalCodeGen = NewConditionalCodeGenerator(gen, gen.conditionalArgAnalyzer, gen.tempVarMgr)
	gen.securityAnalyzer = NewSecurityCallAnalyzer(gen)
	gen.udfAnalyzer = NewUDFTempVarAnalyzer(gen)
	gen.statementAnalyzer = NewStatementConditionalAnalyzer(gen)

	gen.hasSecurityCalls = detectSecurityCalls(program)
	gen.hasStrategyRuntimeAccess = detectStrategyRuntimeAccess(program)

	sessionMemberKeys := gen.builtinHandler.registry.SessionSeriesBuiltinNames()
	usageDetector := NewBuiltinUsageDetectorWithMembers(
		append(
			gen.builtinHandler.CalendarBuiltinNames(),
			"bar_index", "last_bar_index", "last_bar_time", "timenow",
			"time_close", "time_tradingday",
		),
		append(sessionMemberKeys, VolumeIndicatorMemberKeys()...),
	)
	detected := usageDetector.Detect(program)
	gen.hasBarIndexUsage = detected["bar_index"]
	gen.hasLastBarIndex = detected["last_bar_index"]
	gen.hasLastBarTime = detected["last_bar_time"]
	gen.hasTimenow = detected["timenow"]
	gen.builtinSeriesLifecycle = NewCompositeSeriesLifecycle(
		NewCalendarSeriesLifecycle(
			gen.builtinHandler.ResolveCalendarBuiltins(detected),
		),
		NewTimeSeriesLifecycle(
			detected["time_close"],
			detected["time_tradingday"],
		),
		NewSessionSeriesLifecycle(
			detected["session.isfirstbar"],
			detected["session.islastbar"],
			detected["session.isfirstbar_regular"],
			detected["session.islastbar_regular"],
		),
		NewVolumeIndicatorLifecycle(detected),
	)
	gen.seriesInitCoercer = NewSeriesInitCoercer()

	if err := NewLoopNestingValidator().Validate(program); err != nil {
		return nil, err
	}

	body, err := gen.generateProgram(program)
	if err != nil {
		return nil, err
	}

	if strings.Contains(body, "sort.") {
		gen.hasSortUsage = true
	}

	additionalImports := []string{}
	if gen.hasSecurityCalls {
		additionalImports = append(additionalImports, "github.com/quant5-lab/runner/security")
	}
	if gen.hasTickerCalls {
		additionalImports = append(additionalImports, "github.com/quant5-lab/runner/runtime/ticker")
	}
	if gen.hasSortUsage {
		additionalImports = append(additionalImports, "sort")
	}

	code := &StrategyCode{
		UserDefinedFunctions: gen.userDefinedFunctions,
		FunctionBody:         body,
		StrategyName:         gen.strategyConfig.Name,
		AdditionalImports:    additionalImports,
	}

	return code, nil
}

type generator struct {
	imports                  map[string]bool
	variables                map[string]string
	varInits                 map[string]ast.Expression
	constants                map[string]interface{}
	reassignedVars           map[string]bool
	plots                    []string
	strategyConfig           *StrategyConfig
	indent                   int
	userDefinedFunctions     string
	taFunctions              []taFunctionCall
	tupleTAFunctions         []tupleTAFunctionCall
	inSecurityContext        bool
	inArrowFunctionBody      bool
	loopContextStack         *LoopContextStack
	hasSecurityCalls         bool
	hasSecurityExprEvals     bool
	hasStrategyRuntimeAccess bool
	hasBarIndexUsage         bool
	hasLastBarIndex          bool
	hasLastBarTime           bool
	hasTimenow               bool
	hasTickerCalls           bool
	hasSortUsage             bool
	pineVersion              int
	limits                   CodeGenerationLimits
	safetyGuard              RuntimeSafetyGuard
	persistenceEmitter       *VarPersistenceEmitter
	hoistedArrowContexts     []ArrowCallSite

	constantRegistry *ConstantRegistry
	typeSystem       *TypeInferenceEngine
	boolConverter    *BooleanConverter
	registryGuard    *VariableRegistryGuard

	inputHandler               *InputHandler
	inputConstExtractor        *InputConstantExtractor
	mathHandler                *MathHandler
	calendarHandler            *CalendarHandler
	timeframeFuncHandler       *TimeframeFuncCallHandler
	colorHandler               *ColorHandler
	valueHandler               *ValueHandler
	subscriptResolver          *SubscriptResolver
	builtinHandler             *BuiltinIdentifierHandler
	taRegistry                 *TAFunctionRegistry
	compositeIndicatorRegistry *CompositeIndicatorRegistry
	exprAnalyzer               *ExpressionAnalyzer
	tempVarMgr                 *TempVariableManager
	constEvaluator             *validation.WarmupAnalyzer
	plotExprHandler            *PlotExpressionHandler
	barFieldRegistry           *BarFieldSeriesRegistry
	inlineRegistry             *InlineFunctionRegistry
	runtimeOnlyFilter          *RuntimeOnlyFunctionFilter
	inlineConditionRegistry    *InlineConditionHandlerRegistry
	plotCollector              *PlotCollector
	callRouter                 *CallExpressionRouter
	funcSigRegistry            *FunctionSignatureRegistry
	signatureRegistrar         *SignatureRegistrar
	arrowContextLifecycle      *ArrowContextLifecycleManager
	returnValueStorage         *ReturnValueSeriesStorageHandler
	arrowAccessResolver        *ArrowSeriesAccessResolver
	symbolTable                SymbolTable
	literalFormatter           *LiteralFormatter
	tupleIndicatorHandler      *TupleIndicatorHandler
	directionExtractor         *ChainDirectionExtractor

	conditionalArgAnalyzer *ConditionalArgumentAnalyzer
	conditionalCodeGen     *ConditionalCodeGenerator
	securityAnalyzer       *SecurityCallAnalyzer
	udfAnalyzer            *UDFTempVarAnalyzer
	statementAnalyzer      *StatementConditionalAnalyzer
	builtinSeriesLifecycle *CompositeSeriesLifecycle
	seriesInitCoercer      *SeriesInitCoercer
}

func (g *generator) buildPlotOptions(opts PlotOptions) string {
	optionsMap := make([]string, 0)

	if opts.ColorExpr != nil {
		if colorValue := g.evaluateStringConstant(opts.ColorExpr); colorValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"color\": %q", colorValue))
		} else if ident, ok := opts.ColorExpr.(*ast.Identifier); ok {
			if varType, exists := g.variables[ident.Name]; exists && varType == "string" {
				optionsMap = append(optionsMap, fmt.Sprintf("\"color\": %s", ident.Name))
			}
		} else if colorCode := g.evaluateColorCallExpression(opts.ColorExpr); colorCode != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"color\": %s", colorCode))
		}
	}

	if opts.OffsetExpr != nil {
		offsetValue := g.constEvaluator.EvaluateConstant(opts.OffsetExpr)
		if !math.IsNaN(offsetValue) && offsetValue != 0 {
			optionsMap = append(optionsMap, fmt.Sprintf("\"offset\": %d", int(offsetValue)))
		}
	}

	if opts.StyleExpr != nil {
		if styleValue := g.evaluateStringConstant(opts.StyleExpr); styleValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"style\": %q", styleValue))
		}
	}

	if opts.LineWidthExpr != nil {
		linewidthValue := g.constEvaluator.EvaluateConstant(opts.LineWidthExpr)
		if !math.IsNaN(linewidthValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"linewidth\": %d", int(linewidthValue)))
		}
	}

	if opts.TranspExpr != nil {
		transpValue := g.constEvaluator.EvaluateConstant(opts.TranspExpr)
		if !math.IsNaN(transpValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"transp\": %d", int(transpValue)))
		}
	}

	if opts.PaneExpr != nil {
		if paneValue := g.evaluateStringConstant(opts.PaneExpr); paneValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"pane\": %q", paneValue))
		}
	}

	if len(optionsMap) > 0 {
		return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(optionsMap, ", "))
	}
	return "nil"
}

func (g *generator) buildPlotOptionsWithNullColor(opts PlotOptions) string {
	optionsMap := make([]string, 0)
	optionsMap = append(optionsMap, "\"color\": nil")

	if opts.OffsetExpr != nil {
		offsetValue := g.constEvaluator.EvaluateConstant(opts.OffsetExpr)
		if !math.IsNaN(offsetValue) && offsetValue != 0 {
			optionsMap = append(optionsMap, fmt.Sprintf("\"offset\": %d", int(offsetValue)))
		}
	}

	if opts.StyleExpr != nil {
		if styleValue := g.evaluateStringConstant(opts.StyleExpr); styleValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"style\": %q", styleValue))
		}
	}

	if opts.LineWidthExpr != nil {
		linewidthValue := g.constEvaluator.EvaluateConstant(opts.LineWidthExpr)
		if !math.IsNaN(linewidthValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"linewidth\": %d", int(linewidthValue)))
		}
	}

	if opts.TranspExpr != nil {
		transpValue := g.constEvaluator.EvaluateConstant(opts.TranspExpr)
		if !math.IsNaN(transpValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"transp\": %d", int(transpValue)))
		}
	}

	if opts.PaneExpr != nil {
		if paneValue := g.evaluateStringConstant(opts.PaneExpr); paneValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"pane\": %q", paneValue))
		}
	}

	if len(optionsMap) > 0 {
		return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(optionsMap, ", "))
	}
	return "map[string]interface{}{\"color\": nil}"
}

/* color param must be a ready-to-embed Go expression (quoted literal or runtime call) */
func (g *generator) buildPlotOptionsWithColor(opts PlotOptions, color string) string {
	optionsMap := make([]string, 0)
	if color != "" {
		optionsMap = append(optionsMap, fmt.Sprintf("\"color\": %s", color))
	}

	if opts.OffsetExpr != nil {
		offsetValue := g.constEvaluator.EvaluateConstant(opts.OffsetExpr)
		if !math.IsNaN(offsetValue) && offsetValue != 0 {
			optionsMap = append(optionsMap, fmt.Sprintf("\"offset\": %d", int(offsetValue)))
		}
	}

	if opts.StyleExpr != nil {
		if styleValue := g.evaluateStringConstant(opts.StyleExpr); styleValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"style\": %q", styleValue))
		}
	}

	if opts.LineWidthExpr != nil {
		linewidthValue := g.constEvaluator.EvaluateConstant(opts.LineWidthExpr)
		if !math.IsNaN(linewidthValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"linewidth\": %d", int(linewidthValue)))
		}
	}

	if opts.TranspExpr != nil {
		transpValue := g.constEvaluator.EvaluateConstant(opts.TranspExpr)
		if !math.IsNaN(transpValue) {
			optionsMap = append(optionsMap, fmt.Sprintf("\"transp\": %d", int(transpValue)))
		}
	}

	if opts.PaneExpr != nil {
		if paneValue := g.evaluateStringConstant(opts.PaneExpr); paneValue != "" {
			optionsMap = append(optionsMap, fmt.Sprintf("\"pane\": %q", paneValue))
		}
	}

	if len(optionsMap) > 0 {
		return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(optionsMap, ", "))
	}
	return "nil"
}

func (g *generator) evaluateStringConstant(expr ast.Expression) string {
	// Handle string literals
	if lit, ok := expr.(*ast.Literal); ok {
		if strVal, ok := lit.Value.(string); ok {
			return strVal
		}
	}
	// Handle member expressions like plot.style_circles via ConstantResolver
	resolver := NewConstantResolver()
	if strVal, ok := resolver.ResolveToString(expr); ok {
		return strVal
	}
	return ""
}

func (g *generator) evaluateColorCallExpression(expr ast.Expression) string {
	call, ok := expr.(*ast.CallExpression)
	if !ok {
		return ""
	}
	funcName := g.extractFunctionName(call.Callee)
	if !g.colorHandler.CanHandle(funcName) {
		return ""
	}
	code, err := g.colorHandler.GenerateColorCall(funcName, call.Arguments, g)
	if err != nil {
		return ""
	}
	return code
}

type taFunctionCall struct {
	varName  string
	funcName string
	args     []ast.Expression
	call     *ast.CallExpression
}

type tupleTAFunctionCall struct {
	varNames []string
	funcName string
	call     *ast.CallExpression
}

func (g *generator) generateProgram(program *ast.Program) (string, error) {
	if program == nil || len(program.Body) == 0 {
		return g.generatePlaceholder(), nil
	}

	// Initialize safety limits if not already set (for tests)
	if g.limits.MaxStatementsPerPass == 0 {
		g.limits = NewCodeGenerationLimits()
		g.safetyGuard = NewRuntimeSafetyGuard()
		g.persistenceEmitter = NewVarPersistenceEmitter(g.safetyGuard)
	}

	// PRE-PASS: Collect AST constants for expression evaluator
	for _, stmt := range program.Body {
		g.constEvaluator.CollectConstants(stmt)
	}

	// First pass: collect variables, analyze Series requirements, extract strategy name
	statementCounter := NewStatementCounter(g.limits)
	registrar := NewVariableDeclarationRegistrar(g)
	nestedScanner := NewNestedVariableScanner(g)
	for _, stmt := range program.Body {
		if err := statementCounter.Increment(); err != nil {
			return "", err
		}
		// Extract strategy name from indicator() or strategy() calls
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			if call, ok := exprStmt.Expression.(*ast.CallExpression); ok {
				if member, ok := call.Callee.(*ast.MemberExpression); ok {
					// Extract function name from ta.sma or strategy.entry
					obj := ""
					if id, ok := member.Object.(*ast.Identifier); ok {
						obj = id.Name
					}
					prop := ""
					if id, ok := member.Property.(*ast.Identifier); ok {
						prop = id.Name
					}
					funcName := obj + "." + prop

					if funcName == "indicator" || funcName == "strategy" {
						metaHandler := NewMetaFunctionHandler()
						_, _ = metaHandler.GenerateCode(g, call)
					}
				}
				if id, ok := call.Callee.(*ast.Identifier); ok {
					if id.Name == "study" || id.Name == "indicator" || id.Name == "strategy" {
						metaHandler := NewMetaFunctionHandler()
						_, _ = metaHandler.GenerateCode(g, call)
					}
				}
			}
		}

		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if g.tryResolveInputConstant(declarator) {
					continue
				}
				registrar.RegisterDeclarator(declarator)
			}
		}

		if ifStmt, ok := stmt.(*ast.IfStatement); ok {
			nestedScanner.ScanIfBlock(ifStmt)
		}
	}

	// Sync constants to typeSystem and constEvaluator
	for varName, value := range g.constants {
		g.typeSystem.RegisterConstant(varName, value)

		if floatVal, ok := value.(float64); ok {
			g.constEvaluator.AddConstant(varName, floatVal)
		} else if intVal, ok := value.(int); ok {
			g.constEvaluator.AddConstant(varName, float64(intVal))
		}
	}

	// Scan for reassignments (Kind="var") to skip initial assignments (Kind="let")
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			if varDecl.Kind == "var" {
				for _, declarator := range varDecl.Declarations {
					if id, ok := declarator.ID.(*ast.Identifier); ok {
						g.reassignedVars[id.Name] = true
					}
				}
			}
		}
		nestedScanner.ScanReassignments(stmt)
	}

	// Generate user-defined functions at module level
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				id, ok := declarator.ID.(*ast.Identifier)
				if !ok {
					continue
				}
				if arrowFunc, ok := declarator.Init.(*ast.ArrowFunctionExpression); ok {
					g.variables[id.Name] = "function"

					savedIndent := g.indent
					g.indent = 0

					arrowCodegen := NewArrowFunctionCodegen(g)
					funcCode, err := arrowCodegen.Generate(id.Name, arrowFunc)
					if err != nil {
						g.indent = savedIndent
						return "", fmt.Errorf("failed to generate arrow function %s: %w", id.Name, err)
					}

					g.userDefinedFunctions += funcCode
					g.indent = savedIndent
				}
			}
		}
	}

	// Third pass: collect TA function calls for pre-calculation
	statementCounter.Reset()
	for _, stmt := range program.Body {
		if err := statementCounter.Increment(); err != nil {
			return "", err
		}
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if ap, ok := declarator.ID.(*ast.ArrayPattern); ok {
					if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
						funcName := g.extractFunctionName(callExpr.Callee)
						if g.tupleIndicatorHandler.IsCustomHandler(funcName) && len(ap.Elements) > 0 {
							varNames := make([]string, len(ap.Elements))
							for i, el := range ap.Elements {
								varNames[i] = el.Name
							}
							g.tupleTAFunctions = append(g.tupleTAFunctions, tupleTAFunctionCall{
								varNames: varNames,
								funcName: funcName,
								call:     callExpr,
							})
						}
					}
					continue
				}

				if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
					funcName := g.extractFunctionName(callExpr.Callee)
					if g.taRegistry.IsSupported(funcName) {
						if id, ok := declarator.ID.(*ast.Identifier); ok {
							g.taFunctions = append(g.taFunctions, taFunctionCall{
								varName:  id.Name,
								funcName: funcName,
								args:     callExpr.Arguments,
								call:     callExpr,
							})
						}
					}
				}
			}
		}
	}

	code := ""

	code += g.ind() + fmt.Sprintf("strat.CallWithPyramiding(%q, %.0f, %d)\n", g.strategyConfig.Name, g.strategyConfig.InitialCapital, g.strategyConfig.Pyramiding)
	if g.strategyConfig.CommissionType != "" {
		code += g.ind() + fmt.Sprintf("strat.SetCommission(%.10g, %q)\n", g.strategyConfig.CommissionValue, g.strategyConfig.CommissionType)
	}
	if g.strategyConfig.DefaultQtyType != "" {
		code += g.ind() + fmt.Sprintf("strat.SetDefaultQty(%.10g, %q)\n", g.strategyConfig.DefaultQtyValue, g.strategyConfig.DefaultQtyType)
	}
	code += "\n"

	if g.inputHandler != nil && len(g.inputHandler.inputConstants) > 0 {
		code += g.ind() + "// Input constants\n"
		for _, constCode := range g.inputHandler.inputConstants {
			code += g.ind() + constCode
		}
		code += "\n"
	}

	/* Declare internal series for composite indicators using metadata discovery */
	for _, taFunc := range g.taFunctions {
		seriesNames := g.compositeIndicatorRegistry.GetInternalSeriesNames(taFunc.funcName, taFunc.varName, taFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("var %sSeries *series.Series\n", seriesName)
		}
	}
	for _, tupleFunc := range g.tupleTAFunctions {
		seriesNames := g.tupleIndicatorHandler.InternalSeriesNamesFor(tupleFunc.funcName, tupleFunc.varNames[0], tupleFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("var %sSeries *series.Series\n", seriesName)
		}
	}

	code += g.ind() + "// Series storage (ForwardSeriesBuffer paradigm)\n"
	for _, seriesName := range g.barFieldRegistry.AllSeriesNames() {
		code += g.ind() + fmt.Sprintf("var %s *series.Series\n", seriesName)
		// Register as series type (remove "Series" suffix to get variable name)
		if g.symbolTable != nil && len(seriesName) > 6 && seriesName[len(seriesName)-6:] == "Series" {
			varName := seriesName[:len(seriesName)-6]
			g.symbolTable.Register(varName, VariableTypeSeries)
		}
	}
	if g.hasBarIndexUsage {
		code += g.ind() + "var bar_indexSeries *series.Series\n"
		if g.symbolTable != nil {
			g.symbolTable.Register("bar_index", VariableTypeSeries)
		}
	}
	code += g.ind() + "var timeSeries *series.Series\n"
	if g.symbolTable != nil {
		g.symbolTable.Register("time", VariableTypeSeries)
	}
	code += g.builtinSeriesLifecycle.GenerateDeclarations(g.ind())
	g.builtinSeriesLifecycle.GenerateSymbolTableRegistrations(g.symbolTable)

	if len(g.variables) > 0 {
		for varName, varType := range g.variables {
			if varType == "function" {
				if g.symbolTable != nil {
					g.symbolTable.Register(varName, VariableTypeFunction)
				}
				continue
			}
			if varType == "string" {
				code += g.ind() + fmt.Sprintf("var %s string\n", varName)
				if g.symbolTable != nil {
					g.symbolTable.Register(varName, VariableTypeScalar)
				}
				continue
			}
			if varType == "array_series" {
				code += g.ind() + fmt.Sprintf("var %sArraySeries *series.ArraySeries\n", varName)
				continue
			}
			code += g.ind() + fmt.Sprintf("var %sSeries *series.Series\n", varName)
			if g.symbolTable != nil {
				g.symbolTable.Register(varName, VariableTypeSeries)
			}
		}
	}
	code += "\n"

	/* Scan for inline TA calls requiring hoisting (must run before GenerateDeclarations) */
	inlineScanner := NewInlineExpressionScanner(g)
	hoistableCalls := inlineScanner.ScanProgram(program)
	for _, callInfo := range hoistableCalls {
		g.tempVarMgr.GetOrCreate(callInfo)
	}

	if g.hasSecurityCalls {
		code += g.ind() + "// StreamingBarEvaluator for security() expressions\n"
		code += g.ind() + "var secBarEvaluator security.BarEvaluator\n"
		code += "\n"
	}

	tempVarDecls := g.tempVarMgr.GenerateDeclarations()
	if tempVarDecls != "" {
		code += tempVarDecls + "\n"
	}

	hasFixnan := false
	for _, taFunc := range g.taFunctions {
		if taFunc.funcName == "fixnan" {
			hasFixnan = true
			break
		}
	}
	if hasFixnan {
		code += g.ind() + "// State variables for fixnan forward-fill\n"
		for _, taFunc := range g.taFunctions {
			if taFunc.funcName == "fixnan" {
				code += g.ind() + fmt.Sprintf("var fixnanState_%s = math.NaN()\n", taFunc.varName)
			}
		}
		code += "\n"
	}

	/* OHLCV bar fields always initialized (unconditionally populated in bar loop) */
	code += g.ind() + "// Initialize Series storage\n"
	for _, seriesName := range g.barFieldRegistry.AllSeriesNames() {
		code += g.ind() + fmt.Sprintf("%s = series.NewSeries(len(ctx.Data))\n", seriesName)
	}
	if g.hasBarIndexUsage {
		code += g.ind() + "bar_indexSeries = series.NewSeries(len(ctx.Data))\n"
	}
	code += g.ind() + "timeSeries = series.NewSeries(len(ctx.Data))\n"
	code += g.builtinSeriesLifecycle.GenerateInitializations(g.ind())

	/* Initialize internal series for composite indicators using metadata discovery */
	for _, taFunc := range g.taFunctions {
		seriesNames := g.compositeIndicatorRegistry.GetInternalSeriesNames(taFunc.funcName, taFunc.varName, taFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", seriesName)
		}
	}
	for _, tupleFunc := range g.tupleTAFunctions {
		seriesNames := g.tupleIndicatorHandler.InternalSeriesNamesFor(tupleFunc.funcName, tupleFunc.varNames[0], tupleFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", seriesName)
		}
	}

	if len(g.variables) > 0 {
		for varName, varType := range g.variables {
			if varType == "function" || varType == "string" {
				continue
			}
			if varType == "array_series" {
				code += g.ind() + fmt.Sprintf("%sArraySeries = series.NewArraySeries(len(ctx.Data))\n", varName)
				continue
			}
			code += g.ind() + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", varName)
		}
	}

	tempVarInits := g.tempVarMgr.GenerateInitializations()
	if tempVarInits != "" {
		code += tempVarInits
	}
	code += "\n"

	/* Register series in main context for security() variable resolution */
	if len(g.variables) > 0 || len(g.barFieldRegistry.AllSeriesNames()) > 0 {
		code += g.ind() + "// Register series for context hierarchy variable resolution\n"

		/* Register OHLCV bar fields */
		for _, seriesName := range g.barFieldRegistry.AllSeriesNames() {
			code += g.ind() + fmt.Sprintf("ctx.RegisterSeries(%q, %s)\n", seriesName, seriesName)
		}
		if g.hasBarIndexUsage {
			code += g.ind() + `ctx.RegisterSeries("bar_indexSeries", bar_indexSeries)` + "\n"
		}
		code += g.ind() + `ctx.RegisterSeries("timeSeries", timeSeries)` + "\n"
		code += g.builtinSeriesLifecycle.GenerateRegistrations(g.ind())

		/* Register user variables */
		for varName, varType := range g.variables {
			if varType == "function" || varType == "string" || varType == "array_series" {
				continue
			}
			code += g.ind() + fmt.Sprintf("ctx.RegisterSeries(%q, %sSeries)\n", varName, varName)
		}

		code += "\n"
	}

	// StateManager for strategy.* runtime values (Series storage)
	if g.hasStrategyRuntimeAccess {
		code += g.ind() + "sm := strategy.NewStateManager(len(ctx.Data))\n"
		for _, binding := range strategySeriesBindings() {
			code += g.ind() + fmt.Sprintf("%s := sm.%s()\n", binding.varName, binding.accessor)
			code += g.ind() + fmt.Sprintf("ctx.RegisterSeries(%q, %s)\n", binding.varName, binding.varName)
		}
		code += g.ind() + "tradeAccessor := strategy.NewTradeAccessor(strat.GetTradeHistory())\n"
		code += g.ind() + "_ = tradeAccessor\n"
		code += "\n"
	}

	scanner := NewArrowCallSiteScanner(g.variables)
	callSites := scanner.ScanForArrowFunctionCalls(program)

	secDetector := NewArrowSecurityDetector()
	for i := range callSites {
		callSites[i].NeedsSecurity = secDetector.FunctionContainsSecurityCall(callSites[i].FunctionName, program)
	}

	g.hoistedArrowContexts = callSites

	if len(callSites) > 0 {
		hoister := NewArrowContextHoister(g.ind())
		hoistedCode := hoister.GeneratePreLoopDeclarations(callSites)
		if hoistedCode != "" {
			code += g.ind() + "// Pre-allocate ArrowContext (persistent across bars)\n"
			code += hoistedCode
			code += "\n"

			for _, site := range callSites {
				g.arrowContextLifecycle.MarkAsHoisted(site.ContextVar)
			}
		}
	}

	code += g.builtinSeriesLifecycle.GenerateTimezoneSetup(g.ind())
	if g.hasLastBarIndex {
		code += g.ind() + "last_bar_index := float64(len(ctx.Data) - 1)\n"
	}
	if g.hasLastBarTime {
		code += g.ind() + "last_bar_time := float64(ctx.Data[len(ctx.Data)-1].Time * 1000)\n"
	}
	if g.hasTimenow {
		code += g.ind() + "timenow := float64(ctx.Data[len(ctx.Data)-1].Time * 1000)\n"
	}

	// Bar loop for strategy execution
	code += g.ind() + "const maxBars = 1000000\n"
	code += g.ind() + "barCount := len(ctx.Data)\n"
	code += g.ind() + "if barCount > maxBars {\n"
	g.indent++
	code += g.ind() + `fmt.Fprintf(os.Stderr, "Error: bar count (%d) exceeds safety limit (%d)\n", barCount, maxBars)` + "\n"
	code += g.ind() + "os.Exit(1)\n"
	g.indent--
	code += g.ind() + "}\n"
	iterVar := g.safetyGuard.GenerateIterationVariableReference()
	code += g.ind() + fmt.Sprintf("for %s := 0; %s < barCount; %s++ {\n", iterVar, iterVar, iterVar)
	g.indent++
	code += g.ind() + fmt.Sprintf("ctx.BarIndex = %s\n", iterVar)
	code += g.ind() + fmt.Sprintf("bar := ctx.Data[%s]\n", iterVar)
	code += g.ind() + "strat.OnBarUpdate(i, bar.Open, bar.Time)\n"

	code += g.ind() + "closeSeries.Set(bar.Close)\n"
	code += g.ind() + "highSeries.Set(bar.High)\n"
	code += g.ind() + "lowSeries.Set(bar.Low)\n"
	code += g.ind() + "openSeries.Set(bar.Open)\n"
	code += g.ind() + "volumeSeries.Set(bar.Volume)\n"
	if g.hasBarIndexUsage {
		code += g.ind() + fmt.Sprintf("bar_indexSeries.Set(float64(%s))\n", iterVar)
	}
	code += g.ind() + "timeSeries.Set(float64(bar.Time * 1000))\n"
	code += g.builtinSeriesLifecycle.GenerateBarPopulation(g.ind(), iterVar)
	code += "\n"

	/* Sample strategy state before Pine statements execute (ForwardSeriesBuffer paradigm) */
	if g.hasStrategyRuntimeAccess {
		code += g.ind() + "sm.SampleCurrentBar(strat, bar.Close, bar.High, bar.Low)\n"
	}
	code += "\n"

	/* Interleaved emission — period .Set() must precede .Get(0) within the same bar */
	statementCounter.Reset()
	for stmtIdx, stmt := range program.Body {
		stmtCalcs, err := g.tempVarMgr.GenerateCalculationsForStatement(stmtIdx)
		if err != nil {
			return "", fmt.Errorf("failed to generate temp var calculations for statement %d: %w", stmtIdx, err)
		}
		if stmtCalcs != "" {
			code += stmtCalcs
			code += "\n"
		}

		if err := statementCounter.Increment(); err != nil {
			return "", err
		}
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		code += stmtCode
	}

	if g.plotCollector != nil && g.plotCollector.HasPlots() {
		for _, plotStmt := range g.plotCollector.GetPlots() {
			code += g.ind() + plotStmt.code
		}
	}

	if g.hasStrategyRuntimeAccess {
		code += g.ind() + "strat.OnBarMetrics(bar.High, bar.Low)\n"
	}

	code += "\n" + g.ind() + "// Suppress unused variable warnings\n"
	if g.hasSecurityCalls {
		code += g.ind() + "_ = secBarEvaluator\n"
	}
	if g.hasStrategyRuntimeAccess {
		code += g.ind() + "_ = strategy_position_avg_priceSeries\n"
		code += g.ind() + "_ = strategy_position_sizeSeries\n"
		code += g.ind() + "_ = strategy_equitySeries\n"
		code += g.ind() + "_ = strategy_netprofitSeries\n"
		code += g.ind() + "_ = strategy_closedtradesSeries\n"
	}
	for varName, varType := range g.variables {
		if varType == "function" {
			continue
		}
		if varType == "string" {
			code += g.ind() + fmt.Sprintf("_ = %s\n", varName)
			continue
		}
		if varType == "array_series" {
			code += g.ind() + fmt.Sprintf("_ = %sArraySeries\n", varName)
			continue
		}
		/* Skip input constants - they don't have Series versions */
		if g.inputHandler != nil && g.inputHandler.IsInputConstant(varName) {
			continue
		}
		code += g.ind() + fmt.Sprintf("_ = %sSeries\n", varName)
	}
	code += g.builtinSeriesLifecycle.GenerateSuppressUnused(g.ind())
	if g.hasLastBarIndex {
		code += g.ind() + "_ = last_bar_index\n"
	}
	if g.hasLastBarTime {
		code += g.ind() + "_ = last_bar_time\n"
	}
	if g.hasTimenow {
		code += g.ind() + "_ = timenow\n"
	}

	// Advance Series cursors at end of bar loop
	code += "\n" + g.ind() + "// Advance Series cursors\n"

	for _, seriesName := range g.barFieldRegistry.AllSeriesNames() {
		code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %s.Next() }\n", iterVar, seriesName)
	}
	if g.hasBarIndexUsage {
		code += g.ind() + fmt.Sprintf("if %s < barCount-1 { bar_indexSeries.Next() }\n", iterVar)
	}
	code += g.ind() + fmt.Sprintf("if %s < barCount-1 { timeSeries.Next() }\n", iterVar)
	code += g.builtinSeriesLifecycle.GenerateAdvancement(g.ind(), iterVar)

	for varName, varType := range g.variables {
		if varType == "function" || varType == "string" {
			continue
		}
		if varType == "array_series" {
			code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %sArraySeries.Next() }\n", iterVar, varName)
			continue
		}
		if g.inputHandler != nil && g.inputHandler.IsInputConstant(varName) {
			continue
		}
		code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %sSeries.Next() }\n", iterVar, varName)
	}

	// Advance temp variable Series cursors (ForwardSeriesBuffer paradigm)
	tempVarNextCalls := g.tempVarMgr.GenerateNextCalls()
	if tempVarNextCalls != "" {
		code += tempVarNextCalls
	}

	// Advance internal series for composite indicators
	for _, taFunc := range g.taFunctions {
		seriesNames := g.compositeIndicatorRegistry.GetInternalSeriesNames(taFunc.funcName, taFunc.varName, taFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %sSeries.Next() }\n", iterVar, seriesName)
		}
	}
	for _, tupleFunc := range g.tupleTAFunctions {
		seriesNames := g.tupleIndicatorHandler.InternalSeriesNamesFor(tupleFunc.funcName, tupleFunc.varNames[0], tupleFunc.call)
		for _, seriesName := range seriesNames {
			code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %sSeries.Next() }\n", iterVar, seriesName)
		}
	}

	if len(g.hoistedArrowContexts) > 0 {
		for _, site := range g.hoistedArrowContexts {
			code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %s.AdvanceAll() }\n", iterVar, site.ContextVar)
		}
	}

	if g.hasStrategyRuntimeAccess {
		code += g.ind() + fmt.Sprintf("if %s < barCount-1 { sm.AdvanceCursors() }\n", iterVar)
	}

	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (g *generator) generateStatement(node ast.Node) (string, error) {
	switch n := node.(type) {
	case *ast.ExpressionStatement:
		return g.generateExpression(n.Expression)
	case *ast.VariableDeclaration:
		return g.generateVariableDeclaration(n)
	case *ast.IfStatement:
		return g.generateIfStatement(n)
	case *ast.ForStatement:
		return g.generateForStatement(n)
	case *ast.ForInStatement:
		return g.generateForInStatement(n)
	case *ast.WhileStatement:
		return g.generateWhileStatement(n)
	case *ast.BreakStatement:
		return g.ind() + "break\n", nil
	case *ast.ContinueStatement:
		return g.ind() + "continue\n", nil
	default:
		return "", fmt.Errorf("unsupported statement type: %T", node)
	}
}

func (g *generator) generateExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ForStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateForExpressionAsIIFE(e)
	case *ast.ForInStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateForInExpressionAsIIFE(e)
	case *ast.WhileStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateWhileExpressionAsIIFE(e)
	case *ast.IfStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateIfExpressionAsIIFE(e)
	case *ast.CallExpression:
		return g.generateCallExpression(e)
	case *ast.BinaryExpression:
		return g.generateBinaryExpression(e)
	case *ast.LogicalExpression:
		return g.generateLogicalExpression(e)
	case *ast.ConditionalExpression:
		return g.generateConditionalExpression(e)
	case *ast.UnaryExpression:
		return g.generateUnaryExpression(e)
	case *ast.Identifier:
		// In arrow function context or as call argument, return identifier directly
		if g.inArrowFunctionBody {
			// Check if it's a builtin identifier
			if code, resolved := g.builtinHandler.TryResolveIdentifier(e, g.accessScope()); resolved {
				return code, nil
			}
			// Check if it's a function parameter or variable
			if _, exists := g.variables[e.Name]; exists {
				return e.Name, nil
			}
			// Check if it's a constant
			if _, exists := g.constants[e.Name]; exists {
				return e.Name, nil
			}
			return e.Name, nil
		}
		return g.ind() + "// " + e.Name + "\n", nil
	case *ast.Literal:
		return g.generateLiteral(e)
	case *ast.MemberExpression:
		return g.generateMemberExpression(e)
	case *ast.ObjectExpression:
		return "", fmt.Errorf("ObjectExpression should not reach generateExpression - call handlers must use ArgumentExtractor for named arguments")
	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (g *generator) generateCallExpression(call *ast.CallExpression) (string, error) {
	// Lazy-initialize callRouter if not set (for tests)
	if g.callRouter == nil {
		g.callRouter = NewCallExpressionRouter()
	}

	// Delegate to registered handlers via router
	return g.callRouter.RouteCall(g, call)
}

func (g *generator) generateIfStatement(ifStmt *ast.IfStatement) (string, error) {
	condition, err := g.generateConditionExpression(ifStmt.Test)
	if err != nil {
		return "", err
	}

	condition = g.addBoolConversionIfNeeded(ifStmt.Test, condition)

	code := g.ind() + fmt.Sprintf("if %s {\n", condition)
	g.indent++

	bodyCode, err := g.generateIfBody(ifStmt.Consequent)
	if err != nil {
		return "", err
	}
	code += bodyCode

	g.indent--

	alternateCode, err := g.generateIfAlternate(ifStmt.Alternate)
	if err != nil {
		return "", err
	}
	code += alternateCode

	return code, nil
}

func (g *generator) generateForStatement(forStmt *ast.ForStatement) (string, error) {
	counterVar := forStmt.Counter

	g.loopContextStack.Push(counterVar)

	fromCode, err := g.generateArrowFunctionExpression(forStmt.From)
	if err != nil {
		g.loopContextStack.Pop()
		return "", err
	}

	toCode, err := g.generateArrowFunctionExpression(forStmt.To)
	if err != nil {
		g.loopContextStack.Pop()
		return "", err
	}

	stepCode := "1"
	if forStmt.Step != nil {
		stepCode, err = g.generateArrowFunctionExpression(forStmt.Step)
		if err != nil {
			g.loopContextStack.Pop()
			return "", err
		}
	}

	code := g.ind() + fmt.Sprintf("{\n")
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := int(%s)\n", counterVar, fromCode)
	code += g.ind() + fmt.Sprintf("_to := int(%s)\n", toCode)
	code += g.ind() + fmt.Sprintf("_step := int(%s)\n", stepCode)

	code += g.ind() + fmt.Sprintf("if _step == 0 {\n")
	g.indent++
	code += g.ind() + fmt.Sprintf("panic(\"for loop step cannot be zero\")\n")
	g.indent--
	code += g.ind() + fmt.Sprintf("}\n")

	code += g.ind() + fmt.Sprintf("_ascending := _step > 0\n")
	code += g.ind() + fmt.Sprintf("for ; (_ascending && %s <= _to) || (!_ascending && %s >= _to); %s += _step {\n", counterVar, counterVar, counterVar)
	g.indent++

	for _, stmt := range forStmt.Body {
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			g.loopContextStack.Pop()
			return "", err
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	g.loopContextStack.Pop()

	return code, nil
}

func (g *generator) generateForInStatement(forIn *ast.ForInStatement) (string, error) {
	/* Index var enables float64() wrapping in binary expressions; empty string still gates IsInLoop */
	counterVar := ""
	if forIn.IndexVar != "" {
		counterVar = forIn.IndexVar
	}
	g.loopContextStack.Push(counterVar)
	defer g.loopContextStack.Pop()

	collCode, err := g.generateArrowFunctionExpression(forIn.Collection)
	if err != nil {
		return "", fmt.Errorf("for-in collection: %w", err)
	}

	indexVar := "_"
	if forIn.IndexVar != "" {
		indexVar = forIn.IndexVar
	}

	code := g.ind() + fmt.Sprintf("for %s, %s := range %s {\n", indexVar, forIn.ElementVar, collCode)
	g.indent++

	for _, stmt := range forIn.Body {
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("for-in body: %w", err)
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (g *generator) generateWhileStatement(whileStmt *ast.WhileStatement) (string, error) {
	g.loopContextStack.Push("")
	defer g.loopContextStack.Pop()

	condition, err := g.generateConditionExpression(whileStmt.Condition)
	if err != nil {
		return "", fmt.Errorf("while condition: %w", err)
	}
	condition = g.addBoolConversionIfNeeded(whileStmt.Condition, condition)

	guard := NewLoopIterationGuard()

	code := g.ind() + "{\n"
	g.indent++

	code += guard.InitCode(g.ind())
	code += g.ind() + fmt.Sprintf("for %s {\n", condition)
	g.indent++

	code += guard.CheckCode(g.ind())

	for _, stmt := range whileStmt.Body {
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", fmt.Errorf("while body: %w", err)
		}
		if stmtCode != "" {
			code += stmtCode
		}
	}

	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (g *generator) generateBinaryExpression(binExpr *ast.BinaryExpression) (string, error) {
	isInLoop := g.loopContextStack != nil && g.loopContextStack.IsInLoop()
	if g.inArrowFunctionBody || isInLoop {
		formatter := NewBinaryExpressionFormatter(g.generateArrowFunctionExpression)
		return formatter.Format(binExpr)
	}

	/* Series context: Binary expressions should be in condition context */
	return "", fmt.Errorf("binary expression should be used in condition context")
}

func (g *generator) generateArrowFunctionExpression(expr ast.Expression) (string, error) {
	wasInArrow := g.inArrowFunctionBody
	g.inArrowFunctionBody = true
	defer func() { g.inArrowFunctionBody = wasInArrow }()

	switch e := expr.(type) {
	case *ast.Identifier:
		isInLoop := g.loopContextStack != nil && g.loopContextStack.IsInLoop()

		if isInLoop && g.loopContextStack.IsLoopCounter(e.Name) {
			return fmt.Sprintf("float64(%s)", e.Name), nil
		}

		if e.Name == "bar_index" && isInLoop {
			return "bar_indexSeries.GetCurrent()", nil
		}

		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, ArrowScope); resolved {
			return code, nil
		}

		if varType, exists := g.variables[e.Name]; exists {
			if varType == "float" || varType == "float64" || varType == "bool" {
				return e.Name + "Series.GetCurrent()", nil
			}
			if varType == "function" {
				return e.Name, nil
			}
		}

		if constVal, isConstant := g.constants[e.Name]; isConstant {
			if constVal == "input.source" {
				return e.Name + "Series.GetCurrent()", nil
			}
			return e.Name, nil
		}

		return e.Name, nil

	case *ast.Literal:
		return fmt.Sprintf("%v", e.Value), nil

	case *ast.CallExpression:
		return g.generateCallExpression(e)

	case *ast.BinaryExpression:
		return g.generateBinaryExpression(e)

	case *ast.LogicalExpression:
		return g.generateLogicalExpression(e)

	case *ast.MemberExpression:
		if e.Computed && g.subscriptResolver != nil {
			if obj, ok := e.Object.(*ast.Identifier); ok {
				return g.subscriptResolver.ResolveSubscript(obj.Name, e.Property, g), nil
			}
		}
		if code, resolved := g.builtinHandler.TryResolveMemberExpression(e, ArrowScope); resolved {
			return code, nil
		}
		return g.generateMemberExpression(e)

	case *ast.ConditionalExpression:
		return g.generateConditionalExpression(e)

	case *ast.UnaryExpression:
		return g.generateUnaryExpressionInArrowContext(e)

	default:
		return "", fmt.Errorf("unsupported arrow function expression type: %T", expr)
	}
}

func (g *generator) generateUnaryExpressionInArrowContext(unaryExpr *ast.UnaryExpression) (string, error) {
	operandCode, err := g.generateArrowFunctionExpression(unaryExpr.Argument)
	if err != nil {
		return "", err
	}

	op := unaryExpr.Operator
	if op == "not" {
		op = "!"
	}

	return fmt.Sprintf("%s%s", op, operandCode), nil
}

func (g *generator) generateUnaryExpression(unaryExpr *ast.UnaryExpression) (string, error) {
	operandCode, err := g.generateConditionExpression(unaryExpr.Argument)
	if err != nil {
		return "", err
	}

	op := unaryExpr.Operator
	switch op {
	case "not":
		op = "!"
	}

	return fmt.Sprintf("%s%s", op, operandCode), nil
}

func (g *generator) generateLogicalExpression(logExpr *ast.LogicalExpression) (string, error) {
	leftCode, err := g.generateConditionExpression(logExpr.Left)
	if err != nil {
		return "", err
	}

	rightCode, err := g.generateConditionExpression(logExpr.Right)
	if err != nil {
		return "", err
	}

	op := logExpr.Operator
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	}

	return fmt.Sprintf("(%s %s %s)", leftCode, op, rightCode), nil
}

func (g *generator) generateConditionalExpression(condExpr *ast.ConditionalExpression) (string, error) {
	testCode, err := g.generateConditionExpression(condExpr.Test)
	if err != nil {
		return "", err
	}

	testCode = g.addBoolConversionIfNeeded(condExpr.Test, testCode)

	consequentCode, err := g.generateNumericExpression(condExpr.Consequent)
	if err != nil {
		return "", err
	}

	alternateCode, err := g.generateNumericExpression(condExpr.Alternate)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		testCode, consequentCode, alternateCode), nil
}

// addBoolConversionIfNeeded wraps bool Series variables with conversion for boolean contexts
func (g *generator) addBoolConversionIfNeeded(expr ast.Expression, code string) string {
	return g.boolConverter.ConvertBoolSeriesForIfStatement(expr, code)
}

func (g *generator) ensureBooleanOperand(expr ast.Expression, code string) string {
	return g.boolConverter.EnsureBooleanOperand(expr, code)
}

func (g *generator) generateNumericExpression(expr ast.Expression) (string, error) {
	if lit, ok := expr.(*ast.Literal); ok {
		if boolVal, ok := lit.Value.(bool); ok {
			if boolVal {
				return "1.0", nil
			}
			return "0.0", nil
		}
	}

	if g.boolConverter.IsAlreadyBoolean(expr) {
		boolCode, err := g.generateConditionExpression(expr)
		if err != nil {
			return "", err
		}
		boolCode = g.addBoolConversionIfNeeded(expr, boolCode)
		return fmt.Sprintf("func() float64 { if %s { return 1.0 } else { return 0.0 } }()", boolCode), nil
	}

	return g.generateConditionExpression(expr)
}

// generatePlotExpression generates inline code for plot() argument expressions
func (g *generator) generatePlotExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ConditionalExpression:
		condCode, err := g.generateConditionExpression(e.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(e.Test, condCode)

		consequentCode, err := g.generateNumericExpression(e.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(e.Alternate)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
			condCode, consequentCode, alternateCode), nil

	case *ast.Identifier:
		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, BarLoopScope); resolved {
			return code, nil
		}
		return e.Name + "Series.Get(0)", nil

	case *ast.MemberExpression:
		if code, resolved := g.builtinHandler.TryResolveMemberExpression(e, BarLoopScope); resolved {
			return code, nil
		}
		return g.extractSeriesExpression(e), nil

	case *ast.Literal:
		return g.generateNumericExpression(e)

	case *ast.BinaryExpression, *ast.LogicalExpression:
		return g.generateConditionExpression(expr)

	case *ast.CallExpression:
		return g.plotExprHandler.Generate(expr)

	case *ast.ObjectExpression:
		return "", nil

	default:
		return "", fmt.Errorf("unsupported plot expression type: %T", expr)
	}
}

func (g *generator) generateConditionExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ConditionalExpression:
		testCode, err := g.generateConditionExpression(e.Test)
		if err != nil {
			return "", err
		}
		testCode = g.addBoolConversionIfNeeded(e.Test, testCode)

		consequentCode, err := g.generateNumericExpression(e.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(e.Alternate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
			testCode, consequentCode, alternateCode), nil

	case *ast.UnaryExpression:
		operandCode, err := g.generateConditionExpression(e.Argument)
		if err != nil {
			return "", err
		}

		/* Ensure Series values converted to bool before unary operator */
		operandCode = g.ensureBooleanOperand(e.Argument, operandCode)

		op := e.Operator
		switch op {
		case "not":
			op = "!"
		}
		return fmt.Sprintf("%s%s", op, operandCode), nil

	case *ast.LogicalExpression:
		// Handle logical expressions (and, or)
		leftCode, err := g.generateConditionExpression(e.Left)
		if err != nil {
			return "", err
		}
		rightCode, err := g.generateConditionExpression(e.Right)
		if err != nil {
			return "", err
		}

		// Convert float64 Series values to bool for logical operations
		leftCode = g.ensureBooleanOperand(e.Left, leftCode)
		rightCode = g.ensureBooleanOperand(e.Right, rightCode)

		op := e.Operator
		switch op {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		}
		return fmt.Sprintf("(%s %s %s)", leftCode, op, rightCode), nil

	case *ast.BinaryExpression:
		// Special case: bar_index with modulo operator
		if e.Operator == "%" {
			leftIdent, leftIsBarIndex := e.Left.(*ast.Identifier)
			if leftIsBarIndex && leftIdent.Name == "bar_index" {
				right, err := g.generateConditionExpression(e.Right)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("float64(i %% %s)", right), nil
			}
		}

		left, err := g.generateConditionExpression(e.Left)
		if err != nil {
			return "", err
		}

		right, err := g.generateConditionExpression(e.Right)
		if err != nil {
			return "", err
		}

		// Map Pine operators to Go operators
		op := e.Operator
		switch op {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		}

		return fmt.Sprintf("(%s %s %s)", left, op, right), nil

	case *ast.MemberExpression:
		// Use extractSeriesExpression for proper offset handling
		return g.extractSeriesExpression(e), nil

	case *ast.Identifier:
		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, g.accessScope()); resolved {
			return code, nil
		}

		varName := e.Name

		/* Loop counter resolves to float64(counterVar) inside for-loop conditions */
		if g.loopContextStack != nil && g.loopContextStack.IsInLoop() && g.loopContextStack.IsLoopCounter(varName) {
			return fmt.Sprintf("float64(%s)", varName), nil
		}

		if constVal, isConstant := g.constants[varName]; isConstant {
			if constVal == "input.source" {
				return fmt.Sprintf("%sSeries.GetCurrent()", varName), nil
			}
			return varName, nil
		}

		// User-defined variable (ALL use Series storage)
		return fmt.Sprintf("%sSeries.GetCurrent()", varName), nil

	case *ast.Literal:
		switch v := e.Value.(type) {
		case float64:
			return g.literalFormatter.FormatFloat(v), nil
		case bool:
			return g.literalFormatter.FormatBool(v), nil
		case string:
			return g.literalFormatter.FormatString(v), nil
		default:
			formatted, err := g.literalFormatter.FormatGeneric(v)
			if err != nil {
				return "", fmt.Errorf("failed to format literal in expression: %w", err)
			}
			return formatted, nil
		}

	case *ast.CallExpression:
		if hoistedVarName := g.tempVarMgr.GetVarNameForCall(e); hoistedVarName != "" {
			code := hoistedVarName + "Series.GetCurrent()"
			if g.boolConverter.IsBooleanFunction(e) {
				code = "value.IsTrue(" + code + ")"
			}
			return code, nil
		}

		funcName := g.extractFunctionName(e.Callee)

		/* Delegate to inline condition handler registry */
		if g.inlineConditionRegistry.CanHandle(funcName) {
			return g.inlineConditionRegistry.GenerateInline(funcName, e, g)
		}

		/* Fallback to value handler for backward compatibility */
		if g.valueHandler.CanHandle(funcName) {
			return g.valueHandler.GenerateInlineCall(funcName, e.Arguments, g)
		}

		if varType, exists := g.variables[funcName]; exists && varType == "function" {
			return g.callRouter.RouteCall(g, e)
		}

		dispatcher := NewExpressionPositionDispatcher(g.callRouter)
		return dispatcher.Dispatch(g, e)

	case *ast.IfStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateIfExpressionAsIIFE(e)

	case *ast.ForStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateForExpressionAsIIFE(e)

	case *ast.ForInStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateForInExpressionAsIIFE(e)

	case *ast.WhileStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		return cfGenerator.GenerateWhileExpressionAsIIFE(e)

	default:
		return "", fmt.Errorf("unsupported condition expression: %T", expr)
	}
}

func (g *generator) generateVariableDeclaration(decl *ast.VariableDeclaration) (string, error) {
	code := ""
	for _, declarator := range decl.Declarations {
		id, ok := declarator.ID.(*ast.Identifier)
		if !ok {
			tupleCode, err := g.generateTupleDestructuringDeclaration(declarator)
			if err != nil {
				return "", err
			}
			if decl.Persistence != "" {
				arrayPattern, isArray := declarator.ID.(*ast.ArrayPattern)
				if isArray {
					indentedInit := ""
					for _, line := range strings.Split(tupleCode, "\n") {
						if line != "" {
							indentedInit += "\t" + line + "\n"
						}
					}
					var elemNames []string
					for _, elem := range arrayPattern.Elements {
						elemNames = append(elemNames, elem.Name)
					}
					code += g.persistenceEmitter.EmitTupleGuard(g.ind(), elemNames, indentedInit)
					return code, nil
				}
			}
			return tupleCode, err
		}
		varName := id.Name

		/* Persisted declarations must keep their zero-literal init */
		if decl.Kind == "let" && decl.Persistence == "" && g.reassignedVars[varName] {
			if lit, isLiteral := declarator.Init.(*ast.Literal); isLiteral {
				isZero := false
				switch v := lit.Value.(type) {
				case float64:
					isZero = v == 0
				case int:
					isZero = v == 0
				case int64:
					isZero = v == 0
				case string:
					isZero = v == "" || v == "0" || v == "0.0"
				}
				if isZero {
					continue
				}
			}
		}

		if _, ok := declarator.Init.(*ast.ArrowFunctionExpression); ok {
			// Already generated before bar loop - skip here
			continue
		}

		// Check if this is an input.* function call
		if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
			funcName := g.extractFunctionName(callExpr.Callee)

			if constVal, isConst := g.constants[varName]; isConst && constVal == "input.source" {
				funcName = "input.source"
			}

			// Handle input functions
			if IsInputConstantFuncName(funcName) {
				// Already handled in first pass - skip code generation here
				continue
			}

			if funcName == "input.source" {
				sourceSeries := "close"
				if len(callExpr.Arguments) > 0 {
					if id, ok := callExpr.Arguments[0].(*ast.Identifier); ok {
						sourceSeries = id.Name
					}
				}
				if seriesCode, resolved := g.builtinHandler.TryResolveIdentifier(&ast.Identifier{Name: sourceSeries}, BarLoopScope); resolved {
					code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, seriesCode)
				} else {
					code += g.ind() + fmt.Sprintf("// %s = input.source(defval=%s) - using source directly\n", varName, sourceSeries)
				}
				continue
			}
		}

		/* Persisted declarations and input.source bypass constant folding */
		if g.constantRegistry.IsConstant(varName) && decl.Persistence == "" {
			if constValue, exists := g.constants[varName]; exists && constValue == "input.source" {
				/* input.source needs initialization */
			} else {
				continue
			}
		}

		varType := g.inferVariableType(declarator.Init)

		isInLoop := g.loopContextStack != nil && g.loopContextStack.IsInLoop()
		if !isInLoop {
			if g.registryGuard != nil {
				if g.registryGuard.SafeRegister(varName, varType) {
					g.varInits[varName] = declarator.Init
				}
			} else {
				g.variables[varName] = varType
				g.varInits[varName] = declarator.Init
			}
		}

		if varType == "string" {
			if decl.Persistence != "" {
				g.indent++
				stringCode, err := g.generateStringVariableInit(varName, declarator.Init)
				g.indent--
				if err != nil {
					code += g.ind() + fmt.Sprintf("// %s = var string (generation failed: %v)\n", varName, err)
				} else {
					code += g.persistenceEmitter.EmitStringGuard(g.ind(), stringCode)
				}
			} else {
				stringCode, err := g.generateStringVariableInit(varName, declarator.Init)
				if err != nil {
					code += g.ind() + fmt.Sprintf("// %s = string variable (generation failed: %v)\n", varName, err)
				} else {
					code += stringCode
				}
			}
			continue
		}

		if declarator.Init != nil {
			if isInLoop {
				if decl.Persistence != "" {
					g.variables[varName] = varType
					g.indent++
					initCode, err := g.generateVariableInit(varName, declarator.Init)
					g.indent--
					if err != nil {
						return "", err
					}
					code += g.persistenceEmitter.EmitSeriesGuard(g.ind(), varName, initCode)
				} else if _, existsOuter := g.variables[varName]; existsOuter {
					seriesCode, err := g.generateLoopSeriesReassignment(varName, declarator.Init)
					if err != nil {
						return "", err
					}
					code += seriesCode
				} else {
					localCode, err := g.generateLoopLocalVariable(varName, declarator.Init)
					if err != nil {
						return "", err
					}
					code += localCode
				}
			} else if g.inArrowFunctionBody {
				seriesCode, err := g.generateArrowFunctionSeriesInit(varName, declarator.Init)
				if err != nil {
					return "", err
				}
				if decl.Persistence != "" {
					code += g.persistenceEmitter.EmitSeriesGuard(g.ind(), varName, "\t"+seriesCode)
				} else {
					code += seriesCode
				}
			} else {
				// Series context: Use ForwardSeriesBuffer paradigm
				if decl.Persistence != "" {
					g.indent++
					initCode, err := g.generateVariableInit(varName, declarator.Init)
					g.indent--
					if err != nil {
						return "", err
					}
					code += g.persistenceEmitter.EmitSeriesGuard(g.ind(), varName, initCode)
				} else {
					initCode, err := g.generateVariableInit(varName, declarator.Init)
					if err != nil {
						return "", err
					}
					code += initCode
				}
			}
		}
	}
	return code, nil
}

/* generateLoopLocalVariable generates local Go variable assignment inside for loops */
func (g *generator) generateLoopLocalVariable(varName string, initExpr ast.Expression) (string, error) {
	exprCode, err := g.generateArrowFunctionExpression(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate loop local variable %s: %w", varName, err)
	}
	return g.ind() + fmt.Sprintf("%s := %s\n", varName, exprCode), nil
}

/* generateLoopSeriesReassignment generates Series.Set() for reassigning outer Series variables in loops */
func (g *generator) generateLoopSeriesReassignment(varName string, initExpr ast.Expression) (string, error) {
	exprCode, err := g.generateArrowFunctionExpression(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate loop series reassignment %s: %w", varName, err)
	}
	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, exprCode), nil
}

/*
generateArrowFunctionSeriesInit generates Series.Set() for arrow function variables.

Universal ForwardSeriesBuffer paradigm: ALL arrow function variables use Series storage.
This replaces the old scalar assignment approach.
*/
func (g *generator) generateArrowFunctionSeriesInit(varName string, initExpr ast.Expression) (string, error) {
	exprCode, err := g.generateArrowFunctionExpression(initExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate expression for %s: %w", varName, err)
	}

	return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, exprCode), nil
}

func (g *generator) generateArrowFunctionVariableInit(varName string, initExpr ast.Expression) (*ArrowVarInitResult, error) {
	switch expr := initExpr.(type) {
	case *ast.CallExpression:
		funcName := extractCallFunctionName(expr)
		if funcName == "fixnan" || funcName == "ta.fixnan" {
			return g.generateArrowFunctionFixnanInit(varName, expr)
		}

		exprCode, err := g.generateCallExpression(expr)
		if err != nil {
			return nil, err
		}
		assignment := g.ind() + fmt.Sprintf("%s := %s\n", varName, exprCode)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.BinaryExpression:
		exprCode, err := g.generateBinaryExpression(expr)
		if err != nil {
			return nil, err
		}
		assignment := g.ind() + fmt.Sprintf("%s := %s\n", varName, exprCode)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.Identifier:
		assignment := g.ind() + fmt.Sprintf("%s := %s\n", varName, expr.Name)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.Literal:
		assignment := g.ind() + fmt.Sprintf("%s := %v\n", varName, expr.Value)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.MemberExpression:
		exprCode, err := g.generateMemberExpression(expr)
		if err != nil {
			return nil, err
		}
		assignment := g.ind() + fmt.Sprintf("%s := %s\n", varName, exprCode)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.ConditionalExpression:
		condCode, err := g.generateConditionExpression(expr.Test)
		if err != nil {
			return nil, err
		}
		condCode = g.addBoolConversionIfNeeded(expr.Test, condCode)

		consequentCode, err := g.generateNumericExpression(expr.Consequent)
		if err != nil {
			return nil, err
		}
		alternateCode, err := g.generateNumericExpression(expr.Alternate)
		if err != nil {
			return nil, err
		}
		assignment := g.ind() + fmt.Sprintf("%s := func() float64 { if %s { return %s } else { return %s } }()\n",
			varName, condCode, consequentCode, alternateCode)
		return NewArrowVarInitResult("", assignment), nil

	case *ast.UnaryExpression:
		operandCode, err := g.generateArrowFunctionExpression(expr.Argument)
		if err != nil {
			return nil, err
		}
		op := expr.Operator
		if op == "not" {
			op = "!"
		}
		assignment := g.ind() + fmt.Sprintf("%s := %s%s\n", varName, op, operandCode)
		return NewArrowVarInitResult("", assignment), nil

	default:
		return nil, fmt.Errorf("unsupported arrow function variable init expression: %T", initExpr)
	}
}

func (g *generator) generateArrowFunctionFixnanInit(varName string, call *ast.CallExpression) (*ArrowVarInitResult, error) {
	if len(call.Arguments) < 1 {
		return nil, fmt.Errorf("fixnan() requires 1 argument")
	}

	sourceExpr := call.Arguments[0]

	accessor, err := g.createAccessorForFixnan(sourceExpr)
	if err != nil {
		return nil, fmt.Errorf("fixnan: failed to create accessor: %w", err)
	}

	extractor := NewPreambleExtractor()
	preamble := extractor.ExtractFromAccessor(accessor)

	targetSeriesVar := varName + "Series"
	generator := &FixnanIIFEGenerator{}
	iifeCode := generator.GenerateWithSelfReference(accessor, targetSeriesVar)

	assignment := g.ind() + fmt.Sprintf("%s := %s\n", varName, iifeCode)
	return NewArrowVarInitResult(preamble, assignment), nil
}

func (g *generator) createAccessorForFixnan(expr ast.Expression) (AccessGenerator, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		if varType, exists := g.variables[e.Name]; exists && varType == "float" {
			return NewArrowFunctionParameterAccessor(e.Name), nil
		}

		classifier := NewSeriesSourceClassifier()
		sourceInfo := classifier.ClassifyAST(e)
		return CreateAccessGenerator(sourceInfo), nil

	case *ast.CallExpression:
		funcName := extractCallFunctionName(e)

		tempVarName := strings.ReplaceAll(funcName, ".", "_") + "_temp"
		result, err := g.generateArrowFunctionVariableInit(tempVarName, e)
		if err != nil {
			return nil, fmt.Errorf("failed to generate temp var for fixnan source: %w", err)
		}

		// Extract expression code from assignment (format: "tempVar := expression\n")
		exprCode, err := g.generateCallExpression(e)
		if err != nil {
			return nil, fmt.Errorf("failed to generate call expression for accessor: %w", err)
		}

		return &FixnanCallExpressionAccessor{
			tempVarName: tempVarName,
			tempVarCode: result.CombinedCode(),
			exprCode:    exprCode,
		}, nil

	case *ast.BinaryExpression:
		tempVarName := "fixnan_source_temp"
		binaryCode, err := g.generateBinaryExpression(e)
		if err != nil {
			return nil, fmt.Errorf("failed to generate binary expression: %w", err)
		}

		return &FixnanCallExpressionAccessor{
			tempVarName: tempVarName,
			tempVarCode: g.ind() + fmt.Sprintf("%s := %s\n", tempVarName, binaryCode),
			exprCode:    binaryCode,
		}, nil

	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if obj.Name == "ctx" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					fieldName := capitalize(prop.Name)
					return NewOHLCVFieldAccessGenerator(fieldName), nil
				}
			}
		}
		return nil, fmt.Errorf("unsupported member expression in fixnan")

	default:
		return nil, fmt.Errorf("unsupported source expression type for fixnan: %T", expr)
	}
}

// inferVariableType delegates to TypeInferenceEngine
func (g *generator) inferVariableType(expr ast.Expression) string {
	return g.typeSystem.InferType(expr)
}

func (g *generator) generateStringVariableInit(varName string, initExpr ast.Expression) (string, error) {
	switch expr := initExpr.(type) {
	case *ast.Literal:
		if s, ok := expr.Value.(string); ok {
			return g.ind() + fmt.Sprintf("%s = %q\n", varName, s), nil
		}
		return "", fmt.Errorf("unsupported literal type for string variable: %T", expr.Value)

	case *ast.Identifier:
		if hex, found := g.builtinHandler.ResolveColorHex(expr.Name); found {
			return g.ind() + fmt.Sprintf("%s = %q\n", varName, hex), nil
		}
		return "", fmt.Errorf("unsupported string identifier: %s", expr.Name)

	case *ast.ConditionalExpression:
		condCode, err := g.generateConditionExpression(expr.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(expr.Test, condCode)

		consequentCode, err := g.generateStringExpression(expr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateStringExpression(expr.Alternate)
		if err != nil {
			return "", err
		}
		return g.ind() + fmt.Sprintf("%s = func() string { if %s { return %s } else { return %s } }()\n",
			varName, condCode, consequentCode, alternateCode), nil

	case *ast.MemberExpression:
		if hex, found := g.builtinHandler.ResolveMemberExpressionColorHex(expr); found {
			return g.ind() + fmt.Sprintf("%s = %q\n", varName, hex), nil
		}

		if obj, ok := expr.Object.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				if prop, ok := expr.Property.(*ast.Identifier); ok {
					if prop.Name == "long" || prop.Name == "short" {
						return g.ind() + fmt.Sprintf("%s = strategy.%s\n", varName, capitalize(prop.Name)), nil
					}
				}
			}
		}
		return "", fmt.Errorf("unsupported string member expression: %v", expr)

	case *ast.CallExpression:
		funcName := g.extractFunctionName(expr.Callee)
		if g.colorHandler.CanHandle(funcName) {
			colorCode, err := g.colorHandler.GenerateColorCall(funcName, expr.Arguments, g)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%s = %s\n", varName, colorCode), nil
		}
		return "", fmt.Errorf("unsupported call expression for string variable: %s", funcName)

	default:
		return "", fmt.Errorf("unsupported string variable init: %T", initExpr)
	}
}

func (g *generator) generateStringExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Literal:
		if s, ok := e.Value.(string); ok {
			return fmt.Sprintf("%q", s), nil
		}
		return "", fmt.Errorf("unsupported literal type for string expression: %T", e.Value)

	case *ast.Identifier:
		if hex, found := g.builtinHandler.ResolveColorHex(e.Name); found {
			return fmt.Sprintf("%q", hex), nil
		}
		return "", fmt.Errorf("unsupported string identifier: %s", e.Name)

	case *ast.ConditionalExpression:
		condCode, err := g.generateConditionExpression(e.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(e.Test, condCode)

		consequentCode, err := g.generateStringExpression(e.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateStringExpression(e.Alternate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() string { if %s { return %s } else { return %s } }()",
			condCode, consequentCode, alternateCode), nil

	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					if prop.Name == "long" {
						return "strategy.Long", nil
					}
					if prop.Name == "short" {
						return "strategy.Short", nil
					}
				}
			}
			if obj.Name == "color" {
				resolver := NewConstantResolver()
				if colorValue, ok := resolver.ResolveToString(e); ok {
					return fmt.Sprintf("%q", colorValue), nil
				}
			}
		}
		return "", fmt.Errorf("unsupported string member expression: %v", e)

	case *ast.CallExpression:
		funcName := g.extractFunctionName(e.Callee)
		if g.colorHandler.CanHandle(funcName) {
			return g.colorHandler.GenerateColorCall(funcName, e.Arguments, g)
		}
		return "", fmt.Errorf("unsupported call expression for string expression: %s", funcName)

	default:
		return "", fmt.Errorf("unsupported string expression: %T", expr)
	}
}

func (g *generator) generateVariableInit(varName string, initExpr ast.Expression) (string, error) {
	nestedCalls := g.exprAnalyzer.FindNestedCalls(initExpr)

	tempVarCode := ""
	if len(nestedCalls) > 0 {
		deduplicator := NewTempVarInlineDeduplicator(g.tempVarMgr)

		for i := len(nestedCalls) - 1; i >= 0; i-- {
			callInfo := nestedCalls[i]

			if callInfo.Call == initExpr {
				continue
			}

			if g.runtimeOnlyFilter.IsRuntimeOnly(callInfo.FuncName) {
				continue
			}

			isTAFunction := g.taRegistry.IsSupported(callInfo.FuncName)
			containsNestedTA := false
			if !isTAFunction {
				mathNestedCalls := g.exprAnalyzer.FindNestedCalls(callInfo.Call)
				for _, mathNested := range mathNestedCalls {
					if mathNested.Call != callInfo.Call && g.taRegistry.IsSupported(mathNested.FuncName) {
						containsNestedTA = true
						break
					}
				}
			}

			if !isTAFunction && !containsNestedTA {
				continue
			}

			tempVarName := g.tempVarMgr.GetOrCreate(callInfo)

			if !deduplicator.ShouldEmitCalculation(callInfo) {
				continue
			}

			tempCode, err := g.generateVariableFromCall(tempVarName, callInfo.Call)
			if err != nil {
				return "", fmt.Errorf("failed to generate temp var %s: %w", tempVarName, err)
			}
			tempVarCode += tempCode
		}
	}

	switch expr := initExpr.(type) {
	case *ast.CallExpression:
		mainCode, err := g.generateVariableFromCall(varName, expr)
		return tempVarCode + mainCode, err
	case *ast.ConditionalExpression:
		condCode, err := g.generateConditionExpression(expr.Test)
		if err != nil {
			return "", err
		}
		condCode = g.addBoolConversionIfNeeded(expr.Test, condCode)

		consequentCode, err := g.generateNumericExpression(expr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(expr.Alternate)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return %s } else { return %s } }())\n",
			varName, condCode, consequentCode, alternateCode), nil
	case *ast.UnaryExpression:
		if expr.Operator == "not" || expr.Operator == "!" {
			operandCode, err := g.generateConditionExpression(expr.Argument)
			if err != nil {
				return "", err
			}
			boolToFloatExpr := fmt.Sprintf("func() float64 { if !(%s) { return 1.0 } else { return 0.0 } }()", operandCode)
			return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, boolToFloatExpr), nil
		} else {
			/* expression-level generator avoids statement decorations in init */
			operandCode, err := g.generateConditionExpression(expr.Argument)
			if err != nil {
				return "", err
			}
			return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s(%s))\n", varName, expr.Operator, operandCode), nil
		}
	case *ast.Literal:
		switch v := expr.Value.(type) {
		case float64:
			formatted := g.literalFormatter.FormatFloat(v)
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, formatted), nil
		case int:
			formatted := g.literalFormatter.FormatFloat(float64(v))
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, formatted), nil
		case bool:
			val := 0.0
			if v {
				val = 1.0
			}
			formatted := g.literalFormatter.FormatFloat(val)
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, formatted), nil
		case string:
			return g.ind() + fmt.Sprintf("// ERROR: string literal %q cannot be used in series context\n", v), nil
		default:
			return g.ind() + fmt.Sprintf("// ERROR: unsupported literal type\n"), nil
		}
	case *ast.Identifier:
		refName := expr.Name

		if code, resolved := g.builtinHandler.TryResolveIdentifier(expr, g.accessScope()); resolved {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, code), nil
		}

		if _, isConstant := g.constants[refName]; isConstant {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, refName), nil
		}

		accessCode := fmt.Sprintf("%sSeries.GetCurrent()", refName)
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, accessCode), nil
	case *ast.MemberExpression:
		memberCode := g.extractSeriesExpression(expr)

		/* strategy.long/short are string constants — map to numeric for Series storage */
		if obj, ok := expr.Object.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				if prop, ok := expr.Property.(*ast.Identifier); ok {
					if prop.Name == "long" {
						return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(1.0) // strategy.long\n", varName), nil
					} else if prop.Name == "short" {
						return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(-1.0) // strategy.short\n", varName), nil
					}
				}
			}
		}

		if goType, resolved := g.builtinHandler.ResolveMemberExpressionGoType(expr); resolved && goType != GoFloat64 {
			coerced := g.seriesInitCoercer.Coerce(memberCode, goType)
			return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, coerced), nil
		}

		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, memberCode), nil
	case *ast.BinaryExpression:
		/* In security context, need to generate temp series for operands */
		if g.inSecurityContext {
			return g.generateBinaryExpressionInSecurityContext(varName, expr)
		}

		/* Format binary expression with operator precedence awareness */
		formatter := NewBinaryExpressionFormatterWithExtractor(g.extractSeriesExpression)
		binaryCode := formatter.formatWithExtractor(expr)

		varType := g.inferVariableType(expr)
		if varType == "bool" {
			/* Convert bool to float64 for Series storage */
			return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return 1.0 } else { return 0.0 } }())\n", varName, binaryCode), nil
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, binaryCode), nil
	case *ast.LogicalExpression:
		logicalCode, err := g.generateConditionExpression(expr)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return 1.0 } else { return 0.0 } }())\n", varName, logicalCode), nil
	case *ast.IfStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		iifeCode, err := cfGenerator.GenerateIfExpressionAsIIFE(expr)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
	case *ast.ForStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		iifeCode, err := cfGenerator.GenerateForExpressionAsIIFE(expr)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
	case *ast.ForInStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		iifeCode, err := cfGenerator.GenerateForInExpressionAsIIFE(expr)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
	case *ast.WhileStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g)
		iifeCode, err := cfGenerator.GenerateWhileExpressionAsIIFE(expr)
		if err != nil {
			return "", err
		}
		return tempVarCode + g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, iifeCode), nil
	default:
		return "", fmt.Errorf("unsupported init expression: %T", initExpr)
	}
}

func (g *generator) generateVariableFromCall(varName string, call *ast.CallExpression) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	// Check if this is a user-defined function
	if varType, exists := g.variables[funcName]; exists && varType == "function" {
		ctxVarName := g.arrowContextLifecycle.AllocateContextVariable(funcName)

		code := ""

		if !g.arrowContextLifecycle.IsHoisted(ctxVarName) {
			code = g.ind() + fmt.Sprintf("%s := context.NewArrowContext(ctx)\n", ctxVarName)
		}

		callCode, err := g.generateUserDefinedFunctionCallWithContext(call, ctxVarName)
		if err != nil {
			return "", err
		}
		code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, callCode)
		return code, nil
	}

	// Try TA function registry first
	if g.taRegistry.IsSupported(funcName) {
		return g.taRegistry.GenerateInlineTA(g, varName, funcName, call)
	}

	if sharedTASignatures.Contains(funcName) {
		log.Printf("WARNING: TA function %s has no handler — producing NaN stub", funcName)
		return g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName), nil
	}

	// Handle math functions that need Series storage (have TA dependencies)
	mathHandler := NewMathFunctionHandler()
	if mathHandler.CanHandle(funcName) {
		return mathHandler.GenerateCode(g, varName, call)
	}

	if IsSecurityFunction(funcName) {
		if len(call.Arguments) < 3 {
			return g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN()) // security() missing arguments\n", varName), nil
		}

		argExtractor := NewSecurityArgumentExtractor(g)

		symbolResult, err := argExtractor.ExtractSymbol(call.Arguments[0])
		if err != nil {
			return "", fmt.Errorf("failed to extract security symbol: %w", err)
		}

		timeframeResult, err := argExtractor.ExtractTimeframe(call.Arguments[1])
		if err != nil {
			return "", fmt.Errorf("failed to extract security timeframe: %w", err)
		}

		g.hasSecurityCalls = true

		keyBuilder := NewSecurityCacheKeyBuilder()
		keyComponents := keyBuilder.Build(symbolResult, timeframeResult)

		code := g.ind() + fmt.Sprintf("/* security(%s, %s, ...) */\n", symbolResult.Code, timeframeResult.Code)
		code += g.ind() + "{\n"
		g.indent++

		if keyComponents.FormatArgs == "" {
			code += g.ind() + fmt.Sprintf("secKey := %q\n", keyComponents.KeyPattern)
		} else {
			code += g.ind() + fmt.Sprintf("secKey := fmt.Sprintf(%q, %s)\n", keyComponents.KeyPattern, keyComponents.FormatArgs)
		}
		code += g.ind() + "secCtx, secFound := securityContexts[secKey]\n"
		code += g.ind() + "if !secFound {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++

		lookahead := resolveSecurityLookahead(call, g.pineVersion)

		code += g.ind() + "securityBarMapper, mapperFound := securityBarMappers[secKey]\n"
		code += g.ind() + "if !mapperFound {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++

		/* Calculate lookahead for bar mapper */
		code += g.ind() + fmt.Sprintf("secLookahead := %v\n", lookahead)
		code += g.ind() + fmt.Sprintf("if %s == ctx.Timeframe {\n", timeframeResult.Code)
		g.indent++
		code += g.ind() + "secLookahead = true\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "\n"

		/* Context hierarchy setup: link security context → main context */
		code += g.ind() + "if secCtx.GetParent() == nil {\n"
		g.indent++
		code += g.ind() + "barAligner := request.NewSecurityBarMapperAligner(securityBarMapper, secLookahead)\n"
		code += g.ind() + "secCtx.SetParent(ctx, barAligner)\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "\n"

		code += g.ind() + "secBarIdx := securityBarMapper.FindDailyBarIndex(ctx.BarIndex, secLookahead)\n"
		code += g.ind() + "if secBarIdx < 0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++

		exprArg := call.Arguments[2]

		secExprHandler := NewSecurityExpressionHandler(SecurityExpressionConfig{
			IndentFunc:           g.ind,
			IncrementIndent:      func() { g.indent++ },
			DecrementIndent:      func() { g.indent-- },
			SerializeExpr:        g.serializeExpressionForRuntime,
			MarkSecurityExprEval: func() { g.hasSecurityExprEvals = true },
			SymbolTable:          g.symbolTable,
			Generator:            g,
		})
		evalCode, err := secExprHandler.GenerateEvaluationCode(varName, exprArg, "secBarIdx")
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

	switch funcName {
	case "plot":
		opts := ParsePlotOptions(call)

		var plotExpr string
		if len(call.Arguments) > 0 {
			exprCode, err := g.generatePlotExpression(call.Arguments[0])
			if err != nil {
				return "", err
			}
			plotExpr = exprCode
		}

		code := ""
		if plotExpr != "" && opts.ColorExpr != nil {
			if condExpr, ok := opts.ColorExpr.(*ast.ConditionalExpression); ok {
				testCode, err := g.generateConditionExpression(condExpr.Test)
				if err != nil {
					return "", err
				}

				if _, isCall := condExpr.Test.(*ast.CallExpression); isCall {
					testCode = fmt.Sprintf("(%s) != 0", testCode)
				} else {
					testCode = g.addBoolConversionIfNeeded(condExpr.Test, testCode)
				}

				alternateIsNa := false
				if ident, ok := condExpr.Alternate.(*ast.Identifier); ok && ident.Name == "na" {
					alternateIsNa = true
				}

				if alternateIsNa {
					code += g.ind() + fmt.Sprintf("if !(%s) {\n", testCode)
					g.indent++
					colorValue := g.colorHandler.ResolveColorExpression(condExpr.Consequent, g)
					optionsWithColor := g.buildPlotOptionsWithColor(opts, colorValue)
					code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, optionsWithColor)
					g.indent--
					code += g.ind() + "} else {\n"
					g.indent++
					code += g.ind() + "/* Add plot point with null color to mark gap */\n"
					gapOptions := g.buildPlotOptionsWithNullColor(opts)
					code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, gapOptions)
					g.indent--
					code += g.ind() + "}\n"
				} else {
					code += g.ind() + fmt.Sprintf("if %s {\n", testCode)
					g.indent++
					code += g.ind() + "/* Consequent is na - add plot point with null color to mark gap */\n"
					gapOptions := g.buildPlotOptionsWithNullColor(opts)
					code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, gapOptions)
					g.indent--
					code += g.ind() + "} else {\n"
					g.indent++
					colorValue := g.colorHandler.ResolveColorExpression(condExpr.Alternate, g)
					optionsWithColor := g.buildPlotOptionsWithColor(opts, colorValue)
					code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, optionsWithColor)
					g.indent--
					code += g.ind() + "}\n"
				}
			} else {
				options := g.buildPlotOptions(opts)
				code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, options)
			}
		} else if plotExpr != "" {
			options := g.buildPlotOptions(opts)
			code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, %s)\n", opts.Title, plotExpr, options)
		}
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		return code, nil

	case "time":
		/* time(timeframe, session) - session filtering for intraday strategies
		 * Returns bar timestamp if within session, NaN otherwise
		 * Usage: entry_time = time(timeframe.period, "0950-1345")
		 * Check: is_entry_time = na(entry_time) ? false : true
		 */
		handler := NewTimeHandler(g.ind())
		return handler.HandleVariableInit(varName, call), nil

	case "nz":
		/* nz(x, replacement) - replaces NaN with replacement value (default 0) */
		nzCode, err := g.valueHandler.generateNz(call.Arguments, g)
		if err != nil {
			return "", err
		}
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, nzCode), nil

	default:
		if g.colorHandler.CanHandle(funcName) {
			colorCode, err := g.colorHandler.GenerateColorCall(funcName, call.Arguments, g)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, colorCode), nil
		}
		if g.mathHandler.CanHandle(funcName) {
			mathCode, err := g.mathHandler.GenerateMathCall(funcName, call.Arguments, g)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, mathCode), nil
		}
		if g.calendarHandler.CanHandle(funcName) {
			calCode, err := g.calendarHandler.GenerateCalendarCall(funcName, call.Arguments, g)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, calCode), nil
		}
		/* timeframe.change/in_seconds/from_seconds in variable init */
		if g.timeframeFuncHandler.CanHandle(funcName) {
			tfCode, err := g.timeframeFuncHandler.GenerateCode(g, call)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, tfCode), nil
		}

		routedCode, err := g.callRouter.RouteCall(g, call)
		if err != nil {
			return "", fmt.Errorf("failed to route %s: %w", funcName, err)
		}
		if routedCode != "" && !strings.HasPrefix(strings.TrimSpace(routedCode), "//") {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, routedCode), nil
		}

		return g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN()) // TODO: implement %s()\n", varName, funcName), nil
	}
}

/* generateInlineATR generates inline ATR calculation for security() context
 * ATR = RMA(TR, period) where TR = max(H-L, |H-prevC|, |L-prevC|)
 */
func (g *generator) generateInlineATR(varName string, period int) (string, error) {
	var code string

	code += g.ind() + fmt.Sprintf("/* Inline ATR(%d) in security context */\n", period)
	code += g.ind() + "if ctx.BarIndex < 1 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++

	/* Calculate TR for current bar */
	code += g.ind() + "hl := highSeries.GetCurrent() - lowSeries.GetCurrent()\n"
	code += g.ind() + "hc := math.Abs(highSeries.GetCurrent() - closeSeries.Get(1))\n"
	code += g.ind() + "lc := math.Abs(lowSeries.GetCurrent() - closeSeries.Get(1))\n"
	code += g.ind() + "tr := math.Max(hl, math.Max(hc, lc))\n"

	/* RMA smoothing of TR */
	code += g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", period)
	g.indent++
	/* Warmup: use SMA for first period bars - loop uses absolute indices */
	code += g.ind() + "sum := 0.0\n"
	code += g.ind() + "for j := 0; j <= ctx.BarIndex; j++ {\n"
	g.indent++
	code += g.ind() + "if j == 0 {\n"
	g.indent++
	code += g.ind() + "sum += ctx.Data[j].High - ctx.Data[j].Low\n"
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + "hl_j := ctx.Data[j].High - ctx.Data[j].Low\n"
	code += g.ind() + "hc_j := math.Abs(ctx.Data[j].High - ctx.Data[j-1].Close)\n"
	code += g.ind() + "lc_j := math.Abs(ctx.Data[j].Low - ctx.Data[j-1].Close)\n"
	code += g.ind() + "sum += math.Max(hl_j, math.Max(hc_j, lc_j))\n"
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("if ctx.BarIndex == %d-1 {\n", period)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(sum / %d.0)\n", varName, period)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	/* RMA: prevATR + (TR - prevATR) / period */
	code += g.ind() + fmt.Sprintf("alpha := 1.0 / %d.0\n", period)
	code += g.ind() + fmt.Sprintf("prevATR := %sSeries.Get(1)\n", varName)
	code += g.ind() + "atr := prevATR + alpha*(tr - prevATR)\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(atr)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

/* generateBinaryExpressionInSecurityContext handles BinaryExpression with temp series
 * Creates temp series for left/right operands, then combines with operator
 */
func (g *generator) generateBinaryExpressionInSecurityContext(varName string, expr *ast.BinaryExpression) (string, error) {
	var code string

	/* Generate temp series for left operand */
	leftVar := fmt.Sprintf("%s_left", varName)
	code += g.ind() + fmt.Sprintf("%sSeries := series.NewSeries(len(ctx.Data))\n", leftVar)

	leftInit, err := g.generateVariableInit(leftVar, expr.Left)
	if err != nil {
		return "", fmt.Errorf("failed to generate left operand: %w", err)
	}
	code += leftInit

	/* Generate temp series for right operand */
	rightVar := fmt.Sprintf("%s_right", varName)
	code += g.ind() + fmt.Sprintf("%sSeries := series.NewSeries(len(ctx.Data))\n", rightVar)

	rightInit, err := g.generateVariableInit(rightVar, expr.Right)
	if err != nil {
		return "", fmt.Errorf("failed to generate right operand: %w", err)
	}
	code += rightInit

	/* Combine operands with operator */
	combineExpr := fmt.Sprintf("%sSeries.GetCurrent() %s %sSeries.GetCurrent()",
		leftVar, expr.Operator, rightVar)

	/* Check if result is boolean (comparison operators) */
	varType := g.inferVariableType(expr)
	if varType == "bool" {
		code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return 1.0 } else { return 0.0 } }())\n",
			varName, combineExpr)
	} else {
		code += g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, combineExpr)
	}

	return code, nil
}

func (g *generator) extractFunctionName(callee ast.Expression) string {
	switch c := callee.(type) {
	case *ast.Identifier:
		return c.Name
	case *ast.MemberExpression:
		obj := ""
		if id, ok := c.Object.(*ast.Identifier); ok {
			obj = id.Name
		}
		prop := ""
		if id, ok := c.Property.(*ast.Identifier); ok {
			prop = id.Name
		}
		return obj + "." + prop
	default:
		return "unknown"
	}
}

func (g *generator) extractArgIdentifier(expr ast.Expression) string {
	// Handle MemberExpression like close[0]
	if mem, ok := expr.(*ast.MemberExpression); ok {
		if id, ok := mem.Object.(*ast.Identifier); ok {
			// Map Pine builtins to OHLCV fields
			switch id.Name {
			case "close":
				return "Close"
			case "open":
				return "Open"
			case "high":
				return "High"
			case "low":
				return "Low"
			case "volume":
				return "Volume"
			default:
				return id.Name
			}
		}
	}
	// Handle direct Identifier (legacy support)
	if id, ok := expr.(*ast.Identifier); ok {
		// Map Pine builtins to OHLCV fields
		switch id.Name {
		case "close":
			return "Close"
		case "open":
			return "Open"
		case "high":
			return "High"
		case "low":
			return "Low"
		case "volume":
			return "Volume"
		default:
			return id.Name
		}
	}
	return "Close" // Default
}

func (g *generator) extractArgLiteral(expr ast.Expression) int {
	if lit, ok := expr.(*ast.Literal); ok {
		if val, ok := lit.Value.(float64); ok {
			return int(val)
		}
	}
	return 0
}

/* extractStrategyName extracts title from strategy/indicator/study arguments */
func (g *generator) extractStrategyName(args []ast.Expression) string {
	if len(args) == 0 {
		return ""
	}

	if lit, ok := args[0].(*ast.Literal); ok {
		if name, ok := lit.Value.(string); ok {
			return name
		}
	}

	for _, arg := range args {
		if obj, ok := arg.(*ast.ObjectExpression); ok {
			parser := NewPropertyParser()
			if title, ok := parser.ParseString(obj, "title"); ok {
				return title
			}
		}
	}

	return ""
}

func (g *generator) generatePattern(pattern ast.Pattern) string {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return p.Name
	case *ast.ArrayPattern:
		names := make([]string, len(p.Elements))
		for i, elem := range p.Elements {
			names[i] = elem.Name
		}
		return strings.Join(names, ", ")
	default:
		return "unknown"
	}
}

func (g *generator) generateTupleDestructuringDeclaration(declarator ast.VariableDeclarator) (string, error) {
	arrayPattern, ok := declarator.ID.(*ast.ArrayPattern)
	if !ok {
		return "", fmt.Errorf("expected ArrayPattern for tuple destructuring, got %T", declarator.ID)
	}

	if len(arrayPattern.Elements) == 0 {
		return "", fmt.Errorf("empty tuple pattern")
	}

	varNames := make([]string, len(arrayPattern.Elements))
	for i, elem := range arrayPattern.Elements {
		varNames[i] = elem.Name
		g.variables[elem.Name] = "float"
	}

	callExpr, ok := declarator.Init.(*ast.CallExpression)
	if !ok {
		return "", fmt.Errorf("tuple destructuring init must be CallExpression, got %T", declarator.Init)
	}

	funcName := extractCallFunctionName(callExpr)
	detector := NewUserDefinedFunctionDetector(g.variables)

	if detector.IsUserDefinedFunction(funcName) {
		return g.generateUserDefinedFunctionTupleCall(varNames, funcName, callExpr)
	}

	/* Delegate tuple-returning TA functions to specialized handlers */
	if g.tupleIndicatorHandler.CanHandle(funcName) {
		return g.tupleIndicatorHandler.GenerateTupleCode(g, varNames, callExpr)
	}

	/* Route security()/request.security() tuple calls to specialized handler */
	if IsSecurityFunction(funcName) {
		return g.generateTupleSecurityDeclaration(varNames, callExpr)
	}

	initCode, err := g.generateCallExpression(callExpr)
	if err != nil {
		return "", err
	}

	/* Function implemented: use normal assignment */
	if !isTODOComment(initCode) {
		return g.ind() + fmt.Sprintf("%s := %s\n", strings.Join(varNames, ", "), initCode), nil
	}

	/* Function unimplemented: AST-driven graceful degradation.
	 * len(varNames) from ArrayPattern tells us return count - no hardcoded registry.
	 */
	code := g.ind() + fmt.Sprintf("/* %s() - TODO: implement */\n", funcName)
	for _, varName := range varNames {
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	}
	return code, nil
}

func isTODOComment(code string) bool {
	return strings.Contains(code, "// ") && strings.Contains(code, "TODO: implement")
}

func (g *generator) generateUserDefinedFunctionTupleCall(varNames []string, funcName string, callExpr *ast.CallExpression) (string, error) {
	code := ""

	ctxVarName := g.arrowContextLifecycle.AllocateContextVariable(funcName)

	if !g.arrowContextLifecycle.IsHoisted(ctxVarName) {
		code += g.ind() + fmt.Sprintf("%s := context.NewArrowContext(ctx)\n", ctxVarName)
	}

	args := []string{ctxVarName}
	for idx, arg := range callExpr.Arguments {
		argGen := NewArgumentExpressionGenerator(g, funcName, idx)
		argCode, err := argGen.Generate(arg)
		if err != nil {
			return "", fmt.Errorf("failed to generate argument %d: %w", idx, err)
		}
		args = append(args, argCode)
	}

	callCode := fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
	code += g.ind() + fmt.Sprintf("%s := %s\n", strings.Join(varNames, ", "), callCode)

	code += g.returnValueStorage.GenerateStorageStatements(varNames)

	return code, nil
}

func (g *generator) generateUserDefinedFunctionCallWithContext(callExpr *ast.CallExpression, ctxVarName string) (string, error) {
	funcName := extractCallFunctionName(callExpr)

	args := []string{ctxVarName}
	for idx, arg := range callExpr.Arguments {
		argGen := NewArgumentExpressionGenerator(g, funcName, idx)
		argCode, err := argGen.Generate(arg)
		if err != nil {
			return "", fmt.Errorf("failed to generate argument %d: %w", idx, err)
		}
		args = append(args, argCode)
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", ")), nil
}

func (g *generator) extractStringLiteral(expr ast.Expression) string {
	if lit, ok := expr.(*ast.Literal); ok {
		if val, ok := lit.Value.(string); ok {
			return val
		}
	}
	return ""
}

func (g *generator) extractFloatLiteral(expr ast.Expression) float64 {
	if lit, ok := expr.(*ast.Literal); ok {
		if val, ok := lit.Value.(float64); ok {
			return val
		}
	}
	return 0.0
}

func (g *generator) extractDirectionConstant(expr ast.Expression) string {
	if g.directionExtractor == nil {
		g.directionExtractor = NewDefaultDirectionExtractor()
	}
	return g.directionExtractor.Extract(expr)
}

func (g *generator) extractMemberName(expr *ast.MemberExpression) string {
	obj := ""
	if id, ok := expr.Object.(*ast.Identifier); ok {
		obj = id.Name
	}
	prop := ""
	if id, ok := expr.Property.(*ast.Identifier); ok {
		prop = id.Name
	}

	// Map Pine constants to Go runtime constants
	if obj == "strategy" {
		switch prop {
		case "long":
			return "strategy.Long"
		case "short":
			return "strategy.Short"
		}
	}

	return obj + "." + prop
}

func (g *generator) extractSeriesExpression(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.MemberExpression:
		// Handle subscript after function call: func()[offset]
		if call, ok := e.Object.(*ast.CallExpression); ok && e.Computed {
			funcName := g.extractFunctionName(call.Callee)
			varName := strings.ReplaceAll(funcName, ".", "_")

			// Extract offset from subscript
			offset := 0
			if lit, ok := e.Property.(*ast.Literal); ok {
				switch v := lit.Value.(type) {
				case float64:
					offset = int(v)
				case int:
					offset = v
				}
			}

			return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
		}

		if code, resolved := g.builtinHandler.TryResolveMemberExpression(e, BarLoopScope); resolved {
			return code
		}

		if obj, ok := e.Object.(*ast.Identifier); ok {
			varName := obj.Name

			if e.Computed {
				if _, ok := e.Property.(*ast.Literal); !ok {
					// Variable offset like [nA], [length]
					if g.subscriptResolver != nil {
						return g.subscriptResolver.ResolveSubscript(varName, e.Property, g)
					}
					return fmt.Sprintf("%sSeries.Get(0)", varName)
				}
			}

			// Check if it's a strategy constant (strategy.long, strategy.short)
			if prop, ok := e.Property.(*ast.Identifier); ok {
				if varName == "strategy" && (prop.Name == "long" || prop.Name == "short") {
					return g.extractMemberName(e)
				}
			}

			// Check if it's an input constant with subscript
			if funcName, isConstant := g.constants[varName]; isConstant {
				if funcName == "input.source" {
					// input.source defaults to close
					offset := 0
					if e.Computed {
						if lit, ok := e.Property.(*ast.Literal); ok {
							switch v := lit.Value.(type) {
							case float64:
								offset = int(v)
							case int:
								offset = v
							}
						}
					}
					if offset == 0 {
						return "bar.Close"
					}
					return fmt.Sprintf("ctx.Data[i-%d].Close", offset)
				}
				// Other input constants
				return varName
			}

			// User-defined variable with subscript
			offset := 0
			if e.Computed {
				if lit, ok := e.Property.(*ast.Literal); ok {
					switch v := lit.Value.(type) {
					case float64:
						offset = int(v)
					case int:
						offset = v
					}
				}
			}
			if g.variables[varName] == "array_series" {
				return fmt.Sprintf("%sArraySeries.Get(%d)", varName, offset)
			}
			return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
		}

		return g.extractMemberName(e)
	case *ast.Identifier:
		if constVal, isConstant := g.constants[e.Name]; isConstant {
			if constVal == "input.source" {
				return fmt.Sprintf("%sSeries.GetCurrent()", e.Name)
			}
			return e.Name
		}

		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, g.accessScope()); resolved {
			return code
		}

		return g.resolveUserIdentifierAccess(e.Name)
	case *ast.Literal:
		switch v := e.Value.(type) {
		case float64:
			return g.literalFormatter.FormatFloat(v)
		case int:
			return fmt.Sprintf("%d.0", v)
		case bool:
			if v {
				return "1.0"
			}
			return "0.0"
		case string:
			return fmt.Sprintf("%q", v)
		}
	case *ast.BinaryExpression:
		/* Binary expressions should be formatted with operator precedence */
		formatter := NewBinaryExpressionFormatterWithExtractor(g.extractSeriesExpression)
		return formatter.formatWithExtractor(e)
	case *ast.UnaryExpression:
		/* Unary expression like -1, +x */
		operand := g.extractSeriesExpression(e.Argument)
		op := e.Operator
		if op == "not" {
			op = "!"
		}
		return fmt.Sprintf("%s%s", op, operand)
	case *ast.CallExpression:
		return g.extractCallExpression(e)
	case *ast.ObjectExpression:
		return "/* ERROR: ObjectExpression requires ArgumentExtractor */"
	}
	return "0.0"
}

func (g *generator) convertSeriesAccessToPrev(seriesCode string) string {
	// Convert current bar access to previous bar access
	// bar.Close → ctx.Data[i-1].Close
	// sma20Series.Get(0) → sma20Series.Get(1)
	// sma20Series.GetCurrent() → sma20Series.Get(1)

	if seriesCode == "bar.Close" {
		return "ctx.Data[i-1].Close"
	}
	if seriesCode == "bar.Open" {
		return "ctx.Data[i-1].Open"
	}
	if seriesCode == "bar.High" {
		return "ctx.Data[i-1].High"
	}
	if seriesCode == "bar.Low" {
		return "ctx.Data[i-1].Low"
	}
	if seriesCode == "bar.Volume" {
		return "ctx.Data[i-1].Volume"
	}

	// Handle Series.Get(0) → Series.Get(1)
	if strings.HasSuffix(seriesCode, "Series.Get(0)") {
		return strings.Replace(seriesCode, "Series.Get(0)", "Series.Get(1)", 1)
	}

	// Handle Series.GetCurrent() → Series.Get(1)
	if strings.Contains(seriesCode, "Series.GetCurrent()") {
		return strings.ReplaceAll(seriesCode, "Series.GetCurrent()", "Series.Get(1)")
	}

	// For constants (numeric values), return unchanged - they don't need previous bar access
	return seriesCode
}

func (g *generator) convertSeriesAccessToOffset(seriesCode string, offsetVar string) string {
	if strings.HasPrefix(seriesCode, "bar.") {
		field := strings.TrimPrefix(seriesCode, "bar.")
		if seriesName, exists := g.barFieldRegistry.GetSeriesName("bar." + field); exists {
			return fmt.Sprintf("%s.Get(%s)", seriesName, offsetVar)
		}
		seriesName := g.fieldNameToOHLCVSeriesName(field)
		return fmt.Sprintf("%s.Get(%s)", seriesName, offsetVar)
	}

	if strings.Contains(seriesCode, "Series.GetCurrent()") {
		re := regexp.MustCompile(`(\w+Series)\.GetCurrent\(\)`)
		result := re.ReplaceAllString(seriesCode, fmt.Sprintf("$1.Get(%s)", offsetVar))
		return result
	}

	if strings.Contains(seriesCode, "Series.Get(") {
		re := regexp.MustCompile(`(\w+Series)\.Get\([^)]*\)`)
		result := re.ReplaceAllString(seriesCode, fmt.Sprintf("$1.Get(%s)", offsetVar))
		return result
	}

	return seriesCode
}

/* convertSeriesAccessToIntOffset converts series access code to use specific integer offset */
func (g *generator) convertSeriesAccessToIntOffset(seriesCode string, offset int) string {
	offsetStr := fmt.Sprintf("%d", offset)

	if strings.HasPrefix(seriesCode, "bar.") {
		field := strings.TrimPrefix(seriesCode, "bar.")
		if seriesName, exists := g.barFieldRegistry.GetSeriesName("bar." + field); exists {
			return fmt.Sprintf("%s.Get(%d)", seriesName, offset)
		}
		seriesName := g.fieldNameToOHLCVSeriesName(field)
		return fmt.Sprintf("%s.Get(%d)", seriesName, offset)
	}

	if strings.Contains(seriesCode, "Series.GetCurrent()") {
		re := regexp.MustCompile(`(\w+Series)\.GetCurrent\(\)`)
		result := re.ReplaceAllString(seriesCode, fmt.Sprintf("$1.Get(%s)", offsetStr))
		return result
	}

	if strings.Contains(seriesCode, "Series.Get(") {
		re := regexp.MustCompile(`(\w+Series)\.Get\([^)]*\)`)
		result := re.ReplaceAllString(seriesCode, fmt.Sprintf("$1.Get(%s)", offsetStr))
		return result
	}

	return seriesCode
}

func (g *generator) fieldNameToOHLCVSeriesName(fieldName string) string {
	return OHLCVFieldToSeriesName(fieldName)
}

/* extractIntArgument extracts integer argument from AST expression */
func (g *generator) extractIntArgument(expr ast.Expression, argName string) (int, error) {
	if lit, ok := expr.(*ast.Literal); ok {
		switch v := lit.Value.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		default:
			return 0, fmt.Errorf("%s must be integer, got %T", argName, v)
		}
	}

	/* Try constant evaluation */
	value := g.constEvaluator.EvaluateConstant(expr)
	if math.IsNaN(value) {
		return 0, fmt.Errorf("%s must be compile-time constant, got %T", argName, expr)
	}

	return int(value), nil
}

func (g *generator) generateLiteral(lit *ast.Literal) (string, error) {
	switch v := lit.Value.(type) {
	case float64:
		formatted := g.literalFormatter.FormatFloat(v)
		return g.ind() + formatted + "\n", nil
	case string:
		formatted := g.literalFormatter.FormatString(v)
		return g.ind() + formatted + "\n", nil
	case bool:
		formatted := g.literalFormatter.FormatBool(v)
		return g.ind() + formatted + "\n", nil
	default:
		formatted, err := g.literalFormatter.FormatGeneric(v)
		if err != nil {
			return "", fmt.Errorf("failed to format literal: %w", err)
		}
		return g.ind() + formatted + "\n", nil
	}
}

func (g *generator) generateMemberExpression(mem *ast.MemberExpression) (string, error) {
	obj := ""
	if id, ok := mem.Object.(*ast.Identifier); ok {
		obj = id.Name
	}
	prop := ""
	if id, ok := mem.Property.(*ast.Identifier); ok {
		prop = id.Name
	}

	if obj == "syminfo" && prop == "tickerid" {
		return "syminfo_tickerid", nil
	}

	return g.ind() + fmt.Sprintf("// %s.%s\n", obj, prop), nil
}

/* analyzeSeriesRequirements traverses AST to detect variables accessed with [offset > 0] */
func (g *generator) analyzeSeriesRequirements(node ast.Node) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.ExpressionStatement:
		g.analyzeSeriesRequirements(n.Expression)

	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			g.analyzeSeriesRequirements(decl.Init)
		}

	case *ast.CallExpression:
		// Analyze callee
		g.analyzeSeriesRequirements(n.Callee)
		// Analyze arguments
		for _, arg := range n.Arguments {
			g.analyzeSeriesRequirements(arg)
		}

	case *ast.MemberExpression:
		// No longer needed (ALL variables use Series storage)
		// Kept for future optimizations
		g.analyzeSeriesRequirements(n.Property)
		g.analyzeSeriesRequirements(n.Object)

	case *ast.BinaryExpression:
		g.analyzeSeriesRequirements(n.Left)
		g.analyzeSeriesRequirements(n.Right)

	case *ast.ConditionalExpression:
		g.analyzeSeriesRequirements(n.Test)
		g.analyzeSeriesRequirements(n.Consequent)
		g.analyzeSeriesRequirements(n.Alternate)

	case *ast.LogicalExpression:
		g.analyzeSeriesRequirements(n.Left)
		g.analyzeSeriesRequirements(n.Right)
	}
}

func (g *generator) generatePlaceholder() string {
	code := g.ind() + "// Strategy code will be generated here\n"
	code += g.ind() + fmt.Sprintf("strat.CallWithPyramiding(%q, %.0f, %d)\n", g.strategyConfig.Name, g.strategyConfig.InitialCapital, g.strategyConfig.Pyramiding)
	if g.strategyConfig.CommissionType != "" {
		code += g.ind() + fmt.Sprintf("strat.SetCommission(%.10g, %q)\n", g.strategyConfig.CommissionValue, g.strategyConfig.CommissionType)
	}
	if g.strategyConfig.DefaultQtyType != "" {
		code += g.ind() + fmt.Sprintf("strat.SetDefaultQty(%.10g, %q)\n", g.strategyConfig.DefaultQtyValue, g.strategyConfig.DefaultQtyType)
	}
	code += g.ind() + "for i := 0; i < len(ctx.Data); i++ {\n"
	g.indent++
	code += g.ind() + "ctx.BarIndex = i\n"
	code += g.ind() + "strat.OnBarUpdate(i, ctx.Data[i].Open, ctx.Data[i].Time)\n"
	code += g.ind() + "strat.OnBarMetrics(ctx.Data[i].High, ctx.Data[i].Low)\n"
	g.indent--
	code += g.ind() + "}\n"
	return code
}

func (g *generator) accessScope() AccessScope {
	if g.inArrowFunctionBody {
		return ArrowScope
	}
	return ScopeFromSecurityFlag(g.inSecurityContext)
}

func (g *generator) ind() string {
	indent := ""
	for i := 0; i < g.indent; i++ {
		indent += "\t"
	}
	return indent
}

// indentCode adds the current indentation level to each line of generated code.
// This integrates builder-generated code with the generator's indentation context.
func (g *generator) indentCode(code string) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	indented := make([]string, 0, len(lines))
	currentIndent := g.ind()

	for _, line := range lines {
		if line == "" {
			indented = append(indented, "")
		} else {
			indented = append(indented, currentIndent+line)
		}
	}

	return strings.Join(indented, "\n")
}

// generateSTDEV generates STDEV calculation using two-pass algorithm.
// Pass 1: Calculate mean, Pass 2: Calculate variance from mean.
func (g *generator) generateSTDEV(varName string, period int, accessor AccessGenerator, needsNaN bool) (string, error) {
	var code strings.Builder

	// Add header comment
	code.WriteString(g.ind() + fmt.Sprintf("/* Inline ta.stdev(%d) */\n", period))

	// Warmup check
	code.WriteString(g.ind() + fmt.Sprintf("if ctx.BarIndex < %d-1 {\n", period))
	g.indent++
	code.WriteString(g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName))
	g.indent--
	code.WriteString(g.ind() + "} else {\n")
	g.indent++

	// Pass 1: Calculate mean (inline SMA calculation)
	code.WriteString(g.ind() + "sum := 0.0\n")
	if needsNaN {
		code.WriteString(g.ind() + "hasNaN := false\n")
	}
	code.WriteString(g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period))
	g.indent++

	if needsNaN {
		code.WriteString(g.ind() + fmt.Sprintf("val := %s\n", accessor.GenerateLoopValueAccess("j")))
		code.WriteString(g.ind() + "if math.IsNaN(val) {\n")
		g.indent++
		code.WriteString(g.ind() + "hasNaN = true\n")
		code.WriteString(g.ind() + "break\n")
		g.indent--
		code.WriteString(g.ind() + "}\n")
		code.WriteString(g.ind() + "sum += val\n")
	} else {
		code.WriteString(g.ind() + fmt.Sprintf("sum += %s\n", accessor.GenerateLoopValueAccess("j")))
	}

	g.indent--
	code.WriteString(g.ind() + "}\n")

	// Check for NaN and calculate mean
	if needsNaN {
		code.WriteString(g.ind() + "if hasNaN {\n")
		g.indent++
		code.WriteString(g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName))
		g.indent--
		code.WriteString(g.ind() + "} else {\n")
		g.indent++
	}

	code.WriteString(g.ind() + fmt.Sprintf("mean := sum / %d.0\n", period))

	// Pass 2: Calculate variance
	code.WriteString(g.ind() + "variance := 0.0\n")
	code.WriteString(g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", period))
	g.indent++
	code.WriteString(g.ind() + fmt.Sprintf("diff := %s - mean\n", accessor.GenerateLoopValueAccess("j")))
	code.WriteString(g.ind() + "variance += diff * diff\n")
	g.indent--
	code.WriteString(g.ind() + "}\n")
	code.WriteString(g.ind() + fmt.Sprintf("variance /= %d.0\n", period))
	code.WriteString(g.ind() + fmt.Sprintf("%sSeries.Set(math.Sqrt(variance))\n", varName))

	if needsNaN {
		g.indent--
		code.WriteString(g.ind() + "}\n") // close else (hasNaN check)
	}

	g.indent--
	code.WriteString(g.ind() + "}\n") // close else (warmup check)

	return code.String(), nil
}

// generateRMA generates inline RMA (Relative Moving Average) calculation
// RMA uses alpha = 1/period and maintains state across bars
func (g *generator) generateRMA(varName string, period int, accessor AccessGenerator, needsNaN bool) (string, error) {
	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}
	builder := NewStatefulIndicatorBuilder("ta.rma", varName, NewConstantPeriod(period), accessor, needsNaN, context)
	return g.indentCode(builder.BuildRMA()), nil
}

/* generateRSI generates inline RSI (Relative Strength Index) calculation
 * RSI = 100 - 100/(1+RS) where RS = RMA(gains, period) / RMA(losses, period)
 */
func (g *generator) generateRSI(varName string, period int, accessor AccessGenerator, needsNaN bool) (string, error) {
	var context StatefulIndicatorContext
	if g.inArrowFunctionBody {
		context = NewArrowFunctionIndicatorContext()
	} else {
		context = NewTopLevelIndicatorContext()
	}

	builder := NewRSIIndicatorBuilder(varName, NewConstantPeriod(period), accessor, needsNaN, context)
	return g.indentCode(builder.Build()), nil
}

// generateChange generates inline change calculation
// change(source, offset) = source[0] - source[offset]
func (g *generator) generateChange(varName string, sourceExpr string, offset int) (string, error) {
	code := g.ind() + fmt.Sprintf("/* Inline ta.change(%s, %d) */\n", sourceExpr, offset)
	code += g.ind() + fmt.Sprintf("if i >= %d {\n", offset)
	g.indent++

	// Calculate difference: current - previous
	code += g.ind() + fmt.Sprintf("current := %s\n", sourceExpr)

	// Access previous value - need to adjust sourceExpr for offset
	// If sourceExpr is "bar.Close", previous is "ctx.Data[i-%d].Close"
	// If sourceExpr is "xSeries.GetCurrent()", previous is "xSeries.Get(%d)"
	prevExpr := ""
	if strings.Contains(sourceExpr, "bar.") {
		field := strings.TrimPrefix(sourceExpr, "bar.")
		prevExpr = fmt.Sprintf("ctx.Data[i-%d].%s", offset, field)
	} else if strings.Contains(sourceExpr, "Series.GetCurrent()") {
		seriesName := strings.TrimSuffix(sourceExpr, "Series.GetCurrent()")
		prevExpr = fmt.Sprintf("%sSeries.Get(%d)", seriesName, offset)
	} else {
		prevExpr = fmt.Sprintf("(/* previous value of %s */0.0)", sourceExpr)
	}

	code += g.ind() + fmt.Sprintf("previous := %s\n", prevExpr)
	code += g.ind() + fmt.Sprintf("%sSeries.Set(current - previous)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (g *generator) generateCum(varName string, sourceExpr string) (string, error) {
	code := g.ind() + fmt.Sprintf("/* Inline ta.cum(%s) */\n", sourceExpr)
	code += g.ind() + "{\n"
	g.indent++

	code += g.ind() + fmt.Sprintf("current := %s\n", sourceExpr)
	code += g.ind() + "if math.IsNaN(current) {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + "var prevSum float64\n"
	code += g.ind() + "if i > 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("prevSum = %sSeries.Get(1)\n", varName)
	code += g.ind() + "if math.IsNaN(prevSum) {\n"
	g.indent++
	code += g.ind() + "prevSum = 0.0\n"
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + fmt.Sprintf("%sSeries.Set(prevSum + current)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

func (g *generator) generateValuewhen(varName string, conditionExpr string, sourceExpr string, occurrence int) (string, error) {
	code := g.ind() + fmt.Sprintf("/* Inline valuewhen(%s, %s, %d) */\n", conditionExpr, sourceExpr, occurrence)

	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 {\n", varName)
	g.indent++

	code += g.ind() + "occurrenceCount := 0\n"
	code += g.ind() + "for lookbackOffset := 0; lookbackOffset <= i; lookbackOffset++ {\n"
	g.indent++

	conditionAccess := g.convertSeriesAccessToOffset(conditionExpr, "lookbackOffset")
	isDirectSeriesAccess := strings.Contains(conditionAccess, ".Get(") &&
		!strings.ContainsAny(conditionAccess, "><!=&|")

	if isDirectSeriesAccess {
		code += g.ind() + fmt.Sprintf("if value.IsTrue(%s) {\n", conditionAccess)
	} else {
		code += g.ind() + fmt.Sprintf("if value.IsTrue(func() float64 { if %s { return 1.0 } else { return 0.0 } }()) {\n", conditionAccess)
	}
	g.indent++

	code += g.ind() + fmt.Sprintf("if occurrenceCount == %d {\n", occurrence)
	g.indent++

	sourceAccess := g.convertSeriesAccessToOffset(sourceExpr, "lookbackOffset")
	code += g.ind() + fmt.Sprintf("return %s\n", sourceAccess)

	g.indent--
	code += g.ind() + "}\n"
	code += g.ind() + "occurrenceCount++\n"

	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "}\n"

	code += g.ind() + "return math.NaN()\n"

	g.indent--
	code += g.ind() + "}())\n"

	return code, nil
}

/* generatePivot generates inline delayed pivot detection code.
 * Uses backward-only window scan: at bar i, calculates for bar (i - rightBars).
 * All data access through SeriesBuffer.Get(offset) where offset >= 0 (historical).
 * Supports 2-arg form: pivothigh(left, right) uses high, pivotlow(left, right) uses low.
 */
func (g *generator) generatePivot(varName string, call *ast.CallExpression, isHigh bool) (string, error) {
	var sourceExpr ast.Expression
	var leftBars, rightBars int
	var err error

	if len(call.Arguments) == 2 {
		/* 2-arg form: pivothigh(leftBars, rightBars) - use default source */
		if isHigh {
			sourceExpr = &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "high"}
		} else {
			sourceExpr = &ast.Identifier{NodeType: ast.TypeIdentifier, Name: "low"}
		}
		leftBars, err = g.extractIntArgument(call.Arguments[0], "leftBars")
		if err != nil {
			return "", err
		}
		rightBars, err = g.extractIntArgument(call.Arguments[1], "rightBars")
		if err != nil {
			return "", err
		}
	} else if len(call.Arguments) >= 3 {
		/* 3-arg form: pivothigh(source, leftBars, rightBars) */
		sourceExpr = call.Arguments[0]
		leftBars, err = g.extractIntArgument(call.Arguments[1], "leftBars")
		if err != nil {
			return "", err
		}
		rightBars, err = g.extractIntArgument(call.Arguments[2], "rightBars")
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("pivot requires 2 or 3 arguments")
	}

	if leftBars < 1 || rightBars < 1 {
		return "", fmt.Errorf("pivot leftBars and rightBars must be >= 1, got left=%d right=%d", leftBars, rightBars)
	}

	totalWidth := leftBars + rightBars + 1
	sourceAccess := g.extractSeriesExpression(sourceExpr)
	comparisonOp := ">"
	if !isHigh {
		comparisonOp = "<"
	}

	var code string
	code += g.ind() + fmt.Sprintf("if i >= %d {\n", totalWidth-1)
	g.indent++

	code += g.ind() + fmt.Sprintf("centerValue := %s\n", g.convertSeriesAccessToIntOffset(sourceAccess, rightBars))
	code += g.ind() + "if !math.IsNaN(centerValue) {\n"
	g.indent++
	code += g.ind() + "isPivot := true\n\n"

	for j := 0; j < leftBars; j++ {
		offset := totalWidth - 1 - j
		code += g.ind() + fmt.Sprintf("if leftVal := %s; !math.IsNaN(leftVal) && leftVal %s= centerValue {\n", g.convertSeriesAccessToIntOffset(sourceAccess, offset), comparisonOp)
		g.indent++
		code += g.ind() + "isPivot = false\n"
		g.indent--
		code += g.ind() + "}\n"
	}

	code += g.ind() + "\n"
	for j := 1; j <= rightBars; j++ {
		offset := rightBars - j
		code += g.ind() + fmt.Sprintf("if rightVal := %s; !math.IsNaN(rightVal) && rightVal %s= centerValue {\n", g.convertSeriesAccessToIntOffset(sourceAccess, offset), comparisonOp)
		g.indent++
		code += g.ind() + "isPivot = false\n"
		g.indent--
		code += g.ind() + "}\n"
	}

	code += g.ind() + "\nif isPivot {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(centerValue)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

// collectNestedVariables recursively scans CallExpression arguments for nested function calls
func (g *generator) collectNestedVariables(parentVarName string, call *ast.CallExpression) {
	funcName := g.extractFunctionName(call.Callee)

	// Only collect nested variables for functions that support it (fixnan)
	if funcName != "fixnan" {
		return
	}

	// Scan arguments for nested CallExpression
	for _, arg := range call.Arguments {
		g.scanForNestedCalls(parentVarName, arg)
	}
}

// scanForNestedCalls recursively searches for CallExpression in MemberExpression
func (g *generator) scanForNestedCalls(parentVarName string, expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.MemberExpression:
		if nestedCall, ok := e.Object.(*ast.CallExpression); ok {
			nestedFuncName := g.extractFunctionName(nestedCall.Callee)

			if g.runtimeOnlyFilter.IsRuntimeOnly(nestedFuncName) {
				return
			}

			tempVarName := strings.ReplaceAll(nestedFuncName, ".", "_")

			if _, exists := g.variables[tempVarName]; !exists {
				g.variables[tempVarName] = "float"
			}
		}
		// Recurse into object and property
		g.scanForNestedCalls(parentVarName, e.Object)
		g.scanForNestedCalls(parentVarName, e.Property)

	case *ast.CallExpression:
		// Recurse into arguments
		for _, arg := range e.Arguments {
			g.scanForNestedCalls(parentVarName, arg)
		}
	}
}

// scanForSubscriptedCalls scans any expression for subscripted function calls
func (g *generator) scanForSubscriptedCalls(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		// Check if object is CallExpression with subscript: func()[offset]
		if call, ok := e.Object.(*ast.CallExpression); ok && e.Computed {
			funcName := g.extractFunctionName(call.Callee)
			varName := strings.ReplaceAll(funcName, ".", "_")

			// Register variable for Series initialization
			if _, exists := g.variables[varName]; !exists {
				g.variables[varName] = "float"
			}
		}
		// Recurse
		g.scanForSubscriptedCalls(e.Object)
		g.scanForSubscriptedCalls(e.Property)

	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			g.scanForSubscriptedCalls(arg)
		}

	case *ast.BinaryExpression:
		g.scanForSubscriptedCalls(e.Left)
		g.scanForSubscriptedCalls(e.Right)

	case *ast.UnaryExpression:
		g.scanForSubscriptedCalls(e.Argument)

	case *ast.ConditionalExpression:
		g.scanForSubscriptedCalls(e.Test)
		g.scanForSubscriptedCalls(e.Consequent)
		g.scanForSubscriptedCalls(e.Alternate)
	}
}

func (g *generator) serializeExpressionForRuntime(expr ast.Expression) (string, error) {
	switch exp := expr.(type) {
	case *ast.Identifier:
		return fmt.Sprintf("&ast.Identifier{Name: %q}", exp.Name), nil
	case *ast.Literal:
		if val, ok := exp.Value.(float64); ok {
			return fmt.Sprintf("&ast.Literal{Value: %.1f}", val), nil
		}
		if val, ok := exp.Value.(string); ok {
			return fmt.Sprintf("&ast.Literal{Value: %q}", val), nil
		}
		if val, ok := exp.Value.(bool); ok {
			return fmt.Sprintf("&ast.Literal{Value: %t}", val), nil
		}
		return "", fmt.Errorf("unsupported literal type: %T", exp.Value)
	case *ast.MemberExpression:
		objectCode, err := g.serializeExpressionForRuntime(exp.Object)
		if err != nil {
			return "", err
		}
		propertyCode, err := g.serializeExpressionForRuntime(exp.Property)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("&ast.MemberExpression{Object: %s, Property: %s}", objectCode, propertyCode), nil
	case *ast.CallExpression:
		funcName := g.extractFunctionName(exp.Callee)
		parts := strings.Split(funcName, ".")
		if len(parts) == 1 {
			parts = []string{"ta", parts[0]}
		}
		if len(parts) != 2 {
			return "", fmt.Errorf("unsupported function name format: %s", funcName)
		}

		args := ""
		for i, arg := range exp.Arguments {
			argCode, err := g.serializeExpressionForRuntime(arg)
			if err != nil {
				return "", err
			}
			if i > 0 {
				args += ", "
			}
			args += argCode
		}

		return fmt.Sprintf("&ast.CallExpression{Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: %q}, Property: &ast.Identifier{Name: %q}}, Arguments: []ast.Expression{%s}}",
			parts[0], parts[1], args), nil
	case *ast.BinaryExpression:
		leftCode, err := g.serializeExpressionForRuntime(exp.Left)
		if err != nil {
			return "", err
		}

		rightCode, err := g.serializeExpressionForRuntime(exp.Right)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("&ast.BinaryExpression{Operator: %q, Left: %s, Right: %s}",
			exp.Operator, leftCode, rightCode), nil
	case *ast.ConditionalExpression:
		testCode, err := g.serializeExpressionForRuntime(exp.Test)
		if err != nil {
			return "", err
		}

		consequentCode, err := g.serializeExpressionForRuntime(exp.Consequent)
		if err != nil {
			return "", err
		}

		alternateCode, err := g.serializeExpressionForRuntime(exp.Alternate)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("&ast.ConditionalExpression{Test: %s, Consequent: %s, Alternate: %s}",
			testCode, consequentCode, alternateCode), nil
	default:
		return "", fmt.Errorf("unsupported expression type for runtime serialization: %T", expr)
	}
}

// extractConstValue parses "const varName = VALUE" to extract VALUE
// Deprecated: Use ConstantRegistry.ExtractFromGeneratedCode
func extractConstValue(code string) interface{} {
	var varName string
	var floatVal float64
	var intVal int
	var boolVal bool

	if _, err := fmt.Sscanf(code, "const %s = %f", &varName, &floatVal); err == nil {
		return floatVal
	}
	if _, err := fmt.Sscanf(code, "const %s = %d", &varName, &intVal); err == nil {
		return intVal
	}
	if _, err := fmt.Sscanf(code, "const %s = %t", &varName, &boolVal); err == nil {
		return boolVal
	}
	return nil
}

/* detectSecurityCalls delegates to security package for complete AST analysis */
func detectSecurityCalls(program *ast.Program) bool {
	return len(security.AnalyzeAST(program)) > 0
}

/* detectStrategyRuntimeAccess walks AST to detect strategy.* runtime value access */
func detectStrategyRuntimeAccess(program *ast.Program) bool {
	if program == nil {
		return false
	}

	for _, node := range program.Body {
		if hasStrategyRuntimeInNode(node) {
			return true
		}
	}
	return false
}

func hasStrategyRuntimeInNode(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if hasStrategyRuntimeInExpression(decl.Init) {
				return true
			}
		}
	case *ast.ExpressionStatement:
		return hasStrategyRuntimeInExpression(n.Expression)
	case *ast.IfStatement:
		if hasStrategyRuntimeInExpression(n.Test) {
			return true
		}
		for _, consequent := range n.Consequent {
			if hasStrategyRuntimeInNode(consequent) {
				return true
			}
		}
		for _, alternate := range n.Alternate {
			if hasStrategyRuntimeInNode(alternate) {
				return true
			}
		}
	}
	return false
}

func hasStrategyRuntimeInExpression(expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					runtimeProps := map[string]bool{
						"position_avg_price":   true,
						"position_size":        true,
						"equity":               true,
						"netprofit":            true,
						"closedtrades":         true,
						"opentrades":           true,
						"max_drawdown":         true,
						"max_runup":            true,
						"max_drawdown_percent": true,
						"max_runup_percent":    true,
						"initial_capital":      true,
						"grossprofit":          true,
						"grossloss":            true,
						"wintrades":            true,
						"losstrades":           true,
						"eventrades":           true,
						"openprofit":           true,
						"avg_trade":            true,
						"avg_winning_trade":    true,
						"avg_losing_trade":     true,
						"position_entry_name":  true,
					}
					if runtimeProps[prop.Name] {
						return true
					}
				}
			}
		}
		return hasStrategyRuntimeInExpression(e.Object)
	case *ast.CallExpression:
		if hasStrategyRuntimeInExpression(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if hasStrategyRuntimeInExpression(arg) {
				return true
			}
		}
	case *ast.BinaryExpression:
		return hasStrategyRuntimeInExpression(e.Left) || hasStrategyRuntimeInExpression(e.Right)
	case *ast.LogicalExpression:
		return hasStrategyRuntimeInExpression(e.Left) || hasStrategyRuntimeInExpression(e.Right)
	case *ast.ConditionalExpression:
		return hasStrategyRuntimeInExpression(e.Test) || hasStrategyRuntimeInExpression(e.Consequent) || hasStrategyRuntimeInExpression(e.Alternate)
	}
	return false
}
