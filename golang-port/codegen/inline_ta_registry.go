package codegen

type InlineTAIIFERegistry struct {
	generators map[string]InlineTAIIFEGenerator
}

func NewInlineTAIIFERegistry() *InlineTAIIFERegistry {
	registry := &InlineTAIIFERegistry{
		generators: make(map[string]InlineTAIIFEGenerator),
	}
	registry.registerDefaults()
	return registry
}

func (r *InlineTAIIFERegistry) registerDefaults() {
	r.Register("ta.sma", &SMAIIFEGenerator{})
	r.Register("sma", &SMAIIFEGenerator{})
	r.Register("ta.ema", &EMAIIFEGenerator{})
	r.Register("ema", &EMAIIFEGenerator{})
	r.Register("ta.rma", &RMAIIFEGenerator{})
	r.Register("rma", &RMAIIFEGenerator{})
	r.Register("ta.wma", &WMAIIFEGenerator{})
	r.Register("wma", &WMAIIFEGenerator{})
	r.Register("ta.stdev", &STDEVIIFEGenerator{})
	r.Register("stdev", &STDEVIIFEGenerator{})
}

func (r *InlineTAIIFERegistry) Register(name string, generator InlineTAIIFEGenerator) {
	r.generators[name] = generator
}

func (r *InlineTAIIFERegistry) IsSupported(funcName string) bool {
	_, exists := r.generators[funcName]
	return exists
}

func (r *InlineTAIIFERegistry) Generate(funcName string, accessor AccessGenerator, period int) (string, bool) {
	generator, exists := r.generators[funcName]
	if !exists {
		return "", false
	}
	return generator.Generate(accessor, period), true
}
