package security

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
	"github.com/quant5-lab/runner/datafetcher"
	"github.com/quant5-lab/runner/runtime/context"
	"github.com/quant5-lab/runner/runtime/ticker"
)

/* SecurityPrefetcher orchestrates the security() data prefetch workflow:
 * 1. Analyze AST for security() calls
 * 2. Deduplicate requests (same symbol+timeframe)
 * 3. Fetch OHLCV data via DataFetcher interface
 * 4. Store contexts in cache for O(1) runtime access
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

/* PrefetchRequest represents deduplicated security() call with transformation */
type PrefetchRequest struct {
	CacheKey    string                    // Full symbol (may include modifier prefix)
	BaseSymbol  string                    // Base symbol for fetching
	Timeframe   string                    // Timeframe
	Transformer ticker.BarTransformer     // Transformer for bar conversion
	Expressions map[string]ast.Expression // "sma20" -> ta.sma(close, 20)
}

/* Prefetch executes complete workflow: analyze → fetch → transform → cache contexts */
func (p *SecurityPrefetcher) Prefetch(program *ast.Program, limit int) error {
	calls := AnalyzeAST(program)
	if len(calls) == 0 {
		return nil
	}

	requests := p.deduplicateCallsWithModifiers(calls)

	for _, req := range requests {
		ohlcvData, err := p.fetcher.Fetch(req.BaseSymbol, req.Timeframe, limit)
		if err != nil {
			return fmt.Errorf("fetch %s:%s: %w", req.BaseSymbol, req.Timeframe, err)
		}

		transformedData := req.Transformer.Transform(ohlcvData).Bars

		secCtx := context.New(req.CacheKey, req.Timeframe, len(transformedData))
		for _, bar := range transformedData {
			secCtx.AddBar(bar)
		}

		entry := &CacheEntry{
			Context: secCtx,
		}
		p.cache.Set(req.CacheKey, req.Timeframe, entry)
	}

	return nil
}

/* GetCache returns the populated SecurityCache for runtime lookups */
func (p *SecurityPrefetcher) GetCache() *SecurityCache {
	return p.cache
}

/* deduplicateCallsWithModifiers groups security calls by symbol:timeframe with transformation */
func (p *SecurityPrefetcher) deduplicateCallsWithModifiers(calls []SecurityCall) map[string]*PrefetchRequest {
	requests := make(map[string]*PrefetchRequest)
	extractor := NewSymbolExtractor()

	for _, call := range calls {
		baseSymbol, modifierType := extractor.Extract(call.SymbolExpr)
		if baseSymbol == "" {
			baseSymbol = call.Symbol
		}

		cacheKey := call.Symbol
		key := fmt.Sprintf("%s:%s:%s", cacheKey, baseSymbol, call.Timeframe)

		req, exists := requests[key]
		if !exists {
			req = &PrefetchRequest{
				CacheKey:    cacheKey,
				BaseSymbol:  baseSymbol,
				Timeframe:   call.Timeframe,
				Transformer: ticker.NewTransformer(modifierType),
				Expressions: make(map[string]ast.Expression),
			}
			requests[key] = req
		}

		if call.ExprName != "" {
			req.Expressions[call.ExprName] = call.Expression
		}
	}

	return requests
}
