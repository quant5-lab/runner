package codegen

type PeriodType int

const (
	PeriodCompileTimeConstant PeriodType = iota
	PeriodRuntimeDynamic
)

type PeriodClassifier struct{}

func NewPeriodClassifier() *PeriodClassifier {
	return &PeriodClassifier{}
}

func (c *PeriodClassifier) Classify(periodValue int, periodExpr string) PeriodType {
	if periodValue > 0 && periodExpr == "" {
		return PeriodCompileTimeConstant
	}
	return PeriodRuntimeDynamic
}
