package codegen

type ArrowVarInitResult struct {
	Preamble   string
	Assignment string
}

func NewArrowVarInitResult(preamble, assignment string) *ArrowVarInitResult {
	return &ArrowVarInitResult{
		Preamble:   preamble,
		Assignment: assignment,
	}
}

func (r *ArrowVarInitResult) HasPreamble() bool {
	return r.Preamble != ""
}

func (r *ArrowVarInitResult) CombinedCode() string {
	return r.Preamble + r.Assignment
}
