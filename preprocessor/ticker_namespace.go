package preprocessor

import "github.com/quant5-lab/runner/parser"

/* Pine v4→v5: heikinashi() → ticker.heikinashi(), renko() → ticker.renko(), etc. */
type TickerNamespaceTransformer struct {
	base *NamespaceTransformer
}

func NewTickerNamespaceTransformer() *TickerNamespaceTransformer {
	mappings := map[string]string{
		"heikinashi":  "ticker.heikinashi",
		"heikenashi":  "ticker.heikinashi",
		"renko":       "ticker.renko",
		"kagi":        "ticker.kagi",
		"linebreak":   "ticker.linebreak",
		"pointfigure": "ticker.pointfigure",
	}

	return &TickerNamespaceTransformer{
		base: NewNamespaceTransformer(mappings),
	}
}

func (t *TickerNamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	return t.base.Transform(script)
}
