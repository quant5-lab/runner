package request

import rtcontext "github.com/quant5-lab/runner/runtime/context"

/* SecurityBarMapperAligner adapts SecurityBarMapper to BarAligner interface */
type SecurityBarMapperAligner struct {
	mapper    *SecurityBarMapper
	lookahead bool
}

func NewSecurityBarMapperAligner(mapper *SecurityBarMapper, lookahead bool) *SecurityBarMapperAligner {
	return &SecurityBarMapperAligner{
		mapper:    mapper,
		lookahead: lookahead,
	}
}

func (a *SecurityBarMapperAligner) AlignToParent(childBarIdx int) int {
	return a.mapper.FindDailyBarIndex(childBarIdx, a.lookahead)
}

func (a *SecurityBarMapperAligner) AlignToChild(parentBarIdx int) int {
	return -1
}

var _ rtcontext.BarAligner = (*SecurityBarMapperAligner)(nil)
