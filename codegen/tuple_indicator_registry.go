package codegen

/* TupleIndicatorRegistry maps PineScript tuple functions to code generation specs */
type TupleIndicatorRegistry struct {
	specs map[string]*TupleIndicatorSpec
}

func NewTupleIndicatorRegistry() *TupleIndicatorRegistry {
	r := &TupleIndicatorRegistry{
		specs: make(map[string]*TupleIndicatorSpec),
	}
	r.registerBuiltinIndicators()
	return r
}

func (r *TupleIndicatorRegistry) registerBuiltinIndicators() {
	r.register(&TupleIndicatorSpec{
		FunctionName:    "ta.macd",
		OutputCount:     3,
		RuntimeFunction: "ta.Macd",
		SourceArgIndex:  0,
		PeriodArgCount:  3,
	})

	r.register(&TupleIndicatorSpec{
		FunctionName:    "macd",
		OutputCount:     3,
		RuntimeFunction: "ta.Macd",
		SourceArgIndex:  0,
		PeriodArgCount:  3,
	})

	r.register(&TupleIndicatorSpec{
		FunctionName:    "ta.bb",
		OutputCount:     3,
		RuntimeFunction: "ta.BBands",
		SourceArgIndex:  0,
		PeriodArgCount:  2,
	})

	r.register(&TupleIndicatorSpec{
		FunctionName:    "ta.stoch",
		OutputCount:     2,
		RuntimeFunction: "ta.Stoch",
		SourceArgIndex:  -1,
		PeriodArgCount:  2,
	})
}

func (r *TupleIndicatorRegistry) register(spec *TupleIndicatorSpec) {
	r.specs[spec.FunctionName] = spec
}

func (r *TupleIndicatorRegistry) Lookup(funcName string) *TupleIndicatorSpec {
	return r.specs[funcName]
}

func (r *TupleIndicatorRegistry) IsRegistered(funcName string) bool {
	return r.specs[funcName] != nil
}
