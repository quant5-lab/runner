package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type CUMStateManager struct {
	buf      *series.Series
	computed int
}

func NewCUMStateManager(capacity int) *CUMStateManager {
	return &CUMStateManager{
		buf: series.NewSeries(max(capacity, 1)),
	}
}

func (s *CUMStateManager) ComputeAtBar(secCtx *context.Context, sourceID *ast.Identifier, barIdx int) (float64, error) {
	for s.computed <= barIdx {
		if s.computed > 0 {
			s.buf.Next()
		}

		val, err := evaluateOHLCVAtBar(sourceID, secCtx, s.computed)
		if err != nil {
			return math.NaN(), err
		}

		prev := 0.0
		if s.computed > 0 {
			prev = s.buf.Get(1)
		}

		s.buf.Set(prev + val)
		s.computed++
	}

	return s.buf.Get(s.buf.Position() - barIdx), nil
}
