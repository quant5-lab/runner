package request

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/runtime/context"
)

type BarEvaluator interface {
	EvaluateAtBar(expr ast.Expression, secCtx *context.Context, barIdx int) (float64, error)
}

type StreamingRequest struct {
	ctx        *context.Context
	fetcher    SecurityDataFetcher
	cache      map[string]*context.Context
	evaluator  BarEvaluator
	currentBar int
}

func NewStreamingRequest(ctx *context.Context, fetcher SecurityDataFetcher, evaluator BarEvaluator) *StreamingRequest {
	return &StreamingRequest{
		ctx:       ctx,
		fetcher:   fetcher,
		cache:     make(map[string]*context.Context),
		evaluator: evaluator,
	}
}

func (r *StreamingRequest) SecurityWithExpression(symbol, timeframe string, expr ast.Expression, lookahead bool) (float64, error) {
	cacheKey := buildSecurityKey(symbol, timeframe)

	secCtx, err := r.getOrFetchContext(cacheKey, symbol, timeframe)
	if err != nil {
		return math.NaN(), err
	}

	currentTime := r.getCurrentTime()
	secBarIdx := r.findMatchingBarIndex(secCtx, currentTime, lookahead)

	if !isValidBarIndex(secBarIdx, secCtx) {
		return math.NaN(), nil
	}

	return r.evaluator.EvaluateAtBar(expr, secCtx, secBarIdx)
}

func (r *StreamingRequest) SetCurrentBar(bar int) {
	r.currentBar = bar
}

func (r *StreamingRequest) ClearCache() {
	r.cache = make(map[string]*context.Context)
}

func (r *StreamingRequest) getOrFetchContext(cacheKey, symbol, timeframe string) (*context.Context, error) {
	if secCtx, cached := r.cache[cacheKey]; cached {
		return secCtx, nil
	}

	secCtx, err := r.fetcher.FetchData(symbol, timeframe, r.ctx.LastBarIndex()+1)
	if err != nil {
		return nil, err
	}

	r.cache[cacheKey] = secCtx
	return secCtx, nil
}

func (r *StreamingRequest) getCurrentTime() int64 {
	currentTimeObj := r.ctx.GetTime(-r.currentBar)
	return currentTimeObj.Unix()
}

func (r *StreamingRequest) findMatchingBarIndex(secCtx *context.Context, currentTime int64, lookahead bool) int {
	for i := 0; i <= secCtx.LastBarIndex(); i++ {
		barTimeObj := secCtx.GetTime(-i)
		barTime := barTimeObj.Unix()

		if barTime <= currentTime {
			return r.adjustForLookahead(i, lookahead)
		}
	}

	return -1
}

func (r *StreamingRequest) adjustForLookahead(barIdx int, lookahead bool) int {
	if lookahead {
		return barIdx + 1
	}
	return barIdx + 2
}

func buildSecurityKey(symbol, timeframe string) string {
	return fmt.Sprintf("%s:%s", symbol, timeframe)
}

func isValidBarIndex(barIdx int, secCtx *context.Context) bool {
	return barIdx >= 0 && barIdx < len(secCtx.Data)
}
