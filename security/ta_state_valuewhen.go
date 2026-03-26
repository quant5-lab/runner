package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/value"
)

// matchRing provides O(1) push and O(1) nth-most-recent lookup for the
// occurrence-indexed valuewhen semantics.
type matchRing struct {
	values []float64
	head   int
	count  int
}

func newMatchRing(capacity int) *matchRing {
	return &matchRing{values: make([]float64, capacity)}
}

func (r *matchRing) push(v float64) {
	r.values[r.head] = v
	r.head = (r.head + 1) % len(r.values)
	if r.count < len(r.values) {
		r.count++
	}
}

// nthMostRecent returns (value, true) where 0 is the latest push,
// or (NaN, false) when fewer than n+1 values have been pushed.
func (r *matchRing) nthMostRecent(n int) (float64, bool) {
	if n >= r.count {
		return math.NaN(), false
	}
	idx := (r.head - 1 - n + len(r.values)*2) % len(r.values)
	return r.values[idx], true
}

type ValuewhenStateManager struct {
	condExpr   ast.Expression
	srcExpr    ast.Expression
	occurrence int
	ring       *matchRing
	buf        forwardBuffer
	evaluator  BarEvaluator
}

func newValuewhenStateManager(condExpr, srcExpr ast.Expression, occurrence, capacity int, evaluator BarEvaluator) *ValuewhenStateManager {
	return &ValuewhenStateManager{
		condExpr:   condExpr,
		srcExpr:    srcExpr,
		occurrence: occurrence,
		ring:       newMatchRing(occurrence + 1),
		buf:        newForwardBuffer(capacity),
		evaluator:  evaluator,
	}
}

func (s *ValuewhenStateManager) ComputeAtBar(secCtx *context.Context, barIdx int) (float64, error) {
	if barIdx < 0 || barIdx >= len(secCtx.Data) {
		return math.NaN(), nil
	}
	if s.buf.growsFor(len(secCtx.Data)) {
		s.buf.reallocate(len(secCtx.Data))
		s.ring = newMatchRing(s.occurrence + 1)
	}
	s.buf.advanceTo(barIdx, func(bar int) float64 {
		return s.evalBar(secCtx, bar)
	})
	return s.buf.at(barIdx), nil
}

func (s *ValuewhenStateManager) evalBar(secCtx *context.Context, bar int) float64 {
	condVal, err := s.evaluator.EvaluateAtBar(s.condExpr, secCtx, bar)
	if err != nil {
		return math.NaN()
	}
	if value.IsTrue(condVal) {
		srcVal, err := s.evaluator.EvaluateAtBar(s.srcExpr, secCtx, bar)
		if err != nil {
			srcVal = math.NaN()
		}
		s.ring.push(srcVal)
	}
	if result, ok := s.ring.nthMostRecent(s.occurrence); ok {
		return result
	}
	return math.NaN()
}
