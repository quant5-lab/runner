package codegen

import "fmt"

// TAIndicatorBuilder builds technical analysis indicator code
type TAIndicatorBuilder struct {
	indicatorName string
	varName       string
	period        int
	warmupChecker *WarmupChecker
	loopGen       *LoopGenerator
	accumulator   AccumulatorStrategy
	indenter      CodeIndenter
}

func NewTAIndicatorBuilder(name, varName string, period int, accessor AccessGenerator, needsNaN bool) *TAIndicatorBuilder {
	return &TAIndicatorBuilder{
		indicatorName: name,
		varName:       varName,
		period:        period,
		warmupChecker: NewWarmupChecker(period),
		loopGen:       NewLoopGenerator(period, accessor, needsNaN),
		indenter:      NewCodeIndenter(),
	}
}

func (b *TAIndicatorBuilder) WithAccumulator(acc AccumulatorStrategy) *TAIndicatorBuilder {
	b.accumulator = acc
	return b
}

func (b *TAIndicatorBuilder) BuildHeader() string {
	return b.indenter.Line(fmt.Sprintf("/* Inline %s(%d) */", b.indicatorName, b.period))
}

func (b *TAIndicatorBuilder) BuildWarmupCheck() string {
	return b.warmupChecker.GenerateCheck(b.varName, &b.indenter)
}

func (b *TAIndicatorBuilder) BuildInitialization() string {
	if b.accumulator == nil {
		return ""
	}

	code := ""
	initCode := b.accumulator.Initialize()
	if initCode != "" {
		code += b.indenter.Line(initCode)
	}
	return code
}

func (b *TAIndicatorBuilder) BuildLoop(loopBody func(valueExpr string) string) string {
	code := b.loopGen.GenerateForwardLoop(&b.indenter)

	b.indenter.IncreaseIndent()
	valueAccess := b.loopGen.GenerateValueAccess()

	if b.loopGen.RequiresNaNCheck() && b.accumulator.NeedsNaNGuard() {
		code += b.indenter.Line(fmt.Sprintf("val := %s", valueAccess))
		code += b.indenter.Line("if math.IsNaN(val) {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line("hasNaN = true")
		code += b.indenter.Line("break")
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
		code += loopBody("val")
	} else {
		code += loopBody(valueAccess)
	}

	b.indenter.DecreaseIndent()
	code += b.indenter.Line("}")

	return code
}

func (b *TAIndicatorBuilder) BuildFinalization(resultExpr string) string {
	code := ""

	if b.accumulator.NeedsNaNGuard() {
		code += b.indenter.Line("if hasNaN {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(math.NaN())", b.varName))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("} else {")
		b.indenter.IncreaseIndent()
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(%s)", b.varName, resultExpr))
		b.indenter.DecreaseIndent()
		code += b.indenter.Line("}")
	} else {
		code += b.indenter.Line(fmt.Sprintf("%sSeries.Set(%s)", b.varName, resultExpr))
	}

	return code
}

func (b *TAIndicatorBuilder) CloseBlock() string {
	b.indenter.DecreaseIndent()
	return b.indenter.Line("}")
}

// Build generates complete TA indicator code
func (b *TAIndicatorBuilder) Build() string {
	b.indenter.IncreaseIndent() // Start at indent level 1

	code := b.BuildHeader()
	code += b.BuildWarmupCheck()

	b.indenter.IncreaseIndent()
	code += b.BuildInitialization()

	if b.accumulator != nil {
		code += b.BuildLoop(func(val string) string {
			return b.indenter.Line(b.accumulator.Accumulate(val))
		})

		finalizeCode := b.accumulator.Finalize(b.period)
		code += b.BuildFinalization(finalizeCode)
	}

	code += b.CloseBlock()

	return code
}

// CodeIndenter implements Indenter interface
type CodeIndenter struct {
	level int
	tab   string
}

func NewCodeIndenter() CodeIndenter {
	return CodeIndenter{level: 0, tab: "\t"}
}

func (c *CodeIndenter) Line(code string) string {
	indent := ""
	for i := 0; i < c.level; i++ {
		indent += c.tab
	}
	return indent + code + "\n"
}

func (c *CodeIndenter) Indent(fn func() string) string {
	c.level++
	result := fn()
	c.level--
	return result
}

func (c *CodeIndenter) CurrentLevel() int {
	return c.level
}

func (c *CodeIndenter) IncreaseIndent() {
	c.level++
}

func (c *CodeIndenter) DecreaseIndent() {
	if c.level > 0 {
		c.level--
	}
}
