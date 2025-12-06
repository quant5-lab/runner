package preprocessor

import "github.com/quant5-lab/runner/parser"

// MathNamespaceTransformer adds math. prefix to mathematical functions
// Examples: abs() → math.abs(), max() → math.max(), sqrt() → math.sqrt()
type MathNamespaceTransformer struct {
	mappings map[string]string
}

func NewMathNamespaceTransformer() *MathNamespaceTransformer {
	return &MathNamespaceTransformer{
		mappings: map[string]string{
			"abs":              "math.abs",
			"acos":             "math.acos",
			"asin":             "math.asin",
			"atan":             "math.atan",
			"avg":              "math.avg",
			"ceil":             "math.ceil",
			"cos":              "math.cos",
			"exp":              "math.exp",
			"floor":            "math.floor",
			"log":              "math.log",
			"log10":            "math.log10",
			"max":              "math.max",
			"min":              "math.min",
			"pow":              "math.pow",
			"random":           "math.random",
			"round":            "math.round",
			"round_to_mintick": "math.round_to_mintick",
			"sign":             "math.sign",
			"sin":              "math.sin",
			"sqrt":             "math.sqrt",
			"tan":              "math.tan",
			"todegrees":        "math.todegrees",
			"toradians":        "math.toradians",
		},
	}
}

func (t *MathNamespaceTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	visitor := &functionRenamer{mappings: t.mappings}
	for _, stmt := range script.Statements {
		visitor.visitStatement(stmt)
	}
	return script, nil
}
