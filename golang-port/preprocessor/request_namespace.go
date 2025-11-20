package preprocessor

import "github.com/borisquantlab/pinescript-go/parser"

// RequestNamespaceTransformer adds request. prefix to data request functions
// Examples: security() → request.security(), financial() → request.financial()
type RequestNamespaceTransformer struct {
	mappings map[string]string
}

func NewRequestNamespaceTransformer() *RequestNamespaceTransformer {
	return &RequestNamespaceTransformer{
		mappings: map[string]string{
			"security":  "request.security",
			"financial": "request.financial",
			"quandl":    "request.quandl",
			"splits":    "request.splits",
			"dividends": "request.dividends",
			"earnings":  "request.earnings",
		},
	}
}

func (t *RequestNamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	visitor := &functionRenamer{mappings: t.mappings}
	for _, stmt := range script.Statements {
		visitor.visitStatement(stmt)
	}
	return script, nil
}
