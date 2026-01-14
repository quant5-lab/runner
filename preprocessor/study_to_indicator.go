package preprocessor

import "github.com/quant5-lab/runner/parser"

/* Pine v4→v5: study() → indicator() */
type StudyToIndicatorTransformer struct {
	base *SimpleRenameTransformer
}

func NewStudyToIndicatorTransformer() *StudyToIndicatorTransformer {
	mappings := map[string]string{"study": "indicator"}

	return &StudyToIndicatorTransformer{
		base: NewSimpleRenameTransformer(mappings),
	}
}

func (t *StudyToIndicatorTransformer) Transform(script *parser.Script) (*parser.Script, error) {
	return t.base.Transform(script)
}
