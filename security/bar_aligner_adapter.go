package security

import rtcontext "github.com/quant5-lab/runner/runtime/context"

type BarIndexMapperAligner struct {
	mapper *BarIndexMapper
}

func NewBarIndexMapperAligner(mapper *BarIndexMapper) *BarIndexMapperAligner {
	return &BarIndexMapperAligner{mapper: mapper}
}

func (a *BarIndexMapperAligner) AlignToParent(childBarIdx int) int {
	return a.mapper.GetMainBarIndexForSecurityBar(childBarIdx)
}

func (a *BarIndexMapperAligner) AlignToChild(parentBarIdx int) int {
	return -1
}

var _ rtcontext.BarAligner = (*BarIndexMapperAligner)(nil)
