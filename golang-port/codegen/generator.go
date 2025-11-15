package codegen

import (
	"encoding/json"
	"fmt"

	"github.com/borisquantlab/pinescript-go/ast"
)

/* GenerateStrategyCodeFromAST converts parsed Pine ESTree to Go runtime code */
func GenerateStrategyCodeFromAST(program *ast.Program) (*StrategyCode, error) {
	gen := &generator{
		imports:   make(map[string]bool),
		variables: make(map[string]string),
	}

	body, err := gen.generateProgram(program)
	if err != nil {
		return nil, err
	}

	code := &StrategyCode{
		FunctionBody: body,
	}

	return code, nil
}

type generator struct {
	imports   map[string]bool
	variables map[string]string
	plots     []string // Track plot variables
	indent    int
}

func (g *generator) generateProgram(program *ast.Program) (string, error) {
	if program == nil || len(program.Body) == 0 {
		return g.generatePlaceholder(), nil
	}

	// First pass: collect variables
	for _, stmt := range program.Body {
		if varDecl, ok := stmt.(*ast.VariableDeclaration); ok {
			for _, declarator := range varDecl.Declarations {
				g.variables[declarator.ID.Name] = "float64"
			}
		}
	}

	code := ""
	
	// Initialize strategy
	code += g.ind() + "strat.Call(\"Generated Strategy\", 10000)\n\n"
	
	// Declare series variables (will be updated per bar)
	if len(g.variables) > 0 {
		code += g.ind() + "// Series variables\n"
		for varName := range g.variables {
			code += g.ind() + fmt.Sprintf("var %s float64\n", varName)
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
				code += g.ind() + fmt.Sprintf("collector.Add(%q, bar.Time, %s, nil)\n", plotTitle, plotVar)
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
		// Crossover functions - TODO: implement
		code += g.ind() + fmt.Sprintf("// %s() - TODO: implement\n", funcName)
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

func (g *generator) generateConditionExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
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
		// Extract variable name
		// MemberExpression can be: close[0], sma20[0], bar.Close, etc.
		if obj, ok := e.Object.(*ast.Identifier); ok {
			// Check if it's a Pine built-in series variable
			switch obj.Name {
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
			default:
				// It's a user-defined variable like sma20
				return obj.Name, nil
			}
		}
		return "bar.Close", nil
		
	case *ast.Identifier:
		return e.Name, nil
		
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
		g.variables[varName] = "float64" // Assume float64 for now
		
		// Generate initialization from init expression
		if declarator.Init != nil {
			initCode, err := g.generateVariableInit(varName, declarator.Init)
			if err != nil {
				return "", err
			}
			code += initCode
		}
	}
	return code, nil
}

func (g *generator) generateVariableInit(varName string, initExpr ast.Expression) (string, error) {
	switch expr := initExpr.(type) {
	case *ast.CallExpression:
		// Handle function calls like ta.sma(close, 20)
		return g.generateVariableFromCall(varName, expr)
	case *ast.Literal:
		// Simple literal assignment
		return g.ind() + fmt.Sprintf("%s = %.2f\n", varName, expr.Value), nil
	case *ast.Identifier:
		// Reference to another variable
		return g.ind() + fmt.Sprintf("%s = %s\n", varName, expr.Name), nil
	case *ast.MemberExpression:
		// Member access like strategy.long
		memberName := g.extractMemberName(expr)
		return g.ind() + fmt.Sprintf("%s = %s\n", varName, memberName), nil
	default:
		return "", fmt.Errorf("unsupported init expression: %T", initExpr)
	}
}

func (g *generator) generateVariableFromCall(varName string, call *ast.CallExpression) (string, error) {
	funcName := g.extractFunctionName(call.Callee)
	
	switch funcName {
	case "ta.sma":
		// ta.sma(source, length)
		if len(call.Arguments) < 2 {
			return "", fmt.Errorf("ta.sma requires 2 arguments")
		}
		
		// Get source and length
		source := g.extractArgIdentifier(call.Arguments[0])
		lengthVal := g.extractArgLiteral(call.Arguments[1])
		
		code := g.ind() + fmt.Sprintf("// Calculate SMA for %s\n", varName)
		code += g.ind() + fmt.Sprintf("if i >= %d - 1 {\n", lengthVal)
		g.indent++
		code += g.ind() + "sum := 0.0\n"
		code += g.ind() + fmt.Sprintf("for j := 0; j < %d; j++ {\n", lengthVal)
		g.indent++
		code += g.ind() + fmt.Sprintf("sum += ctx.Data[i-j].%s\n", source)
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%s = sum / %.1f\n", varName, float64(lengthVal))
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%s = 0.0 // NaN warmup\n", varName)
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

func (g *generator) generatePlaceholder() string {
	code := g.ind() + "// Strategy code will be generated here\n"
	code += g.ind() + "strat.Call(\"Generated Strategy\", 10000)\n\n"
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
