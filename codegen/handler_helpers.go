package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

/* Helper functions shared across TA handlers */

/* extractTAArgumentsAST extracts source and period from TA arguments, returning AST node for ClassifyAST() */
func extractTAArgumentsAST(g *generator, call *ast.CallExpression, funcName string) (ast.Expression, int, error) {
	if len(call.Arguments) < 2 {
		return nil, 0, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceASTExpr := call.Arguments[0]
	periodArg := call.Arguments[1]

	/* Try literal period first (fast path) */
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriod(periodLit)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", funcName, err)
		}
		return sourceASTExpr, period, nil
	}

	/* Try compile-time constant evaluation (handles variables + expressions) */
	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return sourceASTExpr, int(periodValue), nil
	}

	return nil, 0, fmt.Errorf("%s period must be compile-time constant (got %T that evaluates to NaN)", funcName, periodArg)
}

/* extractSinglePeriodArgument extracts period from single-argument TA functions */
func extractSinglePeriodArgument(g *generator, call *ast.CallExpression, funcName string) (int, error) {
	if len(call.Arguments) < 1 {
		return 0, fmt.Errorf("%s requires 1 argument (period)", funcName)
	}

	periodArg := call.Arguments[0]

	/* Fast path: literal period */
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriod(periodLit)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", funcName, err)
		}
		return period, nil
	}

	/* Compile-time constant evaluation (handles input() variables) */
	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return int(periodValue), nil
	}

	return 0, fmt.Errorf("%s period must be compile-time constant (got %T)", funcName, periodArg)
}

/* extractPeriod converts a literal to an integer period value */
func extractPeriod(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric, got %T", v)
	}
}

/* generateCrossDetection generates code for crossover/crossunder detection */
func generateCrossDetection(g *generator, varName string, call *ast.CallExpression, isCrossunder bool) (string, error) {
	if len(call.Arguments) < 2 {
		funcName := "ta.crossover"
		if isCrossunder {
			funcName = "ta.crossunder"
		}
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	series1 := g.extractSeriesExpression(call.Arguments[0])
	series2 := g.extractSeriesExpression(call.Arguments[1])

	prev1Var := varName + "_prev1"
	prev2Var := varName + "_prev2"

	var code string
	var description string
	var condition string

	if isCrossunder {
		description = fmt.Sprintf("// Crossunder: %s crosses below %s\n", series1, series2)
		condition = fmt.Sprintf("if %s < %s && %s >= %s { return 1.0 } else { return 0.0 }", series1, series2, prev1Var, prev2Var)
	} else {
		description = fmt.Sprintf("// Crossover: %s crosses above %s\n", series1, series2)
		condition = fmt.Sprintf("if %s > %s && %s <= %s { return 1.0 } else { return 0.0 }", series1, series2, prev1Var, prev2Var)
	}

	code += g.ind() + description
	code += g.ind() + "if i > 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev1Var, g.convertSeriesAccessToPrev(series1))
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, g.convertSeriesAccessToPrev(series2))
	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { %s }())\n", varName, condition)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

/* detectV4InputType detects Pine v4 input() type parameter, returns normalized v5 function name */
func detectV4InputType(call *ast.CallExpression) string {
	for _, arg := range call.Arguments {
		objExpr, ok := arg.(*ast.ObjectExpression)
		if !ok {
			continue
		}

		for _, prop := range objExpr.Properties {
			keyId, ok := prop.Key.(*ast.Identifier)
			if !ok || keyId.Name != "type" {
				continue
			}

			memExpr, ok := prop.Value.(*ast.MemberExpression)
			if !ok {
				continue
			}

			objId, ok := memExpr.Object.(*ast.Identifier)
			if !ok || objId.Name != "input" {
				continue
			}

			propId, ok := memExpr.Property.(*ast.Identifier)
			if !ok {
				continue
			}

			/* Map v4 type names to v5 function names */
			switch propId.Name {
			case "session":
				return "input.session"
			case "integer":
				return "input.int"
			case "float":
				return "input.float"
			case "bool":
				return "input.bool"
			case "string":
				return "input.string"
			}
		}
	}

	return ""
}

/* inferInputTypeFromLiteral infers input type from first literal argument value */
func inferInputTypeFromLiteral(call *ast.CallExpression) string {
	if len(call.Arguments) == 0 {
		return ""
	}

	lit, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return ""
	}

	switch v := lit.Value.(type) {
	case float64:
		if v == float64(int(v)) {
			return "input.int"
		}
		return "input.float"
	case int:
		return "input.int"
	default:
		return ""
	}
}
