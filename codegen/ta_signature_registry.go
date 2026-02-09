package codegen

/* Immutable singleton — safe for concurrent reads, no writes after init */
var sharedTASignatures = NewTASignatureRegistry()

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
		RegisterUtilitySignatures(),
		RegisterTupleSignatures(),
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

func (r *TASignatureRegistry) IsTupleFunction(functionName string) bool {
	metadata, exists := r.functionsByName[functionName]
	return exists && metadata.IsTuple
}

func (r *TASignatureRegistry) AllFunctionNames() []string {
	names := make([]string, 0, len(r.functionsByName))
	for name := range r.functionsByName {
		names = append(names, name)
	}
	return names
}

/*
	Returns true when first series arg needs promotion to *series.Series for historical lookback

(source+length pattern: has both series and scalar args in the overload)
*/
func (r *TASignatureRegistry) NeedsSourcePromotion(functionName string, argCount int) bool {
	metadata, exists := r.functionsByName[functionName]
	if !exists {
		return false
	}
	overload, found := metadata.FindOverload(argCount)
	if !found {
		return false
	}
	hasSeries := false
	hasScalar := false
	for _, arg := range overload.Arguments {
		if arg.Classification == TAArgImplicitOHLC {
			continue
		}
		if arg.Classification.IsSeries() {
			hasSeries = true
		}
		if arg.Classification.IsScalar() {
			hasScalar = true
		}
	}
	return hasSeries && hasScalar
}

/* Source-only functions with implicit lookback (e.g. swma reads 4 bars despite no length param) */
func (r *TASignatureRegistry) IsSourceOnlyLookback(functionName string) bool {
	metadata, exists := r.functionsByName[functionName]
	if !exists {
		return false
	}
	return metadata.SourceOnlyLookback
}
