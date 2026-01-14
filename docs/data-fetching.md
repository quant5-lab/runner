# request.security() Module Architecture

## Evidence-Based Design (from PineTS)

Analyzed legacy implementation: `/PineTS/src/utils/SecurityCallAnalyzer.class.ts` and `/PineTS/src/namespaces/PineRequest.ts`

**Proven Pattern**:
1. **Pre-analysis**: Static AST scan extracts `{symbol, timeframe, expressionName}` tuples
2. **Prefetch**: Async fetch ALL required data before strategy execution
3. **Cache**: Store fetched contexts + evaluated expressions
4. **Runtime**: Lookup cached values (zero I/O during bar loop)

## Module Structure (Go)

```
golang-port/
├── security/
│   ├── analyzer.go              # AST scanner (SRP: detect security calls)
│   ├── prefetcher.go            # Orchestrator (SRP: coordinate prefetch)
│   ├── cache.go                 # Storage (SRP: context + expression caching)
│   └── evaluator.go             # Calculator (SRP: evaluate expressions in security context)
│
├── datafetcher/
│   ├── fetcher.go               # Interface (DIP: abstract provider)
│   ├── file_fetcher.go          # Local JSON (current need)
│   └── remote_fetcher.go        # HTTP API (future extension)
│
└── runtime/request/
    └── request.go               # Runtime API (thin facade, delegates to cache)
```

## Data Flow

```
┌──────────────┐
│ Pine Source  │
└──────┬───────┘
       │
       v
┌──────────────────┐    1. Analyze AST
│ SecurityAnalyzer │───────> [{symbol, tf, expr}...]
└──────┬───────────┘
       │
       v
┌──────────────────┐    2. Fetch Data (async)
│ SecurityPrefetch │────────┐
└──────────────────┘        │
                            v
                    ┌───────────────┐
                    │ DataFetcher   │ (interface)
                    │ ┌───────────┐ │
                    │ │ FileFetch │ │ (impl: read JSON + sleep)
                    │ └───────────┘ │
                    └───────┬───────┘
                            │
                            v
                    ┌───────────────┐
                    │ SecurityCache │ {key -> Context + ExprValues}
                    └───────┬───────┘
                            │
       ┌────────────────────┘
       │
       v
┌──────────────────┐    3. Runtime Lookup (zero I/O)
│ Bar Loop         │
│  ├─ GetSecurity  │────> Cache.Get(key) -> value
│  └─ (no fetch)   │
└──────────────────┘
```

## Component Responsibilities

### 1. SecurityAnalyzer (analyzer.go)
**SRP**: Detect `request.security()` calls in AST

```go
type SecurityCall struct {
    Symbol     string
    Timeframe  string
    Expression ast.Expression // Store AST node for later evaluation
}

func AnalyzeAST(program *ast.Program) []SecurityCall
```

**Why**: Separates detection logic from fetching/caching

---

### 2. SecurityPrefetcher (prefetcher.go)
**SRP**: Orchestrate prefetch workflow

```go
type Prefetcher struct {
    fetcher DataFetcher
    cache   *SecurityCache
}

func (p *Prefetcher) Prefetch(calls []SecurityCall, mainCtx *context.Context) error
```

**Flow**:
1. Deduplicate `{symbol, timeframe}` pairs
2. Async fetch via `DataFetcher.Fetch()`
3. Create security contexts
4. Evaluate expressions (delegate to `Evaluator`)
5. Store in cache

**Why**: Single orchestrator prevents scattered coordination logic

---

### 3. DataFetcher (datafetcher/fetcher.go)
**DIP**: Abstract data source

```go
type DataFetcher interface {
    Fetch(symbol, timeframe string, limit int) ([]context.OHLCV, error)
}
```

**Implementations**:

**FileFetcher** (datafetcher/file_fetcher.go):
```go
type FileFetcher struct {
    dataDir string
    latency time.Duration // Simulate network delay
}

func (f *FileFetcher) Fetch(symbol, tf string, limit int) ([]context.OHLCV, error) {
    time.Sleep(f.latency) // Async simulation
    data := readJSON(f.dataDir + "/" + symbol + "_" + tf + ".json")
    return data, nil
}
```

**RemoteFetcher** (future - datafetcher/remote_fetcher.go):
```go
type RemoteFetcher struct {
    baseURL string
    client  *http.Client
}

func (r *RemoteFetcher) Fetch(symbol, tf string, limit int) ([]context.OHLCV, error) {
    resp := r.client.Get(r.baseURL + "/api/ohlcv?symbol=" + symbol + "&tf=" + tf)
    return parseJSON(resp.Body), nil
}
```

