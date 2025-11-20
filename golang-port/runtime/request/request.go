package request

import (
	"fmt"
	"math"

	"github.com/borisquantlab/pinescript-go/runtime/context"
)

/* Lookahead constants */
const (
	LookaheadOn  = "barmerge.lookahead_on"
	LookaheadOff = "barmerge.lookahead_off"
)

/* Gaps constants */
const (
	GapsOn  = "barmerge.gaps_on"
	GapsOff = "barmerge.gaps_off"
)

/* SecurityDataFetcher interface for fetching multi-timeframe data */
type SecurityDataFetcher interface {
	/* FetchData fetches OHLCV data for symbol and timeframe */
	FetchData(symbol, timeframe string, limit int) (*context.Context, error)
}

/* Request implements request.security() for multi-timeframe data */
type Request struct {
	ctx        *context.Context
	fetcher    SecurityDataFetcher
	cache      map[string]*context.Context
	exprCache  map[string][]float64
	currentBar int
}

/* NewRequest creates a new request handler */
func NewRequest(ctx *context.Context, fetcher SecurityDataFetcher) *Request {
	return &Request{
		ctx:       ctx,
		fetcher:   fetcher,
		cache:     make(map[string]*context.Context),
		exprCache: make(map[string][]float64),
	}
}

/* Security fetches data from another timeframe/symbol and evaluates expression */
func (r *Request) Security(symbol, timeframe string, exprFunc func(*context.Context) []float64, lookahead bool) (float64, error) {
	cacheKey := fmt.Sprintf("%s:%s", symbol, timeframe)

	// Check context cache
	secCtx, cached := r.cache[cacheKey]
	if !cached {
		// Fetch data for security timeframe
		var err error
		secCtx, err = r.fetcher.FetchData(symbol, timeframe, r.ctx.LastBarIndex()+1)
		if err != nil {
			return math.NaN(), err
		}
		r.cache[cacheKey] = secCtx
	}

	// Check expression cache
	exprValues, exprCached := r.exprCache[cacheKey]
	if !exprCached {
		// Calculate expression in security context
		exprValues = exprFunc(secCtx)
		r.exprCache[cacheKey] = exprValues
	}

	// Get current bar time from main context
	currentTimeObj := r.ctx.GetTime(-r.currentBar)
	currentTime := currentTimeObj.Unix()

	// Find matching bar in security context
	secIdx := r.findMatchingBar(secCtx, currentTime, lookahead)
	if secIdx < 0 || secIdx >= len(exprValues) {
		return math.NaN(), nil
	}

	return exprValues[secIdx], nil
}

/* SecurityLegacy for backward compatibility with tests */
func (r *Request) SecurityLegacy(symbol, timeframe string, expression []float64, lookahead bool) (float64, error) {
	// Wrap pre-calculated array in function
	exprFunc := func(secCtx *context.Context) []float64 {
		return expression
	}
	return r.Security(symbol, timeframe, exprFunc, lookahead)
}

/* SetCurrentBar updates current bar index for context alignment */
func (r *Request) SetCurrentBar(bar int) {
	r.currentBar = bar
}

/* ClearCache clears security data and expression caches */
func (r *Request) ClearCache() {
	r.cache = make(map[string]*context.Context)
	r.exprCache = make(map[string][]float64)
}

/* findMatchingBar finds the bar index in security context that matches current time */
func (r *Request) findMatchingBar(secCtx *context.Context, currentTime int64, lookahead bool) int {
	// Simplified: find bar where time <= currentTime
	// With lookahead: use next bar
	// Without lookahead: use confirmed bar (2 bars back)

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
