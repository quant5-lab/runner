package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* UserDefinedFunctionHandler generates calls to user-defined arrow functions.
 * Handles proper ctx parameter passing and argument marshaling.
 */
type UserDefinedFunctionHandler struct{}

func (h *UserDefinedFunctionHandler) CanHandle(funcName string) bool {
	return false
}

func (h *UserDefinedFunctionHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	// In arrow function context, check if this is an unprefixed TA function
	if g.inArrowFunctionBody {
		if h.isUnprefixedTAFunction(funcName) {
			taHandler := &TAIndicatorCallHandler{}
			return taHandler.generateArrowFunctionTACall(g, call)
		}
	}

	// Check if this is a user-defined function
	varType, exists := g.variables[funcName]
	if !exists || varType != "function" {
		return "", nil // Not a user-defined function, let next handler try
	}

	// Generate arguments
	var args []string
	args = append(args, "ctx") // First parameter is always ctx

	for _, arg := range call.Arguments {
		argCode, err := h.generateArgumentExpression(g, arg)
		if err != nil {
			return "", fmt.Errorf("failed to generate argument: %w", err)
		}
		args = append(args, argCode)
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", ")), nil
}

func (h *UserDefinedFunctionHandler) isUnprefixedTAFunction(funcName string) bool {
	switch funcName {
	case "sma", "ema", "stdev", "rma", "wma":
		return true
	default:
		return false
	}
}

func (h *UserDefinedFunctionHandler) generateArgumentExpression(g *generator, expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// Check if it's a builtin identifier (like close, open, high, low)
		if code, resolved := g.builtinHandler.TryResolveIdentifier(e, g.inSecurityContext); resolved {
			// Builtin identifiers need Series access for user function calls
			// Convert "bar.Close" to "closeSeries.Get(0)"
			switch e.Name {
			case "close":
				return "closeSeries.Get(0)", nil
			case "open":
				return "openSeries.Get(0)", nil
			case "high":
				return "highSeries.Get(0)", nil
			case "low":
				return "lowSeries.Get(0)", nil
			case "volume":
				return "volumeSeries.Get(0)", nil
			default:
				// Non-bar-field builtin, use as-is
				return code, nil
			}
		}
		// Function parameter, variable, or constant - return name directly
		return e.Name, nil

	case *ast.Literal:
		switch v := e.Value.(type) {
		case float64:
			return fmt.Sprintf("%.1f", v), nil
		case int:
			return fmt.Sprintf("%d.0", v), nil
		default:
			return fmt.Sprintf("%v", v), nil
		}

	case *ast.CallExpression:
		return g.generateCallExpression(e)

	case *ast.BinaryExpression:
		left, err := h.generateArgumentExpression(g, e.Left)
		if err != nil {
			return "", err
		}
		right, err := h.generateArgumentExpression(g, e.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right), nil

	case *ast.MemberExpression:
		return g.generateMemberExpression(e)

	default:
		return "", fmt.Errorf("unsupported argument expression type: %T", expr)
	}
}
