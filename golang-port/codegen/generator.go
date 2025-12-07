package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/validation"
)

/* StrategyCode holds generated Go code for strategy execution */
type StrategyCode struct {
	FunctionBody string // executeStrategy() function body
	StrategyName string // Pine Script strategy name
}

/* GenerateStrategyCodeFromAST converts parsed Pine ESTree to Go runtime code */
func GenerateStrategyCodeFromAST(program *ast.Program) (*StrategyCode, error) {
	gen := &generator{
		imports:      make(map[string]bool),
		variables:    make(map[string]string),
		varInits:     make(map[string]ast.Expression),
		constants:    make(map[string]interface{}),
		strategyName: "Generated Strategy",
		limits:       NewCodeGenerationLimits(),
		safetyGuard:  NewRuntimeSafetyGuard(),
	}

	// Initialize handlers
	gen.inputHandler = NewInputHandler()
	gen.mathHandler = NewMathHandler()
	gen.subscriptResolver = NewSubscriptResolver()
	gen.builtinHandler = NewBuiltinIdentifierHandler()
	gen.taRegistry = NewTAFunctionRegistry()
	gen.exprAnalyzer = NewExpressionAnalyzer(gen)       // Expression analysis for temp vars
	gen.tempVarMgr = NewTempVariableManager(gen)        // ForwardSeriesBuffer temp var manager
	gen.constEvaluator = validation.NewWarmupAnalyzer() // Compile-time constant evaluator
	gen.plotExprHandler = NewPlotExpressionHandler(gen) // Inline TA/math in plot() expressions

	body, err := gen.generateProgram(program)
	if err != nil {
		return nil, err
	}

	code := &StrategyCode{
		FunctionBody: body,
		StrategyName: gen.strategyName,
	}

	return code, nil
}

type generator struct {
	imports           map[string]bool
	variables         map[string]string
	varInits          map[string]ast.Expression // Variable init expressions for constant resolution
	constants         map[string]interface{}    // Input constants (input.float, input.int, etc)
	plots             []string                  // Track plot variables
	strategyName      string                    // Strategy name from indicator() or strategy()
	indent            int
	taFunctions       []taFunctionCall // List of TA function calls to pre-calculate
	inSecurityContext bool             // Flag when generating code inside security() context
	limits            CodeGenerationLimits
	safetyGuard       RuntimeSafetyGuard
	inputHandler      *InputHandler
	mathHandler       *MathHandler
	subscriptResolver *SubscriptResolver
	builtinHandler    *BuiltinIdentifierHandler  // Resolves Pine built-in identifiers
	taRegistry        *TAFunctionRegistry        // Registry for TA function handlers
	exprAnalyzer      *ExpressionAnalyzer        // Finds nested TA calls in expressions
	tempVarMgr        *TempVariableManager       // Manages temp Series variables (ForwardSeriesBuffer)
	constEvaluator    *validation.WarmupAnalyzer // Compile-time constant expression evaluator
	plotExprHandler   *PlotExpressionHandler     // Handles inline TA/math in plot() expressions
}

type taFunctionCall struct {
	varName  string
	funcName string
	args     []ast.Expression
}

func (g *generator) generateProgram(program *ast.Program) (string, error) {
	if program == nil || len(program.Body) == 0 {
		return g.generatePlaceholder(), nil
	}

	// Initialize safety limits if not already set (for tests)
	if g.limits.MaxStatementsPerPass == 0 {
		g.limits = NewCodeGenerationLimits()
		g.safetyGuard = NewRuntimeSafetyGuard()
	}

	// PRE-PASS: Collect AST constants for expression evaluator
	for _, stmt := range program.Body {
		g.constEvaluator.CollectConstants(stmt)
	}

	// First pass: collect variables, analyze Series requirements, extract strategy name
	statementCounter := NewStatementCounter(g.limits)
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
						// Extract title from first argument or from 'title=' named parameter
						strategyName := g.extractStrategyName(call.Arguments)
						if strategyName != "" {
							g.strategyName = strategyName
						}
					}
				}
				// Handle v4 'study()' and v5 'indicator()' as Identifier calls
				if id, ok := call.Callee.(*ast.Identifier); ok {
					if id.Name == "study" || id.Name == "indicator" || id.Name == "strategy" {
						// Extract title from first argument or from 'title=' named parameter
						strategyName := g.extractStrategyName(call.Arguments)
						if strategyName != "" {
							g.strategyName = strategyName
						}
					}
				}
			}
		}

		// Collect variable declarations
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				varName := declarator.ID.Name

				// Check if this is an input.* function call
				if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
					funcName := g.extractFunctionName(callExpr.Callee)

					// Generate input constants immediately (if handler exists)
					if g.inputHandler != nil {
						// Handle Pine v4 generic input() - infer type from first arg
						if funcName == "input" && len(callExpr.Arguments) > 0 {
							if lit, ok := callExpr.Arguments[0].(*ast.Literal); ok {
								// Check if value is float or int
								switch v := lit.Value.(type) {
								case float64:
									if v == float64(int(v)) {
										funcName = "input.int"
									} else {
										funcName = "input.float"
									}
								case int:
									funcName = "input.int"
								}
							}
						}

						if funcName == "input.float" {
							code, _ := g.inputHandler.GenerateInputFloat(callExpr, varName)
							if code != "" {
								// Extract value from generated code: "const varName = 1.23"
								if val := extractConstValue(code); val != nil {
									g.constants[varName] = val
								}
							}
							continue
						}
						if funcName == "input.int" {
							code, _ := g.inputHandler.GenerateInputInt(callExpr, varName)
							if code != "" {
								if val := extractConstValue(code); val != nil {
									g.constants[varName] = val
								}
							}
							continue
						}
						if funcName == "input.bool" {
							g.inputHandler.GenerateInputBool(callExpr, varName)
							continue
						}
						if funcName == "input.string" {
							g.inputHandler.GenerateInputString(callExpr, varName)
							continue
						}
						if funcName == "input.session" {
							g.inputHandler.GenerateInputSession(callExpr, varName)
							continue
						}
					}
					if funcName == "input.source" {
						// input.source is an alias to an existing series
						// Don't add to variables - handle specially in codegen
						g.constants[varName] = funcName
						continue
					}

					// Collect nested function variables (fixnan(pivothigh()[1]))
					g.collectNestedVariables(varName, callExpr)
				}

				// Scan ALL initializers for subscripted function calls: pivothigh()[1]
				g.scanForSubscriptedCalls(declarator.Init)

				varType := g.inferVariableType(declarator.Init)
				g.variables[varName] = varType
			}
		}
	}

	// Sync input constants to constEvaluator AFTER first pass collects them
	for varName, value := range g.constants {
		if floatVal, ok := value.(float64); ok {
			g.constEvaluator.AddConstant(varName, floatVal)
		} else if intVal, ok := value.(int); ok {
			g.constEvaluator.AddConstant(varName, float64(intVal))
		}
	}

	// Pre-analyze security() calls to register temp vars BEFORE declarations
	g.preAnalyzeSecurityCalls(program)

	// Second pass: No longer needed (ALL variables use Series storage)
	// Kept for future optimizations if needed

	// Third pass: collect TA function calls for pre-calculation
	statementCounter.Reset()
	for _, stmt := range program.Body {
		if err := statementCounter.Increment(); err != nil {
			return "", err
		}
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
					funcName := g.extractFunctionName(callExpr.Callee)
					if funcName == "ta.sma" || funcName == "ta.ema" || funcName == "ta.rma" ||
						funcName == "ta.rsi" || funcName == "ta.atr" || funcName == "ta.stdev" ||
						funcName == "ta.change" || funcName == "ta.pivothigh" || funcName == "ta.pivotlow" ||
						funcName == "fixnan" {
						g.taFunctions = append(g.taFunctions, taFunctionCall{
							varName:  declarator.ID.Name,
							funcName: funcName,
							args:     callExpr.Arguments,
						})
					}
				}
			}
		}
	}

	code := ""

	// Initialize strategy
	code += g.ind() + fmt.Sprintf("strat.Call(%q, 10000)\n\n", g.strategyName)

	// Generate input constants
	if g.inputHandler != nil && len(g.inputHandler.inputConstants) > 0 {
		code += g.ind() + "// Input constants\n"
		for _, constCode := range g.inputHandler.inputConstants {
			code += g.ind() + constCode
		}
		code += "\n"
	}

	// Declare ALL variables as Series (ForwardSeriesBuffer paradigm)
	if len(g.variables) > 0 {
		code += g.ind() + "// ALL variables use Series storage (ForwardSeriesBuffer paradigm)\n"
		for varName := range g.variables {
			code += g.ind() + fmt.Sprintf("var %sSeries *series.Series\n", varName)
		}
		code += "\n"

		// Declare temp variables for nested TA calls (managed by TempVariableManager)
		tempVarDecls := g.tempVarMgr.GenerateDeclarations()
		if tempVarDecls != "" {
			code += tempVarDecls + "\n"
		}

		// Declare state variables for fixnan
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

		// Initialize ALL Series before bar loop
		code += g.ind() + "// Initialize Series storage\n"
		for varName := range g.variables {
			code += g.ind() + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", varName)
		}

		// Initialize temp variable Series (ForwardSeriesBuffer paradigm)
		tempVarInits := g.tempVarMgr.GenerateInitializations()
		if tempVarInits != "" {
			code += tempVarInits
		}
		code += "\n"
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
	code += g.ind() + "strat.OnBarUpdate(i, bar.Open, bar.Time)\n\n"

	// Generate statements inside bar loop
	statementCounter.Reset()
	for _, stmt := range program.Body {
		if err := statementCounter.Increment(); err != nil {
			return "", err
		}
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		code += stmtCode
	}

	// Suppress unused variable warnings
	code += "\n" + g.ind() + "// Suppress unused variable warnings\n"
	for varName := range g.variables {
		code += g.ind() + fmt.Sprintf("_ = %sSeries\n", varName)
	}

	// Advance Series cursors at end of bar loop
	code += "\n" + g.ind() + "// Advance Series cursors\n"
	for varName := range g.variables {
		code += g.ind() + fmt.Sprintf("if %s < barCount-1 { %sSeries.Next() }\n", iterVar, varName)
	}

	// Advance temp variable Series cursors (ForwardSeriesBuffer paradigm)
	tempVarNextCalls := g.tempVarMgr.GenerateNextCalls()
	if tempVarNextCalls != "" {
		code += tempVarNextCalls
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
	default:
		return "", fmt.Errorf("unsupported statement type: %T", node)
	}
}

