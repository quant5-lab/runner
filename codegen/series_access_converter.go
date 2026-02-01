package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// CallVarLookup resolves temp variable name for CallExpression (decoupled from TempVariableManager)
// Returns empty string if no temp var exists for the call
type CallVarLookup func(*ast.CallExpression) string

// SeriesAccessConverter transforms AST expressions by converting series variable identifiers
// to their historical access form (e.g., "sum" → "sumSeries.Get(offset)")
// Responsibility: AST transformation for type-aware series access
type SeriesAccessConverter struct {
	symbolTable   SymbolTable
	offset        string
	lookupCallVar CallVarLookup
}

// NewSeriesAccessConverter creates a converter with symbol type information
func NewSeriesAccessConverter(symbolTable SymbolTable, offset string, lookupCallVar CallVarLookup) *SeriesAccessConverter {
	return &SeriesAccessConverter{
		symbolTable:   symbolTable,
		offset:        offset,
		lookupCallVar: lookupCallVar,
	}
}

// ConvertExpression traverses AST and generates Go code with proper series access
func (c *SeriesAccessConverter) ConvertExpression(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		return c.convertIdentifier(e)

	case *ast.MemberExpression:
		return c.convertMemberExpression(e)

	case *ast.CallExpression:
		return c.convertCallExpression(e)

	case *ast.BinaryExpression:
		return c.convertBinaryExpression(e)

	case *ast.LogicalExpression:
		return c.convertLogicalExpression(e)

	case *ast.UnaryExpression:
		return c.convertUnaryExpression(e)

	case *ast.ConditionalExpression:
		return c.convertConditionalExpression(e)

	case *ast.Literal:
		return c.convertLiteral(e)

	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (c *SeriesAccessConverter) convertIdentifier(id *ast.Identifier) (string, error) {
	name := id.Name

	// Built-in OHLCV fields handled separately
	if c.isBuiltinField(name) {
		return c.convertBuiltinField(name), nil
	}

	// Check if it's a series variable
	if c.symbolTable.IsSeries(name) {
		// Special case: offset "0" means current bar in recursive phase
		// Use scalar variable directly (current value), not historical buffer
		if c.offset == "0" {
			return name, nil
		}
		return fmt.Sprintf("%sSeries.Get(%s)", name, c.offset), nil
	}

	// Scalar variable or constant - use directly
	return name, nil
}

func (c *SeriesAccessConverter) convertMemberExpression(mem *ast.MemberExpression) (string, error) {
	// Handle patterns like: bar.Close, strategy.equity, syminfo.tickerid
	obj, ok := mem.Object.(*ast.Identifier)
	if !ok {
		return "", fmt.Errorf("complex member expression not supported: %T", mem.Object)
	}

	prop, ok := mem.Property.(*ast.Identifier)
	if !ok {
		return "", fmt.Errorf("complex member property not supported: %T", mem.Property)
	}

	// Built-in namespaced access (bar.Close → ctx.Data[i-offset].Close)
	if obj.Name == "bar" {
		return c.convertBuiltinField(prop.Name), nil
	}

	// Pass through other member expressions (strategy.equity, syminfo.tickerid, etc.)
	return fmt.Sprintf("%s.%s", obj.Name, prop.Name), nil
}

func (c *SeriesAccessConverter) convertCallExpression(call *ast.CallExpression) (string, error) {
	// Check if call has materialized temp variable - reuse instead of regenerating
	if c.lookupCallVar != nil {
		if tempVarName := c.lookupCallVar(call); tempVarName != "" {
			return fmt.Sprintf("%sSeries.Get(%s)", tempVarName, c.offset), nil
		}
	}

	// Function names should not be converted with series access logic
	// They are either builtin functions (abs → math.Abs) or user-defined functions
	var funcCode string
	if id, ok := call.Callee.(*ast.Identifier); ok {
		// Simple function name - map Pine functions to Go equivalents
		funcCode = c.mapFunctionName(id.Name)
	} else {
		// Complex callee expression (e.g., member expression) - convert it
		var err error
		funcCode, err = c.ConvertExpression(call.Callee)
		if err != nil {
			return "", fmt.Errorf("converting callee: %w", err)
		}
	}

	// Convert arguments with series access logic
	args := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		argCode, err := c.ConvertExpression(arg)
		if err != nil {
			return "", fmt.Errorf("converting argument %d: %w", i, err)
		}
		args[i] = argCode
	}

	return fmt.Sprintf("%s(%s)", funcCode, joinArgs(args)), nil
}

