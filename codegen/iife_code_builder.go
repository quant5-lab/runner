package codegen

import "fmt"

type IIFECodeBuilder struct {
	warmupPeriod     int
	warmupExpression string
	body             string
}

func NewIIFECodeBuilder() *IIFECodeBuilder {
	return &IIFECodeBuilder{}
}

func (b *IIFECodeBuilder) WithWarmupCheck(period int) *IIFECodeBuilder {
	b.warmupPeriod = period - 1
	b.warmupExpression = ""
	return b
}

func (b *IIFECodeBuilder) WithWarmupCheckPeriodExpression(period PeriodExpression, baseOffset int) *IIFECodeBuilder {
	if period.IsConstant() {
		b.warmupPeriod = period.AsInt() - 1 + baseOffset
		b.warmupExpression = ""
	} else {
		b.warmupPeriod = -1
		if baseOffset > 0 {
			b.warmupExpression = fmt.Sprintf("%s-1+%d", period.AsIntCast(), baseOffset)
		} else {
			b.warmupExpression = fmt.Sprintf("%s-1", period.AsIntCast())
		}
	}
	return b
}

func (b *IIFECodeBuilder) WithBody(body string) *IIFECodeBuilder {
	b.body = body
	return b
}

func (b *IIFECodeBuilder) Build() string {
	code := "func() float64 { "

	if b.warmupExpression != "" {
		code += fmt.Sprintf("if ctx.BarIndex < %s { return math.NaN() }; ", b.warmupExpression)
	} else if b.warmupPeriod > 0 {
		code += fmt.Sprintf("if ctx.BarIndex < %d { return math.NaN() }; ", b.warmupPeriod)
	}

	code += b.body
	code += " }()"
	return code
}
