package codegen

type PreambleProvider interface {
	GetPreamble() string
}

type PreambleExtractor struct{}

func NewPreambleExtractor() *PreambleExtractor {
	return &PreambleExtractor{}
}

func (e *PreambleExtractor) ExtractFromAccessor(accessor AccessGenerator) string {
	if provider, ok := accessor.(PreambleProvider); ok {
		return provider.GetPreamble()
	}
	return ""
}
