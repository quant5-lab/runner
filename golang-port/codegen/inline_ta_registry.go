package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/codegen/series_naming"
)

type InlineTAIIFEGenerator interface {
	Generate(accessor AccessGenerator, period int, sourceHash string) string
}

type InlineTAIIFERegistry struct {
	generators map[string]InlineTAIIFEGenerator
}

func NewInlineTAIIFERegistry() *InlineTAIIFERegistry {
	r := &InlineTAIIFERegistry{generators: make(map[string]InlineTAIIFEGenerator)}
	r.registerDefaults()
	return r
}

func (r *InlineTAIIFERegistry) registerDefaults() {
	windowNamer := series_naming.NewWindowBasedNamer()
	statefulNamer := series_naming.NewStatefulIndicatorNamer()

	r.Register("ta.sma", &SMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("sma", &SMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.wma", &WMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("wma", &WMAIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.stdev", &STDEVIIFEGenerator{namingStrategy: windowNamer})
	r.Register("stdev", &STDEVIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.highest", &HighestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("highest", &HighestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.lowest", &LowestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("lowest", &LowestIIFEGenerator{namingStrategy: windowNamer})
	r.Register("ta.change", &ChangeIIFEGenerator{namingStrategy: windowNamer})
	r.Register("change", &ChangeIIFEGenerator{namingStrategy: windowNamer})

	r.Register("ta.ema", &EMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("ema", &EMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("ta.rma", &RMAIIFEGenerator{namingStrategy: statefulNamer})
	r.Register("rma", &RMAIIFEGenerator{namingStrategy: statefulNamer})
}

func (r *InlineTAIIFERegistry) Register(name string, generator InlineTAIIFEGenerator) {
	r.generators[name] = generator
}

func (r *InlineTAIIFERegistry) IsSupported(funcName string) bool {
	_, ok := r.generators[funcName]
	return ok
}

func (r *InlineTAIIFERegistry) Generate(funcName string, accessor AccessGenerator, period int, sourceHash string) (string, bool) {
	gen, ok := r.generators[funcName]
	if !ok {
		return "", false
	}
	return gen.Generate(accessor, period, sourceHash), true
}

// Generators

type SMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type EMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type RMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type WMAIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type STDEVIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type HighestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type LowestIIFEGenerator struct{ namingStrategy series_naming.Strategy }

type ChangeIIFEGenerator struct{ namingStrategy series_naming.Strategy }

func (g *SMAIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("sma", period, sourceHash)

	builder := NewTAIndicatorBuilder("ta.sma", varName, period, accessor, false)
	builder.WithAccumulator(NewSumAccumulator())
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *EMAIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("ema", period, sourceHash)

	builder := NewStatefulIndicatorBuilder("ta.ema", varName, period, accessor, false, context)
	statefulCode := builder.BuildEMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *RMAIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("rma", period, sourceHash)

	builder := NewStatefulIndicatorBuilder("ta.rma", varName, period, accessor, false, context)
	statefulCode := builder.BuildRMA()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *WMAIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	context := NewArrowFunctionIndicatorContext()
	varName := g.namingStrategy.GenerateName("wma", period, sourceHash)

	builder := NewTAIndicatorBuilder("ta.wma", varName, period, accessor, false)
	builder.WithAccumulator(NewWeightedSumAccumulator(period))
	statefulCode := builder.Build()
	seriesAccess := context.GenerateSeriesAccess(varName, 0)

	return fmt.Sprintf("func() float64 {\n\t%s\n\treturn %s\n}()", statefulCode, seriesAccess)
}

func (g *STDEVIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	body := fmt.Sprintf("sum := 0.0; for j := 0; j < %d; j++ { sum += %s }; ", period, accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("mean := sum / %d.0; ", period)
	body += fmt.Sprintf("variance := 0.0; for j := 0; j < %d; j++ { diff := %s - mean; variance += diff * diff }; ", period, accessor.GenerateLoopValueAccess("j"))
	body += fmt.Sprintf("return math.Sqrt(variance / %d.0)", period)

	return NewIIFECodeBuilder().WithWarmupCheck(period).WithBody(body).Build()
}

func (g *HighestIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	body := fmt.Sprintf("highest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j > 0; j-- { v := %s; if v > highest { highest = v } }; ", period-1, accessor.GenerateLoopValueAccess("j"))
	body += "return highest"

	return NewIIFECodeBuilder().WithWarmupCheck(period).WithBody(body).Build()
}

func (g *LowestIIFEGenerator) Generate(accessor AccessGenerator, period int, sourceHash string) string {
	body := fmt.Sprintf("lowest := %s; ", accessor.GenerateInitialValueAccess(period))
	body += fmt.Sprintf("for j := %d; j > 0; j-- { v := %s; if v < lowest { lowest = v } }; ", period-1, accessor.GenerateLoopValueAccess("j"))
	body += "return lowest"

	return NewIIFECodeBuilder().WithWarmupCheck(period).WithBody(body).Build()
}

func (g *ChangeIIFEGenerator) Generate(accessor AccessGenerator, offset int, sourceHash string) string {
	if offset <= 0 {
		offset = 1
	}

	body := fmt.Sprintf("current := %s; ", accessor.GenerateLoopValueAccess("0"))
	body += fmt.Sprintf("previous := %s; ", accessor.GenerateLoopValueAccess(fmt.Sprintf("%d", offset)))
	body += "return current - previous"

	return NewIIFECodeBuilder().WithWarmupCheck(offset + 1).WithBody(body).Build()
}
