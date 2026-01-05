package security

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/series"
)

type SeriesCachingEvaluator struct {
	delegate      BarEvaluator
	seriesCache   map[string]*series.Series
	contextHashes map[*context.Context]string
}

func NewSeriesCachingEvaluator(delegate BarEvaluator) *SeriesCachingEvaluator {
	return &SeriesCachingEvaluator{
		delegate:      delegate,
		seriesCache:   make(map[string]*series.Series),
		contextHashes: make(map[*context.Context]string),
	}
}

func (e *SeriesCachingEvaluator) EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error) {
	// Delegate to StreamingRequest which handles offset extraction via HistoricalOffsetExtractor
	// No offset manipulation at this layer - series caching happens in StreamingRequest.getOrBuildSeries()
	return e.delegate.EvaluateAtBar(expr, secCtx, barIdx)
}

// extractMemberExpression unwraps CallExpression layers to find nested MemberExpression
func (e *SeriesCachingEvaluator) extractMemberExpression(expr ast.Expression) *ast.MemberExpression {
	if memberExpr, ok := expr.(*ast.MemberExpression); ok {
		return memberExpr
	}

	if callExpr, ok := expr.(*ast.CallExpression); ok {
		for _, arg := range callExpr.Arguments {
			if memberExpr, ok := arg.(*ast.MemberExpression); ok {
				// Check if it has numeric offset property
				if _, isLiteral := memberExpr.Property.(*ast.Literal); isLiteral {
					return memberExpr
				}
			}
		}
	}

	return nil
}

// removeOffset creates new expression with offset removed from MemberExpression
// For fixnan(pivothigh()[1]), returns fixnan(pivothigh())
func (e *SeriesCachingEvaluator) removeOffset(fullExpr ast.Expression, memberExpr *ast.MemberExpression, baseExpr ast.Expression) ast.Expression {
	// If expr is directly the MemberExpression, return base
	if _, ok := fullExpr.(*ast.MemberExpression); ok {
		return baseExpr
	}

	// If expr is CallExpression wrapping MemberExpression, replace argument
	if callExpr, ok := fullExpr.(*ast.CallExpression); ok {
		// Create new CallExpression with baseExpr instead of memberExpr
		newCall := &ast.CallExpression{
			Callee:    callExpr.Callee,
			Arguments: make([]ast.Expression, len(callExpr.Arguments)),
		}
		copy(newCall.Arguments, callExpr.Arguments)

		// Find and replace the MemberExpression argument with baseExpr
		for i, arg := range newCall.Arguments {
			if arg == memberExpr {
				newCall.Arguments[i] = baseExpr
				break
			}
		}
		return newCall
	}

	// Fallback: return baseExpr
	return baseExpr
}

func (e *SeriesCachingEvaluator) getOrBuildSeries(expr ast.Expression, secCtx *context.Context) (*series.Series, error) {
	ctxHash := e.getContextHash(secCtx)
	cacheKey := fmt.Sprintf("%s:%p", ctxHash, expr)

	if cached, found := e.seriesCache[cacheKey]; found {
		fmt.Printf("[CACHE] Using cached series for key=%s\n", cacheKey)
		return cached, nil
	}

	fmt.Printf("[CACHE] Building NEW series for key=%s, secCtx.Data len=%d, expr type=%T\n", cacheKey, len(secCtx.Data), expr)

	seriesBuffer := series.NewSeries(len(secCtx.Data))
	nanCount := 0
	validCount := 0
	firstValid := -1
	lastValid := -1

	for barIdx := 0; barIdx < len(secCtx.Data); barIdx++ {
		value, err := e.delegate.EvaluateAtBar(expr, secCtx, barIdx)
		if err != nil {
			fmt.Printf("[CACHE] ERROR at barIdx=%d: %v\n", barIdx, err)
			return nil, err
		}

		seriesBuffer.Set(value)
		if math.IsNaN(value) {
			nanCount++
		} else {
			validCount++
			if firstValid == -1 {
				firstValid = barIdx
			}
			lastValid = barIdx
		}
		if barIdx < len(secCtx.Data)-1 {
			seriesBuffer.Next()
		}
	}

	fmt.Printf("[CACHE] Series built: %d NaN, %d valid values (first valid: bar %d, last valid: bar %d)\n", nanCount, validCount, firstValid, lastValid)
	e.seriesCache[cacheKey] = seriesBuffer
	return seriesBuffer, nil
}

func (e *SeriesCachingEvaluator) getContextHash(secCtx *context.Context) string {
	if hash, found := e.contextHashes[secCtx]; found {
		return hash
	}

	hash := fmt.Sprintf("%p", secCtx)
	e.contextHashes[secCtx] = hash
	return hash
}
