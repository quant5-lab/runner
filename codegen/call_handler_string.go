package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type StringNamespaceHandler struct {
	registry   *StringFunctionRegistry
	generators []StringCodeGenerator
}

func NewStringNamespaceHandler() *StringNamespaceHandler {
	return &StringNamespaceHandler{
		registry: NewStringFunctionRegistry(),
		generators: []StringCodeGenerator{
			&SimpleStringGenerator{},
			&TwoArgStringGenerator{},
			&SubstringGenerator{},
			&ReplaceGenerator{},
			&RepeatGenerator{},
			&ToStringGenerator{},
			&FormatGenerator{},
			&FormatTimeGenerator{},
		},
	}
}

func (h *StringNamespaceHandler) CanHandle(funcName string) bool {
	return h.registry.IsRegistered(funcName)
}

func (h *StringNamespaceHandler) GenerateCode(
	gen *generator,
	call *ast.CallExpression,
) (string, error) {
	funcName := extractCallFunctionName(call)

	if !h.CanHandle(funcName) {
		return "", nil
	}

	signature, exists := h.registry.GetSignature(funcName)
	if !exists {
		return "", fmt.Errorf("unregistered function: %s", funcName)
	}

	if err := signature.ValidateArgCount(len(call.Arguments)); err != nil {
		return "", err
	}

	parser := NewStringArgumentParser(gen)
	args, err := parser.ParseArguments(call)
	if err != nil {
		return "", fmt.Errorf("%s: %w", funcName, err)
	}

	for _, generator := range h.generators {
		if generator.CanGenerate(funcName) {
			return generator.Generate(gen, funcName, args)
		}
	}

	return "", fmt.Errorf("no generator found for %s", funcName)
}
