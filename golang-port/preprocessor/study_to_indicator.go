package preprocessor

import "github.com/quant5-lab/runner/parser"

// StudyToIndicatorTransformer renames study() to indicator()
// This is a simple function name replacement (v4 → v5)
type StudyToIndicatorTransformer struct{}

func NewStudyToIndicatorTransformer() *StudyToIndicatorTransformer {
	return &StudyToIndicatorTransformer{}
}

func (t *StudyToIndicatorTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	visitor := &functionRenamer{mappings: map[string]string{"study": "indicator"}}
	for _, stmt := range script.Statements {
		visitor.visitStatement(stmt)
	}
	return script, nil
}
