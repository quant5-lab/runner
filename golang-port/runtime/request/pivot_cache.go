package request

import (
	"fmt"

	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ta"
)

/* PivotResultCache stores computed pivot arrays for reuse (DRY) */
type PivotResultCache struct {
	cache map[string][]float64
}

/* NewPivotResultCache creates a new result cache */
func NewPivotResultCache() *PivotResultCache {
	return &PivotResultCache{
		cache: make(map[string][]float64),
	}
}

/* ComputeOrRetrieve calculates pivot values if not cached, returns cached otherwise */
func (c *PivotResultCache) ComputeOrRetrieve(
	pivotType PivotFunctionType,
	source []float64,
	leftBars int,
	rightBars int,
) []float64 {
	key := c.buildCacheKey(pivotType, leftBars, rightBars, len(source))

	if cached, found := c.cache[key]; found {
		return cached
	}

	result := c.computePivotArray(pivotType, source, leftBars, rightBars)
	c.cache[key] = result
	return result
}

/* Clear removes all cached results */
func (c *PivotResultCache) Clear() {
	c.cache = make(map[string][]float64)
}

/* buildCacheKey generates unique key for pivot computation parameters */
func (c *PivotResultCache) buildCacheKey(
	pivotType PivotFunctionType,
	leftBars int,
	rightBars int,
	sourceLen int,
) string {
	typeStr := "high"
	if pivotType == PivotTypeLow {
		typeStr = "low"
	}
	return fmt.Sprintf("pivot_%s_%d_%d_%d", typeStr, leftBars, rightBars, sourceLen)
}

/* computePivotArray executes the actual pivot calculation using runtime ta package */
func (c *PivotResultCache) computePivotArray(
	pivotType PivotFunctionType,
	source []float64,
	leftBars int,
	rightBars int,
) []float64 {
	if pivotType == PivotTypeHigh {
		return ta.Pivothigh(source, leftBars, rightBars)
	}
	return ta.Pivotlow(source, leftBars, rightBars)
}

/* ExtractSourceSeries retrieves the appropriate source series from context */
func ExtractSourceSeries(
	pivotType PivotFunctionType,
	secCtx *context.Context,
	customSource []float64,
) []float64 {
	if customSource != nil && len(customSource) > 0 {
		return customSource
	}

	if pivotType == PivotTypeHigh {
		return extractHighSeries(secCtx)
	}
	return extractLowSeries(secCtx)
}

/* extractHighSeries builds high price array from context */
func extractHighSeries(secCtx *context.Context) []float64 {
	dataLen := len(secCtx.Data)
	result := make([]float64, dataLen)

	for i := 0; i < dataLen; i++ {
		result[i] = secCtx.Data[i].High
	}

	return result
}

/* extractLowSeries builds low price array from context */
func extractLowSeries(secCtx *context.Context) []float64 {
	dataLen := len(secCtx.Data)
	result := make([]float64, dataLen)

	for i := 0; i < dataLen; i++ {
		result[i] = secCtx.Data[i].Low
	}

	return result
}
