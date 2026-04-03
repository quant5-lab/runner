package codegen

import "fmt"

/* Period value abstraction — constant, runtime variable, or computed expression */
type PeriodExpression interface {
	IsConstant() bool

	/* -1 sentinel when not compile-time constant */
	AsInt() int

	AsGoExpr() string
	AsIntCast() string
	AsFloat64Cast() string

	/* Series naming key (_rma_20_ vs _rma_runtime_ vs _rma_computed_) */
	AsSeriesNamePart() string
}

type ConstantPeriod struct {
	value int
}

func NewConstantPeriod(value int) *ConstantPeriod {
	return &ConstantPeriod{value: value}
}

func (p *ConstantPeriod) IsConstant() bool {
	return true
}

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

/* Pre-rendered Go expression from arbitrary Pine arithmetic (e.g., _length/2, round(sqrt(n))) */
type ComputedPeriod struct {
	goExpression string
}

func NewComputedPeriod(goExpression string) *ComputedPeriod {
	return &ComputedPeriod{goExpression: goExpression}
}

func (p *ComputedPeriod) IsConstant() bool {
	return false
}

func (p *ComputedPeriod) AsInt() int {
	return -1
}

func (p *ComputedPeriod) AsGoExpr() string {
	return p.goExpression
}

func (p *ComputedPeriod) AsIntCast() string {
	return fmt.Sprintf("int(%s)", p.goExpression)
}

func (p *ComputedPeriod) AsFloat64Cast() string {
	return fmt.Sprintf("float64(%s)", p.goExpression)
}

func (p *ComputedPeriod) AsSeriesNamePart() string {
	return "computed"
}