func (g *generator) generateExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
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
		return g.ind() + "// " + e.Name + "\n", nil
	case *ast.Literal:
		return g.generateLiteral(e)
	case *ast.MemberExpression:
		return g.generateMemberExpression(e)
	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (g *generator) generateCallExpression(call *ast.CallExpression) (string, error) {
	// Extract function name
	funcName := ""
	switch callee := call.Callee.(type) {
	case *ast.Identifier:
		funcName = callee.Name
	case *ast.MemberExpression:
		// Handle ta.sma, strategy.entry, etc.
		obj := ""
		if id, ok := callee.Object.(*ast.Identifier); ok {
			obj = id.Name
		}
		prop := ""
		if id, ok := callee.Property.(*ast.Identifier); ok {
			prop = id.Name
		}
		funcName = obj + "." + prop
	}

	// Handle specific Pine functions
	code := ""
	switch funcName {
	case "indicator", "strategy":
		// Strategy/indicator initialization - skip in bar loop
		return "", nil
	case "plot":
		// Plot function - add to collector
		opts := ParsePlotOptions(call)

		// Generate expression for the plot value
		var plotExpr string
		if opts.Variable != "" {
			// Simple variable reference
			plotExpr = opts.Variable + "Series.Get(0)"
		} else if len(call.Arguments) > 0 {
			// Inline expression - generate numeric code for it
			exprCode, err := g.generatePlotExpression(call.Arguments[0])
			if err != nil {
				return "", err
			}
			plotExpr = exprCode
		}

		if plotExpr != "" {
			code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, nil)\n", opts.Title, plotExpr)
		}
	case "ta.sma":
		// SMA calculation - handled in variable declaration
		return "", nil
	case "strategy.entry":
		// strategy.entry(id, direction, qty)
		if len(call.Arguments) >= 2 {
			entryID := g.extractStringLiteral(call.Arguments[0])
			direction := g.extractDirectionConstant(call.Arguments[1])
			qty := 1.0
			if len(call.Arguments) >= 3 {
				qty = g.extractFloatLiteral(call.Arguments[2])
			}

			code += g.ind() + fmt.Sprintf("strat.Entry(%q, %s, %.0f)\n", entryID, direction, qty)
		}
	case "strategy.close":
		// strategy.close(id)
		if len(call.Arguments) >= 1 {
			entryID := g.extractStringLiteral(call.Arguments[0])
			code += g.ind() + fmt.Sprintf("strat.Close(%q, bar.Close, bar.Time)\n", entryID)
		}
	case "strategy.close_all":
		// strategy.close_all()
		code += g.ind() + "strat.CloseAll(bar.Close, bar.Time)\n"
	case "ta.crossover", "ta.crossunder":
		// Crossover functions - handled in variable declaration
		return "", nil
	case "ta.stdev", "ta.change", "ta.pivothigh", "ta.pivotlow", "fixnan":
		// TA functions - handled in variable declaration
		return "", nil
	case "valuewhen":
		// Value functions - handled in variable declaration
		return "", nil
	default:
		code += g.ind() + fmt.Sprintf("// %s() - TODO: implement\n", funcName)
	}

	return code, nil
}

