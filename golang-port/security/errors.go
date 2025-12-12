package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SecurityError struct {
	Type    string
	Message string
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func newUnsupportedExpressionError(expr ast.Expression) error {
	return &SecurityError{
		Type:    "UnsupportedExpression",
		Message: fmt.Sprintf("expression type %T not supported", expr),
	}
}

func newUnsupportedFunctionError(funcName string) error {
	return &SecurityError{
		Type:    "UnsupportedFunction",
		Message: fmt.Sprintf("function %s not implemented", funcName),
	}
}

func newUnknownIdentifierError(name string) error {
	return &SecurityError{
		Type:    "UnknownIdentifier",
		Message: fmt.Sprintf("identifier %s not recognized", name),
	}
}

func newBarIndexOutOfRangeError(barIdx, maxBars int) error {
	return &SecurityError{
		Type:    "BarIndexOutOfRange",
		Message: fmt.Sprintf("bar index %d exceeds data length %d", barIdx, maxBars),
	}
}

func newInsufficientArgumentsError(funcName string, expected, got int) error {
	return &SecurityError{
		Type:    "InsufficientArguments",
		Message: fmt.Sprintf("%s requires %d arguments, got %d", funcName, expected, got),
	}
}

func newInvalidArgumentTypeError(funcName string, argIdx int, expected string) error {
	return &SecurityError{
		Type:    "InvalidArgumentType",
		Message: fmt.Sprintf("%s argument %d must be %s", funcName, argIdx, expected),
	}
}
