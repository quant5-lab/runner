package security

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

// UDFFunc is a generated UDF body bound to a secondary symbol's ArrowContext.
// arrowCtx.Context.BarIndex is always the current secondary bar index when called.
type UDFFunc func(arrowCtx *context.ArrowContext) float64

// UDFBarEvaluator evaluates a UDF against a secondary symbol's bar series,
// caching each bar's result so it is computed at most once.
type UDFBarEvaluator struct {
	replayCtx *context.Context      // private cursor; shares Data slice with the source Context
	arrowCtx  *context.ArrowContext // owns local series for self-referential state
	udfFn     UDFFunc
	cache     []float64
	nextBar   int
}

// NewUDFBarEvaluator barCount must equal len(secCtx.Data).
func NewUDFBarEvaluator(barCount int, secCtx *context.Context, udfFn UDFFunc) *UDFBarEvaluator {
	replayCtx := context.NewReplayContext(secCtx)
	cache := make([]float64, barCount)
	for i := range cache {
		cache[i] = math.NaN()
	}
	return &UDFBarEvaluator{
		replayCtx: replayCtx,
		arrowCtx:  context.NewArrowContext(replayCtx),
		udfFn:     udfFn,
		cache:     cache,
	}
}

// EvaluateAtBar returns the cached UDF output for barIdx.
// expr and secCtx are unused — they satisfy the BarEvaluator interface but the
// UDF's data source is bound to replayCtx at construction time.
// Returns (NaN, nil) for any barIdx outside [0, barCount).
func (e *UDFBarEvaluator) EvaluateAtBar(_ ast.Expression, _ *context.Context, barIdx int) (float64, error) {
	lastBar := len(e.cache) - 1
	for e.nextBar <= barIdx && e.nextBar <= lastBar {
		e.replayCtx.BarIndex = e.nextBar
		val := e.udfFn(e.arrowCtx)
		e.cache[e.nextBar] = val
		if e.nextBar < lastBar {
			e.arrowCtx.AdvanceAll()
		}
		e.nextBar++
	}
	if barIdx < 0 || barIdx >= len(e.cache) {
		return math.NaN(), nil
	}
	return e.cache[barIdx], nil
}