func (g *generator) generateIfStatement(ifStmt *ast.IfStatement) (string, error) {
	// Generate condition expression
	condition, err := g.generateConditionExpression(ifStmt.Test)
	if err != nil {
		return "", err
	}

	// If the condition accesses a bool Series variable, add != 0 conversion
	condition = g.addBoolConversionIfNeeded(ifStmt.Test, condition)

	code := g.ind() + fmt.Sprintf("if %s {\n", condition)
	g.indent++

	// Generate consequent (body) statements
	hasValidBody := false
	for _, stmt := range ifStmt.Consequent {
		// Parser limitation: indented blocks sometimes parsed incorrectly
		// Skip expression-only statements in if body (likely parsing artifacts)
		if exprStmt, ok := stmt.(*ast.ExpressionStatement); ok {
			// Check if expression is non-call (BinaryExpression, LogicalExpression, etc.)
			switch exprStmt.Expression.(type) {
			case *ast.CallExpression:
				// Valid call statement - generate
			case *ast.Identifier, *ast.Literal:
				// Simple expression - skip (parsing artifact)
				continue
			case *ast.BinaryExpression, *ast.LogicalExpression, *ast.ConditionalExpression:
				// Condition expression in body - skip (parsing artifact)
				continue
			}
		}

		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		if stmtCode != "" {
			code += stmtCode
			hasValidBody = true
		}
	}

	// If no valid body statements, add comment
	if !hasValidBody {
		code += g.ind() + "// TODO: if body statements\n"
	}

	g.indent--
	code += g.ind() + "}\n"

	// TODO: Handle alternate (else) if needed

	return code, nil
}

func (g *generator) generateBinaryExpression(binExpr *ast.BinaryExpression) (string, error) {
	// Binary expressions should be handled in condition context
	// This is just a fallback - shouldn't be called directly
	return "", fmt.Errorf("binary expression should be used in condition context")
}

func (g *generator) generateUnaryExpression(unaryExpr *ast.UnaryExpression) (string, error) {
	// Generate the operand
	operandCode, err := g.generateConditionExpression(unaryExpr.Argument)
	if err != nil {
		return "", err
	}

	// Map Pine unary operators to Go operators
	op := unaryExpr.Operator
	switch op {
	case "not":
		op = "!"
	}

	return fmt.Sprintf("%s%s", op, operandCode), nil
}

func (g *generator) generateLogicalExpression(logExpr *ast.LogicalExpression) (string, error) {
	// Generate left expression
	leftCode, err := g.generateConditionExpression(logExpr.Left)
	if err != nil {
		return "", err
	}

	// Generate right expression
	rightCode, err := g.generateConditionExpression(logExpr.Right)
	if err != nil {
		return "", err
	}

	// Map Pine logical operators to Go operators
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
	// Generate test condition
	testCode, err := g.generateConditionExpression(condExpr.Test)
	if err != nil {
		return "", err
	}

	// If the test accesses a bool Series variable, add != 0 conversion
	testCode = g.addBoolConversionIfNeeded(condExpr.Test, testCode)

	// Generate consequent (true branch)
	consequentCode, err := g.generateConditionExpression(condExpr.Consequent)
	if err != nil {
		return "", err
	}

	// Generate alternate (false branch)
	alternateCode, err := g.generateConditionExpression(condExpr.Alternate)
	if err != nil {
		return "", err
	}

	// Generate Go ternary-style code using if-else expression
	// Go doesn't have ternary operator, so we use a function-like pattern
	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		testCode, consequentCode, alternateCode), nil
}

// addBoolConversionIfNeeded checks if the expression accesses a bool Series variable
// and wraps the code with != 0 conversion for use in boolean contexts
func (g *generator) addBoolConversionIfNeeded(expr ast.Expression, code string) string {
	needsConversion := false

	// Check if this is a simple identifier that maps to a bool variable
	if ident, ok := expr.(*ast.Identifier); ok {
		if varType, exists := g.variables[ident.Name]; exists && varType == "bool" {
			needsConversion = true
		}
	}

	// Check if this is a member expression (e.g., signal[0]) that accesses a bool Series
	if member, ok := expr.(*ast.MemberExpression); ok {
		if ident, ok := member.Object.(*ast.Identifier); ok {
			if varType, exists := g.variables[ident.Name]; exists && varType == "bool" {
				needsConversion = true
			}
		}
	}

	if needsConversion {
		return fmt.Sprintf("%s != 0", code)
	}
	return code
}

// generateNumericExpression generates code for expressions that must produce float64 values
// Converts boolean literals to 1.0 (true) or 0.0 (false)
func (g *generator) generateNumericExpression(expr ast.Expression) (string, error) {
	// Special handling for boolean literals: convert to float
	if lit, ok := expr.(*ast.Literal); ok {
		if boolVal, ok := lit.Value.(bool); ok {
			if boolVal {
				return "1.0", nil
			}
			return "0.0", nil
		}
	}

	// For all other expressions, use generateConditionExpression which produces values
	return g.generateConditionExpression(expr)
}

// generatePlotExpression generates inline code for plot() argument expressions
// Handles ternary expressions, identifiers, and literals as immediate values
func (g *generator) generatePlotExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ConditionalExpression:
		// Handle ternary: test ? consequent : alternate
		// Generate as inline func() float64 expression
		condCode, err := g.generateConditionExpression(e.Test)
		if err != nil {
			return "", err
		}
		// Add != 0 conversion for Series variables used in boolean context
		if _, ok := e.Test.(*ast.Identifier); ok {
			condCode = condCode + " != 0"
		} else if _, ok := e.Test.(*ast.MemberExpression); ok {
			condCode = condCode + " != 0"
		}

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
		// Variable reference - use Series.Get(0)
		return e.Name + "Series.Get(0)", nil

	case *ast.MemberExpression:
		// Member expression like close[0]
		return g.extractSeriesExpression(e), nil

	case *ast.Literal:
		// Direct literal value
		return g.generateNumericExpression(e)

	case *ast.BinaryExpression, *ast.LogicalExpression:
		// Mathematical or logical expression
		return g.generateConditionExpression(expr)

	case *ast.CallExpression:
		// Inline TA/math functions: plot(sma(close, 20)), plot(math.max(high, low))
		return g.plotExprHandler.Generate(expr)

	default:
		return "", fmt.Errorf("unsupported plot expression type: %T", expr)
	}
}

