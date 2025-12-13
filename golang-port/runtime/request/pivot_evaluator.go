package request

import (
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

/* PivotEvaluator handles pivot function evaluation within security() context (SRP) */
type PivotEvaluator struct {
	detector *PivotDetector
	cache    *PivotResultCache
}

/* NewPivotEvaluator creates evaluator with detector and cache dependencies */
func NewPivotEvaluator() *PivotEvaluator {
	return &PivotEvaluator{
		detector: NewPivotDetector(),
		cache:    NewPivotResultCache(),
	}
}

/* TryEvaluate attempts to evaluate expression as pivot function call */
func (e *PivotEvaluator) TryEvaluate(
	expr ast.Expression,
	secCtx *context.Context,
	barIdx int,
) (float64, bool) {
	pivotInfo, detected := e.detector.DetectPivotCall(expr)
	if !detected {
		return math.NaN(), false
	}

	sourceSeries := ExtractSourceSeries(pivotInfo.Type, secCtx, nil)

	pivotArray := e.cache.ComputeOrRetrieve(
		pivotInfo.Type,
		sourceSeries,
		pivotInfo.LeftBars,
		pivotInfo.RightBars,
	)

	targetIdx := barIdx
	if pivotInfo.HasOffset {
		targetIdx = barIdx + pivotInfo.Offset
	}

	if targetIdx < 0 || targetIdx >= len(pivotArray) {
		return math.NaN(), true
	}

	return pivotArray[targetIdx], true
}

/* ClearCache forwards cache clearing to internal cache */
func (e *PivotEvaluator) ClearCache() {
	e.cache.Clear()
}
