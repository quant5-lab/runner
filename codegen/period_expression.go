package codegen

import "fmt"

/* PeriodExpression represents period value in TA indicators - either compile-time constant or runtime variable */
type PeriodExpression interface {
	/* IsConstant returns true if period is known at compile time */
	IsConstant() bool

	/* AsInt returns integer value if constant, -1 if runtime */
	AsInt() int

	/* AsGoExpr returns Go expression string for code generation */
	AsGoExpr() string

	/* AsIntCast returns int(expr) for loop conditions */
	AsIntCast() string

	/* AsFloat64Cast returns float64(expr) for calculations */
	AsFloat64Cast() string

	/* AsSeriesNamePart returns string for series naming (_rma_20_ vs _rma_runtime_) */
	AsSeriesNamePart() string
}

/* ConstantPeriod represents compile-time constant period (e.g., 20) */
type ConstantPeriod struct {
	value int
}

func NewConstantPeriod(value int) *ConstantPeriod {
	return &ConstantPeriod{value: value}
}

func (p *ConstantPeriod) IsConstant() bool {
	return true
}

/* Value returns the integer value for compile-time constants */
func (p *ConstantPeriod) Value() int {
	return p.value
}

func (p *ConstantPeriod) AsInt() int {
	return p.value
}

func (p *ConstantPeriod) AsGoExpr() string {
	return fmt.Sprintf("%d", p.value)
}

func (p *ConstantPeriod) AsIntCast() string {
	return fmt.Sprintf("%d", p.value)
}

func (p *ConstantPeriod) AsFloat64Cast() string {
	return fmt.Sprintf("float64(%d)", p.value)
}

func (p *ConstantPeriod) AsSeriesNamePart() string {
	return fmt.Sprintf("%d", p.value)
}

/* RuntimePeriod represents runtime variable period (e.g., len parameter) */
type RuntimePeriod struct {
	variableName string
}

func NewRuntimePeriod(variableName string) *RuntimePeriod {
	return &RuntimePeriod{variableName: variableName}
}

func (p *RuntimePeriod) IsConstant() bool {
	return false
}

func (p *RuntimePeriod) AsInt() int {
	return -1
}

func (p *RuntimePeriod) AsGoExpr() string {
	return p.variableName
}

func (p *RuntimePeriod) AsIntCast() string {
	return fmt.Sprintf("int(%s)", p.variableName)
}

func (p *RuntimePeriod) AsFloat64Cast() string {
	return fmt.Sprintf("float64(%s)", p.variableName)
}

func (p *RuntimePeriod) AsSeriesNamePart() string {
	return "runtime"
}
