package codegen

type StatefulTAGenerator struct {
	builder *StatefulIndicatorBuilder
}

func NewStatefulRMAGenerator(varName string, period int, accessor AccessGenerator, context StatefulIndicatorContext) *StatefulTAGenerator {
	builder := NewStatefulIndicatorBuilder(
		"ta.rma",
		varName,
		period,
		accessor,
		false,
		context,
	)
	return &StatefulTAGenerator{builder: builder}
}

func NewStatefulEMAGenerator(varName string, period int, accessor AccessGenerator, context StatefulIndicatorContext) *StatefulTAGenerator {
	builder := NewStatefulIndicatorBuilder(
		"ta.ema",
		varName,
		period,
		accessor,
		false,
		context,
	)
	return &StatefulTAGenerator{builder: builder}
}

func (g *StatefulTAGenerator) GenerateRMA() string {
	return g.builder.BuildRMA()
}

func (g *StatefulTAGenerator) GenerateEMA() string {
	return g.builder.BuildEMA()
}
