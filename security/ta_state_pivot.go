package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ta/pivot"
)

type pivotKind int

const (
	pivotKindHigh pivotKind = iota
	pivotKindLow
)

type PivotStateManager struct {
	detector   *pivot.DelayedDetector
	sourceExpr ast.Expression
	buf        forwardBuffer
	evaluator  BarEvaluator
}

func newPivotStateManager(kind pivotKind, leftBars, rightBars, capacity int, sourceExpr ast.Expression, evaluator BarEvaluator) *PivotStateManager {
	var detector *pivot.DelayedDetector
	if kind == pivotKindHigh {
		detector = pivot.NewDelayedHigh(leftBars, rightBars)
	} else {
		detector = pivot.NewDelayedLow(leftBars, rightBars)
	}
	return &PivotStateManager{
		detector:   detector,
		sourceExpr: sourceExpr,
		buf:        newForwardBuffer(capacity),
		evaluator:  evaluator,
	}
}

func (s *PivotStateManager) ComputeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return math.NaN(), nil
	}
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
	}
	extractor := s.makeExtractor(secCtx)
	s.buf.advanceTo(barIdx, func(bar int) float64 {
		return s.detector.DetectAtCurrentBar(bar, extractor)
	})
	return s.buf.at(barIdx), nil
}

func (s *PivotStateManager) makeExtractor(secCtx *context.Context) pivot.ValueExtractor {
	return func(idx int) float64 {
		val, err := s.evaluator.EvaluateAtBar(s.sourceExpr, secCtx, idx)
		if err != nil {
			return math.NaN()
		}
		return val
	}
}
