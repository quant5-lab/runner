package preprocessor

import "github.com/quant5-lab/runner/parser"

/* Pine v4→v5: security() → request.security() */
type RequestNamespaceTransformer struct {
	base *NamespaceTransformer
}

func NewRequestNamespaceTransformer() *RequestNamespaceTransformer {
	mappings := map[string]string{
		"security":  "request.security",
		"financial": "request.financial",
		"quandl":    "request.quandl",
		"splits":    "request.splits",
		"dividends": "request.dividends",
		"earnings":  "request.earnings",
	}

	return &RequestNamespaceTransformer{
		base: NewNamespaceTransformer(mappings),
	}
}

func (t *RequestNamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	return t.base.Transform(script)
}
