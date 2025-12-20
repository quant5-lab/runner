package codegen

import "github.com/quant5-lab/runner/ast"

type SignatureRegistrar struct {
	registry *FunctionSignatureRegistry
	mapper   *ParameterSignatureMapper
}

func NewSignatureRegistrar(registry *FunctionSignatureRegistry) *SignatureRegistrar {
	return &SignatureRegistrar{
		registry: registry,
		mapper:   NewParameterSignatureMapper(),
	}
}

func (r *SignatureRegistrar) RegisterArrowFunction(funcName string, params []ast.Identifier, paramUsage map[string]ParameterUsageType, returnType string) {
	signatureTypes := r.mapper.MapUsageToSignatureTypes(params, paramUsage)
	r.registry.Register(funcName, signatureTypes, returnType)
}
