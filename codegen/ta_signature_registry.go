package codegen

type TASignatureRegistry struct {
	functionsByName map[string]TAFunctionMetadata
}

func NewTASignatureRegistry() *TASignatureRegistry {
	registry := &TASignatureRegistry{
		functionsByName: make(map[string]TAFunctionMetadata),
	}
	registry.registerAllSignatures()
	return registry
}

func (r *TASignatureRegistry) registerAllSignatures() {
	allSignatures := [][]TAFunctionMetadata{
		RegisterMovingAverageSignatures(),
		RegisterVolatilitySignatures(),
		RegisterPivotSignatures(),
		RegisterOscillatorSignatures(),
		RegisterOverlaySignatures(),
		RegisterStatisticsSignatures(),
		RegisterMinMaxSignatures(),
	}

	for _, signatureGroup := range allSignatures {
		for _, metadata := range signatureGroup {
			r.functionsByName[metadata.FunctionName] = metadata
		}
	}
}

func (r *TASignatureRegistry) Lookup(functionName string) (TAFunctionMetadata, bool) {
	metadata, exists := r.functionsByName[functionName]
	return metadata, exists
}

func (r *TASignatureRegistry) Contains(functionName string) bool {
	_, exists := r.functionsByName[functionName]
	return exists
}

func (r *TASignatureRegistry) AllFunctionNames() []string {
	names := make([]string, 0, len(r.functionsByName))
	for name := range r.functionsByName {
		names = append(names, name)
	}
	return names
}
