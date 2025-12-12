package preprocessor

import "github.com/quant5-lab/runner/parser"

// Transformer transforms Pine AST (v4 → v5 migrations, etc.)
// Each transformer implements a single responsibility (SOLID principle)
type Transformer interface {
	Transform(script *parser.Script) (*parser.Script, error)
}

// Pipeline orchestrates multiple transformers in sequence
// Open/Closed: add new transformers without modifying existing code
type Pipeline struct {
	transformers []Transformer
}

// NewPipeline creates an empty pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{transformers: []Transformer{}}
}

// Add appends a transformer to the pipeline (method chaining)
func (p *Pipeline) Add(t Transformer) *Pipeline {
	p.transformers = append(p.transformers, t)
	return p
}

// Run executes all transformers sequentially
func (p *Pipeline) Run(script *parser.Script) (*parser.Script, error) {
	for _, t := range p.transformers {
		var err error
		script, err = t.Transform(script)
		if err != nil {
			return nil, err
		}
	}
	return script, nil
}

// NewV4ToV5Pipeline creates a configured pipeline for Pine v4→v5 migration
func NewV4ToV5Pipeline() *Pipeline {
	return NewPipeline().
		Add(NewTANamespaceTransformer()).
		Add(NewMathNamespaceTransformer()).
		Add(NewRequestNamespaceTransformer()).
		Add(NewStudyToIndicatorTransformer())
}