**Why**: Easy to swap implementations without changing consumers

---

### 4. SecurityEvaluator (evaluator.go)
**SRP**: Calculate expression in security context

```go
func EvaluateExpression(expr ast.Expression, secCtx *context.Context) ([]float64, error)
```

**Example**: `sma(close, 20)` in daily context
1. Extract `close` series from `secCtx.Data`
2. Call `ta.Sma(closeSeries, 20)`
3. Return array of values

**Why**: Isolates expression execution logic

---

### 5. SecurityCache (cache.go)
**SRP**: Store fetched contexts + evaluated expressions

```go
type CacheEntry struct {
    Context    *context.Context
    Expressions map[string][]float64 // expressionName -> values
}

type SecurityCache struct {
    entries map[string]*CacheEntry // "symbol:timeframe" -> entry
}
```

**Why**: Single source of truth for cached data

---

### 6. Request.Security (runtime/request/request.go)
**SRP**: Runtime lookup facade

```go
func (r *Request) Security(symbol, timeframe, exprName string) (float64, error) {
    entry := r.cache.Get(symbol, timeframe)
    values := entry.Expressions[exprName]
    idx := r.findMatchingBarIndex(...)
    return values[idx], nil
}
```

**Why**: Thin API layer, business logic in separate modules

---

## Workflow Integration

### Build-Time
```go
// cmd/pinescript-builder/main.go
calls := security.AnalyzeAST(program)
codeGen.GeneratePrefetchCall(calls) // Inject prefetch before bar loop
```

### Generated Code
```go
func main() {
    // 1. Prefetch (BEFORE bar loop)
    fetcher := datafetcher.NewFileFetcher("./data", 50*time.Millisecond)
    prefetcher := security.NewPrefetcher(fetcher, cache)
    
    calls := []security.SecurityCall{
        {Symbol: "BTCUSDT", Timeframe: "1D", Expression: smaExpr},
    }
    prefetcher.Prefetch(calls, mainCtx)
    
    // 2. Bar Loop (cache hit only)
    for i := 0; i < len(bars); i++ {
        val, _ := reqHandler.Security("BTCUSDT", "1D", "daily_sma20")
        // ... use val
    }
}
```

---

## Design Rationale

### Why Prefetch Pattern?
- **Performance**: Zero I/O in bar loop (proven in PineTS)
- **Determinism**: All data fetched before execution
- **Parallelization**: Async fetch multiple symbols/timeframes

### Why Separate Analyzer?
- **SRP**: Detection ≠ execution
- **Testability**: Mock AST trees easily
- **Reusability**: Can analyze without executing

### Why DataFetcher Interface?
- **DIP**: High-level code independent of data source
- **Extensibility**: Add Binance/Polygon/CSV without touching core
- **Testing**: Mock fetcher returns deterministic data

### Why Expression Evaluator?
- **SRP**: Evaluation logic isolated from caching/fetching
- **Complexity**: TA calculations require full context access
- **Reusability**: Can evaluate arbitrary expressions

### Why Cache Module?
- **SRP**: Single storage responsibility
- **Concurrency**: Can add mutex for thread safety
- **Observability**: Single point for cache stats/debugging

---

## File Organization

```
security/
├── types.go          # SecurityCall, CacheEntry structs
├── analyzer.go       # AnalyzeAST()
├── prefetcher.go     # Prefetcher struct + Prefetch()
├── evaluator.go      # EvaluateExpression()
├── cache.go          # SecurityCache struct
└── analyzer_test.go  # Unit tests for each module

datafetcher/
├── fetcher.go        # Interface definition
├── file_fetcher.go   # FileFetcher implementation
└── file_fetcher_test.go
```

**Why**: 
- Grouped by functional domain (SOLID)
- Each file has single responsibility
- Easy to navigate: `security/analyzer.go` = "where analysis happens"
- Test files colocated with implementation

---

## Naming Conventions

| Entity | Naming | Example |
|--------|--------|---------|
| Interface | Noun (capability) | `DataFetcher` |
| Struct | Noun | `FileFetcher` |
| Method | Verb + Object | `Fetch()`, `AnalyzeAST()` |
| Package | Domain noun (lowercase) | `security`, `datafetcher` |

**Why**: Self-documenting code, no WHAT comments needed

---

## Error Handling

