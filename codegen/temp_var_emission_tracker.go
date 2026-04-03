package codegen

type TempVarEmissionTracker struct {
	emittedVars map[string]bool
}

func NewTempVarEmissionTracker() *TempVarEmissionTracker {
	return &TempVarEmissionTracker{
		emittedVars: make(map[string]bool),
	}
}

func (t *TempVarEmissionTracker) MarkAsEmitted(varName string) {
	t.emittedVars[varName] = true
}

func (t *TempVarEmissionTracker) WasEmitted(varName string) bool {
	return t.emittedVars[varName]
}

func (t *TempVarEmissionTracker) Reset() {
	t.emittedVars = make(map[string]bool)
}
