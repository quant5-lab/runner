package codegen

type TAArgumentPattern int

const (
	TAPatternExplicitSourceAndLength TAArgumentPattern = iota
	TAPatternSingleArgIsLength
	TAPatternSingleArgIsSource
)

type TAFunctionSignature struct {
	ArgumentPattern TAArgumentPattern
	DefaultSource   string
}

type TAFunctionSignatureRegistry struct {
	signatures map[string]TAFunctionSignature
}

func NewTAFunctionSignatureRegistry() *TAFunctionSignatureRegistry {
	registry := &TAFunctionSignatureRegistry{
		signatures: make(map[string]TAFunctionSignature),
	}

	highestSig := TAFunctionSignature{
		ArgumentPattern: TAPatternSingleArgIsLength,
		DefaultSource:   "high",
	}
	lowestSig := TAFunctionSignature{
		ArgumentPattern: TAPatternSingleArgIsLength,
		DefaultSource:   "low",
	}
	changeSig := TAFunctionSignature{
		ArgumentPattern: TAPatternSingleArgIsSource,
		DefaultSource:   "",
	}
	explicitSig := TAFunctionSignature{
		ArgumentPattern: TAPatternExplicitSourceAndLength,
		DefaultSource:   "",
	}

	registry.signatures["highest"] = highestSig
	registry.signatures["ta.highest"] = highestSig

	registry.signatures["lowest"] = lowestSig
	registry.signatures["ta.lowest"] = lowestSig

	registry.signatures["highestbars"] = highestSig
	registry.signatures["ta.highestbars"] = highestSig

	registry.signatures["lowestbars"] = lowestSig
	registry.signatures["ta.lowestbars"] = lowestSig

	registry.signatures["change"] = changeSig
	registry.signatures["ta.change"] = changeSig

	registry.signatures["sma"] = explicitSig
	registry.signatures["ta.sma"] = explicitSig

	registry.signatures["ema"] = explicitSig
	registry.signatures["ta.ema"] = explicitSig

	registry.signatures["rma"] = explicitSig
	registry.signatures["ta.rma"] = explicitSig

	registry.signatures["wma"] = explicitSig
	registry.signatures["ta.wma"] = explicitSig

	registry.signatures["stdev"] = explicitSig
	registry.signatures["ta.stdev"] = explicitSig

	registry.signatures["rsi"] = explicitSig
	registry.signatures["ta.rsi"] = explicitSig

	return registry
}

func (r *TAFunctionSignatureRegistry) GetSignature(functionName string) (TAFunctionSignature, bool) {
	sig, exists := r.signatures[functionName]
	return sig, exists
}

func (r *TAFunctionSignatureRegistry) HasOptionalSource(functionName string) bool {
	sig, exists := r.signatures[functionName]
	if !exists {
		return false
	}
	return sig.ArgumentPattern == TAPatternSingleArgIsLength
}