func (c *SeriesAccessConverter) mapFunctionName(name string) string {
	// Map Pine function names to Go equivalents
	switch name {
	case "abs":
		return "math.Abs"
	case "max":
		return "math.Max"
	case "min":
		return "math.Min"
	case "pow":
		return "math.Pow"
	case "sqrt":
		return "math.Sqrt"
	case "log":
		return "math.Log"
	case "log10":
		return "math.Log10"
	case "exp":
		return "math.Exp"
	case "ceil":
		return "math.Ceil"
	case "floor":
		return "math.Floor"
	case "round":
		return "math.Round"
	case "sign":
		return "math.Copysign(1.0,"
	default:
		// Already prefixed (math.Abs) or user-defined function - pass through
		return name
	}
}

func (c *SeriesAccessConverter) convertBinaryExpression(bin *ast.BinaryExpression) (string, error) {
	left, err := c.ConvertExpression(bin.Left)
	if err != nil {
		return "", fmt.Errorf("converting left side: %w", err)
	}

	right, err := c.ConvertExpression(bin.Right)
	if err != nil {
		return "", fmt.Errorf("converting right side: %w", err)
	}

	return fmt.Sprintf("(%s %s %s)", left, bin.Operator, right), nil
}

func (c *SeriesAccessConverter) convertLogicalExpression(logical *ast.LogicalExpression) (string, error) {
	left, err := c.ConvertExpression(logical.Left)
	if err != nil {
		return "", fmt.Errorf("converting left side: %w", err)
	}

	right, err := c.ConvertExpression(logical.Right)
	if err != nil {
		return "", fmt.Errorf("converting right side: %w", err)
	}

	// Convert logical operators to Go syntax
	operator := logical.Operator
	if operator == "and" {
		operator = "&&"
	} else if operator == "or" {
		operator = "||"
	}

	return fmt.Sprintf("(%s %s %s)", left, operator, right), nil
}

func (c *SeriesAccessConverter) convertUnaryExpression(unary *ast.UnaryExpression) (string, error) {
	operand, err := c.ConvertExpression(unary.Argument)
	if err != nil {
		return "", fmt.Errorf("converting operand: %w", err)
	}

	return fmt.Sprintf("%s%s", unary.Operator, operand), nil
}

func (c *SeriesAccessConverter) convertConditionalExpression(cond *ast.ConditionalExpression) (string, error) {
	test, err := c.ConvertExpression(cond.Test)
	if err != nil {
		return "", fmt.Errorf("converting test: %w", err)
	}

	consequent, err := c.ConvertExpression(cond.Consequent)
	if err != nil {
		return "", fmt.Errorf("converting consequent: %w", err)
	}

	alternate, err := c.ConvertExpression(cond.Alternate)
	if err != nil {
		return "", fmt.Errorf("converting alternate: %w", err)
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()", test, consequent, alternate), nil
}

func (c *SeriesAccessConverter) convertLiteral(lit *ast.Literal) (string, error) {
	switch v := lit.Value.(type) {
	case float64:
		return fmt.Sprintf("%g", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case string:
		return fmt.Sprintf("%q", v), nil
	default:
		return "", fmt.Errorf("unsupported literal type: %T", v)
	}
}

func (c *SeriesAccessConverter) isBuiltinField(name string) bool {
	builtins := map[string]bool{
		"open": true, "high": true, "low": true, "close": true,
		"volume": true, "hl2": true, "hlc3": true, "ohlc4": true, "hlcc4": true,
	}
	return builtins[name]
}

func (c *SeriesAccessConverter) convertBuiltinField(field string) string {
	if field == "open" || field == "high" || field == "low" || field == "close" || field == "volume" {
		seriesNameMap := map[string]string{
			"open":   "openSeries",
			"high":   "highSeries",
			"low":    "lowSeries",
			"close":  "closeSeries",
			"volume": "volumeSeries",
		}
		return fmt.Sprintf("%s.Get(%s)", seriesNameMap[field], c.offset)
	}

	switch field {
	case "hl2":
		return fmt.Sprintf("(highSeries.Get(%s) + lowSeries.Get(%s)) / 2", c.offset, c.offset)
	case "hlc3":
		return fmt.Sprintf("(highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s)) / 3", c.offset, c.offset, c.offset)
	case "ohlc4":
		return fmt.Sprintf("(openSeries.Get(%s) + highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s)) / 4", c.offset, c.offset, c.offset, c.offset)
	case "hlcc4":
		return fmt.Sprintf("(highSeries.Get(%s) + lowSeries.Get(%s) + closeSeries.Get(%s) + closeSeries.Get(%s)) / 4", c.offset, c.offset, c.offset, c.offset)
	}

	return field
}

func joinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		result += ", " + args[i]
	}
	return result
}