```go
// Prefetch errors: fail fast (before bar loop)
if err := prefetcher.Prefetch(calls, ctx); err != nil {
    return fmt.Errorf("prefetch failed: %w", err)
}

// Runtime errors: return NaN (graceful degradation)
val, err := req.Security(...)
if err != nil {
    log.Warn("security lookup failed", "err", err)
    return math.NaN()
}
```

**Why**: 
- Prefetch: Data missing = cannot proceed
- Runtime: Cache miss = log + continue with NaN

---

## Extension Points

### Adding Remote Fetcher
1. Implement `DataFetcher` interface in `remote_fetcher.go`
2. Pass to `Prefetcher` constructor
3. **Zero changes** to analyzer/cache/runtime

### Adding TSDB Fetcher
1. Implement `DataFetcher` interface in `questdb_fetcher.go` or `timescale_fetcher.go`
2. Pass to `Prefetcher` constructor
3. **Zero changes** to analyzer/cache/runtime

### Adding Redis Cache
1. Implement `CacheStorage` interface (new)
2. Inject into `SecurityCache`
3. **Zero changes** to fetcher/analyzer

### Supporting `request.dividends()`
1. Add `DividendCall` type
2. Add `DividendAnalyzer`
3. Reuse same `DataFetcher` interface
4. **Parallel module**, no coupling

---

## Testing Strategy

```go
// Analyzer: Pure function, easy to test
func TestAnalyzeAST(t *testing.T) {
    ast := parseCode("ma20 = security(tickerid, '1D', sma(close, 20))")
    calls := AnalyzeAST(ast)
    assert.Equal(t, "tickerid", calls[0].Symbol)
}

// FileFetcher: Mock filesystem
func TestFileFetcher(t *testing.T) {
    fetcher := NewFileFetcher("/tmp/test-data", 0)
    data, _ := fetcher.Fetch("BTC", "1h", 100)
    assert.Len(t, data, 100)
}

// Prefetcher: Mock DataFetcher interface
func TestPrefetcher(t *testing.T) {
    mockFetcher := &MockFetcher{...}
    prefetcher := NewPrefetcher(mockFetcher, cache)
    err := prefetcher.Prefetch(calls, ctx)
    assert.NoError(t, err)
}
```

---

## Performance Characteristics

| Operation | Complexity | Notes |
|-----------|------------|-------|
| AnalyzeAST | O(n) | n = AST nodes, single pass |
| Prefetch | O(k) | k = unique symbol-timeframe pairs, parallel |
| FileFetch (each) | O(1) + I/O | Read JSON file |
| Cache Lookup | O(1) | Map access |
| Runtime Security | O(log m) | m = bars, binary search for time match |

**Why prefetch wins**: 
- 3 security calls = 3 fetches before loop
- 1000 bars × 3 calls = **3000 cache hits** (zero I/O)
- Total: 3 I/O vs 3000 I/O

---

## Migration Path

### Phase 1: Core Infrastructure (NOW)
- [ ] `security/analyzer.go` - AST detection
- [ ] `datafetcher/file_fetcher.go` - Local JSON
- [ ] `security/cache.go` - Storage
- [ ] Unit tests for each

### Phase 2: Integration (NEXT)
- [ ] `security/prefetcher.go` - Orchestration
- [ ] `security/evaluator.go` - Expression execution
- [ ] Codegen integration (inject prefetch call)
- [ ] E2E test: daily-lines.pine

### Phase 3: Polish (LATER)
- [ ] `datafetcher/remote_fetcher.go` - HTTP API
- [ ] `datafetcher/questdb_fetcher.go` - TSDB implementation
- [ ] Concurrency safety (mutex in cache)
- [ ] Metrics/observability
- [ ] Performance benchmarks

---

## Counter-Suggestion: Why NOT inline everything?

**Alternative**: Put all logic in `runtime/request/request.go`

**Rejected because**:
- 500+ line file (violates SRP)
- Cannot test analyzer without full runtime
- Cannot swap fetcher without editing core
- Cannot reuse evaluator for other features
- Tight coupling = fragile code

**Evidence**: PineTS separated analyzer into standalone class, proven maintainability

---

## TSDB Integration for Production

### TSDB Selection

| TSDB | Query Latency | Throughput | Best For |
|------|---------------|------------|----------|
| **QuestDB** | 1-5ms | 4M rows/s | Financial OHLCV, SQL |
| **TimescaleDB** | 5-20ms | 1M rows/s | PostgreSQL ecosystem |
| **ClickHouse** | 10-50ms | 10M rows/s | Large-scale analytics |

