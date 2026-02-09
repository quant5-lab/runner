package codegen

import "strings"

/* Immutable singleton — safe for concurrent reads, no writes after init */
var sharedTupleIndicatorRegistry = NewTupleIndicatorRegistry()

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
	r.registerWithBareAlias(&TupleIndicatorSpec{
		FunctionName:    "ta.macd",
		OutputCount:     3,
		RuntimeFunction: "ta.Macd",
		SourceArgIndex:  0,
		PeriodArgCount:  3,
	})

	r.registerWithBareAlias(&TupleIndicatorSpec{
		FunctionName:    "ta.bb",
		OutputCount:     3,
		RuntimeFunction: "ta.BBands",
		SourceArgIndex:  0,
		PeriodArgCount:  2,
	})

	r.registerWithBareAlias(&TupleIndicatorSpec{
		FunctionName:    "ta.stoch",
		OutputCount:     2,
		RuntimeFunction: "ta.Stoch",
		SourceArgIndex:  -1,
		PeriodArgCount:  2,
	})
}

/* registerWithBareAlias registers ta.X and automatically derives bare X alias */
func (r *TupleIndicatorRegistry) registerWithBareAlias(spec *TupleIndicatorSpec) {
	r.specs[spec.FunctionName] = spec
	if i := strings.LastIndex(spec.FunctionName, "."); i >= 0 {
		bare := spec.FunctionName[i+1:]
		r.specs[bare] = &TupleIndicatorSpec{
			FunctionName:    bare,
			OutputCount:     spec.OutputCount,
			RuntimeFunction: spec.RuntimeFunction,
			SourceArgIndex:  spec.SourceArgIndex,
			PeriodArgCount:  spec.PeriodArgCount,
		}
	}
}

func (r *TupleIndicatorRegistry) Lookup(funcName string) *TupleIndicatorSpec {
	return r.specs[funcName]
}

func (r *TupleIndicatorRegistry) IsRegistered(funcName string) bool {
	return r.specs[funcName] != nil
}
