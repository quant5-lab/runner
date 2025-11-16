package security

import (
	"fmt"

	"github.com/borisquantlab/pinescript-go/ast"
	"github.com/borisquantlab/pinescript-go/datafetcher"
	"github.com/borisquantlab/pinescript-go/runtime/context"
)

/* SecurityPrefetcher orchestrates the security() data prefetch workflow:
 * 1. Analyze AST for security() calls
 * 2. Deduplicate requests (same symbol+timeframe)
 * 3. Fetch data via DataFetcher interface
 * 4. Evaluate expressions in security contexts
 * 5. Populate SecurityCache for O(1) runtime lookups
 */
type SecurityPrefetcher struct {
	fetcher datafetcher.DataFetcher
	cache   *SecurityCache
}

/* NewSecurityPrefetcher creates prefetcher with specified fetcher implementation */
func NewSecurityPrefetcher(fetcher datafetcher.DataFetcher) *SecurityPrefetcher {
	return &SecurityPrefetcher{
		fetcher: fetcher,
		cache:   NewSecurityCache(),
	}
}

/* PrefetchRequest represents deduplicated security() call */
type PrefetchRequest struct {
	Symbol      string
	Timeframe   string
	Expressions map[string]ast.Expression // "sma20" -> ta.sma(close, 20)
}

/* Prefetch executes complete workflow: analyze → fetch → evaluate → cache */
func (p *SecurityPrefetcher) Prefetch(program *ast.Program, limit int) error {
	/* Step 1: Analyze AST for security() calls */
	calls := AnalyzeAST(program)
	if len(calls) == 0 {
		return nil // No security() calls - skip prefetch
	}

	/* Step 2: Deduplicate requests (group by symbol:timeframe) */
	requests := p.deduplicateCalls(calls)

	/* Step 3: Fetch data and evaluate expressions */
	for key, req := range requests {
		/* Fetch OHLCV data for symbol+timeframe */
		ohlcvData, err := p.fetcher.Fetch(req.Symbol, req.Timeframe, limit)
		if err != nil {
			return fmt.Errorf("fetch %s:%s: %w", req.Symbol, req.Timeframe, err)
		}

		/* Create security context from fetched data */
		secCtx := context.New(req.Symbol, req.Timeframe, len(ohlcvData))
		for _, bar := range ohlcvData {
			secCtx.AddBar(bar)
		}

		/* Create cache entry with context */
		entry := &CacheEntry{
			Context:     secCtx,
			Expressions: make(map[string][]float64),
		}

		/* Store entry in cache */
		p.cache.Set(req.Symbol, req.Timeframe, entry)

		/* Evaluate all expressions for this symbol+timeframe */
		for exprName, exprAST := range req.Expressions {
			values, err := EvaluateExpression(exprAST, secCtx)
			if err != nil {
				return fmt.Errorf("evaluate %s %s: %w", key, exprName, err)
			}

			/* Store evaluated expression in cache entry */
			err = p.cache.SetExpression(req.Symbol, req.Timeframe, exprName, values)
			if err != nil {
				return fmt.Errorf("cache expression %s %s: %w", key, exprName, err)
			}
		}
	}

	return nil
}

/* GetCache returns the populated SecurityCache for runtime lookups */
func (p *SecurityPrefetcher) GetCache() *SecurityCache {
	return p.cache
}

/* deduplicateCalls groups security calls by symbol:timeframe */
func (p *SecurityPrefetcher) deduplicateCalls(calls []SecurityCall) map[string]*PrefetchRequest {
	requests := make(map[string]*PrefetchRequest)

	for _, call := range calls {
		key := fmt.Sprintf("%s:%s", call.Symbol, call.Timeframe)

		/* Get or create request for this symbol+timeframe */
		req, exists := requests[key]
		if !exists {
			req = &PrefetchRequest{
				Symbol:      call.Symbol,
				Timeframe:   call.Timeframe,
				Expressions: make(map[string]ast.Expression),
			}
			requests[key] = req
		}

		/* Add expression to request (use exprName as key) */
		if call.ExprName != "" {
			req.Expressions[call.ExprName] = call.Expression
		}
	}

	return requests
}