### DataFetcher Interface Extension

```go
type DataFetcher interface {
    Fetch(symbol, timeframe string, limit int) ([]OHLCV, error)
    FetchRange(symbol, timeframe string, start, end time.Time) ([]OHLCV, error)
    FetchBatch(requests []FetchRequest) (map[string][]OHLCV, error)
}
```

### QuestDB Implementation

```go
// datafetcher/questdb_fetcher.go
type QuestDBFetcher struct {
    pool *pgxpool.Pool
}

func NewQuestDBFetcher(connStr string) (*QuestDBFetcher, error) {
    config, _ := pgxpool.ParseConfig(connStr)
    config.MaxConns = 20
    config.MinConns = 5
    config.MaxConnLifetime = 1 * time.Hour
    pool, _ := pgxpool.NewWithConfig(context.Background(), config)
    return &QuestDBFetcher{pool: pool}, nil
}

func (f *QuestDBFetcher) Fetch(symbol, timeframe string, limit int) ([]OHLCV, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    query := `
        SELECT timestamp, open, high, low, close, volume
        FROM ohlcv
        WHERE symbol = $1 AND timeframe = $2
        ORDER BY timestamp DESC
        LIMIT $3
    `
    
    rows, err := f.pool.Query(ctx, query, symbol, timeframe, limit)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    defer rows.Close()
    
    results := make([]OHLCV, 0, limit)
    for rows.Next() {
        var bar OHLCV
        err := rows.Scan(&bar.Time, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.Volume)
        if err != nil {
            return nil, err
        }
        results = append(results, bar)
    }
    return results, nil
}

func (f *QuestDBFetcher) FetchBatch(requests []FetchRequest) (map[string][]OHLCV, error) {
    results := make(map[string][]OHLCV)
    errChan := make(chan error, len(requests))
    
    for _, req := range requests {
        go func(r FetchRequest) {
            data, err := f.Fetch(r.Symbol, r.Timeframe, r.Limit)
            if err != nil {
                errChan <- err
                return
            }
            key := fmt.Sprintf("%s:%s", r.Symbol, r.Timeframe)
            results[key] = data
            errChan <- nil
        }(req)
    }
    
    for range requests {
        if err := <-errChan; err != nil {
            return nil, err
        }
    }
    return results, nil
}
```

### Query Optimization

**Time range queries (preferred over LIMIT)**:
```go
func (f *QuestDBFetcher) FetchRange(symbol, timeframe string, start, end time.Time) ([]OHLCV, error) {
    query := `
        SELECT timestamp, open, high, low, close, volume
        FROM ohlcv
        WHERE symbol = $1 
          AND timeframe = $2
          AND timestamp >= $3 
          AND timestamp < $4
        ORDER BY timestamp DESC
    `
    rows, _ := f.pool.Query(ctx, query, symbol, timeframe, start, end)
    // ...
}
```

### TSDB Schema

```sql
CREATE TABLE ohlcv (
    timestamp   TIMESTAMP,
    symbol      SYMBOL,
    timeframe   SYMBOL,
    open        DOUBLE,
    high        DOUBLE,
    low         DOUBLE,
    close       DOUBLE,
    volume      DOUBLE
) TIMESTAMP(timestamp) PARTITION BY DAY;

CREATE INDEX idx_symbol_time ON ohlcv (symbol, timestamp);
```

### Performance Expectations

| Operation | QuestDB | TimescaleDB |
|-----------|---------|-------------|
| Single symbol (1000 bars) | 1-3ms | 5-10ms |
| 10 symbols (parallel) | 5-15ms | 20-40ms |
| 100 symbols (parallel) | 30-80ms | 100-200ms |

### Cache Optimization

```go
type CacheEntry struct {
    Times       []int64
    Data        map[string][]float64
    indexCache  map[int64]int
    maxCacheSize int
}

func (c *CacheEntry) Get(exprName string, timestamp int64) (float64, error) {
    if idx, ok := c.indexCache[timestamp]; ok {
        return c.Data[exprName][idx], nil
    }
    
    idx := sort.Search(len(c.Times), func(i int) bool {
        return c.Times[i] >= timestamp
    })
    
    if idx >= len(c.Times) {
        return 0, fmt.Errorf("timestamp not found")
    }
    
    if len(c.indexCache) < c.maxCacheSize {
        c.indexCache[timestamp] = idx
    }
    
    return c.Data[exprName][idx], nil
}
```
