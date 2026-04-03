package codegen

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

func extractTAArgumentsAST(g *generator, call *ast.CallExpression, funcName string) (ast.Expression, int, error) {
	if len(call.Arguments) < 2 {
		return nil, 0, fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceASTExpr := call.Arguments[0]
	periodArg := call.Arguments[1]

	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriod(periodLit)
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", funcName, err)
		}
		return sourceASTExpr, period, nil
	}

	if periodValue := tryExtractInputIntValue(periodArg); periodValue > 0 {
		return sourceASTExpr, periodValue, nil
	}

	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return sourceASTExpr, int(periodValue), nil
	}

	return nil, 0, fmt.Errorf("%s period must be compile-time constant (got %T that evaluates to NaN)", funcName, periodArg)
}

func extractSinglePeriodArgument(g *generator, call *ast.CallExpression, funcName string) (int, error) {
	if len(call.Arguments) < 1 {
		return 0, fmt.Errorf("%s requires 1 argument (period)", funcName)
	}

	periodArg := call.Arguments[0]

	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriod(periodLit)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", funcName, err)
		}
		return period, nil
	}

	if periodValue := tryExtractInputIntValue(periodArg); periodValue > 0 {
		return periodValue, nil
	}

	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return int(periodValue), nil
	}

	return 0, fmt.Errorf("%s period must be compile-time constant (got %T)", funcName, periodArg)
}

/* evaluatePeriodExpression converts period argument to PeriodEvaluationResult.
 * Single source of truth for period evaluation cascade.
 * Returns CompileTimeConstant for literals/input.int/resolvable expressions.
 * Returns RuntimeDynamic for non-constant expressions (variables, ternaries).
 * Returns Failed for validation errors (non-positive literals). */
func evaluatePeriodExpression(g *generator, periodArg ast.Expression) PeriodEvaluationResult {
	if periodLit, ok := periodArg.(*ast.Literal); ok {
		period, err := extractPeriodFromLiteral(periodLit)
		if err != nil {
			return NewFailedPeriodEvaluation(err.Error())
		}
		return NewCompileTimeConstantPeriod(period)
	}

	if periodValue := tryExtractInputIntValue(periodArg); periodValue > 0 {
		return NewCompileTimeConstantPeriod(periodValue)
	}

	periodValue := g.constEvaluator.EvaluateConstant(periodArg)
	if !math.IsNaN(periodValue) && periodValue > 0 {
		return NewCompileTimeConstantPeriod(int(periodValue))
	}

	return NewRuntimeDynamicPeriod(periodArg)
}

func extractTAArgumentsWithDynamic(g *generator, call *ast.CallExpression, funcName string) (ast.Expression, PeriodEvaluationResult) {
	if len(call.Arguments) < 2 {
		return nil, NewFailedPeriodEvaluation(fmt.Sprintf("%s requires at least 2 arguments", funcName))
	}

	return call.Arguments[0], evaluatePeriodExpression(g, call.Arguments[1])
}

func extractSinglePeriodWithDynamic(g *generator, call *ast.CallExpression, funcName string) PeriodEvaluationResult {
	if len(call.Arguments) < 1 {
		return NewFailedPeriodEvaluation(fmt.Sprintf("%s requires 1 argument (period)", funcName))
	}

	return evaluatePeriodExpression(g, call.Arguments[0])
}

func extractPeriod(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		if v <= 0 {
			return 0, fmt.Errorf("period must be positive, got %.0f", v)
		}
		return int(v), nil
	case int:
		if v <= 0 {
			return 0, fmt.Errorf("period must be positive, got %d", v)
		}
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric, got %T", v)
	}
}

func extractIntegerValue(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("value must be non-negative, got %.0f", v)
		}
		return int(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("value must be non-negative, got %d", v)
		}
		return v, nil
	default:
		return 0, fmt.Errorf("value must be numeric, got %T", v)
	}
}

func tryExtractInputIntValue(expr ast.Expression) int {
	call, ok := expr.(*ast.CallExpression)
	if !ok {
		return 0
	}

	funcName := extractFunctionNameFromCall(call)
	if funcName != "input.int" && funcName != "input" {
		return 0
	}

	if len(call.Arguments) == 0 {
		return 0
	}

	firstArg := call.Arguments[0]
	if lit, ok := firstArg.(*ast.Literal); ok {
		switch v := lit.Value.(type) {
		case float64:
			if v > 0 {
				return int(v)
			}
		case int:
			if v > 0 {
				return v
			}
		}
	}

	return 0
}

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
	// Ensure prev2 is float64 when comparing with series (which returns float64)
	prev2Value := g.convertSeriesAccessToPrev(series2)
	prev2Value = ensureFloat64Literal(prev2Value)
	code += g.ind() + fmt.Sprintf("%s := %s\n", prev2Var, prev2Value)
	code += g.ind() + fmt.Sprintf("%sSeries.Set(func() float64 { %s }())\n", varName, condition)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
	g.indent--
	code += g.ind() + "}\n"

	return code, nil
}

/* ensureFloat64Literal ensures a numeric literal has .0 suffix for Go type safety */
func ensureFloat64Literal(s string) string {
	// Don't modify non-numeric values (series accesses, bar. accesses, etc.)
	if strings.Contains(s, "Series") || strings.Contains(s, "bar.") || strings.Contains(s, "ctx.") {
		return s
	}
	// Don't modify if already has decimal or scientific notation
	if strings.Contains(s, ".") || strings.Contains(s, "e") || strings.Contains(s, "E") {
		return s
	}
	// Try to parse as number - if successful, add .0 suffix
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s + ".0"
	}
	return s
}
