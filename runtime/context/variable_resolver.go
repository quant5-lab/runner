package context

import "github.com/quant5-lab/runner/runtime/series"

type VariableResolutionResult struct {
	Series       *series.Series
	SourceBarIdx int
	Found        bool
}

type VariableResolver interface {
	Resolve(name string, targetBarIdx int) VariableResolutionResult
}

type RecursiveResolver struct {
	localRegistry SeriesRegistry
	barAligner    BarAligner
	parent        VariableResolver
}

func NewRecursiveResolver(
	localRegistry SeriesRegistry,
	barAligner BarAligner,
	parent VariableResolver,
) *RecursiveResolver {
	return &RecursiveResolver{
		localRegistry: localRegistry,
		barAligner:    barAligner,
		parent:        parent,
	}
}

func (r *RecursiveResolver) Resolve(name string, targetBarIdx int) VariableResolutionResult {
	if series, found := r.localRegistry.Get(name); found {
		return VariableResolutionResult{
			Series:       series,
			SourceBarIdx: targetBarIdx,
			Found:        true,
		}
	}

	if r.parent == nil {
		return VariableResolutionResult{Found: false}
	}

	parentBarIdx := r.barAligner.AlignToParent(targetBarIdx)
	return r.parent.Resolve(name, parentBarIdx)
}
