package codegen

/* TupleIndicatorSpec defines contract for tuple-returning indicator code generation */
type TupleIndicatorSpec struct {
	FunctionName    string
	OutputCount     int
	RuntimeFunction string
	SourceArgIndex  int
	PeriodArgCount  int
	/* ImplicitSources specifies series to inject from context (e.g. ["high", "low", "close"] for DMI) */
	ImplicitSources []string
}

func (s *TupleIndicatorSpec) Validate() error {
	if s.OutputCount < 2 {
		return errInvalidOutputCount(s.FunctionName, s.OutputCount)
	}
	if s.RuntimeFunction == "" {
		return errMissingRuntimeFunction(s.FunctionName)
	}
	return nil
}
