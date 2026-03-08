package codegen

type FunctionParameterType int

const (
	ParamTypeScalar FunctionParameterType = iota
	ParamTypeSeries
	ParamTypeString
)

type FunctionSignature struct {
	Name       string
	Parameters []FunctionParameterType
	ReturnType string
}

type FunctionSignatureRegistry struct {
	signatures map[string]*FunctionSignature
}

func NewFunctionSignatureRegistry() *FunctionSignatureRegistry {
	return &FunctionSignatureRegistry{
		signatures: make(map[string]*FunctionSignature),
	}
}

func (r *FunctionSignatureRegistry) Register(funcName string, paramTypes []FunctionParameterType, returnType string) {
	r.signatures[funcName] = &FunctionSignature{
		Name:       funcName,
		Parameters: paramTypes,
		ReturnType: returnType,
	}
}

func (r *FunctionSignatureRegistry) Get(funcName string) (*FunctionSignature, bool) {
	sig, exists := r.signatures[funcName]
	return sig, exists
}

func (r *FunctionSignatureRegistry) GetParameterType(funcName string, paramIndex int) (FunctionParameterType, bool) {
	sig, exists := r.signatures[funcName]
	if !exists || paramIndex >= len(sig.Parameters) {
		return ParamTypeScalar, false
	}
	return sig.Parameters[paramIndex], true
}