func (g *generator) generateConditionExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ConditionalExpression:
		// Handle ternary expressions in condition context
		testCode, err := g.generateConditionExpression(e.Test)
		if err != nil {
			return "", err
		}
		consequentCode, err := g.generateConditionExpression(e.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateConditionExpression(e.Alternate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
			testCode, consequentCode, alternateCode), nil

	case *ast.UnaryExpression:
		// Handle unary expressions (-x, +x, !x, not x)
		operandCode, err := g.generateConditionExpression(e.Argument)
		if err != nil {
			return "", err
		}
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
		op := e.Operator
		switch op {
		case "and":
			op = "&&"
		case "or":
			op = "||"
		}
		return fmt.Sprintf("(%s %s %s)", leftCode, op, rightCode), nil

	case *ast.BinaryExpression:
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

		return fmt.Sprintf("%s %s %s", left, op, right), nil

	case *ast.MemberExpression:
		// Use extractSeriesExpression for proper offset handling
		return g.extractSeriesExpression(e), nil

	case *ast.Identifier:
		// Special built-in identifiers
		if e.Name == "na" {
			return "math.NaN()", nil
		}
		varName := e.Name

		// Check if it's a Pine built-in series variable
		switch varName {
		case "close":
			return "bar.Close", nil
		case "open":
			return "bar.Open", nil
		case "high":
			return "bar.High", nil
		case "low":
			return "bar.Low", nil
		case "volume":
			return "bar.Volume", nil
		}

		// Check if it's an input constant
		if _, isConstant := g.constants[varName]; isConstant {
			return varName, nil
		}

		// User-defined variable (ALL use Series storage)
		return fmt.Sprintf("%sSeries.GetCurrent()", varName), nil

	case *ast.Literal:
		switch v := e.Value.(type) {
		case float64:
			return fmt.Sprintf("%.2f", v), nil
		case bool:
			return fmt.Sprintf("%t", v), nil
		case string:
			return fmt.Sprintf("%q", v), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}

	case *ast.CallExpression:
		// Handle inline function calls in conditions (e.g., na(time(...)))
		funcName := g.extractFunctionName(e.Callee)

		switch funcName {
		case "na":
			// na(expr) checks if value is NaN
			if len(e.Arguments) >= 1 {
				argCode, err := g.generateConditionExpression(e.Arguments[0])
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("math.IsNaN(%s)", argCode), nil
			}
			return "true", nil

		case "time":
			// time() function returns timestamp or NaN
			handler := NewTimeHandler(g.ind())
			return handler.HandleInlineExpression(e.Arguments), nil

		case "math.min", "math.max", "math.pow", "math.abs", "math.sqrt",
			"math.floor", "math.ceil", "math.round", "math.log", "math.exp":
			mathHandler := NewMathHandler()
			return mathHandler.GenerateMathCall(funcName, e.Arguments, g)

		case "ta.dev", "dev":
			if len(e.Arguments) < 2 {
				return "", fmt.Errorf("dev requires 2 arguments (source, length)")
			}
			sourceExpr := g.extractSeriesExpression(e.Arguments[0])
			lengthExpr := g.extractSeriesExpression(e.Arguments[1])
			return fmt.Sprintf("(func() float64 { length := int(%s); if ctx.BarIndex < length-1 { return math.NaN() }; sum := 0.0; for j := 0; j < length; j++ { sum += %s }; mean := sum / float64(length); devSum := 0.0; for j := 0; j < length; j++ { devSum += math.Abs(%s - mean) }; return devSum / float64(length) }())", lengthExpr, sourceExpr, sourceExpr), nil

		case "ta.crossover", "crossover", "ta.crossunder", "crossunder":
			if len(e.Arguments) < 2 {
				return "", fmt.Errorf("%s requires 2 arguments", funcName)
			}

			arg1Call, isCall1 := e.Arguments[0].(*ast.CallExpression)
			arg2Call, isCall2 := e.Arguments[1].(*ast.CallExpression)

			if !isCall1 || !isCall2 {
				return "", fmt.Errorf("%s requires CallExpression arguments for inline generation", funcName)
			}

			inline1, err := g.plotExprHandler.Generate(arg1Call)
			if err != nil {
				return "", fmt.Errorf("%s arg1 inline generation failed: %w", funcName, err)
			}
			inline2, err := g.plotExprHandler.Generate(arg2Call)
			if err != nil {
				return "", fmt.Errorf("%s arg2 inline generation failed: %w", funcName, err)
			}

			if funcName == "ta.crossover" || funcName == "crossover" {
				return fmt.Sprintf("(func() bool { if ctx.BarIndex == 0 { return false }; curr1 := (%s); curr2 := (%s); prevBarIdx := ctx.BarIndex; ctx.BarIndex--; prev1 := (%s); prev2 := (%s); ctx.BarIndex = prevBarIdx; return curr1 > curr2 && prev1 <= prev2 }())",
					inline1, inline2, inline1, inline2), nil
			}
			return fmt.Sprintf("(func() bool { if ctx.BarIndex == 0 { return false }; curr1 := (%s); curr2 := (%s); prevBarIdx := ctx.BarIndex; ctx.BarIndex--; prev1 := (%s); prev2 := (%s); ctx.BarIndex = prevBarIdx; return curr1 < curr2 && prev1 >= prev2 }())",
				inline1, inline2, inline1, inline2), nil

		default:
			return "", fmt.Errorf("unsupported inline function in condition: %s", funcName)
		}

	default:
		return "", fmt.Errorf("unsupported condition expression: %T", expr)
	}
}

func (g *generator) generateVariableDeclaration(decl *ast.VariableDeclaration) (string, error) {
	code := ""
	for _, declarator := range decl.Declarations {
		varName := declarator.ID.Name

		// Check if this is an input.* function call
		if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
			funcName := g.extractFunctionName(callExpr.Callee)

			// Handle input functions
			if funcName == "input.float" || funcName == "input.int" ||
				funcName == "input.bool" || funcName == "input.string" ||
				funcName == "input.session" {
				// Already handled in first pass - skip code generation here
				continue
			}

			if funcName == "input.source" {
				// input.source(defval=close) means varName is an alias to close
				// Generate comment only - actual usage will reference source directly
				code += g.ind() + fmt.Sprintf("// %s = input.source() - using source directly\n", varName)
				continue
			}
		}

		// Determine variable type based on init expression
		varType := g.inferVariableType(declarator.Init)
		g.variables[varName] = varType
		g.varInits[varName] = declarator.Init // Store for constant resolution in extractTAArguments

		// Skip string variables (Series storage is float64 only)
		if varType == "string" {
			code += g.ind() + fmt.Sprintf("// %s = string variable (not implemented)\n", varName)
			continue
		}

		// Generate initialization from init expression
		if declarator.Init != nil {
			// ALL variables use same initialization path (ForwardSeriesBuffer paradigm)
			initCode, err := g.generateVariableInit(varName, declarator.Init)
			if err != nil {
				return "", err
			}
			code += initCode
		}
	}
	return code, nil
}

