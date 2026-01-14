package codegen

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

/* Helper functions shared across TA handlers */

/* extractTAArgumentsAST extracts source AST expression and period from standard TA function arguments.
 * Returns AST node directly for use with ClassifyAST() to avoid code generation artifacts.
 * Supports: literals (14), variables (sr_len), expressions (round(sr_n / 2))
 */
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
