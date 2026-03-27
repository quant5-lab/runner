package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func extractTAArguments(call *ast.CallExpression, inputConstantsMap ...map[string]float64) (ast.Expression, int, error) {
	if len(call.Arguments) < 2 {
		funcName := extractCallFunctionName(call.Callee)
		return nil, 0, newInsufficientArgumentsError(funcName, 2, len(call.Arguments))
	}

	period, err := extractNumberLiteral(call.Arguments[1], inputConstantsMap...)
	if err != nil {
		return nil, 0, err
	}

	return call.Arguments[0], int(period), nil
}

func buildTACacheKey(funcName, sourceName string, period int) string {
	return fmt.Sprintf("%s_%s_%d", funcName, sourceName, period)
}

func buildValuewhenCacheKey(conditionExpr ast.Expression, sourceExpr ast.Expression, occurrence int) string {
	return fmt.Sprintf("valuewhen_%s_%s_%d", expressionKey(conditionExpr), expressionKey(sourceExpr), occurrence)
}

func extractSourceOnlyArgument(call *ast.CallExpression, funcName string) (ast.Expression, error) {
	if len(call.Arguments) < 1 {
		return nil, newInsufficientArgumentsError(funcName, 1, 0)
	}
	return call.Arguments[0], nil
}

func extractPeriodArgument(call *ast.CallExpression, funcName string) (int, error) {
	if len(call.Arguments) < 1 {
		return 0, newMissingArgumentError(funcName, "period")
	}

	lit, ok := call.Arguments[0].(*ast.Literal)
	if !ok {
		return 0, newInvalidArgumentError(funcName, "period", "literal")
	}

	periodFloat, ok := lit.Value.(float64)
	if !ok {
		return 0, newInvalidArgumentError(funcName, "period", "number")
	}

	return int(periodFloat), nil
}

func extractChangeArguments(call *ast.CallExpression, inputConstantsMap ...map[string]float64) (ast.Expression, int, error) {
	if len(call.Arguments) < 1 {
		return nil, 0, newInsufficientArgumentsError("change", 1, 0)
	}
	if len(call.Arguments) < 2 {
		return call.Arguments[0], 1, nil
	}
	length, err := extractNumberLiteral(call.Arguments[1], inputConstantsMap...)
	if err != nil {
		return nil, 0, err
	}
	return call.Arguments[0], int(length), nil
}

func extractLinregArguments(call *ast.CallExpression, inputConstantsMap ...map[string]float64) (ast.Expression, int, int, error) {
	sourceExpr, length, err := extractTAArguments(call, inputConstantsMap...)
	if err != nil {
		return nil, 0, 0, err
	}
	offset := 0
	if len(call.Arguments) >= 3 {
		v, err2 := extractNumberLiteral(call.Arguments[2], inputConstantsMap...)
		if err2 != nil {
			return nil, 0, 0, err2
		}
		offset = int(v)
	}
	return sourceExpr, length, offset, nil
}

func extractTwoExpressionArguments(call *ast.CallExpression, funcName string) (ast.Expression, ast.Expression, error) {
	if len(call.Arguments) < 2 {
		return nil, nil, newInsufficientArgumentsError(funcName, 2, len(call.Arguments))
	}
	return call.Arguments[0], call.Arguments[1], nil
}

func extractSingleExpressionArgument(call *ast.CallExpression, funcName string) (ast.Expression, error) {
	if len(call.Arguments) < 1 {
		return nil, newInsufficientArgumentsError(funcName, 1, 0)
	}
	return call.Arguments[0], nil
}

func extractValuewhenArguments(call *ast.CallExpression, inputConstantsMap ...map[string]float64) (ast.Expression, ast.Expression, int, error) {
	funcName := extractCallFunctionName(call.Callee)

	if len(call.Arguments) < 3 {
		return nil, nil, 0, newInsufficientArgumentsError(funcName, 3, len(call.Arguments))
	}

	conditionExpr := call.Arguments[0]
	sourceExpr := call.Arguments[1]

	occurrence, err := extractNumberLiteral(call.Arguments[2], inputConstantsMap...)
	if err != nil {
		return nil, nil, 0, err
	}

	return conditionExpr, sourceExpr, int(occurrence), nil
}
