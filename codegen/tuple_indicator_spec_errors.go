package codegen

import "fmt"

func errInvalidOutputCount(funcName string, count int) error {
	return fmt.Errorf("tuple indicator %s: invalid output count %d (minimum 2)", funcName, count)
}

func errMissingRuntimeFunction(funcName string) error {
	return fmt.Errorf("tuple indicator %s: missing runtime function mapping", funcName)
}

func errIndicatorNotRegistered(funcName string) error {
	return fmt.Errorf("tuple indicator %s: not registered", funcName)
}

func errOutputCountMismatch(funcName string, expected, actual int) error {
	return fmt.Errorf("tuple indicator %s: expected %d outputs, got %d", funcName, expected, actual)
}
