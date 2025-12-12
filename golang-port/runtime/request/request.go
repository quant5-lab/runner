package request

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/runtime/context"
)

const (
	LookaheadOn  = "barmerge.lookahead_on"
	LookaheadOff = "barmerge.lookahead_off"
)

const (
	GapsOn  = "barmerge.gaps_on"
	GapsOff = "barmerge.gaps_off"
)

type SecurityDataFetcher interface {
	FetchData(symbol, timeframe string, limit int) (*context.Context, error)
}

type Request struct {
	ctx        *context.Context
	fetcher    SecurityDataFetcher
	cache      map[string]*context.Context
	exprCache  map[string][]float64
	currentBar int
}

func NewRequest(ctx *context.Context, fetcher SecurityDataFetcher) *Request {
	return &Request{
		ctx:       ctx,
		fetcher:   fetcher,
		cache:     make(map[string]*context.Context),
		exprCache: make(map[string][]float64),
	}
}

func (r *Request) Security(symbol, timeframe string, exprFunc func(*context.Context) []float64, lookahead bool) (float64, error) {
	cacheKey := fmt.Sprintf("%s:%s", symbol, timeframe)

	secCtx, cached := r.cache[cacheKey]
	if !cached {
		var err error
		secCtx, err = r.fetcher.FetchData(symbol, timeframe, r.ctx.LastBarIndex()+1)
		if err != nil {
			return math.NaN(), err
		}
		r.cache[cacheKey] = secCtx
	}

	exprValues, exprCached := r.exprCache[cacheKey]
	if !exprCached {
		exprValues = exprFunc(secCtx)
		r.exprCache[cacheKey] = exprValues
	}

	currentTimeObj := r.ctx.GetTime(-r.currentBar)
	currentTime := currentTimeObj.Unix()

	secIdx := r.findMatchingBar(secCtx, currentTime, lookahead)
	if secIdx < 0 || secIdx >= len(exprValues) {
		return math.NaN(), nil
	}

	return exprValues[secIdx], nil
}

func (r *Request) SecurityLegacy(symbol, timeframe string, expression []float64, lookahead bool) (float64, error) {
	exprFunc := func(secCtx *context.Context) []float64 {
		return expression
	}
	return r.Security(symbol, timeframe, exprFunc, lookahead)
}

func (r *Request) SetCurrentBar(bar int) {
	r.currentBar = bar
}

func (r *Request) ClearCache() {
	r.cache = make(map[string]*context.Context)
	r.exprCache = make(map[string][]float64)
}

func (r *Request) findMatchingBar(secCtx *context.Context, currentTime int64, lookahead bool) int {
	for i := 0; i <= secCtx.LastBarIndex(); i++ {
		barTimeObj := secCtx.GetTime(-i)
		barTime := barTimeObj.Unix()
		if barTime <= currentTime {
			if lookahead {
				return i + 1
			}
			return i + 2
		}
	}

	return -1
}