func (g *generator) inferVariableType(expr ast.Expression) string {
	if expr == nil {
		return "float64"
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok {
			if obj.Name == "syminfo" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					if prop.Name == "tickerid" {
						return "string"
					}
				}
			}
		}
		return "float64"
	case *ast.BinaryExpression:
		// Comparison operators produce bool
		if e.Operator == ">" || e.Operator == "<" || e.Operator == ">=" ||
			e.Operator == "<=" || e.Operator == "==" || e.Operator == "!=" {
			return "bool"
		}
		return "float64"
	case *ast.LogicalExpression:
		// and/or produce bool
		return "bool"
	case *ast.UnaryExpression:
		// Boolean negation produces bool
		if e.Operator == "not" || e.Operator == "!" {
			return "bool"
		}
		// Numeric unary preserves operand type
		return g.inferVariableType(e.Argument)
	case *ast.CallExpression:
		funcName := g.extractFunctionName(e.Callee)
		if funcName == "ta.crossover" || funcName == "ta.crossunder" {
			return "bool"
		}
		return "float64"
	case *ast.ConditionalExpression:
		// Ternary type depends on consequent/alternate
		return g.inferVariableType(e.Consequent)
	default:
		return "float64"
	}
}

func (g *generator) generateVariableInit(varName string, initExpr ast.Expression) (string, error) {
	// STEP 1: Detect nested TA calls and generate temp vars INLINE (same statement)
	// Example: rma(max(change(x), 0), 9) →
	//   1. Generate change_xxxSeries.Set()
	//   2. Generate max_yyySeries.Set()
	//   3. Generate rma using max_yyySeries reference
	nestedCalls := g.exprAnalyzer.FindNestedCalls(initExpr)

	tempVarCode := ""
	if len(nestedCalls) > 0 {
		// Process nested calls in REVERSE order (innermost first)
		// Example: rma(max(change(x), 0), 9) returns [rma, max, change]
		//   Must process change → max → rma so dependencies exist when referenced
		for i := len(nestedCalls) - 1; i >= 0; i-- {
			callInfo := nestedCalls[i]

			// Skip the outermost call (that's the main variable being generated)
			if callInfo.Call == initExpr {
				continue
			}

			// Create temp vars for:
			// 1. TA functions (ta.sma, ta.ema, etc.)
			// 2. Math functions that contain TA calls (e.g., max(change(x), 0))
			isTAFunction := g.taRegistry.IsSupported(callInfo.FuncName)
			containsNestedTA := false
			if !isTAFunction {
				// Check if this math function contains TA calls
				mathNestedCalls := g.exprAnalyzer.FindNestedCalls(callInfo.Call)
				for _, mathNested := range mathNestedCalls {
					if mathNested.Call != callInfo.Call && g.taRegistry.IsSupported(mathNested.FuncName) {
						containsNestedTA = true
						break
					}
				}
			}

			if !isTAFunction && !containsNestedTA {
				continue // Pure math function - inline OK
			}

			// Create temp var for this nested call
			tempVarName := g.tempVarMgr.GetOrCreate(callInfo)

			// Generate calculation code for temp var
			tempCode, err := g.generateVariableFromCall(tempVarName, callInfo.Call)
			if err != nil {
				return "", fmt.Errorf("failed to generate temp var %s: %w", tempVarName, err)
			}
			tempVarCode += tempCode
		}
	}

	// STEP 2: Process the main expression (extractSeriesExpression now uses temp var refs)
	switch expr := initExpr.(type) {
	case *ast.CallExpression:
		// Handle function calls like ta.sma(close, 20)
		mainCode, err := g.generateVariableFromCall(varName, expr)
		return tempVarCode + mainCode, err
	case *ast.ConditionalExpression:
		// Handle ternary: test ? consequent : alternate
		condCode, err := g.generateConditionExpression(expr.Test)
		if err != nil {
			return "", err
		}
		// If the test accesses a bool Series variable, add != 0 conversion
		condCode = g.addBoolConversionIfNeeded(expr.Test, condCode)

		// For consequent and alternate: generate as numeric expressions
		// Convert boolean literals to float: true→1.0, false→0.0
		consequentCode, err := g.generateNumericExpression(expr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateNumericExpression(expr.Alternate)
		if err != nil {
			return "", err
		}
		// Generate inline conditional with Series.Set() (ALL variables use Series)
		return g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return %s } else { return %s } }())\n",
			varName, condCode, consequentCode, alternateCode), nil
	case *ast.UnaryExpression:
		// Handle unary expressions: not x, -x, +x
		if expr.Operator == "not" || expr.Operator == "!" {
			// Boolean negation: not na(x) → convert boolean to float (1.0 or 0.0)
			operandCode, err := g.generateConditionExpression(expr.Argument)
			if err != nil {
				return "", err
			}
			// Convert boolean expression to float: true→1.0, false→0.0
			boolToFloatExpr := fmt.Sprintf("func() float64 { if !(%s) { return 1.0 } else { return 0.0 } }()", operandCode)
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, boolToFloatExpr), nil
		} else {
			// Numeric unary: -x, +x (get numeric value, not condition)
			operandCode, err := g.generateExpression(expr.Argument)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s(%s))\n", varName, expr.Operator, operandCode), nil
		}
	case *ast.Literal:
		// Simple literal assignment
		// Note: Pine Script doesn't have true constants for non-input literals
		// String literals assigned to variables are unusual and not typically used in series context
		// For session strings, use input.session() instead
		switch v := expr.Value.(type) {
		case float64:
			return g.ind() + fmt.Sprintf("%sSeries.Set(%.2f)\n", varName, v), nil
		case int:
			return g.ind() + fmt.Sprintf("%sSeries.Set(%.2f)\n", varName, float64(v)), nil
		case bool:
			val := 0.0
			if v {
				val = 1.0
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%.2f)\n", varName, val), nil
		case string:
			// String literals cannot be stored in numeric Series
			// Generate const declaration instead
			return g.ind() + fmt.Sprintf("// ERROR: string literal %q cannot be used in series context\n", v), nil
		default:
			return g.ind() + fmt.Sprintf("// ERROR: unsupported literal type\n"), nil
		}
	case *ast.Identifier:
		refName := expr.Name

		// Try builtin identifier resolution first
		if code, resolved := g.builtinHandler.TryResolveIdentifier(expr, g.inSecurityContext); resolved {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, code), nil
		}

		// Check if it's an input constant
		if _, isConstant := g.constants[refName]; isConstant {
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, refName), nil
		}

		// User-defined variable (ALL use Series)
		accessCode := fmt.Sprintf("%sSeries.GetCurrent()", refName)
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, accessCode), nil
	case *ast.MemberExpression:
		// Member access like strategy.long or close[1] (use Series.Set())
		memberCode := g.extractSeriesExpression(expr)

		// Strategy constants (strategy.long, strategy.short) need numeric conversion for Series
		if obj, ok := expr.Object.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				if prop, ok := expr.Property.(*ast.Identifier); ok {
					if prop.Name == "long" {
						return g.ind() + fmt.Sprintf("%sSeries.Set(1.0) // strategy.long\n", varName), nil
					} else if prop.Name == "short" {
						return g.ind() + fmt.Sprintf("%sSeries.Set(-1.0) // strategy.short\n", varName), nil
					}
				}
			}
		}

		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, memberCode), nil
	case *ast.BinaryExpression:
		// Binary expression like sma20[1] > ema50[1] or SMA + EMA
		/* In security context, need to generate temp series for operands */
		if g.inSecurityContext {
			return g.generateBinaryExpressionInSecurityContext(varName, expr)
		}

		// Normal context: compile-time evaluation
		binaryCode := g.extractSeriesExpression(expr)
		varType := g.inferVariableType(expr)
		if varType == "bool" {
			// Convert bool to float64 for Series storage
			return g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return 1.0 } else { return 0.0 } }())\n", varName, binaryCode), nil
		}
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, binaryCode), nil
	case *ast.LogicalExpression:
		// Logical expression like (a and b) or (c and d) → bool needs float64 conversion
		logicalCode, err := g.generateConditionExpression(expr)
		if err != nil {
			return "", err
		}
		// Convert bool to float64 for Series storage
		return g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return 1.0 } else { return 0.0 } }())\n", varName, logicalCode), nil
	default:
		return "", fmt.Errorf("unsupported init expression: %T", initExpr)
	}
}

