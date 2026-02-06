package codegen

type PeriodTypeQualifier int

const (
	PeriodSimpleInt PeriodTypeQualifier = iota
	PeriodSeriesInt
)

func (q PeriodTypeQualifier) AllowsRuntimeDynamic() bool {
	return q == PeriodSeriesInt
}

func (q PeriodTypeQualifier) String() string {
	if q == PeriodSimpleInt {
		return "simple int"
	}
	return "series int"
}

type TAFunctionPeriodSpec struct {
	FunctionName      string
	PeriodQualifier   PeriodTypeQualifier
	ParameterPosition int
}

type PeriodRequirementRepository struct {
	specs map[string]TAFunctionPeriodSpec
}

func NewPeriodRequirementRepository() *PeriodRequirementRepository {
	repo := &PeriodRequirementRepository{
		specs: make(map[string]TAFunctionPeriodSpec),
	}
	repo.registerBuiltinSpecs()
	return repo
}

func (r *PeriodRequirementRepository) registerBuiltinSpecs() {
	/* PineScript Reference: simple int = compile-time constant only */
	simpleIntIndicators := map[string]int{
		"ta.rsi": 1,
		"ta.atr": 0,
		"ta.rma": 1,
		"ta.dmi": 1,
		"ta.tsi": 1,
		"ta.kc":  1,
		"ta.ema": 1,
	}

	/* PineScript Reference: series int = runtime dynamic allowed */
	seriesIntIndicators := map[string]int{
		"ta.sma":     1,
		"ta.stdev":   1,
		"ta.wma":     1,
		"ta.vwma":    1,
		"ta.hma":     1,
		"ta.highest": 1,
		"ta.lowest":  1,
		"ta.bb":      1,
		"ta.macd":    1,
	}

	for fn, pos := range simpleIntIndicators {
		r.specs[fn] = TAFunctionPeriodSpec{
			FunctionName:      fn,
			PeriodQualifier:   PeriodSimpleInt,
			ParameterPosition: pos,
		}
	}

	for fn, pos := range seriesIntIndicators {
		if fn == "ta.highest" || fn == "ta.lowest" {
			pos = 0
		}
		r.specs[fn] = TAFunctionPeriodSpec{
			FunctionName:      fn,
			PeriodQualifier:   PeriodSeriesInt,
			ParameterPosition: pos,
		}
	}
}

func (r *PeriodRequirementRepository) GetSpec(functionName string) (TAFunctionPeriodSpec, bool) {
	spec, exists := r.specs[functionName]
	return spec, exists
}

func (r *PeriodRequirementRepository) AllowsRuntimeDynamic(functionName string) bool {
	spec, exists := r.specs[functionName]
	if !exists {
		return false
	}
	return spec.PeriodQualifier.AllowsRuntimeDynamic()
}
