package codegen

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/borisquantlab/pinescript-go/ast"
)

/* StrategyCode holds generated Go code for strategy execution */
type StrategyCode struct {
	FunctionBody       string // executeStrategy() function body
	StrategyName       string // Pine Script strategy name
	NeedsSeriesPreCalc bool   // Whether TA pre-calculation imports are needed
}

/* GenerateStrategyCodeFromAST converts parsed Pine ESTree to Go runtime code */
func GenerateStrategyCodeFromAST(program *ast.Program) (*StrategyCode, error) {
	gen := &generator{
		imports:      make(map[string]bool),
		variables:    make(map[string]string),
		strategyName: "Generated Strategy",
	}

	body, err := gen.generateProgram(program)
	if err != nil {
		return nil, err
	}

	code := &StrategyCode{
		FunctionBody:       body,
		StrategyName:       gen.strategyName,
		NeedsSeriesPreCalc: gen.needsSeriesPreCalc,
	}

	return code, nil
}

type generator struct {
	imports            map[string]bool
	variables          map[string]string
	plots              []string        // Track plot variables
	strategyName       string          // Strategy name from indicator() or strategy()
	indent             int
	needsSeriesPreCalc bool                 // Flag if we need series pre-calculation
	taFunctions        []taFunctionCall     // List of TA function calls to pre-calculate
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

	// First pass: collect variables, analyze Series requirements, extract strategy name
	for _, stmt := range program.Body {
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
				varType := g.inferVariableType(declarator.Init)
				g.variables[varName] = varType
			}
		}
	}

	// Second pass: No longer needed (ALL variables use Series storage)
	// Kept for future optimizations if needed

	// Third pass: collect TA function calls for pre-calculation
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
					funcName := g.extractFunctionName(callExpr.Callee)
					if funcName == "ta.sma" || funcName == "ta.ema" || funcName == "ta.rma" ||
						funcName == "ta.rsi" || funcName == "ta.atr" || funcName == "ta.stdev" ||
						funcName == "ta.change" || funcName == "ta.pivothigh" || funcName == "ta.pivotlow" {
						g.needsSeriesPreCalc = true
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

	// Declare ALL variables as Series (ForwardSeriesBuffer paradigm)
	if len(g.variables) > 0 {
		code += g.ind() + "// ALL variables use Series storage (ForwardSeriesBuffer paradigm)\n"
		for varName := range g.variables {
			code += g.ind() + fmt.Sprintf("var %sSeries *series.Series\n", varName)
		}
		code += "\n"

		// Initialize ALL Series before bar loop
		code += g.ind() + "// Initialize Series storage\n"
		for varName := range g.variables {
			code += g.ind() + fmt.Sprintf("%sSeries = series.NewSeries(len(ctx.Data))\n", varName)
		}
		code += "\n"
	}

	// Pre-calculate TA functions using runtime library
	if g.needsSeriesPreCalc && len(g.taFunctions) > 0 {
		code += g.ind() + "// Pre-calculate TA functions using runtime library\n"
		
		// Extract source series (close, high, low, open, etc.)
		sourcesNeeded := make(map[string]bool)
		for _, taFunc := range g.taFunctions {
			if len(taFunc.args) > 0 {
				sourceName := g.extractArgIdentifier(taFunc.args[0])
				if sourceName != "" {
					sourcesNeeded[sourceName] = true
				}
			}
		}
		
		for source := range sourcesNeeded {
			code += g.ind() + fmt.Sprintf("%sSeries := make([]float64, len(ctx.Data))\n", strings.ToLower(source))
			code += g.ind() + "for i := range ctx.Data {\n"
			g.indent++
			code += g.ind() + fmt.Sprintf("%sSeries[i] = ctx.Data[i].%s\n", strings.ToLower(source), source)
			g.indent--
			code += g.ind() + "}\n"
		}
		code += "\n"
		
		// Call runtime TA functions
		for _, taFunc := range g.taFunctions {
			funcCode, err := g.generateTAPreCalc(taFunc)
			if err != nil {
				return "", err
			}
			code += funcCode
		}
		code += "\n"
	}

	// Bar loop for strategy execution
	code += g.ind() + "for i := 0; i < len(ctx.Data); i++ {\n"
	g.indent++
	code += g.ind() + "ctx.BarIndex = i\n"
	code += g.ind() + "bar := ctx.Data[i]\n"
	code += g.ind() + "strat.OnBarUpdate(i, bar.Open, bar.Time)\n\n"

	// Generate statements inside bar loop
	for _, stmt := range program.Body {
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
		code += g.ind() + fmt.Sprintf("if i < len(ctx.Data)-1 { %sSeries.Next() }\n", varName)
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
		if len(call.Arguments) > 0 {
			// Get plot value (first argument) - could be identifier or member expression like sma20[0]
			plotVar := ""
			switch arg := call.Arguments[0].(type) {
			case *ast.Identifier:
				plotVar = arg.Name
			case *ast.MemberExpression:
				// Handle sma20[0] → extract "sma20"
				if id, ok := arg.Object.(*ast.Identifier); ok {
					plotVar = id.Name
				}
			}

			plotTitle := plotVar // Default title

			// Check for title in second argument (object expression)
			if len(call.Arguments) > 1 {
				if obj, ok := call.Arguments[1].(*ast.ObjectExpression); ok {
					for _, prop := range obj.Properties {
						if keyID, ok := prop.Key.(*ast.Identifier); ok && keyID.Name == "title" {
							if valLit, ok := prop.Value.(*ast.Literal); ok {
								if title, ok := valLit.Value.(string); ok {
									plotTitle = title
								}
							}
						}
					}
				}
			}

			if plotVar != "" {
				// ALL variables use Series storage
				plotExpr := plotVar + "Series.Get(0)"
				code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, nil)\n", plotTitle, plotExpr)
			}
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
	case "ta.stdev", "ta.change", "ta.pivothigh", "ta.pivotlow":
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
	for _, stmt := range ifStmt.Consequent {
		stmtCode, err := g.generateStatement(stmt)
		if err != nil {
			return "", err
		}
		code += stmtCode
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
		// ALL variables use Series storage
		varName := e.Name
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

	default:
		return "", fmt.Errorf("unsupported condition expression: %T", expr)
	}
}

func (g *generator) generateVariableDeclaration(decl *ast.VariableDeclaration) (string, error) {
	code := ""
	for _, declarator := range decl.Declarations {
		varName := declarator.ID.Name

		// Determine variable type based on init expression
		varType := g.inferVariableType(declarator.Init)
		g.variables[varName] = varType

		// Check if this is a TA function call
		isTAFunc := false
		if callExpr, ok := declarator.Init.(*ast.CallExpression); ok {
			funcName := g.extractFunctionName(callExpr.Callee)
			if funcName == "ta.sma" || funcName == "ta.ema" || funcName == "ta.rma" || 
			   funcName == "ta.rsi" || funcName == "ta.atr" || funcName == "ta.stdev" ||
			   funcName == "ta.change" || funcName == "ta.pivothigh" || funcName == "ta.pivotlow" {
				isTAFunc = true
			}
		}

		// Generate initialization from init expression
		if declarator.Init != nil {
			// For TA functions that are pre-calculated
			if isTAFunc {
				// Store in Series (ALL variables use Series)
				code += g.ind() + fmt.Sprintf("%sSeries.Set(%sArray[i])\n", varName, varName)
			} else {
				// Regular variable initialization (use Series.Set())
				initCode, err := g.generateVariableInit(varName, declarator.Init)
				if err != nil {
					return "", err
				}
				code += initCode
			}
		}
	}
	return code, nil
}

func (g *generator) inferVariableType(expr ast.Expression) string {
	if expr == nil {
		return "float64"
	}

	switch e := expr.(type) {
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
	switch expr := initExpr.(type) {
	case *ast.CallExpression:
		// Handle function calls like ta.sma(close, 20)
		return g.generateVariableFromCall(varName, expr)
	case *ast.ConditionalExpression:
		// Handle ternary: test ? consequent : alternate
		condCode, err := g.generateConditionExpression(expr.Test)
		if err != nil {
			return "", err
		}
		// If the test accesses a bool Series variable, add != 0 conversion
		condCode = g.addBoolConversionIfNeeded(expr.Test, condCode)
		
		consequentCode, err := g.generateConditionExpression(expr.Consequent)
		if err != nil {
			return "", err
		}
		alternateCode, err := g.generateConditionExpression(expr.Alternate)
		if err != nil {
			return "", err
		}
		// Generate inline conditional with Series.Set() (ALL variables use Series)
		return g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s { return %s } else { return %s } }())\n",
			varName, condCode, consequentCode, alternateCode), nil
	case *ast.Literal:
		// Simple literal assignment (use Series.Set())
		return g.ind() + fmt.Sprintf("%sSeries.Set(%.2f)\n", varName, expr.Value), nil
	case *ast.Identifier:
		// Reference to another variable (ALL use Series)
		refName := expr.Name
		accessCode := fmt.Sprintf("%sSeries.GetCurrent()", refName)
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, accessCode), nil
	case *ast.MemberExpression:
		// Member access like strategy.long or close[1] (use Series.Set())
		memberCode := g.extractSeriesExpression(expr)
		return g.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, memberCode), nil
	case *ast.BinaryExpression:
		// Binary expression like sma20[1] > ema50[1] → bool needs float64 conversion
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

	switch funcName {
	case "ta.sma", "ta.ema", "ta.rma", "ta.rsi", "ta.atr", "ta.stdev", "ta.change", "ta.pivothigh", "ta.pivotlow":
		// TA functions - registered in third pass, just return empty here
		return "", nil

	case "ta.crossover":
		// ta.crossover(series1, series2) - series1 crosses ABOVE series2
		if len(call.Arguments) < 2 {
			return "", fmt.Errorf("ta.crossover requires 2 arguments")
		}

		series1 := g.extractSeriesExpression(call.Arguments[0])
		series2 := g.extractSeriesExpression(call.Arguments[1])

		// Need previous values for both series
		prev1Var := varName + "_prev1"
		prev2Var := varName + "_prev2"

		code := g.ind() + fmt.Sprintf("// Crossover: %s crosses above %s\n", series1, series2)
		code += g.ind() + "if i > 0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%s := %s\n", prev1Var, g.convertSeriesAccessToPrev(series1))
		code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, g.convertSeriesAccessToPrev(series2))
		code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s > %s && %s <= %s { return 1.0 } else { return 0.0 } }())\n",
			varName, series1, series2, prev1Var, prev2Var)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
		g.indent--
		code += g.ind() + "}\n"

		return code, nil

	case "ta.crossunder":
		// ta.crossunder(series1, series2) - series1 crosses BELOW series2
		if len(call.Arguments) < 2 {
			return "", fmt.Errorf("ta.crossunder requires 2 arguments")
		}

		series1 := g.extractSeriesExpression(call.Arguments[0])
		series2 := g.extractSeriesExpression(call.Arguments[1])

		// Need previous values for both series
		prev1Var := varName + "_prev1"
		prev2Var := varName + "_prev2"

		code := g.ind() + fmt.Sprintf("// Crossunder: %s crosses below %s\n", series1, series2)
		code += g.ind() + "if i > 0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%s := %s\n", prev1Var, g.convertSeriesAccessToPrev(series1))
		code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, g.convertSeriesAccessToPrev(series2))
		code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { if %s < %s && %s >= %s { return 1.0 } else { return 0.0 } }())\n",
			varName, series1, series2, prev1Var, prev2Var)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
		g.indent--
		code += g.ind() + "}\n"

		return code, nil

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
		code += g.ind() + "secCtx.BarIndex = secBarIdx\n"
		
		/* Evaluate expression in security context by generating variable init */
		/* Create temporary series variable for expression result */
		exprArg := call.Arguments[2]
		secTempVar := fmt.Sprintf("secTmp_%s", varName)
		
		/* Generate series declaration for temporary variable */
		code += g.ind() + fmt.Sprintf("%sSeries := series.NewSeries(1000)\n", secTempVar)
		
		/* Store original context reference, temporarily use secCtx */
		code += g.ind() + "origCtx := ctx\n"
		code += g.ind() + "ctx = secCtx\n"
		
		/* Generate the expression evaluation using normal code generation */
		exprInit, err := g.generateVariableInit(secTempVar, exprArg)
		if err != nil {
			return "", fmt.Errorf("failed to generate security expression: %w", err)
		}
		code += exprInit
		
		/* Restore original context */
		code += g.ind() + "ctx = origCtx\n"
		
		/* Extract value from temporary series */
		code += g.ind() + fmt.Sprintf("secValue := %sSeries.GetCurrent()\n", secTempVar)
		code += g.ind() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
		
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		
		return code, nil

	default:
		return g.ind() + fmt.Sprintf("// %s = %s() - TODO: implement\n", varName, funcName), nil
	}
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
	
	// First argument is positional title (simple case)
	if lit, ok := args[0].(*ast.Literal); ok {
		if name, ok := lit.Value.(string); ok {
			return name
		}
	}
	
	// Search for 'title=' named parameter in ObjectExpression
	for _, arg := range args {
		if obj, ok := arg.(*ast.ObjectExpression); ok {
			for _, prop := range obj.Properties {
				// Check if key is 'title'
				keyName := ""
				if id, ok := prop.Key.(*ast.Identifier); ok {
					keyName = id.Name
				} else if lit, ok := prop.Key.(*ast.Literal); ok {
					if name, ok := lit.Value.(string); ok {
						keyName = name
					}
				}
				
				if keyName == "title" {
					// Extract value
					if lit, ok := prop.Value.(*ast.Literal); ok {
						if name, ok := lit.Value.(string); ok {
							return name
						}
					}
				}
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
		// Handle series subscript like close[0], close[1], sma20[0], sma20[1]
		if obj, ok := e.Object.(*ast.Identifier); ok {
			varName := obj.Name

			// Extract offset from subscript
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

			// Check if it's a Pine built-in series
			switch varName {
			case "close":
				if offset == 0 {
					return "bar.Close"
				}
				// Historical builtin: use ctx.Data[i-offset]
				return fmt.Sprintf("ctx.Data[i-%d].Close", offset)
			case "open":
				if offset == 0 {
					return "bar.Open"
				}
				return fmt.Sprintf("ctx.Data[i-%d].Open", offset)
			case "high":
				if offset == 0 {
					return "bar.High"
				}
				return fmt.Sprintf("ctx.Data[i-%d].High", offset)
			case "low":
				if offset == 0 {
					return "bar.Low"
				}
				return fmt.Sprintf("ctx.Data[i-%d].Low", offset)
			case "volume":
				if offset == 0 {
					return "bar.Volume"
				}
				return fmt.Sprintf("ctx.Data[i-%d].Volume", offset)
			default:
				// User-defined variable (ALL use Series storage)
				return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
			}
		}
		return g.extractMemberName(e)
	case *ast.Identifier:
		// User-defined variable (ALL use Series storage)
		varName := e.Name
		return fmt.Sprintf("%sSeries.GetCurrent()", varName)
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

/* generateTAPreCalc generates runtime library call for TA function */
func (g *generator) generateTAPreCalc(taFunc taFunctionCall) (string, error) {
	code := ""
	
	// Use "Array" suffix for TA function results to distinguish from Series
	arrayName := taFunc.varName + "Array"
	
	switch taFunc.funcName {
	case "ta.sma", "ta.ema", "ta.rma":
		// ta.sma(source, period)
		if len(taFunc.args) < 2 {
			return "", fmt.Errorf("%s requires 2 arguments", taFunc.funcName)
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		period := g.extractArgLiteral(taFunc.args[1])
		
		funcMap := map[string]string{"ta.sma": "Sma", "ta.ema": "Ema", "ta.rma": "Rma"}
		funcName := funcMap[taFunc.funcName]
		
		code += g.ind() + fmt.Sprintf("%s := ta.%s(%sSeries, %d)\n", 
			arrayName, funcName, source, period)
		
	case "ta.rsi":
		// ta.rsi(source, period)
		if len(taFunc.args) < 2 {
			return "", fmt.Errorf("ta.rsi requires 2 arguments")
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		period := g.extractArgLiteral(taFunc.args[1])
		
		code += g.ind() + fmt.Sprintf("%s := ta.Rsi(%sSeries, %d)\n", 
			arrayName, source, period)
		
	case "ta.atr":
		// ta.atr(period)
		if len(taFunc.args) < 1 {
			return "", fmt.Errorf("ta.atr requires 1 argument")
		}
		
		period := g.extractArgLiteral(taFunc.args[0])
		
		code += g.ind() + fmt.Sprintf("%s := ta.Atr(highSeries, lowSeries, closeSeries, %d)\n", 
			arrayName, period)
		
	case "ta.stdev":
		// ta.stdev(source, period)
		if len(taFunc.args) < 2 {
			return "", fmt.Errorf("ta.stdev requires 2 arguments")
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		period := g.extractArgLiteral(taFunc.args[1])
		
		code += g.ind() + fmt.Sprintf("%s := ta.Stdev(%sSeries, %d)\n", 
			arrayName, source, period)
		
	case "ta.change":
		// ta.change(source)
		if len(taFunc.args) < 1 {
			return "", fmt.Errorf("ta.change requires 1 argument")
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		
		code += g.ind() + fmt.Sprintf("%s := ta.Change(%sSeries)\n", 
			arrayName, source)
		
	case "ta.pivothigh":
		// ta.pivothigh(source, leftBars, rightBars)
		if len(taFunc.args) < 3 {
			return "", fmt.Errorf("ta.pivothigh requires 3 arguments")
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		leftBars := g.extractArgLiteral(taFunc.args[1])
		rightBars := g.extractArgLiteral(taFunc.args[2])
		
		code += g.ind() + fmt.Sprintf("%s := ta.Pivothigh(%sSeries, %d, %d)\n", 
			arrayName, source, leftBars, rightBars)
		
	case "ta.pivotlow":
		// ta.pivotlow(source, leftBars, rightBars)
		if len(taFunc.args) < 3 {
			return "", fmt.Errorf("ta.pivotlow requires 3 arguments")
		}
		
		source := strings.ToLower(g.extractArgIdentifier(taFunc.args[0]))
		leftBars := g.extractArgLiteral(taFunc.args[1])
		rightBars := g.extractArgLiteral(taFunc.args[2])
		
		code += g.ind() + fmt.Sprintf("%s := ta.Pivotlow(%sSeries, %d, %d)\n", 
			arrayName, source, leftBars, rightBars)
		
	default:
		return "", fmt.Errorf("unsupported TA function: %s", taFunc.funcName)
	}
	
	return code, nil
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
