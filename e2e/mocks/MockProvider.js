/**
 * MockProvider - Deterministic data provider for E2E tests
 *
 * Provides 100% predictable candle data for regression testing.
 * Benefits:
 * - No network dependencies (fast, reliable)
 * - Exact expected values can be calculated
 * - Tests never flaky
 * - Can test edge cases easily
 */

export class MockProvider {
  constructor(config = {}) {
    this.dataPattern = config.dataPattern || 'linear'; // 'linear', 'constant', 'random', 'edge', 'sawtooth', 'bullish', 'bearish'
    this.basePrice = config.basePrice || 1;
    this.amplitude = config.amplitude || 10; // For sawtooth pattern
    this.supportedTimeframes = ['1m', '5m', '15m', '30m', '1h', '4h', 'D', 'W', 'M'];
  }

  /**
   * Generate deterministic candle data
   * @param {string} symbol - Symbol name (ignored in mock)
   * @param {string} timeframe - Timeframe (used for timestamp calculation)
   * @param {number} limit - Number of candles to generate
   * @returns {Array} Array of candle objects
   */
  async getMarketData(symbol, timeframe, limit = 100) {
    const candles = [];
    const now = Math.floor(Date.now() / 1000); // Current Unix timestamp
    const timeframeSeconds = this.getTimeframeSeconds(timeframe);

    for (let i = 0; i < limit; i++) {
      const price = this.generatePrice(i);

      /* For sawtooth pattern, high/low should match close to create clear pivots */
      const high = this.dataPattern === 'sawtooth' ? price : price + 1;
      const low = this.dataPattern === 'sawtooth' ? price : price - 1;

      candles.push({
        time: now - (limit - 1 - i) * timeframeSeconds, // Work backwards from now
        open: price,
        high,
        low,
        close: price,
        volume: 1000 + i,
      });
    }

    return candles;
  }

  /**
   * Generate price based on pattern
   */
  generatePrice(index) {
    switch (this.dataPattern) {
      case 'linear':
        // close = [1, 2, 3, 4, 5, ...]
        return this.basePrice + index;

      case 'constant':
        // close = [100, 100, 100, ...]
        return this.basePrice;

      case 'random':
        // Deterministic "random" using index as seed
        return this.basePrice + ((index * 7) % 50);

      case 'sawtooth': {
        // Zigzag pattern creates clear pivot highs and lows
        // Pattern: 100, 105, 110, 105, 100, 95, 100, 105, 110...
        // Cycle: [0, 5, 10, 5, 0, -5] repeating
        const cycle = index % 6;
        const offsets = [0, 5, 10, 5, 0, -5];
        return this.basePrice + offsets[cycle];
      }

      case 'edge': {
        // Test edge cases: 0, negative, very large
        const patterns = [0, -100, 0.0001, 999999, NaN];
        return patterns[index % patterns.length];
      }

      case 'bullish': {
        // Uptrend with small dips: creates long entries
        // Pattern oscillates ABOVE baseline, trending up
        const trend = index * 0.5; // Gradual uptrend
        const cycle = index % 4;
        const offsets = [0, 2, 1, 3]; // Small oscillation
        return this.basePrice + trend + offsets[cycle];
      }

      case 'bearish': {
        // Downtrend with small bounces: creates short entries
        // Pattern oscillates BELOW baseline, trending down
        const trend = -index * 0.5; // Gradual downtrend
        const cycle = index % 4;
        const offsets = [0, -2, -1, -3]; // Small oscillation
        return this.basePrice + trend + offsets[cycle];
      }

      default:
        return this.basePrice + index;
    }
  }

  /**
   * Convert timeframe to seconds
   */
  getTimeframeSeconds(timeframe) {
    const map = {
      '1m': 60,
      '5m': 300,
      '15m': 900,
      '30m': 1800,
      '1h': 3600,
      '4h': 14400,
      D: 86400,
      W: 604800,
      M: 2592000, // ~30 days
    };
    return map[timeframe] || 86400;
  }
}

/**
 * MockProviderManager - Wraps MockProvider to match ProviderManager interface
 */
export class MockProviderManager {
  constructor(config = {}) {
    this.mockProvider = new MockProvider(config);
  }

  async getMarketData(symbol, timeframe, limit) {
    return await this.mockProvider.getMarketData(symbol, timeframe, limit);
  }

  // Implement other ProviderManager methods if needed
  getStats() {
    return {
      totalRequests: 0,
      cacheHits: 0,
      cacheMisses: 0,
      byProvider: { Mock: { requests: 0, symbols: new Set() } },
    };
  }
}