func (g *generator) generateVariableFromCall(varName string, call *ast.CallExpression) (string, error) {
	funcName := g.extractFunctionName(call.Callee)

	// Try TA function registry first
	if g.taRegistry.IsSupported(funcName) {
		return g.taRegistry.GenerateInlineTA(g, varName, funcName, call)
	}

	// Handle math functions that need Series storage (have TA dependencies)
	mathHandler := NewMathFunctionHandler()
	if mathHandler.CanHandle(funcName) {
		return mathHandler.GenerateCode(g, varName, call)
	}

	switch funcName {
	case "request.security", "security":
		/* security(symbol, timeframe, expression) - runtime evaluation with cached context
		 * 1. Lookup security context from prefetch cache
		 * 2. Find matching bar index using timestamp alignment
		 * 3. Evaluate expression in security context at that bar
		 */
		if len(call.Arguments) < 3 {
			return g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN()) // security() missing arguments\n", varName), nil
		}

		/* Extract symbol and timeframe literals */
		symbolExpr := call.Arguments[0]
		timeframeExpr := call.Arguments[1]

		/* Get symbol string (tickerid → ctx.Symbol, literal → "BTCUSDT") */
		symbolStr := ""
		if id, ok := symbolExpr.(*ast.Identifier); ok {
			if id.Name == "tickerid" {
				symbolStr = "ctx.Symbol"
			} else {
				symbolStr = fmt.Sprintf("%q", id.Name)
			}
		} else if mem, ok := symbolExpr.(*ast.MemberExpression); ok {
			/* syminfo.tickerid */
			_ = mem
			symbolStr = "ctx.Symbol"
		} else if lit, ok := symbolExpr.(*ast.Literal); ok {
			if s, ok := lit.Value.(string); ok {
				symbolStr = fmt.Sprintf("%q", s)
			}
		}

		/* Get timeframe string */
		timeframeStr := ""
		if lit, ok := timeframeExpr.(*ast.Literal); ok {
			if s, ok := lit.Value.(string); ok {
				tf := strings.Trim(s, "'\"") /* Strip Pine string quotes */
				/* Normalize: D→1D, W→1W, M→1M */
				if tf == "D" {
					tf = "1D"
				} else if tf == "W" {
					tf = "1W"
				} else if tf == "M" {
					tf = "1M"
				}
				timeframeStr = tf /* Use normalized value directly without quoting yet */
			}
		}

		if symbolStr == "" || timeframeStr == "" {
			return g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN()) // security() unresolved symbol/timeframe\n", varName), nil
		}

		/* Build cache key using normalized timeframe */
		cacheKey := fmt.Sprintf("%%s:%s", timeframeStr)
		if symbolStr == "ctx.Symbol" {
			cacheKey = fmt.Sprintf("%s:%s", "%s", timeframeStr)
		} else {
			cacheKey = fmt.Sprintf("%s:%s", strings.Trim(symbolStr, `"`), timeframeStr)
		}

		/* Generate runtime lookup and evaluation */
		code := g.ind() + fmt.Sprintf("/* security(%s, %s, ...) */\n", symbolStr, timeframeStr)
		code += g.ind() + "{\n"
		g.indent++

		code += g.ind() + fmt.Sprintf("secKey := fmt.Sprintf(%q, %s)\n", cacheKey, symbolStr)
		code += g.ind() + "secCtx, secFound := securityContexts[secKey]\n"
		code += g.ind() + "if !secFound {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++

		/* Find bar index using timestamp */
		code += g.ind() + "secBarIdx := context.FindBarIndexByTimestamp(secCtx, ctx.Data[ctx.BarIndex].Time)\n"
		code += g.ind() + "if secBarIdx < 0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++

		/* Evaluate expression directly in security context (O(1) per-bar access) */
		exprArg := call.Arguments[2]

		/* Handle simple identifier access: close, open, high, low, volume */
		if ident, ok := exprArg.(*ast.Identifier); ok {
			fieldName := ident.Name
			switch fieldName {
			case "close":
				code += g.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Close)\n", varName)
			case "open":
				code += g.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Open)\n", varName)
			case "high":
				code += g.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].High)\n", varName)
			case "low":
				code += g.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Low)\n", varName)
			case "volume":
				code += g.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Volume)\n", varName)
			default:
				code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN()) // Unknown identifier: %s\n", varName, fieldName)
			}
		} else if callExpr, ok := exprArg.(*ast.CallExpression); ok {
			/* Handle TA function calls: ta.sma(close, 20), ta.ema(close, 10) */
			/* Create temporary series variable for inline TA result */
			secTempVar := fmt.Sprintf("secTmp_%s", varName)
			code += g.ind() + fmt.Sprintf("%sSeries := series.NewSeries(len(secCtx.Data))\n", secTempVar)

			/* Store original context, switch to security context */
			code += g.ind() + "origCtx := ctx\n"
			code += g.ind() + "ctx = secCtx\n"
			code += g.ind() + "ctx.BarIndex = secBarIdx\n"

			/* Set security context flag for inline TA */
			g.inSecurityContext = true

			/* Generate inline TA calculation */
			exprInit, err := g.generateVariableInit(secTempVar, callExpr)
			if err != nil {
				return "", fmt.Errorf("failed to generate security expression: %w", err)
			}
			code += exprInit

			/* Clear security context flag */
			g.inSecurityContext = false

			/* Restore original context */
			code += g.ind() + "ctx = origCtx\n"

			/* Extract value from temporary series */
			code += g.ind() + fmt.Sprintf("secValue := %sSeries.GetCurrent()\n", secTempVar)
			code += g.ind() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
		} else {
			/* Complex expression (BinaryExpression, ConditionalExpression, etc.) */
			/* Create temporary series variable for expression result */
			secTempVar := fmt.Sprintf("secTmp_%s", varName)
			code += g.ind() + fmt.Sprintf("%sSeries := series.NewSeries(len(secCtx.Data))\n", secTempVar)

			/* Store original context, switch to security context */
			code += g.ind() + "origCtx := ctx\n"
			code += g.ind() + "ctx = secCtx\n"
			code += g.ind() + "ctx.BarIndex = secBarIdx\n"

			/* Set security context flag */
			g.inSecurityContext = true

			/* Generate expression evaluation using full generateVariableInit */
			exprInit, err := g.generateVariableInit(secTempVar, exprArg)
			if err != nil {
				return "", fmt.Errorf("failed to generate security expression: %w", err)
			}
			code += exprInit

			/* Clear security context flag */
			g.inSecurityContext = false

			/* Restore original context */
			code += g.ind() + "ctx = origCtx\n"

			/* Extract value from temporary series */
			code += g.ind() + fmt.Sprintf("secValue := %sSeries.GetCurrent()\n", secTempVar)
			code += g.ind() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
		}

		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"

		return code, nil

	default:
		// Check if it's a math function
		if strings.HasPrefix(funcName, "math.") && g.mathHandler != nil {
			mathCode, err := g.mathHandler.GenerateMathCall(funcName, call.Arguments, g)
			if err != nil {
				return "", err
			}
			return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, mathCode), nil
		}
		return g.ind() + fmt.Sprintf("// %s = %s() - TODO: implement\n", varName, funcName), nil

	case "time":
		/* time(timeframe, session) - session filtering for intraday strategies
		 * Returns bar timestamp if within session, NaN otherwise
		 * Usage: entry_time = time(timeframe.period, "0950-1345")
		 * Check: is_entry_time = na(entry_time) ? false : true
		 */
		handler := NewTimeHandler(g.ind())
		return handler.HandleVariableInit(varName, call), nil
	}
}

