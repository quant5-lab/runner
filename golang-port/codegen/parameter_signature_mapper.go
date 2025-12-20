package codegen

import "github.com/quant5-lab/runner/ast"

type ParameterSignatureMapper struct{}

func NewParameterSignatureMapper() *ParameterSignatureMapper {
	return &ParameterSignatureMapper{}
}

func (m *ParameterSignatureMapper) MapUsageToSignatureTypes(params []ast.Identifier, usageTypes map[string]ParameterUsageType) []FunctionParameterType {
	signatureTypes := make([]FunctionParameterType, 0, len(params))

	for _, param := range params {
		signatureTypes = append(signatureTypes, m.mapSingleParameter(param.Name, usageTypes))
	}

	return signatureTypes
}

func (m *ParameterSignatureMapper) mapSingleParameter(paramName string, usageTypes map[string]ParameterUsageType) FunctionParameterType {
	if usageTypes[paramName] == ParameterUsageSeries {
		return ParamTypeSeries
	}
	return ParamTypeScalar
}
