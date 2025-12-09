package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

func extractTAArguments(call *ast.CallExpression) (*ast.Identifier, int, error) {
	if len(call.Arguments) < 2 {
		funcName := extractCallFunctionName(call.Callee)
		return nil, 0, newInsufficientArgumentsError(funcName, 2, len(call.Arguments))
	}

	sourceID, ok := call.Arguments[0].(*ast.Identifier)
	if !ok {
		funcName := extractCallFunctionName(call.Callee)
		return nil, 0, newInvalidArgumentTypeError(funcName, 0, "identifier")
	}

	period, err := extractNumberLiteral(call.Arguments[1])
	if err != nil {
		return nil, 0, err
	}

	return sourceID, int(period), nil
}

func buildTACacheKey(funcName, sourceName string, period int) string {
	return fmt.Sprintf("%s_%s_%d", funcName, sourceName, period)
}
