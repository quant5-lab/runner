package codegen

import "fmt"

type IIFECodeBuilder struct {
	warmupPeriod int
	body         string
}

func NewIIFECodeBuilder() *IIFECodeBuilder {
	return &IIFECodeBuilder{}
}

func (b *IIFECodeBuilder) WithWarmupCheck(period int) *IIFECodeBuilder {
	b.warmupPeriod = period - 1
	return b
}

func (b *IIFECodeBuilder) WithBody(body string) *IIFECodeBuilder {
	b.body = body
	return b
}

func (b *IIFECodeBuilder) Build() string {
	code := "func() float64 { "
	if b.warmupPeriod > 0 {
		code += fmt.Sprintf("if ctx.BarIndex < %d { return math.NaN() }; ", b.warmupPeriod)
	}
	code += b.body
	code += " }()"
	return code
}
