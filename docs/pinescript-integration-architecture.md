## Objective

Enable direct `.pine` file import and execution using **pynescript → PineTS transpilation bridge**.

### TODO

- [ ] Non-functional Refactoring of plot adapter added at 4286558781c8215bb3c3255726084233dfb5db7a - [ ] commit to make it testable with tests, and decoupled from string template where it's being injected
- [ ] Implement remaining logic of `security()` method
- [ ] Extend `security()` method to support both higher and lower timeframes (will require adjusting of PineTS source code in a sibling dir, committing PineTS to our repository and rebuilding the PineTS)
- [ ] Fix failing tests by fitting tests and mocks to newly adjusted code
- [ ] Increase test coverage to 80
- [ ] Increase test coverage to 95
- [ ] Debug and fix any issues with `daily-lines` strategy on any timeframe
- [ ] Debug and fix any issues with `rolling-cagr` streategy on any timeframe
- [ ] Design and plan extension of existing code which is necessary for BB strategies v7, 8, 9
  - [ ] Implement Pine Script `strategy.*` to trading signals
  - [ ] Handle Pine Script `alert()` conditions

## Performance Optimization Strategy

### Baseline Metrics

- Current system: <1s for 100 candles (pure JavaScript)
- Target with .pine import: <2s for 100 candles (includes transpilation)

### Optimization TODO

- [ ] Implement in-memory AST cache (avoid re-parsing)
- [ ] Pre-transpile strategies at container startup
- [ ] Use persistent Python process pool
- [ ] Measure and profile each pipeline stage
- [ ] Add performance monitoring to StrategyExecutor

### Extra: Testing & Validation

- [ ] Create test suite for pynescript transpilation
- [ ] Add example `.pine` strategies (EMA, RSI, MACD)
- [ ] Validate transpiled code against original behavior
- [ ] Benchmark performance: `.pine` vs inline JavaScript
- [ ] Test edge cases: complex indicators, nested conditionals
- [ ] Verify all Pine Script v5 technical analysis functions

### Extra: Production Deployment

- [ ] Optimize Docker image size (multi-stage build)
- [ ] Add transpilation result caching (Redis/filesystem)
- [ ] Implement error recovery and fallback strategies
- [ ] Create monitoring for parser service health
- [ ] Add strategy execution timeouts
- [ ] Document deployment procedures