/* generateInlineTA generates inline TA calculation for security() context */
func (g *generator) generateInlineTA(varName string, funcName string, call *ast.CallExpression) (string, error) {
	/* Normalize function name (handle both v4 and v5 syntax) */
	normalizedFunc := funcName
	if !strings.HasPrefix(funcName, "ta.") {
		normalizedFunc = "ta." + funcName
	}

	/* ATR special case: requires 1 argument (period only) */
	if normalizedFunc == "ta.atr" {
		if len(call.Arguments) < 1 {
			return "", fmt.Errorf("ta.atr requires 1 argument (period)")
		}
		periodArg, ok := call.Arguments[0].(*ast.Literal)
		if !ok {
			return "", fmt.Errorf("ta.atr period must be literal")
		}
		// Handle both int and float64 literals
		var period int
		switch v := periodArg.Value.(type) {
		case float64:
			period = int(v)
		case int:
			period = v
		default:
			return "", fmt.Errorf("ta.atr period must be numeric")
		}
		return g.generateInlineATR(varName, period)
	}

	/* Extract source and period arguments */
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceExpr := g.extractSeriesExpression(call.Arguments[0])

	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessGen := CreateAccessGenerator(sourceInfo)

	periodArg, ok := call.Arguments[1].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("%s period must be literal", funcName)
	}

	// Handle both int and float64 literals
	var period int
	switch v := periodArg.Value.(type) {
	case float64:
		period = int(v)
	case int:
		period = v
	default:
		return "", fmt.Errorf("%s period must be numeric", funcName)
	}

	// Use TAIndicatorBuilder for all indicators
	needsNaN := sourceInfo.IsSeriesVariable()

	var code string

	switch normalizedFunc {
	case "ta.sma":
		builder := NewTAIndicatorBuilder("ta.sma", varName, period, accessGen, needsNaN)
		builder.WithAccumulator(NewSumAccumulator())
		code = g.indentCode(builder.Build())

	case "ta.ema":
		builder := NewTAIndicatorBuilder("ta.ema", varName, period, accessGen, needsNaN)
		code = g.indentCode(builder.BuildEMA())

	case "ta.stdev":
		builder := NewTAIndicatorBuilder("ta.stdev", varName, period, accessGen, needsNaN)
		code = g.indentCode(builder.BuildSTDEV())

	default:
		return "", fmt.Errorf("inline TA not implemented for %s", funcName)
	}

	return code, nil
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
	code += g.ind() + "hl := ctx.Data[ctx.BarIndex].High - ctx.Data[ctx.BarIndex].Low\n"
	code += g.ind() + "hc := math.Abs(ctx.Data[ctx.BarIndex].High - ctx.Data[ctx.BarIndex-1].Close)\n"
	code += g.ind() + "lc := math.Abs(ctx.Data[ctx.BarIndex].Low - ctx.Data[ctx.BarIndex-1].Close)\n"
	code += g.ind() + "tr := math.Max(hl, math.Max(hc, lc))\n"

	/* RMA smoothing of TR */
	code += g.ind() + fmt.Sprintf("if ctx.BarIndex < %d {\n", period)
	g.indent++
	/* Warmup: use SMA for first period bars */
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
	// Handle strategy.long, strategy.short
	if mem, ok := expr.(*ast.MemberExpression); ok {
		if prop, ok := mem.Property.(*ast.Identifier); ok {
			switch prop.Name {
			case "long":
				return "strategy.Long"
			case "short":
				return "strategy.Short"
			}
		}
	}
	return "strategy.Long"
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

		// Try builtin member expression resolution (close[1], strategy.position_avg_price, etc.)
		if code, resolved := g.builtinHandler.TryResolveMemberExpression(e, false); resolved {
			return code
		}

		// Check for built-in namespaces like timeframe.* and syminfo.*
		if obj, ok := e.Object.(*ast.Identifier); ok {
			varName := obj.Name

			if varName == "syminfo" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					switch prop.Name {
					case "tickerid":
						return "syminfo_tickerid"
					}
				}
			}

			// Handle timeframe.* built-ins
			if varName == "timeframe" {
				if prop, ok := e.Property.(*ast.Identifier); ok {
					switch prop.Name {
					case "ismonthly":
						return "ctx.IsMonthly"
					case "isdaily":
						return "ctx.IsDaily"
					case "isweekly":
						return "ctx.IsWeekly"
					case "period":
						return "ctx.Timeframe"
					}
				}
			}

			// Handle series subscript with variable offset
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
			return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
		}

		return g.extractMemberName(e)
	case *ast.Identifier:
		// Check if it's an input constant
		if _, isConstant := g.constants[e.Name]; isConstant {
			return e.Name
		}

		// Try builtin identifier resolution first
		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, g.inSecurityContext); resolved {
			return code
		}

		// User-defined variables use Series storage (ForwardSeriesBuffer paradigm)
		return fmt.Sprintf("%sSeries.GetCurrent()", e.Name)
	case *ast.Literal:
		// Numeric literal
		switch v := e.Value.(type) {
		case float64:
			return fmt.Sprintf("%.2f", v)
		case int:
			return fmt.Sprintf("%d", v)
		}
	case *ast.BinaryExpression:
		// Arithmetic expression like sma20 * 1.02
		left := g.extractSeriesExpression(e.Left)
		right := g.extractSeriesExpression(e.Right)
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right)
	case *ast.UnaryExpression:
		// Unary expression like -1, +x
		operand := g.extractSeriesExpression(e.Argument)
		op := e.Operator
		if op == "not" {
			op = "!"
		}
		return fmt.Sprintf("%s%s", op, operand)
	case *ast.CallExpression:
		// Function call like math.pow(x, y) or ta.sma(close, 20)
		funcName := g.extractFunctionName(e.Callee)

		// PRIORITY 1: Check if temp var exists (even for math functions)
		// This handles cases where math functions have TA dependencies:
		// max(change(x), 0) needs temp var because change() is TA
		existingVar := g.tempVarMgr.GetVarNameForCall(e)
		if existingVar != "" {
			// Temp var already generated, use it
			return fmt.Sprintf("%sSeries.GetCurrent()", existingVar)
		}

		// PRIORITY 2: Try inline math (only if no TA dependencies)
		// Pure math functions like max(2, 3) or abs(-5) can be inlined
		if (strings.HasPrefix(funcName, "math.") ||
			funcName == "max" || funcName == "min" || funcName == "abs" ||
			funcName == "sqrt" || funcName == "floor" || funcName == "ceil" ||
			funcName == "round" || funcName == "log" || funcName == "exp") && g.mathHandler != nil {
			mathCode, err := g.mathHandler.GenerateMathCall(funcName, e.Arguments, g)
			if err != nil {
				// Return error placeholder
				return "0.0"
			}
			return mathCode
		}

		// PRIORITY 3: Legacy fallback for TA functions
		// Assume function result stored in series variable
		// (This will fail if variable doesn't exist - needs temp var generation)
		varName := strings.ReplaceAll(funcName, ".", "_")
		return fmt.Sprintf("%sSeries.GetCurrent()", varName)
	}
	return "0.0"
}

