package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ArrayConstructorHandler struct{}

func NewArrayConstructorHandler() *ArrayConstructorHandler {
	return &ArrayConstructorHandler{}
}

func (h *ArrayConstructorHandler) CanHandle(funcName string) bool {
	classifier := NewArrayConstructorClassifier()
	return classifier.IsArrayConstructor(funcName)
}

func (h *ArrayConstructorHandler) GenerateCode(g *generator, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	switch funcName {
	case "array.new_float", "array.new_int", "array.new_bool":
		return h.generateNewArray(g, call)
	case "array.from":
		return h.generateFrom(g, call)
	}

	return "", nil
}

func (h *ArrayConstructorHandler) generateNewArray(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "[]float64{}", nil
	}

	sizeCode, err := g.generateArrowFunctionExpression(call.Arguments[0])
	if err != nil {
		return "", fmt.Errorf("array.new: size: %w", err)
	}

	if len(call.Arguments) == 1 {
		return fmt.Sprintf("make([]float64, int(%s))", sizeCode), nil
	}

	initialCode, err := g.generateArrowFunctionExpression(call.Arguments[1])
	if err != nil {
		return "", fmt.Errorf("array.new: initial value: %w", err)
	}

	return fmt.Sprintf("arrayops.NewArrayWithValue(int(%s), %s)", sizeCode, initialCode), nil
}

func (h *ArrayConstructorHandler) generateFrom(g *generator, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) == 0 {
		return "[]float64{}", nil
	}

	elemCodes := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		code, err := g.generateArrowFunctionExpression(arg)
		if err != nil {
			return "", fmt.Errorf("array.from: element %d: %w", i, err)
		}
		elemCodes[i] = code
	}

	return fmt.Sprintf("[]float64{%s}", joinStrings(elemCodes, ", ")), nil
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
