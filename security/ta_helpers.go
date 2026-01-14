package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func extractTAArguments(call *ast.CallExpression, inputConstantsMap ...map[string]float64) (*ast.Identifier, int, error) {
	if len(call.Arguments) < 2 {
		funcName := extractCallFunctionName(call.Callee)
		return nil, 0, newInsufficientArgumentsError(funcName, 2, len(call.Arguments))
	}

	sourceID, ok := call.Arguments[0].(*ast.Identifier)
	if !ok {
		funcName := extractCallFunctionName(call.Callee)
		return nil, 0, newInvalidArgumentTypeError(funcName, 0, "identifier")
	}

	period, err := extractNumberLiteral(call.Arguments[1], inputConstantsMap...)
	if err != nil {
		return nil, 0, err
	}

	return sourceID, int(period), nil
}

func buildTACacheKey(funcName, sourceName string, period int) string {
	return fmt.Sprintf("%s_%s_%d", funcName, sourceName, period)
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