func (g *generator) convertSeriesAccessToPrev(seriesCode string) string {
	// Convert current bar access to previous bar access
	// bar.Close → ctx.Data[i-1].Close
	// sma20Series.Get(0) → sma20Series.Get(1)

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

	// For non-Series user variables, return 0.0 (shouldn't happen in crossover with Series)
	return "0.0"
}

func (g *generator) generateLiteral(lit *ast.Literal) (string, error) {
	switch v := lit.Value.(type) {
	case float64:
		return g.ind() + fmt.Sprintf("%.2f\n", v), nil
	case string:
		return g.ind() + fmt.Sprintf("%q\n", v), nil
	case bool:
		return g.ind() + fmt.Sprintf("%t\n", v), nil
	default:
		jsonBytes, _ := json.Marshal(v)
		return g.ind() + string(jsonBytes) + "\n", nil
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
	code += g.ind() + fmt.Sprintf("strat.Call(%q, 10000)\n\n", g.strategyName)
	code += g.ind() + "for i := 0; i < len(ctx.Data); i++ {\n"
	g.indent++
	code += g.ind() + "ctx.BarIndex = i\n"
	code += g.ind() + "strat.OnBarUpdate(i, ctx.Data[i].Open, ctx.Data[i].Time)\n"
	code += g.ind() + "// Strategy logic placeholder\n"
	g.indent--
	code += g.ind() + "}\n"
	return code
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
// RMA uses alpha = 1/period (vs EMA's 2/(period+1))
func (g *generator) generateRMA(varName string, period int, accessor AccessGenerator, needsNaN bool) (string, error) {
	builder := NewTAIndicatorBuilder("ta.rma", varName, period, accessor, needsNaN)
	return g.indentCode(builder.BuildRMA()), nil
}

// generateRSI generates inline RSI (Relative Strength Index) calculation
// TODO: Implement RSI inline generation
func (g *generator) generateRSI(varName string, period int, accessor AccessGenerator, needsNaN bool) (string, error) {
	return "", fmt.Errorf("ta.rsi inline generation not yet implemented")
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
		// Fallback for complex expressions
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

// generatePivot generates inline pivot high/low detection
// TODO: Implement pivot inline generation
func (g *generator) generatePivot(varName string, call *ast.CallExpression, isHigh bool) (string, error) {
	return "", fmt.Errorf("ta.pivot inline generation not yet implemented")
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
		// Check if object is a CallExpression: pivothigh()[1]
		if nestedCall, ok := e.Object.(*ast.CallExpression); ok {
			nestedFuncName := g.extractFunctionName(nestedCall.Callee)
			// Use funcName-based naming to match extractSeriesExpression
			tempVarName := strings.ReplaceAll(nestedFuncName, ".", "_")

			// Register nested variable for Series initialization
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

/* preAnalyzeSecurityCalls scans AST for ALL expressions with nested TA calls,
 * registers temp vars BEFORE declaration phase to prevent "undefined: ta_sma_XXX" errors.
 *
 * CRITICAL: Must run AFTER first pass (collects constants) but BEFORE code generation.
 *
 * Bug Fix #1: security(syminfo.tickerid, 'D', sma(close, 20)) generates inline TA code
 * that references ta_sma_20_XXXSeries, but if temp var not pre-registered, declaration
 * phase misses it → compile error.
 *
 * Bug Fix #2: sma(close, 50) > sma(close, 200) in BinaryExpression also needs temp vars
 * for both sma() calls to avoid "undefined: ta_sma_XXX" errors.
 *
 * FILTER: Only create temp vars for TA functions (ta.sma, ta.ema, etc.), not math functions.
 */
func (g *generator) preAnalyzeSecurityCalls(program *ast.Program) {
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if declarator.Init != nil {
					// Scan ALL expressions for nested TA calls (not just security())
					nestedCalls := g.exprAnalyzer.FindNestedCalls(declarator.Init)
					// Register temp vars in REVERSE order (innermost first)
					for i := len(nestedCalls) - 1; i >= 0; i-- {
						callInfo := nestedCalls[i]

						// Create temp vars for:
						// 1. TA functions (ta.sma, ta.ema, etc.)
						isTAFunction := g.taRegistry.IsSupported(callInfo.FuncName)

						// 2. Math functions that contain TA calls (e.g., max(change(x), 0))
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

						if isTAFunction || containsNestedTA {
							g.tempVarMgr.GetOrCreate(callInfo)
						}
					}
				}
			}
		}
	}
}

// extractConstValue parses "const varName = VALUE" to extract VALUE
func extractConstValue(code string) interface{} {
	// Parse: "const bblenght = 46\n"
	var varName string
	var floatVal float64
	var intVal int

	if _, err := fmt.Sscanf(code, "const %s = %f", &varName, &floatVal); err == nil {
		return floatVal
	}
	if _, err := fmt.Sscanf(code, "const %s = %d", &varName, &intVal); err == nil {
		return intVal
	}
	return nil
}
